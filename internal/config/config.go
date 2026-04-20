package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AmoSubdomain          string `yaml:"amo_subdomain"`
	AmoAPIKey             string `yaml:"amo_api_key"`
	GoogleTableID         string `yaml:"google_table_id"`
	GoogleCredentialsPath string `yaml:"google_credentials_path"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
