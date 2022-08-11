package entity

const DefaultFileName = `dependencies.md`

type Dependency struct {
	Name    string
	URL     string
	Version string
	Desc    string
}
