package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/verify"
	yamlv3 "gopkg.in/yaml.v3"
)

var stdoutIsTerminal = func() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func (r ConfigResult) Pretty() api.Text {
	body := renderMergedConfigYAML(r.configTrace(), stdoutIsTerminal())
	return clicky.Text(body, "font-mono")
}

func (r ConfigResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(prunedConfigValue(r.Merged))
}

func (r ConfigResult) MarshalYAML() ([]byte, error) {
	return []byte(renderMergedConfigYAML(r.configTrace(), false)), nil
}

func (r ConfigResult) configTrace() verify.GavelConfigTrace {
	return verify.GavelConfigTrace{
		TargetPath: r.TargetPath,
		TargetDir:  r.TargetDir,
		GitRoot:    r.GitRoot,
		Sources:    r.Sources,
		Merged:     r.Merged,
	}
}

type configSourceRef struct {
	Key      string
	Comment  string
	Suppress bool
}

type configProvenanceNode struct {
	Source *configSourceRef
	Fields map[string]*configProvenanceNode
	Items  []*configProvenanceNode
}

func renderMergedConfigYAML(trace verify.GavelConfigTrace, annotate bool) string {
	var provenance *configProvenanceNode
	if annotate {
		provenance = buildConfigProvenance(trace)
	}

	root, ok := buildYAMLNode(reflect.ValueOf(trace.Merged), provenance, nil)
	if !ok {
		return "{}\n"
	}

	doc := &yamlv3.Node{
		Kind:    yamlv3.DocumentNode,
		Content: []*yamlv3.Node{root},
	}

	var buf strings.Builder
	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "{}\n"
	}
	_ = enc.Close()

	out := buf.String()
	if out == "" {
		return "{}\n"
	}
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out
}

func buildConfigProvenance(trace verify.GavelConfigTrace) *configProvenanceNode {
	builtIn := &configSourceRef{
		Key:     "built-in-defaults",
		Comment: "from built-in defaults",
	}

	current := verify.GavelConfig{Verify: verify.DefaultVerifyConfig()}
	currentTree := normalizedConfigValue(current)
	provenance := buildProvenanceTree(currentTree, builtIn)

	for _, source := range trace.Sources {
		ref := configSourceReference(trace, source)
		next := verify.MergeGavelConfig(current, source.Config)
		nextTree := normalizedConfigValue(next)
		provenance = mergeProvenanceTree(provenance, currentTree, nextTree, ref)
		current = next
		currentTree = nextTree
	}

	return provenance
}

func configSourceReference(trace verify.GavelConfigTrace, source verify.GavelConfigSource) *configSourceRef {
	ref := &configSourceRef{
		Key: source.Origin + "|" + source.Path,
	}

	displayPath := source.Path
	if home, err := os.UserHomeDir(); err == nil {
		if displayPath == home+"/.gavel.yaml" {
			displayPath = "~/.gavel.yaml"
		}
	}

	switch source.Origin {
	case "git-root":
		ref.Suppress = true
	case "user-home":
		ref.Comment = "from ~/.gavel.yaml"
	case "target-directory":
		ref.Comment = "from " + displayPath
	case "parent-directory":
		ref.Comment = "from " + displayPath
	default:
		if source.Path != "" {
			ref.Comment = "from " + displayPath
		} else {
			ref.Comment = "from " + source.Origin
		}
	}

	if trace.GitRoot != "" && source.Path == trace.GitRoot+"/.gavel.yaml" {
		ref.Suppress = true
	}

	return ref
}

func normalizedConfigValue(v any) any {
	data, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}

	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]any{}
	}

	pruned := pruneEmpty(parsed)
	if pruned == nil {
		return map[string]any{}
	}
	return pruned
}

func prunedConfigValue(v any) any {
	return normalizedConfigValue(v)
}

func buildProvenanceTree(v any, source *configSourceRef) *configProvenanceNode {
	switch typed := v.(type) {
	case map[string]any:
		node := &configProvenanceNode{
			Fields: make(map[string]*configProvenanceNode, len(typed)),
		}
		for key, child := range typed {
			node.Fields[key] = buildProvenanceTree(child, source)
		}
		return node
	case []any:
		node := &configProvenanceNode{
			Items: make([]*configProvenanceNode, 0, len(typed)),
		}
		for _, child := range typed {
			node.Items = append(node.Items, buildProvenanceTree(child, source))
		}
		return node
	default:
		return &configProvenanceNode{Source: source}
	}
}

func mergeProvenanceTree(prev *configProvenanceNode, before, after any, source *configSourceRef) *configProvenanceNode {
	switch next := after.(type) {
	case map[string]any:
		beforeMap, _ := before.(map[string]any)
		node := &configProvenanceNode{
			Fields: make(map[string]*configProvenanceNode, len(next)),
		}
		for key, child := range next {
			var prevChild *configProvenanceNode
			var beforeChild any
			if prev != nil {
				prevChild = prev.Fields[key]
			}
			if beforeMap != nil {
				beforeChild = beforeMap[key]
			}
			if beforeMap == nil {
				node.Fields[key] = buildProvenanceTree(child, source)
				continue
			}
			if _, ok := beforeMap[key]; !ok {
				node.Fields[key] = buildProvenanceTree(child, source)
				continue
			}
			node.Fields[key] = mergeProvenanceTree(prevChild, beforeChild, child, source)
		}
		return node
	case []any:
		beforeSlice, _ := before.([]any)
		node := &configProvenanceNode{
			Items: make([]*configProvenanceNode, 0, len(next)),
		}
		nextBeforeIndex := 0
		for _, child := range next {
			match := -1
			for i := nextBeforeIndex; i < len(beforeSlice); i++ {
				if reflect.DeepEqual(beforeSlice[i], child) {
					match = i
					break
				}
			}
			if match >= 0 && prev != nil && match < len(prev.Items) {
				node.Items = append(node.Items, prev.Items[match])
				nextBeforeIndex = match + 1
				continue
			}
			node.Items = append(node.Items, buildProvenanceTree(child, source))
		}
		return node
	default:
		if prev != nil && reflect.DeepEqual(before, after) {
			return prev
		}
		return &configProvenanceNode{Source: source}
	}
}

