package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// deliverTargetFlags choose a delivery target: a live session, or a new run of a prompt.
type deliverTargetFlags struct {
	session, prompt, group, runOptions, edits string
	fresh                                     bool
}

func (f *deliverTargetFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.session, "session", "", "Captain session ID or Amp Orb thread ID or URL to deliver into")
	cmd.Flags().StringVar(&f.prompt, "prompt", "", "Prompt a new managed session runs")
	cmd.Flags().StringVar(&f.group, "group", "", "Reuse the live session recorded for this group key (with --prompt)")
	cmd.Flags().StringVar(&f.runOptions, "run-options", "", "Run options as JSON: presets, spec, variables, instructions (with --prompt)")
	cmd.Flags().BoolVar(&f.fresh, "fresh", false, "Always start a new managed session, which becomes the --group's session (with --prompt)")
	cmd.Flags().StringVar(&f.edits, "edits", "", "direct (default) lets the agent edit files; propose has it answer with a proposed patch")
	cmd.MarkFlagsMutuallyExclusive("session", "prompt")
	cmd.MarkFlagsOneRequired("session", "prompt")
}

func (f deliverTargetFlags) target() (comments.DeliverTarget, error) {
	if f.session != "" {
		if f.group != "" || f.runOptions != "" || f.fresh {
			return comments.DeliverTarget{}, errors.New("--group, --run-options and --fresh apply to --prompt only")
		}
		target := comments.DeliverTarget{Kind: "session", SessionID: f.session, Edits: f.edits}
		return target, target.Validate()
	}
	target := comments.DeliverTarget{Kind: "run", Prompt: f.prompt, GroupKey: f.group, Fresh: f.fresh, Edits: f.edits}
	if f.runOptions != "" {
		target.RunOptions = json.RawMessage(f.runOptions)
	}
	return target, target.Validate()
}

func printJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func commentsDeliverCommand() *cobra.Command {
	var flags deliverTargetFlags
	cmd := &cobra.Command{Use: "deliver <id>", Short: "Send a comment to an agent session or Amp thread, or start a managed session", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		target, err := flags.target()
		if err != nil {
			return err
		}
		service, err := ui.OpenComments(cmd.Context())
		if err != nil {
			return err
		}
		thread, err := service.Get(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		delivered, err := service.Deliver(cmd.Context(), comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: uuid.NewString(), Target: target})
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), delivered)
	}}
	flags.register(cmd)
	return cmd
}

// agentSettings is a target's settings layers and the comment agent that
// applies to it, each field traced to its layer: the shape of the editor's
// GET /api/file/{id}/settings.
type agentSettings struct {
	settings.Resolved
	Agent verify.ResolvedCommentAgent `json:"agent"`
}

func commentsAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "agents", Short: "Configure the comment agent new human threads on a file are delivered to (comments.agents in its settings layers)"}
	cmd.AddCommand(commentsAgentGetCommand(), commentsAgentSetCommand(), commentsAgentDeleteCommand())
	return cmd
}

func commentsAgentGetCommand() *cobra.Command {
	var target agentTargetFlags
	cmd := &cobra.Command{Use: "get <path>", Short: "Show a file's settings layers and the comment agent that applies to it", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		resolved, service, err := target.resolve(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printAgentSettings(cmd, service, resolved)
	}}
	target.register(cmd)
	return cmd
}

func commentsAgentSetCommand() *cobra.Command {
	var target agentTargetFlags
	var layer agentLayerFlags
	var fields agentFieldFlags
	cmd := &cobra.Command{Use: "set <path>", Short: "Set the comment agent at a settings layer: its default, or with --page its entry for the file", Long: "Replaces the whole entry with the fields the flags give: a field no flag sets is removed and inherits again. (PUT /api/v1/comments/agent merges a patch instead.)", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		agent, err := fields.agent(cmd)
		if err != nil {
			return err
		}
		return saveCommentAgent(cmd, target, layer, args[0], agent)
	}}
	target.register(cmd)
	layer.register(cmd)
	fields.register(cmd)
	return cmd
}

