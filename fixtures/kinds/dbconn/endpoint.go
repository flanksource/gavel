package dbconn

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/flanksource/commons-db/models"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
)

// Endpoint is the database/sql driver name and URL of a SQL Server or
// PostgreSQL connection, SQL Server keyword DSNs rewritten as URLs with their
// extra parameters preserved.
func Endpoint(connection *models.Connection) (string, *url.URL, error) {
	if connection.Type == models.ConnectionTypeSQLServer {
		cfg, err := msdsn.Parse(connection.URL)
		if err != nil {
			return "", nil, errors.New("invalid SQL Server connection DSN")
		}
		endpoint := cfg.URL()
		params := endpoint.Query()
		for key, value := range cfg.Parameters {
			switch key {
			case "server", "port", "user id", "password", "database":
				continue
			}
			params.Set(key, value)
		}
		endpoint.RawQuery = params.Encode()
		return "sqlserver", endpoint, nil
	}
	if connection.Type != models.ConnectionTypePostgres {
		return "", nil, fmt.Errorf("existing connection type %q is unsupported", connection.Type)
	}
	endpoint, err := url.Parse(connection.URL)
	if err != nil || endpoint.Hostname() == "" || !strings.HasPrefix(endpoint.Scheme, "postgres") {
		return "", nil, errors.New("invalid PostgreSQL connection URL")
	}
	if endpoint.Port() == "" {
		endpoint.Host = net.JoinHostPort(endpoint.Hostname(), "5432")
	}
	return "postgres", endpoint, nil
}
