package jira

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anmitsu/go-shlex"
)

type Credentials struct {
	Username    string
	APIToken    string `json:"-"`
	AccessToken string `json:"-"`
	CloudID     string
}

var twgVariable = regexp.MustCompile(`^TWG_[A-Z0-9_]+$`)

func credentialsFromEnv(getenv func(string) string) (Credentials, bool, error) {
	cloudID := firstEnv(getenv, "JIRA_CLOUD_ID", "ATLASSIAN_CLOUD_ID")
	for _, group := range []struct{ oauth, token, user []string }{
		{[]string{"JIRA_OAUTH_ACCESS_TOKEN"}, []string{"JIRA_API_TOKEN", "JIRA_TOKEN"}, []string{"JIRA_EMAIL", "JIRA_USER", "JIRA_USERNAME"}},
		{[]string{"ATLASSIAN_OAUTH_ACCESS_TOKEN"}, []string{"ATLASSIAN_API_TOKEN"}, []string{"ATLASSIAN_EMAIL", "ATLASSIAN_USER"}},
		{[]string{"TWG_OAUTH_ACCESS_TOKEN"}, []string{"TWG_TOKEN"}, []string{"TWG_USER"}},
	} {
		if token := firstEnv(getenv, group.oauth...); token != "" {
			return Credentials{AccessToken: token, CloudID: cloudID}, true, nil
		}
		if token := firstEnv(getenv, group.token...); token != "" {
			user := firstEnv(getenv, group.user...)
			if user == "" {
				return Credentials{}, true, fmt.Errorf("%s requires %s", group.token[0], strings.Join(group.user, " or "))
			}
			return Credentials{Username: user, APIToken: token, CloudID: cloudID}, true, nil
		}
		if firstEnv(getenv, group.user...) != "" {
			return Credentials{}, true, fmt.Errorf("%s requires %s", group.user[0], group.token[0])
		}
	}
	return Credentials{}, false, nil
}

func parseTWGExport(output string) (map[string]string, error) {
	values := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if assignment, ok := strings.CutPrefix(line, "export "); ok {
			key, literal, assigned := strings.Cut(assignment, "=")
			if assigned && twgVariable.MatchString(key) && strings.HasPrefix(literal, "$'") {
				value, err := decodeTWGLiteral(literal)
				if err != nil {
					return nil, err
				}
				values[key] = value
				continue
			}
		}
		words, err := shlex.Split(line, true)
		if err == nil && len(words) == 5 && words[0] == "unset" && twgVariable.MatchString(words[1]) && words[2] == "2>/dev/null" && words[3] == "||" && words[4] == "true" {
			delete(values, words[1])
			continue
		}
		if err != nil || len(words) != 2 || words[0] != "export" {
			return nil, fmt.Errorf("invalid TWG credential export statement")
		}
		key, value, ok := strings.Cut(words[1], "=")
		if !ok || !twgVariable.MatchString(key) {
			return nil, fmt.Errorf("invalid TWG credential export assignment")
		}
		values[key] = value
	}
	return values, nil
}

func decodeTWGLiteral(literal string) (string, error) {
	if !strings.HasSuffix(literal, "'") || len(literal) < 3 {
		return "", fmt.Errorf("invalid TWG credential export literal")
	}
	var value strings.Builder
	for remaining := literal[2 : len(literal)-1]; remaining != ""; {
		decoded, multibyte, rest, err := strconv.UnquoteChar(remaining, '\'')
		if err != nil {
			return "", fmt.Errorf("invalid TWG credential export literal")
		}
		if multibyte {
			value.WriteRune(decoded)
		} else {
			value.WriteByte(byte(decoded))
		}
		remaining = rest
	}
	return value.String(), nil
}

func firstEnv(getenv func(string) string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func discoverCredentials(ctx context.Context, options ClientOptions) (Credentials, string, error) {
	site := options.Site
	if site == "" {
		site = firstEnv(os.Getenv, "JIRA_URL", "JIRA_BASE_URL", "JIRA_SITE", "ATLASSIAN_SITE", "ATLASSIAN_URL", "TWG_SITE", "TWG_DOMAIN")
	}
	if options.Credentials.APIToken != "" || options.Credentials.AccessToken != "" || options.Credentials.Username != "" {
		return options.Credentials, site, nil
	}
	if options.Profile == "" {
		credentials, found, err := credentialsFromEnv(os.Getenv)
		if found || err != nil {
			return credentials, site, err
		}
	}
	values, err := exportTWGCredentials(ctx, options)
	if err != nil {
		return Credentials{}, "", err
	}
	credentials, found, err := credentialsFromEnv(func(key string) string { return values[key] })
	if err != nil {
		return Credentials{}, "", err
	}
	if !found {
		return Credentials{}, "", fmt.Errorf("TWG exported no Jira credentials; configure JIRA_API_TOKEN and JIRA_EMAIL or run twg login")
	}
	credentials.CloudID = firstEnv(os.Getenv, "JIRA_CLOUD_ID", "ATLASSIAN_CLOUD_ID")
	if site == "" {
		site = values["TWG_SITE"]
	}
	return credentials, site, nil
}

func exportTWGCredentials(ctx context.Context, options ClientOptions) (map[string]string, error) {
	binary, err := exec.LookPath("twg")
	if err != nil {
		return nil, fmt.Errorf("Jira credentials not configured: set JIRA_API_TOKEN and JIRA_EMAIL, or install and authenticate twg: %w", err)
	}
	args := []string{"env"}
	if options.Site != "" {
		args = append(args, "--site", options.Site)
	}
	if options.Profile != "" {
		args = append(args, "--profile", options.Profile)
	}
	args = append(args, "--exec", "--", binary, "env", "--export")
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, binary, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("discover Jira credentials from TWG: %w", err)
	}
	return parseTWGExport(string(output))
}
