package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Port       string `yaml:"SERVER_PORT" env:"SERVER_PORT" env-default:"18000"`
	Host       string `yaml:"SERVER_HOST" env:"SERVER_HOST" env-default:"localhost"`
	Socks5Port string `yaml:"SOCKS5_PORT" env:"SOCKS5_PORT" env-default:"10800"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("./config/config.yaml", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
