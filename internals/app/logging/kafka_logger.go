package logging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type LogMessage struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
}
type KafkaLogger struct {
	writer  *kafka.Writer
	Service string
}

func NewKafkaLogger(broker, topic, service string) *KafkaLogger {
	return &KafkaLogger{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		},
		Service: service,
	}
}

func (k *KafkaLogger) Log(level string, message string) error {
	logMsg := LogMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   k.Service,
	}

	value, err := json.Marshal(logMsg)
	if err != nil {
		return err
	}

	key := []byte(logMsg.Timestamp)

	return k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)
}

func (k *KafkaLogger) Close() error {
	return k.writer.Close()
}
