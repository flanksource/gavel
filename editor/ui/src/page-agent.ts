import { useCallback, useEffect, useState } from "react";
import type { AISpecRuntimeSpec } from "@flanksource/clicky-ui/ai";
import { request } from "./api";
import type { AgentSession } from "./agent-sessions";
import { runTargetOf, type RunRequest, type RunTarget } from "./run-request";

/** The captain prompt a comment agent runs when its entry names none (verify.DefaultCommentAgentPrompt). */
export const DEFAULT_AGENT_PROMPT = "editor-comments";

/**
 * A comment agent as one settings layer configures it (verify.CommentAgent).
 * Every field is optional: a layer sets only the fields it names and the rest
 * are inherited from the layers above; `enabled: false` opts out.
 */
export type CommentAgent = {
  enabled?: boolean;
  assign?: boolean;
  prompt?: string;
  presets?: string[];
  spec?: AISpecRuntimeSpec;
  variables?: Record<string, unknown>;
  instructions?: string;
  /** `direct` (default when unset): the agent edits the project's files; `propose`: it answers with a patch a human accepts or rejects. */
  edits?: CommentEdits;
};

/** How a comment agent changes the project (verify.CommentAgent.Edits). */
export type CommentEdits = "direct" | "propose";

/** What a settings layer position is to the file's project (settings.Kind). */
export type SettingsKind = "user" | "parent" | "project" | "path";

/** One directory of the file's .gavel.yaml chain and what can be saved there (settings.Position). */
export type SettingsPosition = {
  kind: SettingsKind;
  /** The folder relative to the project for a path position; the absolute directory for a parent. */
  path?: string;
  dir: string;
  files: string[];
  /** Whether the position takes a database layer. */
  db: boolean;
};

/** A settings layer: the position and file ("" for the database layer) a PUT names (settings.Ref). */
export type LayerRef = { kind: SettingsKind; path?: string; file?: string };

/** One configuration layer, in the order it applies (settings.Layer). */
export type SettingsLayer = LayerRef & {
  /** The file's absolute path, or "db" for a database layer. */
  location: string;
  /** The repository HEAD a DB-mode file row was synced at; "none" outside git. */
  gitSha?: string;
  config: { comments?: { agents?: { default?: CommentAgent; pages?: Record<string, CommentAgent> } } } & Record<string, unknown>;
};

/** The entry that set a resolved agent field: a layer's default, or its entry for `page`. */
export type CommentAgentSource = { layer: string; page?: string };

/** A file's agent and, per field, the entry that last set it (verify.ResolvedCommentAgent). */
export type ResolvedCommentAgent = {
  agent: CommentAgent;
  provenance: Partial<Record<keyof CommentAgent, CommentAgentSource>>;
  found: boolean;
};

/** GET /api/file/{id}/settings, and the answer to PUT /api/file/{id}/settings/agent. */
export type FileSettings = {
  scope: string;
  target: string;
  positions: SettingsPosition[];
  layers: SettingsLayer[];
  merged: Record<string, unknown>;
  provenance: Record<string, string>;
  agent: ResolvedCommentAgent;
  /** The agent session `gavel edit` was started from: new comments go to it, ahead of `agent`. */
  session?: string;
};

/** The body of PUT /api/file/{id}/settings/agent: `agent` null removes the entry. */
export type SaveAgentBody = { layer: LayerRef; page: boolean; agent: CommentAgent | null };

/** How settings.Layer.Label names a layer in provenance. */
export function layerLabelOf(layer: SettingsLayer): string {
  return layer.file ? layer.location : `db:${layer.kind}:${layer.path ?? ""}`;
}

const AGENT_FIELDS = ["enabled", "assign", "edits", "prompt", "presets", "spec", "variables", "instructions"] as const;

/** A field of CommentAgent. */
export type AgentField = (typeof AGENT_FIELDS)[number];

function fieldValue(agent: CommentAgent, field: AgentField): string {
  switch (field) {
    case "enabled":
      return agent.enabled ? "on" : "off";
    case "assign":
      return agent.assign ? "yes" : "no";
    case "edits":
      return agent.edits ?? "direct";
    case "prompt":
      return agent.prompt ?? DEFAULT_AGENT_PROMPT;
    case "presets":
      return (agent.presets ?? []).join(", ");
    case "spec":
      return JSON.stringify(agent.spec ?? {});
    case "variables":
      return JSON.stringify(agent.variables ?? {});
    case "instructions":
      return agent.instructions ?? "";
  }
}

/** A field of the effective agent, its value and the badge naming the layer entry that set it. */
export type EffectiveField = { field: AgentField; value: string; badge: string; title: string };

/**
 * The effective agent's configured fields, each badged with where it comes from:
 * the file's basename or "DB", "(page)" for a page entry, and the short git sha
 * of a file synced in DB mode. A provenance entry naming a layer outside the
 * chain is a server contract break and throws.
 */
