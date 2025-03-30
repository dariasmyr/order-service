package orderbook

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

func (tree *RBTree) UpdateOrderVolume(order *Order) {
	node := tree.findNode(order.Price, order.Type)
	if node != nil {
		node.OrderList.UpdateOrderVolume(order)
	}
}

func (tree *RBTree) Remove(order *Order) {
	nodeToDelete := tree.findNode(order.Price, order.Type)
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

		// Case 1: Node to be deleted has no children. Simply remove it from the tree.
		if nodeToDelete.Left == nil && nodeToDelete.Right == nil {
			if nodeToDelete.Parent != nil {
				// If the node to be deleted has a parent, set the appropriate child of the parent node to nil
				if nodeToDelete.Parent.Left == nodeToDelete {
					nodeToDelete.Parent.Left = nil // Set the left child of the parent node to nil
				} else {
					nodeToDelete.Parent.Right = nil
				}
			} else {
				// If the node to be deleted is the root node, set the root to nil
				tree.Root = nil
			}
		} else {
			// Case 2: Node to be deleted has one or two children
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

func (tree *RBTree) findNode(price float64, orderType string) *Node {
	currentNode := tree.Root
	for currentNode != nil {
		if price == currentNode.Key {
			return currentNode
		} else if price < currentNode.Key {
			if orderType == "bid" {
				currentNode = currentNode.Right
			} else if orderType == "ask" {
				currentNode = currentNode.Left
			}
		} else {
			if orderType == "bid" {
				currentNode = currentNode.Left
			} else if orderType == "ask" {
				currentNode = currentNode.Right
			}
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
