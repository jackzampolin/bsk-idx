package indexer

import (
	"log"
	"sync"
)

// GetAllZonefiles saves the current zonefiles to the mongo database
func (idx *Indexer) GetAllZonefiles() {
	// chan handles map[zonefileHash]zonefile
	zonefileHashZonefileChan := make(chan map[string]string, 0)
	go idx.handleZonefileHashChan(zonefileHashZonefileChan)

	// this is a map[zonefileHash]name
	zonefileHashNameMap := make(map[string]string, 0)
	// chan handles map[zonefileHash]name
	zonefileHashNameChan := make(chan map[string]string, 0)
	go idx.handleZonefileHashNameChan(zonefileHashNameMap, zonefileHashNameChan, zonefileHashZonefileChan)

	// Limit number of calls to core to idx.Concurrency
	sem := make(chan struct{}, idx.Conc)
	var wg sync.WaitGroup
	for _, name := range idx.names {
		sem <- struct{}{}
		go idx.fetchNameDetails(name, zonefileHashNameChan, sem, &wg)
	}

	wg.Wait()
}

// handleZonefileHashChan ranges over the zonefileHashChan and writes name -> zonefile mappings to the database
func (idx *Indexer) handleZonefileHashChan(zonefileHashChan chan map[string]string) {
	for zonefileHashes := range zonefileHashChan {
		keys := make([]string, 0)
		for k := range zonefileHashes {
			keys = append(keys, k)
		}
		ret, err := idx.GetZonefiles(keys)
		if err != nil {
			panic(err)
		}
		zonefiles := ret.Decode()
		for zfh, zf := range zonefiles {
			err := idx.DB.UpsertNameZonefile(zonefileHashes[zfh], zf)
			if err != nil {
				log.Printf("[zonefiles] Failed to insert or update name zonefile: %s %s\n", zonefileHashes[zfh], err)
			}
		}
	}
}

func (idx *Indexer) handleZonefileHashNameChan(zonefileHashNameMap map[string]string, zonefileHashNameChan chan map[string]string, zonefileHashZonefileChan chan map[string]string) {
	for ret := range zonefileHashNameChan {
		if len(ret) > 1 {
			panic("NOT SUPPOSED TO BE HERE: More than 1 map[zfh]name coming down zonefileHashChan")
		}
		for k, v := range ret {
			zonefileHashNameMap[k] = v
		}
		if len(zonefileHashNameMap) >= 100 {
			zonefileHashZonefileChan <- zonefileHashNameMap
			zonefileHashNameMap = make(map[string]string, 0)
		}
	}
}

func (idx *Indexer) fetchNameDetails(name string, zonefileHashNameChan chan map[string]string, sem chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	res, err := idx.GetNameBlockchainRecord(name)
	if err != nil {
		panic(err)
	}
	if res.Record.ValueHash != "" {
		zonefileHashNameChan <- map[string]string{res.Record.ValueHash: name}
	}
	<-sem
	wg.Done()
}
