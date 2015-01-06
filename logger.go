package main

import "log"

// Log is a global logger object
type Log struct {
	verboseOutput bool
}

func NewLog(verbose bool) *Log {
	return &Log{verboseOutput: verbose}
}

func (l *Log) Log(msg string) {
	if l.verboseOutput {
		log.Println(msg)
	}
}

func (l *Log) Logf(msg string, args ...interface{}) {
	if l.verboseOutput {
		log.Printf(msg, args...)
	}
}

func (l *Log) checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
