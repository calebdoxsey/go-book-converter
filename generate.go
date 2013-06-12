package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/badgerodon/go/dom"
	"github.com/badgerodon/go/dom/css"
	. "github.com/badgerodon/go/dom/dsl"
	"code.google.com/p/appengine-go/example/moustachio/resize"
)

var assets = "../golang-book-web/assets/"
var url = regexp.MustCompile("http(s)?://[^\\s)]+")

type Heading struct {
	level int
	text string
}

type Generator struct {
	fonts map[string]string
	output dom.Node
}
func (this *Generator) HandleStyles(root dom.Node) error {
	parents := make(map[string]string)

	for _, style := range css.Find(root, "style") {
		name, ok := style.Get("name")
		if !ok {
			continue
		}
		for _, tp := range css.Find(style, "text-properties") {
			fontName, ok := tp.Get("font-name")
			if ok {
				this.fonts[name] = fontName
			}
		}
		p, ok := style.Get("parent-style-name")
		if ok {
			parents[name] = p
		}
	}


	for c, p := range parents {
		if _, ok := this.fonts[c]; !ok {
			if f, ok := this.fonts[p]; ok {
				this.fonts[c] = f
			}
		}
	}

	return nil
}
func (this *Generator) HandleContent(root dom.Node) error {
	c := this.output
	inTable := false
	dom.Visit(root, func(node dom.Node, visit func()) {
		switch n := node.(type) {
		case *dom.Text:
			c.Append(T(n.Content))
		case *dom.Element:
			switch n.Tag {
			case "footer":
			case "h":
				ol, ok := n.Get("outline-level")
				if !ok {
					ol = "1"
				}
				p := c
				c = E("h" + ol)
				p.Append(c)
				p.Append(T("\n\n"))
				visit()
				c = p
			case "master-styles":
			case "image":
				if name, ok := n.Get("href"); ok {
					c.Append(E("img", A("class", "block"), A("src", "assets/img/" + path.Base(name))))
				}
			case "line-break":
				if o, ok := c.(*dom.Element); ok {
					if o.Tag == "pre" {
						c.Append(T("\n"))
					} else {
						c.Append(E("br"))
					}
				} else {
					c.Append(E("br"))
				}
			case "list":
				p := c
				c = E("ul")
				p.Append(c)
				visit()
				c = p
			case "list-item":
				p := c
				c = E("li")
				p.Append(c)
				visit()
				c = p
			case "p":
				if inTable {
					if this.IsBlock(n) {
						c.Set("class", "code")
					}
					visit()
				} else {
					p := c
					if this.IsBlock(n) {
						c = E("pre")
						p.Append(c)
					} else {
						c = E("p")
						p.Append(c)
						p.Append(T("\n"))
					}
					visit()
					c = p
				}
			case "s":
				cnt := 1
				if s, ok := n.Get("c"); ok {
					if i, err := strconv.Atoi(s); err == nil {
						cnt = i
					}
				}
				for i := 0; i < cnt; i++ {
					c.Append(T(" "))
				}
			case "span":
				if this.IsBlock(n) {
					p := c
					c = E("code")
					p.Append(c)
					visit()
					c = p
				} else {
					visit()
				}
			case "tab":
				c.Append(T("    "))
			case "table":
				inTable = true
				p := c
				c = E("table")
				p.Append(c)
				visit()
				c = p
				inTable = false
			case "table-row":
				p := c
				c = E("tr")
				p.Append(c)
				visit()
				c = p
			case "table-cell":
				p := c
				c = E("td")
				p.Append(c)
				visit()
				c = p
			default:
				visit()
			}
			break
		default:
			visit()
		}
	})
	return nil
}

func (this *Generator) IsBlock(node dom.Node) bool {
	v, ok := node.Get("style-name")
	if ok {
		f, ok := this.fonts[v]
		if ok {
			if f == "Consolas" || f == "Consolas1" {
				return true
			}
		}
	}
	return false
}

