package rbtree

type Color int

const (
	RED Color = iota
	BLACK
)

// Node represents a node in the red-black tree.
type Node struct {
	Key    string
	Color  Color
	Left   *Node
	Right  *Node
	Parent *Node
}

type RBTree struct {
	Root *Node
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

func (tree *RBTree) Insert(key string) {
	newNode := &Node{
		Key:   key,
		Color: RED,
	}

	if tree.Root == nil {
		tree.Root = newNode
	} else {
		currentNode := tree.Root
		var parentNode *Node
		for currentNode != nil {
			parentNode = currentNode
			if key < currentNode.Key {
				currentNode = currentNode.Left
			} else {
				currentNode = currentNode.Right
			}
		}

		newNode.Parent = parentNode
		if newNode.Key < parentNode.Key {
			parentNode.Left = newNode
		} else {
			parentNode.Right = newNode
		}
	}

	tree.fixInsert(newNode)
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
