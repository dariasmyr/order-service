package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log/slog"
	"os"
)

type Producer struct {
	logger   *slog.Logger
	producer *kafka.Producer
}

func NewProducer(
	logger *slog.Logger,
	boostrapServers string,
	saslUsername string,
	saslPassword string,
	caLocation string,
	sslPassword string,

) (*Producer, error) {
	if _, err := os.Stat(caLocation); os.IsNotExist(err) {
		return nil, fmt.Errorf("CA file not found at path: %s", caLocation)
	} else if err != nil {
		return nil, fmt.Errorf("error checking CA file: %v", err)
	} else {
		fmt.Printf("CA file found at path: %s\n", caLocation)
	}

	config := &kafka.ConfigMap{
		"bootstrap.servers":                     boostrapServers,
		"sasl.username":                         saslUsername,
		"sasl.password":                         saslPassword,
		"acks":                                  "all",
		"enable.idempotence":                    false,
		"sasl.mechanism":                        "PLAIN",
		"security.protocol":                     "SASL_SSL",
		"ssl.endpoint.identification.algorithm": "none",
		"ssl.key.password":                      sslPassword,
		"ssl.ca.location":                       caLocation,
		"client.id":                             "order-producer",
		//"debug":                                 "all", // For debug messages
	}

	producer, err := kafka.NewProducer(config)

	if err != nil {
		return nil, err
	}

	return &Producer{logger, producer}, nil
}

func (p *Producer) Close() {
	p.producer.Close()
}

func (p *Producer) CreateTopic(ctx context.Context, topic string, numPartitions int, replicationFactor int) error {
	admin, err := kafka.NewAdminClientFromProducer(p.producer)
	if err != nil {
		return fmt.Errorf("error creating admin client from producer: %v", err)
	}
	defer admin.Close()

	results, topicErr := admin.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		}},
		nil,
	)

	if topicErr != nil {
		return fmt.Errorf("error creating topic: %v", topicErr)
	}

	for _, result := range results {
		switch result.Error.Code() {
		case kafka.ErrNoError:
			p.logger.Info("Topic created successfully", "topic", topic)
			return nil
		case kafka.ErrTopicAlreadyExists:
			p.logger.Info("Topic %s already exists", topic)
			return nil
		default:
			return fmt.Errorf("error ensuring topic: %s", result.Error.Code())
		}
	}

	return nil
}

func (p *Producer) SendMessage(topic string, key string, value string) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}

	err := p.producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %s", err)
	}

	return nil
}

type DeliveryHandlerFunc func(delivery *kafka.Message)

func (p *Producer) StartListeningWithHandler(handler DeliveryHandlerFunc) {
	go func() {
		for e := range p.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				handler(ev)
			}
		}
	}()
}

func (p *Producer) Flush(timeoutMs int) {
	p.producer.Flush(timeoutMs)
}