export function effectiveFieldsOf(settings: FileSettings): EffectiveField[] {
  const { agent, provenance } = settings.agent;
  return AGENT_FIELDS.flatMap((field) => {
    const source = provenance[field];
    if (!source) return [];
    const layer = settings.layers.find((candidate) => layerLabelOf(candidate) === source.layer);
    if (!layer) throw new Error(`the ${field} of the page agent comes from ${source.layer}, which is not one of the file's settings layers`);
    const sha = layer.gitSha && layer.gitSha !== "none" ? ` @${layer.gitSha.slice(0, 7)}` : "";
    const name = layer.file ? layer.location.split("/").pop()! : "DB";
    return [{ field, value: fieldValue(agent, field), badge: `${name}${source.page ? " (page)" : ""}${sha}`, title: source.layer }];
  });
}

/**
 * The toolbar label: the launching session new threads go to (by its live
 * title, else its short id), else the prompt that works them and whether it
 * assigns, or that none does.
 */
export function agentLabel(settings: FileSettings, sessions: readonly AgentSession[]): string {
  if (settings.session) {
    const live = sessions.find((session) => session.id === settings.session);
    return live ? `${live.source} · ${live.title}` : `Session ${settings.session.slice(0, 8)}`;
  }
  const { agent, found } = settings.agent;
  if (!found) return "No page agent";
  if (!agent.enabled) return "Page agent off";
  return `${agent.prompt || DEFAULT_AGENT_PROMPT} · auto${agent.assign ? " + assign" : ""}`;
}

/** The positions nearest the file first: the picker's order, the reverse of the server's precedence order. */
export function positionsNearestFirst(positions: readonly SettingsPosition[]): SettingsPosition[] {
  return [...positions].reverse();
}

const POSITION_NAMES: Record<SettingsKind, (position: SettingsPosition) => string> = {
  user: () => "User (~)",
  parent: (position) => `Parent ${position.dir}`,
  project: () => "Project",
  path: (position) => `Folder ${position.path}`,
};

/** The picker's name for a position; the nearest is the file's own folder, labelled as the page's. */
export function positionName(position: SettingsPosition, nearest: boolean): string {
  const name = POSITION_NAMES[position.kind](position);
  return nearest ? `This page · ${name}` : name;
}

/** The file choice that saves to the position's database layer. */
export const DB_CHOICE = "::db";
/** The file choice that saves to a new .gavel.<name>.yaml snippet. */
export const NEW_SNIPPET_CHOICE = "::new";

/** What can be saved at a position: each existing file, its database layer when it takes one, and a new snippet. */
export function fileChoicesOf(position: SettingsPosition): { value: string; label: string }[] {
  return [
    ...position.files.map((file) => ({ value: file, label: file })),
    ...(position.db ? [{ value: DB_CHOICE, label: "Database" }] : []),
    { value: NEW_SNIPPET_CHOICE, label: "New snippet…" },
  ];
}

/** Why `name` cannot name a .gavel.<name>.yaml snippet, or undefined. */
export function snippetNameProblem(name: string): string | undefined {
  const trimmed = name.trim();
  if (!trimmed) return "Name the snippet.";
  if (/[/.]/.test(trimmed)) return "A snippet name cannot contain “/” or “.”.";
  return undefined;
}

export function snippetFileName(name: string): string {
  const problem = snippetNameProblem(name);
  if (problem) throw new Error(problem);
  return `.gavel.${name.trim()}.yaml`;
}

/** The layer a PUT names for a position and file choice; the database layer carries no file. */
export function layerRefOf({ position, choice, snippet }: { position: SettingsPosition; choice: string; snippet: string }): LayerRef {
  const file = choice === DB_CHOICE ? undefined : choice === NEW_SNIPPET_CHOICE ? snippetFileName(snippet) : choice;
  return { kind: position.kind, ...(position.path ? { path: position.path } : {}), ...(file ? { file } : {}) };
}

/**
 * The key of the page in a layer's comments.agents.pages at the position: its
 * path relative to the position's folder. Undefined for the user's and parent
 * directories, which the target URI does not locate the page in, and for a
 * folder the page does not lie beneath.
 */
export function pageKeyOf(target: string, position: SettingsPosition): string | undefined {
  const marker = "/files/";
  const at = target.indexOf(marker);
  if (at < 0) throw new Error(`settings target ${JSON.stringify(target)} has no ${marker} segment`);
  const inProject = target.slice(at + marker.length);
  if (position.kind === "project") return inProject;
  if (position.kind !== "path" || !position.path) return undefined;
  return inProject.startsWith(`${position.path}/`) ? inProject.slice(position.path.length + 1) : undefined;
}

