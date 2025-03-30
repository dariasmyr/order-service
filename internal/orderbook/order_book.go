package orderbook

import (
	"math"
)

type OrderBook struct {
	SymbolID           string
	BidTree            *RBTree
	AskTree            *RBTree
	MatchOrdersCount   int
	AddedOrdersCount   int
	RemovedOrdersCount int
}

func NewOrderBook(symbolId string) *OrderBook {
	return &OrderBook{
		SymbolID: symbolId,
		BidTree:  &RBTree{},
		AskTree:  &RBTree{},
	}
}

func (ob *OrderBook) RemoveOrder(order *Order) {
	if order.Type == "bid" {
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
	if order.Type == "bid" {
		ob.BidTree.UpdateOrderVolume(order)
	} else {
		ob.AskTree.UpdateOrderVolume(order)
	}
}

func (ob *OrderBook) InsertOrder(order *Order) {
	var oppositeTree *RBTree
	var sameTree *RBTree

	if order.Volume <= 0 || order.Price <= 0 {
		return
	}

	if order.Type == "ask" {
		oppositeTree = ob.BidTree // Buyers
		sameTree = ob.AskTree     // Sellers
	} else {
		oppositeTree = ob.AskTree // Sellers
		sameTree = ob.BidTree     // Buyers
	}

	// Check if opposite tree exists and has best orders to match
	bestOppositeOrder := oppositeTree.BestPriceOrder
	if bestOppositeOrder != nil &&
		((order.Type == "bid" && order.Price >= bestOppositeOrder.Price) ||
			(order.Type == "ask" && order.Price <= bestOppositeOrder.Price)) {
		// Order can aggressive, match it with opposite best orders
		ob.MatchOrder(order, sameTree, oppositeTree)
	} else {
		// Order is passive, insert it into the same tree
		sameTree.Insert(order)

		// Check if this order has the best price, and update the best price order accordingly
		if sameTree.BestPriceOrder == nil ||
			(order.Type == "ask" && order.Price < sameTree.BestPriceOrder.Price) ||
			(order.Type == "bid" && order.Price > sameTree.BestPriceOrder.Price) {
			// If the order has a better price, set it as the best price order
			sameTree.BestPriceOrder = order
		}

		// If the order is not best order, update the best price order
		if sameTree.BestPriceOrder != order {
			sameTree.UpdateBestPriceOrder() // Update the best price level
		}
		ob.AddedOrdersCount++
	}
}

func (ob *OrderBook) MatchOrder(order *Order, sameTree *RBTree, oppositeTree *RBTree) {
	for order.Volume > 0 && order.Price > 0 {
		// Check if the opposite tree has best orders to match
		bestOppositeOrder := oppositeTree.BestPriceOrder
		if bestOppositeOrder == nil {
			// Insert the order into the tree
			sameTree.Insert(order)
			// No opposite orders, update best price order
			sameTree.BestPriceOrder = order
			ob.AddedOrdersCount++
			break // No opposite orders, exit
		}

		// Check if the best opposite order can be matched with the aggressive order
		if (order.Type == "bid" && order.Price < bestOppositeOrder.Price) ||
			(order.Type == "ask" && order.Price > bestOppositeOrder.Price) {
			// Cannot match, insert the order into the tree
			sameTree.Insert(order)

			// Update the best price order if the new order's price is better
			if (order.Type == "bid" && (sameTree.BestPriceOrder == nil || order.Price > sameTree.BestPriceOrder.Price)) ||
				(order.Type == "ask" && (sameTree.BestPriceOrder == nil || order.Price < sameTree.BestPriceOrder.Price)) {
				sameTree.BestPriceOrder = order
			}

			ob.AddedOrdersCount++
			break
		}

		// Compute trade quantity (minimum of the two orders)
		tradeQty := math.Min(order.Volume, bestOppositeOrder.Volume)
		if tradeQty <= 0 {
			// No trade possible as both orders have zero volume
			break
		}

		// Update volumes of both orders
		order.Volume -= tradeQty
		bestOppositeOrder.Volume -= tradeQty

		// Update match count
		ob.MatchOrdersCount++

		// Remove the opposite order if its volume is zero
		if bestOppositeOrder.Volume <= 0 {
			oppositeTree.Remove(bestOppositeOrder)
			ob.RemovedOrdersCount++
			break
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
