import { useEffect, useState, type ReactNode } from "react";
import { Button, Modal } from "@flanksource/clicky-ui";
import {
  SpecRuntimeBar,
  type AISpecRuntimeValue,
} from "@flanksource/clicky-ui/ai";

import type {
  AgentSessionGrouping,
  AgentSessionRuntime,
} from "../contracts/agent-session-store";
import {
  runtimeOfSpecValue,
  specValueOfRuntime,
  type RuntimeCatalog,
} from "./runtime-families";

function describeError(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/** `run` starts a session for a thread; `configure` only saves the runtime new sessions use. */
export type RunAgentDialogMode = "run" | "configure" | "create" | "assign" | "assign-all";

const COPY: Record<
  RunAgentDialogMode,
  { title: string; submit: string; busy: string }
> = {
  create: {
    title: "New agent session",
    submit: "Create session",
    busy: "Creating…",
  },
  run: {
    title: "Run agent",
    submit: "Start session and run",
    busy: "Starting…",
  },
  configure: {
    title: "Agent runtime",
    submit: "Save runtime",
    busy: "Saving…",
  },
  assign: {
    title: "Assign to a new agent",
    submit: "Start session and send",
    busy: "Sending…",
  },
  "assign-all": {
    title: "Assign all to a new session",
    submit: "Start session and send all",
    busy: "Sending…",
  },
};

function intro(
  mode: RunAgentDialogMode,
  grouping: AgentSessionGrouping,
): string {
  if (mode === "create")
    return "Create a managed Captain session and use it for this page's comments. Existing comments are not sent.";
  if (mode === "configure") {
    return "Every new agent session starts with this runtime. Sessions already running keep theirs.";
  }
  if (mode === "assign")
    return "Starts a new agent session with this runtime and sends it this thread. The page's agent settings stay as they are.";
  if (mode === "assign-all")
    return "Starts a new agent session with this runtime and sends it every open thread on this page in one message. The page's agent settings stay as they are.";
  return grouping === "page"
    ? "No agent session is running for this page. This starts the page's session; later comments on the page are sent into it."
    : "No agent session is running. This starts the project's shared session; later comments are sent into it.";
}

function RuntimeChooser({
  catalog,
  current,
  mode,
  grouping,
  busy,
  error,
  onRun,
  onCancel,
}: {
  catalog: RuntimeCatalog;
  current: AgentSessionRuntime | undefined;
  mode: RunAgentDialogMode;
  grouping: AgentSessionGrouping;
  busy: boolean;
  error: string | null;
  onRun: (runtime: AgentSessionRuntime) => void;
  onCancel: () => void;
}) {
  const [value, setValue] = useState<AISpecRuntimeValue>(() =>
    specValueOfRuntime(catalog.families, current),
  );
  const chosen = runtimeOfSpecValue(value, current?.presets);

  return (
    <div className="space-y-4 text-sm">
      <p className="text-muted-foreground">{intro(mode, grouping)}</p>
      {/* Permissions, workspace and commit settings come from the session prompt, so only the model row is offered. */}
      <SpecRuntimeBar
        value={value}
        onChange={setValue}
        models={catalog.models}
        families={catalog.families}
        sections={["model"]}
        ariaLabel="Agent runtime"
      />
      {error && (
        <p role="alert" className="text-xs text-destructive">
          {error}
        </p>
      )}
      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
        <Button
          disabled={busy || !chosen}
          onClick={() => {
            if (chosen) onRun(chosen);
          }}
        >
          {busy ? COPY[mode].busy : COPY[mode].submit}
        </Button>
      </div>
    </div>
  );
}

/**
 * Chooses the runtime new agent sessions start with, through clicky-ui's
 * runtime bar fed by captain's agent-mode runtimes and model catalog.
 */
export function RunAgentDialog({
  open,
  mode,
  grouping,
  current,
  error,
  busy,
  onRun,
  onClose,
  fields,
  canSubmit = true,
  loadCatalog,
}: {
  open: boolean;
  mode: RunAgentDialogMode;
  grouping: AgentSessionGrouping;
  /** The configured runtime, preselected when present. */
  current: AgentSessionRuntime | undefined;
  error: string | null;
  busy: boolean;
  onRun: (runtime: AgentSessionRuntime) => void;
  onClose: () => void;
  fields?: ReactNode;
  canSubmit?: boolean;
  /** Loads the configured Gavel dashboard's runtimes and models. */
  loadCatalog: () => Promise<RuntimeCatalog>;
}) {
  const [catalog, setCatalog] = useState<RuntimeCatalog | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setCatalog(null);
    setLoadError(null);
    loadCatalog().then(setCatalog, (cause: unknown) => setLoadError(describeError(cause)));
  }, [open, loadCatalog]);

  return (
    <Modal open={open} onClose={onClose} title={COPY[mode].title} size="lg">
      {loadError ? (
        <p role="alert" className="text-sm text-destructive">
          {loadError}
        </p>
      ) : catalog === null ? (
        <p role="status" className="text-sm text-muted-foreground">
          Loading runtimes from captain…
        </p>
      ) : catalog.families.length === 0 ? (
        <p role="alert" className="text-sm text-destructive">
          Captain reports no available agent runtime.
        </p>
      ) : (
        <>
          {fields}
          {!canSubmit ? null : (
            <RuntimeChooser
              catalog={catalog}
              current={current}
              mode={mode}
              grouping={grouping}
              busy={busy}
              error={error}
              onRun={onRun}
              onCancel={onClose}
            />
          )}
        </>
      )}
    </Modal>
  );
}
