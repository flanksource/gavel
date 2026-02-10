package claudehistory

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

// map[header:Init Behavior multiSelect:false options:[map[description:Automatically initialize PostgreSQL database cluster if data directory is missing or empty label:Run initdb to create a new cluster] map[description:Initialize and automatically optimize configuration label:Run initdb and apply pg-tune] map[description:Initialize with specific locale, encoding, or authentication settings label:Run initdb with custom settings]] question:What should --auto-init do when the PostgreSQL data directory doesn't exist?] map[header:Existing Data multiSelect:false options:[map[description:Do nothing if database already exists, proceed with start label:Skip initialization and continue] map[description:Fail with error message if database already exists label:Show error and exit] map[description:Prompt user before proceeding label:Ask user for confirmation]] question:What should happen if the data directory already exists?]]
type Questions struct {
	Questions []Question `json:"questions"`
}

type ExitPlanMode struct {
	Plan string `json:"plan"`
}

type Question struct {
	Description string   `json:"description"`
	Label       string   `json:"label"`
	Header      string   `json:"header,omitempty"`
	MultiSelect bool     `json:"multiSelect,omitempty"`
	Options     []Option `json:"options,omitempty"`
}

type Option struct {
	Description string `json:"description"`
	Label       string `json:"label"`
}

type TodoWrite struct {
	Todos []Todo `json:"todos"`
}

type Todo struct {
	ActiveForm string `json:"activeForm,omitempty"`
	Content    string `json:"content,omitempty"`
	Status     string `json:"status,omitempty"`
}

type Bash struct {
	Command string `json:"command"`
	CWD     string `json:"cwd,omitempty"`
}

func (b Bash) ToolName() string {
	return "Bash"
}

func (b Bash) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "💻", Iconify: "codicon:terminal", Style: "muted"}).Append(" Bash", "text-green-600 font-medium")
	if b.CWD != "" {
		text = text.Append(" (", "text-gray-400").Append(b.CWD, "text-gray-500").Append(")", "text-gray-400")
	}
	if b.Command != "" {
		text = text.NewLine().Add(clicky.CodeBlock(b.Command, "bash"))
	}
	return text
}

type Read struct {
	Path   string  `json:"path"`
	Limit  float64 `json:"limit,omitempty"`
	Offset float64 `json:"offset,omitempty"`
}

func (r Read) ToolName() string {
	return "Read"
}

func (r Read) Pretty() api.Text {
	text := clicky.Text("").Add(icons.File).Append(" Read", "text-blue-600 font-medium")
	if r.Path != "" {
		text = text.Append(": ", "text-gray-600").Append(r.Path, "text-gray-800")
		if r.Limit > 0 || r.Offset > 0 {
			text = text.Append(fmt.Sprintf(" [%.0f:%.0f]", r.Offset, r.Limit), "text-gray-500")
		}
	}
	return text
}

type Write struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
}

func (w Write) ToolName() string {
	return "Write"
}

func (w Write) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "✏️", Iconify: "codicon:edit", Style: "muted"}).Append(" Write", "text-orange-600 font-medium")
	if w.Path != "" {
		text = text.Append(": ", "text-gray-600").Append(w.Path, "text-gray-800")
	}
	if w.Content != "" {
		preview := w.Content
		if len(preview) > 100 {
			preview = preview[:97] + "..."
		}
		text = text.NewLine().Append("Content: ", "text-gray-500").Append(preview, "text-gray-700")
	}
	return text
}

type Edit struct {
	Path      string `json:"path"`
	OldString string `json:"old_string,omitempty"`
	NewString string `json:"new_string,omitempty"`
}

func (e Edit) ToolName() string {
	return "Edit"
}

func (e Edit) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "✏️", Iconify: "codicon:edit", Style: "muted"}).Append(" Edit", "text-purple-600 font-medium")
	if e.Path != "" {
		text = text.Append(": ", "text-gray-600").Append(e.Path, "text-gray-800")
	}
	if e.OldString != "" && e.NewString != "" {
		oldPreview := e.OldString
		newPreview := e.NewString
		if len(oldPreview) > 50 {
			oldPreview = oldPreview[:47] + "..."
		}
		if len(newPreview) > 50 {
			newPreview = newPreview[:47] + "..."
		}
		text = text.NewLine().Append("Replace: ", "text-gray-500").Append(oldPreview, "text-red-600").
			Append(" → ", "text-gray-400").Append(newPreview, "text-green-600")
	}
	return text
}

