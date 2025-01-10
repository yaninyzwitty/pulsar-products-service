package pkg

import (
	"io"
	"log/slog"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   Server `yaml:"server"`
	Pulsar   Pulsar `yaml:"pulsar"`
	Database DB     `yaml:"database"`
}

type Server struct {
	Port int `yaml:"port"`
}

type Pulsar struct {
	URI       string `yaml:"uri"`
	TopicName string `yaml:"topic_name"`
}

type DB struct {
	User     string `yaml:"username"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

func (c *Config) LoadConfig(file io.Reader) error {
	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("failed to read file", "error", err)
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		return err
	}
	return nil

}
