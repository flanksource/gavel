package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/flanksource/clicky/schemaform"

	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/query"
)

// Roles are the option fields a lookup query's columns fill, each read from
// the column of its own name unless QueryBinding.Columns renames it.
var Roles = []string{"value", "label", "description", "icon", "badges", "tone"}

// SearchParam is the param a lookup query binds the search text to (`@q`).
const SearchParam = "q"

// QueryBinding answers an `x-query` field by running SQL — `@q` bound to the
// search text — on one connection: the value of the linked sibling property
// at From (resolved through Connections), or Connection when From is unset.
type QueryBinding struct {
	SQL         string
	Columns     map[string]string
	From        []string
	Connection  *dbconn.Resolved
	Connections *dbconn.Connections
}

func (b QueryBinding) validate() error {
	switch {
	case strings.TrimSpace(b.SQL) == "":
		return errors.New("x-query.sql is required")
	case len(b.From) > 0 && b.Connections == nil:
		return errors.New("a linked x-query needs the fixture's connections to resolve its param")
	case len(b.From) == 0 && b.Connection == nil:
		return errors.New("x-query has no connection: link a param or resolve one")
	case len(b.From) > 0 && b.Connection != nil:
		return errors.New("x-query has both a linked param and a fixed connection")
	}
	for role := range b.Columns {
		if !slices.Contains(Roles, role) {
			return fmt.Errorf("x-query.columns.%s: not an option field (%s)", role, strings.Join(Roles, ", "))
		}
	}
	return nil
}

func (b QueryBinding) resolve(ctx context.Context, req Request) (Result, error) {
	resolved := b.Connection
	if len(b.From) > 0 {
		ref := linkedValue(req, b.From)
		if ref == "" {
			return Result{Label: "Choose " + strings.Join(b.From, ".") + " first"}, nil
		}
		var err error
		if resolved, err = b.Connections.Resolve(ref); err != nil {
			return Result{}, fmt.Errorf("x-query connection %s=%q: %w", strings.Join(b.From, "."), ref, err)
		}
	}
	rows, truncated, err := query.LookupRows(ctx, resolved, query.LookupQuery{
		SQL: b.SQL, Params: map[string]any{SearchParam: req.Query}, Limit: MaxOptions,
	})
	if err != nil {
		return Result{}, fmt.Errorf("x-query on %s: %w", resolved.Ref, err)
	}
	options, err := RowOptions(rows, b.Columns)
	if err != nil {
		return Result{}, err
	}
	return Result{LookupResult: schemaform.LookupResult{Options: options, Truncated: truncated, Total: len(options)}}, nil
}

// linkedValue is the linked param's value: the browser's `connection` scope
// param, else the terminal form's value at path.
func linkedValue(req Request, path []string) string {
	if scoped, ok := req.Scope["connection"]; ok {
		text, _ := scoped.(string)
		return text
	}
	var value any = req.Values
	for _, key := range path {
		object, _ := value.(map[string]any)
		value = object[key]
	}
	text, _ := value.(string)
	return text
}

// RowOptions maps lookup rows to options: the `value` column (else the one
// column no other role reads) is the value; label (else the value),
// description, icon and tone are text; badges is a string, a comma list, or
// a JSON array of strings or {label, icon?, tone?} objects. columns renames a
// role's column; a renamed column a row lacks is an error.
func RowOptions(rows []map[string]any, columns map[string]string) ([]schemaform.Option, error) {
	options := make([]schemaform.Option, 0, len(rows))
	for i, row := range rows {
		option, err := rowOption(row, columns)
		if err != nil {
			return nil, fmt.Errorf("lookup row %d: %w", i, err)
		}
		options = append(options, option)
	}
	return options, nil
}

func rowOption(row map[string]any, columns map[string]string) (schemaform.Option, error) {
	if len(row) == 0 {
		return schemaform.Option{}, errors.New("the query returned no columns")
	}
	cells := map[string]any{}
	for _, role := range Roles {
		column := role
		if renamed, ok := columns[role]; ok {
			if _, present := row[renamed]; !present {
				return schemaform.Option{}, fmt.Errorf("x-query.columns.%s names column %q, which the query does not return (%s)", role, renamed, strings.Join(sortedKeys(row), ", "))
			}
			column = renamed
		}
		if value, present := row[column]; present {
			cells[role] = value
		}
	}
	if _, ok := cells["value"]; !ok {
		value, err := soleColumn(row, columns)
		if err != nil {
			return schemaform.Option{}, err
		}
		cells["value"] = value
	}
	return cellOption(cells)
}

