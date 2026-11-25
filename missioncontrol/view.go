// Reads the panels of a Mission Control View, such as the graphs of substrate's Prometheus
// metrics: timeseries rows grouped into series, and number and gauge panels as one value.
package missioncontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/incident-commander/sdk/client"
)

// View is what a Mission Control View's panels show.
type View struct {
	MissionControl  string    `json:"missionControl"`
	Title           string    `json:"title,omitempty"`
	RefreshError    string    `json:"refreshError,omitempty"`
	LastRefreshedAt time.Time `json:"lastRefreshedAt"`
	Panels          []Panel   `json:"panels"`
}

// Panel is one panel of a View: Series for a timeseries panel, Value for a number or gauge.
type Panel struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Unit        string   `json:"unit,omitempty"`
	Series      []Series `json:"series,omitempty"`
	// Value is a number or gauge panel's value; nil when its query returned no rows.
	Value *float64 `json:"value,omitempty"`
	// Unsupported says why gavel shows the panel without its data.
	Unsupported string `json:"unsupported,omitempty"`
}

// Series is one line of a timeseries panel: the points sharing one set of labels.
type Series struct {
	Labels map[string]string `json:"labels"`
	Points []Point           `json:"points"`
}

// Point is one sample of a Series.
type Point struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// viewResponse is the part of Mission Control's rendered View gavel reads.
type viewResponse struct {
	Title           string          `json:"title"`
	RefreshError    string          `json:"refreshError"`
	LastRefreshedAt time.Time       `json:"lastRefreshedAt"`
	Panels          []panelResponse `json:"panels"`
}

type panelResponse struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        string           `json:"type"`
	Number      *unitConfig      `json:"number"`
	Gauge       *unitConfig      `json:"gauge"`
	Timeseries  *timeseriesKeys  `json:"timeseries"`
	Rows        []map[string]any `json:"rows"`
}

type unitConfig struct {
	Unit string `json:"unit"`
}

type timeseriesKeys struct {
	TimeKey  string `json:"timeKey"`
	ValueKey string `json:"valueKey"`
}

// View reads the panels of the View named namespace/name, its queries templated with variables,
// which they read as var.<key>.
func (s *Source) View(ctx context.Context, view string, variables map[string]string) (View, error) {
	namespace, name, ok := strings.Cut(view, "/")
	if !ok || namespace == "" || name == "" || strings.Contains(name, "/") {
		return View{}, fmt.Errorf("view %q must be namespace/name", view)
	}
	what := "view " + view
	if variables == nil {
		variables = map[string]string{}
	}
	r, err := s.Client.R(ctx).Post("/view/"+namespace+"/"+name, map[string]any{"variables": variables})
	if err != nil {
		return View{}, s.explain(what, err)
	}
	body, err := r.AsString()
	if err != nil {
		return View{}, s.explain(what, err)
	}
	if r.StatusCode == http.StatusNotFound {
		return View{}, fmt.Errorf("view %s not found in Mission Control at %s: %w", view, s.URL, client.NewServerError(r.StatusCode, []byte(body)))
	}
	if !r.IsOK() {
		return View{}, s.explain(what, client.NewServerError(r.StatusCode, []byte(body)))
	}
	var rendered viewResponse
	if err := json.Unmarshal([]byte(body), &rendered); err != nil {
		return View{}, s.explain(what, fmt.Errorf("decode the view: %w", err))
	}
	panels := make([]Panel, 0, len(rendered.Panels))
	for _, panel := range rendered.Panels {
		panels = append(panels, readPanel(panel))
	}
	return View{
		MissionControl: s.URL, Title: rendered.Title, RefreshError: rendered.RefreshError,
		LastRefreshedAt: rendered.LastRefreshedAt, Panels: panels,
	}, nil
}

