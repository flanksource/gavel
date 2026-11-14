import { useState, type ReactNode } from "react";
import { Badge, Button, SegmentedControl, Switch } from "@flanksource/clicky-ui";
import { SpecRuntimeBar } from "@flanksource/clicky-ui/ai";
import { AgentIdentity } from "@flanksource/clicky-ui/chat";

import type { CommentEdits } from "../comments/gavel-delivery";
import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import { RuntimeIdentity } from "./RuntimeIdentity";
import { runtimeOfSpecValue, specValueOfRuntime, type RuntimeCatalog } from "./runtime-families";

/** A live agent session the picker lists; the host maps its own session shape onto this. */
export type AgentPickerSession = {
  id: string;
  title: string;
  /** The agent that runs it (claude, codex); names the identity when no `model` is known. */
  source: string;
  model?: string | undefined;
  cwd?: string | undefined;
  status?: string | undefined;
  /** The session has ended and delivering to it resumes it; shown in place of its status. */
  resumable?: boolean | undefined;
};

export type AgentPickerProps = {
  heading?: ReactNode;
  busy?: boolean | undefined;
  /** Shown in place of the settings rows while the host is still loading them. */
  notice?: ReactNode;
  error?: string | null | undefined;
  grouping?: { value: AgentSessionGrouping; onChange: (grouping: AgentSessionGrouping) => void } | undefined;
  /**
   * The runtime new sessions start with. With `onChoose` the host opens its own
   * chooser; otherwise the row expands the runtime bar inline from `catalog`.
   */
  runtime?: {
    value: AgentSessionRuntime | undefined;
    ariaLabel?: string | undefined;
    /** What the row says while no runtime is set. */
    emptyLabel?: string | undefined;
    onChoose?: (() => void) | undefined;
    onChange?: ((runtime: AgentSessionRuntime) => void) | undefined;
    catalog?: RuntimeCatalog | undefined;
    /** Offers a button that unsets the runtime while one is set. */
    onClear?: (() => void) | undefined;
    clearLabel?: string | undefined;
  } | undefined;
  autoRun?: { checked: boolean; onChange: (enabled: boolean) => void; disabled?: boolean; title?: string; label?: string } | undefined;
  /** How the auto-run agent changes the project: it edits files itself, or answers with a patch a human accepts. */
  edits?: { value: CommentEdits; onChange: (edits: CommentEdits) => void; disabled?: boolean; title?: string } | undefined;
  /** Omitted where only the settings are edited. */
  sessions?: {
    items: AgentPickerSession[];
    /** `error` shows neither the loading nor the empty message; the host passes the failure as `error`. */
    state: "loading" | "ready" | "error";
    /** The session comments go to; with `onNew`, `undefined` means a new session. Without `onNew` nothing is marked. */
    selectedId?: string | undefined;
    onSelect: (session: AgentPickerSession) => void;
    selectLabel: (session: AgentPickerSession) => string;
    /** Offers a "New session" choice; choosing it shows the runtime bar. */
    onNew?: () => void;
    onAll?: () => void;
    /** Page-level "assign every open thread": to a listed session, or to a new one. */
    assignAll?: {
      /** Open threads that would be sent; the buttons are disabled at zero. */
      count: number;
      onSession: (session: AgentPickerSession) => void;
      onNew: () => void;
    } | undefined;
  } | undefined;
};

function InlineRuntimeBar({ runtime, catalog, onChange }: { runtime: AgentSessionRuntime | undefined; catalog: RuntimeCatalog | undefined; onChange: (runtime: AgentSessionRuntime) => void }) {
  if (!catalog) return <p role="status" className="text-muted-foreground">Loading models…</p>;
  return (
    // Permissions, workspace and commit settings come from the session prompt, so only the model row is offered.
    <SpecRuntimeBar
      // An unset runtime means the host's default model, so the bar names none instead of seeding the catalog's default.
      value={runtime ? specValueOfRuntime(catalog.families, runtime) : { mode: "agent" }}
      onChange={(value) => {
        const next = runtimeOfSpecValue(value, runtime?.presets);
        if (next) onChange(next);
      }}
      models={catalog.models}
      families={catalog.families}
      sections={["model"]}
      ariaLabel="Agent runtime"
    />
  );
}

function SessionRow({ session, label, pressed, onSelect, assignAll }: { session: AgentPickerSession; label: string; pressed: boolean | undefined; onSelect: () => void; assignAll?: { disabled: boolean; onAssign: () => void } | undefined }) {
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="sm"
        className="h-auto min-w-0 flex-1 justify-start gap-2 py-2 text-left"
        aria-label={label}
        {...(pressed === undefined ? {} : { "aria-pressed": pressed })}
        title={`${session.title}\n${session.cwd ?? ""}\n${session.id}`}
        onClick={onSelect}
      >
        <span className="flex min-w-0 flex-1 flex-col gap-1">
          <span className="truncate">{session.title}</span>
          <span className="flex min-w-0 flex-wrap items-center gap-2 text-xs text-muted-foreground">
            <AgentIdentity {...(session.model ? { model: session.model } : { provider: session.source })} identityId={session.id} />
            <span className="truncate">{session.cwd?.split("/").filter(Boolean).at(-1) ?? session.id.slice(0, 12)}</span>
          </span>
        </span>
        <Badge variant="soft" tone={session.resumable ? "neutral" : "info"} clickToCopy={false}>
          {session.resumable ? "resumes" : (session.status ?? "running")}
        </Badge>
      </Button>
      {assignAll && (
        <Button size="sm" variant="outline" aria-label={`Assign all to ${session.title}`} disabled={assignAll.disabled} onClick={assignAll.onAssign}>
          Assign all
        </Button>
      )}
    </div>
  );
}

