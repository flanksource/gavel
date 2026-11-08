package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/verify"
)

type SlotsOptions struct {
	Project bool `json:"project,omitempty" flag:"project" help:"Show only the global pool and this repository's"`
}

func init() {
	cmd := clicky.AddNamedCommandWithContext("slots", rootCmd, SlotsOptions{}, runSlots)
	cmd.Short = "List the machine-wide slots every gavel process holds or waits for"
	cmd.Long = "Lists each slot pool — global, and one per repository shared by its worktrees — with the processes holding a slot and those queued for one. " +
		"Running gavel processes also publish their test and lint progress, which `telltale top` shows live."
}

func runSlots(ctx context.Context, opts SlotsOptions) (any, error) {
	dirs, err := commandDirs(ctx, "")
	if err != nil {
		return nil, err
	}
	pools, err := slots.Inspect(dirs.User.Cache)
	if err != nil {
		return nil, err
	}
	if opts.Project {
		_, id, err := slots.ProjectDir(ctx, dirs.Invocation)
		if err != nil {
			return nil, err
		}
		pools = projectPools(pools, "project/"+id)
	}
	if len(pools) > 0 {
		return slotRows(pools, time.Now()), nil
	}
	config, err := verify.LoadGavelConfig(dirs.Invocation)
	if err != nil {
		return nil, err
	}
	limits, err := config.SlotLimits()
	if err != nil {
		return nil, err
	}
	if !limits.Enabled() {
		return clicky.Text("No slots configured; see slots in ~/.gavel.yaml"), nil
	}
	return clicky.Text("No slot taken yet"), nil
}

func projectPools(pools []slots.PoolStatus, project string) []slots.PoolStatus {
	var kept []slots.PoolStatus
	for _, pool := range pools {
		if pool.Name == "global" || pool.Name == project {
			kept = append(kept, pool)
		}
	}
	return kept
}

// slotRow is one held slot, one queued process, or an idle pool.
type slotRow struct {
	Pool    string `json:"pool"`
	Use     string `json:"use"`
	Project string `json:"project,omitempty"`
	Slot    string `json:"slot"`
	State   string `json:"state"`
	PID     int    `json:"pid,omitempty"`
	Age     string `json:"age,omitempty"`
	Command string `json:"command,omitempty"`
}

func (slotRow) Columns() []api.ColumnDef {
	var columns []api.ColumnDef
	for _, name := range []string{"Pool", "Use", "Project", "Slot", "State", "PID", "Age", "Command"} {
		columns = append(columns, clicky.Column(name).Label(name).Build())
	}
	return columns
}

func (r slotRow) Row() map[string]any {
	row := map[string]any{"Pool": r.Pool, "Use": r.Use, "Project": r.Project, "Slot": r.Slot, "State": r.State, "Age": r.Age, "Command": r.Command, "PID": ""}
	if r.PID > 0 {
		// A string, or the table abbreviates it as a quantity (70K).
		row["PID"] = strconv.Itoa(r.PID)
	}
	return row
}

func slotRows(pools []slots.PoolStatus, now time.Time) []slotRow {
	var rows []slotRow
	for _, pool := range pools {
		base := slotRow{Pool: pool.Name, Use: fmt.Sprintf("%d/%d", len(pool.Held), pool.Limit), Project: repoOf(pool.Project)}
		if len(pool.Waiting) > 0 {
			base.Use += fmt.Sprintf(" +%d waiting", len(pool.Waiting))
		}
		for _, held := range pool.Held {
			row := base
			row.Slot, row.State, row.PID, row.Command = fmt.Sprint(held.Index), "held", held.Owner.PID, held.Owner.Command
			row.Age = age(now, held.Owner.Started)
			rows = append(rows, row)
		}
		for _, waiter := range pool.Waiting {
			row := base
			row.Slot, row.State, row.PID, row.Command = "-", "waiting", waiter.PID, waiter.Lease.Command
			row.Age = age(now, waiter.Lease.Since)
			rows = append(rows, row)
		}
		if len(pool.Held)+len(pool.Waiting) == 0 {
			row := base
			row.Slot, row.State = "-", "idle"
			rows = append(rows, row)
		}
	}
	return rows
}

// repoOf is the checkout a git common directory belongs to.
func repoOf(common string) string {
	if filepath.Base(common) == ".git" {
		return filepath.Dir(common)
	}
	return common
}

func age(now, since time.Time) string {
	if since.IsZero() {
		return ""
	}
	return now.Sub(since).Round(time.Second).String()
}
