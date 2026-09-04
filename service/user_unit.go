package service

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type unitData struct {
	ExecStart        string
	WorkingDirectory string
}

const userUnitTemplate = `[Unit]
Description=Gavel PR UI (pr list --all --ui)
After=default.target

[Service]
Type=simple
WorkingDirectory={{.WorkingDirectory}}
ExecStart={{.ExecStart}}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`

func renderUserUnit(shell, bin, home string) (string, error) {
	t, err := template.New("unit").Parse(userUnitTemplate)
	if err != nil {
		return "", fmt.Errorf("parse unit template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, unitData{
		ExecStart:        renderSystemdCommand(userShellInvocation(shell, bin)),
		WorkingDirectory: quoteSystemdArgument(home),
	}); err != nil {
		return "", fmt.Errorf("render unit template: %w", err)
	}
	return buf.String(), nil
}

func renderSystemdCommand(arguments []string) string {
	quoted := make([]string, len(arguments))
	for i, argument := range arguments {
		quoted[i] = quoteSystemdArgument(argument)
	}
	return strings.Join(quoted, " ")
}

func quoteSystemdArgument(argument string) string {
	argument = strings.ReplaceAll(argument, "$", "$$")
	argument = strings.ReplaceAll(argument, "%", "%%")
	return strconv.Quote(argument)
}
