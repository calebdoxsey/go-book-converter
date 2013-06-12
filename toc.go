package main

import (
	"fmt"
	"github.com/badgerodon/go/dom"
	"github.com/badgerodon/go/dom/css"
	. "github.com/badgerodon/go/dom/dsl"
)

func (this *Generator) GetTOC() dom.Node {
	headings := make([]Heading, 0)
	for _, h := range css.Find(this.output, "h1, h2") {
		lvl := 0
		if e, ok := h.(*dom.Element); ok {
			if e.Tag == "h2" {
				lvl = 1
			}
		}
		headings = append(headings, Heading{lvl, dom.TextContent(h)})
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
