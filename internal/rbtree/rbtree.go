package rbtree

const (
	RED   = true
	BLACK = false
)

// Node represents a node in the red-black tree.
type Node struct {
	price  float64
	size   float64
	color  bool
	left   *Node
	right  *Node
	parent *Node
}

type RBTree struct {
	root *Node
}

func (t *RBTree) leftRotate(x *Node) {
	y := x.right
	x.right = y.left

	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent

	if x.parent == nil {
		t.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}

	y.left = x
	x.parent = y
}

func (t *RBTree) rightRotate(y *Node) {
	x := y.left
	y.left = x.right

	if x.right != nil {
		x.right.parent = y
	}
	x.parent = y.parent

	if y.parent == nil {
		t.root = x
	} else if y == y.parent.right {
		y.parent.right = x
	} else {
		y.parent.left = x
	}

	x.right = y
	y.parent = x
}

func (t *RBTree) insert(price, size float64) {
	newNode := &Node{price: price, size: size, color: RED}
	var y *Node
	x := t.root

	for x != nil {
		y = x
		if newNode.price < x.price {
			x = x.left
		} else {
			x = x.right
		}
	}

	newNode.parent = y
	if y == nil {
		t.root = newNode
	} else if newNode.price < y.price {
		y.left = newNode
	} else {
		y.right = newNode
	}

	t.fixInsert(newNode)
}

func (t *RBTree) fixInsert(z *Node) {
	for z.parent != nil && z.parent.color == RED {
		if z.parent == z.parent.parent.left {
			y := z.parent.parent.right
			if y != nil && y.color == RED {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					z = z.parent
					t.leftRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.rightRotate(z.parent.parent)
			}
		} else {
			y := z.parent.parent.left
			if y != nil && y.color == RED {
				z.parent.color = BLACK
				y.color = BLACK
				z.parent.parent.color = RED
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					z = z.parent
					t.rightRotate(z)
				}
				z.parent.color = BLACK
				z.parent.parent.color = RED
				t.leftRotate(z.parent.parent)
			}
		}
	}
	t.root.color = BLACK
}

func (t *RBTree) getMin() *Node {
	x := t.root
	if x == nil {
		return nil
	}
	for x.left != nil {
		x = x.left
	}
	return x
}

func (t *RBTree) getMax() *Node {
	x := t.root
	if x == nil {
		return nil
	}
	for x.right != nil {
		x = x.right
	}
	return x
}
