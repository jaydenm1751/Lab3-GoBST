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

type inOrderIter struct {
	stack []*node
}

func (it *inOrderIter) pushLeft(n *node) {
	for n != nil {
		it.stack = append(it.stack, n)
		n = n.left
	}
}

func newIter(root *node) *inOrderIter {
	it := &inOrderIter{}
	it.pushLeft(root)
	return it
}
func (t *Tree) Equal(other *Tree) bool {
	it1 := newIter(t.root)
	it2 := newIter(other.root)

	for {
		v1, valid1 := it1.next()
		v2, valid2 := it2.next()

		if valid1 != valid2 {
			return false
		}
		if !valid1 && !valid2 {
			return true
		}
		if v1 != v2 {
			return false
		}

	}
}

func (it *inOrderIter) next() (int, bool) {
	if (len(it.stack) == 0) {
		return 0, false
	}
	n := it.stack[len(it.stack) - 1]
	it.stack = it.stack[:len(it.stack) - 1]
	val := n.key
	it.pushLeft(n.right)
	return val, true
}

// func equalNodes(a, b *node) bool {
// 	if a == nil && b == nil { return true }
// 	if a == nil || b == nil { return false }

// 	if a.key != b.key { return false }

// 	return equalNodes(a.left, b.left) && equalNodes(a.right, b.right)
// }

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

