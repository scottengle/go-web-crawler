go-web-crawler
==============

A toy web crawler written in Go for a "Dev Day" over the course of about 12 hours.

Requirements
============

3rd party dependencies have not been vendored yet. The following packages are
required to build. You should be able to "go get" these:

[Gorp] (https://github.com/coopernurse/gorp)

[Sqlite3] (https://github.com/mattn/go-sqlite3)

[go-html-transform] (https://code.google.com/p/go-html-transform)

[go.net/html] (https://code.google.com/p/go.net/html)

[robotstxt-go] (https://github.com/temoto/robotstxt-go)

Usage
=====

Build the project:

    go build

Basic usage with defaults and provided start URL:

    ./go-web-crawler -start http://www.startpage.com

Use -help to get a list of commands:

    ./go-web-crawler -help

Known Issues
============

* There are no known issues at this time

License
=======

[MIT License] (LICENSE)
