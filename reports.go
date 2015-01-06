package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ReportItem is used by all reports
type ReportItem struct {
	URL       string `json:"url"`
	ParentURL string `json:"parent_url"`
}

// GenerateInboundLinksReport creates a report showing the inbound link frequency for crawled pages
func GenerateInboundLinksReport(asJSON bool) {

	container := make(map[string][]string)

	items, err := GetReportData()
	if err != nil {
		Logger.checkErr(err, "Couldn't generate report")
		return
	}

	for _, item := range items {
		links := AppendIfMissing(container[item.URL], item.ParentURL)
		container[item.URL] = links
	}

	OutputData(container, "Page\tPages Linking To Page", asJSON)
}

// GenerateParentChildLinksReport creates a report showing the connections between
// pages and the links they contain
func GenerateParentChildLinksReport(asJSON bool) {

	container := make(map[string][]string)

	items, err := GetReportData()
	if err != nil {
		Logger.checkErr(err, "Couldn't generate report")
		return
	}

	for _, item := range items {
		links := AppendIfMissing(container[item.ParentURL], item.URL)
		container[item.ParentURL] = links
	}

	OutputData(container, "Page\tLinks On Page", asJSON)
}

// OutputData outputs report data, optionally as a JSON encoded string
func OutputData(data map[string][]string, tabularHeader string, asJSON bool) {

	if asJSON {
		results, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("[go-web-crawler] Report:\n%s", results)
		return
	}

	fmt.Println(tabularHeader)
	for key, val := range data {
		for _, item := range val {
			fmt.Printf("%s\t%s\n", key, item)
		}
	}

}

// GetReportData gets the report data from the last crawler run
func GetReportData() ([]ReportItem, error) {
	dbmap := connect()
	defer disconnect(dbmap)

	var items []ReportItem

	tx, _ := dbmap.Begin()
	_, err := tx.Select(&items, "SELECT URL, ParentURL from links ORDER BY ParentURL")
	tx.Commit()

	if err != nil {
		return nil, err
	}

	return items, nil
}

// AppendIfMissing adds an element to the slice if the element is not already present
func AppendIfMissing(slice []string, item string) []string {
	for _, ele := range slice {
		if ele == item {
			return slice
		}
	}
	slice = append(slice, item)
	sort.Strings(slice)

	return slice
}