func (this *Generator) RemoveEmpty() {
	for _, n := range css.Find(this.output, "h1, h2, h3, h4, h5, h6, p") {
		if len(n.Children()) == 0 {
			n.Parent().Remove(n)
		}
	}
}
func (this *Generator) CollapseHeaders() {
	for _, h := range css.Find(this.output, "h1, h2, h3, h4, h5, h6") {
		for {
			p := h.Parent()
			if p == nil {
				break
			}
			el, ok := p.(*dom.Element)
			if !ok {
				break
			}
			if el.Tag == "li" || el.Tag == "ul" {
				dom.Replace(p, h)
			} else {
				break
			}
		}
	}
}
func (this *Generator) RemoveFrontMatter() {
	h := css.First(this.output, "h1")
	for _, n := range css.PrevAll(h) {
		n.Parent().Remove(n)
	}
}
func (this *Generator) AddHeaderAnchors() {
	cnt := 1
	for _, h := range css.Find(this.output, "h1, h2") {
		el, ok := h.(*dom.Element)
		if ok {
			if el.Tag == "h1" {
				cnt = 1
			} else if el.Tag == "h2" {
				el.Set("id", fmt.Sprint("section", cnt))
				cnt++
			}
		}
	}
}
func (this *Generator) AddLinks() {
	dom.Visit(this.output, func(node dom.Node, visit func()) {
		if t, ok := node.(*dom.Text); ok {
			is := url.FindAllStringIndex(t.Content, -1)
			if len(is) > 0 {
				f := F()
				f.Append(T(t.Content[:is[0][0]]))
				f.Append(E("a", A("href", t.Content[is[0][0]:is[0][1]]), A("target", "_blank"), t.Content[is[0][0]:is[0][1]]))
				f.Append(T(t.Content[is[0][1]:]))
				dom.Replace(node, f)
			}
		}
		visit()
	})
}
func (this *Generator) CleanLinks() {
	for _, n := range css.Find(this.output, "code a") {
		p := n
		for {
			p = p.Parent()
			if _, ok := p.(*dom.Element); ok {
				break
			}
		}
		dom.Replace(p, n)
	}
}
func (this *Generator) Export(name string, h1 io.Reader) {
	h2, err := os.Create(assets + "img/" + name)
	if err != nil {
		log.Fatalln(err)
	}
	defer h2.Close()
	img, err := png.Decode(h1)
	if err != nil {
		log.Fatalln(err)
	}
	r := img.Bounds()
	p := r.Size()
	w := 500
	if p.X > w {
		h := int((float64(w) / float64(p.X)) * float64(p.Y))
		img = resize.Resize(img, r, w, h)
	}
	err = png.Encode(h2, img)
	if err != nil {
		log.Fatalln(err)
	}
}

func (this *Generator) GetTOC() dom.Node {
	headings := make([]Heading, 0)
	for _, h := range css.Find(this.output, "h1, h2") {
		lvl := 0
		if e, ok := h.(*dom.Element); ok {
			if e.Tag == "h2" {
				lvl = 1
			}
		}
		headings = append(headings, Heading{lvl,dom.TextContent(h)})
	}

	ol := E("ol", A("id", "toc"))
	var h1 dom.Node = F()
	h1c := 0
	h2c := 1
	for _, h := range headings {
		if h.level == 0 {
			h1c++
			h1 = E("ol")
			ol.Append(
				E("li",
					E("a", A("href", fmt.Sprint(h1c)),
						h.text,
					),
					h1,
				),
			)
			h2c = 1
		}
		if h.level == 1 {
			h1.Append(
				E("li",
					E("a", A("href", fmt.Sprint(h1c, "#section", h2c)),
						h.text,
					),
				),
			)
			h2c++
		}
	}
	return F(
		E("h1", A("class", "title"), "An Introduction to Programming in Go"),
		E("img", A("src", "assets/img/cover.png"), A("class", "block"), A("title", "Cover")),
		E("h2", "Installers"),
		E("p", "Installs Go and a text editor.",
			E("ul",
				E("li",
					E("a", A("href", "/installers/go-install.exe"), "Windows"),
				),
				E("li",
					"OSX (",
						E("a", A("href", "/installers/go-install-x86.pkg"), "32 bit"),
						", ",
						E("a", A("href", "/installers/go-install-x64.pkg"), "64 bit"),
					")",
				),
			),
		),
		E("h2", "The Book"),
		E("p",
			E("i",
				"An Introduction to Programming in Go",
			), ".", E("br"),
			"Copyright ", H("&copy;"), " 2012 by Caleb Doxsey", E("br"),
			"ISBN: 978-1478355823",
		),
		E("p",
			"This book is available for purchase at Amazon.com in ",
			E("a", A("href", "http://www.amazon.com/An-Introduction-Programming-Go-ebook/dp/B0095MCNAO/"), "Kindle"),
			" or ",
			E("a", A("href", "http://www.amazon.com/An-Introduction-Programming-Caleb-Doxsey/dp/1478355824"), "Paperback"),
			". It is available for free online below or in ",
			E("a", A("href", "/assets/pdf/gobook.pdf"), "PDF form"),
			".",
		),
		E("p",
			"Questions, comments, corrections or concerns can be sent to ",
			E("a", A("href", "mailto:admin@golang-book.com"), "Caleb Doxsey"),
			".",
		),
		E("h2", "Table of Contents"),
		ol,
	)
}
func (this *Generator) MergeCode() {
	for _, n := range css.Find(this.output, "pre + pre") {
		var prev dom.Node
		for _, next := range n.Parent().Children() {
			if next == n {
				prev.Append(T("\n"))
				for _, c := range next.Children() {
					prev.Append(c)
				}
				next.Parent().Remove(next)
				break
			}
			prev = next
		}
	}
}

