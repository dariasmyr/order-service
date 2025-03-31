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
	leaf           *Node
	len            int
}

func (tree *RBTree) Len() int {
	return tree.len
}

func (tree *RBTree) Leaf() *Node {
	return tree.leaf
}

func (tree *RBTree) rotateLeft(n *Node) *Node {
	rChild := n.Right
	n.Right = rChild.Left

	if rChild.Left != tree.leaf {
		rChild.Left.Parent = n
	}

	rChild.Parent = n.Parent
	if n.Parent == tree.leaf {
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

	if lChild.Right != tree.leaf {
		lChild.Right.Parent = n
	}

	lChild.Parent = n.Parent
	if n.Parent == tree.leaf {
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
	nodeToInsert := tree.findNode(order.Price, order.Type)

	// Case: nodeToInsert found and its not leaf - simply insert the order
	if nodeToInsert != tree.leaf {
		nodeToInsert.OrderList.Append(order)
		return
	}

	// Case: nodeToInsert not found: search for the parent
	parent := tree.leaf
	current := tree.Root

	// Standart BTS search: loop utnil reach the leaf of the current node
	for current != tree.leaf {
		parent = current
		if order.Compare(current.Key) < 0 {
			current = current.Left // Move to the left child (new order is less than the current node)
		} else {
			current = current.Right // Move to the right child (new order is greater than the current node)
		}
	}

	// Parent found: create a new node
	nodeToInsert = &Node{
		Key:       order.Price,
		OrderList: &OrderList{}, // Create a new order list and append the order
		Color:     RED,          // New nodes always RED
		Parent:    parent,
		Left:      tree.leaf,
		Right:     tree.leaf,
	}
	nodeToInsert.OrderList.Append(order)

	// Insert node to tree
	// Case: Parent is leaf, set node to root
	if parent == tree.leaf {
		tree.Root = nodeToInsert
	} else if order.Compare(parent.Key) < 0 {
		parent.Left = nodeToInsert // Case: Parent is not leaf, set node to root
	} else {
		parent.Right = nodeToInsert // Case: Parent is not leaf, set node to root
	}

	// As we deleted BestPriceOrder, find the new (first node on the tree: for bids tree is in descending order, for asks tree is ia ascending order) node for (O(log N))
	if tree.BestPriceOrder == nil || order.Compare(tree.BestPriceOrder.Price) < 0 {
		tree.BestPriceOrder = order
	}

	tree.len++

	// Rebalance the tree after inserting the new node
	tree.fixInsert(nodeToInsert)
}

func (tree *RBTree) fixInsert(nodeToInsert *Node) {
	// While the parent node is red, we need to restructuring the tree and / or recoloring nodes.
	for nodeToInsert.Parent.Color == RED {
		// Parent node is the left child of the grandparent node.
		if nodeToInsert.Parent == nodeToInsert.Parent.Parent.Left {
			rightUncle := nodeToInsert.Parent.Parent.Right
			// Case 1: The uncle (right) node is red. Recolor the parent and uncle nodes to black and the grandparent node to red.
			if rightUncle.Color == RED {
				nodeToInsert.Parent.Color = BLACK
				rightUncle.Color = BLACK
				nodeToInsert.Parent.Parent.Color = RED
				// Move to grandparentNode to continue recoloring
				nodeToInsert = nodeToInsert.Parent.Parent
			} else {
				// Case 2: New Node is the right child of the grandparent node. Perform a left rotation on the parent node.
				if nodeToInsert == nodeToInsert.Parent.Right {
					// Move to parent node to perform rotation
					nodeToInsert = nodeToInsert.Parent
					tree.rotateLeft(nodeToInsert)
				}

				// Case 3: New Node is the left child of the grandparent node. Perform a right rotation on the grandparent node.
				nodeToInsert.Parent.Color = BLACK
				nodeToInsert.Parent.Parent.Color = RED
				// Move to grandparent node to perform rotation
				tree.rotateRight(nodeToInsert.Parent.Parent)
			}
		} else {
			// Parent node is the right child of the grandparent node.
			leftUncle := nodeToInsert.Parent.Parent.Left
			if leftUncle.Color == RED {
				nodeToInsert.Parent.Color = BLACK
				leftUncle.Color = BLACK
				nodeToInsert.Parent.Parent.Color = RED
				// Move to grandparentNode to continue recoloring
				nodeToInsert = nodeToInsert.Parent.Parent
			} else {
				// Case 2: New Node is the left child of the grandparent node. Perform a right rotation on the parent node.
				if nodeToInsert == nodeToInsert.Parent.Left {
					// Move to parent node to perform rotation
					nodeToInsert = nodeToInsert.Parent
					tree.rotateRight(nodeToInsert)
				}

				// Case 3: New Node is the right child of the grandparent node. Perform a left rotation on the grandparent node.
				nodeToInsert.Parent.Color = BLACK
				nodeToInsert.Parent.Parent.Color = RED
				// Move to grandparent node to perform rotation
				tree.rotateLeft(nodeToInsert.Parent.Parent)
			}
		}
	}

	tree.Root.Color = BLACK
}

func (tree *RBTree) UpdateOrderVolume(order *Order) {
	node := tree.findNode(order.Price, order.Type)
	if node != tree.leaf {
		node.OrderList.UpdateOrderVolume(order)
	}
}

func (tree *RBTree) Remove(order *Order) {
	// Find node with order to delete
	node := tree.findNode(order.Price, order.Type)
	if node == tree.leaf {
		return // Node not found
	}

	// Remove the order from the order list
	node.OrderList.Remove(order)

	// Check if the order list is empty, if yes - delete the node
	if node.OrderList.Head == nil {
		// Order list is empty (no orders at this price level), remove the node
		// Save the original color of the node to be deleted
		originalColor := node.Color

		var replacement *Node

		if node.Left == tree.leaf {
			// Node to be deleted has right child. Replace the deleted node with it.
			replacement = node.Right
			tree.transplant(node, node.Right)
		} else if node.Right == tree.leaf {
			// Node to be deleted has left child. Replace the deleted node with it.
			replacement = node.Left
			tree.transplant(node, node.Left)
		} else {
			// Node to be deleted has two children. Find the successor node and replace the deleted node with it.
			successor := tree.minimum(node.Right)
			originalColor = successor.Color

			replacement = successor.Right

			if successor != node.Right {
				tree.transplant(successor, successor.Right)
				successor.Right = node.Right
				successor.Right.Parent = successor
			} else {
				replacement.Parent = successor
			}

			tree.transplant(node, successor)
			successor.Left = node.Left
			successor.Left.Parent = successor
			successor.Color = node.Color
		}

		tree.len--

		if originalColor == BLACK {
			tree.fixRemove(replacement)
		}

		// As we deleted BestPriceOrder, find the new (first node on the tree: for bids tree is in descending order, for asks tree is ia ascending order) node for (O(log N))
		if tree.BestPriceOrder.ID == order.ID {
			if tree.Root == tree.leaf {
				tree.BestPriceOrder = nil
			} else {
				tree.BestPriceOrder = tree.minimum(tree.Root).OrderList.Head.Order
			}
		}
	} else {
		// Node is not deleted, so we make next after removed order a new best opposite order (for O(1))
		tree.BestPriceOrder = node.OrderList.Head.Order
	}
}

func (tree *RBTree) transplant(nodeToBeReplaced, nodeThatReplaces *Node) {
	if nodeToBeReplaced.Parent == tree.leaf {
		tree.Root = nodeThatReplaces
	} else if nodeToBeReplaced == nodeToBeReplaced.Parent.Left {
		nodeToBeReplaced.Parent.Left = nodeThatReplaces
	} else {
		nodeToBeReplaced.Parent.Right = nodeThatReplaces
	}
	nodeThatReplaces.Parent = nodeToBeReplaced.Parent
}

func (tree *RBTree) fixRemove(nodeToBalance *Node) {
	for nodeToBalance != tree.Root && nodeToBalance.Color == BLACK {
		if nodeToBalance == nodeToBalance.Parent.Left {
			sibling := nodeToBalance.Parent.Right
			if sibling.Color == RED {
				// Case 1: Sibling is red. Perform a rotation to move the red sibling into place.
				sibling.Color = BLACK
				nodeToBalance.Parent.Color = RED
				tree.rotateLeft(nodeToBalance.Parent)
				sibling = nodeToBalance.Parent.Right
			}

			// Case 2: Sibling is black, and both of its children are black.
			if sibling.Left.Color == BLACK && sibling.Right.Color == BLACK {
				sibling.Color = RED
				nodeToBalance = nodeToBalance.Parent
			} else {
				// Case 3: Sibling is black, and its left child is red, or right child is red.
				if sibling.Right.Color == BLACK {
					// Case 3a: Sibling's left child is red.
					sibling.Left.Color = BLACK
					sibling.Color = RED
					tree.rotateRight(sibling)
					sibling = nodeToBalance.Parent.Right
				}

				// Case 3b: Sibling's right child is red.
				sibling.Color = nodeToBalance.Parent.Color
				nodeToBalance.Parent.Color = BLACK
				sibling.Right.Color = BLACK
				tree.rotateLeft(nodeToBalance.Parent)
				nodeToBalance = tree.Root
			}
		} else {
			// Symmetric cases for when the node to fix is the right child.
			sibling := nodeToBalance.Parent.Left

			if sibling.Color == RED {
				// Case 1: Sibling is red. Perform a rotation to move the red sibling into place.
				sibling.Color = BLACK
				nodeToBalance.Parent.Color = RED
				tree.rotateRight(nodeToBalance.Parent)
				sibling = nodeToBalance.Parent.Left
			}

			// Case 2: Sibling is black, and both of its children are black.
			if sibling.Left.Color == BLACK && sibling.Right.Color == BLACK {
				sibling.Color = RED
				nodeToBalance = nodeToBalance.Parent
			} else {
				// Case 3: Sibling is black, and its left child is red, or right child is red.
				if sibling.Left.Color == BLACK {
					// Case 3a: Sibling's right child is red.
					sibling.Right.Color = BLACK
					sibling.Color = RED
					tree.rotateLeft(sibling)
					sibling = nodeToBalance.Parent.Left
				}

				// Case 3b: Sibling's left child is red.
				sibling.Color = nodeToBalance.Parent.Color
				nodeToBalance.Parent.Color = BLACK
				sibling.Left.Color = BLACK
				tree.rotateRight(nodeToBalance.Parent)
				nodeToBalance = tree.Root
			}
		}
	}

	tree.Root.Color = BLACK
}

func (tree *RBTree) minimum(node *Node) *Node {
	for node.Left != tree.leaf {
		node = node.Left
	}
	return node
}

func (tree *RBTree) findNode(price float64, orderType string) *Node {
	currentNode := tree.Root
	for currentNode != tree.leaf {
		compareResult := orderType == "bid" && price > currentNode.Key || orderType == "ask" && price < currentNode.Key

		if price == currentNode.Key {
			return currentNode
		}

		if compareResult {
			currentNode = currentNode.Left
		} else {
			currentNode = currentNode.Right
		}
	}
	return tree.leaf
}
