package main

import (
	"fmt"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
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
`)
}
