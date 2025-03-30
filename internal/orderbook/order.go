package orderbook

type Order struct {
	Type       string
	ID         string
	Price      float64
	Volume     float64
	UpdateType string
}

func (o *Order) Compare(otherPrice float64) int {
	switch o.Type {
	case "ask": // asks (sellers) are sorted in ascending order
		if o.Price < otherPrice {
			return -1 // receiver order price is less than other
		} else if o.Price > otherPrice {
			return 1 // pl is greater than other
		}
	case "bid": // bids (buyers) are sorted in descending order
		if o.Price > otherPrice {
			return -1 // receiver order price is greater than other
		} else if o.Price < otherPrice {
			return 1 // pl is less than other
		}
	}
	return 0 // receiver order price is equal to other
}
