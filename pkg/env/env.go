package env

import (
	"github.com/caarlos0/env/v11"
)

type config struct {
	ServerPort       int    `env:"SERVER_PORT"`
	DatabaseHost     string `env:"DATABASE_HOST"`
	DatabaseName     string `env:"DATABASE_NAME"`
	DatabaseAppName  string `env:"DATABASE_APP_NAME"`
	DatabaseUser     string `env:"DATABASE_USER"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`
}

func EnvVars() (*config, error) {
	var cfg config
	cfg, err := env.ParseAs[config]()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
