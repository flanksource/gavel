import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  agentLabel,
  deliverTargetOf,
  effectiveFieldsOf,
  existingAgentAt,
  fileChoicesOf,
  layerLabelOf,
  layerRefOf,
  loadViewerChoice,
  pageKeyOf,
  positionName,
  positionsNearestFirst,
  runOf,
  saveViewerChoice,
  savedAgentOf,
  snippetFileName,
  snippetNameProblem,
  DB_CHOICE,
  NEW_SNIPPET_CHOICE,
  type AgentField,
  type CommentAgent,
  type FileSettings,
  type SettingsLayer,
  type SettingsPosition,
} from "./page-agent";

const scope = "gavel://project/acme";
const target = `${scope}/files/slides/arb/16-security.mdx`;
const home = "/home/moshe";
const root = "/repo";

const positions: SettingsPosition[] = [
  { kind: "user", dir: home, files: [".gavel.yaml"], db: false },
  { kind: "parent", path: "/", dir: "/", files: [], db: false },
  { kind: "project", dir: root, files: [".gavel.yaml", ".gavel.agents.yaml"], db: false },
  { kind: "path", path: "slides", dir: `${root}/slides`, files: [], db: true },
  { kind: "path", path: "slides/arb", dir: `${root}/slides/arb`, files: [".gavel.yaml"], db: false },
];

const projectLayer: SettingsLayer = {
  kind: "project",
  file: ".gavel.agents.yaml",
  location: `${root}/.gavel.agents.yaml`,
  gitSha: "0123456789abcdef",
  config: { comments: { agents: { default: { enabled: true, prompt: "slides-review" } } } },
};
const folderLayer: SettingsLayer = {
  kind: "path",
  path: "slides/arb",
  file: ".gavel.yaml",
  location: `${root}/slides/arb/.gavel.yaml`,
  gitSha: "none",
  config: { comments: { agents: { pages: { "16-security.mdx": { assign: true, instructions: "Keep the tables" } } } } },
};
const dbLayer: SettingsLayer = { kind: "path", path: "slides", location: "db", config: { comments: { agents: { default: { presets: ["fast"] } } } } };

const settings = (overrides: Partial<FileSettings> = {}): FileSettings => ({
  scope,
  target,
  positions,
  layers: [projectLayer, dbLayer, folderLayer],
  merged: {},
  provenance: {},
  agent: {
    agent: { enabled: true, assign: true, prompt: "slides-review", presets: ["fast"], instructions: "Keep the tables" },
    provenance: {
      enabled: { layer: projectLayer.location },
      prompt: { layer: projectLayer.location },
      presets: { layer: "db:path:slides" },
      assign: { layer: folderLayer.location, page: "16-security.mdx" },
      instructions: { layer: folderLayer.location, page: "16-security.mdx" },
    },
    found: true,
  },
  ...overrides,
});

describe("layerLabelOf", () => {
  it("names a file layer by its location and a database layer by kind and path, as the server's provenance does", () => {
    expect([layerLabelOf(projectLayer), layerLabelOf(dbLayer)]).toEqual([`${root}/.gavel.agents.yaml`, "db:path:slides"]);
  });
});

describe("effectiveFieldsOf", () => {
  it("badges each configured field with its file basename or DB, (page) for a page entry and the short git sha", () => {
    expect(effectiveFieldsOf(settings())).toEqual([
      { field: "enabled", value: "on", badge: ".gavel.agents.yaml @0123456", title: `${root}/.gavel.agents.yaml` },
      { field: "assign", value: "yes", badge: ".gavel.yaml (page)", title: `${root}/slides/arb/.gavel.yaml` },
      { field: "prompt", value: "slides-review", badge: ".gavel.agents.yaml @0123456", title: `${root}/.gavel.agents.yaml` },
      { field: "presets", value: "fast", badge: "DB", title: "db:path:slides" },
      { field: "instructions", value: "Keep the tables", badge: ".gavel.yaml (page)", title: `${root}/slides/arb/.gavel.yaml` },
    ]);
  });

  it("shows whether the agent edits directly or proposes patches, badged with the layer that set it", () => {
    const proposing = settings({
      agent: { agent: { enabled: true, edits: "propose" }, provenance: { enabled: { layer: projectLayer.location }, edits: { layer: folderLayer.location, page: "16-security.mdx" } }, found: true },
    });
    expect(effectiveFieldsOf(proposing)).toEqual([
      { field: "enabled", value: "on", badge: ".gavel.agents.yaml @0123456", title: `${root}/.gavel.agents.yaml` },
      { field: "edits", value: "propose", badge: ".gavel.yaml (page)", title: `${root}/slides/arb/.gavel.yaml` },
    ]);
  });

  it("is empty when no layer configures the page", () => {
    const none = settings({ agent: { agent: {}, provenance: {}, found: false } });
    expect(effectiveFieldsOf(none)).toEqual([]);
  });

  it("refuses a provenance entry naming a layer that is not in the chain", () => {
    const broken = settings({ agent: { agent: { enabled: true }, provenance: { enabled: { layer: "/elsewhere/.gavel.yaml" } }, found: true } });
    expect(() => effectiveFieldsOf(broken)).toThrow(/\/elsewhere\/\.gavel\.yaml/);
  });
});

