package indexer

import (
	"net/url"

	"github.com/miekg/dns"
)

// IndexerDB is the database driver interface for the Indexer
type DB interface {
	UpsertNameZonefile(name, zonefile string) error
	FetchZonefile(name string) (NameZonefile, error)
}

// NameZonefile represents a return from the database for fetching a name/zonefile pair
// convinence methods for pulling out different resource records
type NameZonefile interface {
	URI() ([]*dns.URI, error)
	URL() ([]*url.URL, error)
	TXT() ([]*dns.TXT, error)
}