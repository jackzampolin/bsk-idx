package cmd

import (
	"encoding/json"

	"github.com/blockstack/blockstack.go/blockstack"
)

// Config represents the configuration struct
type Config struct {
	DB  DBConfig  `json:"db"`
	BSK BSKConfig `json:"bsk"`
	IDX IDXConfig `json:"idx"`
}

// JSON renders json
func (c *Config) JSON() string {
	byt, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(byt)
}

// BSKConfig represents the blockstack core node config
type BSKConfig struct {
	Host  blockstack.ServerConfig   `json:"host"`
	Hosts []blockstack.ServerConfig `json:"hosts"`
}

// DBConfig represents the backing database (MongoDB)
type DBConfig struct {
	Connection   string `json:"connection"`
	Database     string `json:"database"`
	BatchSize    int    `json:"batchSize"`
	BatchWorkers int    `json:"batchSize"`
}

// IDXConfig represents indexer specific configuration
type IDXConfig struct {
	Port        int `json:"port"`
	Concurrency int `json:"concurrency"`
}
