import { useMemo, useState } from "react";
import { PromptRunEditor, type AIPromptRunValue } from "@flanksource/clicky-ui/ai";
import { Button, Field, InputField, Modal, SegmentedControl, Select, Switch } from "@flanksource/clicky-ui/components";
import { Badge } from "@flanksource/clicky-ui/data";
import { UiRobotAi } from "@flanksource/clicky-ui/icons";
import { request } from "./api";
import type { AgentSession } from "./agent-sessions";
import { useRuntimeCatalogs } from "./NewAgentSessionDialog";
import {
  DEFAULT_AGENT_PROMPT,
  NEW_SNIPPET_CHOICE,
  agentLabel,
  effectiveFieldsOf,
  existingAgentAt,
  fileChoicesOf,
  layerRefOf,
  pageKeyOf,
  positionName,
  positionsNearestFirst,
  runOf,
  savedAgentOf,
  snippetNameProblem,
  type AgentField,
  type CommentAgent,
  type CommentEdits,
  type EffectiveField,
  type FileSettings,
  type LayerRef,
  type SaveAgentBody,
  type SettingsPosition,
  type ViewerChoice,
} from "./page-agent";
import { buildRunRequest, findPrompt, promptLabel, variablesSchemaOf } from "./run-request";

type DialogProps = {
  fileId: string;
  agent: FileSettings;
  choice: ViewerChoice;
  sessions: readonly AgentSession[];
  onChoose: (choice: ViewerChoice) => void;
  onSaved: () => Promise<void>;
};

/**
 * The open page's agent: a toolbar button naming the captain prompt that works
 * new threads here, opening a dialog that shows where each field comes from,
 * saves the agent to a chosen settings layer, and holds the viewer's own choice
 * of where manual hand-offs go.
 */
export function PageAgentControl(props: DialogProps) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button size="sm" variant={props.agent.agent.agent.enabled ? "outline" : "ghost"} onClick={() => setOpen(true)} title={`Page agent for ${props.agent.target}`}>
        <UiRobotAi className="mr-1 size-4" />
        {agentLabel(props.agent)}
      </Button>
      {open && <PageAgentDialog {...props} onSaved={async () => { await props.onSaved(); setOpen(false); }} onClose={() => setOpen(false)} />}
    </>
  );
}

/** The effective agent's configured fields, each badged with the settings layer entry that set it. */
function EffectiveAgent({ settings }: { settings: FileSettings }) {
  const fields = effectiveFieldsOf(settings);
  if (fields.length === 0) return <p className="text-sm text-muted-foreground">No settings layer configures an agent for this page yet.</p>;
  return (
    <dl className="grid grid-cols-[auto_1fr_auto] items-center gap-x-3 gap-y-1 text-sm">
      {fields.map(({ field, value, badge, title }) => (
        <div key={field} className="col-span-3 grid grid-cols-subgrid items-center">
          <dt className="text-muted-foreground">{field}</dt>
          <dd className="min-w-0 truncate font-mono text-xs" title={value}>
            {value || "—"}
          </dd>
          <dd>
            <Badge tone="neutral" variant="soft" clickToCopy={false}>
              <span title={title}>{badge}</span>
            </Badge>
          </dd>
        </div>
      ))}
    </dl>
  );
}

/** Where a form field's value comes from, that it was edited or will inherit, and the way to clear it from the layer's entry. */
function FieldState({ field, source, state, onInherit }: { field: AgentField; source?: EffectiveField; state?: "edited" | "inherit"; onInherit?: (() => void) | undefined }) {
  return (
    <span className="flex shrink-0 items-center gap-1">
      {state === "edited" && <Badge tone="warning" variant="soft" clickToCopy={false}>edited</Badge>}
      {state === "inherit" && <Badge tone="info" variant="soft" clickToCopy={false}>inherits</Badge>}
      {!state && source && (
        <Badge tone="neutral" variant="soft" clickToCopy={false}>
          <span title={source.title}>{source.badge}</span>
        </Badge>
      )}
      {onInherit && (
        <Button size="sm" variant="ghost" onClick={onInherit} title={`Remove ${field} from this entry so it inherits again`}>
          Inherit
        </Button>
      )}
    </span>
  );
}

