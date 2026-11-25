package clicky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/comments"
)

// OpenAPI adds the streaming and binary routes to the generated entity document.
func OpenAPI(next http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &documentResponse{header: http.Header{}, status: 200}
		next.ServeHTTP(response, r)
		if response.status >= 400 {
			for key, value := range response.header {
				w.Header()[key] = value
			}
			w.WriteHeader(response.status)
			if _, err := w.Write(response.body.Bytes()); err != nil {
				return
			}
			return
		}
		var document map[string]any
		if err := json.Unmarshal(response.body.Bytes(), &document); err != nil {
			writeError(w, fmt.Errorf("read generated comments OpenAPI: %w", err))
			return
		}
		paths, ok := document["paths"].(map[string]any)
		if !ok {
			writeError(w, fmt.Errorf("generated document has no paths"))
			return
		}
		if err := describeCreateAndList(paths, prefix); err != nil {
			writeError(w, err)
			return
		}
		for path, item := range HTTPPaths(prefix) {
			paths[path] = item
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(document); err != nil {
			return
		}
	})
}
func HTTPPaths(prefix string) map[string]any {
	id := map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}
	parameter := func(name, kind string, required bool) map[string]any {
		return map[string]any{"name": name, "in": "query", "required": required, "schema": map[string]string{"type": kind}}
	}
	response := func(media string) map[string]any {
		return map[string]any{"200": map[string]any{"description": "Authorized comment response", "content": map[string]any{media: map[string]any{"schema": map[string]string{"type": "string"}}}}}
	}
	paths := map[string]any{
		prefix + "/comments/batch":                         map[string]any{"post": map[string]any{"operationId": "comments-batch", "summary": "Atomically apply versioned comment actions", "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "minItems": 1, "maxItems": 1000, "items": map[string]any{"type": "object", "required": []string{"id", "expectedVersion", "action"}, "properties": map[string]any{"id": map[string]string{"type": "string"}, "expectedVersion": map[string]string{"type": "integer", "format": "int64"}, "action": map[string]string{"type": "string"}, "bodyMarkdown": map[string]string{"type": "string"}, "messageId": map[string]string{"type": "string"}, "parentId": map[string]string{"type": "string"}, "content": map[string]any{}, "target": map[string]string{"type": "object"}, "assignee": map[string]string{"type": "object"}, "summary": map[string]string{"type": "string"}, "evidence": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}, "patch": map[string]string{"type": "string", "description": "Unified diff a reply proposes; exclusive with edits"}, "edits": comments.FileEditsSchema(), "mode": map[string]any{"type": "string", "enum": []string{"apply", "stage"}}}}}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated comment aggregates in action order"}, "409": map[string]any{"description": "Stale version; no changes committed"}}}},
		prefix + "/comments/events":                        map[string]any{"get": map[string]any{"operationId": "comments-events", "summary": "Follow commit-ordered comment events", "parameters": []any{parameter("scope", "string", true), parameter("after", "integer", false), map[string]any{"name": "Last-Event-ID", "in": "header", "schema": map[string]string{"type": "integer"}}}, "responses": response("text/event-stream")}},
		prefix + "/comments/{id}/attachments":              map[string]any{"post": map[string]any{"operationId": "comments-upload", "summary": "Attach bounded binary evidence", "parameters": []any{id, parameter("name", "string", true), parameter("expectedVersion", "integer", true), parameter("messageId", "string", false)}, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/octet-stream": map[string]any{"schema": map[string]string{"type": "string", "format": "binary"}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated comment and attachment metadata"}, "409": map[string]any{"description": "Thread is final, or expectedVersion is newer than the thread; changes since expectedVersion (an agent delivery claiming the thread) do not conflict"}}}},
		prefix + "/comments/{id}/attachments/{attachment}": map[string]any{"get": map[string]any{"operationId": "comments-download", "summary": "Download authorized attachment bytes", "parameters": []any{id, map[string]any{"name": "attachment", "in": "path", "required": true, "schema": map[string]string{"type": "string"}}}, "responses": response("application/octet-stream")}},
	}
	for path, item := range deliveryPaths(id, parameter) {
		paths[prefix+path] = item
	}
	for path, item := range readPaths() {
		paths[prefix+path] = item
	}
	paths[prefix+"/comments/{id}/proposals/{replyId}/diff"] = proposalDiffPath(id)
	return paths
}

