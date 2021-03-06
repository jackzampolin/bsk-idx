package indexer

import (
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/miekg/dns"
)

const (
	profilesCollection  = "profiles"
	zonefilesCollection = "zonefiles"
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

	sync.Mutex
}

// UpsertNameZonefile takes a name and a zonefile and inserts it as {"_id": name, "zonefile": zonefile}
func (mdb *MongoDB) UpsertNameZonefile(name, zonefile string) error {
	session := mdb.Session.Clone()
	defer session.Close()
	_, err := session.DB(mdb.Database).C(zonefilesCollection).Upsert(bson.M{"_id": name}, bson.M{"_id": name, "zonefile": zonefile})
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
	findFilter := bson.M{"_id": name}
	err := session.DB(mdb.Database).C(zonefilesCollection).Find(findFilter).One(zf)
	if err != nil {
		return zf, err
	}
	return zf, err
}

// UpsertProfile takes a name and a profile and inserts it as {"_id": name, "profile": profile}
func (mdb *MongoDB) UpsertProfile(name string, profile Profile) error {
	session := mdb.Session.Clone()
	defer session.Close()
	upsertFilter := bson.M{"_id": name}
	upsertData := bson.M{"_id": name, "profile": profile}
	_, err := session.DB(mdb.Database).C(profilesCollection).Upsert(upsertFilter, upsertData)
	if err != nil {
		return err
	}
	return nil
}

// ZonefilesCount returns the count of all zonefiles
func (mdb *MongoDB) ZonefilesCount() int {
	session := mdb.Session.Clone()
	defer session.Close()
	count, err := session.DB(mdb.Database).C(zonefilesCollection).Count()
	if err != nil {
		return 0
	}
	return count
}

// ProfilesCount returns the count of all zonefiles
func (mdb *MongoDB) ProfilesCount() int {
	session := mdb.Session.Clone()
	defer session.Close()
	count, err := session.DB(mdb.Database).C(profilesCollection).Count()
	if err != nil {
		return 0
	}
	return count
}

// NameProfileMongo models a name profile pairing
type NameProfileMongo struct {
	Name    string  `bson:"_id"`
	Profile Profile `bson:"profile"`
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
