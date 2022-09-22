package p2p

import (
	"sip-go/common/msg"
	"strconv"
	"strings"
	"sync"
)

type Info struct {
	ccId int

	ccName string
	ccIp   string
	ccPort int

	csName string
	csIp   string
	csPort int

	targetIp   string
	targetPort int

	makeHoleTime int

	message string

	waitChan chan uint8
}

func (that *Info) updateInfo(csAddr string) {
	split := strings.Split(csAddr, ":")
	csIp := split[0]
	csPort, _ := strconv.Atoi(split[1])

	that.csIp = csIp
	that.csPort = csPort
}

var (
	ccMap     = make(map[int]*Info)
	ccMapLock = sync.Mutex{}
)

func infoInit(ccId int, ccAddr string, p2pCcRequest *msg.P2pCcRequest) *Info {
	split := strings.Split(ccAddr, ":")
	ccIp := split[0]
	ccPort, _ := strconv.Atoi(split[1])

	ccMapLock.Lock()
	defer ccMapLock.Unlock()
	info := &Info{
		ccId:   ccId,
		ccName: p2pCcRequest.CcName, ccIp: ccIp, ccPort: ccPort,
		csName: p2pCcRequest.CsName, csIp: "", csPort: 0,
		targetIp: p2pCcRequest.TargetIp, targetPort: p2pCcRequest.TargetPort,
		makeHoleTime: p2pCcRequest.MakeHoleTime,
		message:      "not start",
		waitChan:     make(chan uint8, 0),
	}
	ccMap[ccId] = info
	return info
}

func infoUnInit(ccId int) {
	ccMapLock.Lock()
	defer ccMapLock.Unlock()
	delete(ccMap, ccId)
}

func getInfo(ccId int) *Info {
	ccMapLock.Lock()
	defer ccMapLock.Unlock()

	if info, exist := ccMap[ccId]; exist {
		return info
	}
	return nil
}
