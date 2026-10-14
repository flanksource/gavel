// Authorized pushes for the hosted push server: SSH keys identify users, an Authorizer
// decides where they may push, and received refs are reported without running any hook.
package serve

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

// Authorizer maps SSH keys to users and decides who may push to which repository.
type Authorizer interface {
	UserForKey(ctx context.Context, key ssh.PublicKey) (uuid.UUID, bool)
	AuthorizePush(ctx context.Context, user uuid.UUID, repoPath string) (Target, error)
}

// Target is where an authorized push lands.
type Target struct {
	OrgID   uuid.UUID
	RepoID  uuid.UUID
	BareDir string
}

// RefUpdate is one ref a push changed. Old is empty for a created ref, New for a deleted one.
type RefUpdate struct {
	Name, Old, New string
}

// Push is a received push. Out is the pusher's terminal, for messages about the push.
type Push struct {
	OrgID, RepoID, UserID uuid.UUID
	Refs                  []RefUpdate
	Out                   io.Writer
}

// pushUserKey holds the user an Authorizer resolved the session's key to.
type pushUserKey struct{}

// authorizeKey resolves the key's user and keeps it on the connection.
func (s *Server) authorizeKey(ctx ssh.Context, key ssh.PublicKey) bool {
	user, ok := s.opts.Authorizer.UserForKey(ctx, key)
	if ok {
		ctx.SetValue(pushUserKey{}, user)
	}
	return ok
}

// receiveAuthorized receives a push into the repository the Authorizer chose, with hooks
// disabled, and reports the refs that changed to OnPushed.
func (s *Server) receiveAuthorized(sess ssh.Session, repoPath string) int {
	user, _ := sess.Context().Value(pushUserKey{}).(uuid.UUID)
	target, err := s.opts.Authorizer.AuthorizePush(sess.Context(), user, repoPath)
	if err != nil {
		logger.Infof("Refused push of %s by %s: %v", repoPath, user, err)
		// Deliberately the same for a missing repository and a forbidden one.
		fmt.Fprintln(sess.Stderr(), "repository not found or access denied")
		return 1
	}
	if err := ensureBareRepo(target.BareDir); err != nil {
		fmt.Fprintf(sess.Stderr(), "failed to init repo: %v\n", err)
		return 1
	}
	before, err := refs(target.BareDir)
	if err != nil {
		fmt.Fprintf(sess.Stderr(), "failed to read refs: %v\n", err)
		return 1
	}
	// No hook in the repository runs on the host: pushed code is never executed here.
	cmd := exec.Command("git", "-c", "core.hooksPath="+os.DevNull, "receive-pack", target.BareDir)
	cmd.Stdin = sess
	cmd.Stdout = sess
	cmd.Stderr = sess.Stderr()
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(sess.Stderr(), "git receive-pack failed: %v\n", err)
		return 1
	}
	after, err := refs(target.BareDir)
	if err != nil {
		fmt.Fprintf(sess.Stderr(), "failed to read refs: %v\n", err)
		return 1
	}
	if s.opts.OnPushed != nil {
		s.opts.OnPushed(sess.Context(), Push{OrgID: target.OrgID, RepoID: target.RepoID, UserID: user, Refs: changedRefs(before, after), Out: sess.Stderr()})
	}
	return 0
}

// refs reads every ref of a bare repository and the object it points at.
func refs(bareDir string) (map[string]string, error) {
	out, err := exec.Command("git", "--git-dir", filepath.Clean(bareDir), "for-each-ref", "--format=%(refname) %(objectname)").Output()
	if err != nil {
		return nil, err
	}
	refs := map[string]string{}
	for _, line := range strings.Split(string(bytes.TrimSpace(out)), "\n") {
		if name, object, ok := strings.Cut(line, " "); ok {
			refs[name] = object
		}
	}
	return refs, nil
}

// changedRefs lists the refs created, moved or deleted between two snapshots, by name.
func changedRefs(before, after map[string]string) []RefUpdate {
	var changed []RefUpdate
	for name, object := range after {
		if before[name] != object {
			changed = append(changed, RefUpdate{Name: name, Old: before[name], New: object})
		}
	}
	for name, object := range before {
		if _, ok := after[name]; !ok {
			changed = append(changed, RefUpdate{Name: name, Old: object})
		}
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].Name < changed[j].Name })
	return changed
}
