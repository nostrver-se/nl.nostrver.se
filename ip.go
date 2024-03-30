package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/boltdb/bolt"
	es_bolt "github.com/fiatjaf/eventstore/bolt"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/puzpuzpuz/xsync/v3"
)

var (
	dbs         = xsync.NewMapOf[string, *es_bolt.BoltBackend]()
	idMapBucket = []byte("idMap")
)

func storeEventForCountryDB(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	db := getDatabaseForCountry(country)
	return db.SaveEvent(ctx, event)
}

func trackEventOnGlobalDB(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	return globalDB.Update(func(txn *bolt.Tx) error {
		bucket := txn.Bucket(idMapBucket)
		id, _ := hex.DecodeString(event.ID)
		err := bucket.Put(id[0:8], []byte(country))
		if err != nil {
			log.Error().Err(err).Str("id", event.ID).Msg("failed to save id on global")
		}
		return nil
	})
}

func rejectEventForCountry(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	if country != "" && slices.Contains(s.BlockedCountries, country) == true {
		return true, fmt.Sprintf("The country %s is blocked.", country)
	}

	if country == "" {
		return true, "We can't determine your country."
	}

	return false, ""
}

func rejectIfAlreadyHaveInAnyOtherDB(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	globalDB.View(func(txn *bolt.Tx) error {
		conn := khatru.GetConnection(ctx)
		country := getCountryCode(conn.Request)

		bucket := txn.Bucket(idMapBucket)
		id, _ := hex.DecodeString(event.ID)
		existing := bucket.Get(id[0:8])

		if existing != nil && country != string(existing) {
			reject = true
			msg = "event already exists in " + string(existing)
		}
		return nil
	})
	return reject, msg
}

func queryEventForCountryDB(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	db := getDatabaseForCountry(country)

	return db.QueryEvents(ctx, filter)
}

func deleteEventForCountryDB(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	db := getDatabaseForCountry(country)

	return db.DeleteEvent(ctx, event)
}

func getDatabaseForCountry(countryCode string) *es_bolt.BoltBackend {
	db, _ := dbs.LoadOrCompute(countryCode, func() *es_bolt.BoltBackend {
		db := &es_bolt.BoltBackend{
			MaxLimit: 500,
			Path:     s.DatabaseDir + "/" + countryCode,
		}
		if err := db.Init(); err != nil {
			// log.Fatal().Err(err).Msg("failed to initialize database")
			return nil
		}
		// log.Debug().Str("path", db.Path).Msg("initialized database")
		return db
	})
	return db
}

func getCountryCode(r *http.Request) string {
	ip := getRemoteIPAndParse(r)

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	if err := mm.Lookup(ip, &record); err != nil {
		return ""
	}

	return record.Country.ISOCode
}

func getRemoteIPAndParse(r *http.Request) net.IP {
	if xffh := r.Header.Get("X-Forwarded-For"); xffh != "" {
		for _, v := range strings.Split(xffh, ",") {
			if ip := net.ParseIP(strings.TrimSpace(v)); ip != nil && ip.IsGlobalUnicast() && !isPrivate(ip) {
				return ip
			}
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return net.ParseIP(ip)
}

var privateMasks = parseCIDRs(
	"127.0.0.0/8",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
)

func parseCIDRs(ips ...string) []net.IPNet {
	masks := make([]net.IPNet, len(ips))
	for i, cidr := range ips {
		_, netw, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Fatal().Str("ip", cidr).Msg("failed to convert cidr range to mask")
			return nil
		}
		masks[i] = *netw
	}
	return masks
}

func isPrivate(ip net.IP) bool {
	for _, mask := range privateMasks {
		if mask.Contains(ip) {
			return true
		}
	}
	return false
}
