package models

import "time"

// OrderBookMessage type represents an order book message
type OrderBookMessage struct {
	Type         string       `json:"type"`          // Message type, always "book"
	SymbolID     string       `json:"symbol_id"`     // Symbol identifier (pair, e.g., "BITSTAMP_SPOT_BTC_USD")
	Sequence     int64        `json:"sequence"`      // Sequence number for the pair
	TimeExchange time.Time    `json:"time_exchange"` // Time of the order book snapshot from the exchange
	TimeCoinAPI  time.Time    `json:"time_coinapi"`  // Time when order book was received from CoinAPI
	IsSnapshot   bool         `json:"is_snapshot"`   // Indicates if this is a snapshot
	Asks         []OrderLevel `json:"asks"`          // Ask levels in the order book
	Bids         []OrderLevel `json:"bids"`          // Bid levels in the order book
}

// OrderLevel represents a single order level in the order book
type OrderLevel struct {
	Price float64 `json:"price"` // Price of the order level
	Size  float64 `json:"size"`  // Size (volume) of the order at this price level
}
