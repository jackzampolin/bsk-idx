package indexer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/blockstack/blockstack.go/blockstack"
)

const (
	namePageSize = 100
	idxPrefix    = "[indexer]"
)

// NewIndexer creates a new Indexer
func NewIndexer(cfg *Config, names []string) *Indexer {
	return &Indexer{
		BSK:  blockstack.NewClient(cfg.BSK.Host),
		DB:   NewMongoDB(cfg),
		Conc: cfg.IDX.Concurrency,
		ST:   NewStats(cfg.IDX.StatsPort),

		names: names,

		// TODO: Bring these in through config
		retries: 3,
		timeout: 1 * time.Second,

		config: cfg.IDX,
	}
}

func (idx *Indexer) reset() {
	idx.names = []string{}
	idx.ST.reset()
}

// Indexer is the main stuct for this application
type Indexer struct {
	BSK  *blockstack.Client
	DB   DB
	ST   *Stats
	Conc int

	names []string

	// Number of retries and backoff time for blockstack calls
	retries int
	timeout time.Duration

	config IDXConfig

	sync.Mutex
}

// Index is the main operation
func (idx *Indexer) Index() {

	if _, err := os.Stat(idx.config.NameFile); err == nil {
		names := make([]string, 0)
		nm, err := ioutil.ReadFile(idx.config.NameFile)
		if err != nil {
			idx.log(idxPrefix, "Names file exists but is unreadable, fetching names from network")
			err = json.Unmarshal(nm, &names)
			if err != nil {
				idx.log(idxPrefix, "Names file exists but is unparsable, fetching names from network")
			}
		}
		idx.names = names
	}

	// If the names were not loaded from file we need to do an
	// initial name sync to populate the list of names
	if len(idx.names) < 100 {
		log.Println("[indexer] fetching names...")
		idx.GetAllNames()
		log.Println("[indexer] fetching zonefiles...")
		idx.GetAllZonefiles()
		log.Println("[indexer] fetching round finished...")
	}

	// Kick off that loop in the background to keep zonefiles updated
	go nameIndexLoop(idx)

	//
	for {
		log.Println("[indexer] resolving all names associated with indexer")
		idx.ResolveIndexerNames()
	}
}

func (idx *Indexer) log(prefix, message string) {
	log.Printf("%s %s", prefix, message)
}

func nameIndexLoop(idx *Indexer) {
	for {
		idx.reset()
		log.Println("[indexer] fetching names...")
		idx.GetAllNames()
		log.Println("[indexer] fetching zonefiles...")
		idx.GetAllZonefiles()
		log.Println("[indexer] fetching round finished...")
	}
}
