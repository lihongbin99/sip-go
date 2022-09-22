package config

import (
	"encoding/json"
	"os"
	"sip-go/common/logger"
)

type ClientConfigType struct {
	ClientName    string              `json:"client_name"`
	ServerConnect ServerConnectConfig `json:"server"`
	LogLevel      string              `json:"log_level"`
	Security      SecurityConfig      `json:"security"`
	Proxy         []ProxyConfig       `json:"proxy"`
	P2p           []P2pConfig         `json:"p2p"`

	PublicKey []byte
}

type ServerConnectConfig struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type ProxyConfig struct {
	ClientName string `json:"client_name"`
	LocalPort  int    `json:"local_port"`
	RemoteIp   string `json:"remote_ip"`
	RemotePort int    `json:"remote_port"`
}

type P2pConfig struct {
	ClientName   string `json:"client_name"`
	LocalPort    int    `json:"local_port"`
	RemoteIp     string `json:"remote_ip"`
	RemotePort   int    `json:"remote_port"`
	MakeHoleTime int    `json:"make_hole_time"`
}

var (
	ClientConfig ClientConfigType
)

func ClientConfigInit() {
	// 默认的配置文件路径
	if File == "" {
		File = "config/client.json"
	}

	data, err := os.ReadFile(File)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(data, &ClientConfig); err != nil {
		panic(err)
	}

	// 设置日志
	logger.SetLogLevel(ClientConfig.LogLevel)

	// 加载 RAS 密钥
	publicKey, err := os.ReadFile(ClientConfig.Security.PublicKey)
	if err != nil {
		panic(err)
	}
	ClientConfig.PublicKey = publicKey
}
