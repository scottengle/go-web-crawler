package main

import (
	"time"
)

// Link represents a link between two pages
type Link struct {
	Parent    string `json:"parent"`
	ParentURL string `json:"parent_url"`
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	Created   int64  `json:"-"`
}

// NewLink returns a new pointer to a Link
func NewLink(parentID, parentURL, url string) Link {
	return Link{
		URL:       url,
		Parent:    parentID,
		ParentURL: parentURL,
		Created:   time.Now().UnixNano(),
	}
}
