package msg

import "time"

type PingMessage struct {
	Date time.Time `json:"date"`
}

func (t *PingMessage) GetMessageType() uint32 {
	return PingMessageType
}

type PongMessage struct {
	Date time.Time `json:"date"`
}

func (t *PongMessage) GetMessageType() uint32 {
	return PongMessageType
}
