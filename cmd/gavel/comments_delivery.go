package main

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// deliverTargetFlags choose a delivery target: a live session, or a new run of a prompt.
type deliverTargetFlags struct {
	session, prompt, group, runOptions string
	fresh                              bool
}

func (f *deliverTargetFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.session, "session", "", "Live Captain session to deliver into")
	cmd.Flags().StringVar(&f.prompt, "prompt", "", "Prompt a new managed session runs")
	cmd.Flags().StringVar(&f.group, "group", "", "Reuse the live session recorded for this group key (with --prompt)")
	cmd.Flags().StringVar(&f.runOptions, "run-options", "", "Run options as JSON: presets, spec, variables, instructions (with --prompt)")
	cmd.Flags().BoolVar(&f.fresh, "fresh", false, "Always start a new managed session, which becomes the --group's session (with --prompt); on a route, every routed comment starts its own session")
	cmd.MarkFlagsMutuallyExclusive("session", "prompt")
	cmd.MarkFlagsOneRequired("session", "prompt")
}

func (f deliverTargetFlags) target() (comments.DeliverTarget, error) {
	if f.session != "" {
		if f.group != "" || f.runOptions != "" || f.fresh {
			return comments.DeliverTarget{}, errors.New("--group, --run-options and --fresh apply to --prompt only")
		}
		return comments.DeliverTarget{Kind: "session", SessionID: f.session}, nil
	}
	target := comments.DeliverTarget{Kind: "run", Prompt: f.prompt, GroupKey: f.group, Fresh: f.fresh}
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
	cmd := &cobra.Command{Use: "deliver <id>", Short: "Send a comment thread to a live agent session, or start a managed session for it", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
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

func commentsRoutesCommand() *cobra.Command {
	routes := &cobra.Command{Use: "routes", Short: "Manage routes that deliver new comments to agents automatically"}
	var scope string
	list := &cobra.Command{Use: "list", Short: "List a scope's comment routes", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		service, err := ui.OpenComments(cmd.Context())
		if err != nil {
			return err
		}
		rows, err := service.Routes(cmd.Context(), scope)
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), rows)
	}}
	list.Flags().StringVar(&scope, "scope", "", "Scope URI, e.g. gavel://project/<name>")
	_ = list.MarkFlagRequired("scope")
	remove := &cobra.Command{Use: "delete <route-id>", Short: "Delete a comment route", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		service, err := ui.OpenComments(cmd.Context())
		if err != nil {
			return err
		}
		return service.DeleteRoute(cmd.Context(), args[0])
	}}
	routes.AddCommand(list, commentsRouteSetCommand(), remove)
	return routes
}

func commentsRouteSetCommand() *cobra.Command {
	var flags deliverTargetFlags
	var route comments.Route
	var disabled bool
	cmd := &cobra.Command{Use: "set", Short: "Create or replace a comment route; the longest matching prefix wins", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		target, err := flags.target()
		if err != nil {
			return err
		}
		route.Target, route.Enabled = target, !disabled
		service, err := ui.OpenComments(cmd.Context())
		if err != nil {
			return err
		}
		saved, err := service.SaveRoute(cmd.Context(), route)
		if err != nil {
			return err
		}
		return printJSON(cmd.OutOrStdout(), saved)
	}}
	cmd.Flags().StringVar(&route.ID, "id", "", "Route to replace; empty creates one")
	cmd.Flags().StringVar(&route.ScopeURI, "scope", "", "Scope URI, e.g. gavel://project/<name>")
	cmd.Flags().StringVar(&route.TargetPrefix, "prefix", "", "Target URI prefix inside the scope")
	cmd.Flags().StringVar(&route.WorkflowID, "workflow", comments.HumanReview().ID, "Workflow of the comments the route delivers")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Opt the prefix out of delivery")
	_ = cmd.MarkFlagRequired("scope")
	_ = cmd.MarkFlagRequired("prefix")
	flags.register(cmd)
	return cmd
}
