package io

import (
	"sip-go/common/logger"
	"sip-go/common/msg"
	"sync"
)

type Message struct {
	Message msg.Message
	Err     error
}

var (
	log = logger.NewLog("IO")

	clients     = make(map[string]*TCP)
	clientsLock = sync.Mutex{}
)

func RegisterClient(name string, tcp *TCP) bool {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	if _, exist := clients[name]; !exist {
		clients[name] = tcp
		return true
	}
	return false
}

func RemoveClient(name string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	delete(clients, name)
}

func GetClient(name string) *TCP {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	if tcp, exist := clients[name]; exist {
		return tcp
	}
	return nil
}
