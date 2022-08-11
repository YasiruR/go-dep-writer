package main

import (
	"flag"
)

func main() {
	parseArgs()
	initReader()
	go parse(modFile)
	deps := dependencyList()

	generateTable(deps)
}

func parseArgs() {
	mf := flag.String(`modfile`, `go.mod`, `relative file path of the go.mod file`)
	of := flag.String(`output`, `dependencies.md`, `relative file path of the output`)
	flag.Parse()
	modFile, outputFile = *mf, *of
}
