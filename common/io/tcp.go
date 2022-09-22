package io

import (
	"fmt"
	"net"
	"sip-go/common"
	"sip-go/common/config"
	"sip-go/common/msg"
	"sip-go/common/security"
	"sip-go/common/utils"
	"strconv"
	"time"
)

type TCP struct {
	net.Conn
	buf []byte
	Key []byte
	Iv  []byte

	Interval int

	Id int
}

var (
	bufLen uint32 = 65 * 1024
)

func NewTCP(conn net.Conn) *TCP {
	return &TCP{conn, make([]byte, bufLen), nil, nil, 0, 0}
}

func NewTCPByKey(conn net.Conn, key, iv []byte) *TCP {
	return &TCP{conn, make([]byte, bufLen), key, iv, 0, 0}
}

func (that *TCP) ReadByLen(maxReadLen uint32, timeout time.Time) ([]byte, error) {
	_ = that.SetReadDeadline(timeout)

	var buf []byte = nil
	if maxReadLen <= bufLen {
		buf = that.buf[:maxReadLen]
	} else {
		log.Warn("create big buf", maxReadLen)
		buf = make([]byte, maxReadLen)
	}

	var readSum uint32 = 0
	for readSum < maxReadLen {
		readLen, err := that.Read(buf[readSum:])
		if err != nil {
			return nil, err
		}
		readSum += uint32(readLen)
	}

	_ = that.SetReadDeadline(time.Time{})
	return buf, nil
}

func (that *TCP) ServerInit() error {
	// 接受协议
	protocol, err := that.ReadByLen(common.ProtocolLen, time.Now().Add(3*time.Second))
	if err != nil {
		return fmt.Errorf("read protocol error: %v", err)
	}
	protocolLen := len(protocol)
	if string(protocol[:protocolLen-3]) != common.AppName {
		return fmt.Errorf("protocol error")
	}
	clientVersion :=
		strconv.Itoa(int(protocol[protocolLen-3])) + "." +
			strconv.Itoa(int(protocol[protocolLen-2])) + "." +
			strconv.Itoa(int(protocol[protocolLen-1]))
	log.Trace("client version:", clientVersion)

	// 返回协议结果
	if _, err = that.Write(common.Protocol); err != nil {
		return fmt.Errorf("write protocol error: %v", err)
	}

	// 接受密钥
	message, err := that.ReadByLen(256, time.Now().Add(8*time.Second))
	if err != nil {
		return fmt.Errorf("read key error: %v", err)
	}
	message, err = security.DecryptRSA(message, config.ServerConfig.PrivateKey)
	if err != nil {
		return fmt.Errorf("DecryptRSA error: %v", err)
	}
	if len(message) != 32 {
		return fmt.Errorf("key error len: %d", len(message))
	}
	that.Key = message[:16]
	that.Iv = message[16:]

	// 返回密钥结果(心跳时间)
	that.Interval = config.ServerConfig.Interval
	encrypt, err := security.AesEncrypt([]byte(strconv.Itoa(that.Interval)), that.Key, that.Iv)
	if err != nil {
		return fmt.Errorf("aes encrypt error: %v", err)
	}
	if _, err = that.Write(encrypt); err != nil {
		return fmt.Errorf("write interval error: %v", err)
	}

	return nil
}

