package query

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flanksource/commons-db/cmd/query/profiles"
	"github.com/flanksource/commons-db/connection/store"
	"github.com/flanksource/commons-db/models"
	cdbquery "github.com/flanksource/commons-db/query"
	"github.com/flanksource/commons-db/query/profilestore"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
)

// providerForConnection is the commons-db SQL provider key each connection
// type is read through.
var providerForConnection = map[string]string{
	models.ConnectionTypeSQLServer:  "sqlserver",
	models.ConnectionTypePostgres:   "postgres",
	models.ConnectionTypeMySQL:      "mysql",
	models.ConnectionTypeSQLite:     "sqlite",
	models.ConnectionTypeClickHouse: "clickhouse",
}

// providerFor is the SQL provider key connection is read through.
func providerFor(connection *models.Connection) (string, error) {
	if provider := providerForConnection[connection.Type]; provider != "" {
		return provider, nil
	}
	return "", fmt.Errorf("connection %q is a %s connection, which no query provider reads; set provider.type on the profile",
		connection.Name, connection.Type)
}

// LookupQuery is a lookup's SQL, the params it binds (`q`, the search text,
// among them) and the most rows it reads.
type LookupQuery struct {
	SQL    string
	Params map[string]any
	Limit  int
}

// bareParam matches `@name`, a lookup's shorthand for `{{.params.name}}`,
// not preceded by another `@` (a SQL Server `@@global`) or a word character.
var bareParam = regexp.MustCompile(`(^|[^@\w])@([A-Za-z_]\w*)\b`)

// LookupRows reads lookup.SQL on resolved through the same provider mapping
// and param binding a query step uses, refusing a SQL program: a lookup is one
// SELECT. `@name` is read as `{{.params.name}}` for each param supplied.
// truncated says Limit stopped the read short of what the query had to give.
func LookupRows(ctx context.Context, resolved *dbconn.Resolved, lookup LookupQuery) (rows []map[string]any, truncated bool, err error) {
	if strings.TrimSpace(lookup.SQL) == "" {
		return nil, false, errors.New("lookup sql is empty")
	}
	if lookup.Limit <= 0 {
		return nil, false, fmt.Errorf("lookup limit must be positive, got %d", lookup.Limit)
	}
	provider, err := providerFor(resolved.Connection)
	if err != nil {
		return nil, false, err
	}
	sql := bareParam.ReplaceAllStringFunc(lookup.SQL, func(match string) string {
		groups := bareParam.FindStringSubmatch(match)
		if _, supplied := lookup.Params[groups[2]]; !supplied {
			return match
		}
		return groups[1] + "{{.params." + groups[2] + "}}"
	})
	profile := cdbquery.Profile{Name: "lookup", Query: sql, Limits: &cdbquery.RowLimits{MaxExportRows: lookup.Limit}}
	profile.Provider.Type = provider
	profile.Provider.Connection = resolved.Ref
	if err := declareParams(&profile, lookup.Params); err != nil {
		return nil, false, err
	}
	result, err := cdbquery.Execute(resolved.Context.Wrap(ctx), profile, lookup.Params)
	if err != nil {
		return nil, false, err
	}
	if len(result.Batches) > 0 {
		return nil, false, errors.New("lookup sql ran as a SQL program: a lookup is a single SELECT")
	}
	return result.Rows, result.Truncated, nil
}

// buildProfile is the profile the step runs and the params it runs with: the
// query text, an inline profile body, or a stored profile with its imports
// resolved, every step param declared so it binds as a driver argument.
func buildProfile(ctx context.Context, step *fixtures.StepContext, cfg Config) (cdbquery.Profile, map[string]any, error) {
	profile, refParams, err := decodeQuery(ctx, step, cfg.Query)
	if err != nil {
		return cdbquery.Profile{}, nil, err
	}
	if profile.Name == "" {
		profile.Name = stepName(step)
	}
	params := make(map[string]any, len(refParams)+len(cfg.Params))
	for k, v := range refParams {
		params[k] = v
	}
	for k, v := range cfg.Params {
		params[k] = v
	}
	if err := declareParams(&profile, params); err != nil {
		return cdbquery.Profile{}, nil, err
	}
	if cfg.Limit > 0 {
		limits := cdbquery.RowLimits{}
		if profile.Limits != nil {
			limits = *profile.Limits
		}
		limits.MaxExportRows = cfg.Limit
		profile.Limits = &limits
	}
	return profile, params, nil
}

