package events

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
)

// gracePeriod is how long the bot may remain disconnected from the gateway
// before the /healthz endpoint reports unhealthy. This gives discordgo's
// automatic reconnect logic (e.g. after a "bad handshake" error) a chance to
// recover before Docker restarts the container.
const gracePeriod = 90 * time.Second

var (
	connected        atomic.Bool
	lastDisconnected atomic.Int64
)

func init() {
	// Start in a "just disconnected" state so the grace period applies from
	// process startup until the first successful gateway connection.
	lastDisconnected.Store(time.Now().Unix())
}

// OnConnect marks the bot as connected to the Discord gateway.
func OnConnect(s *discordgo.Session, event *discordgo.Connect) {
	connected.Store(true)
}

// OnDisconnect marks the bot as disconnected, e.g. due to a websocket
// handshake failure, and records when it happened.
func OnDisconnect(s *discordgo.Session, event *discordgo.Disconnect) {
	connected.Store(false)
	lastDisconnected.Store(time.Now().Unix())
}

// OnResumed marks the bot as connected after resuming a dropped session.
func OnResumed(s *discordgo.Session, event *discordgo.Resumed) {
	connected.Store(true)
}

// StartHealthServer starts an HTTP server exposing /healthz, used by the
// Docker HEALTHCHECK to detect a bot that is stuck failing to (re)connect to
// the Discord gateway.
func StartHealthServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if connected.Load() || time.Since(time.Unix(lastDisconnected.Load(), 0)) < gracePeriod {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("disconnected from Discord gateway"))
	})

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("health server error: %v", err)
		}
	}()
}
