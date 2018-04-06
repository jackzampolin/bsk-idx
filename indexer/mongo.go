package indexer

import (
	"log"
	"net/url"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/miekg/dns"
)

// NewMongoDB returns a connected instance of the MongoDB Driver
func NewMongoDB(cfg *Config) *MongoDB {
	session, err := mgo.Dial(cfg.DB.Connection)
	if err != nil {
		log.Fatal("Failed to connect to mongodb at", cfg.DB.Connection, "...")
	}
	return &MongoDB{
		Connection: cfg.DB.Connection,
		Database:   cfg.DB.Database,
		Session:    session,
	}
}

// MongoDB is an implementation of the DB interface
type MongoDB struct {
	Connection string
	Database   string
	Session    *mgo.Session
}

// UpsertNameZonefile takes a name and a zonefile and inserts it as {"_id": name, "zonefile": zonefile}
func (mdb *MongoDB) UpsertNameZonefile(name, zonefile string) error {
	session := mdb.Session.Clone()
	defer session.Close()
	_, err := session.DB(mdb.Database).C("zonefiles").Upsert(bson.M{"_id": name}, bson.M{"_id": name, "zonefile": zonefile})
	if err != nil {
		return err
	}
	return nil
}

// FetchZonefile returns a name/zonefile pairing
func (mdb *MongoDB) FetchZonefile(name string) (NameZonefile, error) {
	session := mdb.Session.Clone()
	defer session.Close()
	zf := &NameZonefileMongo{}
	err := session.DB(mdb.Database).C("zonefiles").Find(bson.M{"_id": name}).One(zf)
	if err != nil {
		return zf, err
	}
	return zf, err
}

// NameZonefileMongo represents a name zonefile pairing
// Contains methods for pulling out different types of resource records
type NameZonefileMongo struct {
	Name     string `bson:"_id"`
	Zonefile string `bson:"zonefile"`
}

// URI returns the URI records from a zonefile
func (nz *NameZonefileMongo) URI() ([]*dns.URI, error) {
	out := make([]*dns.URI, 0)
	for x := range dns.ParseZone(strings.NewReader(nz.Zonefile), "", "") {
		if x.Error != nil {
			return out, x.Error
		}
		uri, ok := x.RR.(*dns.URI)
		if ok {
			out = append(out, uri)
		}
	}
	return out, nil
}

// TXT returns the TXT records from a zonefile
func (nz *NameZonefileMongo) TXT() ([]*dns.TXT, error) {
	out := make([]*dns.TXT, 0)
	for x := range dns.ParseZone(strings.NewReader(nz.Zonefile), "", "") {
		if x.Error != nil {
			return out, x.Error
		}
		txt, ok := x.RR.(*dns.TXT)
		if ok {
			out = append(out, txt)
		}
	}
	return out, nil
}

func (nz *NameZonefileMongo) URL() ([]*url.URL, error) {
	out := make([]*url.URL, 0)
	uri, err := nz.URI()
	if err != nil {
		return out, err
	}
	for _, u := range uri {
		ur, err := url.Parse(u.Target)
		if err != nil {
			continue
		}
		out = append(out, ur)
	}
	return out, nil
}
