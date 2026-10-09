package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
)

type TodosStepsOptions struct {
	Refs []string `args:"true"`
}

func init() {
	clicky.AddNamedCommand("steps", todosCmd, TodosStepsOptions{}, func(opts TodosStepsOptions) (any, error) {
		return nil, runTodosSteps(opts.Refs)
	})
}

func runTodosSteps(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("steps accepts at most one TODO ID or alias")
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if len(args) == 0 {
		host, err := lifecycle.NewHost(nil, workDir, lifecycle.HostCLI)
		if err != nil {
			return err
		}
		fmt.Println(renderLifecycleSteps(host.Def.Definition()).ANSI())
		return nil
	}
	ctx := context.Background()
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	todoList, err := ResolveRequested(ctx, provider, workDir, args, todos.DiscoveryFilters{})
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
	}
	if len(todoList) != 1 {
		return fmt.Errorf("expected exactly one TODO matching %q, found %d", args[0], len(todoList))
	}
	host, err := lifecycle.NewHost(provider, workDir, lifecycle.HostCLI)
	if err != nil {
		return err
	}
	states, err := host.Steps(ctx, todoList[0])
	if err != nil {
		return err
	}
	fmt.Println(renderTodoSteps(todoList[0].Title, states).ANSI())
	return nil
}

// renderLifecycleSteps describes the definition itself: every step, its prompt
// reference, whether it is auxiliary, and its `when` predicate.
func renderLifecycleSteps(def lifecycle.Lifecycle) api.Text {
	tree := clicky.Text("Lifecycle "+def.Name, "text-blue-600 font-bold")
	for _, step := range def.Steps {
		line := clicky.Text("  "+step.Name, "text-green-600 font-bold").Append("  "+step.Prompt, "text-gray-500")
		if step.Auxiliary {
			line = line.Append("  auxiliary", "text-yellow-600")
		}
		if when := strings.Join(strings.Fields(step.When), " "); when != "" {
			line = line.Append("\n      when: "+when, "text-gray-400")
		}
		tree = tree.Add(api.Text{Content: "\n"}).Add(line)
	}
	return tree
}

// renderTodoSteps answers "what can I do with this todo now": each step's
// applicability and why, the one the lifecycle would run next, and the state
// of the step's last run when it has one.
func renderTodoSteps(title string, states []lifecycle.StepState) api.Text {
	tree := clicky.Text(title, "text-blue-600 font-bold")
	for _, state := range states {
		style := "text-gray-500"
		if state.Applicable {
			style = "text-green-600 font-bold"
		}
		line := clicky.Text("  "+state.Step.Name, style)
		if state.Suggested {
			line = line.Append("  next", "text-blue-600 font-bold")
		}
		line = line.Append("\n      "+state.Reason, "text-gray-400")
		if state.LastRun != nil {
			line = line.Append("\n      last run: "+state.LastRun.State, "text-gray-500")
		}
		tree = tree.Add(api.Text{Content: "\n"}).Add(line)
	}
	return tree
}
