package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture worker lookup relay", func() {
	var (
		relay    *promptRelay
		prompter *workerPrompter
		notFound = errors.New("no such field")
		// resolve stands in for the worker's lookup registry.
		resolve RemoteLookup
	)

	BeforeEach(func() {
		saved := WorkerLookups
		WorkerLookups = LookupRelay{NotFound: notFound, Resolve: func(ctx context.Context, request json.RawMessage) (json.RawMessage, error) {
			return resolve(ctx, request)
		}}
		ctx, cancel := context.WithCancel(context.Background())
		answers, parent := io.Pipe()
		relay = newPromptRelay(ctx, nil, parent)
		var err error
		prompter, err = startWorkerPrompter(ctx, relay.handle, answers)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			prompter.stopLookups()
			cancel()
			Expect(relay.close()).To(Succeed())
			Expect(parent.Close()).To(Succeed())
			WorkerLookups = saved
		})
	})

	It("answers a parent's lookup with what the worker resolves for its request", func() {
		resolve = func(_ context.Context, request json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(fmt.Sprintf(`{"echo":%s}`, request)), nil
		}

		result, err := relay.lookup(context.Background(), json.RawMessage(`{"Query":"ord"}`))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(MatchJSON(`{"echo":{"Query":"ord"}}`))
	})

	It("keeps a field the worker does not hold identifiable as not found", func() {
		resolve = func(context.Context, json.RawMessage) (json.RawMessage, error) {
			return nil, fmt.Errorf("token abc: %w", notFound)
		}

		_, err := relay.lookup(context.Background(), json.RawMessage(`{}`))

		Expect(err).To(MatchError(notFound))
		Expect(err).To(MatchError("token abc: no such field"))
	})

	It("ends the worker's lookup when the parent gives up on it, and keeps answering later lookups", func() {
		started, ended := make(chan struct{}), make(chan error, 1)
		resolve = func(ctx context.Context, _ json.RawMessage) (json.RawMessage, error) {
			close(started)
			<-ctx.Done()
			ended <- ctx.Err()
			return nil, ctx.Err()
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-started
			cancel()
		}()

		_, err := relay.lookup(ctx, json.RawMessage(`{}`))

		Expect(err).To(MatchError(context.Canceled))
		Eventually(ended).Should(Receive(MatchError(context.Canceled)))
		resolve = func(context.Context, json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"ok":true}`), nil
		}
		result, err := relay.lookup(context.Background(), json.RawMessage(`{}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(MatchJSON(`{"ok":true}`))
	})
})
