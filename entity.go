package main

type dependency struct {
	name    string
	url     string
	version string
	desc    string
}

type gitRes struct {
	Description string `json:"description"`
	URL         string `json:"url"`
}
