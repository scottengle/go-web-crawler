package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const helpText = `
go-web-crawler

go-web-crawler is a simple web crawler that will index URL's contained on pages
fetched over http (or https). It will ignore URLs that do not have domain names
(IP addresses) and will only index URLs with .com, .org, .net, .edu or .us
suffixes. It will also obey robots.txt directives per the Google spec:
https://developers.google.com/webmasters/control-crawl-index/docs/robots_txt

Usage: %s [options]

Options:
   -help               show this help page
   -start              the starting URL to index
   -max-workers=10     maximum number of goroutines to run concurrently
`

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
	flag.Parse()

	if *start == "" {
		panic("No starting URL!")
	}

	fmt.Printf("Started go-web-crawler at %s with %d max workers.", *start, *maxWorkers)

	select {
	case <-time.After(5 * time.Second):
		return 1
	}
}
