package httpx

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky/shutdown"
	"github.com/flanksource/commons-db/observability"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/har"
	commonshttp "github.com/flanksource/commons/http"
	"github.com/flanksource/commons/http/middlewares/tracing"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
)

var harRegistry = har.NewRegistry(nil)

func init() {
	shutdown.AddHookWithPriority("HTTP HAR", shutdown.PriorityCritical, func() {
		if err := harRegistry.Flush(); err != nil {
			logger.Errorf("flush HTTP HAR: %v", err)
		}
	})
}

// Configure attaches shared logging, property-driven HAR capture, and Gavel's
// HTTP traffic observer before the client makes its first request.
func Configure(ctx context.Context, feature string, client *commonshttp.Client) error {
	collector, _, level, err := harRegistry.For(feature)
	if err != nil {
		return fmt.Errorf("configure %s HTTP capture: %w", feature, err)
	}
	log := commonscontext.LoggerFromContext(ctx)
	client.WithLogger(log)
	logLevel := max(log.GetLevel(), logger.GetLogger("http").GetLevel(), logger.GetLogger("http."+feature).GetLevel())
	if config, ok := commonshttp.TraceConfigForLogLevel(logLevel); ok {
		client.TraceToStdout(config)
	}
	if value := properties.String("", "http."+feature+".log", "http.log"); value != "" {
		config := tracing.TraceConfigFromString(value)
		config.RedactedHeaders = append(config.RedactedHeaders, "Authorization")
		client.TraceToStdout(config)
	}
	if collector != nil {
		if level == har.Metadata {
			client.Use(collector.MetadataMiddleware())
		} else {
			client.HARCollector(collector)
		}
	}
	observability.ApplyHTTPClientObservability(ctx, Feature, client, nil)
	return nil
}
