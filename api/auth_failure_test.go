package api

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"golang.org/x/oauth2"
)

func TestDeviceDenialDoesNotLeakCredentials(t *testing.T) {
	server := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	server.DenyLogin = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := NewAuthClient(ctx, AuthConfig{Issuer: server.URL, ClientID: "cli"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.BeginLogin(ctx)
	if err == nil || strings.Contains(err.Error(), clerktest.DeviceCode) {
		t.Fatalf("unsafe device error: %v", err)
	}
	if want := "device authorization with Clerk failed; check the public OAuth application and Device Authorization Grant settings"; err.Error() != want {
		t.Fatalf("device error = %q, want %q", err, want)
	}
	device := &oauth2.DeviceAuthResponse{DeviceCode: "secret-device-code", Expiry: time.Now().Add(time.Minute), Interval: 1}
	if _, _, err := client.CompleteLogin(ctx, device); err == nil || strings.Contains(err.Error(), "secret-device-code") {
		t.Fatalf("unsafe polling error: %v", err)
	}
	cancelled, stop := context.WithCancel(ctx)
	stop()
	if _, _, err := client.CompleteLogin(cancelled, device); err == nil {
		t.Fatal("cancelled login succeeded")
	}
}
