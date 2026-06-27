package listings_transport_kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	core_kafka "listing-service/internal/core/transport/kafka"
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(config core_kafka.ProducerConfig) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BrokersString(),
		"partitioner":       "random",
	})
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{producer: p}, nil
}

func (p *Producer) Publish(ctx context.Context, message core_kafka.Message) error {
	body, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	if err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &message.Topic,
			Partition: kafka.PartitionAny,
		},
		Value: body,
		Key:   []byte(message.Key),
	}, deliveryChan); err != nil {
		return fmt.Errorf("produce message: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case event := <-deliveryChan:
		msg := event.(*kafka.Message)
		if msg.TopicPartition.Error != nil {
			return msg.TopicPartition.Error
		}
	}

	return nil
}

func (p *Producer) Close() {
	p.producer.Flush(core_kafka.FlushTimeOut)
	p.producer.Close()
}