type Grep struct {
	OutputMode string `json:"output_mode,omitempty"`
	Glob       string `json:"glob,omitempty"`
	Count      int    `json:"-n,omitempty"`
}

func (g Grep) ToolName() string {
	return "Grep"
}

func (g Grep) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" Grep", "text-yellow-600 font-medium")
	if g.Glob != "" {
		text = text.Append(": ", "text-gray-600").Append(g.Glob, "text-gray-800")
	}
	if g.OutputMode != "" {
		text = text.Append(" (", "text-gray-400").Append(g.OutputMode, "text-gray-500").Append(")", "text-gray-400")
	}
	return text
}

type Task struct {
	Description  string `json:"description"`
	Prompt       string `json:"prompt,omitempty"`
	SubAgentType string `json:"subagent_type,omitempty"`
}

func (t Task) ToolName() string {
	return "Task"
}

func (t Task) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Package).Append(" Task", "text-indigo-600 font-medium")
	if t.Description != "" {
		text = text.Append(": ", "text-gray-600").Append(t.Description, "text-gray-800")
	}
	if t.SubAgentType != "" {
		text = text.Append(" (", "text-gray-400").Append(t.SubAgentType, "text-gray-500").Append(")", "text-gray-400")
	}
	return text
}

type MultiEdit struct {
	Edits []Edit `json:"edits"`
}

func (m MultiEdit) ToolName() string {
	return "MultiEdit"
}

func (m MultiEdit) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "✏️", Iconify: "codicon:edit", Style: "muted"}).Append(" MultiEdit", "text-purple-600 font-medium")
	if len(m.Edits) > 0 {
		text = text.Append(fmt.Sprintf(" (%d edits)", len(m.Edits)), "text-gray-500")
	}
	return text
}

type Glob struct {
	Pattern string `json:"pattern"`
}

func (g Glob) ToolName() string {
	return "Glob"
}

func (g Glob) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" Glob", "text-cyan-600 font-medium")
	if g.Pattern != "" {
		text = text.Append(": ", "text-gray-600").Append(g.Pattern, "text-gray-800")
	}
	return text
}

type WebFetch struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt,omitempty"`
}

func (w WebFetch) ToolName() string {
	return "WebFetch"
}

func (w WebFetch) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Cloud).Append(" WebFetch", "text-blue-600 font-medium")
	if w.URL != "" {
		text = text.Append(": ", "text-gray-600").Append(w.URL, "text-blue-700 underline")
	}
	if w.Prompt != "" {
		prompt := w.Prompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "..."
		}
		text = text.NewLine().Append("Prompt: ", "text-gray-500").Append(prompt, "text-gray-700")
	}
	return text
}

type BashOutput struct {
	BashId string `json:"bash_id"`
}

func (b BashOutput) ToolName() string {
	return "BashOutput"
}

func (b BashOutput) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "📋", Iconify: "codicon:output", Style: "muted"}).Append(" BashOutput", "text-green-600 font-medium")
	if b.BashId != "" {
		text = text.Append(" [", "text-gray-400").Append(b.BashId, "text-gray-600").Append("]", "text-gray-400")
	}
	return text
}

type KillShell struct {
	ShellId string `json:"shell_id"`
}

func (k KillShell) ToolName() string {
	return "KillShell"
}

func (k KillShell) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🔴", Iconify: "codicon:debug-stop", Style: "muted"}).Append(" KillShell", "text-red-600 font-medium")
	if k.ShellId != "" {
		text = text.Append(" [", "text-gray-400").Append(k.ShellId, "text-gray-600").Append("]", "text-gray-400")
	}
	return text
}

type WebSearch struct {
	Query string `json:"query"`
}

func (w WebSearch) ToolName() string {
	return "WebSearch"
}

func (w WebSearch) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" WebSearch", "text-purple-600 font-medium")
	if w.Query != "" {
		text = text.Append(": ", "text-gray-600").Append(w.Query, "text-gray-800")
	}
	return text
}

type Skill struct {
	Command string `json:"command"`
}

func (s Skill) ToolName() string {
	return "Skill"
}

func (s Skill) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Info).Append(" Skill", "text-teal-600 font-medium")
	if s.Command != "" {
		text = text.Append(": ", "text-gray-600").Append(s.Command, "text-gray-800")
	}
	return text
}

