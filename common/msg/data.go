package msg

type DataMessage struct {
	Data []byte
}

func (t *DataMessage) GetMessageType() uint32 {
	return DataMessageType
}
