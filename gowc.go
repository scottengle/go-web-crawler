package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
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
   -max-workers=1           maximum number of goroutines to run concurrently
   -max-queue-size=100      maximum number of queued page requests
   -max-runtime-seconds=10  maximum number of seconds to run the crawler
   -run-report-1            generate the "Inbound Link Frequency" report
   -run-report-2            generate the "Parent-Child Links" report
`

var PageQueue chan PageRequest
var QuitChannel chan struct{}

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
	maxWorkers := flag.Int("max-workers", 1, "Maximum number of indexing workers")
	maxQueue := flag.Int("max-queue-size", 100, "Maximum number of queued page requests")
	maxSeconds := flag.Int("max-runtime-seconds", 10, "Maximum number of seconds to run the crawler")
	report1 := flag.Bool("run-report-1", false, "Run report 1 without crawling")
	report2 := flag.Bool("run-report-2", false, "Run report 2 without crawling")
	flag.Parse()

	if *report1 {
		log.Println("[go-web-crawler] Running Report 1 - Inbound Link Counts Per Page")
		RunReport1()
		return 0
	}

	if *report2 {
		log.Println("[go-web-crawler] Running Report 2 - Parent/Child Link Connections")
		RunReport2()
		return 0
	}

	if *start == "" {
		panic("[go-web-crawler] No starting URL!")
	}

	if *maxQueue < 1 {
		log.Println("[go-web-crawler] Invalid max-queue-size. Reverting to default of 100.")
		*maxQueue = 100
	}

	if *maxWorkers < 1 {
		log.Println("[go-web-crawler] Invalid max-workers. Reverting to default of 1.")
		*maxWorkers = 10
	}

	if *maxSeconds < 1 {
		log.Println("[go-web-crawler] Invalid max-runtime-seconds. Reverting to default of 10.")
		*maxSeconds = 10
	}

	log.Printf("[go-web-crawler] Started at page %s with %d indexers and with max queue of %d and a running time of %d seconds", *start, *maxWorkers, *maxQueue, *maxSeconds)

	dbmap := initDb()
	defer dbmap.Db.Close()

	PageQueue = make(chan PageRequest, *maxQueue)
	QuitChannel = make(chan struct{})
	StartDispatcher(*maxWorkers, *maxSeconds)

	startPage := PageRequest{Href: *start}
	PageQueue <- startPage

	select {
	case <-QuitChannel:
		log.Println("[go-web-crawler] Quit signal received. Quitting.")
		return 0
	}
}

func initDb() *gorp.DbMap {
	dbmap := connect()

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via script
	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	return dbmap
}

func connect() *gorp.DbMap {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "post_db.bin")
	checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	dbmap.AddTableWithName(Link{}, "links").SetKeys(true, "ID")

	return dbmap
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

type ReportItem struct {
	URL   string `json:"url"`
	Count int    `json:"count"`
}

type ReportItem2 struct {
	URL    string `json:"url"`
	Parent string `json:"parent"`
}

func RunReport1() {
	dbmap := connect()

	var items []ReportItem

	tx, _ := dbmap.Begin()
	_, err := tx.Select(&items, "SELECT URL, COUNT(*) AS Count from links GROUP BY URL ORDER BY Count DESC")
	checkErr(err, "Couldn't generate report")

	results, _ := json.MarshalIndent(items, "", "  ")
	log.Printf("[go-web-crawler] Report:\n%s", results)
	tx.Commit()

	dbmap.Db.Close()

	return
}

func RunReport2() {
	dbmap := connect()

	var items2 []ReportItem2

	tx, _ := dbmap.Begin()
	_, err := tx.Select(&items2, "SELECT URL, Parent from links ORDER BY Parent")
	checkErr(err, "Couldn't generate 2nd report")

	log.Printf("[go-web-crawler] Report 2:\n")
	for _, item2 := range items2 {
		fmt.Printf("Link: %s -> %s\n", item2.Parent, item2.URL)
	}
	tx.Commit()

	dbmap.Db.Close()

	return
}
