package core_s3

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config — S3-совместимое Object Storage. Подразумевается внешний
// managed-сервис (например, S3 в Timeweb Cloud) — этот чарт его не
// самохостит, аналогично global.kafkaBrokers (см. helm/.../values.yaml).
type Config struct {
	Endpoint  string `envconfig:"ENDPOINT"   required:"true"`
	Region    string `envconfig:"REGION"     required:"true"`
	Bucket    string `envconfig:"BUCKET"     required:"true"`
	AccessKey string `envconfig:"ACCESS_KEY" required:"true"`
	SecretKey string `envconfig:"SECRET_KEY" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("S3", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get S3 config: %w", err)
		panic(err)
	}

	return config
}
