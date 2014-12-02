package main

import (
	"log"
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
			case page := <-idxr.Page:
				log.Printf("[Indexer %d]: Receiveds request for page %s\n", idxr.ID, page.Href)

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
