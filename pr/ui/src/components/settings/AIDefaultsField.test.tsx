import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type {
  PromptSpecDetail,
  PromptSpecSavePayload,
} from "@flanksource/clicky-ui/ai";
import { AIDefaultsField } from "./AIDefaultsField";

vi.mock("@flanksource/clicky-ui/ai", async () => {
  return {
    PromptPickerField: (props: {
      title: string;
      sources?: string[];
      loadDetail: () => Promise<PromptSpecDetail>;
      saveDetail: (payload: PromptSpecSavePayload) => Promise<PromptSpecDetail>;
    }) => {
      return (
        <div>
          <span>{props.title}</span>
          <span data-testid="sources">{props.sources?.join(",")}</span>
          <button
            type="button"
            onClick={async () => {
              const detail = await props.loadDetail();
              const { spec, body } = detail;
              if (!spec || body === undefined) throw new Error(detail.parseError);
              await props.saveDetail({
                source: "inline",
                spec,
                body,
              });
            }}
          >
            Round-trip defaults
          </button>
          <button
            type="button"
            onClick={async () => {
              await props.saveDetail({
                source: "inline",
                spec: {
                  model: "agent:sonnet",
                  prompt: { system: "Always verify." },
                },
                body: "Default user prompt",
              });
            }}
          >
            Save defaults
          </button>
        </div>
      );
    },
  };
});

describe("AIDefaultsField", () => {
  it("uses an inline-only PromptPicker and preserves system and user prompts", async () => {
    const onChange = vi.fn();
    render(
      <AIDefaultsField
        value={{
          model: "claude-haiku-4-5",
          prompt: { system: "Be concise.", user: "Summarize." },
        }}
        onChange={onChange}
      />
    );

    expect(screen.getByTestId("sources").textContent).toBe("inline");

    await fireEvent.click(
      screen.getByRole("button", { name: "Round-trip defaults" })
    );
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith({
        model: "claude-haiku-4-5",
        prompt: { system: "Be concise.", user: "Summarize." },
      });
    });
    onChange.mockClear();

    await fireEvent.click(
      screen.getByRole("button", { name: "Save defaults" })
    );
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith({
        model: "agent:sonnet",
        prompt: { system: "Always verify.", user: "Default user prompt" },
      });
    });
  });
});
