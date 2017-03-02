package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ListenAddr []string `json:"listen"`
	Bootstrap  []RHost  `json:"bootstrap"`
	Secret     string   `json:"secret"`
}

type RHost struct {
	PeerID string `json:"peer"`
	Addr   string `json:"address"`
}

func ReadCfg(path string) (*Config, error) {
	var cfg Config
	fi, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fi, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
