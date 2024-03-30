package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

var idMapBucket = []byte("idMap")

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

func rejectEventForCountryDB(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	if country != "" && strings.Contains(blockedCountries, country) == true {
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

func rejectFilterForCountryDB(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	if country == "" {
		return true, "We can't determine your country."
	}

	return false, ""
}

func deleteEventForCountryDB(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	db := getDatabaseForCountry(country)

	return db.DeleteEvent(ctx, event)
}
