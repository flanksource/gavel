import { useEffect, useRef } from "react";
import { Button, useCommentContext } from "@flanksource/clicky-ui";
import { CommentDraftCard } from "@flanksource/clicky-ui/comments";
import type { Selection } from "./model";
import { sendToHost } from "./useHostBridge";

export function FloatingDraft({ selection }: { selection: Selection }) {
  const ctx = useCommentContext();
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const content = ref.current;
    if (!content) return;
    let pending = 0;
    const measure = () => {
      cancelAnimationFrame(pending);
      pending = requestAnimationFrame(() => {
        const height = content.getBoundingClientRect().height + 16;
        sendToHost({
          type: "draft-size",
          height: document.querySelector("[data-floating-ui-portal]")
            ? Math.max(height, 480)
            : height,
        });
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
        config={ctx.config}
        onDiscard={() => sendToHost({ type: "draft", active: false })}
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
