package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type View struct {
	Path       string                  `json:"path"`
	Revision   string                  `json:"revision"`
	Source     verify.LifecycleConfig  `json:"source"`
	Inherited  lifecycle.Lifecycle     `json:"inherited"`
	Effective  lifecycle.Lifecycle     `json:"effective"`
	FileSource *verify.LifecycleConfig `json:"fileSource,omitempty"`
}

type Edit struct {
	Revision   string                  `json:"revision"`
	Source     verify.LifecycleConfig  `json:"source"`
	FileSource *verify.LifecycleConfig `json:"fileSource,omitempty"`
}

func Inspect(kind, dir string) (View, error) { return Preview(kind, dir, nil) }

var saveMu sync.Mutex

func Preview(kind, dir string, edit *Edit) (View, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return View{}, err
	}
	view := View{Path: filepath.Join(dir, verify.GavelConfigFileName)}
	trace, err := verify.LoadGavelConfigTrace(dir)
	if err != nil {
		return view, err
	}
	base, err := Default(kind)
	if err != nil {
		return view, err
	}
	hash := sha256.New()
	for _, layer := range trace.Sources {
		_, _ = hash.Write([]byte(layer.Path + "\x00" + layer.Raw))
		config, err := Config(layer.Config, kind)
		if err != nil {
			return view, err
		}
		overlay, err := lifecycle.Read(config, filepath.Dir(layer.Path))
		if err != nil {
			return view, err
		}
		if config.File != "" {
			data, err := utils.ReadWithin(filepath.Dir(layer.Path), config.File)
			if err != nil {
				return view, err
			}
			_, _ = hash.Write(data)
		}
		if layer.Path == view.Path {
			view.Source = config
			continue
		}
		base = lifecycle.Merge(base, overlay)
	}
	view.Revision = hex.EncodeToString(hash.Sum(nil))
	if edit != nil && edit.Revision != view.Revision {
		return view, fmt.Errorf("lifecycle configuration changed; reload before saving")
	}
	view.Inherited = base
	if edit != nil {
		view.Source = edit.Source
	}
	overlay, err := lifecycle.Read(view.Source, dir)
	if err != nil {
		return view, err
	}
	if view.Source.File != "" {
		data, err := json.Marshal(overlay)
		if err != nil {
			return view, err
		}
		var source verify.LifecycleConfig
		if err := json.Unmarshal(data, &source); err != nil {
			return view, err
		}
		view.FileSource = &source
		if edit != nil && edit.FileSource != nil {
			if edit.FileSource.File != "" {
				return view, fmt.Errorf("nested lifecycle file references are unsupported")
			}
			view.FileSource = edit.FileSource
			overlay, err = lifecycle.Read(*edit.FileSource, dir)
			if err != nil {
				return view, err
			}
		}
	}
	def := lifecycle.Merge(base, overlay)
	if err := Validate(kind, def); err != nil {
		return view, err
	}
	view.Effective, err = lifecycle.Select(def)
	if err != nil {
		return view, err
	}
	_, err = lifecycle.New(view.Effective)
	return view, err
}

func Save(kind, dir string, edit Edit) (View, error) {
	saveMu.Lock()
	defer saveMu.Unlock()
	view, err := Preview(kind, dir, &edit)
	if err != nil {
		return view, err
	}
	data, err := utils.ReadWithin(filepath.Dir(view.Path), verify.GavelConfigFileName)
	if err != nil && !os.IsNotExist(err) {
		return view, err
	}
	var document yaml.Node
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &document); err != nil {
			return view, err
		}
	} else {
		document = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return view, fmt.Errorf("configuration must be a YAML mapping")
	}
	root := document.Content[0]
	section := mappingValue(root, kind)
	if section == nil {
		section = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: kind}, section)
	}
	if section.Kind != yaml.MappingNode {
		return view, fmt.Errorf("%s configuration must be a mapping", kind)
	}
	var node yaml.Node
	if err := node.Encode(edit.Source); err != nil {
		return view, err
	}
	for i := 0; i < len(section.Content); i += 2 {
		if section.Content[i].Value == "lifecycle" {
			section.Content[i+1] = &node
			return writeDocument(kind, dir, view, document)
		}
	}
	section.Content = append(section.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "lifecycle"}, &node)
	return writeDocument(kind, dir, view, document)
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func writeDocument(kind, dir string, view View, document yaml.Node) (View, error) {
	data, err := yaml.Marshal(&document)
	if err != nil {
		return View{}, err
	}
	root, err := os.OpenRoot(filepath.Dir(view.Path))
	if err != nil {
		return view, err
	}
	defer root.Close()
	if view.FileSource != nil {
		path, err := utils.ResolveWithin(root.Name(), view.Source.File)
		if err != nil {
			return view, err
		}
		path, err = filepath.Rel(root.Name(), path)
		if err != nil {
			return view, err
		}
		fileData, err := yaml.Marshal(view.FileSource)
		if err != nil {
			return view, err
		}
		if err := writeAtomic(root, path, fileData); err != nil {
			return view, err
		}
	}
	if err := writeAtomic(root, verify.GavelConfigFileName, data); err != nil {
		return View{}, err
	}
	return Inspect(kind, dir)
}

func writeAtomic(root *os.Root, path string, data []byte) (returnErr error) {
	mode := os.FileMode(0600)
	if info, err := root.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	temp := filepath.Join(filepath.Dir(path), ".lifecycle-"+uuid.NewString())
	file, err := root.OpenFile(temp, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err := root.Remove(temp); err != nil && !os.IsNotExist(err) {
			returnErr = errors.Join(returnErr, fmt.Errorf("remove temporary lifecycle file: %w", err))
		}
	}()
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return root.Rename(temp, path)
}
