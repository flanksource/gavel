package native_test

import (
	"os"
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deadPID is a PID that cannot be running: the kernel allocates from 1 upwards
// and never hands out this one while the test process itself is alive.
const deadPID = 0x7FFFFFFE

func TestRunOwnerAliveness(t *testing.T) {
	self, err := native.LocalOwner()
	require.NoError(t, err)
	require.Equal(t, int64(os.Getpid()), self.PID)

	now := time.Now()
	fresh := now.Add(-native.OwnerHeartbeatInterval)
	expired := now.Add(-native.OwnerHeartbeatTTL - time.Minute)
	remote := func(heartbeat *time.Time) *native.RunOwner {
		return &native.RunOwner{
			HostID: "some-other-host", PID: 4242,
			StartedAt: now.Add(-time.Hour), Token: uuid.New(), HeartbeatAt: heartbeat,
		}
	}

	tests := []struct {
		name  string
		owner *native.RunOwner
		alive bool
	}{
		{name: "a run nothing ever claimed is not alive", owner: nil},
		{
			name:  "a claim without a host is not alive",
			owner: &native.RunOwner{PID: int64(os.Getpid()), StartedAt: now, Token: uuid.New()},
		},
		{
			name:  "this process is alive",
			owner: &self,
			alive: true,
		},
		{
			name: "a local pid that is not running is not alive",
			owner: &native.RunOwner{
				HostID: captaindb.LocalHostID(), PID: deadPID,
				StartedAt: now.Add(-time.Hour), Token: uuid.New(),
			},
		},
		{
			name: "a recycled local pid is not alive: the start time does not match",
			owner: &native.RunOwner{
				HostID: captaindb.LocalHostID(), PID: int64(os.Getpid()),
				StartedAt: self.StartedAt.Add(-24 * time.Hour), Token: uuid.New(),
			},
		},
		{name: "a remote owner with a fresh heartbeat is alive", owner: remote(&fresh), alive: true},
		{name: "a remote owner whose heartbeat expired is not alive", owner: remote(&expired)},
		{name: "a remote owner that never reported is not alive", owner: remote(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alive, reason := test.owner.Alive(now)
			assert.Equal(t, test.alive, alive, "reason: %s", reason)
			assert.NotEmpty(t, reason, "an ownership verdict must say why")
		})
	}
}

func TestPromptRunLinkOwner(t *testing.T) {
	host, pid, startedAt, token := "host-a", int64(321), time.Now().UTC(), uuid.New()

	t.Run("an unclaimed link has no owner", func(t *testing.T) {
		assert.Nil(t, native.PromptRunLink{}.Owner())
	})

	t.Run("a partially written claim is not an owner", func(t *testing.T) {
		assert.Nil(t, native.PromptRunLink{OwnerHostID: &host, OwnerPID: &pid}.Owner())
	})

	t.Run("a complete claim round-trips", func(t *testing.T) {
		owner := native.PromptRunLink{
			OwnerHostID: &host, OwnerPID: &pid,
			OwnerStartedAt: &startedAt, OwnerToken: &token,
		}.Owner()
		require.NotNil(t, owner)
		assert.Equal(t, native.RunOwner{HostID: host, PID: pid, StartedAt: startedAt, Token: token}, *owner)
		assert.Equal(t, "host-a pid 321", owner.Describe())
	})
}
