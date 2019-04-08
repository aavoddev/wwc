package main

import (
	"./tree"
	
	"strings"
	"html/template"
	"io/ioutil"
	"regexp"
	"gopkg.in/russross/blackfriday.v2"
)

type Handler interface {
	Handle(*tree.Node) (string, template.HTML)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type indexh struct {
	tpl *template.Template
}

func (s *indexh) Handle(n *tree.Node) (string, template.HTML) {
	b := new(strings.Builder)
	var data = pagedot{
		Sidebar: sbdot{
			Base: n.P(),
			Cur: n,
		},
		Title: Namefix(n.P().N()),
	}
	
	err := s.tpl.Execute(b, data)
	check(err)
	return data.Title, template.HTML(b.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type mdh struct {}

func (s *mdh) Handle(n *tree.Node) (string, template.HTML) {
	var title string
	cont, err := ioutil.ReadFile(n.Path())
	check(err)
	// Assumes text file is using Unix \n
	re := regexp.MustCompile(`(?m)\A\s*(.*?)\s*\n===*\n\s*\n`)
	
	ti := re.FindSubmatch(cont)
	if len(ti) > 0 {
		title = string(ti[1])
	} else {
		title = ""
	}

	return title, template.HTML(blackfriday.Run(cont))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type htmlh struct {}

func (s *htmlh) Handle(n *tree.Node) (string, template.HTML) {
	var title string
	cont, err := ioutil.ReadFile(n.Path())
	check(err)
	// Assumes text file is using Unix `\n`
	re := regexp.MustCompile(`(?i)\A(?:.*\n){0,32}.*<title>\s*([^<]+?)\s*<\/title>`)
	ti := re.FindSubmatch(cont)
	if len(ti) > 0 {
		title = string(ti[1])
	} else {
		title = ""
	}
	return title, template.HTML(cont)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type sitemaph struct {
	tpl *template.Template
}

func (s *sitemaph) Handle(n *tree.Node) (string, template.HTML) {
	b := new(strings.Builder)
	var data = pagedot{
		Sidebar: sbdot{
			Base: n.Root(),
			Cur: n.Root(), 
		},
		Title: "Sitemap",
	}
	
	err := s.tpl.Execute(b, data)
	check(err)
	return data.Title, template.HTML(b.String())
}