func (that *TCP) ClientInit(registerType msg.RegisterType) error {
	// 发送协议
	if _, err := that.Write(common.Protocol); err != nil {
		return fmt.Errorf("write protocol error: %v", err)
	}

	// 接受协议结果
	protocol, err := that.ReadByLen(common.ProtocolLen, time.Now().Add(8*time.Second))
	if err != nil {
		return fmt.Errorf("read protocol error: %v", err)
	}
	protocolLen := len(protocol)
	if string(protocol[:protocolLen-3]) != common.AppName {
		return fmt.Errorf("protocol error")
	}
	serverVersion :=
		strconv.Itoa(int(protocol[protocolLen-3])) + "." +
			strconv.Itoa(int(protocol[protocolLen-2])) + "." +
			strconv.Itoa(int(protocol[protocolLen-1]))
	log.Trace("server version:", serverVersion)

	// 发送密钥
	if that.Key == nil || that.Iv == nil {
		that.Key, that.Iv = security.GenerateAES()
	}
	message := make([]byte, 32)
	copy(message[0:16], that.Key)
	copy(message[16:], that.Iv)
	message, err = security.EncryptRSA(message, config.ClientConfig.PublicKey)
	if err != nil {
		return fmt.Errorf("EncryptRSA error: %v", err)
	}
	if _, err = that.Write(message); err != nil {
		return fmt.Errorf("write key error: %v", err)
	}

	// 接收密钥结果(心跳时间)
	message, err = that.ReadByLen(16, time.Now().Add(8*time.Second))
	if err != nil {
		return fmt.Errorf("read key result error: %v", err)
	}
	decrypt, err := security.AesDecrypt(message, that.Key, that.Iv)
	if err != nil {
		return fmt.Errorf("aes decrypt error: %v", err)
	}
	interval, err := strconv.Atoi(string(decrypt))
	if err != nil {
		return fmt.Errorf("interval error: %v", err)
	}
	that.Interval = interval

	// 发送注册消息
	_ = that.WriteMessage(
		&msg.RegisterMessage{Name: config.ClientConfig.ClientName, RegisterType: registerType},
	)

	// 接收注册结果
	registerResultMessage := &msg.RegisterResultMessage{}
	if err = that.ReadToMessage(time.Now().Add(8*time.Second), registerResultMessage); err != nil {
		return fmt.Errorf("read register message error: %v", err)
	}
	if registerResultMessage.Message != config.SUCCESS {
		return fmt.Errorf("register error: %v", err)
	}
	that.Id = registerResultMessage.Id

	return nil
}

func (that *TCP) WriteMessage(message msg.Message) error {
	// 解析
	var data []byte
	var err error
	if message.GetMessageType() == msg.DataMessageType {
		data = message.(*msg.DataMessage).Data
	} else {
		data, err = msg.ToByte(message)
		if err != nil {
			return err
		}
	}
	if len(data) <= 0 {
		return nil
	}

	// 加密
	if data, err = security.AesEncrypt(data, that.Key, that.Iv); err != nil {
		return err
	}
	if len(data) <= 0 {
		return nil
	}

	// 发送消息类型
	if _, err = that.Write(utils.I2b32(message.GetMessageType())); err != nil {
		return err
	}
	// 发送消息长度
	if _, err = that.Write(utils.I2b32(uint32(len(data)))); err != nil {
		return err
	}
	// 发送消息
	if _, err = that.Write(data); err != nil {
		return err
	}
	return nil
}

func (that *TCP) ReadMessage(timeout time.Time) Message {
	// 读取前缀
	data, err := that.ReadByLen(8, timeout)
	if err != nil {
		return Message{Err: err}
	}

	// 获取消息类型
	messageType, err := utils.B2i32(data[:4])
	if err != nil {
		return Message{Err: err}
	}
	message, err := msg.NewMessage(messageType)
	if err != nil {
		return Message{Err: err}
	}

	// 获取消息长度
	messageLen, err := utils.B2i32(data[4:8])
	if err != nil {
		return Message{Err: err}
	}
	if messageLen <= 0 {
		return Message{Err: fmt.Errorf("message len: %d", messageLen)}
	}

	// 读取消息
	data, err = that.ReadByLen(messageLen, timeout)
	if err != nil {
		return Message{Err: err}
	}

	// 解密
	if data, err = security.AesDecrypt(data, that.Key, that.Iv); err != nil {
		return Message{Err: err}
	}

	// 解析
	if message.GetMessageType() == msg.DataMessageType {
		message.(*msg.DataMessage).Data = data
	} else {
		if err = msg.ToObj(data, message); err != nil {
			return Message{Err: err}
		}
	}

	return Message{Message: message, Err: nil}
}

func (that *TCP) ReadToMessage(timeout time.Time, message msg.Message) error {
	// 读取前缀
	data, err := that.ReadByLen(8, timeout)
	if err != nil {
		return err
	}

	// 获取消息类型
	messageType, err := utils.B2i32(data[:4])
	if err != nil {
		return err
	}
	if message.GetMessageType() != messageType {
		return fmt.Errorf("message type error")
	}

	// 获取消息长度
	messageLen, err := utils.B2i32(data[4:8])
	if err != nil {
		return err
	}
	if messageLen <= 0 {
		return fmt.Errorf("message len: %d", messageLen)
	}

	// 读取消息
	data, err = that.ReadByLen(messageLen, timeout)
	if err != nil {
		return err
	}

	// 解密
	if data, err = security.AesDecrypt(data, that.Key, that.Iv); err != nil {
		return err
	}

	// 解析
	if err = msg.ToObj(data, message); err != nil {
		return err
	}

	return nil
}
