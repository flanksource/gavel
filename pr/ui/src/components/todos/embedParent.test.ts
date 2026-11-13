import { describe, expect, it, vi } from 'vitest';
import { EMBED_MESSAGE_SOURCE, embedAttachment, embedParentOrigin, postToEmbedParent, type EmbedFrame } from './embedParent';

const extensionOrigin = 'chrome-extension://abcdefghijklmnopabcdefghijklmnop';
const pageOrigin = 'http://localhost:5173';

function frame({ ancestors, referrer = '' }: { ancestors?: string[]; referrer?: string }): EmbedFrame & { parent: { postMessage: ReturnType<typeof vi.fn> } } {
  const parent = { postMessage: vi.fn() };
  return {
    parent,
    location: ancestors ? { ancestorOrigins: ancestors as unknown as DOMStringList } : {},
    document: { referrer },
  } as unknown as EmbedFrame & { parent: typeof parent };
}

function message(data: unknown, source: unknown, origin: string): MessageEvent {
  return { data, source, origin } as MessageEvent;
}

describe('embedParentOrigin', () => {
  it('uses the immediate parent from location.ancestorOrigins, including a chrome-extension relay page', () => {
    expect(embedParentOrigin(frame({ ancestors: [extensionOrigin, 'https://app.example.com'], referrer: `${pageOrigin}/x` }))).toBe(extensionOrigin);
  });

  it('falls back to the referrer origin when ancestorOrigins is unavailable', () => {
    expect(embedParentOrigin(frame({ referrer: `${pageOrigin}/checkout?a=1` }))).toBe(pageOrigin);
  });

  it('fails when the page is not framed', () => {
    const win = frame({ ancestors: [pageOrigin] });
    (win as { parent: unknown }).parent = win;
    expect(() => embedParentOrigin(win)).toThrow(/must be framed/);
  });

  it.each([
    ['no ancestor and no referrer', frame({})],
    ['an opaque ancestor origin', frame({ ancestors: ['null'] })],
  ])('fails with %s', (_, win) => {
    expect(() => embedParentOrigin(win)).toThrow(/parent origin/);
  });
});

describe('postToEmbedParent', () => {
  it('tags the message and targets only the resolved parent origin', () => {
    const win = frame({ ancestors: [extensionOrigin] });
    postToEmbedParent(win, extensionOrigin, { type: 'todo-created', ref: 't-1' });
    expect(win.parent.postMessage).toHaveBeenCalledWith({ source: EMBED_MESSAGE_SOURCE, type: 'todo-created', ref: 't-1' }, extensionOrigin);
  });
});

describe('embedAttachment', () => {
  const blob = new Blob(['png']);
  const data = { source: EMBED_MESSAGE_SOURCE, type: 'attachment', blob, name: 'shot.png' };

  it('accepts an attachment posted by the parent window from the parent origin', () => {
    const win = frame({ ancestors: [extensionOrigin] });
    expect(embedAttachment(message(data, win.parent, extensionOrigin), win, extensionOrigin)).toEqual({ blob, name: 'shot.png' });
  });

  it.each([
    ['another window', (win: EmbedFrame) => ({ win, source: {}, origin: extensionOrigin, data })],
    ['the parent window under a different origin', (win: EmbedFrame) => ({ win, source: win.parent, origin: pageOrigin, data })],
    ['a foreign message source tag', (win: EmbedFrame) => ({ win, source: win.parent, origin: extensionOrigin, data: { ...data, source: 'other' } })],
    ['a missing blob', (win: EmbedFrame) => ({ win, source: win.parent, origin: extensionOrigin, data: { ...data, blob: undefined } })],
  ])('ignores a message from %s', (_, build) => {
    const { win, source, origin, data: payload } = build(frame({ ancestors: [extensionOrigin] }));
    expect(embedAttachment(message(payload, source, origin), win, extensionOrigin)).toBeNull();
  });
});
