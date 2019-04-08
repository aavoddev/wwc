package tree

import (
	"container/list"
	"regexp"
)

func (n *Node) Up() (chan *Node) {
	c := make(chan *Node)
	go func() {
		c <- n
		cur := n
		for cur.p != nil {
			cur = cur.p
			c <- cur
		}
		close(c)
	}()
	return c
}

type walkDo interface {
	do(*Node) walkDo
}

type walkFinner interface {
	walkFin()
}

func (n *Node) _walk(wdo walkDo, depth int) {
	if depth != 0 {
		depth--
		for etmp := n.sub.Front(); etmp != nil; etmp = etmp.Next() {
			e := etmp.Value.(*Node)

			nwdo := wdo.do(e)
			if e.dir {
				e._walk(nwdo, depth)
			}
		}
	}
}

// Walk down tree `t` with a maximum depth of `depth`,
// -1 `depth` for a full walk of `t`
// Depth first, never returns element before its parent
func (n *Node) walk(wdo walkDo, depth int) {
	wdo = wdo.do(n)
	n._walk(wdo, depth)
	if wfin, ok := wdo.(walkFinner); ok {
		wfin.walkFin()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// `c` for returning pair of old and new tree elements respectively
// `bas` for current tree base
type wdup struct {
	bas *Node
	t	*Tree
}

func (wdo wdup) do(sube *Node) walkDo {
	nn := new(Node)
	*nn = *sube
	
	nn.sub = list.New()
	nn.t = wdo.t
	nn.new = true
	nn.rel = sube
	sube.rel = nn
	
	if wdo.bas != nil {
		nn.p = wdo.bas
		nn.pe = wdo.bas.sub.PushBack(nn)
		
		wdo.bas = nn
	} else {
		wdo.t.root = nn
		wdo.bas = nn
	}
	return wdo
}

// Returns a duplicate of Tree `t`
func (ot *Tree) Dup() *Tree {
	nt := new(Tree)
	
	nt.loc = ot.loc
	ot.root.walk(wdup{t: nt}, -1)
	
	return nt
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type wdown struct {
	c chan *Node	// for returning each descendent element
}

func (wd wdown) do(sube *Node) walkDo {
	wd.c <- sube
	return wd
}

func (wd wdown) walkFin() {
	close(wd.c)
}

// Perform a walk of n, sending elements through returned chan.
// A depth of 0 is a no-op, A depth of -1 performs a full walk of `n`
func (n *Node) Down(depth int) (chan *Node) {
	wd := wdown {
		c: make(chan *Node),
	}

	go n.walk(wd, depth)
	return wd.c
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
/* Obsoleted as implemented in wwc now as Entries
type wsub struct {
	c		chan *Node	// For returning each sub Node
	first	bool		// For skipping first
}

func (wd wsub) do(sube *Node) walkDo {
	if wd.first {
		wd.c <- sube
	}
	wd.first = false
	return wd
}

func (wd wsub) walkFin() {
	close(wd.c)
}

// Return all sub Nodes for `n`
func (n *Node) Sub() (chan *Node) {
	wd := wsub {
		c: make(chan *Node),
		first: true,
	}

	go n.walk(wd, 1)
	return wd.c
}
*/
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type wsearch struct {
	c	chan *Node
	re	*regexp.Regexp
}

func (wdo wsearch) do(n *Node) walkDo {
	if wdo.re.MatchString(n.n) {
		wdo.c <- n
	}
	return wdo
}

func (wfin wsearch) walkFin() {
	close(wfin.c)
}

// Search for elements in `n` matching regex `res`,
// with a maximum search depth of `depth`. Return results on `c`
func (n *Node) Search(res string, depth int) (chan *Node) {
	wdo := wsearch {
		c:	make(chan *Node),
		re:	regexp.MustCompile(res),
	}
	go n.walk(wdo, depth)
	return wdo.c
}
