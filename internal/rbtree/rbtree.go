package rbtree

type Color int

const (
	RED Color = iota
	BLACK
)

type PriceLevel struct {
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

func (ob *OrderBook) InsertOrder(order *PriceLevel) {
	if order.Type == "ask" {
		ob.AskTree.Insert(order)
	} else {
		ob.BidTree.Insert(order)
	}
}

func (pl *PriceLevel) Compare(other *PriceLevel) int {
	switch pl.Type {
	case "ask":
		if pl.Price < other.Price {
			return -1 // pl is less than other
		} else if pl.Price > other.Price {
			return 1 // pl is greater than other
		}
	case "bid":
		if pl.Price > other.Price {
			return -1 // pl is less than other
		} else if pl.Price < other.Price {
			return 1 // pl is greater than other
		}
	}
	return 0
}

// Node represents a node in the red-black tree.
type Node struct {
	PriceLevel *PriceLevel
	Key        string
	Color      Color
	Left       *Node
	Right      *Node
	Parent     *Node
}

type RBTree struct {
	Root           *Node
	BestPriceLevel *PriceLevel
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

func (tree *RBTree) Insert(order *PriceLevel) {
	newNode := &Node{
		PriceLevel: order,
		Key:        order.ID,
		Color:      RED,
	}

	if tree.Root == nil {
		tree.Root = newNode
	} else {
		currentNode := tree.Root
		var parentNode *Node
		for currentNode != nil {
			parentNode = currentNode
			if order.Compare(currentNode.PriceLevel) < 0 {
				currentNode = currentNode.Left
			} else {
				currentNode = currentNode.Right
			}
		}

		newNode.Parent = parentNode
		if order.Compare(parentNode.PriceLevel) < 0 {
			parentNode.Left = newNode
		} else {
			parentNode.Right = newNode
		}
	}

	tree.fixInsert(newNode)

	if tree.BestPriceLevel == nil || order.Compare(tree.BestPriceLevel) < 0 {
		tree.BestPriceLevel = order
	}
}

func (tree *RBTree) fixInsert(n *Node) {
	parentNode := n.Parent

	// While the parent node is red, we need to restructuring the tree and / or recoloring nodes.
	for parentNode != nil && parentNode.Color == RED {
		grandparentNode := parentNode.Parent

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

	tree.Root.Color = BLACK
}

func (tree *RBTree) Remove(order *PriceLevel) {
	nodeToDelete := tree.findNode(order.ID)
	if nodeToDelete == nil {
		return // Node not found
	}

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

	tree.UpdateBestPriceLevel()
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

func (tree *RBTree) findNode(orderID string) *Node {
	currentNode := tree.Root
	for currentNode != nil {
		if orderID == currentNode.Key {
			return currentNode
		} else if orderID < currentNode.Key {
			currentNode = currentNode.Left
		} else {
			currentNode = currentNode.Right
		}
	}
	return nil
}

func (tree *RBTree) UpdateBestPriceLevel() {
	if tree.Root != nil {
		tree.BestPriceLevel = nil
		return
	}

	tree.BestPriceLevel = tree.findBestPriceLevel(tree.Root)
}

func (tree *RBTree) findBestPriceLevel(n *Node) *PriceLevel {
	if n == nil {
		return nil
	}

	bestPriceNode := n.PriceLevel
	if n.Left != nil {
		leftBest := tree.findBestPriceLevel(n.Left)
		if leftBest != nil && leftBest.Compare(bestPriceNode) < 0 {
			bestPriceNode = leftBest
		}
	}

	if n.Right != nil {
		rightBest := tree.findBestPriceLevel(n.Right)
		if rightBest != nil && rightBest.Compare(bestPriceNode) < 0 {
			bestPriceNode = rightBest
		}
	}

	return bestPriceNode
}
