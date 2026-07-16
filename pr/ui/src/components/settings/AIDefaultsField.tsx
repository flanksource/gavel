import { useCallback } from "react";
import {
  PromptPickerField,
  type PromptPickerValue,
  type PromptSpecDetail,
  type PromptSpecSavePayload,
  type SpecRuntimeFamily,
} from "@flanksource/clicky-ui/ai";
import type { ChatModel } from "@flanksource/clicky-ui/chat";

export type AIDefaultsValue = Record<string, unknown>;

interface Props {
  value?: AIDefaultsValue;
  onChange: (next: AIDefaultsValue | undefined) => void;
  models?: ChatModel[];
  families?: SpecRuntimeFamily[];
}

const INLINE_SOURCE = ["inline"] as const;

function promptDetail(value: AIDefaultsValue | undefined): PromptSpecDetail {
  const spec = structuredClone(value ?? {});
  const prompt =
    spec.prompt && typeof spec.prompt === "object"
      ? { ...(spec.prompt as Record<string, unknown>) }
      : {};
  const body = typeof prompt.user === "string" ? prompt.user : "";
  delete prompt.user;
  if (Object.keys(prompt).length > 0) spec.prompt = prompt;
  else delete spec.prompt;
  return {
    id: "ai",
    source: value && Object.keys(value).length > 0 ? "inline" : "default",
    spec,
    body,
    raw: JSON.stringify(value ?? {}),
  };
}

function valueFromPayload(
  payload: Extract<PromptSpecSavePayload, { spec: unknown }>
): AIDefaultsValue {
  const spec = structuredClone(payload.spec) as AIDefaultsValue;
  const prompt =
    spec.prompt && typeof spec.prompt === "object"
      ? { ...(spec.prompt as Record<string, unknown>) }
      : {};
  if (payload.body) prompt.user = payload.body;
  if (Object.keys(prompt).length > 0) spec.prompt = prompt;
  else delete spec.prompt;
  return spec;
}

export function AIDefaultsField({ value, onChange, models, families }: Props) {
  const loadDetail = useCallback(async () => promptDetail(value), [value]);
  const saveDetail = useCallback(
    async (payload: PromptSpecSavePayload) => {
      if (payload.source === "default") {
        onChange(undefined);
        return promptDetail(undefined);
      }
      if (payload.source !== "inline" || !("spec" in payload)) {
        throw new Error(
          "AI defaults support structured inline configuration only"
        );
      }
      const next = valueFromPayload(payload);
      onChange(next);
      return promptDetail(next);
    },
    [onChange]
  );
  const syncPickerValue = useCallback(
    (next: PromptPickerValue | undefined) => {
      if (next === undefined) onChange(undefined);
    },
    [onChange]
  );

  return (
    <PromptPickerField
      value={value ? { inline: JSON.stringify(value) } : undefined}
      onChange={syncPickerValue}
      title="AI defaults"
      description="Defaults inherited by every AI operation"
      loadDetail={loadDetail}
      saveDetail={saveDetail}
      models={models}
      families={families}
      sources={INLINE_SOURCE}
    />
  );
}
