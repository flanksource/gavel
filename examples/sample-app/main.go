// Command sample-app is a tiny HTTP service used as a runnable target for the
// gavel fixture examples in ../. It exposes GET /health returning
// {"status":"ok"} so the smoke-test fixture can probe it, and carries a couple
// of fast tests the precommit / pre-release fixtures run and lint.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

// healthHandler reports service liveness as JSON.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func newServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	// ReadHeaderTimeout is set explicitly (rather than http.ListenAndServe) so
	// the server is not flagged by gosec G114.
	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	srv := newServer(*addr)
	log.Printf("sample-app listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
