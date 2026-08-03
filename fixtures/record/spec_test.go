package record

import (
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unmarshalSpec decodes a `record:` value the way fixture frontmatter does,
// through goccy, so the BytesUnmarshaler dispatch is part of what is tested.
func unmarshalSpec(t *testing.T, body string) *Spec {
	t.Helper()
	var doc struct {
		Record *Spec `yaml:"record"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(body), &doc))
	return doc.Record
}

func TestSpecUnmarshalShorthands(t *testing.T) {
	t.Run("bare kind", func(t *testing.T) {
		spec := unmarshalSpec(t, "record: ansi\n")
		assert.Equal(t, []Kind{KindANSI}, spec.Kinds())
	})

	t.Run("list of kinds", func(t *testing.T) {
		spec := unmarshalSpec(t, "record: [ansi, http]\n")
		assert.Equal(t, []Kind{KindANSI, KindHTTP}, spec.Kinds())
		assert.Equal(t, HTTPConnect, spec.HTTP.Mode, "the shorthand should still get the documented default mode")
		assert.Equal(t, ScopeFile, spec.HTTP.Scope)
	})

	t.Run("none opts a test out of a file-level default", func(t *testing.T) {
		spec := unmarshalSpec(t, "record: none\n")
		assert.True(t, spec.IsEmpty(), "an explicit opt-out must start nothing")
	})

	t.Run("unknown kind fails loud", func(t *testing.T) {
		var doc struct {
			Record *Spec `yaml:"record"`
		}
		err := yaml.Unmarshal([]byte("record: tcpdump\n"), &doc)
		require.Error(t, err, "a typo must not silently record nothing")
		assert.Contains(t, err.Error(), "tcpdump")
	})
}

func TestSpecUnmarshalFullForm(t *testing.T) {
	spec := unmarshalSpec(t, `
record:
  ansi:
    width: 120
    height: 40
    interval: 100ms
    maxBytes: 4MiB
  http:
    mode: mitm
    hosts: ["*.github.com"]
    bodies: 64KiB
    requireEntries: 1
  sql:
    dsn: postgres://localhost/gavel
    params: true
  clients: {}
`)

	require.NotNil(t, spec.ANSI)
	assert.Equal(t, 120, spec.ANSI.Width)
	assert.Equal(t, 40, spec.ANSI.Height)
	assert.Equal(t, 100*time.Millisecond, spec.ANSI.Interval)
	assert.Equal(t, Size(4*1024*1024), spec.ANSI.MaxBytes, "4MiB is binary, not 4e6")

	require.NotNil(t, spec.HTTP)
	assert.Equal(t, HTTPMITM, spec.HTTP.Mode, "an explicit mode must survive ApplyDefaults")
	assert.Equal(t, []string{"*.github.com"}, spec.HTTP.Hosts)
	assert.Equal(t, Size(64*1024), spec.HTTP.Bodies)
	assert.Equal(t, 1, spec.HTTP.RequireEntries)

	require.NotNil(t, spec.SQL)
	assert.Equal(t, SQLProxy, spec.SQL.Mode, "sql defaults to the proxy, the only mode that sees a child process")
	assert.True(t, spec.SQL.Params)

	assert.Equal(t, []Kind{KindANSI, KindHTTP, KindSQL, KindClients}, spec.Kinds())
}

func TestSpecEnabledIsNilSafe(t *testing.T) {
	var spec *Spec
	assert.False(t, spec.Enabled(KindHTTP))
	assert.Empty(t, spec.Kinds())
	assert.True(t, spec.IsEmpty())

	off := &Spec{HTTP: &HTTPOptions{Mode: HTTPOff}, SQL: &SQLOptions{Mode: SQLOff}}
	assert.False(t, off.Enabled(KindHTTP), "mode: off must not start a proxy")
	assert.False(t, off.Enabled(KindSQL))
}

func TestParseShorthand(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  []Kind
	}{
		{"empty", "", nil},
		{"comma separated", "ansi,http", []Kind{KindANSI, KindHTTP}},
		{"space separated", "ansi sql", []Kind{KindANSI, KindSQL}},
		{"uppercase", "ANSI", []Kind{KindANSI}},
		{"all", "all", []Kind{KindANSI, KindHTTP, KindSQL}},
		{"off", "off", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := Parse(tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.want, spec.Kinds())
		})
	}

	_, err := Parse("ansi,bogus")
	require.Error(t, err)
}

func TestMissingReportsUnimplementedRecorders(t *testing.T) {
	spec, err := Parse("ansi,http,sql,clients")
	require.NoError(t, err)

	// Every recorder registers itself from its own init, so nothing is missing
	// today. Unregistering one is what keeps the guard honest: the next kind
	// added to the enum must be named rather than silently skipped.
	assert.Empty(t, spec.Missing(), "every declared recorder has an implementation")

	delete(Implemented, KindSQL)
	t.Cleanup(func() { Implemented[KindSQL] = true })
	assert.Equal(t, []Kind{KindSQL}, spec.Missing(),
		"a recorder with no implementation must be named, not silently skipped")

	var undeclared *Spec
	assert.Empty(t, undeclared.Missing(), "a fixture that records nothing is never in error")
}

func TestParseDistinguishesUndeclaredFromOptOut(t *testing.T) {
	undeclared, err := Parse("")
	require.NoError(t, err)
	assert.Nil(t, undeclared, "an absent value must stay nil so a run-wide --record can fill it in")

	optOut, err := Parse("none")
	require.NoError(t, err)
	require.NotNil(t, optOut, "an explicit `none` must outrank a run-wide --record")
	assert.True(t, optOut.IsEmpty())
}

func TestSizeUnmarshal(t *testing.T) {
	cases := []struct {
		yaml string
		want Size
	}{
		{"1024", 1024},
		{`"4MiB"`, 4 * 1024 * 1024},
		{`"64KiB"`, 64 * 1024},
		{`"1MB"`, 1000 * 1000},
	}
	for _, tc := range cases {
		t.Run(tc.yaml, func(t *testing.T) {
			var size Size
			require.NoError(t, yaml.Unmarshal([]byte(tc.yaml), &size))
			assert.Equal(t, tc.want, size)
		})
	}

	var size Size
	require.Error(t, yaml.Unmarshal([]byte(`"huge"`), &size))
}
