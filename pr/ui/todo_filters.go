package ui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func todoFiltersFromRequest(r *http.Request) (todos.DiscoveryFilters, error) {
	status := types.Status(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		return todos.DiscoveryFilters{}, nil
	}
	// Filtering accepts every known status, including the run projections a
	// caller may not write.
	if !types.IsKnownStatus(status) {
		return todos.DiscoveryFilters{}, fmt.Errorf("invalid status %q", status)
	}
	return todos.DiscoveryFilters{IncludeStatuses: []types.Status{status}}, nil
}

func parseTodoNewPayload(r *http.Request) (todoNewPayload, []todoAttachmentSummary, error) {
	var payload todoNewPayload
	var attachments []todoAttachmentSummary
	contentType := strings.ToLower(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return payload, nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		if r.MultipartForm != nil {
			if err := applyTodoNewValues(&payload, r.MultipartForm.Value, true); err != nil {
				return payload, nil, err
			}
			stored, err := persistMultipartAttachments(r.MultipartForm)
			if err != nil {
				return payload, nil, err
			}
			attachments = stored
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return payload, nil, fmt.Errorf("invalid form: %w", err)
		}
		if err := applyTodoNewValues(&payload, r.PostForm, true); err != nil {
			return payload, nil, err
		}
	case strings.HasPrefix(contentType, "application/json"):
		if r.ContentLength != 0 {
			if err := decodeTodoRequest(r, &payload); err != nil {
				return payload, nil, err
			}
		}
	case contentType == "":
		// Query-only create requests are valid.
	default:
		return payload, nil, fmt.Errorf("unsupported content type %q", r.Header.Get("Content-Type"))
	}

	if err := applyTodoNewValues(&payload, r.URL.Query(), false); err != nil {
		return payload, nil, err
	}
	return payload, attachments, nil
}

func parseTodoUpdatePayload(r *http.Request) (todoUpdatePayload, []todoAttachmentSummary, error) {
	var payload todoUpdatePayload
	var attachments []todoAttachmentSummary
	contentType := strings.ToLower(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return payload, nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		if r.MultipartForm != nil {
			applyTodoUpdateValues(&payload, r.MultipartForm.Value)
			stored, err := persistMultipartAttachments(r.MultipartForm)
			if err != nil {
				return payload, nil, err
			}
			attachments = stored
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return payload, nil, fmt.Errorf("invalid form: %w", err)
		}
		applyTodoUpdateValues(&payload, r.PostForm)
	case strings.HasPrefix(contentType, "application/json"), contentType == "":
		// A body with no explicit Content-Type is treated as JSON — many clients
		// (and httptest.NewRequest) omit the header. A bodyless request decodes
		// nothing and falls through to the later "operation is required"
		// validation, so query-only PATCHes still work.
		if r.ContentLength != 0 {
			if err := decodeTodoRequest(r, &payload); err != nil {
				return payload, nil, err
			}
		}
	default:
		return payload, nil, fmt.Errorf("unsupported content type %q", r.Header.Get("Content-Type"))
	}
	return payload, attachments, nil
}

func applyTodoUpdateValues(payload *todoUpdatePayload, values map[string][]string) {
	assignString := func(target *string, keys ...string) {
		if value, ok := firstTodoUpdateValue(values, keys...); ok {
			*target = strings.TrimSpace(value)
		}
	}
	assignPointer := func(target **string, trim bool, keys ...string) {
		if value, ok := firstTodoUpdateValue(values, keys...); ok {
			if trim {
				value = strings.TrimSpace(value)
			}
			*target = &value
		}
	}

	assignString(&payload.Dir, "dir")
	assignString(&payload.Ref, "ref")
	if value, ok := firstTodoUpdateValue(values, "status"); ok {
		payload.Status = types.Status(strings.TrimSpace(value))
	}
	if value, ok := firstTodoUpdateValue(values, "priority", "severity"); ok {
		payload.Priority = types.Priority(strings.TrimSpace(value))
	}
	assignPointer(&payload.Title, true, "title", "name")
	assignPointer(&payload.Body, false, "body", "description", "text")
	assignString(&payload.Comment, "comment")
}

func firstTodoUpdateValue(values map[string][]string, keys ...string) (string, bool) {
	for _, key := range keys {
		if vals, ok := values[key]; ok {
			if len(vals) == 0 {
				return "", true
			}
			return vals[0], true
		}
	}
	return "", false
}

func applyTodoNewValues(payload *todoNewPayload, values map[string][]string, overwrite bool) error {
	assignString := func(target *string, keys ...string) {
		if !overwrite && strings.TrimSpace(*target) != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = value
		}
	}
	assignPriority := func(target *types.Priority, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Priority(value)
		}
	}
	assignStatus := func(target *types.Status, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Status(value)
		}
	}

	assignString(&payload.Dir, "dir")
	assignString(&payload.Title, "title", "name")
	assignString(&payload.Body, "body", "description", "text")
	assignPriority(&payload.Priority, "priority", "severity")
	assignStatus(&payload.Status, "status")
	if !overwrite && payload.AutoSave != nil {
		return nil
	}
	if raw := firstTodoNewValue(values, "autoSave", "autosave", "auto_save"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid autoSave %q", raw)
		}
		payload.AutoSave = &parsed
	}
	return nil
}

func firstTodoNewValue(values map[string][]string, keys ...string) string {
	for _, key := range keys {
		for _, value := range values[key] {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}
