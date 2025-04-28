package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"os"
)

type Producer struct {
	producer *kafka.Producer
}

func NewProducer() (*Producer, error) {

	caLocation := "pem/ca-root.pem"

	if _, err := os.Stat(caLocation); os.IsNotExist(err) {
		return nil, fmt.Errorf("CA file not found at path: %s", caLocation)
	} else if err != nil {
		return nil, fmt.Errorf("error checking CA file: %v", err)
	} else {
		fmt.Printf("CA file found at path: %s\n", caLocation)
	}

	config := &kafka.ConfigMap{
		"bootstrap.servers":                     "localhost:29092",
		"sasl.username":                         "sa",
		"sasl.password":                         "000000",
		"acks":                                  "all",
		"enable.idempotence":                    false,
		"sasl.mechanism":                        "PLAIN",
		"security.protocol":                     "SASL_SSL",
		"ssl.endpoint.identification.algorithm": "none",
		"ssl.key.password":                      "supersecret",
		"ssl.ca.location":                       caLocation,
		"client.id":                             "go-producer",
		//"debug":                                 "all",
	}

	producer, err := kafka.NewProducer(config)

	if err != nil {
		return nil, err
	}

	return &Producer{producer}, nil
}

func (s *Producer) Close() {
	s.producer.Close()
}

// SendMessage отправляет сообщение в Kafka
func (s *Producer) SendMessage(topic string, key string, value string) error {
	// Создаем сообщение Kafka
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}

	// Отправляем сообщение в Kafka
	err := s.producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %s", err)
	}

	return nil
}

// StartListening для обработки событий
func (s *Producer) StartListening() {
	go func() {
		for e := range s.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Produced message to topic %s: key = %-10s value = %s\n", *ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()
}
