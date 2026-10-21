package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// AskToolName is the one way a todo prompt asks a person something. The native
// AskUserQuestion is denied in every todo prompt, so a question never ends the
// run: it blocks inside it until the answer arrives.
const AskToolName = "todo_ask"

func askTool() api.ToolDefinition {
	question := object([]string{"text"}, map[string]any{
		"text":    str("The question that blocks progress"),
		"context": str("Why it matters: the files, trade-offs or evidence behind it"),
		"options": strs("Suggested answers, if any"),
	})
	return api.ToolDefinition{
		Name:        AskToolName,
		Description: "Ask the person running this TODO questions only they can answer, and wait for the answers.",
		Group:       Group,
		InputSchema: object([]string{"questions"}, map[string]any{
			"questions": map[string]any{"type": "array", "minItems": 1, "items": question},
		}),
		// The tool itself runs without approval: what it does is raise its own
		// question request through the run's broker.
		DefaultPermission: api.ToolPolicyAllow,
		Handler:           ask,
		ReadOnlyHint:      workspaceReadOnly(),
	}
}

type askInput struct {
	Questions []struct {
		Text    string   `json:"text"`
		Context string   `json:"context"`
		Options []string `json:"options"`
	} `json:"questions"`
}

func ask(ctx context.Context, input map[string]any) (any, error) {
	var in askInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if len(in.Questions) == 0 {
		return nil, fmt.Errorf("todo_ask needs at least one question")
	}
	questions := make([]api.TerminalQuestion, len(in.Questions))
	for i, question := range in.Questions {
		if strings.TrimSpace(question.Text) == "" {
			return nil, fmt.Errorf("todo_ask question %d has no text", i+1)
		}
		questions[i] = api.TerminalQuestion{Text: question.Text, Context: question.Context, Options: question.Options}
	}
	approve, ok := api.ApprovalFuncFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("todo_ask requires an approval broker")
	}
	decision, err := approve(ctx, api.ApprovalRequest{
		Tool: AskToolName, Input: input, Kind: api.ApprovalKindQuestion, Questions: questions,
	})
	if err != nil {
		return nil, fmt.Errorf("todo_ask: %w", err)
	}
	if !decision.Allow {
		return nil, fmt.Errorf("the questions were declined: %s", strings.TrimSpace(decision.Message))
	}
	answers, err := api.AnswersForQuestions(questions, decision.UpdatedInput["answers"])
	if err != nil {
		return nil, fmt.Errorf("todo_ask answers: %w", err)
	}
	out := make([]map[string]any, len(answers))
	for i, answer := range answers {
		out[i] = map[string]any{"question": answer.Question.Text, "answer": answer.Choices}
	}
	return out, nil
}
