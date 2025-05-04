package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"order-service/config"
	kafkaProducer "order-service/internal/kafka"
	"order-service/internal/lib/logger/sl"
	"order-service/internal/orderbook"
	"order-service/models/api"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/websocket"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type SubscriptionMessage struct {
	Type                  string   `json:"type"`
	Heartbeat             bool     `json:"heartbeat"`
	Apikey                string   `json:"apikey"`
	SubscribeDataType     []string `json:"subscribe_data_type"`
	SubscribeFilterSymbol []string `json:"subscribe_filter_symbol_id"`
}

func main() {
	var wg sync.WaitGroup

	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)
	logger.Info("Starting order matcher", "env", cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		logger.Info("shutting down")
		cancel()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	kafkaService, err := kafkaProducer.NewProducer(
		logger,
		cfg.Kafka.BoostrapServerURL,
		cfg.Kafka.SASLUsername,
		cfg.Kafka.SASLPassword,
		cfg.Kafka.CALocation,
		cfg.Kafka.SSLPassword,
	)
	if err != nil {
		logger.Error("Error creating Kafka producer", "error", sl.Err(err))
		// call runtime.Goexit() instead of log.Fatal (os.Exit()) so as to call deferred funcs
		runtime.Goexit()
	}
	defer func() {
		logger.Info("Start closing Kafka producer")
		kafkaService.Close()
		logger.Info("Closed Kafka producer")
	}()

	kafkaService.StartListeningWithHandler(func(message *kafka.Message) {
		if message.TopicPartition.Error != nil {
			logger.Error("Failed to deliver message: ", "topic", message.TopicPartition)
		}
	})

	kafkaService.ProduceAsync(ctx)

	headers := http.Header{}
	originURL, err := url.Parse("http://localhost/")
	if err != nil {
		log.Fatal("Invalid origin URL:", err)
	}
	locationURL, err := url.Parse("wss://ws.coinapi.io/v1/")
	if err != nil {
		log.Fatal("Invalid location URL:", err)
	}

	ws, err := websocket.DialConfig(&websocket.Config{
		Location: locationURL,
		Origin:   originURL,
		Version:  13,
		Header:   headers,
	})
	if err != nil {
		log.Fatal("Failed to connect to WebSocket:", err)
	} else {
		logger.Info("Connected to WebSocket")
	}

	subscription := SubscriptionMessage{
		Type:                  "hello",
		Heartbeat:             true,
		Apikey:                cfg.APIKey,
		SubscribeDataType:     []string{"book_l3"},
		SubscribeFilterSymbol: []string{"COINBASE_SPOT_BTC_USDT", "COINBASE_SPOT_ETH_USDT"},
	}

	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		log.Fatal("Failed to encode subscription message:", err)
	}

	_, err = ws.Write(subscriptionJSON)
	if err != nil {
		log.Fatal("Failed to send subscription:", err)
	} else {
		logger.Info("Subscription sent")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-stop
		close(stop)
		logger.Info("Gracefully stopping")
		// Call cancel() to trigger ctx.Done() and initiate listeners to exit the processes
		cancel()
		err := ws.Close()
		if err != nil {
			logger.Error("Error closing WebSocket:", "err", err)
		}
		logger.Info("WebSocket stopped")
		kafkaService.Flush(cfg.Kafka.FlushTimeoutMs)
		logger.Info("Successfully flushed Kafka producer")
	}()

	reader := bufio.NewReader(ws)
	decoder := json.NewDecoder(reader)

	messageChannel := make(chan api.OrderBookMessage)

	btcOrderBook := orderbook.NewOrderBook("COINBASE_SPOT_BTC_USDT", logger, kafkaService)
	err = kafkaService.CreateTopic(ctx, "COINBASE_SPOT_BTC_USDT", 3, 3)
	if err != nil {
		log.Fatalf("Error creating topic: %s", err)
	}
	ethOrderBook := orderbook.NewOrderBook("COINBASE_SPOT_ETH_USDT", logger, kafkaService)
	err = kafkaService.CreateTopic(ctx, "COINBASE_SPOT_ETH_USDT", 3, 3)
	if err != nil {
		log.Fatalf("Error creating topic: %s", err)
	}

	totalMessageCount := 0

	// Stats cron, can be removed later
	go func() {
		timer := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				logger.Info("Context cancelled, exiting stats loop...")
				return
			case <-timer.C:
				addedCount, removedCount, matchCount := btcOrderBook.GetOrderStats()
				logger.Info("BTC OrderBook Stats", "total_messages", totalMessageCount, "match", matchCount, "added", addedCount, "removed", removedCount)
				if btcOrderBook.BidTree.BestPriceOrder != nil {
					logger.Info("BTC Best Bid", "price", btcOrderBook.BidTree.BestPriceOrder.Price, "volume", btcOrderBook.BidTree.BestPriceOrder.Volume)
					if err != nil {
						logger.Error("Failed to send message:", err)
					}
				}
				if btcOrderBook.AskTree.BestPriceOrder != nil {
					logger.Info("BTC Best Ask", "price", btcOrderBook.AskTree.BestPriceOrder.Price, "volume", btcOrderBook.AskTree.BestPriceOrder.Volume)
					if err != nil {
						logger.Error("Failed to send message:", err)
					}
				}

				addedCount, removedCount, matchCount = ethOrderBook.GetOrderStats()
				logger.Info("ETH OrderBook Stats", "total_messages", totalMessageCount, "match", matchCount, "added", addedCount, "removed", removedCount)
				if ethOrderBook.BidTree.BestPriceOrder != nil {
					logger.Info("ETH Best Bid", "price", ethOrderBook.BidTree.BestPriceOrder.Price, "volume", ethOrderBook.BidTree.BestPriceOrder.Volume)
				}
				if ethOrderBook.AskTree.BestPriceOrder != nil {
					logger.Info("ETH Best Ask", "price", ethOrderBook.AskTree.BestPriceOrder.Price, "volume", ethOrderBook.AskTree.BestPriceOrder.Volume)
				}
			}
		}
	}()

	go processOrderBook(ctx, messageChannel, btcOrderBook, &totalMessageCount)

	go processOrderBook(ctx, messageChannel, ethOrderBook, &totalMessageCount)

	for {
		timer := time.NewTicker(5 * time.Second)
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, exiting socket connection...")
			close(messageChannel)
			wg.Wait()
			return
		case <-timer.C:
			logger.Warn("Warning: No data received for 5 seconds")
			close(messageChannel)
			wg.Wait()
			return
		default:
			var orderMessage api.OrderBookMessage
			err = decoder.Decode(&orderMessage)
			if err != nil {
				logger.Error("JSON decode error:", sl.Err(err))
			}

			if orderMessage.Type == "book_l3" {
				//fmt.Printf("Received book l3 order message: %+v\n", orderMessage)
				messageChannel <- orderMessage
			}
		}
	}
}

