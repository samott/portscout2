package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Db struct {
		Url string `yaml:"url"`
		PageSize uint `yaml:"pageSize"`
	} `yaml:"db"`

	Tree struct {
		PortsDir    string `yaml:"portsDir"`
		MakeCmd     string `yaml:"makeCmd"`
		MakeThreads int    `yaml:"makeThreads"`
	} `yaml:"tree"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)

	var config Config

	if err != nil {
		return nil, err
	}

	yaml.Unmarshal(data, &config)

	return &config, nil
}
