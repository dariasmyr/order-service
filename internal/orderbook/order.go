package orderbook

type OrderType int

const (
	BID OrderType = iota
	ASK
)

type Order struct {
	ID         string    `json:"id"`          // 16B
	Price      float64   `json:"price"`       // 8B
	Volume     float64   `json:"size"`        // 8B
	UpdateType string    `json:"update_type"` // 16B
	Type       OrderType `json:"type"`        // 8B
}

func (o *Order) Compare(otherPrice float64) int {
	switch o.Type {
	case ASK: // asks (sellers) are sorted in ascending order
		if o.Price < otherPrice {
			return -1 // receiver order price is less than other
		} else if o.Price > otherPrice {
			return 1 // pl is greater than other
		}
	case BID: // bids (buyers) are sorted in descending order
		if o.Price > otherPrice {
			return -1 // receiver order price is greater than other
		} else if o.Price < otherPrice {
			return 1 // pl is less than other
		}
	}
	return 0 // receiver order price is equal to other
}
