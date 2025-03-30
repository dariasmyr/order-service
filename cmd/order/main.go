package main

import (
	"bufio"
	"context"
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"order-service/config"
	"order-service/internal/orderbook"
	"order-service/models/api"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	cfg := config.MustLoad()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := setupLogger(cfg.Env)
	logger.Info("Starting order matcher", "env", cfg.Env)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

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
		SubscribeFilterSymbol: []string{"COINBASE_SPOT_BTC_USD"},
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

	go func() {
		<-stop
		logger.Info("Gracefully stopping")
		cancel()
		err := ws.Close()
		if err != nil {
			logger.Error("Error closing WebSocket:", "err", err)
		}
	}()

	reader := bufio.NewReader(ws)
	decoder := json.NewDecoder(reader)

	ob := orderbook.NewOrderBook("BTC_USDT")

	totalMessageCount := 0

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				addedCount, removedCount, matchCount := ob.GetOrderStats()
				logger.Info("Order book stats", "match_count", matchCount, "added_count", addedCount, "removed_count", removedCount, "total_message_count", totalMessageCount)
			}
		}
	}()
	processOrderMessage(ctx, decoder, ob, &totalMessageCount)
}

func processOrderMessage(ctx context.Context, decoder *json.Decoder, ob *orderbook.OrderBook, totalMessageCount *int) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			log.Println("Warning: No data received for 5 seconds")
		default:
			var orderMessage api.OrderBookMessage
			err := decoder.Decode(&orderMessage)
			if err != nil {
				log.Println("JSON decode error:", err)
			}

			if orderMessage.Type == "book_l3" {
				*totalMessageCount++
				processOrders(orderMessage, ob)
			}
		}
	}
}

func processOrders(orderMessage api.OrderBookMessage, ob *orderbook.OrderBook) {
	for _, ask := range orderMessage.Asks {
		processOrderUpdate(ask, "ask", ob)
	}
	for _, bid := range orderMessage.Bids {
		processOrderUpdate(bid, "bid", ob)
	}
}

// Логика для обработки каждого ордера в обновлениях
func processOrderUpdate(order api.OrderLevel, orderType string, ob *orderbook.OrderBook) {
	orderObj := orderbook.Order{
		ID:     order.ID,
		Price:  order.Price,
		Volume: order.Size,
		Type:   orderType,
	}

	switch order.UpdateType {
	case "ADD":
		ob.InsertOrder(&orderObj)
	case "DELETE":
		ob.RemoveOrder(&orderObj)
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
