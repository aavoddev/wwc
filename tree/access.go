package tree

// Return location of `t` in filesystem
func (t *Tree) Loc() string { return t.loc }

// Set location of `t` in filesystem
func (t *Tree) SetLoc(loc string) { t.loc = loc }

// Return root of Tree `t`
func (t *Tree) Root() *Node { return t.root }

// Return root of Tree that `n` is part of
func (n *Node) Root() *Node { return n.t.root }

// Return true if `n` is a directory, false otherwise
func (n *Node) Dir() bool { return n.dir }

// Return name for `n`
func (n *Node) N() string { return n.n }

// Set name for `n`
func (n *Node) SetN(nam string) { n.n = nam }

// Return parent element of `n`
func (n *Node) P() *Node { return n.p }

// Return the relative `Node`
func (n *Node) Rel() *Node { return n.rel }

// Set relative attributes for `fn` and `tn` to point to each other.
// Set new attribute to true in `tn`
func (fn *Node) SetRel(tn *Node) {
	fn.rel = tn
	tn.rel = fn
	tn.new = true
}

// Return new Node whose parent is `n`, 
// if `dir` is true, new Node is directory
func (n *Node) Creat(dir bool) *Node {
	nn := newn(n.t)
	nn.dir = dir
	nn.t = n.t
	nn.p = n
	nn.pe = n.sub.PushBack(nn)
	return nn
}

// Remove element pointer to by `n` from it's parent
func (n *Node) Remove() {
	n.p.sub.Remove(n.pe)
}

// Move Node `n` to be a sub-Node of `np`
func (n *Node) Move(np *Node) {
	n.Remove()
	n.p = np
	n.t = np.t
	n.pe = np.sub.PushBack(n)
}
