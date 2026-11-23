import { useId, useState, type ReactNode } from "react";
import { Button, DropdownMenu, Field, InputField, SegmentedControl, Switch, type DropdownMenuItem } from "@flanksource/clicky-ui";
import { AgentIdentity } from "@flanksource/clicky-ui/chat";

import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import type { SessionTarget } from "../host/site-settings";
import { InlineRuntimeBar, type AgentPickerSession } from "./AgentPicker";
import type { RuntimeCatalog } from "./runtime-families";
import { SessionBadge } from "./SessionBadge";

export type PinnedTarget = Exclude<SessionTarget, "auto">;
export type StartSessionRequest = { target: PinnedTarget; prefix: string; runtime: AgentSessionRuntime | undefined };
export type UseSessionRequest = { target: PinnedTarget; session: AgentPickerSession };

export type EnableCardProps = {
  /** The page's host, port included; the page agent's session name starts as `gavel-${host}`. */
  host: string;
  enabled: {
    value: boolean;
    onChange: (enabled: boolean) => void;
    disabled?: boolean | undefined;
    /** Why the switch is disabled; its tooltip. */
    disabledReason?: string | undefined;
  };
  /** Header controls right of the switch (the popup's Settings button). */
  actions?: ReactNode;
  /** Rendered between the header and the routes (the popup's Project row). */
  context?: ReactNode;
  /** Shown in place of the context, routes and footnote while the site cannot route comments yet (first-run setup, unreachable dashboard). */
  setup?: ReactNode;
  /** Where this site's comments go by default. */
  target: { value: SessionTarget; onChange: (target: SessionTarget) => void };
  /** The model a started session runs on; unset is gavel's default. */
  runtime: { value: AgentSessionRuntime | undefined; catalog: RuntimeCatalog | undefined; onChange: (runtime: AgentSessionRuntime) => void };
  sessions: {
    /** The ids of the sessions each pinned target currently sends comments to. */
    page?: string | undefined;
    global?: string | undefined;
    /** The sessions gavel offers to choose instead of starting one. */
    items: AgentPickerSession[];
    state: "loading" | "ready" | "error";
  };
  /** How Gavel groups the sessions it starts itself; edited in the Automatic route. Omitted where it cannot be changed (the panel). `menu` is the host's "This site" menu, shown beside the control. */
  grouping?: { value: AgentSessionGrouping; onChange: (grouping: AgentSessionGrouping) => void; menu?: ReactNode } | undefined;
  /** Starts a session named from `prefix` and makes it the target's session; a rejection is shown inline. */
  onStart: (request: StartSessionRequest) => Promise<void>;
  /** Makes an already running session the target's session; a rejection is shown inline. */
  onUse: (request: UseSessionRequest) => Promise<void>;
  /** Why a session cannot be started yet (for example no project is chosen); it disables Start and is shown beside it. */
  startBlocked?: string | undefined;
  status?: ReactNode;
  error?: string | null | undefined;
  /** Rendered last, below a divider (the popup's sync status; the panel's agent-edits control). */
  footer?: ReactNode;
};

const ROUTES: Record<SessionTarget, { label: string; whenOn: string }> = {
  page: { label: "This site's agent", whenOn: "this site's agent" },
  global: { label: "Global session", whenOn: "the global session" },
  auto: { label: "Automatic", whenOn: "a session Gavel starts for you" },
};
const GROUPING_SUMMARY: Record<AgentSessionGrouping, string> = {
  page: "Gavel starts a session for each page",
  shared: "Gavel starts one session for all pages",
};
const muted = "text-muted-foreground";

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

function defaultName(target: PinnedTarget, host: string): string {
  return target === "global" ? "gavel-global" : `gavel-${host}`;
}

function SessionIdentity({ session }: { session: AgentPickerSession }) {
  return <AgentIdentity {...(session.model ? { model: session.model } : { provider: session.source })} identityId={session.id} />;
}

