package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/azuredevops"
)

func PRFromURL(ref string) (*PR, error) {
	u, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return nil, fmt.Errorf("invalid PR URL: %w", err)
	}
	if u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("PR link requires an HTTPS pull request URL without credentials, query or fragment")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	valid := u.Host == "github.com" && len(parts) == 4 && parts[0] != "" && parts[1] != "" && parts[2] == "pull"
	if u.Host == "dev.azure.com" || strings.HasSuffix(u.Host, ".visualstudio.com") {
		_, _, err := azuredevops.ParsePRURL(strings.TrimSuffix(u.String(), "/"))
		valid = err == nil
	}
	if !valid {
		return nil, fmt.Errorf("unsupported PR URL %q", ref)
	}
	number, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || number <= 0 {
		return nil, fmt.Errorf("invalid PR number in URL %q", ref)
	}
	return &PR{Number: number, URL: strings.TrimSuffix(u.String(), "/")}, nil
}

func (p *PR) Validate() error {
	if p == nil {
		return nil
	}
	parsed, err := PRFromURL(p.URL)
	if err != nil {
		return err
	}
	if p.Number != parsed.Number {
		return fmt.Errorf("PR number %d does not match URL number %d", p.Number, parsed.Number)
	}
	return nil
}

func (p PR) Value() (driver.Value, error) {
	data, err := json.Marshal(p)
	return string(data), err
}

func (p *PR) Scan(value any) error {
	switch data := value.(type) {
	case []byte:
		return json.Unmarshal(data, p)
	case string:
		return json.Unmarshal([]byte(data), p)
	default:
		return fmt.Errorf("cannot scan PR link from %T", value)
	}
}
