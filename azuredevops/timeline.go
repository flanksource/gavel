package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/pr/model"
)

type timelineOptions struct {
	model.StatusOptions
	Project string
}

func (c *Client) loadTimeline(ctx context.Context, run *model.WorkflowRun, opts timelineOptions) error {
	path := c.projectPath(opts.Project, fmt.Sprintf("build/builds/%d", run.DatabaseID))
	data, _, err := c.request(ctx, http.MethodGet, path+"/timeline", nil, nil)
	if err != nil {
		return err
	}
	var timeline struct {
		Records *[]timelineRecord `json:"records"`
	}
	if err = json.Unmarshal(data, &timeline); err != nil {
		return fmt.Errorf("decode Azure timeline: %w", err)
	}
	if timeline.Records == nil {
		return fmt.Errorf("azure build %d timeline is missing records", run.DatabaseID)
	}
	records := latestRecords(*timeline.Records)
	byID := map[string]timelineRecord{}
	for _, record := range records {
		byID[record.ID] = record
	}
	for _, record := range records {
		if record.Type != "Job" {
			continue
		}
		status, conclusion, err := executionState(record.State, record.Result)
		if err != nil {
			return fmt.Errorf("job %s: %w", record.ID, err)
		}
		job := model.Job{NativeID: record.ID, Name: record.Name, Status: status, Conclusion: conclusion, StartedAt: record.Started, CompletedAt: record.Finished, URL: run.URL + "&view=logs&j=" + url.QueryEscape(record.ID)}
		for _, task := range records {
			if task.Type != "Task" || !belongsToJob(task, record.ID, byID) {
				continue
			}
			status, conclusion, err := executionState(task.State, task.Result)
			if err != nil {
				return fmt.Errorf("task %s: %w", task.ID, err)
			}
			step := model.Step{Name: task.Name, Number: task.Order, Status: status, Conclusion: conclusion}
			if opts.Logs && model.IsFailureConclusion(conclusion) && task.Log != nil {
				step.Logs, err = c.logTail(ctx, path, task.Log.ID, opts.TailLogs)
				if err != nil {
					return err
				}
			}
			job.Steps = append(job.Steps, step)
		}
		if opts.Logs && model.IsFailureConclusion(job.Conclusion) && len(job.Steps) == 0 && record.Log != nil {
			job.Logs, err = c.logTail(ctx, path, record.Log.ID, opts.TailLogs)
			if err != nil {
				return err
			}
		}
		run.Jobs = append(run.Jobs, job)
	}
	return nil
}

func latestRecords(records []timelineRecord) []timelineRecord {
	latest := map[string]timelineRecord{}
	for _, record := range records {
		key := record.ID
		if record.Identifier != "" {
			key = record.Type + ":" + record.ParentID + ":" + record.Identifier
		}
		if old, ok := latest[key]; !ok || record.Attempt >= old.Attempt {
			latest[key] = record
		}
	}
	out := make([]timelineRecord, 0, len(latest))
	for _, r := range latest {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func belongsToJob(task timelineRecord, jobID string, records map[string]timelineRecord) bool {
	seen := map[string]bool{}
	for id := task.ParentID; id != "" && !seen[id]; id = records[id].ParentID {
		if id == jobID {
			return true
		}
		seen[id] = true
	}
	return false
}

func (c *Client) logTail(ctx context.Context, path string, id int64, tail int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("azure timeline has invalid log ID")
	}
	logs, err := collection[logReference](ctx, c, path+"/logs", nil)
	if err != nil {
		return "", err
	}
	var found *logReference
	for i := range logs {
		if logs[i].ID == id {
			found = &logs[i]
			break
		}
	}
	if found == nil {
		return "", fmt.Errorf("azure build log %d was not found", id)
	}
	start := max(int64(0), found.Lines-int64(tail))
	if tail == 0 {
		start = 0
	}
	data, _, err := c.request(ctx, http.MethodGet, path+"/logs/"+strconv.FormatInt(id, 10), url.Values{"startLine": {strconv.FormatInt(start, 10)}, "endLine": {strconv.FormatInt(found.Lines, 10)}}, nil)
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"), "\n")
	if tail > 0 && len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return strings.Join(lines, "\n"), nil
}
