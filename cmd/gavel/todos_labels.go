package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/spf13/cobra"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
)

var (
	todoLabelColor       string
	todoLabelIcon        string
	todoLabelDescription string
	todoLabelGlobal      bool
)

var todosLabelsCmd = &cobra.Command{
	Use:          "labels",
	Aliases:      []string{"label", "tags"},
	SilenceUsage: true,
	Short:        "Inspect and edit how TODO labels are coloured and iconified",
	Long: `Labels carry a colour, an icon and a description so a backlog can be scanned by tag.

A definition is global (every workspace) or scoped to this workspace, which shadows
the global one. Well-known names (bug, security, docs, …) have built-in defaults, and
anything still undefined gets a stable colour derived from its name — so labels are
never colourless, and defining one is always an override rather than a repaint.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
}

var todosLabelsSetCmd = &cobra.Command{
	Use:          "set <name>",
	SilenceUsage: true,
	Short:        "Define or update a label's colour, icon, and description",
	Example: `  gavel todos labels set flaky --color amber --icon warning
  gavel todos labels set security --color rose --icon lock --global
  gavel todos labels set area --color teal --description "Which subsystem"`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosLabelsSet,
}

var todosLabelsRemoveCmd = &cobra.Command{
	Use:          "rm <name>",
	Aliases:      []string{"remove", "delete"},
	SilenceUsage: true,
	Short:        "Retire a label from this project, stripping it from every TODO",
	Long: `Removing a label from this project deletes its definition AND strips it from every
TODO in the workspace, so a retired label stops reappearing in filters and facets.

--global removes the definition shared by every workspace instead. That scope spans
every project on this machine, so it is presentation only: TODOs keep the label and
it falls back to its built-in or derived colour.`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosLabelsRemove,
}

// TodosLabelsListOptions is the clicky-registered option surface for
// `todos labels list`, so the command gains -o json/yaml alongside the table.
type TodosLabelsListOptions struct {
	Global bool `flag:"global" description:"List only the definitions that apply to every workspace"`
}

func init() {
	todosCmd.AddCommand(todosLabelsCmd)
	todosLabelsCmd.AddCommand(todosLabelsSetCmd, todosLabelsRemoveCmd)

	todosLabelsSetCmd.Flags().StringVar(&todoLabelColor, "color", "",
		"Palette hue: "+strings.Join(labels.PaletteStrings(), ", ")+" (default: derived from the name)")
	todoLabelsIconUsage := "Icon name from the clicky registry, e.g. debug, lock, performance"
	todosLabelsSetCmd.Flags().StringVar(&todoLabelIcon, "icon", "", todoLabelsIconUsage)
	todosLabelsSetCmd.Flags().StringVar(&todoLabelDescription, "description", "", "What the label means")
	todosLabelsSetCmd.Flags().BoolVar(&todoLabelGlobal, "global", false,
		"Apply to every workspace instead of just this one")
	todosLabelsRemoveCmd.Flags().BoolVar(&todoLabelGlobal, "global", false,
		"Remove the global definition instead of this workspace's override")

	// AddNamedCommand, not AddCommand: the latter derives the subcommand name
	// from the handler ("todos-labels-list"), which reads wrong under `labels`.
	listCmd := clicky.AddNamedCommand("list", todosLabelsCmd, TodosLabelsListOptions{}, runTodosLabelsList)
	listCmd.Aliases = []string{"ls"}
	listCmd.Short = "List label definitions and how many TODOs use each"
}

// labelDefinitionRow is one rendered row of `todos labels list`: the definition
// plus how many TODOs currently carry it.
type labelDefinitionRow struct {
	labels.Definition
	Todos int `json:"todos"`
}

func (r labelDefinitionRow) PrettyRow(opts interface{}) map[string]api.Text {
	row := r.Definition.PrettyRow(opts)
	if r.Todos > 0 {
		row["Todos"] = clicky.Text(fmt.Sprintf("%d", r.Todos), "order-4 text-muted")
	}
	return row
}

func runTodosLabelsList(opts TodosLabelsListOptions) (any, error) {
	store, err := todoLabelStore()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	definitions, err := store.LabelDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := store.LabelCounts(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]labelDefinitionRow, 0, len(definitions))
	for _, definition := range definitions {
		if opts.Global && definition.Scope != labels.ScopeGlobal {
			continue
		}
		rows = append(rows, labelDefinitionRow{Definition: definition, Todos: counts[definition.Name]})
	}

	// Labels in use lead, so a long taxonomy still opens on what matters.
	sort.SliceStable(rows, func(i, j int) bool {
		if (rows[i].Todos > 0) != (rows[j].Todos > 0) {
			return rows[i].Todos > 0
		}
		return rows[i].Name < rows[j].Name
	})
	return rows, nil
}

func runTodosLabelsSet(_ *cobra.Command, args []string) error {
	store, err := todoLabelStore()
	if err != nil {
		return err
	}

	name := labels.Normalize(args[0])
	if name == "" {
		return fmt.Errorf("label name is required")
	}

	// An omitted colour keeps the hue the label already rendered with, so
	// "describe this label" never doubles as "repaint it".
	color := labels.Hash(name)
	if raw := strings.TrimSpace(todoLabelColor); raw != "" {
		if color, err = labels.ParseColor(raw); err != nil {
			return err
		}
	}

	icon := labels.Normalize(todoLabelIcon)
	if icon != "" && !labels.IsIcon(icon) {
		return fmt.Errorf("unknown icon %q; %s", todoLabelIcon, suggestIcons(icon))
	}

	saved, err := store.SetLabelDefinition(context.Background(), labels.Definition{
		Name:        name,
		Color:       color,
		Icon:        icon,
		Description: strings.TrimSpace(todoLabelDescription),
	}, todoLabelGlobal)
	if err != nil {
		return err
	}

	rendered, err := clicky.Format(labelDefinitionRow{Definition: saved})
	if err != nil {
		return err
	}
	fmt.Println(rendered)
	return nil
}

func runTodosLabelsRemove(_ *cobra.Command, args []string) error {
	store, err := todoLabelStore()
	if err != nil {
		return err
	}
	removal, err := store.DeleteLabelDefinition(context.Background(), labels.Normalize(args[0]), todoLabelGlobal)
	if err != nil {
		return err
	}

	if todoLabelGlobal {
		fmt.Printf("%s. TODOs keep the label — a global definition is presentation only,\n"+
			"so it now renders with the built-in or derived colour.\n", removal)
		return nil
	}
	fmt.Printf("%s.\n", removal)
	return nil
}

// todoLabelStore resolves the provider's label-definition capability. Label
// presentation lives in PostgreSQL, so a provider without it says so rather
// than reporting an empty taxonomy.
func todoLabelStore() (todos.LabelDefinitionProvider, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return nil, err
	}
	store, ok := provider.(todos.LabelDefinitionProvider)
	if !ok {
		return nil, fmt.Errorf("TODO provider does not support label definitions; native PostgreSQL storage is required")
	}
	return store, nil
}

// suggestIcons offers the registry names sharing a prefix with the miss, so a
// typo is corrected in one step instead of by dumping hundreds of names.
func suggestIcons(attempt string) string {
	var near []string
	for _, name := range labels.IconNames() {
		if strings.HasPrefix(name, attempt[:min(len(attempt), 3)]) {
			near = append(near, name)
		}
	}
	if len(near) == 0 {
		return fmt.Sprintf("the clicky icon registry has %d names", len(labels.IconNames()))
	}
	if len(near) > 12 {
		near = near[:12]
	}
	return "did you mean: " + strings.Join(near, ", ")
}
