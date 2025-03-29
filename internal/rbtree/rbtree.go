package rbtree

import (
	"fmt"
	"math"
)

type Color int

const (
	RED Color = iota
	BLACK
)

type Order struct {
	Type   string
	Status string
	ID     string
	Price  float64
	Volume float64
}

type OrderBook struct {
	BidTree *RBTree
	AskTree *RBTree
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		BidTree: &RBTree{},
		AskTree: &RBTree{},
	}
}

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
	}
}

func (pl *Order) Compare(otherPrice float64) int {
	switch pl.Type {
	case "ask": // asks (sellers) are sorted in ascending order
		if pl.Price < otherPrice {
			return -1 // pl is less than other
		} else if pl.Price > otherPrice {
			return 1 // pl is greater than other
		}
	case "bid": // bids (buyers) are sorted in descending order
		if pl.Price > otherPrice {
			return -1 // pl is less than other
		} else if pl.Price < otherPrice {
			return 1 // pl is greater than other
		}
	}
	return 0 // pl is equal to other
}

// Node represents a node in the red-black tree.
type Node struct {
	OrderList *OrderList
	Key       float64
	Color     Color
	Left      *Node
	Right     *Node
	Parent    *Node
}

type RBTree struct {
	Root           *Node
	BestPriceOrder *Order
}

func (tree *RBTree) rotateLeft(n *Node) *Node {
	rChild := n.Right
	n.Right = rChild.Left

	if rChild.Left != nil {
		rChild.Left.Parent = n
	}

	rChild.Parent = n.Parent
	if n.Parent == nil {
		tree.Root = rChild
	} else if n == n.Parent.Left {
		n.Parent.Left = rChild
	} else {
		n.Parent.Right = rChild
	}

	rChild.Left = n
	n.Parent = rChild
	return rChild
}

func (tree *RBTree) rotateRight(n *Node) *Node {
	lChild := n.Left
	n.Left = lChild.Right

	if lChild.Right != nil {
		lChild.Right.Parent = n
	}

	lChild.Parent = n.Parent
	if n.Parent == nil {
		tree.Root = lChild
	} else if n == n.Parent.Right {
		n.Parent.Right = lChild
	} else {
		n.Parent.Left = lChild
	}

	lChild.Right = n
	n.Parent = lChild
	return lChild
}

func (tree *RBTree) Insert(order *Order) {
	// Search for the node to insert the new order in the tree
	currentNode := tree.Root

	if currentNode == nil {
		// If the tree is empty, create a new node and insert the order
		newNode := &Node{
			Key:       order.Price,
			OrderList: &OrderList{}, // Create a new order list and append the order
			Color:     RED,
		}
		tree.Root = newNode
		newNode.OrderList.Append(order)
		tree.UpdateBestPriceOrder() // Update the best price level
		return
	}

	var parentNode *Node
	for currentNode != nil {
		parentNode = currentNode
		if order.Compare(currentNode.Key) < 0 {
			currentNode = currentNode.Left // Move to the left child (new order is less than the current node)
		} else if order.Compare(currentNode.Key) > 0 {
			currentNode = currentNode.Right // Move to the right child (new order is greater than the current node)
		} else {
			// Found the node with the same price level, append the order to the order list
			currentNode.OrderList.Append(order)
			tree.UpdateBestPriceOrder() // Update the best price level
			return
		}
	}

	// If the node to insert the new order is not found, create a new node and insert the order
	newNode := &Node{
		Key:       order.Price,
		OrderList: &OrderList{}, // Create a new order list and append the order
		Color:     RED,
	}
	newNode.OrderList.Append(order)

	// Add the new node to the parent node
	newNode.Parent = parentNode

	if order.Compare(parentNode.Key) < 0 {
		parentNode.Left = newNode
	} else {
		parentNode.Right = newNode
	}

	tree.fixInsert(newNode) // Fix the tree after inserting the new node

	tree.UpdateBestPriceOrder() // Update the best price level
}

