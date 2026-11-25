package dbconn

import (
	"regexp"
	"strings"
	"sync"

	"github.com/flanksource/clicky/schemaform"
	"github.com/flanksource/commons-db/models"
	cdbschema "github.com/flanksource/commons-db/query/schema"
)

// Trait badge icons, runtime names the clicky-ui icon provider resolves.
const (
	viaIcon   = "ph:link-thin"
	proxyIcon = "ph:plugs-connected-thin"
	svcIcon   = "ph:network-thin"
)

// endpointURL is `scheme://rest`; serviceHost an in-cluster service name
// (`svc.ns`, `svc.ns.svc`, `svc.ns.svc.cluster.local`), not a public DNS name.
var (
	endpointURL = regexp.MustCompile(`(?i)^([a-z][a-z0-9+.-]*)://(.*)$`)
	serviceHost = regexp.MustCompile(`(?i)^[a-z0-9-]+\.[a-z0-9-]+(\.svc(\.cluster\.local)?)?$`)
)

// databaseTypes are the connection types whose endpoint path names a database.
var databaseTypes = map[string]bool{
	models.ConnectionTypePostgres: true, models.ConnectionTypeSQLServer: true,
	models.ConnectionTypeClickHouse: true, models.ConnectionTypeMySQL: true,
}

// typeIcons is commons-db's connection type -> runtime icon name map, the
// connection form's `x-enum-icons`.
var typeIcons = sync.OnceValue(func() map[string]string {
	props, _ := cdbschema.ConnectionSource()["properties"].(cdbschema.Schema)
	field, _ := props["type"].(cdbschema.Schema)
	icons, _ := field["x-enum-icons"].(map[string]string)
	return icons
})

// TypeIcon is the runtime icon name of a connection type: commons-db's, else
// the type itself, which is how the clicky-ui picker names an unmapped type.
func TypeIcon(connectionType string) string {
	if icon := typeIcons()[connectionType]; icon != "" {
		return icon
	}
	return connectionType
}

// TypeLabel is how a connection type reads as a group heading: `sql_server`
// is `Sql Server`, as clicky-ui titleizes a type its schema does not label.
func TypeLabel(connectionType string) string {
	words := strings.Split(connectionType, "_")
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// PickerOption is connection as a picker lists it: its connection:// value,
// `name (type)` label, type icon and group, the endpoint it reaches (host,
// port and database — never credentials) as the description, and how it is
// reached (proxy, svc, via <connection>) as badges. all are the connections
// listed beside it, which a `connection://` delegation is followed through.
func PickerOption(connection *models.Connection, all []*models.Connection) schemaform.Option {
	description, badges := targetOf(connection, all, map[string]bool{})
	return schemaform.Option{
		Value: PickerValue(connection), Label: PickerLabel(connection), Icon: TypeIcon(connection.Type),
		Description: description, Badges: badges, Group: TypeLabel(connection.Type),
	}
}

func targetOf(connection *models.Connection, all []*models.Connection, seen map[string]bool) (string, []schemaform.Badge) {
	seen[connection.Name] = true
	delegated := strings.TrimPrefix(connection.Properties["connection"], refPrefix)
	if connection.URL == "" && delegated != "" {
		via := []schemaform.Badge{{Label: "via " + delegated, Icon: viaIcon, Tone: "indigo"}}
		for _, target := range all {
			if target.Name == delegated && !seen[target.Name] {
				primary, badges := targetOf(target, all, seen)
				return primary, append(via, badges...)
			}
		}
		return delegated, via
	}
	point, ok := endpointOf(connection.URL)
	if !ok {
		if connection.Type == models.ConnectionTypeKubernetes {
			return "current kube context", nil
		}
		return "", nil
	}
	var badges []schemaform.Badge
	hostname := point.hostname()
	if point.scheme == "portforward" {
		badges = append(badges, schemaform.Badge{Label: "proxy", Icon: proxyIcon, Tone: "violet"})
	}
	if point.scheme == "host" || serviceHost.MatchString(hostname) {
		badges = append(badges, schemaform.Badge{Label: "svc", Icon: svcIcon, Tone: "teal"})
	}
	database := strings.TrimPrefix(point.path, "/")
	switch {
	case point.host == "":
		return point.path, badges
	case databaseTypes[connection.Type] && database != "":
		return point.host + " · " + database, badges
	}
	return point.host, badges
}

// endpoint is where a connection URL points: host is `host` or `host:port`,
// never userinfo; path the database or path with its leading slash.
type endpoint struct {
	scheme, host, path string
}

func (e endpoint) hostname() string {
	if i := strings.LastIndex(e.host, ":"); i != -1 && !strings.Contains(e.host[i:], "]") {
		return e.host[:i]
	}
	return e.host
}

// endpointOf reads a URL, a SQL Server ADO keyword string or a bare file path
// (a sqlite database), dropping any credentials it carries.
func endpointOf(raw string) (endpoint, bool) {
	match := endpointURL.FindStringSubmatch(raw)
	switch {
	case raw == "":
		return endpoint{}, false
	case match == nil && strings.Contains(raw, "="):
		return keywordEndpoint(raw), true
	case match == nil:
		return endpoint{scheme: "file", path: raw}, true
	}
	rest := match[2]
	authority, tail := rest, ""
	if i := strings.IndexAny(rest, "/?#"); i != -1 {
		authority, tail = rest[:i], rest[i:]
	}
	if at := strings.LastIndex(authority, "@"); at != -1 {
		authority = authority[at+1:]
	}
	if i := strings.IndexAny(tail, "?#"); i != -1 {
		tail = tail[:i]
	}
	return endpoint{scheme: strings.ToLower(match[1]), host: authority, path: tail}, true
}

// keywordEndpoint reads `server=host;port=1433;database=db;…`.
func keywordEndpoint(raw string) endpoint {
	fields := map[string]string{}
	for _, pair := range strings.Split(raw, ";") {
		if key, value, ok := strings.Cut(pair, "="); ok {
			fields[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		}
	}
	point := endpoint{scheme: "sqlserver", host: fields["server"]}
	if port := fields["port"]; port != "" {
		point.host += ":" + port
	}
	if database := fields["database"]; database != "" {
		point.path = "/" + database
	}
	return point
}
