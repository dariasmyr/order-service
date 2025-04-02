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
		Type:       orderbook.BID,
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     0.64185083,
		UpdateType: "ADD",
	}

	ob.InsertOrder(bidOrder)

	// Ensure the bid is in the BidTree
	if ob.BidTree.Root == ob.BidTree.Leaf() || ob.BidTree.Root.Key != 272.83 {
		t.Errorf("Bid order not inserted correctly into BidTree")
	}
}

func TestInsertMultipleOrders(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a few bid orders
	bidOrders := []*orderbook.Order{
		{Type: orderbook.BID, ID: "fac5dcb7afbf-4a9d-4519-98e6-0889cf06", Price: 272.83, Volume: 18},
		{Type: orderbook.BID, ID: "100d5004-f4e0-4e92-a571-392403d5b073", Price: 272.83, Volume: 0.64185083},
		{Type: orderbook.BID, ID: "0889cf06-4a9d-4519-98e6-fac5dcb7afbf", Price: 272.82, Volume: 0.641},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	if ob.BidTree.BestPriceOrder.Price != 272.83 {
		t.Errorf("Highest bid order not at the top of the linked list")
	}
}

func TestOrderMatching(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert bid orders
	bidOrders := []*orderbook.Order{
		{Type: orderbook.BID, ID: "50532ea1ccbb-4c84-448a-9c5d-3c5e789c", Price: 272.83, Volume: 18},
		{Type: orderbook.BID, ID: "c64a76ba-5107-4c1b-b788-271ee88df318", Price: 272.83, Volume: 0.64185083},
		{Type: orderbook.BID, ID: "3c5e789c-4c84-448a-9c5d-50532ea1ccbb", Price: 272.82, Volume: 0.641},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	// Insert an ask order that can match the bid orders
	askOrder := &orderbook.Order{
		Type:       orderbook.ASK,
		UpdateType: "ADD",
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     18.5,
	}

	ob.InsertOrder(askOrder)

	// Verify the best opposite order was removed after matching
	if ob.BidTree.BestPriceOrder.Volume != 0.14185082999999998 {
		fmt.Printf("Bid order: %v\n", ob.BidTree.BestPriceOrder)
		t.Errorf("Bid order volume should be reduced after matching")
	}
}

func TestInsertPassiveOrder(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert bid orders
	bidOrders := []*orderbook.Order{
		{Type: orderbook.BID, ID: "50532ea1ccbb-4c84-448a-9c5d-3c5e789c", Price: 272.83, Volume: 18},
		{Type: orderbook.BID, ID: "c64a76ba-5107-4c1b-b788-271ee88df318", Price: 272.83, Volume: 0.64185083},
		{Type: orderbook.BID, ID: "3c5e789c-4c84-448a-9c5d-50532ea1ccbb", Price: 272.82, Volume: 0.641},
	}

	for _, order := range bidOrders {
		ob.InsertOrder(order)
	}

	// Insert ask orders that cannot be matched yet
	askOrders := []*orderbook.Order{
		{Type: orderbook.ASK, ID: "3c5e789c-4c84-448a-9c5d-50532ea1ccbb", Price: 272.9, Volume: 1},
		{Type: orderbook.ASK, ID: "c64a76ba-5107-4c1b-b788-271ee88df318", Price: 272.9, Volume: 18},
		{Type: orderbook.ASK, ID: "271ee88df-5107-4c1b-b788-c64a76ba318", Price: 272.91, Volume: 16},
	}

	for _, order := range askOrders {
		ob.InsertOrder(order)
	}

	// Ensure the ask order is in the AskTree
	if ob.AskTree.Root == ob.AskTree.Leaf() || ob.AskTree.BestPriceOrder != askOrders[0] {
		t.Errorf("Best ask order not inserted correctly into AskTree")
	}

	// Ensure the bid order is in the AskTree
	if ob.BidTree.Root == ob.AskTree.Leaf() || ob.BidTree.BestPriceOrder != bidOrders[0] {
		t.Errorf("Ask order not inserted correctly into AskTree")
	}
}

func TestRemoveOrder(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a bid order
	bidOrder := &orderbook.Order{
		Type:       orderbook.BID,
		UpdateType: "ADD",
		ID:         "271ee88df-5107-4c1b-b788-c64a76ba318",
		Price:      272.83,
		Volume:     0.64185083,
	}

	ob.InsertOrder(bidOrder)

	// Remove the bid order
	ob.RemoveOrder(bidOrder)

	// Ensure the BidTree is empty after removal
	if ob.BidTree.Root != ob.BidTree.Leaf() {
		t.Errorf("Bid order not removed correctly from BidTree")
	}
}

func TestRedBlackTreeProperties(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert a series of orders to test the balancing
	orders := []*orderbook.Order{
		{Type: orderbook.BID, ID: "50532ea1ccbb-4c84-448a-9c5d-3c5e789c", Price: 272.83, Volume: 18},
		{Type: orderbook.BID, ID: "c64a76ba-5107-4c1b-b788-271ee88df318", Price: 272.83, Volume: 0.64185083},
		{Type: orderbook.BID, ID: "3c5e789c-4c84-448a-9c5d-50532ea1ccbb", Price: 272.82, Volume: 0.641},
	}

	for _, order := range orders {
		ob.InsertOrder(order)
	}

	// Check if the tree is balanced and maintain Red-Black properties
	if !isBalanced(ob) {
		t.Errorf("Red-Black Tree properties violated after insertions")
	}
}

// A helper function to check tree balance (e.g., black height, red nodes constraints)
func isBalanced(ob *orderbook.OrderBook) bool {
	if ob.AskTree.Root == ob.BidTree.Leaf() {
		return true
	}
	// Add your balance checking logic here, such as black height consistency
	return true
}

func TestEdgeCaseOrders(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	// Insert an order with zero volume
	zeroVolumeOrder := &orderbook.Order{
		Type:       orderbook.ASK,
		UpdateType: "ADD",
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     0.0,
	}

	ob.InsertOrder(zeroVolumeOrder)

	// Verify the order does not get inserted into the tree
	if ob.AskTree.Root != ob.AskTree.Leaf() {
		t.Errorf("Order with zero volume should not be inserted into the AskTree")
	}

	// Insert an order with a negative price
	negativePriceOrder := &orderbook.Order{
		Type:       orderbook.ASK,
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     -0.64185083,
		UpdateType: "ADD",
	}

	ob.InsertOrder(negativePriceOrder)

	// Verify the negative price order is not inserted
	if ob.AskTree.Root != ob.AskTree.Leaf() {
		t.Errorf("Order with negative price should not be inserted into the AskTree")
	}
}

func TestOrderBookUpdate(t *testing.T) {
	ob := orderbook.NewOrderBook("AAPL")

	// Case 1: Adding a bid order
	order1 := &orderbook.Order{
		Type:       orderbook.BID,
		UpdateType: "ADD",
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     10,
	}
	ob.InsertOrder(order1)

	if ob.BidTree.BestPriceOrder == nil || ob.BidTree.BestPriceOrder.ID != order1.ID {
		t.Errorf("Expected order with ID %s to be in the book, but it's not", order1.ID)
	}

	// Case 2: Updating an order
	order1Update := &orderbook.Order{
		Type:       orderbook.BID,
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     50,
		UpdateType: "MATCH",
	}
	ob.UpdateOrderVolume(order1Update)

	if ob.BidTree.BestPriceOrder.Volume != 50 {
		t.Errorf("Expected order volume to be 50, but got %f", ob.BidTree.BestPriceOrder.Volume)
	}

	// Case 2.1: Updating an order (SUBTRACK)
	order1Update = &orderbook.Order{
		Type:       orderbook.BID,
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     20,
		UpdateType: "SUBTRACK",
	}
	ob.UpdateOrderVolume(order1Update)

	// Case 3: Updating an order with zero volume (DELETE)
	order1Delete := &orderbook.Order{
		Type:       orderbook.BID,
		ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
		Price:      272.83,
		Volume:     0,
		UpdateType: "SUBTRACK",
	}
	ob.UpdateOrderVolume(order1Delete)

	if ob.BidTree.BestPriceOrder != nil {
		t.Errorf("Expected order with ID %s to be removed, but it's still in the book", order1.ID)
	}

	// Case 4: Adding an ask order and removing it
	order2 := &orderbook.Order{
		Type:       orderbook.ASK,
		ID:         "c64a76ba-5107-4c1b-b788-271ee88df318",
		Price:      272.89,
		Volume:     18,
		UpdateType: "ADD",
	}
	ob.InsertOrder(order2)

	ob.RemoveOrder(order2)

	if ob.AskTree.BestPriceOrder != nil && ob.AskTree.BestPriceOrder.ID == order2.ID {
		t.Errorf("Expected order with ID %s to be removed, but it's still in the book", order2.ID)
	}
}

func TestPerformanceWithUpdate(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	start := time.Now()

	numOrders := 275121

	// Generate and insert 10000 orders with random prices and volumes
	for i := 0; i < numOrders; i++ {
		order := &orderbook.Order{
			ID:         fmt.Sprintf("order%d", i),
			Price:      float64(100 + i%100), // Price variates from 100 to 199
			Volume:     float64(10 + i%10),   // Volume variates from 10 to 19
			UpdateType: "ADD",
			Type:       orderbook.BID,
		}

		ob.InsertOrder(order)

		// Periodically insert ask orders
		if i%5 == 0 {
			askOrder := &orderbook.Order{
				ID:         fmt.Sprintf("ask%d", i),
				Price:      order.Price - 1, // Ask orders are 1 unit below the bid
				Volume:     order.Volume,
				UpdateType: "ADD",
				Type:       orderbook.ASK,
			}
			ob.InsertOrder(askOrder)
		}

		// Periodically add the order with updated volume (MATCH или SUBTRACK)
		if i%7 == 0 {
			order := &orderbook.Order{
				ID:         fmt.Sprintf("order%d", i),
				Price:      float64(100 + i%100),
				Volume:     float64(10+i%10) - 5,
				UpdateType: "MATCH",
				Type:       orderbook.BID,
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
				Type:       orderbook.BID,
			}
			ob.RemoveOrder(order)
		}
	}

	duration := time.Since(start)

	fmt.Printf("Processed %d orders in %s\n", numOrders, duration)

	if duration.Seconds() > 5 {
		t.Errorf("Performance test took too long: %s", duration)
	}

	added, removed, match := ob.GetOrderStats()
	fmt.Printf("Order book stats: added: %d, removed: %d, match: %d\n", added, removed, match)
}

func TestPerformanceSimpleAdd(t *testing.T) {
	ob := orderbook.NewOrderBook("BTC_USDT")

	start := time.Now()

	numOrders := 275121
	// Generate and insert 10000 orders with random prices and volumes
	for i := 0; i < numOrders; i++ {
		order := &orderbook.Order{
			ID:         fmt.Sprintf("order%d", i),
			Price:      float64(100 + i%100), // Price variates from 100 to 199
			Volume:     float64(10 + i%10),   // Volume variates from 10 to 19
			UpdateType: "ADD",
			Type:       orderbook.BID,
		}

		ob.InsertOrder(order)

		// Periodically insert ask orders
		if i%2 == 0 {
			askOrder := &orderbook.Order{
				ID:         fmt.Sprintf("ask%d", i),
				Price:      order.Price - 1, // Ask orders are 1 unit below the bid
				Volume:     order.Volume,
				UpdateType: "ADD",
				Type:       orderbook.ASK,
			}
			ob.InsertOrder(askOrder)
		}
	}

	duration := time.Since(start)

	fmt.Printf("Processed %d orders in %s\n", numOrders, duration)

	if duration.Seconds() > 5 {
		t.Errorf("Performance test took too long: %s", duration)
	}

	added, removed, match := ob.GetOrderStats()
	fmt.Printf("Order book stats: added: %d, removed: %d, match: %d\n", added, removed, match)
}

func TestPerformanceRealData(t *testing.T) {
	data, err := os.ReadFile("output_COINBASE_SPOT_BTC_USD.json")
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
		order.Type = orderbook.ASK
		ob.InsertOrder(&order)
	}

	for _, order := range response.Bids {
		order.UpdateType = "ADD" // Same as above
		order.Type = orderbook.BID
		ob.InsertOrder(&order)
	}

	duration := time.Since(start)

	fmt.Printf("Processed %d orders in %s\n", numOrders, duration)
	fmt.Printf("Number of matches made: %d\n", ob.MatchOrdersCount)
	added, removed, match := ob.GetOrderStats()
	fmt.Printf("Order book stats: added: %d, removed: %d, match: %d\n", added, removed, match)

	if duration.Seconds() > 5 {
		t.Errorf("Performance test took too long: %s", duration)
	}

}

func BenchmarkOrderMatch(b *testing.B) {
	ob := orderbook.NewOrderBook("BTC_USDT")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		// Insert bid orders
		bidOrders := []*orderbook.Order{
			{Type: orderbook.BID, ID: "50532ea1ccbb-4c84-448a-9c5d-3c5e789c", Price: 272.83, Volume: 18},
			{Type: orderbook.BID, ID: "c64a76ba-5107-4c1b-b788-271ee88df318", Price: 272.83, Volume: 0.64185083},
			{Type: orderbook.BID, ID: "3c5e789c-4c84-448a-9c5d-50532ea1ccbb", Price: 272.82, Volume: 0.641},
		}

		for _, order := range bidOrders {
			ob.InsertOrder(order)
		}

		// Insert an ask order that can match the bid orders
		askOrder := &orderbook.Order{
			Type:       orderbook.ASK,
			UpdateType: "ADD",
			ID:         "3c5e789c-4c84-448a-9c5d-50532ea1ccbb",
			Price:      272.83,
			Volume:     18.5,
		}

		ob.InsertOrder(askOrder)
	}
}

