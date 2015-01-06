package main

import (
	"time"
)

// WorkerQueue is a channel to hold the PageRequest channel for indexers
var WorkerQueue chan chan PageRequest
var queueLocked bool

// StartDispatcher allocates, initializes and starts the specified number of workers
func StartDispatcher(numWorkers int, maxSeconds int) {

	WorkerQueue = make(chan chan PageRequest, numWorkers)
	workers := make([]Indexer, 0)

	for i := 0; i < numWorkers; i++ {

		indexer := NewIndexer(i+1, WorkerQueue)

		Logger.Logf("[Dispatcher] Starting Indexer %d", indexer.ID)

		indexer.Start()

		workers = append(workers, indexer)

	}

	go func() {
		for {
			select {
			case pageRequest := <-PageQueue:
				Logger.Logf("[Dispatcher] Received page request for %s\n", pageRequest.Href)

				go func(pageRequest PageRequest) {
					worker := <-WorkerQueue

					Logger.Logf("[Dispatcher] Dispatching page request for %s", pageRequest.Href)

					worker <- pageRequest

				}(pageRequest)
			}
		}
	}()

	go func(maxSeconds int, quit chan struct{}, workers []Indexer) {
		select {
		case <-time.After(time.Duration(maxSeconds) * time.Second):

			for _, worker := range workers {
				worker.Stop()
			}

			quit <- struct{}{}
		}
	}(maxSeconds, QuitChannel, workers)
}
