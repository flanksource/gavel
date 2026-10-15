// `gavel ci upload`: sends a results file from a GitHub Actions job to a hosted gavel,
// authenticated by the job's OIDC token for that server.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	ciCmd := &cobra.Command{Use: "ci", Short: "Work with a hosted gavel from CI"}
	var server, file string
	upload := &cobra.Command{
		Use:   "upload",
		Short: "Upload a gavel results file from GitHub Actions to a hosted gavel",
		Long: "Upload a gavel results file (gavel test --json-file) from a GitHub Actions job. The job " +
			"authenticates with its OIDC token, so the workflow needs permissions: id-token: write, and the " +
			"repository must be linked to a gavel organization through the gavel GitHub App.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if server == "" || file == "" {
				return errors.New("--server and --file are required")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			client := &http.Client{Timeout: time.Minute}
			token, err := requestActionsOIDCToken(ctx, client, os.Getenv, strings.TrimRight(server, "/"))
			if err != nil {
				return err
			}
			runID, err := uploadCIResults(ctx, client, server, token, file)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uploaded %s as run %s.\n", file, runID)
			return nil
		},
	}
	upload.Flags().StringVar(&server, "server", "", "The hosted gavel's URL, which is also the OIDC token's audience")
	upload.Flags().StringVar(&file, "file", "", "The results JSON to upload")
	ciCmd.AddCommand(upload)
	rootCmd.AddCommand(ciCmd)
}

// oidcTokenRequestURL is GitHub's token request URL asking for a token with audience.
func oidcTokenRequestURL(requestURL, audience string) (string, error) {
	parsed, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("parse ACTIONS_ID_TOKEN_REQUEST_URL: %w", err)
	}
	query := parsed.Query()
	query.Set("audience", audience)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

// requestActionsOIDCToken asks GitHub Actions for this job's OIDC token for audience.
func requestActionsOIDCToken(ctx context.Context, client *http.Client, getenv func(string) string, audience string) (string, error) {
	requestURL, requestToken := getenv("ACTIONS_ID_TOKEN_REQUEST_URL"), getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if requestURL == "" || requestToken == "" {
		return "", errors.New("ACTIONS_ID_TOKEN_REQUEST_URL and ACTIONS_ID_TOKEN_REQUEST_TOKEN are not set: run in GitHub Actions with permissions: id-token: write")
	}
	target, err := oidcTokenRequestURL(requestURL, audience)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+requestToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request the job's OIDC token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request the job's OIDC token: HTTP %d", resp.StatusCode)
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Value == "" {
		return "", errors.New("request the job's OIDC token: GitHub returned no token")
	}
	return body.Value, nil
}

// uploadCIResults posts the results file to server with token and returns the stored run's ID.
func uploadCIResults(ctx context.Context, client *http.Client, server, token, file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read the results: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(server, "/")+"/api/ci/results", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", file, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", file, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated {
		var failure struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &failure) == nil && failure.Error != "" {
			return "", fmt.Errorf("upload %s: HTTP %d: %s", file, resp.StatusCode, failure.Error)
		}
		return "", fmt.Errorf("upload %s: HTTP %d", file, resp.StatusCode)
	}
	var created struct {
		RunID string `json:"runId"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return "", fmt.Errorf("upload %s: the server's reply isn't JSON: %w", file, err)
	}
	return created.RunID, nil
}