/** The agent the layer already holds as its default, or as its entry for the page. */
export function existingAgentAt(
  settings: FileSettings,
  { layer, page, position }: { layer: LayerRef; page: boolean; position: SettingsPosition },
): CommentAgent | undefined {
  const found = settings.layers.find(
    (candidate) => candidate.kind === layer.kind && (candidate.path ?? "") === (layer.path ?? "") && (candidate.file ?? "") === (layer.file ?? ""),
  );
  const agents = found?.config.comments?.agents;
  if (!page) return agents?.default;
  const key = pageKeyOf(settings.target, position);
  return key === undefined ? undefined : agents?.pages?.[key];
}

const isUnset = (value: unknown): boolean =>
  value === undefined || value === "" || (Array.isArray(value) && value.length === 0) || (typeof value === "object" && value !== null && !Array.isArray(value) && Object.keys(value).length === 0);

/**
 * The agent saved to a layer entry. Layers merge field by field, so only the
 * fields the entry already sets and the ones the user touched are written:
 * `form` (which shows the effective agent, inherited fields included) supplies
 * a touched field, an empty touched field is left unset, and a cleared field is
 * dropped so it inherits again. A touched field wins over a cleared one.
 */
export function savedAgentOf({
  entry,
  form,
  touched,
  cleared,
}: {
  entry: CommentAgent | undefined;
  form: CommentAgent;
  touched: ReadonlySet<AgentField>;
  cleared: ReadonlySet<AgentField>;
}): CommentAgent {
  const entries = AGENT_FIELDS.flatMap((field) => {
    const value = touched.has(field) ? form[field] : cleared.has(field) ? undefined : entry?.[field];
    return isUnset(value) ? [] : [[field, value] as const];
  });
  // Each field is read from CommentAgent above, so the rebuilt object is one.
  return Object.fromEntries(entries) as CommentAgent;
}

/** The run a comment agent starts, in the new-session dialog's request shape; empty fields are left out. */
export function runOf(agent: CommentAgent): RunRequest {
  return {
    prompt: agent.prompt || DEFAULT_AGENT_PROMPT,
    ...(agent.presets?.length ? { presets: agent.presets } : {}),
    ...(agent.spec && Object.keys(agent.spec).length > 0 ? { spec: agent.spec } : {}),
    ...(agent.variables && Object.keys(agent.variables).length > 0 ? { variables: agent.variables } : {}),
    ...(agent.instructions ? { instructions: agent.instructions } : {}),
  };
}

/** The delivery target "Send to page agent" hands a thread to, as the server builds it from the agent, including how it edits. */
export function deliverTargetOf(agent: CommentAgent): RunTarget {
  return { ...runTargetOf(runOf(agent)), ...(agent.edits ? { edits: agent.edits } : {}) };
}

/** A viewer's own hand-off choice for a page: a new run of the page's configured agent, or a live session. */
export type ViewerChoice = { kind: "run" } | { kind: "session"; sessionId: string };

const VIEWER_CHOICE = "gavel-edit:page-agent:";

/**
 * The viewer's stored choice for target, else the session `gavel edit` was
 * started from (`launching`), else a run of the configured agent.
 */
export function loadViewerChoice(target: string, launching?: string): ViewerChoice {
  const fallback: ViewerChoice = launching ? { kind: "session", sessionId: launching } : { kind: "run" };
  try {
    const stored = window.localStorage.getItem(VIEWER_CHOICE + target);
    const parsed: unknown = stored ? JSON.parse(stored) : null;
    if (parsed && typeof parsed === "object" && (parsed as ViewerChoice).kind === "session" && typeof (parsed as { sessionId?: unknown }).sessionId === "string")
      return parsed as ViewerChoice;
    if (parsed && typeof parsed === "object" && (parsed as ViewerChoice).kind === "run") return { kind: "run" };
    return fallback;
  } catch {
    // A per-viewer convenience only: unreadable storage hands off to the default.
    return fallback;
  }
}

export function saveViewerChoice(target: string, choice: ViewerChoice) {
  try {
    window.localStorage.setItem(VIEWER_CHOICE + target, JSON.stringify(choice));
  } catch {
    // Storage may be blocked; the choice still applies until the page reloads.
  }
}

/** The file's settings and the viewer's hand-off choice; none until a document is open. */
export function usePageAgent(fileId: string | null, onError: (message: string) => void) {
  const [agent, setAgent] = useState<FileSettings>();
  const [choice, setChoice] = useState<ViewerChoice>({ kind: "run" });
  const refresh = useCallback(async () => {
    if (!fileId) return;
    try {
      const next = await request<FileSettings>(`/api/file/${fileId}/settings`);
      setAgent(next);
      setChoice(loadViewerChoice(next.target, next.session));
    } catch (cause) {
      onError(String(cause));
    }
  }, [fileId]);
  useEffect(() => {
    setAgent(undefined);
    void refresh();
  }, [refresh]);
  const choose = (next: ViewerChoice) => {
    if (!agent) throw new Error("the page agent is still loading");
    saveViewerChoice(agent.target, next);
    setChoice(next);
  };
  return { agent, choice, choose, refresh };
}
