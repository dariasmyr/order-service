package main

import (
	"bufio"
	"context"
	"encoding/json"
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
		SubscribeFilterSymbol: []string{"COINBASE_SPOT_BTC_USD", "COINBASE_SPOT_ETH_USD"},
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

	messageChannel := make(chan api.OrderBookMessage)

	btcOrderBook := orderbook.NewOrderBook("COINBASE_SPOT_BTC_USD")
	ethOrderBook := orderbook.NewOrderBook("COINBASE_SPOT_ETH_USD")

	totalMessageCount := 0

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				addedCount, removedCount, matchCount := btcOrderBook.GetOrderStats()
				logger.Info("BTC OrderBook Stats", "total_messages", totalMessageCount, "match", matchCount, "added", addedCount, "removed", removedCount)
				if btcOrderBook.BidTree.BestPriceOrder != nil {
					logger.Info("BTC Best Bid", "price", btcOrderBook.BidTree.BestPriceOrder.Price, "volume", btcOrderBook.BidTree.BestPriceOrder.Volume)
				}
				if btcOrderBook.AskTree.BestPriceOrder != nil {
					logger.Info("BTC Best Ask", "price", btcOrderBook.AskTree.BestPriceOrder.Price, "volume", btcOrderBook.AskTree.BestPriceOrder.Volume)
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
				//fmt.Printf("Received book l3 order message: %v\n", orderMessage)
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
