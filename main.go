package main

import (
	"flag"
	"github.com/YasiruR/go-dep-writer/markdown"
	"github.com/YasiruR/go-dep-writer/mod"
	"github.com/tryfix/log"
)

func main() {
	modFile, outputFile, user, pw := parseArgs()
	logger := log.Constructor.Log(log.WithColors(true), log.WithLevel("DEBUG"), log.WithFilePath(true))

	p := mod.NewParser(user, pw, logger)
	go p.Parse(modFile)
	deps := p.DependencyList()

	w := markdown.NewWriter(logger)
	w.GenerateTable(outputFile, deps)
}

func parseArgs() (modFile, outputFile, user, pw string) {
	mf := flag.String(`modfile`, `go.mod`, `relative file path of the go.mod file`)
	of := flag.String(`output`, `dependencies.md`, `relative file path of the output`)
	u := flag.String(`user`, ``, `username of github account [optional]`)
	s := flag.String(`secret`, ``, `secret of github api [optional]`)
	flag.Parse()
	return *mf, *of, *u, *s
}
