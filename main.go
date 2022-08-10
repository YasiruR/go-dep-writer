package main

var fileName string

func main() {
	//fileName = `readme.md`
	//generateTable([]dependency{{
	//	name:    `[google-wire](https://github.com/google/wire)`,
	//	version: `v0.5.0`,
	//	desc:    `dependency injection library`,
	//},
	//	{
	//		name:    `[uber-zap](https://github.com/uber-go/zap)`,
	//		version: `v1.0`,
	//		desc:    `test`,
	//	}})

	fetchDeps(`go.mod`)
}
