package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/blockstack/blockstack.go/blockstack"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

const (
	namePageSize    = 100
	namesCollection = "allNames"
)

// NewIndexer creates a new Indexer
func NewIndexer(cfg *Config) *Indexer {
	client, err := mongo.NewClient(cfg.DB.Connection)
	if err != nil {
		log.Fatal("Failed creating mongodb connection", err)
	}
	idx := &Indexer{
		BSK:         blockstack.NewClient(cfg.BSK.Host),
		DB:          client.Database(cfg.DB.Database).Client(),
		DBName:      cfg.DB.Database,
		Concurrency: cfg.IDX.Concurrency,

		stats: make(map[string]int, 0),
		st:    make(chan map[string]int, 0),

		sem: make(chan struct{}, cfg.IDX.Concurrency),

		zonefileHashChan: make(chan map[string]string, 0),

		names:     []string{},
		namesChan: make(chan []string, 0),
	}

	go idx.handleStatsChan()

	return idx
}

// Indexer is the main stuct for this application
type Indexer struct {
	BSK         *blockstack.Client
	DB          *mongo.Client
	DBName      string
	Concurrency int

	stats map[string]int
	st    chan map[string]int

	names     []string
	namesChan chan []string

	zonefileHashChan chan map[string]string

	sem chan struct{}

	sync.Mutex
	sync.WaitGroup
}

// Index is the main operation
func (idx *Indexer) Index() {
	// First fetch info about the namespaces
	log.Println("[indexer] fetching namespace info")
	nsInfo, _ := idx.GetNSInfo()
	log.Println("[indexer] namespace info fetched")

	// Next fetch the list of all names
	log.Println("[indexer] fetching all names in all namespaces")
	idx.GetNames(nsInfo)
	log.Println("[indexer] all names fetched")

	log.Println("[indexer] fetching all zonefiles")
	idx.GetZonefiles()
	log.Println("[indexer] all zonefiles fetched")
}

// HandleZonefileHashChan ranges over the zonefileHashChan
func (idx *Indexer) HandleZonefileHashChan() {
	for zonefileHashes := range idx.zonefileHashChan {
		idx.Record("zonefileHashes/recieved", len(zonefileHashes))
		keys := make([]string, 0)
		for k := range zonefileHashes {
			keys = append(keys, k)
		}
		ret, err := idx.BSK.GetZonefiles(keys)
		if err != nil {
			panic(err)
		}
		zonefiles := ret.Decode()
		idx.Record("zonefiles/returned", len(zonefiles))
		for zfh, zf := range zonefiles {
			idx.DB.Database(idx.DBName).Collection("zonefiles").InsertOne(context.Background(), bson.NewDocument(
				bson.EC.String("_id", zonefileHashes[zfh]),
				bson.EC.String("zonefile", zf),
			))
			idx.Record("db/write_operations", 1)
		}
	}
}

// GetZonefiles saves the current zonefiles to the mongo database
func (idx *Indexer) GetZonefiles() {
	go idx.HandleZonefileHashChan()
	zfhn := make(map[string]string, 0)
	zfhnChan := make(chan map[string]string, 0)

	// When there are 100 zonefile hashes, send them to get_zonefiles
	go func(zfhn map[string]string, zfhnChan chan map[string]string) {
		for ret := range zfhnChan {
			if len(ret) > 1 {
				panic("NOT SUPPOSED TO BE HERE: More than 1 map[zfh]name coming down zonefileHashChan")
			}
			for k, v := range ret {
				zfhn[k] = v
			}
			if len(zfhn) >= 100 {
				idx.zonefileHashChan <- zfhn
				idx.Record("zonefileHashes/sentToZonefile", len(zfhn))
				zfhn = make(map[string]string, 0)
			}
		}
	}(zfhn, zfhnChan)

	// Limit number of calls to core to idx.Concurrency
	sem := make(chan struct{}, idx.Concurrency)
	var wg sync.WaitGroup
	for _, name := range idx.names {
		wg.Add(1)
		sem <- struct{}{}
		idx.Record("nameDetails/requested", 1)
		go func(name string, zfhnChan chan map[string]string, sem chan struct{}, wg *sync.WaitGroup) {
			res, err := idx.BSK.GetNameBlockchainRecord(name)
			if err != nil {
				panic(err)
			}
			idx.Record("nameDetails/returned", 1)
			if res.Record.ValueHash != "" {
				idx.Record("zonefileHashes/returned", 1)
				zfhnChan <- map[string]string{res.Record.ValueHash: name}
			}
			<-sem
			wg.Done()
		}(name, zfhnChan, sem, &wg)
	}

	wg.Wait()
}

