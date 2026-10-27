package launch

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func runTests(ctx context.Context, options testrunner.RunOptions, publish Publish) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	updates := make(chan []parsers.Test, 16)
	options.Context, options.Updates = ctx, updates
	type outcome struct {
		result any
		err    error
	}
	finished := make(chan outcome, 1)
	go func() { result, err := testrunner.Run(options); finished <- outcome{result, err} }()
	var publicationError error
	for {
		select {
		case tree, open := <-updates:
			if !open {
				updates = nil
				continue
			}
			if publicationError == nil {
				publicationError = publish(resultstore.Run{Tests: tree})
				if publicationError != nil {
					cancel()
				}
			}
		case result := <-finished:
			if publicationError != nil {
				return errors.Join(result.err, publicationError)
			}
			if result.result == nil {
				return result.err
			}
			tests, ok := result.result.([]parsers.Test)
			if !ok {
				return fmt.Errorf("test runner returned unexpected result %T", result.result)
			}
			return errors.Join(result.err, publish(resultstore.Run{Tests: tests}))
		}
	}
}
