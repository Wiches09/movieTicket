package database

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(addr string, topic string) *kafka.Writer {
	conn, err := kafka.Dial("tcp", addr)
	if err == nil {
		controller, err := conn.Controller()
		if err == nil {
			var controllerConn *kafka.Conn
			controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
			if err == nil {
				err = controllerConn.CreateTopics(kafka.TopicConfig{
					Topic:             topic,
					NumPartitions:     1,
					ReplicationFactor: 1,
				})
				if err != nil {
					log.Printf("Kafka topic creation info: %v", err)
				} else {
					log.Printf("Successfully created Kafka topic: %s", topic)
				}
				controllerConn.Close()
			}
		}
		conn.Close()
	} else {
		log.Printf("Warning: Failed to dial Kafka broker for topic creation: %v", err)
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(addr),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		MaxAttempts:            10,
		BatchTimeout:           10 * time.Millisecond,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}

	log.Printf("Kafka Writer initialized for topic: %s at %s", topic, addr)
	return writer
}
