package config

import (
	"flag"
)

var (
	SUCCESS = "success"

	File string
)

type SecurityConfig struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func init() {
	flag.StringVar(&File, "c", File, "config file")
}