describe("agentLabel", () => {
  it.each([
    ["nothing configured", { agent: {}, provenance: {}, found: false }, "No page agent"],
    ["a disabled agent", { agent: { enabled: false }, provenance: {}, found: true }, "Page agent off"],
    ["an agent with no enabled flag", { agent: { prompt: "slides-review" }, provenance: {}, found: true }, "Page agent off"],
    ["the default prompt", { agent: { enabled: true }, provenance: {}, found: true }, "editor-comments · auto"],
    ["an assigning agent", { agent: { enabled: true, assign: true, prompt: "slides-review" }, provenance: {}, found: true }, "slides-review · auto + assign"],
  ])("labels %s", (_case, agent, label) => {
    expect(agentLabel(settings({ agent }))).toBe(label);
  });
});

describe("positions", () => {
  it("lists the nearest position first without changing the server's order", () => {
    expect(positionsNearestFirst(positions).map((position) => position.dir)).toEqual([`${root}/slides/arb`, `${root}/slides`, root, "/", home]);
    expect(positions[0]?.kind).toBe("user");
  });

  it.each([
    [positions[4]!, true, "This page · Folder slides/arb"],
    [positions[3]!, false, "Folder slides"],
    [positions[2]!, false, "Project"],
    [positions[2]!, true, "This page · Project"],
    [positions[0]!, false, "User (~)"],
    [positions[1]!, false, "Parent /"],
  ])("names %j (nearest %s) as %s", (position, nearest, name) => {
    expect(positionName(position, nearest)).toBe(name);
  });
});

describe("fileChoicesOf", () => {
  it("offers each existing file, Database only where the position takes one, and always a new snippet", () => {
    expect([fileChoicesOf(positions[2]!), fileChoicesOf(positions[3]!), fileChoicesOf(positions[1]!)]).toEqual([
      [
        { value: ".gavel.yaml", label: ".gavel.yaml" },
        { value: ".gavel.agents.yaml", label: ".gavel.agents.yaml" },
        { value: NEW_SNIPPET_CHOICE, label: "New snippet…" },
      ],
      [
        { value: DB_CHOICE, label: "Database" },
        { value: NEW_SNIPPET_CHOICE, label: "New snippet…" },
      ],
      [{ value: NEW_SNIPPET_CHOICE, label: "New snippet…" }],
    ]);
  });
});

describe("snippet names", () => {
  it.each([
    ["agents", undefined],
    ["  agents  ", undefined],
    ["", "Name the snippet."],
    ["   ", "Name the snippet."],
    ["a/b", "A snippet name cannot contain “/” or “.”."],
    ["a.b", "A snippet name cannot contain “/” or “.”."],
  ])("validates %j", (name, problem) => {
    expect(snippetNameProblem(name)).toBe(problem);
  });

  it("builds .gavel.<name>.yaml and refuses an invalid name", () => {
    expect(snippetFileName(" agents ")).toBe(".gavel.agents.yaml");
    expect(() => snippetFileName("a/b")).toThrow(/cannot contain/);
  });
});

