import { useMemo, useState } from "react";
import { PromptRunEditor, type AIPromptRunValue } from "@flanksource/clicky-ui/ai";
import { Button, Field, Modal, Select, Switch } from "@flanksource/clicky-ui/components";
import { UiRobotAi } from "@flanksource/clicky-ui/icons";
import { request } from "./api";
import type { AgentSession } from "./agent-sessions";
import { useRuntimeCatalogs } from "./NewAgentSessionDialog";
import { agentLabel, originOf, routeBodyOf, runOf, type AgentScope, type CommentRoute, type PageAgent, type ViewerChoice } from "./page-agent";
import { buildRunRequest, findPrompt, promptLabel, variablesSchemaOf } from "./run-request";

type DialogProps = {
  fileId: string;
  agent: PageAgent;
  choice: ViewerChoice;
  sessions: readonly AgentSession[];
  onChoose: (choice: ViewerChoice) => void;
  onSaved: () => Promise<void>;
};

/**
 * The open page's agent: a toolbar button naming the captain prompt that works
 * new threads here, opening a dialog that saves it to the page's or folder's
 * .gavel.yaml, and the viewer's own choice of where manual hand-offs go.
 */
export function PageAgentControl(props: DialogProps) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button size="sm" variant={props.agent.effective?.enabled ? "outline" : "ghost"} onClick={() => setOpen(true)} title={`Configured in ${props.agent.configPath}`}>
        <UiRobotAi className="mr-1 size-4" />
        {agentLabel(props.agent)}
      </Button>
      {open && <PageAgentDialog {...props} onSaved={async () => { await props.onSaved(); setOpen(false); }} onClose={() => setOpen(false)} />}
    </>
  );
}

function PageAgentDialog({ onClose, ...props }: DialogProps & { onClose: () => void }) {
  const { agent, choice, sessions, onChoose } = props;
  const [scope, setScope] = useState<AgentScope>(originOf(agent) === "folder" ? "folder" : "page");
  return (
    <Modal open onClose={onClose} title="Page agent" size="2xl">
      <div className="flex flex-col gap-3">
        <p className="text-sm text-muted-foreground">
          Saved to {agent.configPath}.{originOf(agent) === "inherited" && " This page currently inherits an agent from a .gavel.yaml further up."}
        </p>
        <Field label="Applies to" htmlFor="page-agent-scope" helper={scope === "page" ? "Only this file." : "Every page in this folder without an agent of its own."}>
          <Select
            id="page-agent-scope"
            value={scope}
            onChange={(event) => setScope(event.currentTarget.value as AgentScope)}
            options={[
              { value: "page", label: agent.page ? "This page (configured)" : "This page" },
              { value: "folder", label: agent.folder ? "This folder (configured)" : "This folder" },
            ]}
          />
        </Field>
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
        <AgentForm key={scope} {...props} scope={scope} onClose={onClose} />
      </div>
    </Modal>
  );
}

/** The scope's own saved agent (or the inherited one, to start from), edited and saved to .gavel.yaml. */
function AgentForm({ fileId, agent, scope, onSaved, onClose }: DialogProps & { scope: AgentScope; onClose: () => void }) {
  const { models, families, context } = useRuntimeCatalogs(fileId);
  const existing: CommentRoute | null = scope === "page" ? agent.page : agent.folder;
  const seed = existing ?? agent.effective;
  const seedRun = seed ? runOf(seed) : {};
  const [enabled, setEnabled] = useState(seed?.enabled ?? true);
  const [assign, setAssign] = useState(seed?.assign ?? false);
  const [promptId, setPromptId] = useState(seedRun.prompt);
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

  const act = async (change: () => Promise<unknown>) => {
    setSaving(true);
    setError(undefined);
    try {
      await change();
      await onSaved();
    } catch (cause) {
      setError(String(cause));
    } finally {
      setSaving(false);
    }
  };
  const save = () => act(() => request<CommentRoute>("/api/v1/comments/routes", routeBodyOf({ agent, scope, enabled, assign, run: built.run! }), "PUT"));
  const remove = () =>
    act(async () => {
      const response = await fetch(`/api/v1/comments/routes/${encodeURIComponent(existing!.id)}`, { method: "DELETE", headers: { "Content-Type": "application/json" } });
      if (!response.ok) throw new Error(`${response.status}: ${await response.text()}`);
    });

  const failure = models.error ?? families.error ?? context.error;
  const problems = [failure?.message, built.failure, error].filter((problem): problem is string => !!problem);
  return (
    <>
      <Switch checked={enabled} onChange={setEnabled} label="Auto-assign new threads to this agent" />
      <Switch checked={assign} onChange={setAssign} disabled={!enabled} label="Show the agent as the thread's assignee as soon as it is created" />
      {!models.data || !families.data || !context.data || !prompt ? (
        !failure && <p className="text-sm text-muted-foreground">Loading prompts and AI options…</p>
      ) : (
        <>
          <Field label="Prompt" htmlFor="page-agent-prompt" helper={prompt.description}>
            <Select
              id="page-agent-prompt"
              value={prompt.id}
              onChange={(event) => {
                setPromptId(event.currentTarget.value);
                setValue((current) => ({ ...current, variables: {} }));
              }}
              options={context.data.prompts.map((entry) => ({ value: entry.id, label: promptLabel(entry) }))}
            />
          </Field>
          <PromptRunEditor
            key={prompt.id}
            value={value}
            onChange={setValue}
            presets={context.data.presets}
            models={models.data}
            families={families.data}
            {...(built.schema ? { variablesSchema: built.schema } : {})}
          />
        </>
      )}
      {problems.map((problem) => (
        <p key={problem} role="alert" className="text-sm text-destructive">
          {problem}
        </p>
      ))}
      <div className="flex justify-between gap-2 border-t border-border pt-3">
        <Button variant="ghost" onClick={() => void remove()} disabled={saving || !existing} title={existing ? `Remove this ${scope}'s agent from ${agent.configPath}` : `This ${scope} has no agent of its own`}>
          Remove {scope} agent
        </Button>
        <div className="flex gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={() => void save()} disabled={saving || !built.run}>
            {saving ? "Saving…" : "Save to .gavel.yaml"}
          </Button>
        </div>
      </div>
    </>
  );
}
