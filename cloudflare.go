package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

var cloudflareRanges []*net.IPNet

func updateCloudflareRangesRoutine() {
	for {
		newRanges := make([]*net.IPNet, 0, 30)

		for _, url := range []string{
			"https://www.cloudflare.com/ips-v6/",
			"https://www.cloudflare.com/ips-v4/",
		} {
			resp, err := http.Get(url)
			if err != nil {
				log.Error().Err(err).Msg("failed to fetch cloudflare ips")
				continue
			}
			data, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				_, ipnet, err := net.ParseCIDR(strings.TrimSpace(line))
				if err != nil {
					log.Error().Str("line", line).Err(err).Msg("failed to parse cloudflare ip range")
					continue
				}
				newRanges = append(newRanges, ipnet)
			}
		}
		if len(newRanges) > 0 {
			cloudflareRanges = newRanges
		}

		time.Sleep(time.Hour * 24)
	}
}

func rejectCloudflareEvents(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
	conn := khatru.GetConnection(ctx)
	ip := getRemoteIPAndParse(conn.Request)
	for _, ipnet := range cloudflareRanges {
		if ipnet.Contains(ip) {
			return true, "blastr not allowed"
		}
	}
	return false, ""
}
