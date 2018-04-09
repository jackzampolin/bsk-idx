package indexer

import (
	"net/url"

	"github.com/miekg/dns"
)

// IndexerDB is the database driver interface for the Indexer
type DB interface {
	UpsertNameZonefile(name, zonefile string) error
	ZonefilesCount() int
	ProfilesCount() int
	FetchZonefile(name string) (NameZonefile, error)
	UpsertProfile(name string, profile Profile) error
}

// NameZonefile represents a return from the database for fetching a name/zonefile pair
// convinence methods for pulling out different resource records
type NameZonefile interface {
	URI() ([]*dns.URI, error)
	URL() ([]*url.URL, error)
	TXT() ([]*dns.TXT, error)
}