// describeCreateAndList adds to the generated /comments path what its flags
// and reflection cannot say: a create's attachments (decoded from the body,
// not a flag) and the anchor states a listed thread carries.
func describeCreateAndList(paths map[string]any, prefix string) error {
	item, _ := paths[prefix+"/comments"].(map[string]any)
	create, err := schemaAt(item, "post", "requestBody", "content", "application/json", "schema", "properties")
	if err != nil {
		return fmt.Errorf("generated create schema: %w", err)
	}
	create["delivery"] = draftDeliverySchema()
	create["attachments"] = map[string]any{"type": "array", "maxItems": comments.MaxCreateAttachments, "description": "Files stored on the new thread, in the create's transaction, before it is routed or delivered. Each is checked (plain file name, media type, size up to the host's attachment limit) before anything is written.", "items": map[string]any{
		"type": "object", "required": []string{"filename", "contentType", "data"}, "properties": map[string]any{
			"filename": map[string]string{"type": "string"}, "contentType": map[string]string{"type": "string"}, "data": map[string]string{"type": "string", "format": "byte", "description": "base64"},
		},
	}}
	anchor, err := schemaAt(item, "get", "responses", "200", "content", "application/json", "schema", "properties", "data", "items", "properties", "anchor")
	if err != nil {
		return fmt.Errorf("generated list schema: %w", err)
	}
	anchor["description"] = "Where the thread's patch anchor lies in its target now; only threads whose selector carries a patch, on targets the host has file content for"
	state, err := schemaAt(anchor, "properties", "state")
	if err != nil {
		return fmt.Errorf("generated list anchor schema: %w", err)
	}
	state["enum"] = []string{"matched", "ambiguous", "outdated"}
	return nil
}

// schemaAt is the object at keys below node.
func schemaAt(node map[string]any, keys ...string) (map[string]any, error) {
	for i, key := range keys {
		next, ok := node[key].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("no object at %v", keys[:i+1])
		}
		node = next
	}
	return node, nil
}

// proposalDiffPath describes the patch a reply proposes: comments.ProposalDiff.
func proposalDiffPath(id map[string]any) map[string]any {
	str := map[string]string{"type": "string"}
	replyID := map[string]any{"name": "replyId", "in": "path", "required": true, "schema": str}
	diff := map[string]any{"type": "object", "required": []string{"diff", "commit"}, "properties": map[string]any{
		"diff": str, "commit": str, "truncated": map[string]string{"type": "boolean"}, "binary": map[string]string{"type": "boolean"},
	}}
	return map[string]any{"get": map[string]any{"operationId": "comments-proposal-diff", "summary": "Read the unified diff a reply proposes against its base", "parameters": []any{id, replyID}, "responses": map[string]any{
		"200": map[string]any{"description": "The proposal's diff", "content": map[string]any{"application/json": map[string]any{"schema": diff}}},
		"400": map[string]any{"description": "The reply has no proposal, or the host has no repository for proposals"},
		"404": map[string]any{"description": "No such thread or reply"},
	}}}
}

func draftDeliverySchema() map[string]any {
	return map[string]any{"type": "object", "description": "Draft delivery choices, exclusive with deliver", "properties": map[string]any{
		"defaultSessionId": map[string]string{"type": "string", "description": "Use this session when the draft does not address an agent"},
		"mentionId":        map[string]string{"type": "string", "description": "A selected ID from the target's mentionables"},
	}}
}

