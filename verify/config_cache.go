package verify

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/flanksource/commons/merge"
)

// gavelConfigParses counts .gavel.yaml parses. Tests read it to prove a cached
// load skipped parsing; nothing else depends on it.
var gavelConfigParses atomic.Int64

// gavelConfigLayer is one .gavel.yaml LoadGavelConfig consults, read in full. A
// missing file is part of the fingerprint too: its later appearance must change
// the result.
type gavelConfigLayer struct {
	path   string
	data   []byte
	exists bool
}

func readGavelConfigLayers(paths []string) ([]gavelConfigLayer, error) {
	layers := make([]gavelConfigLayer, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			layers = append(layers, gavelConfigLayer{path: path})
			continue
		}
		if err != nil {
			return nil, err
		}
		layers = append(layers, gavelConfigLayer{path: path, data: data, exists: true})
	}
	return layers, nil
}

type gavelConfigCacheEntry struct {
	layers []gavelConfigLayer
	config GavelConfig
}

// gavelConfigCache memoises LoadGavelConfig per absolute target directory. An
// entry is served only when every consulted file — path, existence and exact
// bytes — matches what it was built from, so the chain is still read on every
// call but parsing, merging and spec-layer validation are skipped. Content
// rather than mtime is compared because mtime granularity (a kernel tick on
// Linux) lets a same-size rewrite go unnoticed. Failed loads are never stored.
//
// One entry per distinct directory ever loaded; long-lived processes load a
// bounded set (the dashboard's projects and their repos), so there is no
// eviction. The stored config is never handed out: callers get a deep copy.
var gavelConfigCache = struct {
	sync.Mutex
	entries map[string]gavelConfigCacheEntry
}{entries: map[string]gavelConfigCacheEntry{}}

func cachedGavelConfig(dir string, layers []gavelConfigLayer) (GavelConfig, bool) {
	gavelConfigCache.Lock()
	defer gavelConfigCache.Unlock()
	entry, ok := gavelConfigCache.entries[dir]
	if !ok || !slices.EqualFunc(entry.layers, layers, func(a, b gavelConfigLayer) bool {
		return a.path == b.path && a.exists == b.exists && bytes.Equal(a.data, b.data)
	}) {
		return GavelConfig{}, false
	}
	return entry.config, true
}

func storeGavelConfig(dir string, layers []gavelConfigLayer, cfg GavelConfig) {
	gavelConfigCache.Lock()
	defer gavelConfigCache.Unlock()
	gavelConfigCache.entries[dir] = gavelConfigCacheEntry{layers: layers, config: cfg}
}

// cloneGavelConfig deep-copies cfg so no caller can reach the cached value.
// merge.Clone copies interface values by reference; the ones a config carries
// (ai.cliArgs, todos.lifecycle.steps) are decoded JSON, so they are copied here
// as JSON trees. Types the merge policy shares (the model provider catalog) stay
// shared, as merge.Clone leaves them.
func cloneGavelConfig(cfg GavelConfig) (GavelConfig, error) {
	policy := configPolicy()
	out := merge.Clone(cfg, policy)
	shared := make(map[reflect.Type]bool, len(policy.Shared))
	for _, value := range policy.Shared {
		shared[reflect.TypeOf(value)] = true
	}
	if err := detachInterfaceValues(reflect.ValueOf(&out).Elem(), shared); err != nil {
		return GavelConfig{}, err
	}
	return out, nil
}

func detachInterfaceValues(v reflect.Value, shared map[reflect.Type]bool) error {
	t := v.Type()
	switch kind := t.Kind(); {
	case shared[t],
		(kind == reflect.Pointer || kind == reflect.Interface) && v.IsNil(),
		(kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map) && !mayHoldInterface(t.Elem().Kind()):
		return nil
	}
	switch t.Kind() {
	case reflect.Pointer:
		return detachInterfaceValues(v.Elem(), shared)
	case reflect.Struct:
		for i := range v.NumField() {
			if !t.Field(i).IsExported() {
				continue
			}
			if err := detachInterfaceValues(v.Field(i), shared); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			if err := detachInterfaceValues(v.Index(i), shared); err != nil {
				return err
			}
		}
	case reflect.Map:
		for iter := v.MapRange(); iter.Next(); {
			value := reflect.New(t.Elem()).Elem()
			value.Set(iter.Value())
			if err := detachInterfaceValues(value, shared); err != nil {
				return err
			}
			v.SetMapIndex(iter.Key(), value)
		}
	case reflect.Interface:
		copied, err := cloneJSONValue(v.Elem().Interface())
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(copied))
	}
	return nil
}

func mayHoldInterface(kind reflect.Kind) bool {
	switch kind {
	case reflect.Pointer, reflect.Struct, reflect.Slice, reflect.Array, reflect.Map, reflect.Interface:
		return true
	}
	return false
}

func cloneJSONValue(value any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, element := range typed {
			copied, err := cloneJSONValue(element)
			if err != nil {
				return nil, err
			}
			out[key] = copied
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, element := range typed {
			copied, err := cloneJSONValue(element)
			if err != nil {
				return nil, err
			}
			out[i] = copied
		}
		return out, nil
	}
	if value == nil || !mayHoldReference(reflect.TypeOf(value).Kind()) {
		return value, nil
	}
	return nil, fmt.Errorf("clone gavel config: cannot deep-copy %T held in an interface field (expected decoded JSON)", value)
}

func mayHoldReference(kind reflect.Kind) bool {
	return mayHoldInterface(kind) || kind == reflect.Chan || kind == reflect.Func || kind == reflect.UnsafePointer
}
