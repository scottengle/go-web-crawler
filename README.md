go-web-crawler
==============

A toy web crawler written in Go for a "Dev Day" over the course of about 12 hours.

Requirements
============

3rd party dependencies have not been vendored yet. The following packages are
required to build. You should be able to "go get" these:

Gorp: github.com/coopernurse/gorp
Sqlite3: "github.com/mattn/go-sqlite3"
go-html-transform: "code.google.com/p/go-html-transform"
go.net/html: "code.google.com/p/go.net/html"
robotstxt-go: "github.com/temoto/robotstxt-go"

Known Issues
============

* The application works reasonably well when using 1 worker. Concurrent workers can cause issues with database locks.
* Log messages are out of control