type McpIconifyGetIcon struct {
	Set  string `json:"set"`
	Icon string `json:"icon"`
}

func (m McpIconifyGetIcon) ToolName() string {
	return "mcp__iconify__get_icon"
}

func (m McpIconifyGetIcon) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🎨", Iconify: "mdi:palette", Style: "muted"}).Append(" Iconify Get Icon", "text-pink-600 font-medium")
	if m.Set != "" && m.Icon != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Set, "text-gray-700").Append("/", "text-gray-400").Append(m.Icon, "text-gray-800")
	}
	return text
}

type McpIconifySearchIcons struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

func (m McpIconifySearchIcons) ToolName() string {
	return "mcp__iconify__search_icons"
}

func (m McpIconifySearchIcons) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" Iconify Search", "text-pink-600 font-medium")
	if m.Query != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Query, "text-gray-800")
	}
	if m.Limit > 0 {
		text = text.Append(fmt.Sprintf(" (limit: %d)", m.Limit), "text-gray-500")
	}
	return text
}

type McpReactIconsGetLibraryIcons struct {
	LibraryPrefix string `json:"libraryPrefix"`
	Limit         int    `json:"limit,omitempty"`
}

func (m McpReactIconsGetLibraryIcons) ToolName() string {
	return "mcp__react-icons__get_library_icons"
}

func (m McpReactIconsGetLibraryIcons) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "⚛️", Iconify: "mdi:react", Style: "muted"}).Append(" React Icons Library", "text-cyan-600 font-medium")
	if m.LibraryPrefix != "" {
		text = text.Append(": ", "text-gray-600").Append(m.LibraryPrefix, "text-gray-800")
	}
	if m.Limit > 0 {
		text = text.Append(fmt.Sprintf(" (limit: %d)", m.Limit), "text-gray-500")
	}
	return text
}

type McpReactIconsSearchIcons struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

func (m McpReactIconsSearchIcons) ToolName() string {
	return "mcp__react-icons__search_icons"
}

func (m McpReactIconsSearchIcons) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" React Icons Search", "text-cyan-600 font-medium")
	if m.Query != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Query, "text-gray-800")
	}
	if m.Limit > 0 {
		text = text.Append(fmt.Sprintf(" (limit: %d)", m.Limit), "text-gray-500")
	}
	return text
}

type McpLucideSearchIcons struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

func (m McpLucideSearchIcons) ToolName() string {
	return "mcp__lucide__search_icons"
}

func (m McpLucideSearchIcons) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" Lucide Search", "text-indigo-600 font-medium")
	if m.Query != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Query, "text-gray-800")
	}
	if m.Limit > 0 {
		text = text.Append(fmt.Sprintf(" (limit: %d)", m.Limit), "text-gray-500")
	}
	return text
}

type McpIcons8GetIconSvg struct {
	IconId string `json:"icon_id"`
}

func (m McpIcons8GetIconSvg) ToolName() string {
	return "mcp__icons8mcp__get_icon_svg"
}

func (m McpIcons8GetIconSvg) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🎨", Iconify: "mdi:palette", Style: "muted"}).Append(" Icons8 Get SVG", "text-orange-600 font-medium")
	if m.IconId != "" {
		text = text.Append(": ", "text-gray-600").Append(m.IconId, "text-gray-800")
	}
	return text
}

type McpIcons8SearchIcons struct {
	Query    string `json:"query"`
	Amount   int    `json:"amount,omitempty"`
	Platform string `json:"platform,omitempty"`
}

func (m McpIcons8SearchIcons) ToolName() string {
	return "mcp__icons8mcp__search_icons"
}

func (m McpIcons8SearchIcons) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Search).Append(" Icons8 Search", "text-orange-600 font-medium")
	if m.Query != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Query, "text-gray-800")
	}
	if m.Platform != "" {
		text = text.Append(" [", "text-gray-400").Append(m.Platform, "text-gray-600").Append("]", "text-gray-400")
	}
	if m.Amount > 0 {
		text = text.Append(fmt.Sprintf(" (limit: %d)", m.Amount), "text-gray-500")
	}
	return text
}

type McpGeminiGenerateContent struct {
	UserPrompt string                   `json:"user_prompt"`
	Model      string                   `json:"model,omitempty"`
	Files      []map[string]interface{} `json:"files,omitempty"`
}

