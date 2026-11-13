import type { AgentPin } from "@flanksource/page-runtime/schema";

type PinnedComment = { anchor?: string | null; parentId?: string | null; element?: { locators?: AgentPin["locators"] } };

export function commentPins(comments: readonly PinnedComment[]): AgentPin[] {
  const pins = new Map<string, AgentPin>();
  for (const comment of comments) {
    if (comment.parentId || !comment.anchor) continue;
    const pin = pins.get(comment.anchor) ?? { anchor: comment.anchor, count: 0, locators: [] };
    pin.count++;
    if (!pin.locators.length && comment.element?.locators?.length) pin.locators = comment.element.locators;
    pins.set(comment.anchor, pin);
  }
  return [...pins.values()];
}
