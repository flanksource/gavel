//go:build ignore

// gen_schema.go writes the documented JSON Schema for .gavel.yaml to
// gavel.schema.json at the repo root. Run it via `go generate .` (wired up by
// the //go:generate directive in example_config.go) or directly:
//
//	go run gen_schema.go
//
// The committed artifact is kept honest by
// verify.TestConfigSchema_GoldenMatchesCommitted.
package main

import (
	"fmt"
	"os"

	"github.com/flanksource/gavel/verify"
)

func main() {
	schema, err := verify.ConfigJSONSchema()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile("gavel.schema.json", []byte(schema), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote gavel.schema.json")
}