func readPanel(rendered panelResponse) Panel {
	panel := Panel{Name: rendered.Name, Description: rendered.Description, Type: rendered.Type}
	var err error
	switch rendered.Type {
	case "timeseries":
		panel.Series, err = series(rendered.Rows, rendered.Timeseries)
	case "number":
		if rendered.Number != nil {
			panel.Unit = strings.TrimSpace(rendered.Number.Unit)
		}
		panel.Value, err = firstValue(rendered.Rows)
	case "gauge":
		if rendered.Gauge != nil {
			panel.Unit = strings.TrimSpace(rendered.Gauge.Unit)
		}
		panel.Value, err = firstValue(rendered.Rows)
	default:
		panel.Unsupported = fmt.Sprintf("gavel doesn't draw %s panels", rendered.Type)
	}
	if err != nil {
		panel.Unsupported = err.Error()
	}
	return panel
}

// series groups timeseries rows by their labels, every column but the time and the value: series
// ordered by their labels, points by time.
func series(rows []map[string]any, keys *timeseriesKeys) ([]Series, error) {
	timeKey, valueKey := "timestamp", "value"
	if keys != nil && keys.TimeKey != "" {
		timeKey = keys.TimeKey
	}
	if keys != nil && keys.ValueKey != "" {
		valueKey = keys.ValueKey
	}
	byLabels := map[string]*Series{}
	for i, row := range rows {
		raw := fmt.Sprint(row[timeKey])
		at, err := parseTime(raw)
		if err != nil {
			return nil, fmt.Errorf("row %d's %s %q isn't an RFC 3339 time or Go's text for one", i+1, timeKey, raw)
		}
		value, err := number(row[valueKey])
		if err != nil {
			return nil, fmt.Errorf("row %d's %s", i+1, err)
		}
		labels := map[string]string{}
		for column, cell := range row {
			if column != timeKey && column != valueKey && cell != nil {
				labels[column] = fmt.Sprint(cell)
			}
		}
		key := labelKey(labels)
		if byLabels[key] == nil {
			byLabels[key] = &Series{Labels: labels}
		}
		byLabels[key].Points = append(byLabels[key].Points, Point{Time: at, Value: value})
	}
	keysInOrder := make([]string, 0, len(byLabels))
	for key := range byLabels {
		keysInOrder = append(keysInOrder, key)
	}
	sort.Strings(keysInOrder)
	all := make([]Series, 0, len(keysInOrder))
	for _, key := range keysInOrder {
		line := byLabels[key]
		sort.SliceStable(line.Points, func(a, b int) bool { return line.Points[a].Time.Before(line.Points[b].Time) })
		all = append(all, *line)
	}
	return all, nil
}

// goTimeText is how Go writes a time as text, which is how Mission Control's SQLite keeps the time
// of a Prometheus sample that a panel's query reads.
const goTimeText = "2006-01-02 15:04:05.999999999 -0700 MST"

// parseTime reads a timeseries row's time, which Mission Control renders as RFC 3339 or as Go's text.
func parseTime(raw string) (time.Time, error) {
	if at, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return at, nil
	}
	return time.Parse(goTimeText, raw)
}

// labelKey is labels as one comparable string, in label-name order.
func labelKey(labels map[string]string) string {
	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	sort.Strings(names)
	var key strings.Builder
	for _, name := range names {
		fmt.Fprintf(&key, "%s=%q,", name, labels[name])
	}
	return key.String()
}

// firstValue is the value column of a number or gauge panel's first row, nil without rows.
func firstValue(rows []map[string]any) (*float64, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	value, err := number(rows[0]["value"])
	if err != nil {
		return nil, fmt.Errorf("row 1's %s", err)
	}
	return &value, nil
}

// number reads a cell Mission Control renders as a JSON number or a numeric string.
func number(cell any) (float64, error) {
	switch v := cell.(type) {
	case float64:
		return v, nil
	case string:
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n, nil
		}
	}
	return 0, fmt.Errorf("value %q isn't a number", fmt.Sprint(cell))
}
