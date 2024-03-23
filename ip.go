package main

import (
	"net"
	"net/http"

	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/puzpuzpuz/xsync/v3"
)

var dbs = xsync.NewMapOf[string, *lmdb.LMDBBackend]()

func getDatabaseForCountry(countryCode string) *lmdb.LMDBBackend {
	db, _ := dbs.LoadOrCompute(countryCode, func() *lmdb.LMDBBackend {
		db := &lmdb.LMDBBackend{
			MaxLimit: 500,
			Path:     s.DatabasePath + "-" + countryCode,
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
	ip := net.ParseIP(r.RemoteAddr)

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