func (m McpGeminiGenerateContent) ToolName() string {
	return "mcp__gemini__generate_content"
}

func (m McpGeminiGenerateContent) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "✨", Iconify: "mdi:sparkles", Style: "muted"}).Append(" Gemini", "text-purple-600 font-medium")
	if m.Model != "" {
		text = text.Append(" [", "text-gray-400").Append(m.Model, "text-gray-600").Append("]", "text-gray-400")
	}
	if m.UserPrompt != "" {
		prompt := m.UserPrompt
		if len(prompt) > 60 {
			prompt = prompt[:57] + "..."
		}
		text = text.NewLine().Append(prompt, "text-gray-700")
	}
	if len(m.Files) > 0 {
		text = text.Append(fmt.Sprintf(" (%d files)", len(m.Files)), "text-gray-500")
	}
	return text
}

type McpPlaywrightBrowserClick struct {
	Element string `json:"element"`
	Ref     string `json:"ref"`
}

func (m McpPlaywrightBrowserClick) ToolName() string {
	return "mcp__playwright__browser_click"
}

func (m McpPlaywrightBrowserClick) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🖱️", Iconify: "mdi:cursor-default-click", Style: "muted"}).Append(" Browser Click", "text-blue-600 font-medium")
	if m.Element != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Element, "text-gray-800")
	}
	return text
}

type McpPlaywrightBrowserClose struct{}

func (m McpPlaywrightBrowserClose) ToolName() string {
	return "mcp__playwright__browser_close"
}

func (m McpPlaywrightBrowserClose) Pretty() api.Text {
	return clicky.Text("").Add(icons.Icon{Unicode: "❌", Iconify: "mdi:close", Style: "muted"}).Append(" Browser Close", "text-red-600 font-medium")
}

type McpPlaywrightBrowserConsoleMessages struct{}

func (m McpPlaywrightBrowserConsoleMessages) ToolName() string {
	return "mcp__playwright__browser_console_messages"
}

func (m McpPlaywrightBrowserConsoleMessages) Pretty() api.Text {
	return clicky.Text("").Add(icons.Icon{Unicode: "📋", Iconify: "codicon:output", Style: "muted"}).Append(" Browser Console", "text-gray-600 font-medium")
}

type McpPlaywrightBrowserEvaluate struct {
	Function string `json:"function"`
}

func (m McpPlaywrightBrowserEvaluate) ToolName() string {
	return "mcp__playwright__browser_evaluate"
}

func (m McpPlaywrightBrowserEvaluate) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "⚙️", Iconify: "codicon:code", Style: "muted"}).Append(" Browser Evaluate", "text-purple-600 font-medium")
	if m.Function != "" {
		preview := m.Function
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		text = text.NewLine().Append(preview, "text-gray-700")
	}
	return text
}

type McpPlaywrightBrowserNavigate struct {
	URL string `json:"url"`
}

func (m McpPlaywrightBrowserNavigate) ToolName() string {
	return "mcp__playwright__browser_navigate"
}

func (m McpPlaywrightBrowserNavigate) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🌐", Iconify: "mdi:web", Style: "muted"}).Append(" Browser Navigate", "text-blue-600 font-medium")
	if m.URL != "" {
		text = text.Append(": ", "text-gray-600").Append(m.URL, "text-blue-700 underline")
	}
	return text
}

type McpPlaywrightBrowserNetworkRequests struct{}

func (m McpPlaywrightBrowserNetworkRequests) ToolName() string {
	return "mcp__playwright__browser_network_requests"
}

func (m McpPlaywrightBrowserNetworkRequests) Pretty() api.Text {
	return clicky.Text("").Add(icons.Icon{Unicode: "🌐", Iconify: "mdi:network", Style: "muted"}).Append(" Browser Network", "text-green-600 font-medium")
}

type McpPlaywrightBrowserPressKey struct {
	Key string `json:"key"`
}

func (m McpPlaywrightBrowserPressKey) ToolName() string {
	return "mcp__playwright__browser_press_key"
}

func (m McpPlaywrightBrowserPressKey) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "⌨️", Iconify: "mdi:keyboard", Style: "muted"}).Append(" Browser Press Key", "text-yellow-600 font-medium")
	if m.Key != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Key, "text-gray-800")
	}
	return text
}

type McpPlaywrightBrowserSnapshot struct{}

func (m McpPlaywrightBrowserSnapshot) ToolName() string {
	return "mcp__playwright__browser_snapshot"
}

