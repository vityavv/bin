package main

import (
//	"github.com/gomarkdown/markdown"
	//"html/template"
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
)

type Rendered struct{
	Name string
	Rendered template.HTML
}
type RenderFunc func(string) (template.HTML, error) // Text content, output, err
var RENDERFUNCS map[string]RenderFunc = map[string]RenderFunc{
	"markdown": func(input string) (template.HTML, error) {
		rendered := blackfriday.Run([]byte(input))
		return template.HTML(rendered), nil
	},
}

