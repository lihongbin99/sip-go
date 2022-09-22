package main

import (
	"fmt"
	"net"
	"sip-go/common/config"
	"sip-go/core/p2p"
)

func startService() {
	if config.ClientConfig.P2p != nil {
		for _, p2pConfig := range config.ClientConfig.P2p {
			go doStartService(p2pConfig.ClientName, p2pConfig.LocalPort, p2pConfig.RemoteIp, p2pConfig.RemotePort, p2pConfig.MakeHoleTime, "p2p")
		}
	}
	if config.ClientConfig.Proxy != nil {
		for _, proxyConfig := range config.ClientConfig.Proxy {
			go doStartService(proxyConfig.ClientName, proxyConfig.LocalPort, proxyConfig.RemoteIp, proxyConfig.RemotePort, 0, "proxy")
		}
	}
}

func doStartService(clientName string, localPort int, remoteIp string, remotePort int, makeHoleTime int, configType string) {
	serviceAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		log.Error("resolve tcp addr error", err)
		return
	}

	serviceListen, err := net.ListenTCP("tcp", serviceAddr)
	if err != nil {
		log.Error("listen tcp error", err)
		return
	}

	for {
		conn, err := serviceListen.AcceptTCP()
		if err != nil {
			log.Error("accept tcp error", err)
			return
		}

		go func() {
			defer func() {
				_ = conn.Close()
			}()

			switch configType {
			case "p2p":
				p2p.CcStartService(serverAddr, conn, config.ClientConfig.ClientName, clientName, remoteIp, remotePort, makeHoleTime)
			case "proxy":
				// TODO 未实现 proxy 模式
			}
		}()
	}
}