function sessionItems(options: { items: AgentPickerSession[]; state: EnableCardProps["sessions"]["state"]; except: string | undefined; onSelect: (session: AgentPickerSession) => void }): { items: DropdownMenuItem[]; footer: ReactNode } {
  const { items, state, except, onSelect } = options;
  const choices = items.filter(session => session.id !== except);
  return {
    items: choices.map((session): DropdownMenuItem => ({
      label: <span className="flex min-w-0 items-center gap-2">
        <span className="min-w-0 truncate">{session.title}</span>
        <SessionIdentity session={session} />
        <span className={muted}>{session.resumable ? "resumes" : (session.status ?? "running")}</span>
      </span>,
      title: `${session.title}\n${session.cwd ?? ""}\n${session.id}`,
      group: "Active sessions",
      onSelect: () => onSelect(session),
    })),
    footer: state === "loading" ? <span role="status" className={`text-xs ${muted}`}>Loading active sessions…</span>
      : state === "ready" && choices.length === 0 ? <span className={`text-xs ${muted}`}>No active sessions.</span> : null,
  };
}

type RouteRowProps = { group: string; route: SessionTarget; selected: boolean; disabled: boolean; onSelect: () => void; children?: ReactNode };

function RouteRow({ group, route, selected, disabled, onSelect, children }: RouteRowProps) {
  return (
    <div data-route={route} className={`p-3 ${selected ? "bg-primary/10" : ""}`}>
      <label className="flex cursor-pointer items-center gap-2.5 font-semibold">
        <input type="radio" name={group} className="size-4 shrink-0 accent-primary" checked={selected} disabled={disabled} onChange={onSelect} />
        <span>{ROUTES[route].label}</span>
      </label>
      {children && <div className="ml-[26px] mt-1 space-y-2.5">{children}</div>}
    </div>
  );
}

/** A remembered id Gavel no longer lists is named by its last six characters. */
function sessionTitle(sessionId: string, items: AgentPickerSession[]): string {
  return items.find(session => session.id === sessionId)?.title ?? `Session ${sessionId.slice(-6)} (not live)`;
}

/** One muted line under an unselected session route: the session and its model, or that none was started. */
function SessionSummary({ sessionId, items }: { sessionId: string | undefined; items: AgentPickerSession[] }) {
  if (!sessionId) return <p className={muted}>Not started</p>;
  const current = items.find(session => session.id === sessionId);
  return (
    <p className={`flex min-w-0 items-center gap-1.5 ${muted}`}>
      <SessionBadge title={sessionTitle(sessionId, items)} />
      {current && <><span aria-hidden="true">·</span><SessionIdentity session={current} /></>}
    </p>
  );
}

type SessionMenu = { idle: boolean; choices: ReturnType<typeof sessionItems> };

function SwitchMenu({ label, menuLabel, variant, menu }: { label: string; menuLabel: string; variant: "outline" | "ghost"; menu: SessionMenu }) {
  return <DropdownMenu label={label} menuLabel={menuLabel} size="sm" variant={variant} align="left" items={menu.choices.items} {...(menu.choices.footer ? { footer: menu.choices.footer } : {})} />;
}

function CurrentSession({ session, sessionId, items, menu, onReplace }: { session: AgentPickerSession | undefined; sessionId: string; items: AgentPickerSession[]; menu: SessionMenu; onReplace: () => void }) {
  return (
    <>
      <SessionSummary sessionId={sessionId} items={items} />
      {session && <p className={muted}>{session.resumable ? "Resumes on the next comment" : (session.status ?? "running")}</p>}
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" disabled={!menu.idle} onClick={onReplace}>New session</Button>
        <SwitchMenu label="Switch" menuLabel="Switch session" variant="outline" menu={menu} />
      </div>
    </>
  );
}

