// The /todos/new?embed=1 form talks to whichever page frames it (today the
// Playground grabber extension's chrome-extension:// relay page) over
// postMessage. Messages only go to, and are only accepted from, that one parent
// window and its origin.
export const EMBED_MESSAGE_SOURCE = 'gavel-react-grab';

export type EmbedOutgoing =
  | { type: 'embed-ready' }
  | { type: 'todo-created' | 'todo-commented' | 'cancel'; ref?: string | undefined };

export type EmbedFrame = {
  parent: Pick<Window, 'postMessage'>;
  location: { ancestorOrigins?: DOMStringList };
  document: { referrer: string };
};

// embedParentOrigin resolves the framing page's origin: Chrome/Safari expose it
// as location.ancestorOrigins[0] (the only source for a chrome-extension parent,
// whose URL origin is opaque); other browsers fall back to the referrer.
export function embedParentOrigin(win: EmbedFrame): string {
  if ((win.parent as unknown) === win) throw new Error('/todos/new?embed=1 must be framed by the page that opened it');
  const ancestor = win.location.ancestorOrigins?.[0];
  const origin = ancestor ?? (win.document.referrer ? new URL(win.document.referrer).origin : '');
  if (!origin || origin === 'null') {
    throw new Error('Cannot determine the embedding parent origin: no location.ancestorOrigins or document.referrer');
  }
  return origin;
}

export function postToEmbedParent(win: EmbedFrame, origin: string, message: EmbedOutgoing) {
  win.parent.postMessage({ source: EMBED_MESSAGE_SOURCE, ...message }, origin);
}

// embedAttachment returns the screenshot the parent handed over, or null for any
// message that is not an attachment from the parent window at its origin.
export function embedAttachment(event: MessageEvent, win: EmbedFrame, origin: string): { blob: Blob; name: string } | null {
  if (event.source !== (win.parent as unknown) || event.origin !== origin) return null;
  const d = event.data as { source?: unknown; type?: unknown; blob?: unknown; name?: unknown } | null;
  if (!d || d.source !== EMBED_MESSAGE_SOURCE || d.type !== 'attachment' || !(d.blob instanceof Blob)) return null;
  return { blob: d.blob, name: typeof d.name === 'string' && d.name ? d.name : 'attachment' };
}