func (this *Generator) GenerateKindle() {
	dom.SelfClosingTags["mbp:pagebreak"] = true
	dom.SelfClosingTags["item"] = true
	dom.SelfClosingTags["itemref"] = true
	dom.SelfClosingTags["reference"] = true

	stylesheet := `

		code, pre {
			font-family: "Courier New", Courier, monospace;
		}
		code {
			font-style: italic;
		}
		pre {
			background: #EAEAEA;
			border: 1px solid #CCC;
			padding: 5px;
			margin: 5px;
		}
		table {
			border-collapse: collapse;
			margin: 20px auto;
		}
			tr td:first-child {
				text-align: center;
			}
			td {
				border: 1px solid #CCC;
				padding: 5px;
			}
		p {
			text-indent: 0;
		}
	`

	files := []string{}

	doc := this.output.Clone()
	// Remap the headings
	for _, n := range css.Find(doc, "h1, h2, h3, h4, h5, h6") {
		el := n.(*dom.Element)
		i, _ := strconv.Atoi(string(el.Tag[1]))
		el.Tag = string(el.Tag[0]) + fmt.Sprint(i + 1)
	}

	RemoveParagraphsFromLists(doc)

	doc.Insert(0, F(
		E("img", A("src", "assets/img/cover.png")),
	))

	// Add images
	for _, n := range css.Find(doc, "img") {
		el := n.(*dom.Element)

		fn, _ := el.Get("src")
		if fn == "" {
			continue
		}
		nfn := path.Base(strings.Replace(fn, ".png", ".jpg", -1))
		files = append(files, nfn)
		fn = "../golang-book-web/" + fn

		r, err := os.Open(fn)
		if err != nil {
			log.Fatalln(err)
		}
		defer r.Close()

		img, err := png.Decode(r)
		if err != nil {
			log.Fatalln(err)
		}

		w, err := os.Create("book/kindle/" + nfn)
		if err != nil {
			log.Fatalln(err)
		}
		defer w.Close()

		err = jpeg.Encode(w, img, &jpeg.Options{
			Quality: 90,
		})
		if err != nil {
			log.Fatalln(err)
		}

		el.Set("src", nfn)
	}

	// Write html
	chapters := make([]string, 0)
	for i, n := range css.Find(doc, "h2") {
		fn := "book/kindle/chapter-" + fmt.Sprint(i+1) + ".htm"
		files = append(files, path.Base(fn))
		chapters = append(chapters, dom.TextContent(n))
		w, err := os.Create(fn)
		if err != nil {
			log.Fatalln(err)
		}
		w.Write([]byte{0xEF,0xBB,0xBF})
		io.WriteString(w, "<html><head><style>" + stylesheet + "</style></head><body>")
		n.Export(w)
		for _, c := range css.NextTill(n, "h2") {
			c.Export(w)
		}
		io.WriteString(w, "</body></html>")
		w.Close()
	}

	w, err := os.Create("book/kindle/front.htm")
	if err != nil {
		log.Fatalln(err)
	}
	defer w.Close()
	files = append([]string{"front.htm"}, files...)
	front := F(
		E("img", A("id", "cover"), A("src", "cover.jpg")),
		E("mbp:pagebreak"),
		E("h2", A("id", "toc"), "Table of Contents"),
	)
	for i, c := range chapters {
		front.Append(
			E("a", A("href", "chapter-" + fmt.Sprint(i+1) + ".htm"), c),
		)
		front.Append(E("br"))
	}
	E("html", E("body", front)).Export(w)

	w, err = os.Create("book/kindle/book.opf")
	if err != nil {
		log.Fatalln(err)
	}
	defer w.Close()
	w.Write([]byte{0xEF,0xBB,0xBF})

	items := F()
	for i, fn := range files {
		mediaType := "application/xhtml+xml"
		if strings.HasSuffix(fn, ".jpg") {
			mediaType = "image/jpeg"
		}
		items.Append(
			E("item", A("id", fmt.Sprint("file", i)), A("href", fn), A("media-type", mediaType)),
		)
	}

	opf := F(
		H(`<?xml version="1.0" encoding="utf-8"?>`),
		E("package", A("xmlns", "http://www.idpf.org/2007/opf"), A("version", "2.0"), A("unique-identifier", "BookId"),
			E("metadata", A("xmlns:dc", "http://purl.org/dc/elements/1.1/"), A("xmlns:opf", "http://www.idpf.org/2007/opf"),
				E("dc:title", "An Introduction to Programming in Go"),
				E("dc:language", "en-us"),
				E("meta", A("name", "cover"), A("content", "file1")),
				E("dc:creator", A("opf:file-as", "Doxsey, Caleb"), A("opf:role", "aut"), "Caleb Doxsey"),
			),
			E("manifest", items),
			E("spine", A("toc", "toc"),
				E("itemref", A("idref", "file0")),
			),
			E("guide",
				E("reference", A("type", "toc"), A("title", "Table of Contents"), A("href", "front.htm#TOC")),
				E("reference", A("type", "cover"), A("title", "Cover Image"), A("href", "front.html#COVER")),
			),
		),
	)
	opf.Export(w)
}

