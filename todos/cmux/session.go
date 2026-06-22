package cmux

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultPollInterval = 500 * time.Millisecond

type HookSession struct {
	SessionID   string `json:"session_id,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	SurfaceID   string `json:"surface_id,omitempty"`
	CWD         string `json:"cwd,omitempty"`
	PID         int    `json:"pid,omitempty"`
	Lifecycle   string `json:"lifecycle,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type SessionStore struct {
	Dir          string
	PollInterval time.Duration
}

func DefaultSessionStore() *SessionStore {
	return &SessionStore{}
}

func (s *SessionStore) WaitForSession(ctx context.Context, agent, cwd string, timeout time.Duration) (*HookSession, error) {
	return s.waitFor(ctx, agent, cwd, timeout, func(HookSession) bool { return true }, "a cmux hook session")
}

func (s *SessionStore) WaitForIdle(ctx context.Context, agent, cwd string, timeout time.Duration) (*HookSession, error) {
	return s.waitFor(ctx, agent, cwd, timeout, func(sess HookSession) bool {
		return strings.EqualFold(sess.Lifecycle, "idle")
	}, "lifecycle=idle")
}

func (s *SessionStore) LatestForCWD(agent, cwd string) (*HookSession, error) {
	sessions, err := s.Read(agent)
	if err != nil {
		return nil, err
	}
	want := cleanPath(cwd)
	var matches []HookSession
	for _, sess := range sessions {
		if cleanPath(sess.CWD) == want {
			matches = append(matches, sess)
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].UpdatedAt > matches[j].UpdatedAt
	})
	return &matches[0], nil
}

func (s *SessionStore) Read(agent string) ([]HookSession, error) {
	data, err := os.ReadFile(s.path(agent))
	if err != nil {
		return nil, err
	}
	return parseHookSessions(data)
}

func (s *SessionStore) path(agent string) string {
	dir := s.Dir
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, ".cmuxterm")
		} else {
			dir = ".cmuxterm"
		}
	}
	return filepath.Join(dir, agent+"-hook-sessions.json")
}

func (s *SessionStore) waitFor(ctx context.Context, agent, cwd string, timeout time.Duration, done func(HookSession) bool, want string) (*HookSession, error) {
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	poll := s.PollInterval
	if poll <= 0 {
		poll = defaultPollInterval
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	var sawSession bool
	for {
		sess, err := s.LatestForCWD(agent, cwd)
		if err == nil && sess != nil {
			sawSession = true
			if done(*sess) {
				return sess, nil
			}
		} else if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		select {
		case <-waitCtx.Done():
			if sawSession {
				return nil, fmt.Errorf("timed out waiting for cmux %s session in %s to reach %s", agent, cwd, want)
			}
			return nil, fmt.Errorf("timed out waiting for cmux %s hook session in %s; enable cmux agent hooks (Claude wrapper in cmux settings or `cmux hooks setup codex`)", agent, cwd)
		case <-ticker.C:
		}
	}
}

func parseHookSessions(data []byte) ([]HookSession, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var out []HookSession
	collectHookSessions(raw, "", &out)
	return out, nil
}

func collectHookSessions(raw any, key string, out *[]HookSession) {
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			collectHookSessions(item, "", out)
		}
	case map[string]any:
		if looksLikeSession(v) {
			sess := sessionFromMap(v)
			if sess.SessionID == "" {
				sess.SessionID = key
			}
			*out = append(*out, sess)
			return
		}
		for childKey, child := range v {
			collectHookSessions(child, childKey, out)
		}
	}
}

func looksLikeSession(m map[string]any) bool {
	return firstString(m, "cwd", "Cwd", "workdir", "workDir") != "" &&
		firstString(m, "lifecycle", "state", "status") != ""
}

func sessionFromMap(m map[string]any) HookSession {
	return HookSession{
		SessionID:   firstString(m, "session_id", "sessionId", "sessionID", "id"),
		WorkspaceID: firstString(m, "workspace_id", "workspaceId", "workspaceID", "workspace"),
		SurfaceID:   firstString(m, "surface_id", "surfaceId", "surfaceID", "surface"),
		CWD:         firstString(m, "cwd", "Cwd", "workdir", "workDir"),
		PID:         firstInt(m, "pid", "PID"),
		Lifecycle:   firstString(m, "lifecycle", "state", "status"),
		UpdatedAt:   firstString(m, "updated_at", "updatedAt", "lastSeenAt", "timestamp"),
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return strings.TrimSpace(typed)
				}
			case fmt.Stringer:
				if strings.TrimSpace(typed.String()) != "" {
					return strings.TrimSpace(typed.String())
				}
			case float64:
				return strconv.FormatInt(int64(typed), 10)
			}
		}
	}
	return ""
}

func firstInt(m map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			switch typed := value.(type) {
			case float64:
				return int(typed)
			case string:
				if n, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func cleanPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}
