package indexer

import (
	"encoding/json"
	"time"

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
	Host    blockstack.ServerConfig   `json:"host"`
	Hosts   []blockstack.ServerConfig `json:"hosts"`
	Retries int                       `json:"retries"`
	Timeout time.Duration             `json:"timeout"`
}

// DBConfig represents the backing database (MongoDB)
type DBConfig struct {
	Connection string `json:"connection"`
	Database   string `json:"database"`
	Driver     string `json:"driver"`
}

// IDXConfig represents indexer specific configuration
type IDXConfig struct {
	StatsPort   int    `json:"statsPort"`
	Concurrency int    `json:"concurrency"`
	NameFile    string `json:"namefile"`
}
