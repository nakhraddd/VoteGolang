
package logging

import (
    "context"
    "time"

    "github.com/segmentio/kafka-go"
)

type KafkaLogger struct {
    writer *kafka.Writer
}

func NewKafkaLogger(broker, topic string) *KafkaLogger {
    return &KafkaLogger{
        writer: &kafka.Writer{
            Addr:         kafka.TCP(broker),
            Topic:        topic,
            Balancer:     &kafka.LeastBytes{},
            RequiredAcks: kafka.RequireAll,
        },
    }
}

func (k *KafkaLogger) Log(message string) error {
    return k.writer.WriteMessages(context.Background(),
        kafka.Message{
            Key:   []byte(time.Now().Format(time.RFC3339)),
            Value: []byte(message),
        },
    )
}

func (k *KafkaLogger) Close() error {
    return k.writer.Close()
}