// deliveryPaths are the agent delivery operations RegisterDeliveryRoutes mounts.
func deliveryPaths(id map[string]any, parameter func(name, kind string, required bool) map[string]any) map[string]any {
	str := map[string]string{"type": "string"}
	target := map[string]any{"type": "object", "required": []string{"kind"}, "description": "Deliver to a Captain session, an Amp Orb thread, or a managed run.", "properties": map[string]any{
		"kind":       map[string]any{"type": "string", "enum": []string{"session", "run"}},
		"sessionId":  map[string]string{"type": "string", "description": "Captain session ID or Amp Orb thread ID or URL"},
		"prompt":     str,
		"runOptions": map[string]string{"type": "object"},
		"groupKey":   str,
		"fresh":      map[string]string{"type": "boolean", "description": "Run targets only. Start a new session and make it the group's current session."},
		"edits":      map[string]any{"type": "string", "enum": []string{"direct", "propose"}, "description": "The edit policy. Inherits from the target's comment agent when omitted."},
	}}
	body := func(schema map[string]any) map[string]any {
		return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
	}
	ok := func(description string) map[string]any {
		return map[string]any{"200": map[string]any{"description": description}, "501": map[string]any{"description": "Agent delivery is not enabled on this host"}}
	}
	deliverResponses := ok("Delivered; the comment with its agentRun")
	deliverResponses["409"] = map[string]any{"description": "Stale expectedVersion, or a delivery is already active"}
	deliverResponses["502"] = map[string]any{"description": "Delivery ran and failed; the failure is recorded in the comment's agentRun"}
	agentResponses := map[string]any{
		"200": map[string]any{"description": "{settings, agent, session?}: the target's settings layers and positions, its resolved comment agent with each field traced to the layer that set it, and session, the live session (gavel edit --session) every new human thread goes to ahead of the agent, when the host sets one", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"settings", "agent"}, "properties": map[string]any{
			"settings": map[string]string{"type": "object"}, "agent": map[string]string{"type": "object"},
			"session": map[string]string{"type": "string", "description": "The live session new human threads are routed to ahead of the agent"},
		}}}}},
		"400": map[string]any{"description": "The target is not one the host's settings cover"},
		"501": map[string]any{"description": "Comment agent settings are not enabled on this host"},
	}
	saveResponses := map[string]any{"403": map[string]any{"description": "Only a person configures comment agents"}}
	for status, response := range agentResponses {
		saveResponses[status] = response
	}
	saveBody := map[string]any{"type": "object", "required": []string{"target", "layer", "agent"}, "properties": map[string]any{
		"target": map[string]any{"type": "object", "required": []string{"scopeUri", "uri"}, "properties": map[string]any{"scopeUri": str, "uri": str}},
		"layer":  map[string]any{"type": "object", "required": []string{"kind"}, "description": "The settings layer to save at: {kind: user|parent|project|path, path?, file?} (no file saves the database layer)", "properties": map[string]any{"kind": str, "path": str, "file": str}},
		"page":   map[string]any{"type": "boolean", "description": "Save the layer's entry for the target instead of its default"},
		"agent":  map[string]any{"type": []string{"object", "null"}, "description": "A JSON merge patch (RFC 7386) of the comments.agents entry, applied under the layer's save lock to the entry it already holds: a present field replaces that field (objects merge), a field set to null is removed so it inherits again; agent null deletes the whole entry"},
	}}
	saveResponses["400"] = map[string]any{"description": "The target is not one the host's settings cover, or the patch names an unknown field, a wrong type or an invalid edits policy"}
	batchBody := map[string]any{"type": "object", "required": []string{"requestId", "target", "threads"}, "properties": map[string]any{
		"requestId": str, "target": target,
		"threads": map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"type": "object", "required": []string{"id", "expectedVersion"}, "properties": map[string]any{"id": str, "expectedVersion": map[string]string{"type": "integer", "format": "int64"}}}},
	}}
	batchResponses := ok("Delivered; the threads' comments in request order, each with the same agentRun")
	batchResponses["400"] = map[string]any{"description": "No threads, a thread twice, a closed thread, or threads of different scopes; nothing is delivered"}
	batchResponses["409"] = map[string]any{"description": "A thread's expectedVersion is stale, or a delivery is already active; nothing is delivered"}
	batchResponses["502"] = deliverResponses["502"]
	return map[string]any{
		"/comments/agent": map[string]any{
			"get": map[string]any{"operationId": "comments-agent", "summary": "Read the comment agent a target's settings configure", "parameters": []any{parameter("scope", "string", true), parameter("uri", "string", true)}, "responses": agentResponses},
			"put": map[string]any{"operationId": "comments-save-agent", "summary": "Patch or remove the comment agent at one settings layer of a target", "requestBody": body(saveBody), "responses": saveResponses},
		},
		"/comments/agent-sessions": map[string]any{
			"get":  map[string]any{"operationId": "comments-agent-sessions", "summary": "List deliverable sessions; group narrows to the current session of that group and requires scope", "parameters": []any{parameter("scope", "string", false), parameter("group", "string", false), parameter("all", "string", false)}, "responses": sessionsResponses(ok)},
			"post": startSessionOperation(body, ok),
		},
		"/comments/mentionables": map[string]any{"get": map[string]any{"operationId": "comments-mentionables", "summary": "List the agents a comment on a target addresses with an @name token: @agent (its comment agent rule), @<runtime family> (a fresh run of its default agent model) and @<session slug>", "parameters": []any{parameter("scope", "string", true), parameter("target", "string", true)}, "responses": mentionablesResponses(ok)}},
		"/comments/deliver":      map[string]any{"post": map[string]any{"operationId": "comments-deliver-batch", "summary": "Deliver several threads of one scope to one target in a single message", "requestBody": body(batchBody), "responses": batchResponses}},
		"/comments/delivery/preview": map[string]any{"post": map[string]any{
			"operationId": "comments-create-delivery-preview", "summary": "Resolve delivery for a new human-review comment without posting it",
			"requestBody": body(map[string]any{"type": "object", "required": []string{"target"}, "properties": map[string]any{"target": map[string]string{"type": "object"}, "bodyMarkdown": str, "deliver": target, "delivery": draftDeliverySchema()}}),
			"responses":   ok("Effective route and existing session, when present. Mention marks a draft-specific override. No route means post only."),
		}},
		"/comments/{id}/deliver":         map[string]any{"post": map[string]any{"operationId": "comments-deliver", "summary": "Deliver a thread to a live agent session or a new managed run", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"expectedVersion", "requestId", "target"}, "properties": map[string]any{"expectedVersion": map[string]string{"type": "integer", "format": "int64"}, "requestId": str, "target": target}}), "responses": deliverResponses}},
		"/comments/{id}/deliver/preview": map[string]any{"post": map[string]any{"operationId": "comments-deliver-preview", "summary": "Render what a run target would start for a thread", "parameters": []any{id}, "requestBody": body(map[string]any{"type": "object", "required": []string{"target"}, "properties": map[string]any{"target": target}}), "responses": ok("Rendered run preview")}},
	}
}

