package indexer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

// GetNames fetches all the names from the blockstack network and stores them in MongoDB as a cache
// TODO: Fix when you are done
func (idx *Indexer) GetNames() {
	if len(idx.names) < 100 {
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

		// Range over the namespaces fetching the names
		for _, ns := range nsInfo.Namespaces() {

			// Record stats for debugging
			idx.ST.Rec("namePages/queued", nsInfo.Pages(ns))

			// Create the namePages in a goroutine
			for page := 0; page < nsInfo.Pages(ns); page++ {
				sem <- struct{}{}
				go idx.namePage(ns, page, namesChan, &wg, sem)
			}
		}

		// Wait for all name pages to return before updating the database
		wg.Wait()
		close(namesChan)
		out, err := json.Marshal(idx.names)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile("names.json", out, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func (idx *Indexer) namePage(ns string, page int, namesChan chan []string, wg *sync.WaitGroup, sem chan struct{}) {
	// Increment the WaitGroup
	wg.Add(1)

	// Record this
	idx.ST.Rec("namePages/created", 1)

	// Fetch the page of names
	names, err := idx.BSK.GetNamesInNamespace(ns, page*namePageSize, namePageSize)
	if err != nil {
		panic(err)
	}

	// Record this
	idx.ST.Rec(fmt.Sprintf("namespace/%s/fetched", ns), len(names.Names))
	idx.ST.Rec("namePages/finished", 1)

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
