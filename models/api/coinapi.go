package api

import "time"

// OrderBookMessage type represents an order book message
type OrderBookMessage struct {
	TimeExchange time.Time    `json:"time_exchange"` // Time of the order book snapshot from the exchange // 24B
	TimeCoinAPI  time.Time    `json:"time_coinapi"`  // Time when order book was received from CoinAPI // 24B (time:  pointer (8B) + wall (8B) + ext (8B))
	Asks         []OrderLevel `json:"asks"`          // Ask levels in the order book // 24B (slice: pointer (8B) + len (8B) + cap (8B))
	Bids         []OrderLevel `json:"bids"`          // Bid levels in the order book // 24B (slice: pointer (8B) + len (8B) + cap (8B))
	Sequence     int64        `json:"sequence"`      // Sequence number for the pair // 8B
	SymbolID     string       `json:"symbol_id"`     // Symbol identifier (pair, e.g., "BITSTAMP_SPOT_BTC_USD") // 16B
	Type         string       `json:"type"`          // Message type, always "book_l3"  // 16B
	IsSnapshot   bool         `json:"is_snapshot"`   // Indicates if this is a snapshot // 1B
}

// OrderLevel represents a single order level in the order book
type OrderLevel struct {
	ID         string  `json:"id"`          // uuid4, 16B
	UpdateType string  `json:"update_type"` // 16B
	Price      float64 `json:"price"`       // Price of the order level // 8B
	Size       float64 `json:"size"`        // Size (volume) of the order at this price level // 8B
}

// UpdateType is the type of update
/*
ADD	Add order to the price queue.
UPDATE	Update the order size in the price queue.
SUBTRACT	Subtract the order size in the price queue.
DELETE	Delete the order from the price queue.
MATCH	Subtract the order size in the price queue (same as SUBTRACT but caused by the match).
*/