type NewSessionFormProps = {
  runtime: EnableCardProps["runtime"];
  startBlocked: string | undefined;
  name: string;
  onName: (name: string) => void;
  /** The title of the session being replaced, when replacing one. */
  replacing: string | undefined;
  menu: SessionMenu;
  onStart: () => void;
  onCancel: () => void;
};

function NewSessionForm({ runtime, startBlocked, name, onName, replacing, menu, onStart, onCancel }: NewSessionFormProps) {
  const id = useId();
  return (
    <>
      <Field label="New session name" htmlFor={id}>
        <InputField id={id} value={name} onChange={onName} />
      </Field>
      <p className={`-mt-1 ${muted}`}>Gavel adds a 6-character suffix.</p>
      <InlineRuntimeBar runtime={runtime.value} catalog={runtime.catalog} onChange={runtime.onChange} />
      {startBlocked && <p className={muted}>{startBlocked}</p>}
      <div className="flex flex-wrap items-center gap-2">
        <Button size="sm" disabled={!menu.idle || name.trim() === "" || startBlocked !== undefined} onClick={onStart}>Start session</Button>
        {replacing !== undefined && <Button size="sm" variant="outline" disabled={!menu.idle} onClick={onCancel}>Cancel</Button>}
      </div>
      <SwitchMenu label="use a running session" menuLabel="Use a running session" variant="ghost" menu={menu} />
      {replacing !== undefined && <p className={`flex flex-wrap items-center gap-1 ${muted}`}>Until it starts, comments keep going to <SessionBadge title={replacing} /></p>}
    </>
  );
}

type Run = (label: string, work: () => Promise<void>) => Promise<void>;
type RouteContext = { card: EnableCardProps; group: string; idle: boolean; run: Run };

/** The page's or the global session route; it keeps its own form state while another route is selected. */
function SessionRoute({ forTarget, route }: { forTarget: PinnedTarget; route: RouteContext }) {
  const { card, group, idle, run } = route;
  const { host, target, sessions, runtime, startBlocked } = card;
  const sessionId = sessions[forTarget];
  const [name, setName] = useState<string>();
  const [replacing, setReplacing] = useState(false);
  const row = { group, route: forTarget, selected: target.value === forTarget, disabled: !idle, onSelect: () => target.onChange(forTarget) };
  if (!row.selected) return <RouteRow {...row}><SessionSummary sessionId={sessionId} items={sessions.items} /></RouteRow>;

  const act = (label: string, work: () => Promise<void>) => void run(label, async () => {
    await work();
    setReplacing(false);
  });
  const menu: SessionMenu = {
    idle,
    choices: sessionItems({ items: sessions.items, state: sessions.state, except: sessionId, onSelect: session => act("Switching session…", () => card.onUse({ target: forTarget, session })) }),
  };
  const current = sessionId ? sessions.items.find(session => session.id === sessionId) : undefined;
  if (sessionId && !replacing) {
    return <RouteRow {...row}><CurrentSession session={current} sessionId={sessionId} items={sessions.items} menu={menu} onReplace={() => setReplacing(true)} /></RouteRow>;
  }
  const value = name ?? defaultName(forTarget, host);
  return (
    <RouteRow {...row}>
      <NewSessionForm
        runtime={runtime}
        startBlocked={startBlocked}
        name={value}
        onName={setName}
        replacing={replacing && sessionId ? sessionTitle(sessionId, sessions.items) : undefined}
        menu={menu}
        onStart={() => act("Starting session…", () => card.onStart({ target: forTarget, prefix: value.trim(), runtime: runtime.value }))}
        onCancel={() => setReplacing(false)}
      />
    </RouteRow>
  );
}