// soleColumn is the value of a row with no `value` column: the one column no
// other role reads. Rows carry no column order, so several such columns are
// ambiguous rather than "the first".
func soleColumn(row map[string]any, columns map[string]string) (any, error) {
	read := map[string]bool{}
	for _, role := range Roles {
		read[role] = true
		if renamed, ok := columns[role]; ok {
			read[renamed] = true
		}
	}
	var candidates []string
	for _, key := range sortedKeys(row) {
		if !read[key] {
			candidates = append(candidates, key)
		}
	}
	if len(candidates) != 1 {
		return nil, fmt.Errorf("no value column among %s: select one `as value` or name it under x-query.columns.value", strings.Join(sortedKeys(row), ", "))
	}
	return row[candidates[0]], nil
}

func cellOption(cells map[string]any) (schemaform.Option, error) {
	text := map[string]string{}
	for _, role := range []string{"value", "label", "description", "icon", "tone"} {
		value, err := cellText(cells[role])
		if err != nil {
			return schemaform.Option{}, fmt.Errorf("%s: %w", role, err)
		}
		text[role] = value
	}
	if cells["value"] == nil {
		return schemaform.Option{}, errors.New("value is null")
	}
	badges, err := parseBadges(cells["badges"])
	if err != nil {
		return schemaform.Option{}, fmt.Errorf("badges: %w", err)
	}
	label := text["label"]
	if label == "" {
		label = text["value"]
	}
	return schemaform.Option{Value: text["value"], Label: label, Description: text["description"],
		Icon: text["icon"], Tone: text["tone"], Badges: badges}, nil
}

func cellText(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return fmt.Sprint(v), nil
	}
	return "", fmt.Errorf("want text, got %T(%v)", value, value)
}

// parseBadges reads a badges cell: a string (a JSON array when it starts with
// `[`, else a comma list), or a driver's already-decoded list.
func parseBadges(value any) ([]schemaform.Badge, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		return parseBadges(string(v))
	case []any:
		return badgeList(v)
	case string:
		text := strings.TrimSpace(v)
		if strings.HasPrefix(text, "[") {
			var items []any
			if err := json.Unmarshal([]byte(text), &items); err != nil {
				return nil, fmt.Errorf("JSON array %q: %w", text, err)
			}
			return badgeList(items)
		}
		var badges []schemaform.Badge
		for _, label := range strings.Split(text, ",") {
			if label = strings.TrimSpace(label); label != "" {
				badges = append(badges, schemaform.Badge{Label: label})
			}
		}
		return badges, nil
	}
	return nil, fmt.Errorf("want a string or a list, got %T(%v)", value, value)
}

func badgeList(items []any) ([]schemaform.Badge, error) {
	badges := make([]schemaform.Badge, 0, len(items))
	for i, item := range items {
		badge, err := badgeOf(item)
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		badges = append(badges, badge)
	}
	return badges, nil
}

func badgeOf(item any) (schemaform.Badge, error) {
	switch v := item.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return schemaform.Badge{}, errors.New("empty label")
		}
		return schemaform.Badge{Label: v}, nil
	case map[string]any:
		fields := map[string]string{}
		for key, raw := range v {
			if key != "label" && key != "icon" && key != "tone" {
				return schemaform.Badge{}, fmt.Errorf("unknown badge field %q (label, icon, tone)", key)
			}
			text, ok := raw.(string)
			if !ok {
				return schemaform.Badge{}, fmt.Errorf("%s must be a string, got %T(%v)", key, raw, raw)
			}
			fields[key] = text
		}
		if strings.TrimSpace(fields["label"]) == "" {
			return schemaform.Badge{}, errors.New("a badge object needs a label")
		}
		return schemaform.Badge{Label: fields["label"], Icon: fields["icon"], Tone: fields["tone"]}, nil
	}
	return schemaform.Badge{}, fmt.Errorf("want a label or {label, icon?, tone?}, got %T(%v)", item, item)
}

func sortedKeys(row map[string]any) []string {
	keys := make([]string, 0, len(row))
	for key := range row {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
