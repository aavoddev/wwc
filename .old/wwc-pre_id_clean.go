package main

import (
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"path/filepath"
	"errors"
	"regexp"
	"container/list"  // Excessive?
	
	"gopkg.in/russross/blackfriday.v2"
)

type tree struct {
	dir	bool
	n	string
	p	*tree
	sub	[]*tree
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func trtoid(t *tree) []string {
	tmp := list.New()
	for c := t; c.p != nil; c = c.p {
		tmp.PushFront(c.n)
	}
	out := make([]string, tmp.Len())
	for e, i := tmp.Front(), 0; e != nil; e, i = e.Next(), i+1 {
		out[i] = e.Value.(string)
	}
	return out
}

// Return pointer to deepest matching tree element and index to 
// deepest matching element of `id`
/*
func trcmpid(t *tree, id []string) (*tree, int) {
	var cid string
	var idi int
	
	base := t
Match:
	for idi, cid = range id {
		for _, cbas := range base.sub {
			if cbas.n == cid {
				base = cbas
				continue Match
			}
		}
		idi--
		break
	}
	return base, idi
} */

// Return index of deepest element of both id1 and id2, -1 if no matches
func cmpid(id1 []string, id2 []string) (int) {
	var out int
	
	for out = 0; out < len(id1); out++ {
		if id1[out] != id2[out] {
			out--
			break
		}
	}
	return out
}

// Return true if `t` is an descendent of `d`
func (t *tree) desc(a *tree) bool {
	tid := trtoid(t)
	aid := trtoid(a)
	
	if cmpid(tid, aid) == len(aid)-1 && len(aid) != len(tid) {
		return true
	}
	return false
}

// Return true if `t` is an antecedent of `d`
func (t *tree) antec(d *tree) bool {
	tid := trtoid(t)
	did := trtoid(d)
	
	if cmpid(tid, did) == len(tid)-1 && len(did) != len(tid) {
		return true
	}
	return false
}

// Navigate to `id` in `root`
func idtotr(id []string, root *tree) *tree {
	base := root
Match:
	for idi, cid := range id {
		for _, cbas := range base.sub {
			if cbas.n == cid {
				base = cbas
				continue Match
			}
		}
		idi--
		break
	}
	return base
}

// Returns error if last element of `id` isn't the only unique element, or
// if last element in `id` already exists in `t`
func (t *tree) add(id []string, dir bool) error {
	if len(id) == 2 && id[0] == "" && id[1] == "" {
		id = id[:1]
	} else if id[len(id)-1] == "" {
		return errors.New("Cannot add empty element")
	}
	dt, di := trcmpid(t, id)
	
	if di != len(id)-2 {
		return errors.New("Number of unique elements in ID must be one")
	}
	if !dt.dir {
		return errors.New("Parent of unique element is not a dir")
	}
	nt := new(tree)
	dt.sub = append(dt.sub, nt)
	nt.n = id[di+1]
	nt.p = dt
	nt.dir = dir
	
	return nil
}

// CURRENT DIRECTORY APPEARS AS A SUBDIRECTORY OF ROOT OF TREE
// Sort before adding to tree?
func (t *tree) rreadfs(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		tokpath := strings.Split(path, fmt.Sprintf("%c", filepath.Separator))
		t.add(tokpath, info.IsDir())
		return nil
	})
}


func (t *tree) sprint() string {
	var out string
	var ct *tree
	
	if t.dir {
		out = fmt.Sprintf("%c", filepath.Separator)
	}
	for ct = t; ct.p.p != nil; ct = ct.p {
		out = fmt.Sprintf("%c%s%s", filepath.Separator, ct.n, out)
	}
	out = fmt.Sprintf("%s%s", ct.n, out)
	
	return out
}

func (t *tree) print() {
	fmt.Println(t.sprint())
}

func (t *tree) rprint() {	
	for _, sube := range t.sub {
		sube.print()
		sube.rprint()
	}
}

func (t *tree) _walk(c chan *tree) {
	for _, sube := range t.sub {
		c <- sube
		if sube.dir {
			sube._walk(c)
		}
	}
}

func (t *tree) walk(c chan *tree) {
	t._walk(c)
	close(c)
}

func (t *tree) search(c chan *tree, res string) {
	re := regexp.MustCompile(res)
	trc := make(chan *tree)
	go t.walk(trc)
	for trce := range trc {
		if re.MatchString(trce.n) {
			c <- trce
		}
	}
	close(c)
}

func compilesidebar(t *tree) {
	/* sdtrees := struct {
		root *tree,
		cur  *tree
	}
	tpl := []byte(`
{{ $cur := .cur }}
{{ $root := .root }}
{{define "sdbranch"}}
	{{range .sub}}
		{{if .}}
{{end}}`)
	*/
}

func compilepage(t *tree) {
	var Title string
	md, err := ioutil.ReadFile(t.sprint())
	check(err)
	
	tire := regexp.MustCompile(`(.*)\n===*\n\s*$`)
	begre := regexp.MustCompile(`(.*\n){2}.*`)
	beg := begre.Find(md)
	ti := tire.FindSubmatch(beg)
	if len(ti) > 0 {
		Title = string(ti[1])
	} else {
		Title = "WHOOPS"
	}
	
	fmt.Println(string(blackfriday.Run(md)))
	fmt.Println(Title)
}

func main() {
	dents := new(tree)
	dents.dir = true
	
	dents.rreadfs(".")
	dents.rprint()
	
/*	mds := make([]*tree, 0)
	c := make(chan *tree)
	go dents.search(c, `.*\.md`)
	for match := range c {
		mds = append(mds, match)
	}
	
	for _, ting := range mds {
		compilepage(ting)
	}
	fmt.Println(trtoid(dents.sub[1].sub[0]))
*/

}
