package main

import (
	"gopkg.in/russross/blackfriday.v2"
	"html/template"
	"net/http"
	"strings"
)

type BasicRendered struct {
	Name     string
	Rendered template.HTML
	Func     string
}
type RenderFunc func(string, []byte, http.ResponseWriter) // filename, Text content, writer
var RENDERFUNCS map[string]RenderFunc = map[string]RenderFunc{
	"markdown": func(filename string, input []byte, w http.ResponseWriter) {
		rendered := blackfriday.Run(input)
		executeTemplate(w, "basicRendered.html", BasicRendered{filename, template.HTML(rendered), "markdown"})
	},
	"sent": func(filename string, input []byte, w http.ResponseWriter) {
		output := "<div>"
		for _, para := range strings.Split(string(input), "\n\n") {
			output += "<div tabindex=\"0\" class=\"slide\">"
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
		executeTemplate(w, "basicRendered.html", BasicRendered{filename, template.HTML(output), "sent"})
	},
}
