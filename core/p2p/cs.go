package p2p

import (
	"fmt"
	"net"
	"sip-go/common/config"
	"sip-go/common/io"
	"sip-go/common/msg"
	"time"
)

func CsStartService(baseTCP *io.TCP, serverAddr *net.TCPAddr, p2pCsRequest *msg.P2pCsRequest) {
	// 创建新连接并且获取返回结果
	id, localConn, local2ServerAddr, remoteAddr, key, iv, err := csNewP2pConnect(serverAddr, p2pCsRequest)
	defer func() {
		if localConn != nil {
			_ = localConn.Close()
		}
	}()

	// 打洞
	if err == nil {
		makeHole(local2ServerAddr, remoteAddr, p2pCsRequest.MakeHoleTime)
	}

	// 向server返回结果
	if err != nil {
		_ = baseTCP.WriteMessage(&msg.P2pCsResponse{CcId: p2pCsRequest.CcId, Message: err.Error()})
	} else {
		_ = baseTCP.WriteMessage(&msg.P2pCsResponse{CcId: p2pCsRequest.CcId, Message: config.SUCCESS})
	}

	// p2p
	var p2pConn net.Conn
	if err == nil {
		p2pConn, err = p2p(local2ServerAddr, remoteAddr)
	}

	if err != nil {
		log.Error("new p2p error:", p2pCsRequest.CcName, remoteAddr, "->",
			p2pCsRequest.TargetIp, p2pCsRequest.TargetPort, ":", err.Error())
		return
	}
	log.Info(id, "new p2p success:", p2pCsRequest.CcName, remoteAddr, "->",
		p2pCsRequest.TargetIp, p2pCsRequest.TargetPort)

	// 传输数据
	p2pTCP := io.NewTCPByKey(p2pConn, key, iv)
	p2pTransferData(localConn, p2pTCP)

	log.Info(id, "p2p finish:", p2pCsRequest.CcName, remoteAddr, "->",
		p2pCsRequest.TargetIp, p2pCsRequest.TargetPort)
}

func csNewP2pConnect(serverAddr *net.TCPAddr, p2pCsRequest *msg.P2pCsRequest) (int, *net.TCPConn, net.Addr, net.Addr, []byte, []byte, error) {
	remoteAddr, _ := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", p2pCsRequest.CcIp, p2pCsRequest.CcPort))

	// 连接本地服务
	localServiceConn, err := createLocalServiceConn(p2pCsRequest.TargetIp, p2pCsRequest.TargetPort)
	if err != nil {
		return p2pCsRequest.CcId, nil, nil, remoteAddr, nil, nil, err
	}

	// 创建新链接
	serverConn, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		return p2pCsRequest.CcId, localServiceConn, nil, remoteAddr, nil, nil, fmt.Errorf(
			"connect to server error: %v", err)
	}
	defer func() {
		_ = serverConn.Close()
	}()

	// 获取本地系统分配的地址
	local2ServerAddr := serverConn.LocalAddr()

	// 初始化连接
	serverTCP := io.NewTCPByKey(serverConn, p2pCsRequest.Key, p2pCsRequest.Iv)
	if err = serverTCP.ClientInit(msg.P2pCs); err != nil {
		return p2pCsRequest.CcId, localServiceConn, local2ServerAddr, remoteAddr, nil, nil, fmt.Errorf(
			"init server connect error: %v", err)
	}

	// 发送新连接参数
	_ = serverTCP.WriteMessage(&msg.P2pCsNewConnectRequest{CcId: p2pCsRequest.CcId})
	// 接收返回结果
	p2pCsNewConnectResponse := &msg.P2pCsNewConnectResponse{}
	if err = serverTCP.ReadToMessage(time.Now().Add(3*time.Second), p2pCsNewConnectResponse); err != nil {
		return p2pCsRequest.CcId, localServiceConn, local2ServerAddr, remoteAddr, nil, nil, fmt.Errorf(
			"read p2p cs new connect response error: %v", err)
	}
	if p2pCsNewConnectResponse.Message != config.SUCCESS {
		return p2pCsRequest.CcId, localServiceConn, local2ServerAddr, remoteAddr, nil, nil, fmt.Errorf(
			"p2p cs new connect response error: %s", p2pCsNewConnectResponse.Message)
	}

	return p2pCsRequest.CcId, localServiceConn, local2ServerAddr, remoteAddr, serverTCP.Key, serverTCP.Iv, nil
}

func createLocalServiceConn(ip string, port int) (*net.TCPConn, error) {
	localServiceAddr, err := net.ResolveTCPAddr("tcp",
		fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return nil, fmt.Errorf("resolve local service tcp addr error: %v", err)
	}
	localServiceConn, err := net.DialTCP("tcp", nil, localServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("connect local service error: %v", err)
	}
	return localServiceConn, nil
}

func makeHole(local2ServerAddr, remoteAddr net.Addr, makeHoleTime int) {
	// 发送探测包
	dialer := net.Dialer{Timeout: time.Duration(makeHoleTime) * time.Millisecond, LocalAddr: local2ServerAddr}
	if conn, err := dialer.Dial("tcp", remoteAddr.String()); err == nil {
		_ = conn.Close()
	}
}
