package main

import (
	"./tree"
	
	"fmt"
	"html/template"
	"regexp"
	"io/ioutil"
)

type sbdot struct {
	Base *tree.Node
	Cur *tree.Node
}

type pagedot struct {
	Title string
	Content template.HTML
	Sidebar sbdot
}

type Entry struct {
	Nam		string				// Name of entry
	Lin		string				// Hyperlink to entry
	Dir		bool				// True if entry is a directory, false otherwise
	Nod		*tree.Node			// corresponding Node for entry
	Attrs	template.HTMLAttr	// Attributes for this entry
}

func Incl(fn string) (template.CSS) {
	out, err := ioutil.ReadFile(fmt.Sprintf("%s%s", wwdir, fn))
	check(err)
	return template.CSS(out)
}

func sbNewDot(Base, Cur *tree.Node) sbdot {
	return sbdot{
		Base: Base,
		Cur: Cur,
	}
}

// Return true if Node pointers are identical
func nodeequal(n1, n2 *tree.Node) bool {
	return n1 == n2
}

// Is the node the site map?
func Issm(nam string) bool {
	return nam == "SITEMAP.sitemap"
}

func Namefix(n string) string {
	fn, _ := filenameext(n)
	return regexp.MustCompile(`_`).ReplaceAllLiteralString(fn, ` `)
}

// Return pretty entries for building the navigation tree
func Entries(n, cn *tree.Node) (chan Entry) {
	c := make(chan Entry)
	
	go func(){
		for e := range n.Down(1) {
			if e != n && !isindex(e.N()) && !Issm(e.N()) {
				var ent Entry
				ent.Nam = Namefix(e.N())
				ent.Dir = e.Dir() // Potentially empty dirs will be valid dirs
				if isindex(e.Rel().N()) {
					ent.Lin = e.Rel().P().String()
				} else {
					ent.Lin = e.Rel().String()
				}
				ent.Nod = e
				if e.Antec(cn) || e == cn {
					ent.Attrs = template.HTMLAttr(`class="antec"`)
				}
				
				c <- ent
			}
		}
		close(c)
	}()
	return c
}
