package bst

type node struct {
	key int
	left *node
	right *node
}

type Tree struct {
	root *node
}

func New() *Tree { return &Tree{} }

func (t *Tree) Insert(k int){
	if t.root == nil {
		t.root = &node{key: k}
		return
	}

	cur := t.root
	for {
		if k == cur.key {
			return
		}
		if k < cur.key {
			if cur.left == nil{
				cur.left = &node{key: k}
				return
			}
			cur = cur.left
		} else {
			if cur.right == nil {
				cur.right = &node{key: k}
				return
			}
			cur = cur.right
		}
	}
}

func (t *Tree) Equal(n *Tree) bool {
	return equalNodes(t.root, n.root)
}

func equalNodes(a, b *node) bool {
	if a == nil && b == nil { return true }
	if a == nil || b == nil { return false }

	if a.key != b.key { return false }

	return equalNodes(a.left, b.left) && equalNodes(a.right, b.right)
}

func (t *Tree) HashValue() int {
    hash := 1
    var inorder func(*node)
    inorder = func(n *node) {
        if n == nil {
            return
        }
        inorder(n.left)
        nv := n.key + 2
        hash = (hash*nv + nv) % 1000
        inorder(n.right)
    }
    inorder(t.root)
    return hash
}

