package core_kafka

type Message struct {
	Topic   string
	Key     string
	Payload any
}

func NewMessage(topic, key string, payload any) Message {
	return Message{
		Topic:   topic,
		Key:     key,
		Payload: payload,
	}
}
