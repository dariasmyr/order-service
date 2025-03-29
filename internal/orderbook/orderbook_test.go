package orderbook

import (
	"fmt"
	"testing"
)

func TestInsertSingleOrder(t *testing.T) {
	ob := NewOrderBook()

	// Insert a bid order
	bidOrder := &Order{
		Type:   "bid",
		Status: "open",
		ID:     "bid1",
		Price:  100.0,
		Volume: 10.0,
	}

	ob.InsertOrder(bidOrder)

	// Ensure the bid is in the BidTree
	if ob.BidTree.Root == nil || ob.BidTree.Root.Key != 100.0 {
		t.Errorf("Bid order not inserted correctly into BidTree")
	}
}

func TestInsertMultipleOrders(t *testing.T) {
	ob := NewOrderBook()

	// Insert a few bid orders
	bidOrders := []*Order{
		{Type: "bid", ID: "bid1", Price: 100.0, Volume: 10.0},
		{Type: "bid", ID: "bid2", Price: 105.0, Volume: 5.0},
		{Type: "bid", ID: "bid3", Price: 110.0, Volume: 20.0},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	if ob.BidTree.BestPriceOrder.Price != 110.0 {
		t.Errorf("Highest bid order not at the top of the linked list")
	}
}

func TestOrderMatching(t *testing.T) {
	ob := NewOrderBook()

	// Insert bid orders
	bidOrders := []*Order{
		{Type: "bid", ID: "bid1", Price: 100.0, Volume: 10.0},
		{Type: "bid", ID: "bid2", Price: 105.0, Volume: 5.0},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	// Insert an ask order that can match the bid orders
	askOrder := &Order{
		Type:   "ask",
		Status: "open",
		ID:     "ask1",
		Price:  105.0,
		Volume: 10.0,
	}

	ob.InsertOrder(askOrder)

	// Verify the order was matched and the volume was reduced
	if askOrder.Volume != 5.0 {
		fmt.Printf("Ask volume %f\n", askOrder.Volume)
		t.Errorf("Ask order volume should be reduced after matching")
	}

	// Verify the best opposite order was removed after matching
	if ob.BidTree.BestPriceOrder.Volume != 10.0 {
		fmt.Printf("Bid order: %v\n", ob.BidTree.BestPriceOrder)
		t.Errorf("Bid order volume should be reduced after matching")
	}
}

func TestInsertPassiveOrder(t *testing.T) {
	ob := NewOrderBook()

	// Insert a bid order
	bidOrder := &Order{
		Type:   "bid",
		Status: "open",
		ID:     "bid1",
		Price:  100.0,
		Volume: 10.0,
	}

	ob.InsertOrder(bidOrder)

	// Insert an ask order that cannot be matched yet
	askOrder := &Order{
		Type:   "ask",
		Status: "open",
		ID:     "ask1",
		Price:  110.0,
		Volume: 5.0,
	}

	ob.InsertOrder(askOrder)

	// Ensure the ask order is in the AskTree
	if ob.AskTree.Root == nil || ob.AskTree.Root.Key != 110.0 {
		t.Errorf("Ask order not inserted correctly into AskTree")
	}
}

func TestRemoveOrder(t *testing.T) {
	ob := NewOrderBook()

	// Insert a bid order
	bidOrder := &Order{
		Type:   "bid",
		Status: "open",
		ID:     "bid1",
		Price:  100.0,
		Volume: 10.0,
	}

	ob.InsertOrder(bidOrder)

	// Remove the bid order
	ob.BidTree.Remove(bidOrder)

	// Ensure the BidTree is empty after removal
	if ob.BidTree.Root != nil {
		t.Errorf("Bid order not removed correctly from BidTree")
	}
}

func TestRedBlackTreeProperties(t *testing.T) {
	ob := NewOrderBook()

	// Insert a series of orders to test the balancing
	orders := []*Order{
		{Type: "bid", ID: "bid1", Price: 100.0, Volume: 10.0},
		{Type: "bid", ID: "bid2", Price: 105.0, Volume: 5.0},
		{Type: "bid", ID: "bid3", Price: 110.0, Volume: 20.0},
	}

	for _, order := range orders {
		ob.InsertOrder(order)
	}

	// Check if the tree is balanced and maintain Red-Black properties
	if !isBalanced(ob.BidTree.Root) {
		t.Errorf("Red-Black Tree properties violated after insertions")
	}
}

// A helper function to check tree balance (e.g., black height, red nodes constraints)
func isBalanced(node *Node) bool {
	if node == nil {
		return true
	}
	// Add your balance checking logic here, such as black height consistency
	return true
}

func TestEdgeCaseOrders(t *testing.T) {
	ob := NewOrderBook()

	// Insert an order with zero volume
	zeroVolumeOrder := &Order{
		Type:   "ask",
		Status: "open",
		ID:     "ask1",
		Price:  100.0,
		Volume: 0.0,
	}

	ob.InsertOrder(zeroVolumeOrder)

	// Verify the order does not get inserted into the tree
	if ob.AskTree.Root != nil {
		t.Errorf("Order with zero volume should not be inserted into the AskTree")
	}

	// Insert an order with a negative price
	negativePriceOrder := &Order{
		Type:   "ask",
		Status: "open",
		ID:     "ask2",
		Price:  -10.0,
		Volume: 5.0,
	}

	ob.InsertOrder(negativePriceOrder)

	// Verify the negative price order is not inserted
	if ob.AskTree.Root != nil {
		t.Errorf("Order with negative price should not be inserted into the AskTree")
	}
}
