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
	"sync"
)

const (
	termSignal = `terminate`
)

var (
	depChan chan dependency
	urlList []string
)

func parse(fileName string) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	wg := &sync.WaitGroup{}

	for scanner.Scan() {
		text := scanner.Text()
		wg.Add(1)
		go func(wg *sync.WaitGroup, text string) {
			defer wg.Done()
			// skip if it is a blank line
			if len(text) == 0 {
				return
			}

			words := strings.Split(text, ` `)
			if len(words) < 2 {
				return
			}

			if words[0] == `module` {
				return
			}

			// add go version to table later
			if words[0] == `go` {
				return
			}

			dep, ok := singleRequire(words)
			if ok {
				depChan <- *dep
				return
			}

			dep, ok = repoURL(words)
			if ok {
				depChan <- *dep
				return
			}
		}(wg, text)
	}

	// send term signal to the channel
	wg.Wait()
	depChan <- dependency{url: termSignal}
}

func singleRequire(words []string) (dep *dependency, ok bool) {
	// includes indirect comment
	if len(words) != 3 && len(words) != 5 {
		return nil, false
	}

	if words[0] != `require` {
		return nil, false
	}

	if words[1] == `(` {
		return nil, false
	}

	return buildDependency(words[1], words[2]), true
}

func repoURL(words []string) (dep *dependency, ok bool) {
	if len(words) != 2 && len(words) != 4 {
		return nil, false
	}

	terms := strings.Split(words[0], `/`)
	if len(terms) == 0 {
		return nil, false
	}

	for _, url := range urlList {
		if terms[0] == url {
			return buildDependency(words[1], words[2]), true
		}
	}

	return nil, false
}

func buildDependency(path, version string) *dependency {
	desc, err := description(path)
	if err != nil {
		fmt.Printf("get description failed - %s\n", err)
	}

	return &dependency{
		name:    depName(path),
		url:     `https://` + path,
		version: depVersion(version),
		desc:    desc,
	}
}

func depName(path string) string {
	terms := strings.Split(path, `/`)
	return terms[len(terms)-1]
}

func depVersion(text string) string {
	return strings.Split(text, `-`)[0]
}

func description(path string) (desc string, err error) {
	// todo use https://api.github.com/repos/golang-jwt/jwt
	res, err := http.Get(`https://` + path)
	if err != nil {
		return ``, fmt.Errorf(`get request to repo failed - %v`, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ``, fmt.Errorf("failed status code: %d", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ``, fmt.Errorf(`reading response body failed - %v`, err)
	}

	var gr gitRes
	err = json.Unmarshal(data, &gr)
	if err != nil {
		return ``, fmt.Errorf(`unmarshal error - %v [%v]`, err, gr)
	}

	return gr.Description, nil
}

func dependencyList() (deps []dependency) {
	for dep := range depChan {
		if dep.url == termSignal {
			return deps
		}
		deps = append(deps, dep)
	}
	return nil
}
