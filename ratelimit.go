package main

import (
	"context"
	"encoding/hex"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/puzpuzpuz/xsync/v3"
)

// this same map stores buckets by IP and by PUBKEY
var buckets = xsync.NewMapOf[string, *atomic.Int32]()

func bucketFillingRoutine() {
	for {
		time.Sleep(time.Hour * 2)

		buckets.Range(func(_ string, bucket *atomic.Int32) bool {
			bucket.Add(1)
			return true
		})
	}
}

func rateLimit(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	conn := khatru.GetConnection(ctx)
	ip := net.ParseIP(strings.Split(conn.Request.RemoteAddr, ":")[0])

	for _, key := range []string{hex.EncodeToString(ip), event.PubKey} {
		bucket, loaded := buckets.LoadOrStore(key, &atomic.Int32{})
		if !loaded {
			bucket.Add(2)
		}

		if bucket.Load() <= 0 {
			return true, "rate-limit reached, must slow down"
		}

		bucket.Add(-1)
	}
	return false, ""
}
