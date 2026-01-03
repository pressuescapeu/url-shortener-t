package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string   `yaml:"env" env:"ENV" env-default:"local"`
	Database   Database `yaml:"database"`
	HTTPServer `yaml:"http_server"`
}

type Database struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	DBName   string `yaml:"dbname" env:"DB_NAME" env-default:"urlshortener"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE" env-default:"disable"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"localhost:8082"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	User        string        `yaml:"user" env:"HTTP_USER" env-required:"true"`
	Password    string        `yaml:"password" env:"HTTP_PASSWORD" env-required:"true"`
}

// MustLoad reads config from YAML file if CONFIG_PATH is set,
// otherwise reads from environment variables
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	var configuration Config

	if configPath != "" {
		// YAML mode - read from file
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file does not exist: %s", configPath)
		}

		if err := cleanenv.ReadConfig(configPath, &configuration); err != nil {
			log.Fatalf("cannot read configuration: %s", err)
		}
	} else {
		// Environment variable mode - read from env
		if err := cleanenv.ReadEnv(&configuration); err != nil {
			log.Fatalf("cannot read configuration from environment: %s", err)
		}
	}

	return &configuration
}
