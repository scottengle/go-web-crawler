package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const helpText = `
go-web-crawler

go-web-crawler is a simple web crawler that will index URL's contained on pages
fetched over http (or https) using a worker queue and a configurable number of 
concurrent workers. It will ignore URLs that do not have domain names
(IP addresses) and will only index URLs with .com, .org, .net, .edu or .us
suffixes. It will also obey robots.txt directives per the Google spec:
https://developers.google.com/webmasters/control-crawl-index/docs/robots_txt

Usage: %s [options]

Options:
   -help                    show this help page
   -start                   the starting URL to index
   -max-workers=10          maximum number of goroutines to run concurrently
   -max-queue-size=100      maximum number of queued page requests
`

var PageQueue chan PageRequest

func main() {
	// show help text if -help flag is specified
	if len(os.Args) == 2 && os.Args[1] == "-help" {
		fmt.Printf(helpText, os.Args[0])
		os.Exit(0)
	}

	for {
		if code := start(); code >= 0 {
			os.Exit(code)
		}
	}
}

func start() int {

	start := flag.String("start", "", "Start indexing on this URL")
	maxWorkers := flag.Int("max-workers", 10, "Maximum number of indexing workers")
	maxQueue := flag.Int("max-queue-size", 100, "Maximum number of queued page requests")
	flag.Parse()

	if *start == "" {
		panic("[go-web-crawler] No starting URL!")
	}

	if *maxQueue < 1 {
		log.Println("[go-web-crawler] Invalid max-queue-size. Reverting to default of 100.")
		*maxQueue = 100
	}

	if *maxWorkers < 1 {
		log.Println("[go-web-crawler] Invalid max-workers. Reverting to default of 10.")
	}

	log.Printf("[go-web-crawler] Started at page %s with %d indexers and with max size page request queue of %d", *start, *maxWorkers, *maxQueue)

	PageQueue = make(chan PageRequest, *maxQueue)

	StartDispatcher(*maxWorkers)

	startPage := PageRequest{Href: "http://www.pearson.com"}
	PageQueue <- startPage

	select {
	case <-time.After(10 * time.Second):
		return 1
	}
}