function AutomaticDetail({ grouping }: { grouping: EnableCardProps["grouping"] }) {
  return (
    <>
      <p className={muted}>Gavel starts sessions for you, on the default model.</p>
      {grouping && (
        <div className="flex items-center justify-between gap-2">
          <span>One session for</span>
          <span className="flex items-center gap-2">
            <SegmentedControl
              size="sm"
              aria-label="One session for"
              value={grouping.value}
              options={[{ id: "page", label: "Each page" }, { id: "shared", label: "All pages" }]}
              onChange={grouping.onChange}
            />
            {grouping.menu}
          </span>
        </div>
      )}
    </>
  );
}

function Routes({ route }: { route: RouteContext }) {
  const { card, group, idle } = route;
  const { target, grouping } = card;
  const selected = target.value === "auto";
  return (
    <div className="divide-y divide-border overflow-hidden rounded-lg border border-border">
      <SessionRoute forTarget="page" route={route} />
      <SessionRoute forTarget="global" route={route} />
      <RouteRow group={group} route="auto" selected={selected} disabled={!idle} onSelect={() => target.onChange("auto")}>
        {selected ? <AutomaticDetail grouping={grouping} /> : <p className={muted}>{grouping ? GROUPING_SUMMARY[grouping.value] : "Gavel starts sessions for you"}</p>}
      </RouteRow>
    </div>
  );
}

function OffBody({ target }: { target: SessionTarget }) {
  return (
    <div className="space-y-1 p-4">
      <p className="font-semibold">Gavel is off on this site</p>
      <p className={muted}>Turn it on to pick elements on the page and send comments to an agent.</p>
      <p className={`!mt-3 rounded-lg bg-muted px-3 py-2 ${muted}`}>When on, comments go to {ROUTES[target].whenOn}.</p>
    </div>
  );
}

/**
 * Where comments from this site go, in one place: switch Gavel on for the host,
 * then pick this site's agent, the global session or Gavel's own routing, each row
 * naming its session. Data-agnostic like `AgentPicker`: the popup and the panel's
 * settings menu each feed it their own sources and decide what starting does.
 */
export function EnableCard(props: EnableCardProps) {
  const { host, enabled, actions, context, setup, error, status, footer } = props;
  const group = useId();
  const [busy, setBusy] = useState("");
  const [failure, setFailure] = useState("");

  const run: Run = async (label, work) => {
    setBusy(label);
    setFailure("");
    try {
      await work();
    } catch (cause) {
      setFailure(describe(cause));
    } finally {
      setBusy("");
    }
  };
  const message = error || failure;

  return (
    <div className="flex flex-col text-xs">
      <header className="flex h-[52px] items-center gap-2 border-b border-border pl-4 pr-2.5">
        <span className="min-w-0 flex-1 truncate text-sm font-semibold" title={host}>{host}</span>
        <span className={muted}>{enabled.value ? "On" : "Off"}</span>
        <span title={enabled.disabled ? enabled.disabledReason : undefined}>
          <Switch aria-label={`Gavel on ${host}`} checked={enabled.value} disabled={enabled.disabled === true} onChange={enabled.onChange} />
        </span>
        {actions}
      </header>
      {setup ? setup : enabled.value ? (
        <>
          {context}
          <div className="px-4 py-3">
            <fieldset className="m-0 min-w-0 border-0 p-0">
              <legend className={`mb-2 p-0 ${muted}`}>Comments from this site go to</legend>
              <Routes route={{ card: props, group, idle: busy === "", run }} />
            </fieldset>
            <p className={`mt-2 ${muted}`}>This is only the default. Each comment can still pick another session from its Runs on menu.</p>
          </div>
        </>
      ) : <OffBody target={props.target.value} />}
      {(busy || message || status) && (
        <div className="space-y-1 px-4 pb-3">
          {busy && <p role="status" className={muted}>{busy}</p>}
          {!busy && status && <p role="status" className={muted}>{status}</p>}
          {message && <p role="alert" className="break-words text-destructive">{message}</p>}
        </div>
      )}
      {footer && <footer className="border-t border-border px-4 py-2">{footer}</footer>}
    </div>
  );
}
