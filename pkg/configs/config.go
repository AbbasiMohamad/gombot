package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

const ConfigPath = "./configs/config.yml"

type Config struct {
	Version       string         `yaml:"version"`
	Requesters    []string       `yaml:"requesters"`
	Approvers     []string       `yaml:"approvers"`
	Token         string         `yaml:"token"`
	Microservices []Microservice `yaml:"microservices"`
}

type Microservice struct {
	Name      string `yaml:"name"`
	GitlabUrl string `yaml:"gitlab_url"`
	Branch    string `yaml:"branch"`
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