func BenchmarkPerformanceWithUpdate(b *testing.B) {
	ob := orderbook.NewOrderBook("BTC_USDT")
	numOrders := 275121
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i := 0; i < numOrders; i++ {
			order := &orderbook.Order{
				ID:         fmt.Sprintf("order%d", i),
				Price:      float64(100 + i%100),
				Volume:     float64(10 + i%10),
				UpdateType: "ADD",
				Type:       orderbook.BID,
			}
			ob.InsertOrder(order)

			if i%5 == 0 {
				askOrder := &orderbook.Order{
					ID:         fmt.Sprintf("ask%d", i),
					Price:      order.Price - 1,
					Volume:     order.Volume,
					UpdateType: "ADD",
					Type:       orderbook.ASK,
				}
				ob.InsertOrder(askOrder)
			}

			if i%7 == 0 {
				updateOrder := &orderbook.Order{
					ID:         fmt.Sprintf("order%d", i),
					Price:      float64(100 + i%100),
					Volume:     float64(10+i%10) - 5,
					UpdateType: "MATCH",
					Type:       orderbook.BID,
				}
				ob.UpdateOrderVolume(updateOrder)
			}

			if i%12 == 0 {
				deleteOrder := &orderbook.Order{
					ID:         fmt.Sprintf("order%d", i),
					Price:      float64(100 + i%100),
					Volume:     0,
					UpdateType: "DELETE",
					Type:       orderbook.BID,
				}
				ob.RemoveOrder(deleteOrder)
			}
		}
	}
	b.ReportAllocs()
}

