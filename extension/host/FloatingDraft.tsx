import { useEffect, useRef, useState } from "react";
import { Button, useCommentContext } from "@flanksource/clicky-ui";
import { CommentDraftCard, type CommentConfig } from "@flanksource/clicky-ui/comments";
import type { DraftBox } from "./bridge-parse";
import type { Selection } from "./model";
import { sendToHost } from "./useHostBridge";

/**
 * The comment card inside the draft frame. While one of its menus is open the
 * host widens the frame to the whole viewport (so the menu is not clipped to
 * the card) and reports `box`, the card's page position, to keep it in place.
 */
export function FloatingDraft({ selection, config, onDiscard, box }: { selection: Selection; config: CommentConfig; onDiscard: () => void; box?: DraftBox | undefined }) {
  const ctx = useCommentContext();
  const ref = useRef<HTMLDivElement>(null);
  const [overlay, setOverlay] = useState(false);
  useEffect(() => {
    document.body.toggleAttribute("data-floating-draft", true);
    return () => document.body.removeAttribute("data-floating-draft");
  }, []);
  useEffect(() => {
    const content = ref.current;
    if (!content) return;
    let pending = 0;
    const measure = () => {
      cancelAnimationFrame(pending);
      pending = requestAnimationFrame(() => {
        const open = !!document.querySelector("[data-floating-ui-portal]");
        setOverlay(open);
        sendToHost({ type: "draft-size", height: content.getBoundingClientRect().height + 16, overlay: open });
      });
    };
    const observer = new ResizeObserver(measure);
    observer.observe(content);
    const menus = new MutationObserver(measure);
    menus.observe(document.body, { childList: true, subtree: true });
    measure();
    return () => {
      observer.disconnect();
      menus.disconnect();
      cancelAnimationFrame(pending);
    };
  }, []);
  return (
    <div
      ref={ref}
      role="dialog"
      aria-label={`New comment on ${selection.label}`}
      className="rounded-lg border border-primary/50 bg-card text-foreground"
      style={overlay && box ? { position: "fixed", left: box.left, top: box.top, width: box.width } : undefined}
    >
      <div className="flex justify-end">
        <Button
          size="sm"
          variant="ghost"
          onClick={() => sendToHost({ type: "minimize" })}
        >
          Minimize
        </Button>
      </div>
      <CommentDraftCard
        key={selection.anchor}
        label={selection.label}
        anchor={selection.anchor}
        config={config}
        onDiscard={() => {
          onDiscard();
          sendToHost({ type: "draft", active: false });
        }}
        createActions={[
          {
            id: "plain",
            label: selection.element.screenshot
              ? "Post with screenshot"
              : "Post comment",
          },
        ]}
        onCreate={async (input) => {
          if (!ctx.callbacks.onCreate)
            throw new Error("Comment creation is not configured");
          await ctx.callbacks.onCreate(input);
          sendToHost({ type: "draft", active: false, posted: true });
          ctx.focusAnchor(selection.anchor);
        }}
      />
    </div>
  );
}