func (n *configProvenanceNode) uniformSource() *configSourceRef {
	if n == nil {
		return nil
	}
	if len(n.Fields) == 0 && len(n.Items) == 0 {
		return n.Source
	}

	var source *configSourceRef
	for _, child := range n.Fields {
		childSource := child.uniformSource()
		if childSource == nil {
			return nil
		}
		if source == nil {
			source = childSource
			continue
		}
		if source.Key != childSource.Key {
			return nil
		}
	}
	for _, child := range n.Items {
		childSource := child.uniformSource()
		if childSource == nil {
			return nil
		}
		if source == nil {
			source = childSource
			continue
		}
		if source.Key != childSource.Key {
			return nil
		}
	}
	if source != nil {
		return source
	}
	return n.Source
}

func buildYAMLNode(v reflect.Value, meta *configProvenanceNode, parent *configSourceRef) (*yamlv3.Node, bool) {
	if !v.IsValid() {
		return nil, false
	}

	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		node := &yamlv3.Node{Kind: yamlv3.MappingNode}
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}

			key, omit := yamlFieldName(field)
			if key == "" || key == "-" {
				continue
			}

			childMeta := (*configProvenanceNode)(nil)
			if meta != nil && meta.Fields != nil {
				childMeta = meta.Fields[key]
			}
			childSource := childMeta.uniformSource()
			nextParent := parent
			if shouldAnnotate(childSource, parent) {
				nextParent = childSource
			}

			childNode, ok := buildYAMLNode(v.Field(i), childMeta, nextParent)
			if !ok {
				if omit {
					continue
				}
				continue
			}

			keyNode := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: key}
			if shouldAnnotate(childSource, parent) {
				keyNode.HeadComment = childSource.Comment
			}
			node.Content = append(node.Content, keyNode, childNode)
		}
		if len(node.Content) == 0 {
			return nil, false
		}
		return node, true
	case reflect.Map:
		if v.Len() == 0 {
			return nil, false
		}
		node := &yamlv3.Node{Kind: yamlv3.MappingNode}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
		})
		for _, key := range keys {
			keyString := fmt.Sprint(key.Interface())
			childMeta := (*configProvenanceNode)(nil)
			if meta != nil && meta.Fields != nil {
				childMeta = meta.Fields[keyString]
			}
			childSource := childMeta.uniformSource()
			nextParent := parent
			if shouldAnnotate(childSource, parent) {
				nextParent = childSource
			}

			childNode, ok := buildYAMLNode(v.MapIndex(key), childMeta, nextParent)
			if !ok {
				continue
			}

			keyNode := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: keyString}
			if shouldAnnotate(childSource, parent) {
				keyNode.HeadComment = childSource.Comment
			}
			node.Content = append(node.Content, keyNode, childNode)
		}
		if len(node.Content) == 0 {
			return nil, false
		}
		return node, true
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, false
		}
		node := &yamlv3.Node{Kind: yamlv3.SequenceNode}
		for i := 0; i < v.Len(); i++ {
			childMeta := (*configProvenanceNode)(nil)
			if meta != nil && i < len(meta.Items) {
				childMeta = meta.Items[i]
			}
			childSource := childMeta.uniformSource()
			nextParent := parent
			if shouldAnnotate(childSource, parent) {
				nextParent = childSource
			}

			childNode, ok := buildYAMLNode(v.Index(i), childMeta, nextParent)
			if !ok {
				continue
			}
			if shouldAnnotate(childSource, parent) {
				childNode.HeadComment = childSource.Comment
			}
			node.Content = append(node.Content, childNode)
		}
		if len(node.Content) == 0 {
			return nil, false
		}
		return node, true
	default:
		if isEmptyConfigValue(v) {
			return nil, false
		}
		return scalarYAMLNode(v), true
	}
}

func yamlFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return field.Name, false
	}

	parts := strings.Split(tag, ",")
	name := parts[0]
	omitEmpty := false
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitEmpty = true
		}
	}
	if name == "" {
		name = field.Name
	}
	return name, omitEmpty
}

func shouldAnnotate(source, parent *configSourceRef) bool {
	if source == nil || source.Suppress || source.Comment == "" {
		return false
	}
	if parent == nil {
		return true
	}
	return parent.Key != source.Key
}

func scalarYAMLNode(v reflect.Value) *yamlv3.Node {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: v.String()}
	case reflect.Bool:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(v.Bool())}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(v.Int(), 10)}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!int", Value: strconv.FormatUint(v.Uint(), 10)}
	case reflect.Float32, reflect.Float64:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(v.Float(), 'f', -1, 64)}
	default:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: fmt.Sprint(v.Interface())}
	}
}

func isEmptyConfigValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isEmptyConfigValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