func processOrderBook(ctx context.Context, messageChannel chan api.OrderBookMessage, ob *orderbook.OrderBook, totalMessageCount *int) {
	for {
		select {
		case <-ctx.Done():
			return
		case orderMessage := <-messageChannel:
			if orderMessage.SymbolID != ob.SymbolID {
				continue
			}

			processOrders(orderMessage, ob, totalMessageCount)
		}
	}
}

func processOrders(orderMessage api.OrderBookMessage, ob *orderbook.OrderBook, totalMessageCount *int) {
	for _, ask := range orderMessage.Asks {
		processOrder(ask, orderbook.ASK, ob, totalMessageCount)
	}
	for _, bid := range orderMessage.Bids {
		processOrder(bid, orderbook.BID, ob, totalMessageCount)
	}
}

func processOrder(order api.OrderLevel, orderType orderbook.OrderType, ob *orderbook.OrderBook, totalMessageCount *int) {
	if order.ID == "" {
		return
	}
	orderObj := orderbook.Order{
		ID:     order.ID,
		Price:  order.Price,
		Volume: order.Size,
		Type:   orderType,
	}

	switch order.UpdateType {
	case "ADD":
		*totalMessageCount++
		ob.InsertOrder(&orderObj)
	case "MATCH":
		*totalMessageCount++
		ob.UpdateOrderVolume(&orderObj)
	case "UPDATE":
		*totalMessageCount++
		ob.UpdateOrderVolume(&orderObj)
	case "SUBTRACT":
		*totalMessageCount++
		ob.UpdateOrderVolume(&orderObj)
	case "DELETE":
		*totalMessageCount++
		ob.RemoveOrder(&orderObj)
	case "":
		*totalMessageCount++
		ob.InsertOrder(&orderObj)
	}
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}
