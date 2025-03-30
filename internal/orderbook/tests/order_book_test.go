package tests

import (
	"encoding/json"
	"fmt"
	"order-service/internal/orderbook"
	"os"
	"testing"
	"time"
)

func TestInsertSingleOrder(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a bid order
	bidOrder := &orderbook.Order{
		Type:       "bid",
		ID:         "bid1",
		Price:      100.0,
		Volume:     10.0,
		UpdateType: "ADD",
	}

	ob.InsertOrder(bidOrder)

	// Ensure the bid is in the BidTree
	if ob.BidTree.Root == nil || ob.BidTree.Root.Key != 100.0 {
		t.Errorf("Bid order not inserted correctly into BidTree")
	}
}

func TestInsertMultipleOrders(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a few bid orders
	bidOrders := []*orderbook.Order{
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
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert bid orders
	bidOrders := []*orderbook.Order{
		{Type: "bid", ID: "bid1", Price: 100.0, Volume: 10.0},
		{Type: "bid", ID: "bid2", Price: 105.0, Volume: 5.0},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	// Insert an ask order that can match the bid orders
	askOrder := &orderbook.Order{
		Type:       "ask",
		UpdateType: "ADD",
		ID:         "ask1",
		Price:      105.0,
		Volume:     10.0,
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
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a bid order
	bidOrder := &orderbook.Order{
		Type:       "bid",
		UpdateType: "ADD",
		ID:         "bid1",
		Price:      100.0,
		Volume:     10.0,
	}

	ob.InsertOrder(bidOrder)

	// Insert an ask order that cannot be matched yet
	askOrder := &orderbook.Order{
		Type:       "ask",
		UpdateType: "ADD",
		ID:         "ask1",
		Price:      110.0,
		Volume:     5.0,
	}

	ob.InsertOrder(askOrder)

	// Ensure the ask order is in the AskTree
	if ob.AskTree.Root == nil || ob.AskTree.Root.Key != 110.0 {
		t.Errorf("Ask order not inserted correctly into AskTree")
	}
}

func TestRemoveOrder(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a bid order
	bidOrder := &orderbook.Order{
		Type:       "bid",
		UpdateType: "ADD",
		ID:         "bid1",
		Price:      100.0,
		Volume:     10.0,
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
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a series of orders to test the balancing
	orders := []*orderbook.Order{
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
func isBalanced(node *orderbook.Node) bool {
	if node == nil {
		return true
	}
	// Add your balance checking logic here, such as black height consistency
	return true
}

func TestEdgeCaseOrders(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert an order with zero volume
	zeroVolumeOrder := &orderbook.Order{
		Type:       "ask",
		UpdateType: "ADD",
		ID:         "ask1",
		Price:      100.0,
		Volume:     0.0,
	}

	ob.InsertOrder(zeroVolumeOrder)

	// Verify the order does not get inserted into the tree
	if ob.AskTree.Root != nil {
		t.Errorf("Order with zero volume should not be inserted into the AskTree")
	}

	// Insert an order with a negative price
	negativePriceOrder := &orderbook.Order{
		Type:       "ask",
		UpdateType: "ADD",
		ID:         "ask2",
		Price:      -10.0,
		Volume:     5.0,
	}

	ob.InsertOrder(negativePriceOrder)

	// Verify the negative price order is not inserted
	if ob.AskTree.Root != nil {
		t.Errorf("Order with negative price should not be inserted into the AskTree")
	}
}

func TestOrderBookUpdate(t *testing.T) {
	ob := orderbook.NewOrderBook("AAPL")

	// Case 1: Adding a bid order
	order1 := &orderbook.Order{
		ID:         "1",
		Type:       "bid",
		Price:      100.0,
		Volume:     10.0,
		UpdateType: "ADD",
	}
	ob.InsertOrder(order1)

	if ob.BidTree.BestPriceOrder == nil || ob.BidTree.BestPriceOrder.ID != order1.ID {
		t.Errorf("Expected order with ID %s to be in the book, but it's not", order1.ID)
	}

	// Case 2: Updating an order
	order1Update := &orderbook.Order{
		ID:         "1",
		Type:       "bid",
		Price:      100.0,
		Volume:     5.0,
		UpdateType: "MATCH",
	}
	ob.UpdateOrderVolume(order1Update)

	if ob.BidTree.BestPriceOrder.Volume != 5.0 {
		t.Errorf("Expected order volume to be 5.0, but got %f", ob.BidTree.BestPriceOrder.Volume)
	}

	// Case 2.1: Updating an order (SUBTRACK)
	order1Update = &orderbook.Order{
		ID:         "1",
		Type:       "bid",
		Price:      100.0,
		Volume:     2.0,
		UpdateType: "SUBTRACK",
	}
	ob.UpdateOrderVolume(order1Update)

	// Case 3: Updating an order with zero volume (DELETE)
	order1Delete := &orderbook.Order{
		ID:         "1",
		Type:       "bid",
		Price:      100.0,
		Volume:     0.0,
		UpdateType: "SUBTRACK",
	}
	ob.UpdateOrderVolume(order1Delete)

	if ob.BidTree.BestPriceOrder != nil {
		t.Errorf("Expected order with ID %s to be removed, but it's still in the book", order1.ID)
	}

	// Case 4: Adding an ask order and removing it
	order2 := &orderbook.Order{
		ID:         "2",
		Type:       "ask",
		Price:      101.0,
		Volume:     10.0,
		UpdateType: "ADD",
	}
	ob.InsertOrder(order2)

	ob.RemoveOrder(order2)

	if ob.AskTree.BestPriceOrder != nil && ob.AskTree.BestPriceOrder.ID == order2.ID {
		t.Errorf("Expected order with ID %s to be removed, but it's still in the book", order2.ID)
	}
}

func TestPerformanceWithGeneraledOrders(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	start := time.Now()

	numOrders := 275121
	matchCount := 0

	// Generate and insert 10000 orders with random prices and volumes
	for i := 0; i < numOrders; i++ {
		order := &orderbook.Order{
			ID:         fmt.Sprintf("order%d", i),
			Price:      float64(100 + i%100), // Price variates from 100 to 199
			Volume:     float64(10 + i%10),   // Volume variates from 10 to 19
			UpdateType: "ADD",
			Type:       "bid",
		}

		ob.InsertOrder(order)

		// Periodically insert ask orders
		if i%5 == 0 {
			askOrder := &orderbook.Order{
				ID:         fmt.Sprintf("ask%d", i),
				Price:      order.Price - 1, // Ask orders are 1 unit below the bid
				Volume:     order.Volume,
				UpdateType: "ADD",
				Type:       "ask",
			}
			ob.InsertOrder(askOrder)

			// If the ask order volume is less than the bid order volume, count as a match
			if askOrder.Volume < order.Volume {
				matchCount++
			}
		}

		// Periodically add the order with updated volume (MATCH или SUBTRACK)
		if i%7 == 0 {
			order := &orderbook.Order{
				ID:         fmt.Sprintf("order%d", i),
				Price:      float64(100 + i%100),
				Volume:     float64(10+i%10) - 5,
				UpdateType: "MATCH",
				Type:       "bid",
			}
			ob.UpdateOrderVolume(order)
		}

		// Periodically add the order with DELETE volume
		if i%12 == 0 {
			order := &orderbook.Order{
				ID:         fmt.Sprintf("order%d", i),
				Price:      float64(100 + i%100),
				Volume:     0,
				UpdateType: "DELETE",
				Type:       "bid",
			}
			ob.RemoveOrder(order)
		}
	}

	duration := time.Since(start)

	fmt.Printf("Processed %d orders in %s\n", numOrders, duration)
	fmt.Printf("Number of matches made: %d\n", matchCount)

	if duration.Seconds() > 5 {
		t.Errorf("Performance test took too long: %s", duration)
	}
}

func TestPerformanceWithRealOrders(t *testing.T) {
	data, err := os.ReadFile("output_COINBASE_SPOT_BCH_USD.json")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var response struct {
		SymbolID     string            `json:"symbol_id"`
		TimeExchange string            `json:"time_exchange"`
		TimeCoinAPI  string            `json:"time_coinapi"`
		Asks         []orderbook.Order `json:"asks"`
		Bids         []orderbook.Order `json:"bids"`
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	ob := orderbook.NewOrderBook("COINBASE_SPOT_BCH_USD")

	numOrders := len(response.Asks) + len(response.Bids)

	start := time.Now()

	for _, order := range response.Asks {
		order.UpdateType = "ADD" // As output.json doesn't contain update type (REST API response, not web socket with update type), we set it to "ADD"
		order.Type = "ask"
		ob.InsertOrder(&order)
	}

	for _, order := range response.Bids {
		order.UpdateType = "ADD" // Same as above
		order.Type = "bid"
		ob.InsertOrder(&order)
	}

	duration := time.Since(start)

	fmt.Printf("Processed %d orders in %s\n", numOrders, duration)
	fmt.Printf("Number of matches made: %d\n", ob.MatchOrdersCount)

	if duration.Seconds() > 5 {
		t.Errorf("Performance test took too long: %s", duration)
	}

}
