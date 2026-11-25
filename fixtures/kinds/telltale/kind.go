package telltale

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

type Kind struct{}

func init()               { fixtures.RegisterKind(Kind{}) }
func (Kind) Name() string { return "telltale" }
func (Kind) Schema() json.RawMessage {
	return json.RawMessage(`{
 "type":"object","additionalProperties":false,"required":["agent"],
 "properties":{
 "server":{"type":"string","description":"Telltale HTTP server; defaults to TELLTALE_URL."},
 "token":{"type":"string","description":"Optional API bearer token."},
 "agent":{"type":"string","minLength":1},
 "operation":{"type":"string","minLength":1,"description":"Live operation name, including extension operations."},
 "capture":{"type":"string","minLength":1,"description":"Capture kind, e.g. jvm_redefine or jvm_trace. Owned until fixture teardown."},
 "duration":{"type":"string","description":"Positive capture duration, e.g. 5s."},
 "params":{"type":"object","additionalProperties":true},
 "allowMutation":{"type":"boolean","default":false}
 },"oneOf":[{"required":["operation"],"not":{"anyOf":[{"required":["capture"]},{"required":["duration"]}]}},{"required":["capture","duration"],"not":{"required":["operation"]}}]
}`)
}

type config struct {
	Server        string                     `json:"server"`
	Token         string                     `json:"token"`
	Agent         string                     `json:"agent"`
	Operation     string                     `json:"operation"`
	Capture       string                     `json:"capture"`
	Duration      string                     `json:"duration"`
	Params        map[string]json.RawMessage `json:"params"`
	AllowMutation bool                       `json:"allowMutation"`
}

func (Kind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	cfg, err := fixtures.KindConfig[config](step)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	if strings.TrimSpace(cfg.Agent) == "" || (cfg.Operation == "") == (cfg.Capture == "") {
		return fixtures.StepOutcome{}, fmt.Errorf("telltale requires agent and exactly one of operation or capture")
	}
	client, err := newClient(step, cfg)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	if cfg.Capture != "" {
		return client.capture(ctx, step, cfg)
	}
	if cfg.Duration != "" {
		return fixtures.StepOutcome{}, fmt.Errorf("duration applies only to captures")
	}
	if cfg.Params == nil {
		cfg.Params = map[string]json.RawMessage{}
	}
	body := map[string]any{"operation": cfg.Operation, "params": cfg.Params, "allowMutation": cfg.AllowMutation}
	raw, err := client.request(ctx, "POST", "/api/agents/"+escape(cfg.Agent)+"/call", body)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("telltale %s: %w", cfg.Operation, err)
	}
	return outcome(step, raw)
}

func outcome(step *fixtures.StepContext, raw []byte) (fixtures.StepOutcome, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("decode Telltale response: %w", err)
	}
	return fixtures.StepOutcome{Result: fixtures.FixtureResult{Name: step.Spec.ID, Type: "telltale", Status: task.StatusPASS, Stdout: string(raw)}, Outputs: map[string]any{"json": value, "stdout": string(raw), "exitCode": 0}}, nil
}

func (Kind) Observe(ctx context.Context, _ *fixtures.StepContext, handle any) (map[string]any, error) {
	run, ok := handle.(*capture)
	if !ok {
		return nil, fmt.Errorf("telltale wait requires a capture step")
	}
	return run.observe(ctx)
}

func (Kind) FailureCause(outputs map[string]any) string {
	if outputs["status"] == "failed" {
		return fmt.Sprintf("Telltale capture failed: %v", outputs["json"])
	}
	return ""
}

func (c *client) capture(ctx context.Context, step *fixtures.StepContext, cfg config) (fixtures.StepOutcome, error) {
	duration, err := time.ParseDuration(cfg.Duration)
	if err != nil || duration <= 0 {
		return fixtures.StepOutcome{}, fmt.Errorf("capture duration must be positive: %q", cfg.Duration)
	}
	params := map[string]any{"enabled": true, "target": map[string]any{"seam": "agent", "agent": cfg.Agent, "allowMutation": cfg.AllowMutation}}
	for key, value := range cfg.Params {
		if key == "target" || key == "enabled" {
			return fixtures.StepOutcome{}, fmt.Errorf("capture params must not override %s", key)
		}
		params[key] = value
	}
	raw, err := c.request(ctx, "POST", "/api/traces", map[string]any{"duration": cfg.Duration, "traces": map[string]any{cfg.Capture: params}})
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	result, err := outcome(step, raw)
	if err != nil {
		return result, err
	}
	var info struct {
		RunID  string `json:"runId"`
		Status string `json:"status"`
	}
	if err = json.Unmarshal(raw, &info); err != nil {
		return fixtures.StepOutcome{}, err
	}
	if info.RunID == "" || info.Status == "" {
		return fixtures.StepOutcome{}, fmt.Errorf("telltale capture response is missing runId or status")
	}
	run := &capture{client: c, id: info.RunID}
	result.Handle = run
	result.Cleanup = run.close
	result.Outputs["runId"] = info.RunID
	result.Outputs["status"] = info.Status
	return result, nil
}
