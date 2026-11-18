import { useEffect, useState } from "react";
import { AgentIdentity } from "@flanksource/clicky-ui/chat";
import type { CommentMentionable } from "@flanksource/clicky-ui/comments";
import type { CreateDeliveryChoice, CreateDeliveryPreview } from "../comments/gavel-delivery";
import type { GavelClient, PageRef } from "./gavel-client";

export type DeliveryPreviewState = { preview?: CreateDeliveryPreview; error?: string };

export function useDeliveryMentionables(client: GavelClient, page: PageRef | undefined, refresh: unknown) {
  const [state, setState] = useState<{ url: string; items: CommentMentionable[]; error?: string }>();
  const url = page?.url;
  const title = page?.title;
  useEffect(() => {
    if (!url || title === undefined) return;
    let current = true;
    client.mentionables({ url, title }).then(
      items => { if (current) setState({ url, items }); },
      (cause: unknown) => {
        if (current) setState({ url, items: [], error: cause instanceof Error ? cause.message : String(cause) });
      },
    );
    return () => { current = false; };
  }, [client, url, title, refresh]);
  return state?.url === url ? state : undefined;
}

export function useDeliveryPreview(options: {
  client: GavelClient;
  page: PageRef | undefined;
  sessionId: string | undefined;
  revision: unknown;
  body?: string;
  mentionId?: string;
  signal?: unknown;
}): DeliveryPreviewState {
  const { client, page, sessionId, revision, body, mentionId, signal } = options;
  const url = page?.url;
  const title = page?.title;
  const [state, setState] = useState<DeliveryPreviewState & { key: string }>();
  const key = JSON.stringify([url, sessionId, body, mentionId, revision]);
  useEffect(() => {
    if (!url || title === undefined) return;
    let current = true;
    const timer = window.setTimeout(() => {
      const delivery: CreateDeliveryChoice = { ...(sessionId ? { defaultSessionId: sessionId } : {}), ...(mentionId ? { mentionId } : {}) };
      client.previewDelivery({ page: { url, title }, delivery, ...(body ? { body } : {}) }).then(
        preview => { if (current) setState({ key, preview }); },
        (cause: unknown) => { if (current) setState({ key, error: cause instanceof Error ? cause.message : String(cause) }); },
      );
    }, body ? 250 : 0);
    return () => { current = false; window.clearTimeout(timer); };
  }, [client, key, url, title, sessionId, body, mentionId, signal]);
  return state?.key === key ? state : {};
}

export function deliveryLabel(state: DeliveryPreviewState): string {
  const { preview, error } = state;
  if (error) return "Target unavailable";
  if (!preview) return "Resolving target…";
  if (preview.session) return preview.session.title;
  const target = preview.route?.target;
  if (!target) return "Post only";
  if (target.kind === "session") return `Unavailable session ${target.sessionId.slice(-6)}`;
  const model = target.runOptions?.spec?.["model"];
  return `New session · ${typeof model === "string" ? model : target.prompt}`;
}

export function DeliveryTargetLine({ state }: { state: DeliveryPreviewState }) {
  const { preview, error } = state;
  const session = preview?.session;
  const source = preview?.route?.source;
  const rule = source && source !== "session" && source !== "remembered-session" && !source.startsWith("mention:") ? source : undefined;
  return (
    <div className="mt-1 flex flex-wrap items-center gap-x-1.5 text-[11px] leading-4 text-muted-foreground" aria-live="polite">
      {error ? <span role="alert">{error}</span> : session ? (
        <>
          <AgentIdentity label={session.title} {...(session.source ? { provider: session.source } : {})} {...(session.model ? { model: session.model } : {})} />
          <code title={session.id}>{session.id.slice(-6)}</code>
          {session.resumable && <span>Resumes</span>}
        </>
      ) : <span>{deliveryLabel(state)}</span>}
      {rule && <span>Rule from {rule}</span>}
    </div>
  );
}

export function DraftDeliveryPreview({ body, mentionId, defaultState, ...options }: {
  client: GavelClient;
  page: PageRef | undefined;
  sessionId: string | undefined;
  revision: unknown;
  body: string;
  mentionId?: string;
  defaultState: DeliveryPreviewState;
}) {
  const override = mentionId !== undefined || body.includes("@");
  const draft = useDeliveryPreview({
    ...options,
    page: override ? options.page : undefined,
    ...(override ? { body } : {}),
    ...(mentionId ? { mentionId } : {}),
  });
  return <DeliveryTargetLine state={override ? draft : defaultState} />;
}
