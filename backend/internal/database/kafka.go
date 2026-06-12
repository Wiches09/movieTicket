package database

import (
	"log"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(addr string, topic string) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(addr),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("Kafka Writer initialized for topic: %s at %s", topic, addr)
	return writer
}
