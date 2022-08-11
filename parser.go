package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	termSignal = `terminate`
	github     = `github.com` // todo env var
	uber       = `go.uber.org`
	goPkg      = `gopkg.in`
	golang     = `golang.org`
)

var (
	modFile string
	token   string
	depChan chan dependency
	urlList []string
	client  *http.Client
)

func initReader() {
	depChan = make(chan dependency, 50)
	urlList = []string{github, uber, goPkg, golang}
	client = &http.Client{}
}

func parseModFile(fileName string) {
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
		trimmed := strings.TrimSpace(terms[0])
		if trimmed == url {
			return buildDependency(strings.TrimSpace(words[0]), words[1]), true
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
	if strings.Contains(text, `+`) {
		return strings.Split(text, `+`)[0]
	}
	return strings.Split(text, `-`)[0]
}

func description(path string) (desc string, err error) {
	terms := strings.Split(path, `/`)
	if len(terms) == 0 {
		return ``, fmt.Errorf(`empty path`)
	}

	switch terms[0] {
	case github:
		return extractDescGithub(`https://api.github.com/repos/` + terms[len(terms)-2] + `/` + terms[len(terms)-1])
	case uber:
		return extractDescGoPkg(`https://` + path)
	default:
		return ``, fmt.Errorf(`path [%s] not supported for desc`, path)
	}
}

func extractDescGithub(url string) (desc string, err error) {
	var res *http.Response
	if token == `` {
		res, err = client.Get(url)
		if err != nil {
			return ``, fmt.Errorf(`get request to github failed - %v`, err)
		}
	} else {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return ``, fmt.Errorf(`creating new request failed - %v`, err)
		}
		req.Header.Add(`Authorization`, `Basic `+token)

		res, err = client.Do(req)
		if err != nil {
			return ``, fmt.Errorf(`get request with auth to github failed - %v`, err)
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ``, fmt.Errorf("status code: %d", res.StatusCode)
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

func extractDescGoPkg(url string) (desc string, err error) {
	res, err := http.Get(url)
	if err != nil {
		return ``, fmt.Errorf(`get request to go pkg failed - %v`, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ``, fmt.Errorf(`status code: %d`, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ``, fmt.Errorf(`reading body failed`)
	}

	reader := strings.NewReader(string(body))
	tokenizer := html.NewTokenizer(reader)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return
			}
			return ``, fmt.Errorf(`tokenizer error - %v`, tokenizer.Err())
		}

		_, hasAttr := tokenizer.TagName()
		if hasAttr {
			for {
				attrKey, attrValue, _ := tokenizer.TagAttr()
				if string(attrKey) == `content` {
					terms := strings.Split(string(attrValue), ` `)
					for _, term := range terms {
						if strings.Contains(term, `https://`+github) {
							return description(strings.ReplaceAll(term, `https://`, ``))
						}
					}
				}
			}
		}
	}
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
