// Command svc is the services-kind test service: an HTTP server that answers
// /healthz, optionally delays before listening, and appends its label to a
// file when it receives SIGTERM so tests can observe teardown order.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", "", "listen address")
	label := flag.String("label", "svc", "label written to --term-file on SIGTERM")
	termFile := flag.String("term-file", "", "file the label is appended to on SIGTERM")
	delay := flag.Duration("delay", 0, "wait before listening")
	flag.Parse()
	if *addr == "" {
		fmt.Fprintln(os.Stderr, "--addr is required")
		os.Exit(2)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	go func() {
		<-signals
		if *termFile != "" {
			f, err := os.OpenFile(*termFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Fprintln(f, *label)
			_ = f.Close()
		}
		os.Exit(0)
	}()

	fmt.Printf("%s booting, upstream=%q\n", *label, os.Getenv("UPSTREAM"))
	time.Sleep(*delay)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, *label) })
	fmt.Printf("%s listening on %s\n", *label, *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
