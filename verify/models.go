package verify

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	chttp "github.com/flanksource/commons/http"
)

type modelListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func fetchModelIDs(url, authHeader, authValue, version string) ([]string, error) {
	req := chttp.NewClient().R(context.Background()).
		Header(authHeader, authValue)
	if version != "" {
		req = req.Header("anthropic-version", version)
	}

	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !resp.IsOK() {
		return nil, fmt.Errorf("models API returned %d", resp.StatusCode)
	}

	var body modelListResponse
	if err := resp.Into(&body); err != nil {
		return nil, err
	}

	var ids []string
	for _, m := range body.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	for _, m := range body.Models {
		name := strings.TrimPrefix(m.Name, "models/")
		if name != "" {
			ids = append(ids, name)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func formatModelHint(adapter Adapter) string {
	models, err := adapter.ListModels()
	if err != nil || len(models) == 0 {
		return ""
	}
	return fmt.Sprintf("Available %s models: %s", adapter.Name(), strings.Join(models, ", "))
}

func containsModelError(msg string) bool {
	lower := strings.ToLower(msg)
	for _, kw := range []string{"model", "not_found", "not found", "invalid", "does not exist"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func getEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
