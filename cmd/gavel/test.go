package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

var uiServer *testui.Server

func runTests(opts testrunner.RunOptions) (any, error) {
	if opts.UI {
		uiServer = startTestUI()
		updates := make(chan []parsers.Test, 16)
		opts.Updates = updates
		uiServer.StreamFrom(updates)
	}

	result, err := testrunner.Run(opts)
	if err != nil {
		return result, err
	}
	if tests, ok := result.([]parsers.Test); ok {
		summary := parsers.Tests(tests).Sum()
		if summary.Failed > 0 {
			exitCode = 1
		}
		if uiServer != nil {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			return nil, nil
		}
	}
	return result, nil
}

func startTestUI() *testui.Server {
	srv := testui.NewServer()
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Printf("Failed to start test UI server: %v\n", err)
		return nil
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)

	go http.Serve(listener, srv.Handler()) //nolint:errcheck

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Test UI at %s\n", url)
	openBrowser(url)
	return srv
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}

func init() {
	testCmd := clicky.AddNamedCommand("test", rootCmd, testrunner.RunOptions{}, runTests)
	testCmd.Flags().SetInterspersed(false)
}
