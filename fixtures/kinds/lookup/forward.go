package lookup

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/fixtures"
)

// An isolated run's form steps register their fields in the worker process;
// the parent forwards each field a relayed prompt names to it.
func init() {
	fixtures.WorkerLookups = fixtures.LookupRelay{Forward: Forward, Resolve: resolveForwarded, NotFound: ErrNotFound}
}

// Forward holds, until release, every registered field schema's
// `x-clicky-lookup` URLs name under the token the URL carries, each lookup of
// it answered by resolve — the worker that registered it.
func Forward(schema json.RawMessage, resolve fixtures.RemoteLookup) (release func(), err error) {
	var doc any
	if err := json.Unmarshal(schema, &doc); err != nil {
		return nil, fmt.Errorf("prompt schema: %w", err)
	}
	urls := lookupURLs(doc, nil)
	var tokens []string
	release = func() {
		for _, token := range tokens {
			Release(token)
		}
	}
	for _, raw := range urls {
		token, pointer, ok, err := ParseURL(raw)
		if err != nil {
			release()
			return nil, err
		}
		if !ok {
			continue
		}
		if err := hold(token, entry{pointer: pointer, binding: forwardedBinding{worker: resolve}}); err != nil {
			release()
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return release, nil
}

func hold(token string, found entry) error {
	if err := found.binding.validate(); err != nil {
		return fmt.Errorf("lookup binding %s: %w", found.pointer, err)
	}
	registry.Lock()
	defer registry.Unlock()
	if _, ok := registry.entries[token]; ok {
		return fmt.Errorf("lookup token %q is already registered", token)
	}
	registry.entries[token] = found
	return nil
}

// lookupURLs is every `x-clicky-lookup` url in a schema, appended to urls.
func lookupURLs(node any, urls []string) []string {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if spec, ok := child.(map[string]any); ok && key == "x-clicky-lookup" {
				if url, ok := spec["url"].(string); ok {
					urls = append(urls, url)
				}
				continue
			}
			urls = lookupURLs(child, urls)
		}
	case []any:
		for _, child := range value {
			urls = lookupURLs(child, urls)
		}
	}
	return urls
}

// forwardedBinding answers a field registered in another process.
type forwardedBinding struct {
	worker fixtures.RemoteLookup
}

func (b forwardedBinding) validate() error {
	if b.worker == nil {
		return fmt.Errorf("a forwarded field needs the process that resolves it")
	}
	return nil
}

func (b forwardedBinding) resolve(ctx context.Context, req Request) (Result, error) {
	request, err := json.Marshal(req)
	if err != nil {
		return Result{}, fmt.Errorf("lookup request: %w", err)
	}
	raw, err := b.worker(ctx, request)
	if err != nil {
		return Result{}, err
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return Result{}, fmt.Errorf("forwarded lookup result: %w", err)
	}
	return result, nil
}

// resolveForwarded answers, in the process that registered the field, a
// request another process forwarded.
func resolveForwarded(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("forwarded lookup request: %w", err)
	}
	result, err := Resolve(ctx, req)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