func (tree *RBTree) fixInsert(n *Node) {
	parentNode := n.Parent

	// While the parent node is red, we need to restructuring the tree and / or recoloring nodes.
	for parentNode != nil && parentNode.Color == RED {
		grandparentNode := parentNode.Parent

		// Check if the grandparent node exists
		if grandparentNode == nil {
			break
		}

		// Parent node is the left child of the grandparent node.
		if parentNode == grandparentNode.Left {
			uncleNode := grandparentNode.Right

			// Case 1: The uncle (right) node is red. Recolor the parent and uncle nodes to black and the grandparent node to red.
			if uncleNode != nil && uncleNode.Color == RED {
				grandparentNode.Color = RED
				parentNode.Color = BLACK
				uncleNode.Color = BLACK
				n = grandparentNode
			} else {
				// Case 2: New Node is the right child of the grandparent node. Perform a left rotation on the parent node.
				if n == parentNode.Right {
					n = parentNode
					tree.rotateLeft(n)
				}

				// Case 3: New Node is the left child of the grandparent node. Perform a right rotation on the grandparent node.
				parentNode.Color = BLACK
				grandparentNode.Color = RED
				tree.rotateRight(grandparentNode)
			}
		} else {
			// Parent node is the right child of the grandparent node.
			uncleNode := grandparentNode.Left

			// Case 1: The uncle (left) node is red. Recolor the parent and uncle nodes to black and the grandparent node to red.
			if uncleNode != nil && uncleNode.Color == RED {
				grandparentNode.Color = RED
				parentNode.Color = BLACK
				uncleNode.Color = BLACK
				n = grandparentNode
			} else {
				// Case 2: New Node is the left child of the grandparent node. Perform a right rotation on the parent node.
				if n == parentNode.Left {
					n = parentNode
					tree.rotateRight(n)
				}

				// Case 3: New Node is the right child of the grandparent node. Perform a left rotation on the grandparent node.
				parentNode.Color = BLACK
				grandparentNode.Color = RED
				tree.rotateLeft(grandparentNode)
			}
		}

		parentNode = n.Parent
	}

	if tree.Root != nil {
		tree.Root.Color = BLACK
	}

	tree.UpdateBestPriceOrder()
}

func (tree *RBTree) Remove(order *Order) {
	nodeToDelete := tree.findNode(order.Price)
	if nodeToDelete == nil {
		return // Node not found
	}

	// Remove the order from the order list
	nodeToDelete.OrderList.Remove(order)

	// Check if the order list is empty
	if nodeToDelete.OrderList.Head == nil {
		// Order list is empty (no orders at this price level), remove the node
		// Save the original color of the node to be deleted
		originalColor := nodeToDelete.Color

		var child *Node
		if nodeToDelete.Left == nil {
			// Simple deletion of a node with one child (right). The child replaces the deleted node.
			child = nodeToDelete.Right
			tree.transplant(nodeToDelete, nodeToDelete.Right)
		} else if nodeToDelete.Right == nil {
			// Simple deletion of a node with one child (left). The child replaces the deleted node.
			child = nodeToDelete.Left
			tree.transplant(nodeToDelete, nodeToDelete.Left)
		} else {
			// Node to be deleted has two children. Find the successor node and replace the deleted node with it.
			successor := tree.minimum(nodeToDelete.Right)
			// Found the successor node. Save its original color and its right child.
			originalColor = successor.Color
			child = successor.Right

			if successor.Parent == nodeToDelete {
				child.Parent = successor
			} else {
				tree.transplant(successor, successor.Right)
				successor.Right = nodeToDelete.Right
				successor.Right.Parent = successor
			}

			tree.transplant(nodeToDelete, successor)
			successor.Left = nodeToDelete.Left
			successor.Left.Parent = successor
			successor.Color = nodeToDelete.Color
		}

		if originalColor == BLACK {
			tree.fixRemove(child)
		}

		tree.UpdateBestPriceOrder()
	}
}

