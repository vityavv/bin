package main

import (
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"strings"
)

type Rendered struct {
	Name     string
	Rendered template.HTML
	Func     string
}
type RenderFunc func([]byte) (template.HTML, error) // Text content, output, err
var RENDERFUNCS map[string]RenderFunc = map[string]RenderFunc{
	"markdown": func(input []byte) (template.HTML, error) {
		rendered := blackfriday.Run(input)
		return template.HTML(rendered), nil
	},
	"sent": func(input []byte) (template.HTML, error) {
		output := "<div>"
		for _, para := range strings.Split(string(input), "\n\n") {
			output += "<div tabindex=\"0\" class=\"slide\">" // TODO: Turn into mini template
			if len(para) > 0 && para[0] == '@' {
				src := para[1:]
				ind := strings.IndexRune(src, '\n')
				if ind > -1 {
					src = src[:ind]
				}
				output += "<img src=\"" + src + "\"/></div>"
				continue
			}
			lines := strings.Split(string(para), "\n")
			for _, line := range lines {
				if len(line) > 0 && line[0] == '\\' {
					line = line[1:]
				} else if len(line) > 0 && line[0] == '#' {
					continue
				}
				output += line + "<br/>"
			}
			output += "</div>"
		}
		output += `</div>
<script src="https://cdn.jsdelivr.net/gh/STRML/textFit@2.4.0/textFit.min.js"></script>
<script src="/static/sent.js"></script>`
		return template.HTML(output), nil
	},
}
