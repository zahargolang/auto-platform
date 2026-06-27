package core_goredis_cache

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config хранит параметры подключения к Redis.
// Все поля читаются из переменных окружения с префиксом "REDIS":
// REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB, REDIS_TIMEOUT.
type Config struct {
	Host     string        `envconfig:"HOST"     required:"true"`
	Port     string        `envconfig:"PORT"     default:"6379"`
	Password string        `envconfig:"PASSWORD" default:""`
	DB       int           `envconfig:"DB"        default:"0"`
	Timeout  time.Duration `envconfig:"TIMEOUT"  default:"2s"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("REDIS", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Redis client config: %w", err)
		panic(err)
	}

	return config
}
