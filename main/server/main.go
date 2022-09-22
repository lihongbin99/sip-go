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
	config.ServerConfigInit()
}

var (
	id  = 0
	log = logger.NewLog("Server")
)

func main() {
	listenAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",
		config.ServerConfig.Listen.Ip, config.ServerConfig.Listen.Port))
	if err != nil {
		panic(err)
	}

	for {
		startServer(listenAddr)
		time.Sleep(time.Minute)
	}
}

func startServer(listenAddr *net.TCPAddr) {
	// 监听连接
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Error("listen tcp error", err)
		return
	}

	log.Info("start server success")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Error("accept tcp error", err)
			break
		}

		id++
		go newConn(id, conn)
	}
	_ = listener.Close()
}

func newConn(id int, conn *net.TCPConn) {
	defer func() {
		_ = conn.Close()
		log.Debug(id, "close client success")
	}()
	log.Debug(id, "connect client success")

	tcp := io.NewTCP(conn)
	tcp.Id = id

	if err := tcp.ServerInit(); err != nil {
		log.Debug("server init error", err)
		return
	}
	log.Debug(id, "server init success")

	// 接收注册消息
	registerMessage := &msg.RegisterMessage{}
	if err := tcp.ReadToMessage(time.Now().Add(8*time.Second), registerMessage); err != nil {
		log.Debug("read register message error:", err)
		return
	}

	// 根据不同的连接类型返回注册结果
	switch registerMessage.RegisterType {
	case msg.RegisterTypeClient:
		// 注册
		if !io.RegisterClient(registerMessage.Name, tcp) {
			_ = tcp.WriteMessage(&msg.RegisterResultMessage{Message: "name exist"})
			return
		}
	case msg.P2pCc:
	case msg.P2pCs:
	default:
		_ = tcp.WriteMessage(&msg.RegisterResultMessage{Message: "register type error"})
		return
	}
	_ = tcp.WriteMessage(&msg.RegisterResultMessage{Id: id, Message: config.SUCCESS})

	// 根据不同的连接类型进行处理
	switch registerMessage.RegisterType {
	case msg.P2pCc:
		p2p.SStartService(tcp)
		return
	case msg.P2pCs:
		p2p.CsNewConnect(tcp)
		return
	}

	log.Info(id, "client register success:", registerMessage.Name)

	// 处理读取请求
	readChan := make(chan io.Message, 8)
	go func(tcp *io.TCP, readChan chan io.Message) {
		defer close(readChan)
		for {
			message := tcp.ReadMessage(time.Time{})
			readChan <- message
			if message.Err != nil {
				break
			}
		}
	}(tcp, readChan)

	// 心跳
	pingTicker := time.NewTicker(time.Duration(tcp.Interval) * time.Millisecond)
	defer pingTicker.Stop()
	lastPingTime := time.Now()
	lastPongTime := time.Now()

	// main
	var err error = nil
	for err == nil {
		select {
		case <-pingTicker.C:
			lastPingTime = time.Now()
			log.Trace(id, "send PingMessage")
			err = tcp.WriteMessage(&msg.PingMessage{Date: lastPingTime})
			go func() {
				time.Sleep(10 * time.Second)
				if lastPongTime.Before(lastPingTime) {
					log.Warn(id, "ping timeout")
					_ = tcp.Close() // 此处直接关闭连接, 让read线程退出方法
				}
			}()
		case message := <-readChan:
			if message.Err != nil {
				err = message.Err
				break
			}
			switch m := message.Message.(type) {
			case *msg.PongMessage:
				log.Trace(id, "receiver Pong", m.Date)
				lastPongTime = time.Now()
			case *msg.P2pCsResponse:
				go p2p.NotifyCsResult(m)
			}
		}
	}

	log.Info(id, "client,", registerMessage.Name, "exit:", err)
	io.RemoveClient(registerMessage.Name)
}