/**
 * The body of an agent menu: how sessions are grouped, the runtime new sessions
 * start with, auto-run, and the live sessions to deliver to. Data-agnostic: the
 * playground's Agent dropdown and the Chrome extension's popup and panel each
 * feed it their own sources and decide what choosing something does.
 */
export function AgentPicker(props: AgentPickerProps) {
  const { heading, busy = false, notice, error, grouping, runtime, autoRun, edits, sessions } = props;
  const [expanded, setExpanded] = useState(false);
  const inline = runtime !== undefined && runtime.onChoose === undefined;
  const newSelected = sessions?.onNew !== undefined && sessions.selectedId === undefined;
  const showBar = inline && (expanded || newSelected);
  return (
    <div className="space-y-3 p-3 text-xs">
      {heading && <h2 className="font-semibold">{heading}</h2>}
      {grouping && (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span>Sessions</span>
          <SegmentedControl
            size="sm"
            aria-label="Session grouping"
            value={grouping.value}
            options={[
              { id: "shared", label: "Shared", disabled: busy },
              { id: "page", label: "Per page", disabled: busy },
            ]}
            onChange={grouping.onChange}
          />
        </div>
      )}
      {runtime && (
        <>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span>Model</span>
            <div className="flex items-center gap-1">
              {runtime.onClear && runtime.value && (
                <Button size="sm" variant="ghost" disabled={busy} onClick={runtime.onClear}>
                  {runtime.clearLabel ?? "Clear"}
                </Button>
              )}
              <Button
                size="sm"
                variant="ghost"
                disabled={busy}
                aria-label={runtime.ariaLabel ?? "Choose model"}
                {...(inline ? { "aria-expanded": showBar } : {})}
                onClick={runtime.onChoose ?? (() => setExpanded((open) => !open))}
              >
                {runtime.value ? <RuntimeIdentity runtime={runtime.value} /> : (runtime.emptyLabel ?? "Choose model…")}
              </Button>
            </div>
          </div>
          {showBar && (
            <InlineRuntimeBar
              runtime={runtime.value}
              catalog={runtime.catalog}
              onChange={(next) => runtime.onChange?.(next)}
            />
          )}
        </>
      )}
      {autoRun && (
        <Switch
          label={autoRun.label ?? "Auto-run new comments on this page"}
          checked={autoRun.checked}
          disabled={busy || autoRun.disabled === true}
          {...(autoRun.title ? { title: autoRun.title } : {})}
          onChange={autoRun.onChange}
        />
      )}
      {edits && (
        <div className="flex flex-wrap items-center justify-between gap-2" {...(edits.title ? { title: edits.title } : {})}>
          <span>Agent edits</span>
          <SegmentedControl
            size="sm"
            aria-label="Agent edits"
            value={edits.value}
            options={[
              { id: "direct", label: "Edit directly", disabled: busy || edits.disabled === true },
              { id: "propose", label: "Propose patches", disabled: busy || edits.disabled === true },
            ]}
            onChange={edits.onChange}
          />
        </div>
      )}
      {notice}
      {error && (
        <p role="alert" className="break-words text-destructive">
          {error}
        </p>
      )}
      {sessions && (
        <>
          <hr className="border-border" />
          <h2 className="font-semibold">Active sessions</h2>
          {sessions.state === "loading" && (
            <p role="status" className="text-muted-foreground">
              Loading active sessions…
            </p>
          )}
          {sessions.state === "ready" && sessions.items.length === 0 && (
            <p className="text-muted-foreground">No active sessions.</p>
          )}
          {sessions.onNew && (
            <Button
              variant="ghost"
              size="sm"
              aria-pressed={newSelected}
              aria-label="New session"
              className="h-auto w-full justify-start gap-2 py-2 text-left"
              onClick={sessions.onNew}
            >
              <span className="truncate">New session</span>
              <span className="text-muted-foreground">runs the model above</span>
            </Button>
          )}
          {sessions.assignAll && (
            <div className="space-y-1">
              <Button
                size="sm"
                variant="outline"
                className="w-full"
                aria-label="Assign all to a new session"
                disabled={busy || sessions.assignAll.count === 0}
                onClick={sessions.assignAll.onNew}
              >
                Assign all to a new session
              </Button>
              <p className="text-muted-foreground">
                {sessions.assignAll.count === 0
                  ? "No open threads on this page"
                  : `${sessions.assignAll.count} open ${sessions.assignAll.count === 1 ? "thread" : "threads"}`}
              </p>
            </div>
          )}
          {sessions.items.map((session) => (
            <SessionRow
              key={session.id}
              session={session}
              label={sessions.selectLabel(session)}
              pressed={sessions.onNew ? sessions.selectedId === session.id : undefined}
              onSelect={() => sessions.onSelect(session)}
              assignAll={sessions.assignAll && { disabled: busy || sessions.assignAll.count === 0, onAssign: () => sessions.assignAll?.onSession(session) }}
            />
          ))}
          {sessions.onAll && (
            <Button size="sm" variant="outline" className="w-full" onClick={sessions.onAll}>
              All sessions…
            </Button>
          )}
        </>
      )}
    </div>
  );
}
