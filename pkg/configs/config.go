package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

const ConfigPath = "./configs/config.yml"

type Config struct {
	Version        string         `yaml:"version"`
	Requesters     []string       `yaml:"requesters"`
	Approvers      []Approvers    `yaml:"approvers"`
	BaleToken      string         `yaml:"baleToken"`
	GitlabToken    string         `yaml:"gitlabToken"`
	Applications   []Application  `yaml:"applications"`
	DatabaseConfig DatabaseConfig `yaml:"database"`
}

type Approvers struct {
	FullName string `yaml:"fullName"`
	Username string `yaml:"username"`
}

type Application struct {
	Name          string `yaml:"name"`
	PersianName   string `yaml:"persianName"`
	Branch        string `yaml:"branch"`
	NeedToApprove bool   `yaml:"needToApprove"`
}

type DatabaseConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Database      string `yaml:"dbname"`
	NeedMigration bool   `yaml:"needMigration"`
}

func LoadConfig(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("failed to load config. check the config path.") //TODO: investigate about panic!
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic("failed to load config. can not unmarshal config file.")
	}
	return &config
}
