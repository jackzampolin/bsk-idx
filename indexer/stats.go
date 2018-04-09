package indexer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Status returns the indexer status
type Status struct {
	NamesIndex     string `json:"namesIndex"`
	ZonefilesIndex string `json:"zonefilesIndex"`
	ProfilesIndex  string `json:"profilesIndex"`

	sync.Mutex
}

func newStatus() *Status {
	return &Status{
		NamesIndex:     "indexing",
		ZonefilesIndex: "indexing",
		ProfilesIndex:  "indexing",
	}
}

// Stats handles the stats about names fetched
type Stats struct {
	Namespaces  map[string]int `json:"namespaces"`
	NameFetch   map[string]int `json:"nameFetch"`
	NameDetails map[string]int `json:"nameDetails"`
	Zonefiles   map[string]int `json:"zonefiles"`
	Profiles    map[string]int `json:"profiles"`
	Status      *Status        `json:"status"`

	// Stats map[string]int `json:"stats"`
	Port int `json:"-"`

	statsChan  chan map[string]int
	statusChan chan string

	sync.Mutex
}

// NewStats returns a copy of the stats struct
func NewStats(port int) *Stats {
	st := &Stats{
		Namespaces:  make(map[string]int, 0),
		NameFetch:   make(map[string]int, 0),
		NameDetails: make(map[string]int, 0),
		Zonefiles:   make(map[string]int, 0),
		Profiles:    make(map[string]int, 0),
		Status:      newStatus(),
		Port:        port,
		statsChan:   make(chan map[string]int, 0),
		statusChan:  make(chan string, 0),
	}
	go st.handleStatsChan()
	go st.handleStatusChan()
	go st.listen()
	return st
}

func (stats *Stats) listen() {
	http.HandleFunc("/idxstats", stats.handleStats)
	http.HandleFunc("/status", stats.handleStatus)
	log.Printf("[stats_server] Listening for signals on port :%d", stats.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", stats.Port), nil))
}

// func (stats *Stats) reset() {
// 	stats.Lock()
// 	stats.Stats = make(map[string]int, 0)
// 	stats.Unlock()
// }

// Rec sends a map[string]int to be reorded in the stats
func (stats *Stats) Rec(k string, v int) {
	rec := map[string]int{k: v}
	stats.statsChan <- rec
}

// UpdateStatus updates the indexer status struct
func (stats *Stats) UpdateStatus(st string) {
	stats.statusChan <- st
}

// Stats returns the JSON Marshaled stats
func (stats *Stats) statsRoute() []byte {
	stats.Lock()
	byt, err := json.Marshal(stats)
	if err != nil {
		log.Fatal(err)
	}
	stats.Unlock()
	return byt
}

func (stats *Stats) statusRoute() []byte {
	stats.Lock()
	byt, err := json.Marshal(stats.Status)
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

func (stats *Stats) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Write(stats.statusRoute())
	return
}

// handleStatsChan handles the stats goroutine
func (stats *Stats) handleStatsChan() {
	for st := range stats.statsChan {
		for k, v := range st {
			stats.Lock()
			path := strings.Split(k, ".")
			switch path[0] {
			case "profiles":
				stats.Profiles[path[1]] += v
			case "zonefiles":
				stats.Zonefiles[path[1]] += v
			case "nameFetch":
				stats.NameFetch[path[1]] += v
			case "namespaces":
				stats.Namespaces[path[1]] = v
			case "nameDetails":
				stats.NameDetails[path[1]] += v
			default:
				log.Println("[stats], failed to record stat", k, v)
			}
			stats.Unlock()
		}
	}
}

// handleStatusChan handles the status channel
func (stats *Stats) handleStatusChan() {
	for st := range stats.statusChan {
		statusUpdate := strings.Split(st, ".")
		stats.Status.Lock()
		switch statusUpdate[0] {
		case "names":
			stats.Status.NamesIndex = statusUpdate[1]
		case "zonefiles":
			stats.Status.ZonefilesIndex = statusUpdate[1]
		case "profiles":
			stats.Status.ProfilesIndex = statusUpdate[1]
		}
		stats.Status.Unlock()
	}
}
