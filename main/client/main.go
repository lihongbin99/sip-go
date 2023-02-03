package main

import (
	"flag"
	"fmt"
	"net"
	"sip-go/common/config"
	"sip-go/common/io"
	"sip-go/common/logger"
	"sip-go/common/msg"
	"sip-go/core/p2p"
	"time"
)

func init() {
	flag.Parse()
	config.ClientConfigInit()
}

var (
	log = logger.NewLog("Client")

	serverAddr *net.TCPAddr
)

func main() {
	// 开启服务
	startService()

	interval := 1
	for {
		success := startClient()
		if success {
			interval = 1
		}
		log.Info("sleep:", interval)
		time.Sleep(time.Duration(interval) * time.Second)
		interval = interval * 2
		if interval > 60 {
			interval = 60
		}
	}
}

func startClient() (success bool) {
	var err error
	// 连接服务器
	serverAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",
		config.ClientConfig.ServerConnect.Ip, config.ClientConfig.ServerConnect.Port))
	if err != nil {
		log.Error("resolve tcp addr error", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		log.Error("dial tcp error", err)
		return
	}
	defer func() {
		_ = conn.Close()
		log.Info("close server success")
	}()
	log.Debug("connect server success")

	tcp := io.NewTCP(conn)

	if err = tcp.ClientInit(msg.RegisterTypeClient); err != nil {
		log.Error("client init error", err)
		return
	}
	log.Debug("client init success")

	newConn(tcp)
	return true
}

func newConn(serverTCP *io.TCP) {
	log.Info(serverTCP.Id, "client register success:", config.ClientConfig.ClientName)

	// 处理读取请求
	readChan := make(chan io.Message, 0)
	go func(tcp *io.TCP, readChan chan io.Message) {
		for {
			message := tcp.ReadMessage(time.Time{})
			readChan <- message
			if message.Err != nil {
				break
			}
		}
	}(serverTCP, readChan)

	// 心跳
	pingTicker := time.NewTicker(time.Duration(serverTCP.Interval+10000) * time.Millisecond)
	defer pingTicker.Stop()
	lastPongTime := time.Now()
	lastPingTime := time.Now()

	var err error = nil
	for err == nil {
		select {
		case <-pingTicker.C:
			if lastPongTime.Before(lastPingTime) {
				err = fmt.Errorf("ping timeout")
			}
			lastPingTime = time.Now()
		case message := <-readChan:
			if message.Err != nil {
				err = message.Err
				break
			}
			switch m := message.Message.(type) {
			case *msg.PingMessage:
				lastPongTime = time.Now()
				log.Trace("receiver PingMessage", m.Date)
				err = serverTCP.WriteMessage(&msg.PongMessage{Date: time.Now()})
			case *msg.P2pCsRequest:
				go p2p.CsStartService(serverTCP, serverAddr, m)
			}
		}
	}

	cleanChan := true
	for cleanChan {
		select {
		case _ = <-readChan:
		default:
			cleanChan = false
		}
	}
	close(readChan)

	log.Info(serverTCP.Id, "client,", config.ClientConfig.ClientName, "exit:", err)
}
