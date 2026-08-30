package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

var todosPromptsCmd = &cobra.Command{
	Use:          "prompts",
	SilenceUsage: true,
	Short:        "List the prompts `gavel todos run --prompt` accepts",
	Long: `List every runnable TODO prompt: the built-ins and any declared under
.gavel.yaml todos.prompts.

A prompt names the instructions; its class names the behaviour. Run implements and
commits; plan and triage do neither. That separation is why a project can add a
prompt without gavel gaining a run mode.`,
	Example: `  gavel todos prompts
  gavel todos run 3f2a1b --prompt triage`,
	Args: cobra.NoArgs,
	RunE: runTodosPrompts,
}

func init() {
	todosCmd.AddCommand(todosPromptsCmd)
}

func runTodosPrompts(_ *cobra.Command, _ []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return fmt.Errorf("load .gavel.yaml: %w", err)
	}
	catalog, err := todoprompt.NewCatalog(cfg.Todos)
	if err != nil {
		return err
	}

	tree := clicky.Text("Todo prompts", "text-blue-600 font-bold")
	for _, def := range catalog.List() {
		line := clicky.Text("  "+def.Name, "text-green-600 font-bold").
			Append("  "+string(def.Class), "text-gray-500").
			Append("  "+def.Origin, "text-gray-400")
		if def.Title != "" {
			line = line.Append("\n      "+def.Title, "text-gray-500")
		}
		tree = tree.Add(api.Text{Content: "\n"}).Add(line)
	}
	fmt.Println(tree.ANSI())
	return nil
}

// resolveCheckConcurrency reads the configured definition-of-done concurrency,
// with the --concurrency flag on top. A bad config is not worth failing a check
// run over, so an unreadable .gavel.yaml falls through to the default here — the
// run itself will report it.
func resolveCheckConcurrency(workDir string) int {
	if checkConcurrency > 0 {
		return checkConcurrency
	}
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return 0
	}
	return cfg.Todos.CheckConcurrency
}
