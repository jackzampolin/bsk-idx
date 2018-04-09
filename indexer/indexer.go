package indexer

import (
	"encoding/json"
	"fmt"
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

		names: &networkNames{n: make([]string, 0)},

		// TODO: Bring these in through config
		retries: 3,
		timeout: 1 * time.Second,

		config: cfg.IDX,
	}
}

// Indexer is the main stuct for this application
type Indexer struct {
	BSK  *blockstack.Client
	DB   DB
	ST   *Stats
	Conc int

	names *networkNames

	// Number of retries and backoff time for blockstack calls
	retries int
	timeout time.Duration

	config IDXConfig

	sync.Mutex
}

// Index is the main operation
func (idx *Indexer) Index() {

	// First try to pull names from the names.json file
	if _, err := os.Stat(idx.config.NameFile); err == nil {
		names := make([]string, 0)
		nm, err := ioutil.ReadFile(idx.config.NameFile)
		if err != nil {
			idx.log(idxPrefix, "Names file exists but is unreadable, fetching names from network")
		}
		err = json.Unmarshal(nm, &names)
		if err != nil {
			idx.log(idxPrefix, "Names file exists but is unparsable, fetching names from network")
		}
		idx.log(idxPrefix, "Reading names from file, kicking off update routine...")
		idx.names.n = names
	}

	nsInfo, _ := idx.GetNSInfo()

	// If the names were not loaded from file we need to do an
	// initial name sync to populate the list of names and write them to the file
	if idx.names.length() < nsInfo.Count()-50 {
		idx.log(idxPrefix, "names file not found, fetching names...")
		idx.GetAllNames()
		idx.log(idxPrefix, fmt.Sprintf("names updated, writing names to file %s...", idx.config.NameFile))
		idx.WriteNamesToFile(idx.config.NameFile)
	}
	// Set name index status to available
	idx.ST.UpdateStatus("names.ready")
	idx.log(idxPrefix, "checking zonefiles...")

	// Kick off name update loop in the background to keep names array current
	go idx.nameIndexLoop()

	// If the zonefile database hasn't been populated then populate it
	if idx.DB.ZonefilesCount() < (idx.names.length() * 2 / 3) {
		idx.log(idxPrefix, "zonefiles not populated, fetching...")
		idx.GetAllZonefiles()
	}

	// Set name index status to available
	idx.ST.UpdateStatus("zonefiles.ready")
	idx.log(idxPrefix, "zonefiles available, kicking off update routine...")

	// Kick off zonefile update loop in the background to keep zonefiles in database updated
	go idx.zonefileIndexLoop()

	idx.log(idxPrefix, "Resolving profiles...")

	// Do an initial profile sync then start profile loop
	if (idx.DB.ZonefilesCount() * 1 / 2) > idx.DB.ProfilesCount() {
		idx.log(idxPrefix, "doing initial profile resolution...")
		idx.ResolveIndexerNames()
	}

	// Set name index status to available
	idx.ST.UpdateStatus("profiles.ready")
	idx.log(idxPrefix, "Kicking off profile update routine")

	// Continually update profiles
	go idx.profileIndexLoop()
}

func (idx *Indexer) log(prefix, message string) {
	log.Printf("%s %s", prefix, message)
}

func (idx *Indexer) nameIndexLoop() {
	ticker := time.NewTicker(idx.config.NameFetchTimeout)
	for _ = range ticker.C {
		idx.log(idxPrefix, "fetching names...")
		idx.GetAllNames()
		idx.log(idxPrefix, fmt.Sprintf("names updated, writing to %s...", idx.config.NameFile))
		idx.WriteNamesToFile(idx.config.NameFile)
		idx.log(idxPrefix, "name file updated")
	}
}

func (idx *Indexer) zonefileIndexLoop() {
	ticker := time.NewTicker(idx.config.ZonefileFetchTimeout)
	for _ = range ticker.C {
		idx.log(idxPrefix, "fetching zonefiles...")
		idx.GetAllZonefiles()
		idx.log(idxPrefix, "zonefiles updated...")
	}
}

func (idx *Indexer) profileIndexLoop() {
	for {
		idx.log(idxPrefix, "resolving all profiles")
		idx.ResolveIndexerNames()
	}
}
