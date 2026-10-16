// Tests for carrying the organization and user scope on a context.
// A context without a scope must resolve to the local organization.
package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestFromContextWithoutScopeIsLocal(t *testing.T) {
	scope := FromContext(context.Background())
	if scope != Local() || scope.OrgID != LocalOrgID || !scope.IsLocal() {
		t.Fatalf("got %+v, want the local scope", scope)
	}
	if LocalOrgID.String() != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("local org id changed: %s", LocalOrgID)
	}
}

func TestScopeRoundTrip(t *testing.T) {
	user := uuid.New()
	want := Scope{OrgID: uuid.New(), UserID: &user, Role: "org:admin", GitHubLogin: "octocat"}
	got := FromContext(WithScope(context.Background(), want))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if got.IsLocal() {
		t.Fatal("a non-local org reported IsLocal")
	}
}