// mentionablesResponses describe GET /comments/mentionables: comments.Mentionable items.
func mentionablesResponses(ok func(string) map[string]any) map[string]any {
	mentionable := map[string]any{"type": "object", "required": []string{"id", "name", "kind"}, "properties": map[string]any{
		"id":   map[string]string{"type": "string", "description": `"agent", "runtime:<family>" or "session:<id>"`},
		"name": map[string]string{"type": "string", "description": "The @name: agent, a runtime family, or a session's captain slug (a managed session without one: its label as a slug; else its id). A session whose name agent or a runtime holds is not listed; sessions sharing a name are listed as <name>-<first 8 of id>"},
		"kind": map[string]any{"type": "string", "enum": []string{"agent"}},
	}}
	responses := ok("Mentionable agents")
	responses["200"] = map[string]any{"description": "Mentionable agents; a scope without a local working directory lists no sessions", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": mentionable}}}}
	responses["400"] = map[string]any{"description": "No scope or target"}
	return responses
}

// startSessionOperation describes POST /comments/agent-sessions: comments.StartSessionRequest.
func startSessionOperation(body func(map[string]any) map[string]any, ok func(string) map[string]any) map[string]any {
	str := map[string]string{"type": "string"}
	request := map[string]any{"type": "object", "required": []string{"scope", "target", "prefix"}, "properties": map[string]any{
		"scope":      str,
		"target":     map[string]string{"type": "string", "description": "The page target URI the session is for, e.g. <scope>/pages/<page key>"},
		"prefix":     map[string]string{"type": "string", "description": "The session name's prefix; the session is titled <prefix>-<6 random lowercase letters or digits>"},
		"prompt":     map[string]string{"type": "string", "description": "The prompt the session runs; the host's comment agent prompt (editor-comments) when omitted"},
		"runOptions": map[string]any{"type": "object", "description": "{spec?, presets?, variables?, instructions?}"},
		"groupKey":   map[string]string{"type": "string", "description": "The delivery group the session becomes the current session of; the target when omitted"},
		"edits":      map[string]any{"type": "string", "enum": []string{"direct", "propose"}, "description": "The edit policy. Inherits from the target's comment agent when omitted."},
	}}
	responses := ok("")
	delete(responses, "200")
	responses["201"] = map[string]any{"description": "The started session, as GET /comments/agent-sessions lists it", "content": map[string]any{"application/json": map[string]any{"schema": agentSessionSchema()}}}
	responses["400"] = map[string]any{"description": "No scope or target, a blank prefix, or an invalid edits policy or runOptions"}
	responses["502"] = map[string]any{"description": "The run ended, or the request or 60s cap ran out, before the session reported its id"}
	return map[string]any{"operationId": "comments-start-agent-session", "summary": "Start a named managed session for a page ahead of its comment threads", "requestBody": body(request), "responses": responses}
}

// sessionsResponses describe GET /comments/agent-sessions: comments.AgentSession items.
func sessionsResponses(ok func(string) map[string]any) map[string]any {
	responses := ok("Agent sessions")
	responses["400"] = map[string]any{"description": "A group filter requires scope"}
	responses["200"] = map[string]any{"description": "Agent sessions", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": agentSessionSchema()}}}}
	return responses
}

// agentSessionSchema is a comments.AgentSession.
func agentSessionSchema() map[string]any {
	str, flag := map[string]string{"type": "string"}, func(description string) map[string]string {
		return map[string]string{"type": "boolean", "description": description}
	}
	return map[string]any{"type": "object", "required": []string{"id", "source", "title", "inCmux"}, "properties": map[string]any{
		"id": str, "source": str, "title": str, "model": str, "cwd": str, "inCmux": map[string]string{"type": "boolean"},
		"status":    map[string]string{"type": "string", "description": `The live session's status, or "ended" for a managed session that is no longer live`},
		"managed":   flag("A session the host launched for the scope's threads"),
		"resumable": flag("An ended managed session: delivering to it resumes it with its recorded prompt and run options"),
		"context":   map[string]any{"type": "object", "properties": map[string]any{"usedTokens": map[string]string{"type": "integer"}, "windowTokens": map[string]string{"type": "integer"}}},
	}}
}

type documentResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *documentResponse) Header() http.Header            { return r.header }
func (r *documentResponse) WriteHeader(status int)         { r.status = status }
func (r *documentResponse) Write(data []byte) (int, error) { return r.body.Write(data) }
