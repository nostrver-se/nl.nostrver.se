package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/fiatjaf/eventstore/bolt"
	"github.com/puzpuzpuz/xsync/v3"
)

var dbs = xsync.NewMapOf[string, *bolt.BoltBackend]()

func getDatabaseForCountry(countryCode string) *bolt.BoltBackend {
	db, _ := dbs.LoadOrCompute(countryCode, func() *bolt.BoltBackend {
		db := &bolt.BoltBackend{
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

// Gets the country code in ISO 3166-1 alpha-2 format.
// On error returns an empty string.
func getCountryCode(r *http.Request) string {
	ip := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])

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
