package mod

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/YasiruR/go-dep-writer/entity"
	"github.com/tryfix/log"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// todo append to existing readme

type Parser struct {
	token      string
	depChan    chan entity.Dependency
	domainList []string
	client     *http.Client
	logger     log.Logger
}

func NewParser(user, pw string, domains []string, logger log.Logger) *Parser {
	var token string
	if user != `` && pw != `` {
		token = base64.StdEncoding.EncodeToString([]byte(user + `:` + pw))
	}

	if domains == nil {
		domains = []string{github, uber, goPkg, golang}
	}

	return &Parser{
		token:      token,
		depChan:    make(chan entity.Dependency, 50),
		domainList: domains,
		client:     &http.Client{},
		logger:     logger,
	}
}

func (p *Parser) Parse(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		p.logger.Fatal(fmt.Sprintf(`opening go mod file failed - %v`, err))
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

			dep, ok := p.singleRequire(words)
			if ok {
				p.depChan <- *dep
				return
			}

			dep, ok = p.repoURL(words)
			if ok {
				p.depChan <- *dep
				return
			}
		}(wg, text)
	}

	// send term signal to the channel
	wg.Wait()
	p.depChan <- entity.Dependency{URL: termSignal}
}

func (p *Parser) singleRequire(words []string) (dep *entity.Dependency, ok bool) {
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

	return p.buildDependency(words[1], words[2]), true
}

func (p *Parser) repoURL(words []string) (dep *entity.Dependency, ok bool) {
	if len(words) != 2 && len(words) != 4 {
		return nil, false
	}

	terms := strings.Split(words[0], `/`)
	if len(terms) == 0 {
		return nil, false
	}

	for _, domain := range p.domainList {
		trimmed := strings.TrimSpace(terms[0])
		if trimmed == domain {
			return p.buildDependency(strings.TrimSpace(words[0]), words[1]), true
		}
	}

	return nil, false
}

func (p *Parser) buildDependency(path, version string) *entity.Dependency {
	// checks if path contains a version in the last component
	terms := strings.Split(path, `/`)
	matched, err := regexp.MatchString("v[0-9]", terms[len(terms)-1])
	if err != nil {
		p.logger.Error(fmt.Sprintf(`regex failed - %v`, err))
	}

	if matched {
		path = ``
		for i, term := range terms {
			if i == len(terms)-2 {
				path += term
				break
			}
			path += term + `/`
		}
	}

	desc, err := p.description(path)
	if err != nil {
		p.logger.Error(fmt.Sprintf("get description failed - %s", err))
	}

	return &entity.Dependency{
		Name:    p.depName(path),
		URL:     `https://` + path,
		Version: p.depVersion(version),
		Desc:    desc,
	}
}

func (p *Parser) depName(path string) string {
	terms := strings.Split(path, `/`)
	return terms[len(terms)-1]
}

func (p *Parser) depVersion(text string) string {
	if strings.Contains(text, `+`) {
		return strings.Split(text, `+`)[0]
	}
	return strings.Split(text, `-`)[0]
}

func (p *Parser) description(path string) (desc string, err error) {
	terms := strings.Split(path, `/`)
	if len(terms) == 0 {
		return ``, fmt.Errorf(`empty path`)
	}

	switch terms[0] {
	case github:
		return p.extractDescGithub(`https://api.github.com/repos/` + terms[len(terms)-2] + `/` + terms[len(terms)-1])
	case uber:
		return p.extractDescGoPkg(`https://` + path)
	default:
		return ``, fmt.Errorf(`path [%s] not supported for desc`, path)
	}
}

func (p *Parser) extractDescGithub(url string) (desc string, err error) {
	var res *http.Response
	if p.token == `` {
		res, err = p.client.Get(url)
		if err != nil {
			return ``, fmt.Errorf(`get request to github failed - %v`, err)
		}
	} else {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return ``, fmt.Errorf(`creating new request failed - %v`, err)
		}
		req.Header.Add(`Authorization`, `Basic `+p.token)

		res, err = p.client.Do(req)
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

	var gr entity.GithubResponse
	err = json.Unmarshal(data, &gr)
	if err != nil {
		return ``, fmt.Errorf(`unmarshal error - %v [%v]`, err, gr)
	}

	return gr.Description, nil
}

func (p *Parser) extractDescGoPkg(url string) (desc string, err error) {
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
							return p.description(strings.ReplaceAll(term, `https://`, ``))
						}
					}
				}
			}
		}
	}
}

func (p *Parser) DependencyList() (deps []entity.Dependency) {
	for dep := range p.depChan {
		if dep.URL == termSignal {
			return deps
		}
		deps = append(deps, dep)
	}
	return nil
}