describe("layerRefOf", () => {
  it.each([
    ["an existing file at a folder", positions[4]!, ".gavel.yaml", "", { kind: "path", path: "slides/arb", file: ".gavel.yaml" }],
    ["the database layer, with no file", positions[3]!, DB_CHOICE, "", { kind: "path", path: "slides" }],
    ["a new snippet at the project, with no path", positions[2]!, NEW_SNIPPET_CHOICE, "agents", { kind: "project", file: ".gavel.agents.yaml" }],
    ["a file in the user's home", positions[0]!, ".gavel.yaml", "", { kind: "user", file: ".gavel.yaml" }],
  ])("names %s", (_case, position, choice, snippet, ref) => {
    expect(layerRefOf({ position, choice, snippet })).toEqual(ref);
  });

  it("refuses a new snippet with an invalid name", () => {
    expect(() => layerRefOf({ position: positions[2]!, choice: NEW_SNIPPET_CHOICE, snippet: "" })).toThrow(/Name the snippet/);
  });
});

describe("pageKeyOf", () => {
  it.each([
    ["a folder", positions[4]!, "16-security.mdx"],
    ["a higher folder", positions[3]!, "arb/16-security.mdx"],
    ["the project root", positions[2]!, "slides/arb/16-security.mdx"],
    ["the user's home", positions[0]!, undefined],
    ["a parent directory", positions[1]!, undefined],
  ])("keys the page at %s", (_case, position, key) => {
    expect(pageKeyOf(target, position)).toBe(key);
  });

  it("is undefined for a position the page does not lie beneath", () => {
    expect(pageKeyOf(target, { kind: "path", path: "other", dir: `${root}/other`, files: [], db: false })).toBeUndefined();
  });
});

describe("existingAgentAt", () => {
  it("finds a layer's default and its entry for the page", () => {
    expect([
      existingAgentAt(settings(), { layer: { kind: "project", file: ".gavel.agents.yaml" }, page: false, position: positions[2]! }),
      existingAgentAt(settings(), { layer: { kind: "path", path: "slides/arb", file: ".gavel.yaml" }, page: true, position: positions[4]! }),
      existingAgentAt(settings(), { layer: { kind: "path", path: "slides" }, page: false, position: positions[3]! }),
    ]).toEqual([
      { enabled: true, prompt: "slides-review" },
      { assign: true, instructions: "Keep the tables" },
      { presets: ["fast"] },
    ]);
  });

  it("is undefined where the layer has no such entry or does not exist yet", () => {
    expect([
      existingAgentAt(settings(), { layer: { kind: "path", path: "slides/arb", file: ".gavel.yaml" }, page: false, position: positions[4]! }),
      existingAgentAt(settings(), { layer: { kind: "project", file: ".gavel.new.yaml" }, page: true, position: positions[2]! }),
    ]).toEqual([undefined, undefined]);
  });
});

describe("savedAgentOf", () => {
  const none = new Set<AgentField>();
  // The form shows the effective agent: enabled and assign inherited from a farther layer, the prompt from this entry.
  const form: CommentAgent = { enabled: true, assign: true, prompt: "slides-review", presets: ["fast"] };
  const entry: CommentAgent = { prompt: "slides-review" };

  it("writes only the entry's own fields when the user touched nothing, never the inherited ones the form shows", () => {
    expect(savedAgentOf({ entry, form, touched: none, cleared: none })).toEqual({ prompt: "slides-review" });
  });

  it("adds a touched field with the form's value, keeping an untouched field of the entry as it was", () => {
    const edited = { ...form, assign: false, prompt: "ignored" };
    expect(savedAgentOf({ entry, form: edited, touched: new Set<AgentField>(["assign"]), cleared: none })).toEqual({ prompt: "slides-review", assign: false });
  });

  it("starts from nothing when the layer has no entry yet", () => {
    expect(savedAgentOf({ entry: undefined, form, touched: new Set<AgentField>(["enabled"]), cleared: none })).toEqual({ enabled: true });
  });

  it("writes the edit mode only once it is touched, and drops it when cleared so it inherits again", () => {
    const withEdits: CommentAgent = { ...form, edits: "propose" };
    expect(savedAgentOf({ entry, form: withEdits, touched: none, cleared: none })).toEqual({ prompt: "slides-review" });
    expect(savedAgentOf({ entry, form: withEdits, touched: new Set<AgentField>(["edits"]), cleared: none })).toEqual({ prompt: "slides-review", edits: "propose" });
    expect(savedAgentOf({ entry: { ...entry, edits: "propose" }, form: withEdits, touched: none, cleared: new Set<AgentField>(["edits"]) })).toEqual({ prompt: "slides-review" });
  });

  it("drops a cleared field so it inherits again, and keeps the entry's others", () => {
    const both: CommentAgent = { enabled: false, prompt: "slides-review", presets: ["fast"] };
    expect(savedAgentOf({ entry: both, form, touched: none, cleared: new Set<AgentField>(["enabled", "presets"]) })).toEqual({ prompt: "slides-review" });
  });

  it("keeps a touched false and drops a touched field the form emptied", () => {
    const emptied: CommentAgent = { enabled: false, presets: [], variables: {}, instructions: "" };
    const touched = new Set<AgentField>(["enabled", "presets", "variables", "instructions"]);
    expect(savedAgentOf({ entry: { presets: ["fast"], instructions: "x" }, form: emptied, touched, cleared: none })).toEqual({ enabled: false });
  });

  it("lets an edit after a reset win over the reset", () => {
    const touched = new Set<AgentField>(["prompt"]);
    expect(savedAgentOf({ entry: { prompt: "a" }, form: { prompt: "b" }, touched, cleared: touched })).toEqual({ prompt: "b" });
  });

  it("does not mutate the entry", () => {
    const original: CommentAgent = { prompt: "slides-review", assign: true };
    savedAgentOf({ entry: original, form, touched: new Set<AgentField>(["prompt"]), cleared: new Set<AgentField>(["assign"]) });
    expect(original).toEqual({ prompt: "slides-review", assign: true });
  });
});

