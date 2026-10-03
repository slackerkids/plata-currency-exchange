package config

import (
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	configPath = "./configs/.env.template" // set to .env in production.
)

type Configuration struct {
	HTTPServerAddress string `env:"HTTP_SERVER_ADDRESS"`
}

func New() (*Configuration, error) {
	if configPath == "" {
		return nil, errors.New("config path is empty")
	}

	conf := &Configuration{}

	if err := godotenv.Load(configPath); err != nil {
		return nil, fmt.Errorf("failed to load data from env file: %s, %w", configPath, err)
	}

	if err := env.Parse(conf); err != nil {
		return nil, fmt.Errorf("failed to parse env to struct: %w", err)
	}

	return conf, nil
}
