package msg

type RegisterType int

const (
	_ RegisterType = iota
	RegisterTypeClient

	P2pCc
	P2pCs
)

type RegisterMessage struct {
	Name         string       `json:"name"`
	RegisterType RegisterType `json:"register_type"`
}

func (t *RegisterMessage) GetMessageType() uint32 {
	return RegisterMessageType
}

type RegisterResultMessage struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

func (t *RegisterResultMessage) GetMessageType() uint32 {
	return RegisterResultMessageType
}
