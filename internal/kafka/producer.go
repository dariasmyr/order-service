package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"log/slog"
	"order-service/internal/lib/logger/sl"
	"order-service/internal/orderbook"
	"os"
)

type Producer struct {
	logger      *slog.Logger
	producer    *kafka.Producer
	messageChan chan *kafka.Message
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
		"acks":                                  "all", // change to "all" if you need write confirmation from all in sync replicas
		"enable.idempotence":                    true,  // true if "acks":"all"
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

	messageChan := make(chan *kafka.Message, 100000)

	return &Producer{logger, producer, messageChan}, nil
}

func (p *Producer) ProduceAsync(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.Error("Context closed, stopping producing messages to kafka")
				return
			case msg := <-p.messageChan:
				if err := p.producer.Produce(msg, nil); err != nil {
					p.logger.Error("Failed to produce message", "error", sl.Err(err))
				}
			}
		}
	}()
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
			p.logger.Info("Topic already exists", "topic", topic)
			return nil
		default:
			return fmt.Errorf("error ensuring topic: %s", result.Error.Code())
		}
	}

	return nil
}

func (p *Producer) BroadcastMessage(event *orderbook.MatchEvent) error {
	if event == nil {
		return fmt.Errorf("nil match event")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &event.SymbolId, Partition: kafka.PartitionAny},
		Value:          payload,
	}

	select {
	case p.messageChan <- msg:
		return nil
	default:
		return fmt.Errorf("producer queue full, dropping message")
	}
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
	if remain := p.producer.Flush(timeoutMs); remain > 0 {
		log.Printf("WARNING: %d messages were not delivered", remain)
	}
}
