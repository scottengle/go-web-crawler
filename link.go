package main

import ()

// Link represents a link between two pages
type Link struct {
	Parent string `json:"parent"`
	ID     string `json:"id"`
	URL    string `json:"url"`
}

// NewLink returns a new pointer to a Link
func NewLink(parent, id, url string) Link {
	return Link{parent, id, url}
}
