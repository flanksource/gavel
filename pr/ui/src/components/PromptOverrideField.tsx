import { useCallback } from 'react';
import {
  PromptPickerField,
  type PromptPickerValue,
  type PromptSpecDetail,
  type PromptSpecSavePayload,
  type SpecRuntimeFamily,
} from '@flanksource/clicky-ui/ai';
import type { ChatModel } from '@flanksource/clicky-ui/chat';

// A prompt override is either inline template text or a path to a .prompt file
// (the union the Go PromptOverride marshals). An unset/empty override means the
// built-in default is used.
export type PromptOverrideValue = PromptPickerValue;

// promptResponseError turns a non-ok settings response into a readable Error.
// Gavel replies with a `{ "error": "..." }` body, so it extracts that message;
// it falls back to trimmed plain text, then to a status-based fallback. This
// keeps repair errors human-readable instead of surfacing raw JSON.
async function promptResponseError(response: Response, fallback: string): Promise<Error> {
  const text = await response.text().catch(() => '');
  const trimmed = text.trim();
  if (trimmed) {
    try {
      const parsed = JSON.parse(trimmed) as { error?: unknown };
      if (parsed && typeof parsed.error === 'string' && parsed.error.trim()) {
        return new Error(parsed.error.trim());
      }
    } catch {
      // Body was not JSON — fall through and surface the plain text as-is.
    }
    return new Error(trimmed);
  }
  return new Error(fallback);
}

interface Props {
  value: PromptOverrideValue | undefined;
  onChange: (next: PromptOverrideValue | undefined) => void;
  description?: string;
  /** Stable prompt id (schema x-prompt-id) — the /api/settings/prompts/{id} key. */
  id: string;
  /** Human label for the prompt, shown in the editor dialog title. */
  title: string;
  /** scope=global | project=<name>, scoping the detail request to one layer. */
  scopeQuery: string;
  /** Model catalog shown by the shared SpecRuntimeEditor dialog. */
  models?: ChatModel[];
  /** Backend/family catalog used to scope models to the selected runtime. */
  families?: SpecRuntimeFamily[];
}

// PromptOverrideField adapts Gavel's settings prompt endpoints to clicky-ui's
// reusable one-line PromptPickerField.
export function PromptOverrideField({ value, onChange, description, id, title, scopeQuery, models, families }: Props) {
  const loadDetail = useCallback(() => {
    return fetch(`/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`)
      .then(async (response) => {
        if (!response.ok) throw await promptResponseError(response, `load failed (${response.status})`);
        return response.json() as Promise<PromptSpecDetail>;
      });
  }, [id, scopeQuery]);

  const saveDetail = useCallback((payload: PromptSpecSavePayload) => {
    return fetch(`/api/settings/prompts/${encodeURIComponent(id)}?${scopeQuery}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    }).then(async (response) => {
      if (!response.ok) throw await promptResponseError(response, `save failed (${response.status})`);
      return response.json() as Promise<PromptSpecDetail>;
    });
  }, [id, scopeQuery]);

  return (
    <PromptPickerField
      value={value}
      onChange={onChange}
      title={title}
      description={description}
      loadDetail={loadDetail}
      saveDetail={saveDetail}
      models={models}
      families={families}
    />
  );
}
