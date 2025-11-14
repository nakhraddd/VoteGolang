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
	writer *kafka.Writer
	ch     chan []byte
}

func NewKafkaLogger(broker, topic string) *KafkaLogger {
	w := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireNone, // 🔥 fully async
	}

	logger := &KafkaLogger{
		writer: w,
		ch:     make(chan []byte, 1000), // buffered channel
	}

	// background async sender
	go func() {
		for msg := range logger.ch {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			_ = logger.writer.WriteMessages(ctx, kafka.Message{
				Value: msg,
			})
			cancel()
		}
	}()

	return logger
}

func (k *KafkaLogger) Log(level, message string) {
	logMsg := LogMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   "vote-service",
	}

	value, _ := json.Marshal(logMsg)

	// Non-blocking send — DROP if overloaded (important!)
	select {
	case k.ch <- value:
	default:
		// buffer full: drop message to avoid blocking
	}
}

func (k *KafkaLogger) Close() error {
	close(k.ch)
	return k.writer.Close()
}
