package orderbook

import (
	"math"
)

type OrderBook struct {
	SymbolID           string  // 16B
	BidTree            *RBTree // 8B
	AskTree            *RBTree // 8B
	MatchOrdersCount   int     // 8B optional field, can be removed for optimisation
	AddedOrdersCount   int     // 8B optional field, can be removed for optimisation
	RemovedOrdersCount int     // 8B optional field, can be removed for optimisation
}

func NewOrderBook(symbolId string) *OrderBook {
	leafNode := &Node{Color: BLACK}
	return &OrderBook{
		SymbolID: symbolId,
		BidTree: &RBTree{
			Root: leafNode,
			leaf: leafNode,
		},
		AskTree: &RBTree{
			Root: leafNode,
			leaf: leafNode,
		},
	}
}

func (ob *OrderBook) RemoveOrder(order *Order) {
	if order.Type == BID {
		ob.BidTree.Remove(order)
	} else {
		ob.AskTree.Remove(order)
	}
	ob.RemovedOrdersCount++
}

func (ob *OrderBook) UpdateOrderVolume(order *Order) {
	if order.Volume <= 0 || order.Price <= 0 {
		ob.RemoveOrder(order)
		return
	}
	if order.Type == BID {
		ob.BidTree.UpdateOrderVolume(order)
	} else {
		ob.AskTree.UpdateOrderVolume(order)
	}
}

func (ob *OrderBook) InsertOrder(order *Order) {
	if order.Volume <= 0 || order.Price <= 0 {
		return
	}

	var oppositeTree *RBTree
	var sameTree *RBTree

	if order.Type == ASK {
		oppositeTree = ob.BidTree // Buyers
		sameTree = ob.AskTree     // Sellers
	} else {
		oppositeTree = ob.AskTree // Sellers
		sameTree = ob.BidTree     // Buyers
	}

	// Check if opposite tree exists and has best orders to match
	if oppositeTree.BestPriceOrder != nil &&
		((order.Type == BID && order.Price >= oppositeTree.BestPriceOrder.Price) ||
			(order.Type == ASK && order.Price <= oppositeTree.BestPriceOrder.Price)) {
		// Order can aggressive, match it with opposite best orders
		ob.MatchOrder(order, sameTree, oppositeTree)
	} else {
		// Order is passive, insert it into the same tree
		sameTree.Insert(order)
		// Check if this order has the best price
		if sameTree.BestPriceOrder == nil {
			sameTree.BestPriceOrder = order
		} else if (order.Type == ASK && order.Price < sameTree.BestPriceOrder.Price) ||
			(order.Type == BID && order.Price > sameTree.BestPriceOrder.Price) {
			// If the order has a better price, set it as the best price order
			sameTree.BestPriceOrder = order
		}
		ob.AddedOrdersCount++
	}
}

func (ob *OrderBook) MatchOrder(order *Order, sameTree *RBTree, oppositeTree *RBTree) {
	for order.Volume > 0 && order.Price > 0 {
		// fmt.Printf("Check for new match %+v \n", order)
		// Check if the opposite tree has best orders to match
		if oppositeTree.BestPriceOrder == nil {
			// Insert the order into the tree
			sameTree.Insert(order)
			// No opposite orders, update best price order
			sameTree.BestPriceOrder = order
			ob.AddedOrdersCount++
			break // No opposite orders, exit
		}

		// Check if the best opposite order can be matched with the aggressive order
		if (order.Type == BID && order.Price < oppositeTree.BestPriceOrder.Price) ||
			(order.Type == ASK && order.Price > oppositeTree.BestPriceOrder.Price) {
			// Cannot match, insert the order into the tree
			sameTree.Insert(order)

			// Update the best price order if the new order's price is better
			if (order.Type == BID && (sameTree.BestPriceOrder == nil || order.Price > sameTree.BestPriceOrder.Price)) ||
				(order.Type == ASK && (sameTree.BestPriceOrder == nil || order.Price < sameTree.BestPriceOrder.Price)) {
				sameTree.BestPriceOrder = order
			}

			ob.AddedOrdersCount++
			break
		}

		// Compute trade quantity (minimum of the two orders)
		tradeQty := math.Min(order.Volume, oppositeTree.BestPriceOrder.Volume)
		if tradeQty <= 0 {
			// No trade possible as both orders have zero volume
			break
		}

		// fmt.Printf("Matched %s order with price %f, volume %f and %s order id %s with price %f and volume %f trade volume %f \n", order.Type, order.Price, order.Volume, oppositeTree.BestPriceOrder.Type, oppositeTree.BestPriceOrder.ID, oppositeTree.BestPriceOrder.Price, oppositeTree.BestPriceOrder.Volume, tradeQty)

		// Update volumes of both orders
		order.Volume -= tradeQty
		oppositeTree.BestPriceOrder.Volume -= tradeQty

		// Update match count
		ob.MatchOrdersCount++

		// Remove the opposite order if its volume is zero
		if oppositeTree.BestPriceOrder.Volume <= 0 {
			oppositeTree.Remove(oppositeTree.BestPriceOrder)
			// fmt.Printf("New best opposite order after match %+v \n", oppositeTree.BestPriceOrder)
			ob.RemovedOrdersCount++
		}

		// Exit if the aggressive order volume is zero
		if order.Volume == 0 {
			break
		}
	}
}

func (ob *OrderBook) GetOrderStats() (int, int, int) {
	return ob.AddedOrdersCount, ob.RemovedOrdersCount, ob.MatchOrdersCount
}
