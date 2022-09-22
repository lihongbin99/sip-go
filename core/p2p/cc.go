package p2p

import (
	"fmt"
	"net"
	"sip-go/common/config"
	"sip-go/common/io"
	"sip-go/common/logger"
	"sip-go/common/msg"
	"time"
)

var (
	log = logger.NewLog("P2P")
)

func CcStartService(serverAddr *net.TCPAddr, localConn *net.TCPConn, thisName, targetName, remoteIp string, remotePort, makeHoleTime int) {
	// 创建新连接并且获取返回结果
	id, local2ServerAddr, myRemoteAddr, remoteAddr, key, iv, err := ccNewP2pConnect(serverAddr, thisName, targetName, remoteIp, remotePort, makeHoleTime)

	// p2p
	var p2pConn net.Conn
	if err == nil {
		p2pConn, err = p2p(local2ServerAddr, remoteAddr)
	}

	if err != nil {
		log.Error("new p2p error:", myRemoteAddr, "->",
			targetName, remoteAddr, "[", remoteIp, remotePort, "]", ":", err.Error())
		return
	}
	log.Info(id, "new p2p success:", myRemoteAddr, "->",
		targetName, remoteAddr, "[", remoteIp, remotePort, "]")

	// 传输数据
	p2pTCP := io.NewTCPByKey(p2pConn, key, iv)
	p2pTransferData(localConn, p2pTCP)

	log.Info(id, "p2p finish:", myRemoteAddr, "->",
		targetName, remoteAddr, "[", remoteIp, remotePort, "]")
}

func ccNewP2pConnect(serverAddr *net.TCPAddr, thisName, targetName, remoteIp string, remotePort, makeHoleTime int) (int, net.Addr, net.Addr, net.Addr, []byte, []byte, error) {
	// 创建新连接
	serverConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return 0, nil, nil, nil, nil, nil, fmt.Errorf("connect server error: %v", err)
	}
	defer func() {
		_ = serverConn.Close()
	}()

	// 获取本地系统分配的地址
	local2ServerAddr := serverConn.LocalAddr()

	// 初始化连接
	serverTCP := io.NewTCP(serverConn)
	if err = serverTCP.ClientInit(msg.P2pCc); err != nil {
		return 0, nil, nil, nil, nil, nil, fmt.Errorf("init server connect error: %v", err)
	}

	// 发送请求
	_ = serverTCP.WriteMessage(&msg.P2pCcRequest{
		CcName: thisName, CsName: targetName,
		TargetIp: remoteIp, TargetPort: remotePort, MakeHoleTime: makeHoleTime,
		Key: serverTCP.Key, Iv: serverTCP.Iv,
	})

	// 接收响应
	p2pCcResponse := &msg.P2pCcResponse{}
	if err = serverTCP.ReadToMessage(time.Now().Add(8*time.Second), p2pCcResponse); err != nil {
		return serverTCP.Id, nil, nil, nil, nil, nil, fmt.Errorf("read p2p response error: %v", err)
	}

	myRemoteAddr, _ := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", p2pCcResponse.YourIp, p2pCcResponse.YourPort))
	remoteAddr, _ := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", p2pCcResponse.TargetIp, p2pCcResponse.TargetPort))

	if p2pCcResponse.Message != config.SUCCESS {
		return serverTCP.Id, nil, myRemoteAddr, remoteAddr, nil, nil, fmt.Errorf("p2p get remote addr error: %v", p2pCcResponse.Message)
	}

	return serverTCP.Id, local2ServerAddr, myRemoteAddr, remoteAddr, serverTCP.Key, serverTCP.Iv, nil
}
