package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"log"
	"os"
)

func generateTable(rows []row) {
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

func rowsToMarkdown(rows []row) (out string) {
	for _, r := range rows {
		out += fmt.Sprintf(`%s        | %s      | %s        `, r.dep, r.version, r.desc)
	}
	return out
}

func write(data []byte) {
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
