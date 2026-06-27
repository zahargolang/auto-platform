package core_kafka

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type ProducerConfig struct {
	Brokers []string `envconfig:"BROKERS" required:"true"`
}

func NewProducerConfig() (ProducerConfig, error) {
	var config ProducerConfig

	if err := envconfig.Process("KAFKA", &config); err != nil {
		return ProducerConfig{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewProducerConfigMust() ProducerConfig {
	config, err := NewProducerConfig()
	if err != nil {
		err = fmt.Errorf("get Kafka producer config: %w", err)
		panic(err)
	}

	return config
}

func (c ProducerConfig) BrokersString() string {
	return strings.Join(c.Brokers, ",")
}
