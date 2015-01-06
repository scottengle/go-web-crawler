package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"code.google.com/p/go-html-transform/h5"
	exphtml "code.google.com/p/go.net/html"
	"github.com/coopernurse/gorp"
	"github.com/temoto/robotstxt-go"
)

// NewIndexer creates a new Indexer worker
func NewIndexer(id int, workerQueue chan chan PageRequest) Indexer {
	indexer := Indexer{
		ID:          id,
		Page:        make(chan PageRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan struct{}),
	}

	return indexer
}

// Indexer is the struct representation of an Indexer
type Indexer struct {
	ID          int
	Page        chan PageRequest
	WorkerQueue chan chan PageRequest
	QuitChan    chan struct{}
}

// Start kicks off an Indexer, which will add itself to the worker queue
// and begin processing requests for page indexes
func (idxr *Indexer) Start() {
	go func() {
		for {
			idxr.WorkerQueue <- idxr.Page

			select {
			case pageRequest := <-idxr.Page:
				log.Printf("[Indexer %d] Received request for page %s\n", idxr.ID, pageRequest.Href)

				idxr.ProcessPage(pageRequest)

			case <-idxr.QuitChan:
				log.Printf("[Indexer %d]: Stopped\n", idxr.ID)
				return
			}
		}
	}()
}

// Stop instructs a page indexer to stop processing page requests
func (idxr *Indexer) Stop() {
	go func() {
		idxr.QuitChan <- struct{}{}
	}()
}

// ProcessPage attempts to index the requested page
func (idxr *Indexer) ProcessPage(pageRequest PageRequest) {
	// attempt to get robots.txt
	u, err := url.Parse(pageRequest.Href)
	if err != nil {

		log.Printf("[Indexer %d] Unable to parse URL: %s\nError: %s", idxr.ID, pageRequest.Href, err.Error())
		return

	}

	rootURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	robotsURL := fmt.Sprintf("%s/robots.txt", rootURL)
	log.Printf("[Indexer %d] Attempting to retrieve robots.txt at %s", idxr.ID, robotsURL)

	authorized, err := idxr.AuthorizeRobot(robotsURL, pageRequest)
	if err != nil {

		// We already printed the error to the output
		return

	}

	if authorized {

		log.Printf("[Indexer %d] Indexing of %s authorized by robots.txt", idxr.ID, pageRequest.Href)

		log.Printf("[Indexer %d] Starting scrape of %s", idxr.ID, pageRequest.Href)

		resp, err := http.Get(pageRequest.Href)

		if err != nil {
			log.Printf("[Indexer %d] Error received while retrieving %s: %s", idxr.ID, pageRequest.Href, err.Error())
			return
		}

		dbmap := connect()
		defer disconnect(dbmap)

		results := idxr.ScrapePage(resp, rootURL, pageRequest.Href, dbmap)
		for _, link := range results {

			result, err := json.Marshal(link)
			if err != nil {
				log.Printf("Indexer %d] Error received during json encoding: %s", idxr.ID, err.Error())
			}

			log.Printf("[Indexer %d] Found result: %s", idxr.ID, result)

			// request processing for the result if it has not been processed
			var processedLinks []Link

			tx, _ := dbmap.Begin()
			_, err = tx.Select(&processedLinks, "SELECT * FROM links WHERE Parent LIKE ?", GetMD5Hash(link.URL))
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}

			results, err := json.Marshal(processedLinks)
			if err == nil {
				log.Printf("[Indexer %d] Found parent %s with data %s", idxr.ID, link.URL, results)
			}

			if len(processedLinks) == 0 {
				newRequest := PageRequest{Href: link.URL}
				PageQueue <- newRequest
			}
		}

	} else {

		log.Printf("[Indexer %d] Indexing of %s not authorized by robots.txt", idxr.ID, pageRequest.Href)
		return

	}
}

// AuthorizeRobot attempts to retrieve the robots.txt file and authorize the robot to crawl the
// page requested.
func (idxr *Indexer) AuthorizeRobot(robotsURL string, pageRequest PageRequest) (bool, error) {

	resp, err := http.Get(robotsURL)
	if err != nil {

		// This situation is "undefined" per the spec, so just return and ignore the URL
		log.Printf("[Indexer %d] Unable to retrieve %s: %s", idxr.ID, robotsURL, err.Error())
		return false, err

	}

	authorizer, err := robotstxt.FromResponse(resp)
	resp.Body.Close()
	if err != nil {

		log.Printf("[Indexer %d] Unable to parse %s: %s", idxr.ID, robotsURL, err.Error())
		return false, err

	}

	log.Printf("[Indexer %d] Received status code %d from %s", idxr.ID, resp.StatusCode, robotsURL)

	return authorizer.TestAgent(pageRequest.Href, "go-web-crawler"), nil
}

// ScrapePage takes a response body and scrapes it for all a tags
func (idxr *Indexer) ScrapePage(resp *http.Response, rootURL, parentURL string, dbmap *gorp.DbMap) []Link {

	tree, _ := h5.New(resp.Body)
	defer resp.Body.Close()

	links := make([]Link, 0)

	tree.Walk(func(node *exphtml.Node) {

		if node.Type == exphtml.ElementNode && node.Data == "a" {

			for _, elem := range node.Attr {
				if elem.Key == "href" && linkIsValid(elem.Val) {

					url := elem.Val

					if strings.HasPrefix(elem.Val, "/") {
						url = fmt.Sprintf("%s%s", rootURL, elem.Val)
					}

					exists := false
					for _, link := range links {
						if url == link.URL {
							exists = true
						}
					}

					if !exists {

						log.Println("Parent: " + parentURL + ", URL: " + url)
						newLink := NewLink(GetMD5Hash(parentURL), parentURL, url)
						existingLink := NewLink(GetMD5Hash(parentURL), parentURL, url)
						log.Printf("Existing Link %v", &existingLink)

						tx, err := dbmap.Begin()

						if err != nil {
							log.Fatalf("[Indexer %d] Transaction was nil. Error: %s. Couldn't insert URL %s with Parent %s", idxr.ID, err.Error(), existingLink.URL, existingLink.Parent)
							continue
						}

						selectErr := tx.SelectOne(&existingLink, "SELECT * FROM links WHERE URL LIKE ? and Parent LIKE ?;", existingLink.URL, existingLink.Parent)
						tx.Commit()

						switch {
						case selectErr == sql.ErrNoRows:
							txInsert, _ := dbmap.Begin()
							insertErr := txInsert.Insert(&newLink)
							checkErr(insertErr, "[Indexer %d] Error inserting link into db")
							if insertErr != nil {
								txInsert.Rollback()
							} else {
								txInsert.Commit()
							}

						case selectErr != nil:
							checkErr(selectErr, "[Indexer %d] Error retrieving link")
						}

						links = append(links, newLink)

					}
				}
			}
		}
	})

	return links
}

// Attempt to eliminate meaningless or invalid URLs
// For now, only accept URLs that start with http
func linkIsValid(link string) bool {

	return strings.HasPrefix(link, "http")

}
