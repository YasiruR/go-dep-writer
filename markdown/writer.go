package markdown

import (
	"fmt"
	"github.com/YasiruR/go-dep-writer/entity"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tryfix/log"
	"os"
)

type Writer struct {
	logger log.Logger
}

func NewWriter(logger log.Logger) *Writer {
	return &Writer{logger: logger}
}

func (w *Writer) GenerateTable(outputFile string, rows []entity.Dependency) {
	extensions := parser.CommonExtensions | parser.Tables
	p := parser.NewWithExtensions(extensions)
	md := []byte(fmt.Sprintf(`
Dependency | Version | Description 
---------- | ------- | -----------  
%s`, w.rowsToMarkdown(rows)))

	html := markdown.ToHTML(md, p, nil)
	if err := w.write(outputFile, html); err != nil {
		w.logger.Error(err)
	}
	w.logger.Trace(`dependency table was constructed successfully`)
	fmt.Printf("%s", md)
}

func (w *Writer) rowsToMarkdown(rows []entity.Dependency) (out string) {
	for _, r := range rows {
		out += fmt.Sprintf("[%s](%s)        | %s      | %s        \n", r.Name, r.URL, r.Version, r.Desc)
	}
	return out
}

func (w *Writer) write(outputFile string, data []byte) error {
	if outputFile == entity.DefaultFileName {
		err := os.WriteFile(outputFile, data, 0644)
		if err != nil {
			return fmt.Errorf(`writing to the output file %s failed - %v`, outputFile, err)
		}

		w.logger.Trace(fmt.Sprintf(`saved output to %s`, outputFile))
		return nil
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		w.logger.Fatal(fmt.Sprintf(`invalid file path [%s] - %v`, outputFile, err))
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("\n%s", string(data))); err != nil {
		w.logger.Fatal(fmt.Sprintf(`updating %s failed - %v`, outputFile, err))
	}

	w.logger.Trace(fmt.Sprintf(`saved output to %s`, outputFile))
	return nil
}
