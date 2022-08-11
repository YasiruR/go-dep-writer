# Go dependency publisher

**go-dep-writer** reads dependencies in go.mod file of a project and writes to
a markdown file in the HTML format.

## Usage

1. `git clone https://github.com/YasiruR/go-dep-writer.git`
2. `go build`
3. `./go-dep-writer -modfile=<go.mod file path>`

## Parameters

`./go-dep-writer [-PARAMS]`

- **modfile** 
  - File path of go.mod
- **user** [optional]
  - Username of the github account. Both user and secret should be provided to
utilize github authentication and access API endpoints. If else, API endpoints are
accessed without authentication and being constrained under the limitation of 
a quota.
- **secret** [optional]
  - Access token to utilize github authentication as described above.
- **append** [optional]
  - If provided, generated table will be appended to the given file. If not provided,
table will be saved to dependencies.md file in the current working directory.
- **domains** [optional]
  - To save imports with specific domains. Only github.com, go.uber.org,
gopkg.in and golang.org are supported (and all enabled by default).

## Sample Output

<table>
<thead>
<tr>
<th>Dependency</th>
<th>Version</th>
<th>Description</th>
</tr>
</thead>

<tbody>
<tr>
<td><a href="https://github.com/caarlos0/env">env</a></td>
<td>v6.9.1</td>
<td>A simple and zero-dependencies library to parse environment variables into structs.</td>
</tr>

<tr>
<td><a href="https://github.com/rs/zerolog">zerolog</a></td>
<td>v1.22.0</td>
<td>Zero Allocation JSON Logger</td>
</tr>

<tr>
<td><a href="https://github.com/logrusorgru/aurora">aurora</a></td>
<td>v0.0.0</td>
<td>Golang ultimate ANSI-colors that supports Printf/Sprintf methods</td>
</tr>

<tr>
<td><a href="https://github.com/gomarkdown/markdown">markdown</a></td>
<td>v0.0.0</td>
<td>markdown parser and HTML renderer for Go</td>
</tr>
</tbody>
</table>
