package configs

import (
	"gopkg.in/yaml.v3"
	"os"
)

const ConfigPath = "./configs/configs.yml"

type Config struct {
	Version    string   `yaml:"version"`
	Requesters []string `yaml:"requesters"`
	Approvers  []string `yaml:"approvers"`
	Token      string   `yaml:"token"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
