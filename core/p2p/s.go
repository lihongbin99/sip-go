package p2p

import (
	"sip-go/common/config"
	"sip-go/common/io"
	"sip-go/common/msg"
	"time"
)

func SStartService(ccTCP *io.TCP) {
	// 接收cc请求参数
	p2pCcRequest := &msg.P2pCcRequest{}
	if err := ccTCP.ReadToMessage(time.Now().Add(3*time.Second), p2pCcRequest); err != nil {
		log.Debug("read p2p cc request error:", err)
		return
	}
	info := infoInit(ccTCP.Id, ccTCP.RemoteAddr().String(), p2pCcRequest)
	log.Debug("p2p cc request ip:", info.ccIp, info.ccPort)

	notifyCs(info, p2pCcRequest.Key, p2pCcRequest.Iv)

	// 向cc返回请求结果
	_ = ccTCP.WriteMessage(&msg.P2pCcResponse{
		YourIp: info.ccIp, YourPort: info.ccPort,
		TargetIp: info.csIp, TargetPort: info.csPort,
		Message: info.message,
	})

	infoUnInit(ccTCP.Id)
	log.Info("p2p:",
		info.ccId, info.ccName, info.ccIp, info.ccPort, "->",
		info.csName, info.csIp, info.csPort, "[", info.targetIp, info.targetPort, "]:",
		info.message,
	)
}

func notifyCs(info *Info, key, iv []byte) {
	// 通知cs
	csClientTCP := io.GetClient(info.csName)
	if csClientTCP == nil {
		info.message = "client not exist"
		return
	}
	_ = csClientTCP.WriteMessage(&msg.P2pCsRequest{
		CcId: info.ccId, CcName: info.ccName, CcIp: info.ccIp, CcPort: info.ccPort,
		TargetIp: info.targetIp, TargetPort: info.targetPort,
		MakeHoleTime: info.makeHoleTime,
		Key:          key, Iv: iv,
	})

	// 等待结果后在返回
	log.Debug(info.ccId, "start wait notify cs result")
	// TODO 此处可能卡死(没返回)
	_ = <-info.waitChan
}

func NotifyCsResult(p2pCsResponse *msg.P2pCsResponse) {
	// 接收cs结果
	info := getInfo(p2pCsResponse.CcId)
	if info == nil {
		log.Error("notify cs result not find ccId:", p2pCsResponse.CcId)
		return
	}
	info.message = p2pCsResponse.Message
	info.waitChan <- 1
	log.Debug(info.ccId, "notify cs result", info.message)
}

func CsNewConnect(csTCP *io.TCP) {
	// 读取参数
	p2pCsNewConnectRequest := &msg.P2pCsNewConnectRequest{}
	if err := csTCP.ReadToMessage(time.Now().Add(3*time.Second), p2pCsNewConnectRequest); err != nil {
		log.Error("read p2p cs new connect request error: %v", err)
		return
	}

	// 设置
	info := getInfo(p2pCsNewConnectRequest.CcId)
	if info == nil {
		log.Error("cs new connect not find ccId:", p2pCsNewConnectRequest.CcId)
		_ = csTCP.WriteMessage(&msg.P2pCsNewConnectResponse{Message: "server from cs new connect not find ccId"})
		return
	}
	info.updateInfo(csTCP.RemoteAddr().String())

	// 返回结果
	_ = csTCP.WriteMessage(&msg.P2pCsNewConnectResponse{
		YourIp: info.csIp, YourPort: info.csPort,
		Message: config.SUCCESS,
	})
}
