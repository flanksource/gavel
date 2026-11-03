package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/spf13/cobra"
)

func init() {
	command := &cobra.Command{Use: "launch", Short: "Launch structured test and lint executions"}
	var root, backend, address string
	var roots []string
	serve := &cobra.Command{Use: "serve", Short: "Serve the launch API", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) (result error) {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		sessions := captaincli.StartSessionHost(ctx)
		defer sessions()
		service, err := newLaunchService(ctx, root, backend, roots)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, service.Close()) }()
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		server := &http.Server{Handler: service.Handler(), ReadHeaderTimeout: 10 * time.Second}
		failure := make(chan error, 1)
		go func() { failure <- server.Serve(listener) }()
		logger.Infof("Launch API: http://%s", listener.Addr())
		select {
		case err := <-failure:
			return err
		case <-ctx.Done():
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return server.Shutdown(shutdown)
		}
	}}
	serve.Flags().StringVar(&root, "root", ".", "Storage root; independent of prepared workspaces")
	serve.Flags().StringVar(&backend, "store", "json", "Result adapter: json or sqlite")
	serve.Flags().StringVar(&address, "addr", "127.0.0.1:9330", "Listen address")
	serve.Flags().StringSliceVar(&roots, "local-root", nil, "Enable local workspaces under these roots")
	command.AddCommand(serve)
	command.AddCommand(launchImportCommand())
	rootCmd.AddCommand(command)
}
