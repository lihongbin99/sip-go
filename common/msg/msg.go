package msg

import "fmt"

type Message interface {
	GetMessageType() uint32
}

const (
	_ uint32 = iota

	DataMessageType

	PingMessageType
	PongMessageType

	RegisterMessageType
	RegisterResultMessageType

	P2pCcRequestType
	P2pCcResponseType

	P2pCsRequestType
	P2pCsResponseType

	P2pCsNewConnectRequestType
	P2pCsNewConnectResponseType
)

func NewMessage(messageType uint32) (message Message, err error) {
	switch messageType {
	case DataMessageType:
		message = &DataMessage{}

	case PingMessageType:
		message = &PingMessage{}
	case PongMessageType:
		message = &PongMessage{}

	case RegisterMessageType:
		message = &RegisterMessage{}
	case RegisterResultMessageType:
		message = &RegisterResultMessage{}

	case P2pCcRequestType:
		message = &P2pCcRequest{}
	case P2pCcResponseType:
		message = &P2pCcResponse{}

	case P2pCsRequestType:
		message = &P2pCsRequest{}
	case P2pCsResponseType:
		message = &P2pCsResponse{}

	case P2pCsNewConnectRequestType:
		message = &P2pCsNewConnectRequest{}
	case P2pCsNewConnectResponseType:
		message = &P2pCsNewConnectResponse{}

	default:
		err = fmt.Errorf("no find message type: %d", messageType)
	}
	return
}
