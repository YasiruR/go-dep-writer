package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"log"
	"os"
)

var outputFile string

func generateTable(rows []dependency) {
	extensions := parser.CommonExtensions | parser.Tables
	p := parser.NewWithExtensions(extensions)
	md := []byte(fmt.Sprintf(`
Dependency | Version | Description 
---------- | ------- | -----------  
%s`, rowsToMarkdown(rows)))

	html := markdown.ToHTML(md, p, nil)
	fmt.Printf("%s", md)
	write(html)
}

func rowsToMarkdown(rows []dependency) (out string) {
	for _, r := range rows {
		out += fmt.Sprintf("[%s](%s)        | %s      | %s        \n", r.name, r.url, r.version, r.desc)
	}
	return out
}

func write(data []byte) {
	err := os.WriteFile(outputFile, data, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
