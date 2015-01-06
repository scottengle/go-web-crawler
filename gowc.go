package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
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
   -report=1            	generate the "Inbound Link Frequency" report
   -report=2            	generate the "Parent/Child Links" report
   -report-format			output report results as either "json" or "tabular"
`

// PageQueue is a channel to queue up page requests
var PageQueue chan PageRequest

// QuitChannel is a signal channel used to stop execution
var QuitChannel chan struct{}

// Lock is a mutex used to protect the database
var Lock = &sync.Mutex{}

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
	maxQueue := flag.Int("max-queue-size", 100, "Maximum number of queued page requests allowed")
	maxSeconds := flag.Int("max-runtime-seconds", 10, "Maximum number of seconds to run the crawler")
	report := flag.Int("report", 0, "Generate a report without crawling")
	format := flag.String("report-format", "json", "Format of the report. Value is ignored if not running reports.")
	flag.Parse()

	if *start == "" && *report == 0 {
		log.Println("[go-web-crawler] No starting URL specified and invalid report.")
		log.Println("[go-web-crawler] Specify a starting URL with -start, or request a report with -report=1 or -report=2.")
		return 0
	}

	if *start == "" && *report < 1 || *report > 2 {
		log.Println("[go-web-crawler] Report requested does not exist. Valid report ID's are: [1, 2]")
		return 0
	}

	if *format != "json" && *format != "tabular" {
		log.Println("[go-web-crawler] Report format requested is not supported. Valid report-format strings are: ['json', 'tabular']")
		return 0
	}

	if *report == 1 {
		log.Println("[go-web-crawler] Running Inbound Link Counts Report")
		GenerateInboundLinksReport(reportFormatIsJSON(*format))
		return 0
	}

	if *report == 2 {
		log.Println("[go-web-crawler] Running Parent/Child Link Report")
		GenerateParentChildLinksReport(reportFormatIsJSON(*format))
		return 0
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

	initDb()

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

func initDb() {
	dbmap := connect()

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via script
	err := dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")

	dbmap.TruncateTables()
	disconnect(dbmap)
}

func connect() *gorp.DbMap {
	Lock.Lock()

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

func disconnect(dbmap *gorp.DbMap) {
	dbmap.Db.Close()
	Lock.Unlock()
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

// GetMD5Hash converts the specified text to an MD5 hash
func GetMD5Hash(text string) string {

	hasher := md5.New()

	hasher.Write([]byte(text))

	return hex.EncodeToString(hasher.Sum(nil))

}

func reportFormatIsJSON(format string) bool {
	return format == "json"
}
