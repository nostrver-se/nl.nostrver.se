package main

import (
	"net"
	"net/http"

	"github.com/fiatjaf/eventstore/lmdb"
)

func getDatabaseForCountry(countryCode string) *lmdb.LMDBBackend {
	return &lmdb.LMDBBackend{
		MaxLimit: 500,
		Path:     s.DatabasePath + "-" + countryCode,
	}
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
