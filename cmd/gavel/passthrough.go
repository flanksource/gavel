package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// splitAtDash separates the positional arguments before -- from the raw
// arguments after it, which are passed through to the underlying CLI. Clicky
// binds every positional token to the args field, while pflag preserves the
// separator index.
func splitAtDash(cmd *cobra.Command, positional []string) (before, after []string, err error) {
	dash := cmd.Flags().ArgsLenAtDash()
	if dash < 0 {
		return positional, nil, nil
	}
	if dash > len(positional) {
		return nil, nil, fmt.Errorf("invalid -- separator index %d for %d arguments", dash, len(positional))
	}
	return append([]string(nil), positional[:dash]...), append([]string(nil), positional[dash:]...), nil
}
