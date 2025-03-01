package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services ServicesConfig `yaml:"services"`
}

type ServicesConfig struct {
	OTLP   OTLPConfig   `yaml:"otlp"`
	Server ServerConfig `yaml:"server"`
}

type OTLPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

func NewConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func getConfigPath() string {
	// ถ้าอยู่ใน Docker จะใช้ docker.yml, ถ้าไม่จะใช้ local.yml
	if isInDocker() {
		return filepath.Join("configs", "api", "docker.yml")
	}
	return filepath.Join("configs", "api", "local.yml")
}

func isInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}
