package orderbook

type OrderNode struct {
	Order *Order
	Next  *OrderNode
	Prev  *OrderNode
}

type OrderList struct {
	Head *OrderNode
	Tail *OrderNode
}

func (ol *OrderList) Append(order *Order) {
	newNode := &OrderNode{Order: order}
	if ol.Tail == nil {
		ol.Head = newNode
		ol.Tail = newNode
	} else {
		ol.Tail.Next = newNode
		newNode.Prev = ol.Tail
		ol.Tail = newNode
	}
}

func (ol *OrderList) Pop() *Order {
	if ol.Head == nil {
		return nil
	}
	order := ol.Head.Order
	ol.Head = ol.Head.Next
	if ol.Head != nil {
		ol.Head.Prev = nil
	}
	return order
}

func (ol *OrderList) Remove(order *Order) {
	node := ol.Head
	for node != nil {
		if node.Order.ID == order.ID {
			if node.Prev != nil {
				node.Prev.Next = node.Next
			} else {
				ol.Head = node.Next
			}
			if node.Next != nil {
				node.Next.Prev = node.Prev
			} else {
				ol.Tail = node.Prev
			}
			return
		}
		node = node.Next
	}
}