describe("runOf and deliverTargetOf", () => {
  it("builds the run target the server builds, omitting empty runOptions", () => {
    expect([
      deliverTargetOf({ enabled: true }),
      deliverTargetOf({ enabled: true, prompt: "slides-review", presets: [], variables: {}, spec: {}, instructions: "" }),
      deliverTargetOf({ enabled: true, prompt: "slides-review", presets: ["fast"], spec: { model: "opus" }, variables: { tone: "dry" }, instructions: "Keep the tables" }),
    ]).toEqual([
      { kind: "run", prompt: "editor-comments" },
      { kind: "run", prompt: "slides-review" },
      { kind: "run", prompt: "slides-review", runOptions: { presets: ["fast"], spec: { model: "opus" }, variables: { tone: "dry" }, instructions: "Keep the tables" } },
    ]);
  });

  it("hands the page agent's edit mode to a manual delivery so a proposing agent never edits directly", () => {
    expect([
      deliverTargetOf({ enabled: true, edits: "propose" }),
      deliverTargetOf({ enabled: true, edits: "direct", prompt: "slides-review" }),
    ]).toEqual([
      { kind: "run", prompt: "editor-comments", edits: "propose" },
      { kind: "run", prompt: "slides-review", edits: "direct" },
    ]);
  });

  it("round-trips a saved agent back into the form's run", () => {
    const agent = { enabled: true, prompt: "slides-review", presets: ["fast"] };
    expect(runOf(agent)).toEqual({ prompt: "slides-review", presets: ["fast"] });
  });
});

describe("viewer choice", () => {
  beforeEach(() => {
    const stored = new Map<string, string>();
    vi.stubGlobal("window", {
      localStorage: { getItem: (key: string) => stored.get(key) ?? null, setItem: (key: string, value: string) => void stored.set(key, value) },
    });
  });
  afterEach(() => vi.unstubAllGlobals());

  it("defaults to the configured run, and remembers a live session per page", () => {
    const before = loadViewerChoice(target);
    saveViewerChoice(target, { kind: "session", sessionId: "s-1" });
    expect([before, loadViewerChoice(target), loadViewerChoice(scope)]).toEqual([{ kind: "run" }, { kind: "session", sessionId: "s-1" }, { kind: "run" }]);
  });

  it("defaults to the launching session until the viewer stores a choice, including a run", () => {
    const launching = "launching-1";
    const before = loadViewerChoice(target, launching);
    saveViewerChoice(target, { kind: "run" });
    expect([before, loadViewerChoice(target, launching)]).toEqual([{ kind: "session", sessionId: launching }, { kind: "run" }]);
  });

  it("ignores a stored value that is not a choice", () => {
    window.localStorage.setItem(`gavel-edit:page-agent:${target}`, JSON.stringify({ kind: "session" }));
    expect(loadViewerChoice(target)).toEqual({ kind: "run" });
  });
});
