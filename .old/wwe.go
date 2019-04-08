package main

import (
	"fmt"
	"strings"
	_ "io/ioutil"
	"os"
	"path/filepath"
	"errors"
)

type tree struct {
	dir	bool
	n	string
	p	*tree
	sub	[]*tree
}

// Delete entry and, subsequently, any subentries.
// 
func (t *tree) del() {
	s := &t.p.sub
	for i, cs := range *s {
		if cs.n == t.n { // either this mess or linked list
			(*s)[len(*s)-1], (*s)[i] = (*s)[i], (*s)[len(*s)-1]
			(*s) = (*s)[:len(*s)-1]
		}
	}
}

// Elements from p, first element mapping to the root of t,
// are added to t.
func (t *tree) add(id []string, dir bool) error {
	var nt *tree
	var fi int
	var cid string
	base := t
	
Match:
	for i, cid := range id {
		for _, cbas := range base.sub {
			if cbas.n == cid {
				base = cbas
				continue Match
			}
		}
		fi = i
		break
	}
	if base.dir == false { 
		return errors.New("Parent isn't a directory") }
	if fi+1 < len(id) && id[fi+1] != id[len(id)-1] && id[fi+1] != "" {
		return errors.New("Nonexistant parent directory in identifier") }
	if fi+1 > len(id) {
		return errors.New("Entry already exists") }
	if cid == "" && base.p != nil {
		return nil }
	
	nt = new(tree)
	base.sub = append(base.sub, nt)
	nt.n = cid
	nt.p = base
	nt.dir = dir
	return nil
}

func (t *tree) print() {
	var out string
	var ct *tree
	
	if t.dir {
		out = fmt.Sprintf("%c", filepath.Separator)
	}
	for ct = t; ct.p.p != nil; ct = ct.p {
		out = fmt.Sprintf("%c%s%s", filepath.Separator, ct.n, out)
	}
	out = fmt.Sprintf("%s%s", ct.n, out)
	
	fmt.Println(out)
}

func (t *tree) rprint() {	
	for _, sube := range t.sub {
		sube.print()
		sube.rprint()
	}
}

// CURRENT DIRECTORY APPEARS AS A SUBDIRECTORY OF ROOT OF TREE
func (t *tree) rreadfs(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		tokpath := strings.Split(path, fmt.Sprintf("%c", filepath.Separator))
		t.add(tokpath, info.IsDir())
		fmt.Println(tokpath)
		return nil
	})
}

func main() {
	dents := new(tree)
	dents.dir = true
	
//	dents.rreadfs("./")

	id := strings.Split("/test/ing/pls/help", fmt.Sprintf("%c", filepath.Separator))
	dents.add(id, true)
	id = strings.Split("./.", fmt.Sprintf("%c", filepath.Separator))
	err := dents.add(id, true)
	fmt.Println(err)

	dents.rprint()
}

/* 	dents := new(tree)
	
	id := strings.Split("/test/ing/pls/help", fmt.Sprintf("%c", filepath.Separator))
	dents.add(id)
	id = strings.Split("/test/ing/WHHATT", fmt.Sprintf("%c", filepath.Separator))
	dents.add(id)
	
	dents.rprint() */
