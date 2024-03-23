package main

import (
	"context"
	_ "embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr/nip11"
	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

//go:embed GeoLite2-Country.mmdb
var maxmindData []byte

var mm *maxminddb.Reader
var (
	s     Settings
	log   = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	relay = khatru.NewRelay()
)

type Settings struct {
	Port         string `envconfig:"PORT" default:"40404"`
	BaseDomain   string `envconfig:"BASE_DOMAIN" required:"true"`
	DatabasePath string `envconfig:"DATABASE_PATH" default:"./db"`
}

func main() {
	mm, _ = maxminddb.FromBytes(maxmindData)
	if mm == nil {
		log.Fatal().Msg("failed to open maxmind db")
		return
	}

	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig")
		return
	}

	// load db
	db.Path = s.DatabasePath
	if err := db.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
		return
	}
	defer db.Close()
	log.Debug().Str("path", db.Path).Msg("initialized database")

	// init relay
	relay.Info.Name = "countries"
	relay.Info.Description = "serves notes according to your nationality"
	relay.Info.Contact = s.RelayContact
	relay.Info.Icon = s.RelayIcon
	relay.Info.Limitation = &nip11.RelayLimitationDocument{
		RestrictedWrites: true,
	}

	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	relay.RejectEvent = append(relay.RejectEvent,
		policies.PreventLargeTags(100),
		policies.PreventTooManyIndexableTags(8, []int{3, 10002}, nil),
		policies.PreventTooManyIndexableTags(1000, nil, []int{3, 10002}),
	)
	relay.RejectFilter = append(relay.RejectFilter, policies.NoSearchQueries)

	// http routes
	relay.Router().HandleFunc("/", homePage)

	log.Info().Msg("running on http://0.0.0.0:" + s.Port)

	server := &http.Server{Addr: ":" + s.Port, Handler: relay}
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
