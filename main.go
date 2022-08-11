package main

import (
	"flag"
	"github.com/YasiruR/go-dep-writer/entity"
	"github.com/YasiruR/go-dep-writer/markdown"
	"github.com/YasiruR/go-dep-writer/mod"
	"github.com/tryfix/log"
	"strings"
)

func main() {
	modFile, outputFile, user, pw, domains := parseArgs()
	logger := log.Constructor.Log(log.WithColors(true), log.WithLevel("DEBUG"), log.WithFilePath(true))

	p := mod.NewParser(user, pw, domains, logger)
	go p.Parse(modFile)
	deps := p.DependencyList()

	w := markdown.NewWriter(logger)
	w.GenerateTable(outputFile, deps)
}

func parseArgs() (modFile, outputFile, user, pw string, domains []string) {
	mf := flag.String(`modfile`, `go.mod`, `relative file path of the go.mod file`)
	u := flag.String(`user`, ``, `username of github account [optional]`)
	s := flag.String(`secret`, ``, `secret of github api [optional]`)
	dl := flag.String(`domains`, ``, `domain list of imports eg:github [optional]`)
	a := flag.String(`append`, ``, `append to an existing readme [optional]`)
	flag.Parse()

	if *a != `` {
		return *mf, *a, *u, *s, domainList(*dl)
	}
	return *mf, entity.DefaultFileName, *u, *s, domainList(*dl)
}

func domainList(input string) (list []string) {
	domains := strings.Split(input, `,`)
	if len(domains) == 0 {
		return nil
	}

	for _, ele := range domains {
		list = append(list, ele)
	}
	return
}