func (m McpPlaywrightBrowserSnapshot) Pretty() api.Text {
	return clicky.Text("").Add(icons.Icon{Unicode: "📸", Iconify: "mdi:camera", Style: "muted"}).Append(" Browser Snapshot", "text-purple-600 font-medium")
}

type McpPlaywrightBrowserTakeScreenshot struct {
	Filename string `json:"filename,omitempty"`
	FullPage bool   `json:"fullPage,omitempty"`
}

func (m McpPlaywrightBrowserTakeScreenshot) ToolName() string {
	return "mcp__playwright__browser_take_screenshot"
}

func (m McpPlaywrightBrowserTakeScreenshot) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "📷", Iconify: "mdi:camera", Style: "muted"}).Append(" Browser Screenshot", "text-indigo-600 font-medium")
	if m.Filename != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Filename, "text-gray-800")
	}
	if m.FullPage {
		text = text.Append(" [full page]", "text-gray-500")
	}
	return text
}

type McpPlaywrightBrowserTripleClick struct {
	Element string `json:"element"`
	Ref     string `json:"ref"`
}

func (m McpPlaywrightBrowserTripleClick) ToolName() string {
	return "mcp__playwright__browser_triple_click"
}

func (m McpPlaywrightBrowserTripleClick) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "🖱️", Iconify: "mdi:cursor-default-click", Style: "muted"}).Append(" Browser Triple Click", "text-blue-600 font-medium")
	if m.Element != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Element, "text-gray-800")
	}
	return text
}

type McpPlaywrightBrowserType struct {
	Element string `json:"element"`
	Ref     string `json:"ref"`
	Text    string `json:"text"`
}

func (m McpPlaywrightBrowserType) ToolName() string {
	return "mcp__playwright__browser_type"
}

func (m McpPlaywrightBrowserType) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "⌨️", Iconify: "mdi:keyboard", Style: "muted"}).Append(" Browser Type", "text-green-600 font-medium")
	if m.Element != "" {
		text = text.Append(": ", "text-gray-600").Append(m.Element, "text-gray-800")
	}
	if m.Text != "" {
		text = text.Append(" → ", "text-gray-400").Append(fmt.Sprintf("\"%s\"", m.Text), "text-gray-700")
	}
	return text
}

type McpPlaywrightBrowserWaitFor struct {
	Time int `json:"time,omitempty"`
}

func (m McpPlaywrightBrowserWaitFor) ToolName() string {
	return "mcp__playwright__browser_wait_for"
}

func (m McpPlaywrightBrowserWaitFor) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Icon{Unicode: "⏳", Iconify: "mdi:timer-sand", Style: "muted"}).Append(" Browser Wait", "text-orange-600 font-medium")
	if m.Time > 0 {
		text = text.Append(fmt.Sprintf(": %ds", m.Time), "text-gray-600")
	}
	return text
}

type Tool interface {
	ToolName() string
	Pretty() api.Text
}

var Tools = []Tool{
	Bash{},
	Read{},
	Write{},
	Edit{},
	Grep{},
	Glob{},
	WebFetch{},
	Skill{},
	MultiEdit{},
	Task{},
	BashOutput{},
	KillShell{},
	WebSearch{},
	McpIconifyGetIcon{},
	McpIconifySearchIcons{},
	McpReactIconsGetLibraryIcons{},
	McpReactIconsSearchIcons{},
	McpLucideSearchIcons{},
	McpIcons8GetIconSvg{},
	McpIcons8SearchIcons{},
	McpGeminiGenerateContent{},
	McpPlaywrightBrowserClick{},
	McpPlaywrightBrowserClose{},
	McpPlaywrightBrowserConsoleMessages{},
	McpPlaywrightBrowserEvaluate{},
	McpPlaywrightBrowserNavigate{},
	McpPlaywrightBrowserNetworkRequests{},
	McpPlaywrightBrowserPressKey{},
	McpPlaywrightBrowserSnapshot{},
	McpPlaywrightBrowserTakeScreenshot{},
	McpPlaywrightBrowserTripleClick{},
	McpPlaywrightBrowserType{},
	McpPlaywrightBrowserWaitFor{},
}

func ParseTool(jsonString string) any {
	// Use uir.NewPolymorphicRegistry to parse the JSON string into the appropriate struct based on the "tool" field.
	return nil
}
