package types

import (
	"fmt"
	"strings"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

// VerifyReportText renders a verification report as an indented tree: the
// verdict line, then every node with its status icon, then the acceptance
// criteria. indent prefixes the whole block so a caller can nest it under a
// TODO row.
//
// A passing node is shown as well as a failing one: `gavel todos check` is
// answering "is this done?", and a tree that lists only what broke leaves the
// reader unable to tell a thorough green run from an empty one.
func VerifyReportText(report captainapi.VerifyReport, indent string) api.Text {
	text := clicky.Text(indent).Add(verifyStateIcon(report.State)).
		Append(" "+string(report.State), verifyStateStyle(report.State))
	if report.Reason != "" {
		text = text.Append(" — "+report.Reason, "text-gray-500")
	}
	text = text.NewLine()
	text = appendVerifyNodes(text, report.Tests, indent+"  ")
	for _, item := range report.Checklist {
		icon := icons.Fail
		if item.Passed != nil && *item.Passed {
			icon = icons.Pass
		}
		text = text.Append(indent + "  ").Add(icon).Append(" " + item.Item)
		if item.Message != "" {
			text = text.Append(" — "+item.Message, "text-gray-500")
		}
		text = text.NewLine()
	}
	return text
}

func appendVerifyNodes(text api.Text, nodes []captainapi.VerifyNode, indent string) api.Text {
	for i := range nodes {
		node := &nodes[i]
		text = text.Append(indent).Add(verifyNodeIcon(node)).Append(" " + node.Name)
		if node.Duration > 0 {
			text = text.Append(fmt.Sprintf(" (%s)", node.Duration.Round(1e6)), "text-gray-500")
		}
		if node.Message != "" {
			text = text.Append(" — "+strings.SplitN(node.Message, "\n", 2)[0], "text-red-500")
		}
		text = text.NewLine()
		text = appendVerifyNodes(text, node.Children, indent+"  ")
	}
	return text
}

func verifyNodeIcon(node *captainapi.VerifyNode) api.Text {
	switch {
	case len(node.Children) > 0:
		return clicky.Text("•", "text-gray-500")
	case node.TimedOut, node.Failed:
		return iconText(icons.Fail)
	case node.Warned:
		return iconText(icons.Warning)
	case node.Skipped:
		return iconText(icons.Skip)
	case node.Passed:
		return iconText(icons.Pass)
	default:
		return clicky.Text("◦", "text-gray-500")
	}
}

func verifyStateIcon(state captainapi.VerifyState) api.Text {
	switch state {
	case captainapi.VerifyStatePassed:
		return iconText(icons.Pass)
	case captainapi.VerifyStateWarned:
		return iconText(icons.Warning)
	case captainapi.VerifyStateSkipped:
		return iconText(icons.Skip)
	case captainapi.VerifyStateFailed, captainapi.VerifyStateErrored,
		captainapi.VerifyStateTimedOut, captainapi.VerifyStateCancelled:
		return iconText(icons.Fail)
	default:
		return clicky.Text("◦", "text-gray-500")
	}
}

func iconText(icon icons.Icon) api.Text { return clicky.Text("").Add(icon) }

func verifyStateStyle(state captainapi.VerifyState) string {
	switch state {
	case captainapi.VerifyStatePassed:
		return "text-green-600"
	case captainapi.VerifyStateWarned:
		return "text-yellow-600"
	case captainapi.VerifyStateFailed, captainapi.VerifyStateErrored,
		captainapi.VerifyStateTimedOut, captainapi.VerifyStateCancelled:
		return "text-red-600"
	default:
		return "text-gray-500"
	}
}