func commentsAgentDeleteCommand() *cobra.Command {
	var target agentTargetFlags
	var layer agentLayerFlags
	cmd := &cobra.Command{Use: "delete <path>", Short: "Remove the comment agent at a settings layer: its default, or with --page its entry for the file", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return saveCommentAgent(cmd, target, layer, args[0], nil)
	}}
	target.register(cmd)
	layer.register(cmd)
	return cmd
}

func saveCommentAgent(cmd *cobra.Command, target agentTargetFlags, layer agentLayerFlags, arg string, agent *verify.CommentAgent) error {
	ref, err := layer.ref()
	if err != nil {
		return err
	}
	resolved, service, err := target.resolve(cmd.Context(), arg)
	if err != nil {
		return err
	}
	if err := settings.SaveCommentAgent(cmd.Context(), service, resolved, ref, layer.page, verify.ReplaceCommentAgent(agent)); err != nil {
		return err
	}
	return printAgentSettings(cmd, service, resolved)
}

func printAgentSettings(cmd *cobra.Command, service settings.Service, target settings.Target) error {
	resolved, err := service.Resolve(cmd.Context(), target)
	if err != nil {
		return err
	}
	agent, err := resolved.CommentAgent(target)
	if err != nil {
		return err
	}
	return printJSON(cmd.OutOrStdout(), agentSettings{Resolved: resolved, Agent: agent})
}

// agentTargetFlags pick the settings target: a file on disk, or with --scope
// a page key of that web project.
type agentTargetFlags struct{ scope string }

func (f *agentTargetFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.scope, "scope", "", "Web project scope URI, e.g. gavel://workspace/<hash>; <path> is then the page key beneath it")
}

// resolve is the settings target arg names and the settings service of the
// process's project catalog mode.
func (f agentTargetFlags) resolve(ctx context.Context, arg string) (settings.Target, settings.Service, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return settings.Target{}, nil, err
	}
	service, err := settings.Open(ctx, catalog)
	if err != nil {
		return settings.Target{}, nil, err
	}
	if f.scope != "" {
		target, err := agents.WebTarget(ctx, catalog, f.scope, f.scope+"/"+arg)
		return target, service, err
	}
	target, err := fileTarget(ctx, catalog, arg)
	return target, service, err
}

// fileTarget is the settings target of the file at name, in the catalog's
// most specific project holding it (none outside every project).
func fileTarget(ctx context.Context, catalog projects.Service, name string) (settings.Target, error) {
	file, err := projects.CanonicalPath(name)
	if err != nil {
		return settings.Target{}, fmt.Errorf("resolve %s: %w", name, err)
	}
	info, err := os.Stat(file)
	if err != nil {
		return settings.Target{}, err
	}
	if info.IsDir() {
		return settings.Target{}, fmt.Errorf("%s is a directory: comment agents apply to files", file)
	}
	project, err := catalog.ForDir(ctx, filepath.Dir(file))
	if errors.Is(err, projects.ErrNotFound) {
		return settings.Target{Path: file}, nil
	}
	if err != nil {
		return settings.Target{}, err
	}
	root, err := projects.CanonicalPath(project.ResolvedDir())
	if err != nil {
		return settings.Target{}, fmt.Errorf("resolve project %q directory: %w", project.Name, err)
	}
	return settings.Target{Path: file, Project: &settings.Project{ID: project.ID, Dir: root}}, nil
}

// agentLayerFlags name the layer a set or delete edits.
type agentLayerFlags struct {
	layer, file string
	db, page    bool
}

