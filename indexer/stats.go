package indexer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Stats handles the stats about names fetched
type Stats struct {
	Stats map[string]int `json:"stats"`
	Port  int

	statsChan chan map[string]int

	sync.Mutex
}

// NewStats returns a copy of the stats struct
func NewStats(port int) *Stats {
	st := &Stats{
		Stats:     make(map[string]int, 0),
		Port:      port,
		statsChan: make(chan map[string]int, 0),
	}
	go st.handleStatsChan()
	go st.listen()
	return st
}

func (stats *Stats) listen() {
	http.HandleFunc("/idxstats", stats.handleStats)
	log.Printf("[stats_server] Listening for signals on port :%d", stats.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", stats.Port), nil))
}

func (stats *Stats) reset() {
	stats.Lock()
	stats.Stats = make(map[string]int, 0)
	stats.Unlock()
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

// handleStats is the handler for the stats route
func (stats *Stats) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Write(stats.statsRoute())
	return
}

// handleStatsChan handles the stats goroutine
func (stats *Stats) handleStatsChan() {
	for st := range stats.statsChan {
		for k, v := range st {
			stats.Lock()
			stats.Stats[k] += v
			stats.Unlock()
		}
	}
}
