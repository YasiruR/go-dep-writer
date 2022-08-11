package main

import (
	"encoding/base64"
	"flag"
)

func main() {
	parseArgs()
	initReader()
	go parseModFile(modFile)
	deps := dependencyList()
	generateTable(deps)
}

func parseArgs() {
	mf := flag.String(`modfile`, `go.mod`, `relative file path of the go.mod file`)
	of := flag.String(`output`, `dependencies.md`, `relative file path of the output`)
	u := flag.String(`user`, ``, `username of github account [optional]`)
	s := flag.String(`secret`, ``, `secret of github api [optional]`)
	flag.Parse()

	modFile, outputFile = *mf, *of
	if *u != `` && *s != `` {
		token = base64.StdEncoding.EncodeToString([]byte(*u + `:` + *s))
	}
}
