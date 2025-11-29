package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env   string       `yaml:"env" env-default:"local"`
	Kafka KafkaStorage `yaml:"kafka"`
}

type KafkaStorage struct {
	Brokers   []string `yaml:"brokers"`
	TopicUser string   `yaml:"topic_user"`
	TopicLink string   `yaml:"topic_link"`
	GroupID   string   `yaml:"group_id"`
}

func MustLoad() *Config {
	configPath := "config/config.yaml"
	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}
