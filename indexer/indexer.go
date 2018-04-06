package indexer

import (
	"log"
	"sync"

	"github.com/blockstack/blockstack.go/blockstack"
)

const (
	namePageSize    = 100
	namesCollection = "allNames"
)

// NewIndexer creates a new Indexer
func NewIndexer(cfg *Config, names []string) *Indexer {
	return &Indexer{
		BSK:  blockstack.NewClient(cfg.BSK.Host),
		DB:   NewMongoDB(cfg),
		Conc: cfg.IDX.Concurrency,
		ST:   NewStats(),

		// zonefileHashChan: make(chan map[string]string, 0),

		names: names,
	}
}

func (idx *Indexer) reset() {
	idx.names = []string{}
	idx.ST = NewStats()
}

// Indexer is the main stuct for this application
type Indexer struct {
	BSK  *blockstack.Client
	DB   DB
	ST   *Stats
	Conc int

	names []string

	zonefileHashChan chan map[string]string

	sync.Mutex
}

// Index is the main operation
func (idx *Indexer) Index() {
	// Do initial sync of names
	log.Println("[indexer] [initial] fetching names...")
	idx.GetNames()
	log.Println("[indexer] [initial] fetching zonefiles...")
	idx.GetZonefiles()
	log.Println("[indexer] [initial] fetching round finished...")

	// Kick off that loop in the background to keep zonefiles updated
	go nameIndexLoop(idx)
}

func nameIndexLoop(idx *Indexer) {
	for {
		idx.reset()
		log.Println("[indexer] fetching names...")
		idx.GetNames()
		log.Println("[indexer] fetching zonefiles...")
		idx.GetZonefiles()
		log.Println("[indexer] fetching round finished...")
	}
}
