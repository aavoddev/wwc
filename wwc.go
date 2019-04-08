package main

import (
	"./tree"
	
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"os"
	"html/template"
	"log"
	"path/filepath"
)

var Handlers map[string]Handler
var wwdir string

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Returns a tuple of extracted filename and extension
func filenameext(f string) (fn string, ext string) {
	re := regexp.MustCompile(`(.*)\.([^\.]+)$`)
	fnext := re.FindStringSubmatch(f)
	if len(fnext) > 0 {
		fn = fnext[1]
		ext = fnext[2]
	} else {
		fn = f
	}
	return
}

func selecthandler(nam string) (h Handler, ok bool) {
	_, ext := filenameext(nam)
	h, ok = Handlers[ext]
	return
}

func isindex(nam string) bool {
	i, err := regexp.MatchString(`^index\.[^\.]+$`, nam)
	check(err)
	return i || nam == "INDEX"
}

// Checks tree `t` for files that don't have handlers, Adds Handler information into `t`,
// and transforms `t` and any related tree to fill in missing indecies
func prep(t *tree.Tree) {
	var nohandlers string
	
	fromn := t.Root().Creat(false)
	ton := t.Root().Rel().Creat(false)
	fromn.SetN("SITEMAP.sitemap")
	ton.SetN("sitemap.html")
	fromn.SetRel(ton)

	c := t.Root().Down(-1)
	for e := range c {
		if e.Dir() {  // Incorrect?
			c := e.Down(1)
		
			var index *tree.Node
			mul := false
			for i := range c {
				if isindex(i.N()) {
					if i == e { continue }
					han, ok := selecthandler(i.N())
					if mul {
						panic("Too many indecies")
					} else if ok {
						index = i
						i.Ud = han
					} else {
						nohandlers = fmt.Sprintln(nohandlers, i)
					}
				}
			}
			
			if index == nil {
				fmt.Println("No valid index for :: ", e)
				fromn := e.Creat(false)
				ton := e.Rel().Creat(false)
				fromn.SetN("INDEX")
				ton.SetN("index.html")
				fromn.SetRel(ton)
				fromn.Ud, _ = selecthandler("INDEX.index")
			}
		} else if !isindex(e.N()) {  // Indexes are the domain of the previous block
			han, ok := selecthandler(e.N())
			if !ok {
				nohandlers = fmt.Sprintln(nohandlers, e)
			}
			e.Ud = han
		}
	}
	if nohandlers != "" {
		log.Fatal("No Handlers for:\n", nohandlers)
	}
}

func rearr(t *tree.Tree) {
	var es []*tree.Node
	c := t.Root().Down(-1)
	for e := range c {  // Don't read and modify at the same time
		es = append(es, e)
	}
	for _, e := range es {
		if !e.Dir() {
			fn, _ := filenameext(e.N())
			if !isindex(e.N()) {
				pn := e.P().Creat(true)
				pn.SetN(fn)
				e.Move(pn)
			}
			e.SetN("index.html")
		}
	}
}

// Make sure everything's where it needs to be
func envcheck() {
	var files = []string{
		"content",
		"static",
		"tpl",
		"tpl/master.tpl",
		"tpl/index.tpl",
		"tpl/sitemap.tpl",
	}
	
	for _, f := range files {
		_, err := os.Stat(f)
		if os.IsNotExist(err) {
			fmt.Printf("err: File %s doesn't exist.", f)
			os.Exit(3)
		}
	}
}

func compilepage(e *tree.Node, master *template.Template) {
	if !e.Dir() {
		var data = pagedot{
			Sidebar: sbdot{
				Base: e.Root(),
				Cur: e,
			},
		}
		err := os.MkdirAll(e.Rel().P().Path(), os.ModePerm)
		check(err)
		out, err := os.Create(e.Rel().Path())
		check(err)
		defer out.Close()
		
		data.Title, data.Content = e.Ud.(Handler).Handle(e)
		err = master.Execute(out, data)
		check(err)
	}
}

func main() {
	envcheck()
	
	fun := template.FuncMap{
		"sbNewDot": sbNewDot,
		"Entries": Entries,
		"nodeequal": nodeequal,
		"Incl": Incl,
		"Issm": Issm,
	}
	master, err := template.New("master.tpl").Funcs(fun).ParseFiles("tpl/master.tpl")
	check(err)
	
	index, err := template.New("index.tpl").Funcs(fun).ParseFiles("tpl/index.tpl")
	check(err)
	
	sitemap, err := template.New("sitemap.tpl").Funcs(fun).ParseFiles("tpl/sitemap.tpl")
	check(err)
	
	// Setup `Handlers`
	Handlers = make(map[string]Handler)
	Handlers["md"] = &mdh{}
	Handlers["html"] = &htmlh{}
	Handlers["index"] = &indexh{ tpl: index }
	Handlers["sitemap"] = &sitemaph{ tpl: sitemap }
	
	contt, err := tree.Fsto("content")
	check(err)
	outcontt := contt.Dup()
	prep(contt)
	
	tmppub, err := ioutil.TempDir("./", ".TEMPpublic")
	check(err)
	tmppub, err = filepath.Abs(tmppub)
	check(err)
	defer os.RemoveAll(tmppub)
	outcontt.SetLoc(tmppub)
	
	rearr(outcontt)
	
	wwdir, err = filepath.Abs(".")
	check(err)
	wwdir = fmt.Sprintf("%s%c", wwdir, filepath.Separator)
	
	for e := range contt.Root().Down(-1) {
		go compilepage(e, master)
	}
	
	static, err := tree.Fsto("static")
	check(err)
	outstatic := static.Dup()
	outstatic.SetLoc(fmt.Sprintf("%s%s", tmppub, "/static"))
	
	for e := range static.Root().Down(-1) {
		if !e.Dir() {
			err := os.MkdirAll(e.Rel().P().Path(), os.ModePerm)
			check(err)
			out, err := os.Create(e.Rel().Path())
			check(err)
			in, err := os.Open(e.Path())
			check(err)
			_, err = io.Copy(out, in)
			check(err)
			out.Close()
		}
	}
	
	oldpub, err := tree.Fsto("public")
	check(err)
	for e := range oldpub.Root().Search(`^[^.]`, 1) {
		if e.N() == "CNAME" {
			continue
		}
		err = os.RemoveAll(e.Path())
		check(err)
	}
	
	frompub, err := tree.Fsto(tmppub)
	check(err)
	topub := frompub.Dup()
	topub.SetLoc("public")
	
	for e := range frompub.Root().Down(-1) {
		if !e.Dir() && e != e.Root() {
			err := os.MkdirAll(e.Rel().P().Path(), os.ModePerm)
			check(err)
			out, err := os.Create(e.Rel().Path())
			check(err)
			in, err := os.Open(e.Path())
			check(err)
			_, err = io.Copy(out, in)
			check(err)
			out.Close()
		}
	}
}
