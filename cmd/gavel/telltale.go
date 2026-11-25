package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/flanksource/commons/logger"
	telltale "github.com/flanksource/telltale/sdk"
	_ "github.com/flanksource/telltale/sdk/commonsdb"
)

func startDashboardTelltale(hosted bool) (func(), error) {
	mode := os.Getenv("TELLTALE_AGENT")
	if mode == "0" || (hosted && mode != "1" && os.Getenv("TELLTALE_URL") == "") {
		return func() {}, nil
	}
	agent, err := telltale.Start(telltale.Options{Service: "gavel", Version: version})
	if err != nil {
		return nil, fmt.Errorf("start Gavel Telltale agent: %w", err)
	}
	telltale.InstrumentDefaultTransport()
	return func() {
		if err := agent.Close(); err != nil {
			logger.Errorf("close Gavel Telltale agent: %v", err)
		}
	}, nil
}

func dashboardHTTPHandler(next http.Handler) http.Handler {
	return telltale.Handler(next)
}