func decodeQuery(ctx context.Context, step *fixtures.StepContext, raw json.RawMessage) (cdbquery.Profile, map[string]any, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if text == "" {
			return cdbquery.Profile{}, nil, errors.New("query is empty")
		}
		return cdbquery.Profile{Query: text}, nil, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return cdbquery.Profile{}, nil, fmt.Errorf("query must be a string or an object: %w", err)
	}
	if !isProfileReference(object) {
		var profile cdbquery.Profile
		if err := decodeStrict(raw, &profile); err != nil {
			return cdbquery.Profile{}, nil, fmt.Errorf("inline query profile: %w", err)
		}
		if profile.QueryFile == "" {
			return profile, nil, nil
		}
		profile, err := profile.ResolveQueryFile(step.Test.SourceDir)
		if err != nil {
			return cdbquery.Profile{}, nil, fmt.Errorf("inline query profile: %w", err)
		}
		return profile, nil, nil
	}
	var ref struct {
		Profile string         `json:"profile"`
		Params  map[string]any `json:"params,omitempty"`
	}
	if err := decodeStrict(raw, &ref); err != nil {
		return cdbquery.Profile{}, nil, fmt.Errorf("query profile reference: %w", err)
	}
	layers, err := layeredProfiles(step)
	if err != nil {
		return cdbquery.Profile{}, nil, err
	}
	resolved, err := profiles.Resolve(ctx, layers, ref.Profile)
	if err != nil {
		return cdbquery.Profile{}, nil, err
	}
	return resolved.Profile, ref.Params, nil
}

// isProfileReference tells {profile: <name>, params: {…}} from an inline
// profile body, which names its query (or provider) and declares params as a
// list.
func isProfileReference(object map[string]json.RawMessage) bool {
	if _, ok := object["profile"]; !ok {
		return false
	}
	for key, value := range object {
		switch key {
		case "profile":
		case "params":
			if len(bytes.TrimSpace(value)) == 0 || bytes.TrimSpace(value)[0] != '{' {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// declareParams adds a string/number/boolean/list declaration for every
// supplied param the profile does not declare, so the SQL provider binds it.
func declareParams(profile *cdbquery.Profile, params map[string]any) error {
	declared := map[string]bool{}
	for _, def := range profile.Params {
		declared[def.Name] = true
	}
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if declared[name] {
			continue
		}
		paramType, err := inferParamType(name, params[name])
		if err != nil {
			return err
		}
		profile.Params = append(profile.Params, cdbquery.ParamDef{Name: name, Type: paramType})
	}
	return nil
}

func inferParamType(name string, value any) (cdbquery.ParamType, error) {
	switch value.(type) {
	case string:
		return cdbquery.ParamTypeString, nil
	case bool:
		return cdbquery.ParamTypeBoolean, nil
	case float64, int, int64, json.Number:
		return cdbquery.ParamTypeNumber, nil
	case []any:
		return cdbquery.ParamTypeList, nil
	}
	return "", fmt.Errorf("param %q is %T: a query param is a string, number, boolean or list", name, value)
}

// layeredProfiles reads stored profiles from the fixture directory's
// .profiles, then <config>/gavel/profiles, then the host's store; the first
// layer defining a name wins.
func layeredProfiles(step *fixtures.StepContext) (profilestore.Store, error) {
	config, err := store.ConfigDir(dbconn.DefaultApp)
	if err != nil {
		return nil, err
	}
	// FileStore literals, not NewFileStore: the constructor migrates legacy
	// files in place, and reading a fixture's profiles must write nothing.
	layers := profileLayers{
		&profiles.FileStore{Dir: filepath.Join(step.WorkDir, ".profiles")},
		&profiles.FileStore{Dir: filepath.Join(config, "profiles")},
	}
	if step.Opts.Profiles != nil {
		layers = append(layers, step.Opts.Profiles)
	}
	return layers, nil
}

// profileLayers is a read-only profilestore.Store over its layers.
type profileLayers []profilestore.Store

func (l profileLayers) List(ctx context.Context) ([]cdbquery.Profile, error) {
	seen := map[string]bool{}
	var all []cdbquery.Profile
	for _, layer := range l {
		items, err := layer.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, profile := range items {
			if !seen[profile.Name] {
				seen[profile.Name] = true
				all = append(all, profile)
			}
		}
	}
	return all, nil
}

func (l profileLayers) Get(ctx context.Context, name string) (cdbquery.Profile, error) {
	for _, layer := range l {
		items, err := layer.List(ctx)
		if err != nil {
			return cdbquery.Profile{}, err
		}
		for _, profile := range items {
			if profile.Name == name {
				return profile, nil
			}
		}
	}
	return cdbquery.Profile{}, fmt.Errorf("profile %q not found in .profiles, the gavel profiles dir or the host profile store", name)
}

var errReadOnlyProfiles = errors.New("query step profile stores are read-only")

func (profileLayers) Save(context.Context, cdbquery.Profile) error { return errReadOnlyProfiles }

func (profileLayers) Update(context.Context, string, cdbquery.Profile, profilestore.UpdateOptions) error {
	return errReadOnlyProfiles
}

func (profileLayers) Delete(context.Context, string) error { return errReadOnlyProfiles }

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
