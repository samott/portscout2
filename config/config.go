package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	PortsDir string `yaml:"portsDir"`
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