func (tree *RBTree) transplant(u, v *Node) {
	if u.Parent == nil {
		tree.Root = v
	} else if u == u.Parent.Left {
		u.Parent.Left = v
	} else {
		u.Parent.Right = v
	}
	if v != nil {
		v.Parent = u.Parent
	}
}

func (tree *RBTree) fixRemove(x *Node) {
	for x != tree.Root && x.Color == BLACK {
		if x == x.Parent.Left {
			w := x.Parent.Right
			if w.Color == RED {
				w.Color = BLACK
				x.Parent.Color = RED
				tree.rotateLeft(x.Parent)
				w = x.Parent.Right
			}

			if w.Left.Color == BLACK && w.Right.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Right.Color == BLACK {
					w.Left.Color = BLACK
					w.Color = RED
					tree.rotateRight(w)
					w = x.Parent.Right
				}

				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Right.Color = BLACK
				tree.rotateLeft(x.Parent)
				x = tree.Root
			}
		} else {
			w := x.Parent.Left
			if w.Color == RED {
				w.Color = BLACK
				x.Parent.Color = RED
				tree.rotateRight(x.Parent)
				w = x.Parent.Left
			}

			if w.Right.Color == BLACK && w.Left.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Left.Color == BLACK {
					w.Right.Color = BLACK
					w.Color = RED
					tree.rotateLeft(w)
					w = x.Parent.Left
				}

				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Left.Color = BLACK
				tree.rotateRight(x.Parent)
				x = tree.Root
			}
		}
	}
	x.Color = BLACK
}

func (tree *RBTree) minimum(node *Node) *Node {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

func (tree *RBTree) findNode(price float64) *Node {
	currentNode := tree.Root
	for currentNode != nil {
		if price == currentNode.Key {
			return currentNode
		} else if price < currentNode.Key {
			currentNode = currentNode.Left
		} else {
			currentNode = currentNode.Right
		}
	}
	return nil
}

func (tree *RBTree) UpdateBestPriceOrder() {
	if tree.Root == nil {
		tree.BestPriceOrder = nil
		return
	}

	currentNode := tree.Root
	for currentNode.Left != nil {
		currentNode = currentNode.Left
	}
	// Best price order is the first order in the order list (oldest one so as to be executed first)
	tree.BestPriceOrder = currentNode.OrderList.Head.Order
}

func (ob *OrderBook) MatchOrder(order *Order, sameTree *RBTree, oppositeTree *RBTree) {
	for order.Volume > 0 {
		// Check if the opposite tree has best orders to match
		bestOppositeOrder := oppositeTree.BestPriceOrder
		if bestOppositeOrder == nil {
			// No opposite orders, insert the aggressive order into the opposite tree
			sameTree.Insert(order)
			break // No opposite orders, exit
		}

		// Check if the best opposite order can be matched with the aggressive order
		if (order.Type == "bid" && order.Price < bestOppositeOrder.Price) ||
			(order.Type == "ask" && order.Price > bestOppositeOrder.Price) {
			sameTree.Insert(order)
			break // No match, exit
		}

		// Compute trade quantity (minimum of the two orders)
		tradeQty := math.Min(order.Volume, bestOppositeOrder.Volume)

		// Update volumes of both orders
		order.Volume -= tradeQty
		bestOppositeOrder.Volume -= tradeQty

		fmt.Printf("Trade executed: %s %f @ %f\n", order.Type, tradeQty, bestOppositeOrder.Price)

		// Remove the opposite order if its volume is zero
		if bestOppositeOrder.Volume == 0 {
			oppositeTree.Remove(bestOppositeOrder)
			break
		}

		// Exit if the aggressive order volume is zero
		if order.Volume == 0 {
			break
		}
	}
}
