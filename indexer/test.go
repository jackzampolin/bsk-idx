package indexer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Zonefiles does the things
// TODO: Remove
func (idx *Indexer) Zonefiles() {
	for _, n := range idx.names {
		profile := idx.GetProfile(n)
		if profile != nil {
			log.Println(profile.getPubKey(), n)
			out := strings.Split(profile.Token, ".")
			for _, o := range out {
				dec, err := jwt.DecodeSegment(o)
				if err != nil {
					panic(err)
				}
				log.Println(string(dec))
			}
		}
	}
}

// GetProfile takes a name and returns the profile associated
// NOTE: This method makes a DB query and an HTTP request
func (idx *Indexer) GetProfile(n string) *ProfileTokenFile {
	// First fetch the zonefile data from the databse
	zf, err := idx.DB.FetchZonefile(n)
	if err != nil {
		idx.ST.Rec("profiles/zonefiles/invalid", 1)
		return nil
	}
	idx.ST.Rec("profiles/zonefile/fetch/success", 1)

	// Pull the URI's URLs from the Zonefile
	urls, err := zf.URL()
	if err != nil {
		idx.ST.Rec("profiles/zonefile/parse/error", 1)
		return nil
	}

	idx.ST.Rec("profiles/zonefiles/parse/success", 1)

	// Loop over all URLs from URI records
	profiles := []*ProfileTokenFile{}
	for _, url := range urls {
		p, err := fetchProfile(url)
		// This error could be an http, ioutil, or unmarshal
		if err != nil {
			log.Println("Failed fetch for", n)
			idx.ST.Rec("profiles/fetch/error", 1)
			continue
		}

		// If there is a profile in the return and there is a ParentPublicKey return
		// This is the path for new profiles
		if len(p) > 0 {
			if p[0].ParentPublicKey != "" {
				profiles = append(profiles, p[0])
				idx.ST.Rec("profiles/fetch/success", 1)
			} else if p[0].Token != "" {
				profiles = append(profiles, p[0])
				idx.ST.Rec("profiles/fetch/success", 1)
			}
			continue
		}

		log.Println(p, err)
		panic("Supposedly Unreachable")
	}

	// Handle conditions
	if len(profiles) > 1 {
		// If there is more than one profile I would like to see that
		log.Println(profiles)
		panic("Supposedly Unreachable MORE THAN ONE PROFILE ON THIS NAME")
	} else if len(profiles) == 0 {
		// If there is no profile, then return nil
		return nil
	}

	// Otherwise return the profile
	return profiles[0]
}

func fetchProfile(u *url.URL) ([]*ProfileTokenFile, error) {
	p := make([]*ProfileTokenFile, 0)
	res, err := http.Get(u.String())
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
