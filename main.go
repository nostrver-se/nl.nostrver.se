package main

import (
	"context"
	_ "embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/zerolog"
	"github.com/sebest/xff"
	"golang.org/x/sync/errgroup"
)

//go:embed GeoLite2-Country.mmdb
var maxmindData []byte

var mm *maxminddb.Reader
var (
	s        Settings
	log      = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	relay    = khatru.NewRelay()
	globalDB *bolt.DB
)

type Settings struct {
	Port         string `envconfig:"PORT" default:"40404"`
	DatabaseDir  string `envconfig:"DATABASE_DIR" default:"./db"`
	RelayContact string `envconfig:"RELAY_CONTACT" required:"false"`
	RelayIcon    string `envconfig:"RELAY_ICON" required:"false"`

	BlockedCountries []string `envconfig:"BLOCKED_COUNTRIES"`
}

func main() {
	// load environment variables
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig")
		return
	}

	// load ip data
	mm, _ = maxminddb.FromBytes(maxmindData)
	if mm == nil {
		log.Fatal().Msg("failed to open maxmind db")
		return
	}

	// open global boltdb
	os.MkdirAll(s.DatabaseDir, 0755)
	globalDB, err = bolt.Open(s.DatabaseDir+"/global", 0644, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open global db")
		return
	}
	if err := globalDB.Update(func(txn *bolt.Tx) error {
		_, err := txn.CreateBucketIfNotExists([]byte("idMap"))
		return err
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to open idMap bucket on global db")
		return
	}

	// start cloudflare thing
	go updateCloudflareRangesRoutine()

	// ratelimiter
	go bucketFillingRoutine()

	relay.Info.Name = "countries"
	relay.Info.Description = "serves notes according to your country"
	relay.Info.Contact = s.RelayContact
	relay.Info.Icon = s.RelayIcon
	relay.Info.Limitation = &nip11.RelayLimitationDocument{}

	relay.StoreEvent = append(relay.StoreEvent,
		storeEventForCountryDB,
		trackEventOnGlobalDB,
	)
	relay.QueryEvents = append(relay.QueryEvents, queryEventForCountryDB)
	relay.DeleteEvent = append(relay.DeleteEvent, deleteEventForCountryDB)
	relay.RejectEvent = append(relay.RejectEvent,
		policies.RestrictToSpecifiedKinds(1),
		rejectEventForCountry,
		rejectCloudflareEvents,
		policies.PreventLargeTags(12),
		rejectIfAlreadyHaveInAnyOtherDB,
		rateLimit,
	)

	relay.RejectFilter = append(relay.RejectFilter,
		policies.NoSearchQueries,
	)

	// http routes
	relay.Router().HandleFunc("/", homePage)

	log.Info().Msg("running on http://0.0.0.0:" + s.Port)

	xffmw, _ := xff.Default()
	server := &http.Server{Addr: ":" + s.Port, Handler: xffmw.Handler(relay)}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(server.ListenAndServe)
	g.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Debug().Err(err).Msg("exit reason")
	}
}
