package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

var (
	lastEvents     [144]string
	lastIndex      int  = -1
	lastEventsFull bool = false
)

func memoryTrack(ctx context.Context, event *nostr.Event) error {
	conn := khatru.GetConnection(ctx)
	country := getCountryCode(conn.Request)

	lastIndex = (lastIndex + 1) % len(lastEvents)
	if lastIndex == len(lastEvents)-1 {
		lastEventsFull = true // means we've completed a full circle
	}
	lastEvents[lastIndex] = country

	return nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	str := `
countries
=========

this is a nostr relay that only serves events from your country.

events you write to it will only be served to people in your country.
you can only read events published by other people in your country.
any event can only exist in one country.

this is all done using a magic property from our internet called the "ip address".
it is not very hard to bypass.

the source code for this relay is available at https://git.fiatjaf.com/countries

you can check the feed of just this relay by visiting:

  - https://nostrrr.com/relay/countries.fiatjaf.com
  - https://coracle.social/relays/countries.fiatjaf.com
`

	if lastIndex >= 0 {
		total := len(lastEvents)

		start := lastIndex
		target := total - 1
		if lastEventsFull {
			if lastIndex == 0 {
				start = total - 1
			}
			target = start
		}

		str += "\n\nlatest events\n=============\n\n"
		seq := 0
		for i := start; i != target || seq == 0; i = (i - 1 + total) % total {
			str += generateFlag(lastEvents[i]) + " "
			seq++
			if seq%12 == 0 {
				str += "\n"
			}
		}
	}

	fmt.Fprint(w, str)
}

func generateFlag(country string) string {
	flag := make([]rune, 2)
	for i, letter := range country {
		offset := letter - 'A'
		flag[i] = rune(0x1F1E6 + offset)
	}
	return string(flag)
}
