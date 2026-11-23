import { useState } from "react";
import { Button, DropdownMenu } from "@flanksource/clicky-ui";
import { SpecRuntimeBar } from "@flanksource/clicky-ui/ai";
import { AgentIdentity, shortContextModelName } from "@flanksource/clicky-ui/chat";
import type { CommentMentionable } from "@flanksource/clicky-ui/comments";
import { runtimeOfSpecValue, specValueOfRuntime, type RuntimeCatalog } from "../agents/runtime-families";
import type { AgentSessionRuntime } from "../contracts/agent-session-store";
import { DraftDeliveryPreview, deliveryLabel } from "./DeliveryPreview";
import { AmpThreadField } from "./AmpThreadField";
import type { AgentSession } from "./gavel-client";

export type DraftTarget = { kind: "mention"; mention: CommentMentionable } | { kind: "run"; runtime: AgentSessionRuntime } | { kind: "session"; sessionId: string };

/** The two sessions a site can send its comments to by default; a draft may pick either one, or any other session, for itself. */
export type DefaultSessions = { global?: string; page?: string };

/**
 * "Runs on": where one draft goes. The page default is what the site's target
 * resolves to; every other entry applies to this draft only and the default is
 * left as it is.
 */
export function CommentDeliveryMenu({ target, onChange, targets, sessions, sessionIds = {}, runtime, catalog, catalogError, preview }: {
  target?: DraftTarget | undefined;
  onChange: (target: DraftTarget | undefined) => void;
  targets: CommentMentionable[];
  sessions?: AgentSession[] | undefined;
  sessionIds?: DefaultSessions | undefined;
  runtime?: AgentSessionRuntime | undefined;
  catalog?: RuntimeCatalog | undefined;
  catalogError?: string | null;
  preview: Parameters<typeof DraftDeliveryPreview>[0];
}) {
  const all = sessions ?? [];
  const titleOf = (id: string) => all.find(session => session.id === id)?.title ?? `Session ${id.slice(-6)}`;
  const others = all.filter(session => session.id !== sessionIds.global && session.id !== sessionIds.page);
  const mentions = targets.filter(mention => !all.some(session => mention.id === `session:${session.id}`));
  const sessionLabel = (id: string) => `${titleOf(id)}${all.find(session => session.id === id)?.source === "amp" ? " (Amp)" : ""}`;
  const chosen = target?.kind === "run" ? `New ${shortContextModelName(target.runtime.spec.model)}`
    : target?.kind === "mention" ? `@${target.mention.name}`
    : target?.kind === "session" ? sessionLabel(target.sessionId) : undefined;
  const chosenSession = target?.kind === "session" ? target.sessionId : undefined;
  // One row beside the comment box's send button: the target menu, then where the comment goes, truncating first.
  return <div data-comment-delivery className="flex min-w-0 items-center gap-1.5">
    <DropdownMenu label={chosen ? `Runs on ${chosen}` : "Runs on"} title={chosen ? `Runs on ${chosen}` : "Choose where this comment runs"} menuLabel="Run this comment on" role="dialog"
      size="xs" variant={target ? "outline" : "ghost"} className="shrink-0" align="left" menuClassName="w-[min(34rem,calc(100vw-1rem))] p-2">
      {close => {
        const choose = (next?: DraftTarget) => { onChange(next); close(); };
        const entry = (options: { key: string; label: string; detail?: string | undefined; pressed: boolean; disabled?: boolean; onClick: () => void }) =>
          <Button key={options.key} size="sm" variant="ghost" className="w-full justify-start gap-2" aria-pressed={options.pressed} disabled={options.disabled === true} onClick={options.onClick}>
            <span className="min-w-0 truncate">{options.label}</span>
            {options.detail && <span className="text-muted-foreground">{options.detail}</span>}
          </Button>;
        const sessionEntry = (key: string, label: string, id: string | undefined) => entry({
          key, label, ...(id ? { detail: titleOf(id) } : { detail: "none started" }),
          pressed: id !== undefined && chosenSession === id, disabled: id === undefined,
          onClick: () => { if (id) choose({ kind: "session", sessionId: id }); },
        });
        return <div className="space-y-2">
          <section className="space-y-1" aria-label="Sessions">
            {entry({ key: "default", label: "Default", detail: deliveryLabel(preview.defaultState), pressed: !target, onClick: () => choose() })}
            {sessionEntry("global", "Global session", sessionIds.global)}
            {sessionEntry("page", "This page's session", sessionIds.page)}
            {others.length > 0 && <div className="max-h-36 space-y-1 overflow-y-auto border-t border-border pt-1" aria-label="Active sessions">
              {others.map(session => <Button key={session.id} size="sm" variant="ghost" className="w-full justify-start gap-2" aria-pressed={chosenSession === session.id}
                title={`${session.title}\n${session.cwd ?? ""}\n${session.id}`} onClick={() => choose({ kind: "session", sessionId: session.id })}>
                <span className="min-w-0 truncate">{sessionLabel(session.id)}</span>
                <AgentIdentity {...(session.model ? { model: session.model } : { provider: session.source })} identityId={session.id} />
                <span className="text-muted-foreground">{session.resumable ? "resumes" : (session.status ?? "running")}</span>
              </Button>)}
            </div>}
          </section>
          <section className="space-y-2 border-y border-border py-2" aria-label="New session">
            {catalog ? <NewCommentRuntime catalog={catalog} runtime={target?.kind === "run" ? target.runtime : runtime} onChoose={value => choose({ kind: "run", runtime: value })} />
              : <p role="status" className="text-xs text-muted-foreground">{catalogError ? "Connect to Gavel to load new session runtimes." : "Loading runtimes…"}</p>}
          </section>
          {mentions.length > 0 && <div className="max-h-36 space-y-1 overflow-y-auto" aria-label="Gavel targets">
            {mentions.map(mention => <Button key={mention.id} size="sm" variant="ghost" className="w-full justify-start"
              aria-pressed={target?.kind === "mention" && target.mention.id === mention.id} onClick={() => choose({ kind: "mention", mention })}>@{mention.name}</Button>)}
          </div>}
          <AmpThreadField label="Use Amp thread" onChoose={sessionId => choose({ kind: "session", sessionId })} />
        </div>;
      }}
    </DropdownMenu>
    {target?.kind === "run" ? <span className="min-w-0 truncate text-[11px] text-muted-foreground" title="Starts a new session when the comment syncs." aria-live="polite">Starts a new session on sync</span>
      : <DraftDeliveryPreview {...preview} compact {...(target?.kind === "mention" ? { mentionId: target.mention.id } : {})}
          {...(target?.kind === "session" ? { explicitSessionId: target.sessionId } : {})} />}
  </div>;
}

function NewCommentRuntime({ catalog, runtime, onChoose }: {
  catalog: RuntimeCatalog;
  runtime: AgentSessionRuntime | undefined;
  onChoose: (runtime: AgentSessionRuntime) => void;
}) {
  const [spec, setSpec] = useState(() => specValueOfRuntime(catalog.families, runtime));
  const choice = runtimeOfSpecValue(spec, runtime?.presets);
  return <>
    <SpecRuntimeBar ariaLabel="New session runtime" value={spec} onChange={setSpec} families={catalog.families} models={catalog.models} />
    <Button size="sm" variant="outline" disabled={!choice} onClick={() => { if (choice) onChoose(choice); }}>Use new session</Button>
  </>;
}
