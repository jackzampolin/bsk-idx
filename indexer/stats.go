package indexer

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Stats handles the stats about names fetched
type Stats struct {
	Stats map[string]int `json:"stats"`

	statsChan chan map[string]int

	sync.Mutex
}

func NewStats() *Stats {
	st := &Stats{
		Stats:     make(map[string]int, 0),
		statsChan: make(chan map[string]int, 0),
	}
	go st.handleStatsChan()
	return st
}

// Rec sends a map[string]int to be reorded in the stats
func (stats *Stats) Rec(k string, v int) {
	rec := map[string]int{k: v}
	stats.statsChan <- rec
}

// Stats returns the JSON Marshaled stats
func (stats *Stats) statsRoute() []byte {
	stats.Lock()
	byt, err := json.Marshal(stats.Stats)
	if err != nil {
		log.Fatal(err)
	}
	stats.Unlock()
	return byt
}

// HandleStats is the handler for the stats route
func (stats *Stats) HandleStats(w http.ResponseWriter, r *http.Request) {
	w.Write(stats.statsRoute())
	return
}

// handleStatsChan handles the stats goroutine
func (stats *Stats) handleStatsChan() {
	for st := range stats.statsChan {
		for k, v := range st {
			stats.Stats[k] += v
		}
	}
}
