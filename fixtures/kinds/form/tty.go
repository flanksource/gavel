package form

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/schemaform"

	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
)

// TTYPrompter answers form steps on the controlling terminal through clicky's
// schema form (huh), the terminal twin of the dashboard's PromptDialog.
type TTYPrompter struct {
	// By is the answerer recorded on every answer (TerminalAnswerer).
	By string
	// Lookup resolves `x-clicky-lookup` fields (a `format: connection`
	// picker); without one such a form fails loudly on the terminal.
	Lookup schemaform.Lookup
}

func (t TTYPrompter) Ask(ctx context.Context, p prompt.Prompt) (prompt.Answer, error) {
	values, err := clicky.PromptForm(ctx, clicky.PromptFormOptions{
		Title: p.Title, Description: p.Description, Schema: p.Schema, Default: p.Default, Lookup: t.Lookup,
	})
	if errors.Is(err, clicky.ErrPromptCancelled) {
		return prompt.Answer{Cancelled: true, At: time.Now(), By: t.By}, nil
	}
	if err != nil {
		return prompt.Answer{}, err
	}
	return prompt.Answer{Values: values, At: time.Now(), By: t.By}, nil
}

// TerminalLookup resolves a terminal form's `x-clicky-lookup` fields: a
// field bound by its step (an `x-query`, or a connection picker in a fixture
// declaring `connections:`) through its lookup token, scoped by the form's
// values; any other `format: connection` picker over Catalog(workDir), scoped
// by the types in its lookup URL.
func TerminalLookup(workDir string, opts dbconn.Options) schemaform.Lookup {
	return func(ctx context.Context, req schemaform.LookupRequest) (schemaform.LookupResult, error) {
		token, pointer, bound, err := lookup.ParseURL(req.Spec.URL)
		if err != nil {
			return schemaform.LookupResult{}, err
		}
		if bound {
			result, err := lookup.Resolve(ctx, lookup.Request{Token: token, Pointer: pointer, Query: req.Query, Values: req.Values})
			return result.LookupResult, err
		}
		if req.Spec.Filter != ConnectionLookupFilter {
			return schemaform.LookupResult{}, fmt.Errorf("no terminal lookup for x-clicky-lookup filter %q", req.Spec.Filter)
		}
		types, err := dbconn.LookupTypes(req.Spec.URL)
		if err != nil {
			return schemaform.LookupResult{}, err
		}
		return lookup.ConnectionOptions(ctx, lookup.ConnectionQuery{WorkDir: workDir, Options: opts, Types: types, Query: req.Query})
	}
}

// TerminalAnswerer names the person at this terminal: $USER (else the OS
// account) with their `git config user.email` when one is set.
func TerminalAnswerer(ctx context.Context) (string, error) {
	name := os.Getenv("USER")
	if name == "" {
		account, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("identify the terminal user: %w", err)
		}
		name = account.Username
	}
	out, err := exec.CommandContext(ctx, "git", "config", "user.email").Output()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit) && exit.ExitCode() == 1:
		// `git config` exits 1 when the key is unset: the user name alone.
		return name, nil
	case err != nil:
		return "", fmt.Errorf("read git config user.email: %w", err)
	}
	if email := strings.TrimSpace(string(out)); email != "" {
		return fmt.Sprintf("%s <%s>", name, email), nil
	}
	return name, nil
}
