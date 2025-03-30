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
		Type:              "hello",
		Heartbeat:         true,
		Apikey:            cfg.APIKey,
		SubscribeDataType: []string{"book_l3"},
		SubscribeFilterSymbol: []string{"COINBASE_SPOT_BTC_USD",
			"ETH-USD"},
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			log.Println("Warning: No data received for 5 seconds")
		default:
			var order api.OrderBookMessage
			if err := decoder.Decode(&order); err != nil {
				log.Println("JSON decode error:", err)
				continue
			}
			logger.Info("Received order", "order", order)
		}
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
