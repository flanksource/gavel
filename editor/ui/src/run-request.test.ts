import { describe, expect, it } from "vitest";
import type { SpecRuntimeFamily } from "@flanksource/clicky-ui/ai";
import {
  assertRunContext,
  availableMode,
  buildRunRequest,
  effectiveRuntimeOf,
  findPrompt,
  modeUnavailable,
  toolMetaOf,
  variablesSchemaOf,
  type PromptSummary,
} from "./run-request";

const HOST = ["dir", "scope", "target", "actor", "gavelExecutable", "message"];

const prompt = (overrides: Partial<PromptSummary>): PromptSummary => ({ id: "p", name: "p", source: "library", ...overrides });

describe("buildRunRequest", () => {
  it("sends the prompt, ordered presets, compacted spec, user variables and trimmed instructions", () => {
    expect(
      buildRunRequest({
        prompt: "library:editor-comments",
        value: { presets: ["fast", "cheap"], spec: { model: "claude-opus-4-6" }, variables: { tone: "terse" } },
        instructions: "  Keep the heading \n",
        hostVariables: HOST,
      }),
    ).toEqual({
      prompt: "library:editor-comments",
      presets: ["fast", "cheap"],
      spec: { model: "claude-opus-4-6" },
      variables: { tone: "terse" },
      instructions: "Keep the heading",
    });
  });

  it("strips host-owned variables and drops blank variables, empty presets, unset spec fields and blank instructions", () => {
    expect(
      buildRunRequest({
        prompt: "",
        value: {
          presets: [],
          spec: { model: "", mode: "" },
          variables: { dir: "/elsewhere", message: "hi", actor: "me", tone: "", count: undefined, flag: false, depth: 0 },
        },
        instructions: "   ",
        hostVariables: HOST,
      }),
    ).toEqual({ variables: { flag: false, depth: 0 } });
  });

  it("is empty when nothing was chosen", () => {
    expect(buildRunRequest({ prompt: "", value: {}, instructions: "", hostVariables: HOST })).toEqual({});
  });
});

describe("variablesSchemaOf", () => {
  it("lists the prompt's declared variables minus host variables, defaulting to string and marking required ones", () => {
    expect(
      variablesSchemaOf(
        [
          { name: "dir", type: "string", required: true },
          { name: "tone", description: "How to write", required: true },
          { name: "retries", type: "int" },
          { name: "dryRun", type: "bool" },
          { name: "message" },
        ],
        HOST,
      ),
    ).toEqual({
      type: "object",
      properties: {
        tone: { type: "string", title: "tone", description: "How to write" },
        retries: { type: "integer", title: "retries" },
        dryRun: { type: "boolean", title: "dryRun" },
      },
      required: ["tone"],
    });
  });

  it("is an empty object schema when the prompt declares nothing the user can set", () => {
    expect([variablesSchemaOf(undefined, HOST), variablesSchemaOf([{ name: "target" }], HOST)]).toEqual([
      { type: "object", properties: {} },
      { type: "object", properties: {} },
    ]);
  });

  it("rejects a variable type it cannot render", () => {
    expect(() => variablesSchemaOf([{ name: "when", type: "datetime" }], HOST)).toThrow('prompt variable when has unsupported type "datetime"');
  });
});

describe("effectiveRuntimeOf", () => {
  const claude: SpecRuntimeFamily = {
    id: "claude",
    label: "Claude",
    provider: "anthropic",
    modes: [{ id: "agent", label: "Agent" }, { id: "api", label: "API" }],
  };
  const codex: SpecRuntimeFamily = { id: "codex", label: "Codex", provider: "openai", modes: [{ id: "agent", label: "Agent" }] };
  const models = [{ id: "gpt-5", provider: "openai", label: "GPT-5" }] as never;

  it("prefers the request's model and mode, then the prompt's, then claude agent", () => {
    const families = [claude, codex];
    expect([
      effectiveRuntimeOf({ spec: { model: "gpt-5", mode: "agent" }, prompt: prompt({ model: "claude", mode: "api" }), families, models }),
      effectiveRuntimeOf({ spec: {}, prompt: prompt({ model: "codex", mode: "agent" }), families, models }),
      effectiveRuntimeOf({ spec: {}, prompt: undefined, families, models }),
    ]).toEqual([
      { model: "gpt-5", mode: "agent", family: "codex" },
      { model: "codex", mode: "agent", family: "codex" },
      { model: "claude", mode: "agent", family: "claude" },
    ]);
  });

  it("falls back to the first family providing the mode when the model names no family, and to none when no family does", () => {
    expect([
      effectiveRuntimeOf({ spec: { model: "unknown-model", mode: "api" }, prompt: undefined, families: [claude, codex], models }).family,
      effectiveRuntimeOf({ spec: { model: "unknown-model", mode: "cmux" }, prompt: undefined, families: [claude, codex], models }).family,
    ]).toEqual(["claude", undefined]);
  });
});

describe("modeUnavailable", () => {
  const claude = {
    label: "Claude",
    modes: [
      { id: "agent", availability: { state: "missing_executable", reason: "claude is not installed.", remediation: "Install Claude Code." } },
      { id: "api", availability: { state: "available" } },
    ],
  };
  const codex = { label: "Codex", modes: [{ id: "agent", availability: { state: "not_authenticated", reason: "Log in to codex." } }] };

  it("explains every family's reason when no family can run the mode, and offers one that can", () => {
    expect([modeUnavailable([claude, codex], "agent"), availableMode([claude, codex])]).toEqual([
      "Claude: claude is not installed. Install Claude Code. Codex: Log in to codex.",
      "api",
    ]);
  });

  it("accepts a mode one family can run, and reports a mode no family provides", () => {
    expect([modeUnavailable([claude, codex], "api"), modeUnavailable([codex], "cmux"), availableMode([codex])]).toEqual([
      undefined,
      "No runtime here provides cmux mode.",
      undefined,
    ]);
  });
});

describe("findPrompt", () => {
  const library = prompt({ id: "library:editor-comments", name: "editor-comments", source: "library" });
  const project = prompt({ id: "project:editor-comments", name: "editor-comments", source: "project" });

  it("matches an id before a name, so a same-named prompt from another source cannot shadow it", () => {
    expect([findPrompt([library, project], "project:editor-comments"), findPrompt([library, project], "editor-comments"), findPrompt([library], "missing")]).toEqual([
      project,
      library,
      undefined,
    ]);
  });
});

describe("toolMetaOf", () => {
  it("groups preview tools by where they came from", () => {
    expect(
      toolMetaOf([
        { name: "reply_to_thread", description: "Reply", source: "thread" },
        { name: "Read", source: "prompt" },
      ]),
    ).toEqual([
      { name: "reply_to_thread", label: "reply_to_thread", group: "Thread tools", description: "Reply" },
      { name: "Read", label: "Read", group: "Prompt tools" },
    ]);
  });
});

describe("assertRunContext", () => {
  const valid = { prompts: [prompt({ id: "library:editor-comments", name: "editor-comments" })], presets: [], defaultPrompt: "editor-comments", hostVariables: HOST, dir: "/repo" };

  it("returns a well-formed context and rejects one missing a field", () => {
    expect(assertRunContext(valid)).toBe(valid);
    expect(() => assertRunContext({ ...valid, presets: undefined })).toThrow("agent-run context is malformed");
  });

  it("rejects a context whose default prompt was not discovered", () => {
    expect(() => assertRunContext({ ...valid, prompts: [] })).toThrow('default prompt "editor-comments" is not among the 0 discovered prompts');
  });
});
