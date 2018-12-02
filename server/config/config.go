package config

import (
	"os"
)

type Configuration struct {
	ServerListenPort string `json:"serverListenPort"`
	NodeID           string `json:"nodeId"`
	Role             string `json:"role"`
	TendermintIP     string `json:"tendermintIP"`
	TendermintPort   string `json:"tendermintPort"`
}

func LoadConfiguration() Configuration {
	var config Configuration
	config.ServerListenPort = getEnv("SERVER_PORT", "8000")
	config.NodeID = getEnv("NODE_ID", "")
	config.Role = getEnv("ROLE", "")
	config.TendermintIP = getEnv("TENDERMINT_IP", "127.0.0.1")
	config.TendermintPort = getEnv("TENDERMINT_PORT", "45000")
	return config
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
