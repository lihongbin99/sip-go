package config

import (
	"encoding/json"
	"os"
	"sip-go/common/logger"
)

type ServerConfigType struct {
	Listen   ListenConfig   `json:"listen"`
	LogLevel string         `json:"log_level"`
	Interval int            `json:"interval"`
	Security SecurityConfig `json:"security"`

	PrivateKey []byte
}

type ListenConfig struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

var (
	ServerConfig ServerConfigType
)

func ServerConfigInit() {
	// 默认的配置文件路径
	if File == "" {
		File = "config/server.json"
	}

	data, err := os.ReadFile(File)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(data, &ServerConfig); err != nil {
		panic(err)
	}

	// 设置日志
	logger.SetLogLevel(ServerConfig.LogLevel)

	// 加载 RAS 密钥
	privateKey, err := os.ReadFile(ServerConfig.Security.PrivateKey)
	if err != nil {
		panic(err)
	}
	ServerConfig.PrivateKey = privateKey
}