function PageAgentDialog({ onClose, ...props }: DialogProps & { onClose: () => void }) {
  const { agent: settings, choice, sessions, onChoose } = props;
  const positions = useMemo(() => positionsNearestFirst(settings.positions), [settings.positions]);
  const [positionIndex, setPositionIndex] = useState(0);
  const position: SettingsPosition | undefined = positions[positionIndex];
  if (!position) throw new Error(`the settings of ${settings.target} list no layer positions`);
  const choices = fileChoicesOf(position);
  const [fileChoice, setFileChoice] = useState(choices[0]!.value);
  const [snippet, setSnippet] = useState("");
  const [page, setPage] = useState(true);

  const snippetProblem = fileChoice === NEW_SNIPPET_CHOICE ? snippetNameProblem(snippet) : undefined;
  const layer = snippetProblem ? undefined : layerRefOf({ position, choice: fileChoice, snippet });
  const where = positionName(position, positionIndex === 0);
  return (
    <Modal open onClose={onClose} title="Page agent" size="2xl">
      <div className="flex flex-col gap-3">
        <EffectiveAgent settings={settings} />
        <Field label="Applies to" htmlFor="page-agent-position" helper="Which settings layer of this page's chain the agent is saved to. Nearer layers override farther ones field by field.">
          <Select
            id="page-agent-position"
            value={String(positionIndex)}
            onChange={(event) => {
              const next = Number(event.currentTarget.value);
              setPositionIndex(next);
              setFileChoice(fileChoicesOf(positions[next]!)[0]!.value);
            }}
            options={positions.map((entry, index) => ({ value: String(index), label: positionName(entry, index === 0) }))}
          />
        </Field>
        <Field label="Save to" htmlFor="page-agent-file" helper={position.dir}>
          <Select id="page-agent-file" value={fileChoice} onChange={(event) => setFileChoice(event.currentTarget.value)} options={choices} />
        </Field>
        {fileChoice === NEW_SNIPPET_CHOICE && (
          <Field label="Snippet name" htmlFor="page-agent-snippet" required helper={snippetProblem ?? `Saved as .gavel.${snippet.trim()}.yaml`}>
            <InputField id="page-agent-snippet" value={snippet} onChange={setSnippet} />
          </Field>
        )}
        <Switch
          checked={page}
          onChange={setPage}
          label={page ? "Only this page (an entry for this file)" : `The default for every page beneath ${where}`}
        />
        <Field label="Your manual hand-offs" htmlFor="page-agent-viewer" helper="Where “Send to page agent” sends a thread for you. Kept in this browser only.">
          <Select
            id="page-agent-viewer"
            value={choice.kind === "session" ? choice.sessionId : ""}
            onChange={(event) => {
              const sessionId = event.currentTarget.value;
              onChoose(sessionId ? { kind: "session", sessionId } : { kind: "run" });
            }}
            options={[{ value: "", label: "A new run of this page agent" }, ...sessions.map((session) => ({ value: session.id, label: `Live session: ${session.title}` }))]}
          />
        </Field>
        <AgentForm
          key={`${positionIndex}|${fileChoice}|${page}`}
          {...props}
          position={position}
          page={page}
          where={where}
          {...(layer ? { layer } : {})}
          {...(snippetProblem ? { layerProblem: snippetProblem } : {})}
          onClose={onClose}
        />
      </div>
    </Modal>
  );
}

