package tree

import (
	"container/list"
	"fmt"
	"path/filepath"
	"sync"
	"os"
	"strings"
)

type Node struct {
	dir bool			// True if node is directory, false otherwise
	n   string			// Name of node
	
	p   *Node			// Parent of node
	pe  *list.Element	// Element in parent's `sub` that corresponds to this node
	sub *list.List		// All subdirectories of node
	
	t	*Tree			// Tree to which this node belongs

	new bool			// True if this node is the result of a `Dup`
	rel *Node			// Pointer to node from or to which this tree was `Dup`ed
	
	Ud	interface{}		// A place to store any data of the user's choice
}

type Tree struct {
	root	*Node			// Root of tree
	loc		string			// Location of tree in the filesystem
	lock	*sync.RWMutex	// Exactly what it says on the tin
}

// Create new node of Tree `t`
func newn(t *Tree) *Node {
	n := new(Node)
	n.sub = list.New()
	n.t = t
	return n
}

// Return true if `a` is an descendent of `d`
func (d *Node) Desc(a *Node) bool {
	if d == a { return false }
	for e := range d.Up() {
		if e == a {
			return true
		}
	}
	return false
}

// Return true if `d` is an antecedent of `a`
func (d *Node) Antec(a *Node) bool {
	return a.Desc(d)
}

func (n *Node) String() string {
	var out string

	c := n.Up()

	if n.dir {
		out = fmt.Sprintf("%c", filepath.Separator)
	}
	for e := range c {
		if e.p != nil {
			out = fmt.Sprintf("%c%s%s", filepath.Separator, e.n, out)
		}
	}

	return out
}

// Returns absolute path to node in filesystem (hopefully)
func (n *Node) Path() string {
	return fmt.Sprintf("%s%s", n.t.loc, n.String())
}

func (n *Node) Rprint() {
	c := n.Down(-1)

	for e := range c {
		fmt.Print(e)
		if e.rel != nil {
			fmt.Print("::", e.rel)
		}
		fmt.Println()
	}
}

func (t *Tree) Rprint() {
	t.root.Rprint()
}

// Merges b2 into b1, panics if two elements match in every way but
// `.dir` differs
// Could this be made into a `walkDo`er?
func (b1 *Node) Merge(b2 *Node) {
Match:
	for b2et := b2.sub.Front(); b2et != nil; b2et = b2et.Next() {
		b2e := b2et.Value.(*Node)
		for b1et := b1.sub.Front(); b1et != nil; b1et = b1et.Next() {
			b1e := b1et.Value.(*Node)
			if b1e.n == b2e.n {
				if b1e.dir == b2e.dir {
					if b1e.dir == true {
						b1e.Merge(b2e)
						continue Match
					}
				} else {
					panic(fmt.Sprintf(
						"Discrepancy between dir vales of entries %s and %s",
						b1e.String(), b2e.String()))
				}
			}
		}
		// No matches
		b2e.p = b1
		pe := b1.sub.PushBack(b2e)
		b2e.pe = pe
	}
}

func (n *Node) Nav(nam string) (*Node) {
	c := n.Down(1)
	
	for e := range c {
		if e.n == nam {
			return e
		}
	}
	return nil
}

func Fsto(root string) (*Tree, error) {
	owd, err := os.Getwd()
	if err != nil { return nil, err }
	rootabs, err := filepath.Abs(root)
	if err != nil { return nil, err }
	os.Chdir(rootabs)
	defer os.Chdir(owd)
	
	t := new(Tree)
	t.loc = rootabs
	t.root = newn(t)
	t.root.dir = true
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if path == "." { return nil }
		
		tokp := strings.Split(path, fmt.Sprintf("%c", filepath.Separator))
		
		base := t.root
		for i, e := range tokp {
			match := base.Nav(e)
			if match != nil {
				base = match
				continue
			}
			
			br := newn(t)
			br.n = e
			br.p = base
			if i == len(tokp)-1 {
				br.dir = info.IsDir()
			}
	
			br.pe = base.sub.PushBack(br)
			base = br
		}
		
		return nil
	})
	return t, nil
}
