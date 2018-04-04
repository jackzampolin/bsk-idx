package cmd

import (
	"encoding/json"
	"log"
	"net/http"
)

// Record sends a map[string]int to be reorded in the stats
func (idx *Indexer) Record(k string, v int) {
	rec := map[string]int{k: v}
	idx.st <- rec
}

// Stats returns the JSON Marshaled stats
func (idx *Indexer) statsRoute() []byte {
	idx.Lock()
	byt, err := json.Marshal(idx.stats)
	if err != nil {
		log.Fatal(err)
	}
	idx.Unlock()
	return byt
}

// handleStats is the handler for the stats route
func (idx *Indexer) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Write(idx.statsRoute())
	return
}

// handleStatsChan handles the stats goroutine
func (idx *Indexer) handleStatsChan() {
	for st := range idx.st {
		for k, v := range st {
			idx.stats[k] += v
		}
	}
}
