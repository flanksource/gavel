package uirexport

import (
	"strings"
	"testing"
)

func TestEnvironmentCanonicalizesEffectiveGoIdentity(t *testing.T) {
	t.Setenv("CGO_ENABLED", "0")
	t.Setenv("GOFLAGS", "-count=2 -tags=z,a -race=false -shuffle=on")
	env, err := Environment(t.TempDir(), "-tags=b,a -race")
	if err != nil {
		t.Fatal(err)
	}
	if env["CGO_ENABLED"] != "0" || env["tags"] != "a,b" || env["race"] != "true" || env["count"] != "2" {
		t.Fatalf("environment = %#v", env)
	}
	if env["GOEXPERIMENT"] != canonicalList(env["GOEXPERIMENT"]) {
		t.Fatalf("GOEXPERIMENT is not canonical: %q", env["GOEXPERIMENT"])
	}
	for _, flag := range []string{"count=2", "race=true", "shuffle=on", "tags=a,b"} {
		if !strings.Contains(env["GOFLAGS"], flag) {
			t.Fatalf("GOFLAGS %q omits %q", env["GOFLAGS"], flag)
		}
	}
}
