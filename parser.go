package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var depChan chan dependency

func fetchDeps(fileName string) (deps []dependency) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		// skip if it is a blank line
		if len(text) == 0 {
			continue
		}

		words := strings.Split(text, ` `)
		if len(words) < 2 {
			continue
		}

		if words[0] == `module` {
			continue
		}

		// add go version to table later
		if words[0] == `go` {
			continue
		}

		dep, ok := singleRequire(words)
		if ok {
			depChan <- *dep
			continue
		}

		for _, char := range text {
			fmt.Println(string(char))
		}
	}

	return
}

func singleRequire(words []string) (dep *dependency, ok bool) {
	// includes indirect comment
	if len(words) != 3 || len(words) != 5 {
		return nil, false
	}

	if words[0] != `require` {
		return nil, false
	}

	if words[1] == `(` {
		return nil, false
	}

	desc, err := description(words[1])
	if err != nil {
		fmt.Printf("get description failed - %s\n", err)
	}

	dep = &dependency{
		name:    depName(words[1]),
		url:     `https://` + words[1],
		version: depVersion(words[2]),
		desc:    desc,
	}

	return dep, true
}

func depName(path string) string {
	terms := strings.Split(path, `/`)
	return terms[len(terms)-1]
}

func depVersion(text string) string {
	return strings.Split(text, `-`)[0]
}

func description(path string) (desc string, err error) {
	res, err := http.Get(`https://` + path)
	if err != nil {
		return ``, fmt.Errorf(`get request to repo failed - %v`, err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ``, fmt.Errorf(`reading response body failed - %v`, err)
	}

	var gr gitRes
	err = json.Unmarshal(data, &gr)
	if err != nil {
		return ``, fmt.Errorf(`unmarshal error - %v`, err)
	}

	return gr.Description, nil
}

func addDep() (deps []dependency) {
	for dep := range depChan {
		if dep.url == `terminate` {
			return deps
		}
		deps = append(deps, dep)
	}
	return nil
}