/** The layer entry's own saved agent (or the effective one, to start from), edited and saved to the chosen layer. */
function AgentForm({
  fileId,
  agent: settings,
  position,
  layer,
  layerProblem,
  page,
  where,
  onSaved,
  onClose,
}: DialogProps & { position: SettingsPosition; layer?: LayerRef; layerProblem?: string; page: boolean; where: string; onClose: () => void }) {
  const { models, families, context } = useRuntimeCatalogs(fileId);
  const existing = layer && existingAgentAt(settings, { layer, page, position });
  // The form shows what applies to the page: the entry's own fields over the effective agent.
  const seed: CommentAgent = { ...settings.agent.agent, ...existing };
  const seedRun = runOf(seed);
  const [enabled, setEnabled] = useState(seed.enabled ?? !settings.agent.found);
  const [assign, setAssign] = useState(seed.assign ?? false);
  const [edits, setEdits] = useState<CommentEdits>(seed.edits ?? "direct");
  const [promptId, setPromptId] = useState(seedRun.prompt);
  // Only touched fields are written, so a field nobody set here keeps inheriting; a first agent is enabled.
  const [touched, setTouched] = useState<ReadonlySet<AgentField>>(new Set(!existing && !settings.agent.found ? ["enabled"] : []));
  const [cleared, setCleared] = useState<ReadonlySet<AgentField>>(new Set());
  const touch = (...fields: AgentField[]) => {
    setTouched((current) => new Set([...current, ...fields]));
    setCleared((current) => new Set([...current].filter((field) => !fields.includes(field))));
  };
  const inherit = (field: AgentField) => {
    setCleared((current) => new Set([...current, field]));
    setTouched((current) => new Set([...current].filter((entry) => entry !== field)));
  };
  const effective = effectiveFieldsOf(settings);
  const fieldState = (field: AgentField) => (
    <FieldState
      field={field}
      source={effective.find((entry) => entry.field === field)}
      state={touched.has(field) ? "edited" : cleared.has(field) ? "inherit" : undefined}
      onInherit={existing?.[field] !== undefined && !cleared.has(field) ? () => inherit(field) : undefined}
    />
  );
  const [value, setValue] = useState<AIPromptRunValue>({
    ...(seedRun.spec ? { spec: seedRun.spec } : {}),
    ...(seedRun.presets ? { presets: seedRun.presets } : {}),
    ...(seedRun.variables ? { variables: seedRun.variables } : {}),
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>();

  const prompt = context.data && (findPrompt(context.data.prompts, promptId) ?? findPrompt(context.data.prompts, context.data.defaultPrompt));
  const built = useMemo(() => {
    if (!prompt || !context.data) return {};
    try {
      return {
        schema: variablesSchemaOf(prompt.variables, context.data.hostVariables),
        run: buildRunRequest({ prompt: prompt.id, value, instructions: seedRun.instructions ?? "", hostVariables: context.data.hostVariables }),
      };
    } catch (cause) {
      return { failure: cause instanceof Error ? cause.message : String(cause) };
    }
  }, [prompt, value, context.data]);

  const act = async (body: SaveAgentBody) => {
    setSaving(true);
    setError(undefined);
    try {
      await request<FileSettings>(`/api/file/${fileId}/settings/agent`, body, "PUT");
      await onSaved();
    } catch (cause) {
      setError(String(cause));
    } finally {
      setSaving(false);
    }
  };
  const saved = built.run && savedAgentOf({ entry: existing, form: { enabled, assign, edits, ...built.run }, touched, cleared });
  const save = () => void act({ layer: layer!, page, agent: saved! });
  const changeValue = (next: AIPromptRunValue) => {
    touch(...(["presets", "spec", "variables"] as const).filter((field) => JSON.stringify(next[field] ?? null) !== JSON.stringify(value[field] ?? null)));
    setValue(next);
  };
  const remove = () => void act({ layer: layer!, page, agent: null });
  // The user's and parent directories cannot be searched for the page's entry, so removal there is left to the server.
  const removable = !!layer && (!!existing || (page && pageKeyOf(settings.target, position) === undefined));

  const failure = models.error ?? families.error ?? context.error;
  const problems = [failure?.message, built.failure, layerProblem, error].filter((problem): problem is string => !!problem);
  return (
    <>
      <div className="flex items-center justify-between gap-2">
        <Switch
          checked={enabled}
          onChange={(next) => {
            setEnabled(next);
            touch("enabled");
          }}
          label="Auto-assign new threads to this agent"
        />
        {fieldState("enabled")}
      </div>
      <div className="flex items-center justify-between gap-2">
        <Switch
          checked={assign}
          onChange={(next) => {
            setAssign(next);
            touch("assign");
          }}
          disabled={!enabled}
          label="Show the agent as the thread's assignee as soon as it is created"
        />
        {fieldState("assign")}
      </div>
      <Field
        label="Agent edits"
       
        helper={edits === "propose" ? "The agent must not edit project files: it answers with a patch you apply, stage, reject or ask to revise." : "The agent edits the project's files itself."}
      >
        <div className="flex items-center justify-between gap-2">
          <SegmentedControl<CommentEdits>
            size="sm"
            aria-label="Agent edits"
            value={edits}
            options={[
              { id: "direct", label: "Edit directly" },
              { id: "propose", label: "Propose patches" },
            ]}
            onChange={(next) => {
              setEdits(next);
              touch("edits");
            }}
          />
          {fieldState("edits")}
        </div>
      </Field>
      {!models.data || !families.data || !context.data || !prompt ? (
        !failure && <p className="text-sm text-muted-foreground">Loading prompts and AI options…</p>
      ) : (
        <>
          <Field label="Prompt" htmlFor="page-agent-prompt" helper={prompt.description ?? `Unset runs ${DEFAULT_AGENT_PROMPT}.`}>
            <div className="flex items-center gap-2">
              <Select
                id="page-agent-prompt"
                value={prompt.id}
                onChange={(event) => {
                  setPromptId(event.currentTarget.value);
                  touch("prompt");
                  changeValue({ ...value, variables: {} });
                }}
                options={context.data.prompts.map((entry) => ({ value: entry.id, label: promptLabel(entry) }))}
              />
              {fieldState("prompt")}
            </div>
          </Field>
          <PromptRunEditor
            key={prompt.id}
            value={value}
            onChange={changeValue}
            presets={context.data.presets}
            models={models.data}
            families={families.data}
            {...(built.schema ? { variablesSchema: built.schema } : {})}
          />
          <div className="flex flex-col gap-1">
            {(["presets", "spec", "variables", "instructions"] as const).map((field) => {
              const state = fieldState(field);
              const shown = existing?.[field] !== undefined || touched.has(field) || cleared.has(field);
              return shown && <div key={field} className="flex items-center gap-2 text-sm"><span className="text-muted-foreground">{field}</span>{state}</div>;
            })}
          </div>
        </>
      )}
      {problems.map((problem) => (
        <p key={problem} role="alert" className="text-sm text-destructive">
          {problem}
        </p>
      ))}
      <div className="flex justify-between gap-2 border-t border-border pt-3">
        <Button variant="ghost" onClick={remove} disabled={saving || !removable} title={removable ? `Remove ${page ? "this page's entry" : "the default"} at ${where}` : `${where} has no such agent`}>
          Remove {page ? "page entry" : "default"}
        </Button>
        <div className="flex gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={save} disabled={saving || !saved || Object.keys(saved).length === 0 || !layer} title={saved && Object.keys(saved).length === 0 ? "Nothing to save: every field would inherit" : undefined}>
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>
    </>
  );
}
