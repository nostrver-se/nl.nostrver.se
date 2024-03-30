package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func deleteOldStuffRoutine() {
	for {
		entries, err := os.ReadDir(s.DatabaseDir)
		if err != nil {
			log.Error().Err(err).Msg("failed to list databases")
			continue
		}

		country := entries[rand.Intn(len(entries))].Name()
		if country == "global" {
			continue
		}

		db := getDatabaseForCountry(country)
		log.Debug().Str("country", country).Msg("deleting old stuff")
		prevMaxLimit := db.MaxLimit
		db.MaxLimit = 2500

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		_ = cancel

		ch, err := db.QueryEvents(ctx, nostr.Filter{Kinds: []int{1}, Limit: 1500})
		if err != nil {
			log.Error().Err(err).Str("country", country).Msg("error querying")
			db.MaxLimit = prevMaxLimit
			continue
		}

		// we will keep the first 500
		count := 0
		for range ch {
			// skip the first 500
			count++
			if count == 500 {
				break
			}
		}
		count = 0
		for evt := range ch {
			// now we delete all of these
			db.DeleteEvent(ctx, evt)
			count++
		}

		log.Debug().Str("country", country).Int("events", count).Msg("deleted")
		db.MaxLimit = prevMaxLimit

		time.Sleep(time.Hour * 7)
	}
}
