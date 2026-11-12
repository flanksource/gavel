package settings

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/verify"
)

// CommentAgent is the comment agent of target's file over r's layers, each
// field traced to the layer label (and page entry) that set it. A layer's page
// keys are relative to its position's directory; a web project's user layer is
// a home directory, so its page keys name nothing in the project and only its
// default applies.
func (r Resolved) CommentAgent(target Target) (verify.ResolvedCommentAgent, error) {
	file, err := commentFile(target)
	if err != nil {
		return verify.ResolvedCommentAgent{}, err
	}
	layers := make([]verify.CommentAgentLayer, 0, len(r.Layers))
	for _, layer := range r.Layers {
		agents := layer.Config.Comments.Agents
		dir := "."
		switch {
		case target.web() && layer.Kind == KindUser:
			agents.Pages = nil
		case target.web() && layer.Kind == KindPath:
			dir = layer.Path
		case !target.web():
			position, err := positionFor(r.Positions, layer.Ref)
			if err != nil {
				return verify.ResolvedCommentAgent{}, err
			}
			dir = filepath.ToSlash(position.Dir)
		}
		layers = append(layers, verify.CommentAgentLayer{Label: layer.Label(), Dir: dir, Agents: agents})
	}
	return verify.ResolveCommentAgent(layers, file)
}

// SaveCommentAgent makes agent the comment agent at layer ref of target's
// chain: the layer's default, or with page its entry for target's file. A nil
// agent removes that entry.
func SaveCommentAgent(ctx context.Context, service Service, target Target, ref Ref, page bool, agent *verify.CommentAgent) error {
	key := ""
	if page {
		positions, _, err := diskChain(target)
		if err != nil {
			return err
		}
		if key, err = pageKey(positions, target, ref); err != nil {
			return err
		}
	}
	return service.Save(ctx, target, ref, func(cfg *verify.GavelConfig) error { return cfg.Comments.Agents.Set(key, agent) })
}

// pageKey is target's file relative to the directory of ref's position, the
// key of its page entry there.
func pageKey(positions []Position, target Target, ref Ref) (string, error) {
	file, err := commentFile(target)
	if err != nil {
		return "", err
	}
	position, err := positionFor(positions, ref)
	if err != nil {
		return "", err
	}
	dir := filepath.ToSlash(position.Dir)
	if target.web() {
		switch position.Kind {
		case KindUser:
			return "", fmt.Errorf("the user's layer holds no page entries for a web project; save its default, or a project or path layer's page entry")
		case KindProject:
			return file, nil
		}
		dir = position.Path
	}
	key, ok := strings.CutPrefix(file, dir+"/")
	if !ok || !verify.IsCommentPageKey(key) {
		return "", fmt.Errorf("%s is not beneath the %s layer at %s", file, position.Kind, position.Dir)
	}
	return key, nil
}

// commentFile is target's file in the namespace its layers' directories use:
// the slash-separated absolute path on disk, or the key in a web project.
func commentFile(target Target) (string, error) {
	if target.web() {
		return target.Key, nil
	}
	absolute, err := filepath.Abs(target.Path)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", target.Path, err)
	}
	return filepath.ToSlash(absolute), nil
}