func (f *agentLayerFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.layer, "layer", "", "Layer position: user, project, a folder relative to the project, or an absolute parent directory")
	cmd.Flags().StringVar(&f.file, "file", "", "Config file at the position: .gavel.yaml or .gavel.<name>.yaml (created when absent)")
	cmd.Flags().BoolVar(&f.db, "db", false, "Edit the position's database layer instead of a file")
	cmd.Flags().BoolVar(&f.page, "page", false, "Edit the layer's entry for the file instead of its default")
	_ = cmd.MarkFlagRequired("layer")
	cmd.MarkFlagsMutuallyExclusive("file", "db")
	cmd.MarkFlagsOneRequired("file", "db")
}

func (f agentLayerFlags) ref() (settings.Ref, error) {
	ref, err := parseLayerRef(f.layer)
	ref.File = f.file
	return ref, err
}

// parseLayerRef maps --layer to a layer position; the file is left unset.
func parseLayerRef(layer string) (settings.Ref, error) {
	switch {
	case layer == "":
		return settings.Ref{}, errors.New("--layer is required: user, project, a project folder or an absolute parent directory")
	case layer == string(settings.KindUser):
		return settings.Ref{Kind: settings.KindUser}, nil
	case layer == string(settings.KindProject):
		return settings.Ref{Kind: settings.KindProject}, nil
	case filepath.IsAbs(layer):
		return settings.Ref{Kind: settings.KindParent, Path: filepath.Clean(layer)}, nil
	}
	folder := path.Clean(filepath.ToSlash(layer))
	if folder == "." || folder == ".." || strings.HasPrefix(folder, "../") {
		return settings.Ref{}, fmt.Errorf("--layer %q is not a folder beneath the project: use project for its root", layer)
	}
	return settings.Ref{Kind: settings.KindPath, Path: folder}, nil
}

// agentFieldFlags are the comment agent fields a set writes.
type agentFieldFlags struct {
	prompt, instructions, runOptions, edits string
	presets                                 []string
	enabled, assign                         bool
}

func (f *agentFieldFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.prompt, "prompt", "", "Captain prompt the run starts (default "+verify.DefaultCommentAgentPrompt+")")
	cmd.Flags().StringSliceVar(&f.presets, "presets", nil, "Runtime presets replacing the prompt frontmatter's presets")
	cmd.Flags().StringVar(&f.instructions, "instructions", "", "Instructions appended to the thread the session is given")
	cmd.Flags().BoolVar(&f.enabled, "enabled", false, "Hand new human threads to the agent; --enabled=false opts the file(s) out")
	cmd.Flags().BoolVar(&f.assign, "assign", false, "Make the agent the thread's assignee as soon as it is created")
	cmd.Flags().StringVar(&f.runOptions, "run-options", "", "Run options as JSON: spec, variables")
	cmd.Flags().StringVar(&f.edits, "edits", "", "direct lets the agent edit files; propose has it answer with a proposed patch a person accepts")
}

func (f agentFieldFlags) agent(cmd *cobra.Command) (*verify.CommentAgent, error) {
	agent := &verify.CommentAgent{Prompt: f.prompt, Presets: f.presets, Instructions: f.instructions, Edits: f.edits}
	if err := agent.Validate(); err != nil {
		return nil, fmt.Errorf("--edits: %w", err)
	}
	if cmd.Flags().Changed("enabled") {
		agent.Enabled = &f.enabled
	}
	if cmd.Flags().Changed("assign") {
		agent.Assign = &f.assign
	}
	if f.runOptions != "" {
		decoder := json.NewDecoder(strings.NewReader(f.runOptions))
		decoder.DisallowUnknownFields()
		var options agents.RunOptions
		if err := decoder.Decode(&options); err != nil {
			return nil, fmt.Errorf("--run-options: %w", err)
		}
		if len(options.Presets) > 0 || options.Instructions != "" {
			return nil, errors.New("--run-options takes spec and variables: pass presets and instructions with --presets and --instructions")
		}
		agent.Spec, agent.Variables = options.Spec, options.Variables
	}
	if reflect.ValueOf(*agent).IsZero() {
		return nil, errors.New("set at least one agent field; delete removes the entry")
	}
	return agent, nil
}
