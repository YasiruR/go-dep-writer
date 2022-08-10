package main

var fileName string

func main() {
	fileName = `readme.md`
	generateTable([]row{{
		dep:     `[google-wire](https://github.com/google/wire)`,
		version: `v0.5.0`,
		desc:    `dependency injection library`,
	}})
}
