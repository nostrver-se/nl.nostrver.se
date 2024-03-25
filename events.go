package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func storeEventForCountryDB(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)
	db := getDatabaseForCountry(country)

	return db.SaveEvent(ctx, event)
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
