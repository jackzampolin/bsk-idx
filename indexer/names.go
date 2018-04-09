package indexer

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

// WriteNamesToFile writes the names on the Indexer into a file
func (idx *Indexer) WriteNamesToFile(file string) {
	out, err := json.Marshal(idx.names)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(file, out, 0644)
	if err != nil {
		panic(err)
	}
}

// GetAllNames fetches all the names from the blockstack network and stores them on the Indexer
func (idx *Indexer) GetAllNames() {
	if len(idx.names) < 10 {
		// fetch Namespace info first then the names from the namespace
		nsInfo, err := idx.GetNSInfo()
		if err != nil {
			panic(err)
		}

		// Create concurrency control
		namesChan := make(chan []string, 0)
		var wg sync.WaitGroup
		sem := make(chan struct{}, idx.Conc)
		go idx.handleNameChan(namesChan)

		// Range over the namespaces fetching the names in seperate goroutines
		for _, ns := range nsInfo.Namespaces() {
			for page := 0; page < nsInfo.Pages(ns); page++ {
				sem <- struct{}{}
				go idx.namePage(ns, page, namesChan, &wg, sem)
			}
		}

		// Wait for all name pages to return before updating the database
		wg.Wait()
		close(namesChan)
	}
}

func (idx *Indexer) namePage(ns string, page int, namesChan chan []string, wg *sync.WaitGroup, sem chan struct{}) {
	// Increment the WaitGroup
	wg.Add(1)

	// Fetch the page of names
	names, err := idx.GetNamesInNamespace(ns, page*namePageSize, namePageSize)
	if err != nil {
		// NOTE: The above call is retried
		panic(err)
	}

	// Send the names to get processed
	namesChan <- names.Names

	// Read from semaphore chan
	<-sem

	// Decrement the waitgroup
	wg.Done()
}

func (idx *Indexer) handleNameChan(namesChan chan []string) {
	for name := range namesChan {
		for _, n := range name {
			idx.names = append(idx.names, n)
		}
	}
}