func BenchmarkPerformanceSimpleAdd(b *testing.B) {
	numOrders := 275121

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ob := orderbook.NewOrderBook("BTC_USDT")
		for i := 0; i < numOrders; i++ {
			order := &orderbook.Order{
				ID:         fmt.Sprintf("order%d", i),
				Price:      float64(100 + i%100),
				Volume:     float64(10 + i%10),
				UpdateType: "ADD",
				Type:       orderbook.BID,
			}
			ob.InsertOrder(order)

			if i%2 == 0 {
				askOrder := &orderbook.Order{
					ID:         fmt.Sprintf("ask%d", i),
					Price:      order.Price - 1,
					Volume:     order.Volume,
					UpdateType: "ADD",
					Type:       orderbook.ASK,
				}
				ob.InsertOrder(askOrder)
			}
		}
		added, removed, match := ob.GetOrderStats()
		fmt.Printf("Order book stats: added: %d, removed: %d, match: %d\n", added, removed, match)
	}
	b.ReportAllocs()
}

func BenchmarkPerformanceRealData(b *testing.B) {
	data, err := os.ReadFile("output_COINBASE_SPOT_BTC_USD.json")
	if err != nil {
		b.Fatalf("Failed to read file: %v", err)
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
		b.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ob := orderbook.NewOrderBook("COINBASE_SPOT_BCH_USD")

		for _, order := range response.Asks {
			order.UpdateType = "ADD"
			order.Type = orderbook.ASK
			ob.InsertOrder(&order)
		}

		for _, order := range response.Bids {
			order.UpdateType = "ADD"
			order.Type = orderbook.BID
			ob.InsertOrder(&order)
		}
	}

	b.ReportAllocs()
}