func ReplacePreWithP(node dom.Node) {
	for _, pre := range css.Find(node, "pre") {
		p := E("p", A("class","code"))
		for _, c := range pre.Children() {
			p.Append(c)
		}
		pos := 0
		parent := pre.Parent()
		for _, c := range parent.Children() {
			if c == pre {
				break
			}
			pos++
		}
		parent.Insert(pos, p)
		parent.Remove(pre)
	}
}

func RemoveParagraphsFromLists(node dom.Node) {
	for _, n := range css.Find(node, "li p") {
		li := n.Parent()
		pos := 0
		for _, c := range li.Children() {
			if c == n {
				break
			}
			pos++
		}
		for _, c := range n.Children() {
			li.Insert(pos, c)
			pos++
		}
		li.Remove(n)
	}
}

func main() {
	r, err := zip.OpenReader("book/gobook.odt")
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Close()

	g := &Generator{make(map[string]string), F()}

	tree := F()

	for _, f := range r.File {
		switch(f.Name) {
		case "content.xml":
			fallthrough
		case "styles.xml":
			h, err := f.Open()
			if err != nil {
				log.Fatalln(err)
			}
			fragment, err := dom.FromXml(xml.NewDecoder(h))
			if err != nil {
				log.Fatalln(err)
			}
			tree.Append(fragment)
			h.Close()
		default:
			if strings.HasPrefix(f.Name, "Pictures/") {
				h, err := f.Open()
				if err != nil {
					log.Fatalln(err)
				}
				g.Export(path.Base(f.Name), h)
				h.Close()
			}
		}
	}

	g.HandleStyles(tree)
	g.HandleContent(tree)
	g.CollapseHeaders()
	g.AddHeaderAnchors()
	g.RemoveFrontMatter()
	g.RemoveEmpty()
	g.AddLinks()
	g.CleanLinks()
	g.MergeCode()

	// Generate the index
	idx, err := os.Create(assets + "htm/index.htm")
	if err != nil {
		log.Fatalln(err)
	}
	g.GetTOC().Export(idx)
	idx.Close()

	hs := css.Find(g.output, "h1")
	for i, h := range hs {
		chapter, err := os.Create(fmt.Sprint(assets + "htm/", i+1, ".htm"))
		if err != nil {
			log.Fatalln(err)
		}
		h.Export(chapter)
		for _, n := range css.NextTill(h, "h1") {
			n.Export(chapter)
		}

		var prev dom.Node = E("a", A("href", fmt.Sprint(i)), H("&larr;"), " Previous")
		var next dom.Node = E("a", A("href", fmt.Sprint(i+2)), "Next ", H("&rarr;"))
		if i == 0 {
			prev = F()
		}
		if i == len(hs)-1 {
			next = F()
		}

		f := F()
		f.Append(
			E("table", A("class", "paging"),
				E("tr",
					E("td", A("class", "prev"), prev),
					E("td", E("a", A("href", "/"), "Index")),
					E("td", A("class", "next"), next),
				),
			),
		)
		f.Export(chapter)
		chapter.Close()
	}

	g.GenerateKindle()
}
