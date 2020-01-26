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
type RenderFunc func([]byte) (template.HTML, error) // Text content, output, err
var RENDERFUNCS map[string]RenderFunc = map[string]RenderFunc{
	"markdown": func(input []byte) (template.HTML, error) {
		rendered := blackfriday.Run(input)
		return template.HTML(rendered), nil
	},
}

