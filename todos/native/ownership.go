package native

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/utils"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	// OwnerHeartbeatInterval is how often the dispatching process refreshes its
	// claim, and OwnerHeartbeatTTL is how long a claim survives without one.
	// Only a remote owner is judged by the heartbeat: a local one is settled
	// exactly by asking the OS whether its PID is still that process.
	OwnerHeartbeatInterval = 30 * time.Second
	OwnerHeartbeatTTL      = 3 * OwnerHeartbeatInterval

	// ownerStartSkew absorbs the difference between the process start time the
	// OS reports (millisecond truncation, and a value that is re-derived on
	// every read) and the one stored at claim time.
	ownerStartSkew = 2 * time.Second
)

// RunOwner identifies the process driving a prompt run. Host plus PID plus the
// process start time is the identity: PIDs are recycled, start times are not.
// Token distinguishes two dispatchers that somehow share the first three.
type RunOwner struct {
	HostID      string     `json:"hostId"`
	PID         int64      `json:"pid"`
	StartedAt   time.Time  `json:"startedAt"`
	Token       uuid.UUID  `json:"token"`
	HeartbeatAt *time.Time `json:"heartbeatAt,omitempty"`
}

var (
	localOwnerOnce sync.Once
	localOwner     RunOwner
	localOwnerErr  error
)

// LocalOwner is this process's dispatcher identity, resolved once. The token is
// minted per process so a recycled PID never reads as the same dispatcher.
func LocalOwner() (RunOwner, error) {
	localOwnerOnce.Do(func() {
		pid := os.Getpid()
		startedAt, err := processStartTime(pid)
		if err != nil {
			localOwnerErr = fmt.Errorf("resolve this process's start time: %w", err)
			return
		}
		localOwner = RunOwner{
			HostID:    captaindb.LocalHostID(),
			PID:       int64(pid),
			StartedAt: startedAt,
			Token:     uuid.New(),
		}
	})
	return localOwner, localOwnerErr
}

func processStartTime(pid int) (time.Time, error) {
	handle, err := process.NewProcess(int32(pid))
	if err != nil {
		return time.Time{}, err
	}
	createdMS, err := handle.CreateTime()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(createdMS).UTC(), nil
}

// Alive reports whether the owning process is still running, and why. An
// unowned run (never claimed, already released, or claimed before ownership
// was recorded at all) is not alive: nothing is driving it to completion.
func (o *RunOwner) Alive(now time.Time) (bool, string) {
	if o == nil || o.HostID == "" || o.PID <= 0 {
		return false, "no dispatcher claimed the run"
	}
	if o.HostID != captaindb.LocalHostID() {
		return o.heartbeatAlive(now)
	}
	if !utils.ProcessAlive(int(o.PID)) {
		return false, fmt.Sprintf("owner pid %d on %s has exited", o.PID, o.HostID)
	}
	startedAt, err := processStartTime(int(o.PID))
	if err != nil {
		// The PID answers to signal 0 but its start time is unreadable, so the
		// claim cannot be settled locally; the heartbeat is the only evidence left.
		return o.heartbeatAlive(now)
	}
	if delta := startedAt.Sub(o.StartedAt); delta > ownerStartSkew || delta < -ownerStartSkew {
		return false, fmt.Sprintf("pid %d on %s belongs to a different process now (started %s, claim recorded %s)",
			o.PID, o.HostID, startedAt.Format(time.RFC3339), o.StartedAt.Format(time.RFC3339))
	}
	return true, fmt.Sprintf("owner pid %d on %s is running", o.PID, o.HostID)
}

func (o *RunOwner) heartbeatAlive(now time.Time) (bool, string) {
	if o.HeartbeatAt == nil {
		return false, fmt.Sprintf("owner pid %d on %s never recorded a heartbeat", o.PID, o.HostID)
	}
	if age := now.Sub(*o.HeartbeatAt); age > OwnerHeartbeatTTL {
		return false, fmt.Sprintf("owner pid %d on %s last reported %s ago", o.PID, o.HostID, age.Round(time.Second))
	}
	return true, fmt.Sprintf("owner pid %d on %s reported %s ago", o.PID, o.HostID, now.Sub(*o.HeartbeatAt).Round(time.Second))
}

// Describe renders the owner for an operator-facing message.
func (o *RunOwner) Describe() string {
	if o == nil || o.HostID == "" {
		return "no owner"
	}
	return fmt.Sprintf("%s pid %d", o.HostID, o.PID)
}

// ClaimPromptRunOwner records this process as the driver of a linked run.
func (r *Repository) ClaimPromptRunOwner(ctx context.Context, promptRunID uuid.UUID, owner RunOwner) error {
	if promptRunID == uuid.Nil {
		return fmt.Errorf("%w: prompt run ID is required", ErrInvalidInput)
	}
	if owner.HostID == "" || owner.PID <= 0 || owner.StartedAt.IsZero() || owner.Token == uuid.Nil {
		return fmt.Errorf("%w: a run owner needs host, pid, start time and token", ErrInvalidInput)
	}
	result := r.db.WithContext(ctx).Exec(`
		UPDATE todo_issue_prompt_runs
		SET owner_host_id = ?, owner_pid = ?, owner_started_at = ?, owner_token = ?, owner_heartbeat_at = now()
		WHERE prompt_run_id = ?`,
		owner.HostID, owner.PID, owner.StartedAt.UTC(), owner.Token, promptRunID,
	)
	if result.Error != nil {
		return fmt.Errorf("claim prompt run %s: %w", promptRunID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: prompt run %s has no issue link to claim", ErrNotFound, promptRunID)
	}
	return nil
}

// TouchPromptRunOwner refreshes the heartbeat of a claim this process holds.
// It touches nothing when the claim has since moved to another process.
func (r *Repository) TouchPromptRunOwner(ctx context.Context, promptRunID uuid.UUID, token uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE todo_issue_prompt_runs
		SET owner_heartbeat_at = now()
		WHERE prompt_run_id = ? AND owner_token = ?`, promptRunID, token,
	)
	if result.Error != nil {
		return false, fmt.Errorf("refresh prompt run %s ownership: %w", promptRunID, result.Error)
	}
	return result.RowsAffected == 1, nil
}

// ReleasePromptRunOwner drops this process's claim once the run is finished, so
// the row reads as unowned rather than as an owner that stopped heartbeating.
func (r *Repository) ReleasePromptRunOwner(ctx context.Context, promptRunID uuid.UUID, token uuid.UUID) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE todo_issue_prompt_runs
		SET owner_host_id = NULL, owner_pid = NULL, owner_started_at = NULL,
		    owner_token = NULL, owner_heartbeat_at = NULL
		WHERE prompt_run_id = ? AND owner_token = ?`, promptRunID, token,
	)
	if result.Error != nil {
		return fmt.Errorf("release prompt run %s ownership: %w", promptRunID, result.Error)
	}
	return nil
}

// PromptRunOwner returns the claim recorded on a run's issue link, or nil when
// the run is unowned.
func (r *Repository) PromptRunOwner(ctx context.Context, promptRunID uuid.UUID) (*RunOwner, error) {
	var links []PromptRunLink
	result := r.db.WithContext(ctx).Raw(
		promptRunLinkSelect+` FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, promptRunID,
	).Scan(&links)
	if result.Error != nil {
		return nil, fmt.Errorf("read prompt run %s ownership: %w", promptRunID, result.Error)
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("%w: prompt run %s has no issue link", ErrNotFound, promptRunID)
	}
	return links[0].Owner(), nil
}
