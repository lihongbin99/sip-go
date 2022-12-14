package p2p

import (
	"fmt"
	"net"
	"sip-go/common/io"
	"sip-go/common/msg"
	"time"
)

func p2p(local2ServerAddr, remoteAddr net.Addr) (net.Conn, error) {
	dialer := net.Dialer{Timeout: 3 * time.Second, LocalAddr: local2ServerAddr}
	remoteConn, err := dialer.Dial("tcp", remoteAddr.String())
	if err != nil {
		return nil, fmt.Errorf("p2p connect error: %v", err)
	}
	return remoteConn, nil
}

func p2pTransferData(localConn *net.TCPConn, p2pTCP *io.TCP) (string, string) {
	stopChan := make(chan uint8, 0)

	var uploadByte uint64 = 0
	var downloadByte uint64 = 0

	go func() {
		defer func() {
			_ = localConn.Close()
			_ = p2pTCP.Close()
			stopChan <- 1
		}()
		buf := make([]byte, 64*1024)
		for {
			if readLen, err := localConn.Read(buf); err != nil {
				return
			} else {
				if err = p2pTCP.WriteMessage(&msg.DataMessage{Data: buf[:readLen]}); err != nil {
					return
				}
				uploadByte += uint64(readLen)
			}
		}
	}()

	go func() {
		defer func() {
			_ = p2pTCP.Close()
			_ = localConn.Close()
			stopChan <- 1
		}()
		for {
			message := p2pTCP.ReadMessage(time.Time{})
			if message.Err != nil {
				return
			}
			switch m := message.Message.(type) {
			case *msg.PingMessage:
				_ = p2pTCP.WriteMessage(&msg.PongMessage{Date: time.Now()})
			case *msg.PongMessage:
				// TODO 检测超时
			case *msg.DataMessage:
				if _, err := localConn.Write(m.Data); err != nil {
					return
				}
				downloadByte += uint64(len(m.Data))
			default:
				log.Error("p2p transfer type error:", message.Message.GetMessageType())
				return
			}
		}
	}()

	_ = <-stopChan
	_ = <-stopChan

	return byteToUnit(uploadByte), byteToUnit(downloadByte)
}

var (
	unit = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
)

func byteToUnit(n uint64) string {
	num := float64(n)
	numUnit := unit[0]

	for i := 1; num >= 1024; i, num = i+1, num/1024 {
		numUnit = unit[i]
	}

	return fmt.Sprintf("%f%s", num, numUnit)
}
