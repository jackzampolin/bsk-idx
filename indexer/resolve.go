package indexer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ResolveIndexerNames loops through the `names` array on the indexer struct and pulls all the profiles for those names.s
func (idx *Indexer) ResolveIndexerNames() {
	// Limit concurrency to idx.Conc and wait for all names to resolve before exiting
	sem := make(chan struct{}, idx.Conc)
	var wg sync.WaitGroup

	// Loop over names and insert them
	for _, n := range idx.names.current() {
		if n != "" {
			sem <- struct{}{}
			wg.Add(1)
			go resolveAndInsert(idx, n, &wg, sem)
		}
	}

	// Wait for all names to be resolved
	wg.Wait()
}

// resolveAndInsert fetches the profile from storage and then inserts that profile into configured DB driver
func resolveAndInsert(idx *Indexer, name string, wg *sync.WaitGroup, sem chan struct{}) {
	profile := idx.GetProfile(name)
	if profile != nil && profile.DecodedToken.Payload.Claim.Type == "Person" {
		err := idx.DB.UpsertProfile(name, profile.DecodedToken.Payload.Claim)
		if err != nil {
			idx.ST.Rec("profiles.insert_error", 1)
		}
		idx.ST.Rec("profiles.inserted", 1)
	}
	<-sem
	wg.Done()
	idx.ST.Rec("zonefiles.resolved", 1)
}

// GetProfile takes a name and returns the profile associated
// NOTE: This method makes a DB query and an HTTP request
func (idx *Indexer) GetProfile(n string) *ProfileTokenFile {
	// First fetch the zonefile data from the databse
	zf, err := idx.DB.FetchZonefile(n)
	if err != nil {
		idx.ST.Rec("profiles.zf_invalid", 1)
		return nil
	}
	idx.ST.Rec("profiles.zf_valid", 1)

	// Pull the URI's URLs from the Zonefile
	urls, err := zf.URL()
	if err != nil {
		idx.ST.Rec("profiles.zf_parse_error", 1)
		return nil
	}

	idx.ST.Rec("profiles.zf_parsed", 1)

	// Loop over all URLs from URI records
	profiles := []*ProfileTokenFile{}
	for _, url := range urls {
		p, err := fetchProfile(url)
		// This error could be an http, ioutil, or unmarshal
		if err != nil {
			idx.ST.Rec("profiles.fetch_error", 1)
			continue
		}

		// If there is a profile in the return and there is a ParentPublicKey return
		// This is the path for new profiles
		if len(p) > 0 {
			if p[0].ParentPublicKey != "" {
				profiles = append(profiles, p[0])
				idx.ST.Rec("profiles.fetch_success", 1)
			} else if p[0].Token != "" {
				profiles = append(profiles, p[0])
				idx.ST.Rec("profiles.fetch_success", 1)
			}
			continue
		}

		log.Println(p, err)
		panic("Supposedly Unreachable")
	}

	// Handle conditions
	if len(profiles) > 1 {
		idx.ST.Rec("profiles.multiple_profiles", 1)
	} else if len(profiles) == 0 {
		// If there is no profile, then return nil
		return nil
	}

	// Otherwise return the profile
	return profiles[0]
}

// fetchProfile takes a URL and returns the JSON that was recieved back
func fetchProfile(u *url.URL) ([]*ProfileTokenFile, error) {
	p := make([]*ProfileTokenFile, 0)
	client := http.DefaultClient
	client.Timeout = 30 * time.Second
	res, err := client.Get(u.String())
	if err != nil {
		return p, err
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(bodyBytes, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}
