package main

import (
	"log"
)

// WorkerQueue is a channel to hold the PageRequest channel for indexers
var WorkerQueue chan chan PageRequest

// StartDispatcher allocates, initializes and starts the specified number of workers
func StartDispatcher(numWorkers int) {
	WorkerQueue = make(chan chan PageRequest, numWorkers)

	for i := 0; i < numWorkers; i++ {
		indexer := NewIndexer(i+1, WorkerQueue)
		log.Printf("[Dispatcher] Starting Indexer %d", indexer.ID)
		indexer.Start()
	}

	go func() {
		for {
			select {
			case pageRequest := <-PageQueue:
				log.Printf("[Dispatcher] Received page request for %s\n", pageRequest)
				go func(pageRequest PageRequest) {
					log.Printf("[Dispatcher] Dispatching page request for %s", pageRequest.Href)
				}(pageRequest)
			}
		}
	}()
}
