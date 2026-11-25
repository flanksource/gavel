package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/workspace"
	"github.com/spf13/cobra"
)

func newCacheCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "cache",
		Short: "Show the local .gavel cache location and size",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dirs, err := workspace.From(cmd.Context())
			if err != nil {
				return err
			}
			return printCacheInfo(cmd, dirs.StateRoot)
		},
	}
	command.AddCommand(&cobra.Command{
		Use:     "compress",
		Short:   "Compress saved test and lint results at or above the configured threshold",
		Long:    "Compress existing settled JSON snapshots. Set gavel.compress.threshold to a byte size (default 64KB) and gavel.compress.default to zstd (default) or gzip.",
		Example: "  gavel cache compress\n  gavel cache compress -P gavel.compress.default=gzip",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dirs, err := workspace.From(cmd.Context())
			if err != nil {
				return err
			}
			count, err := snapshots.CompressCache(cmd.Context(), dirs.StateRoot)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Compressed %d snapshot(s).\n", count)
			return printCacheInfo(cmd, dirs.StateRoot)
		},
	})
	command.AddCommand(newCacheCleanupCommand())
	return command
}

func newCacheCleanupCommand() *cobra.Command {
	var ages []string
	command := &cobra.Command{
		Use:     "cleanup --age <duration|size:duration>",
		Short:   "Remove expired test and lint snapshots from the local cache",
		Long:    "Delete settled snapshots matching any --age rule. Files must be older than the given duration. A size prefix applies the rule only to files at least that large on disk. Runs referenced by last.json or a branch pointer are kept.",
		Example: "  gavel cache cleanup --age 30d --age 10MB:10d",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rules, err := snapshots.ParseCacheAges(ages)
			if err != nil {
				return err
			}
			dirs, err := workspace.From(cmd.Context())
			if err != nil {
				return err
			}
			count, err := snapshots.CleanupCache(cmd.Context(), dirs.StateRoot, rules, time.Now())
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %d snapshot(s).\n", count)
			return printCacheInfo(cmd, dirs.StateRoot)
		},
	}
	command.Flags().StringArrayVar(&ages, "age", nil, "Expiration rule such as 30d or 10MB:10d (repeatable)")
	return command
}

func printCacheInfo(cmd *cobra.Command, workDir string) error {
	stats, err := snapshots.CacheInfo(workDir)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", stats.Path, humanize.Bytes(uint64(stats.Bytes)))
	return err
}

func init() {
	rootCmd.AddCommand(newCacheCommand())
}
