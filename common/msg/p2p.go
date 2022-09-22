package msg

type P2pCcRequest struct {
	CcName       string `json:"cc_name"`
	CsName       string `json:"cs_name"`
	TargetIp     string `json:"target_ip"`
	TargetPort   int    `json:"target_port"`
	MakeHoleTime int    `json:"make_hole_time"`
	Key          []byte `json:"key"`
	Iv           []byte `json:"iv"`
}

func (t *P2pCcRequest) GetMessageType() uint32 {
	return P2pCcRequestType
}

type P2pCcResponse struct {
	YourIp     string `json:"your_ip"`
	YourPort   int    `json:"your_port"`
	TargetIp   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	Message    string `json:"message"`
}

func (t *P2pCcResponse) GetMessageType() uint32 {
	return P2pCcResponseType
}

type P2pCsRequest struct {
	CcId         int    `json:"cc_id"`
	CcName       string `json:"cc_name"`
	CcIp         string `json:"cc_ip"`
	CcPort       int    `json:"cc_port"`
	TargetIp     string `json:"target_ip"`
	TargetPort   int    `json:"target_port"`
	MakeHoleTime int    `json:"make_hole_time"`
	Key          []byte `json:"key"`
	Iv           []byte `json:"iv"`
}

func (t *P2pCsRequest) GetMessageType() uint32 {
	return P2pCsRequestType
}

type P2pCsResponse struct {
	CcId    int    `json:"cc_id"`
	Message string `json:"message"`
}

func (t *P2pCsResponse) GetMessageType() uint32 {
	return P2pCsResponseType
}

type P2pCsNewConnectRequest struct {
	CcId int `json:"cc_id"`
}

func (t *P2pCsNewConnectRequest) GetMessageType() uint32 {
	return P2pCsNewConnectRequestType
}

type P2pCsNewConnectResponse struct {
	YourIp   string `json:"your_ip"`
	YourPort int    `json:"your_port"`
	Message  string `json:"message"`
}

func (t *P2pCsNewConnectResponse) GetMessageType() uint32 {
	return P2pCsNewConnectResponseType
}
