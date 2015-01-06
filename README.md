go-web-crawler
==============

A toy web crawler written in Go.

Requirements
============

3rd party dependencies are not currently vendored. The following packages are
required to build this application. You should be able to "go get" these:

[Gorp] (https://github.com/coopernurse/gorp)

[Sqlite3] (https://github.com/mattn/go-sqlite3)

[goquery] (https://github.com/PuerkitoBio/goquery)

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
