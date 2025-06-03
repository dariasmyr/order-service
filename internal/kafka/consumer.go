package kafka

import (
	"context"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"log/slog"
	"time"
)

type Consumer struct {
	logger   *slog.Logger
	consumer *kafka.Consumer
	ctx      context.Context
}

func NewConsumer(logger *slog.Logger, ctx context.Context) *Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "kafka:9092",
		"group.id":           "group",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})
	if err != nil {
		logger.Error("Failed to create consumer: %v", err)
		return nil
	}

	return &Consumer{
		logger:   logger,
		consumer: c,
		ctx:      ctx,
	}
}

func (c *Consumer) Subscribe(topic string, handler func(msg *kafka.Message) error) (err error) {
	defer func() {
		closeErr := c.consumer.Close()
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = errors.Join(err, closeErr)
			}
		}
	}()

	err = c.consumer.Subscribe(topic, nil)
	if err != nil {
		c.logger.Error("Failed to subscribe to %s: %v", topic, err)
	}

consumerLoop:
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Consumer stopped")
			break consumerLoop
		default:
			msg, readErr := c.consumer.ReadMessage(-1)
			if readErr != nil {
				c.logger.Error("Failed to read message from %s: %v", topic, err)
				continue
			}

			var success bool
			for i := 0; i < len(msg.Value); i++ {
				handlerErr := handler(msg)
				if handlerErr == nil {
					success = true
					break
				}
				c.logger.Error("Failed to handle message %d: %v", i, handlerErr)
				c.logger.Debug("Retry handling message in %d", "time", time.Second*time.Duration(1<<i))
				time.Sleep(time.Second * time.Duration(1<<i))
			}

			if !success {
				c.logger.Error("Message handling failed after retres: %s", string(msg.Value))
				broadcastToDLQ(msg, readErr)
			}

			_, commitErr := c.consumer.CommitMessage(msg)
			if commitErr != nil {
				log.Printf("Commit error: %v", err)
			}
		}
	}

	return err
}