func (idx *Indexer) getNamesToInsert() []*bson.Value {
	out := make([]*bson.Value, 0)
	for _, n := range idx.names {
		out = append(out, bson.VC.String(n))
	}
	return out
}

type namesReturn struct {
	ID        string    `bson:"_id"`
	Names     []string  `bson:"names"`
	Timestamp time.Time `bson:"timestamp"`
}

// GetNames fetches all the names and saves them in the indexer
func (idx *Indexer) GetNames(nsInfo NSInfo) {
	// First try to source names from the database
	idx.fetchDBNames()

	// If there were no names found, fetch them from core
	if len(idx.names) == 0 {
		go idx.handleNameChan()
		for _, ns := range nsInfo.Namespaces() {

			// Record stats for debugging
			idx.Record(fmt.Sprintf("namespace/%s/names", ns), nsInfo[ns])
			idx.Record("namePages/queued", nsInfo.Pages(ns))

			// Create the namePages in a goroutine
			for page := 0; page < nsInfo.Pages(ns); page++ {
				idx.Add(1)
				idx.sem <- struct{}{}
				go idx.namePage(ns, page)
			}
		}

		// Wait for all name pages to return before updating the database
		idx.Wait()
		idx.updateDBNames()
	}
}

func (idx *Indexer) fetchDBNames() {
	nameQueryReturn := &namesReturn{}
	nameQueryBSON := bson.EC.String("_id", "collection")

	err := idx.DB.Database(idx.DBName).Collection(namesCollection).FindOne(
		context.Background(),
		nameQueryBSON,
	).Decode(nameQueryReturn)

	if err != nil {
		log.Println("[indexer] failed fetching names from the database, fetching names from blockstack-core")
		idx.names = []string{}
	} else if len(nameQueryReturn.Names) == 0 {
		log.Println("[indexer] no names returned, fetching names from blockstack-core")
		idx.names = []string{}
	} else if nameQueryReturn.Timestamp.Before(time.Now().Add(-time.Hour * 2)) {
		log.Println("[indexer] cache expired, fetching names from blockstack-core")
		idx.names = []string{}
	} else {
		log.Println("[indexer] got names from database")
		idx.names = nameQueryReturn.Names
	}
}

func (idx *Indexer) updateDBNames() {
	nameQueryBSON := bson.EC.String("_id", "collection")
	nr := &namesReturn{}

	doc := idx.DB.Database(idx.DBName).Collection(namesCollection).FindOneAndUpdate(
		context.Background(),
		nameQueryBSON,
		bson.NewDocument(
			nameQueryBSON,
			bson.EC.ArrayFromElements("names", idx.getNamesToInsert()...),
			bson.EC.DateTime("timestamp", int64(time.Now().Second())),
		))
	if err != nil {
		log.Printf("[database] %#v\n", err)
		log.Println("[database]", err.Decode(nr))
	}
}

func (idx *Indexer) namePage(ns string, page int) {
	idx.Record("namePages/created", 1)
	names, err := idx.BSK.GetNamesInNamespace(ns, page*namePageSize, namePageSize)
	if err != nil {
		panic(err)
	}
	idx.Record(fmt.Sprintf("namespace/%s/fetched", ns), len(names.Names))
	idx.namesChan <- names.Names
	<-idx.sem
	idx.Done()
	idx.Record("namePages/finished", 1)
}

func (idx *Indexer) handleNameChan() {
	for name := range idx.namesChan {
		for _, n := range name {
			idx.names = append(idx.names, n)
		}
	}
}

func (idx *Indexer) zeroCollection(name string) error {
	ret, err := idx.DB.Database(idx.DBName).Collection(name).Count(context.Background(), nil)
	if err != nil {
		return err
	}
	if ret > 0 {
		_, err := idx.DB.Database(idx.DBName).RunCommand(context.Background(), bson.EC.String("drop", name))
		if err != nil {
			return err
		}
	}
	return nil
}
