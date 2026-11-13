import { Button, DOCUMENT_ANCHOR, type CommentConfig, type CommentCreateInput } from "@flanksource/clicky-ui";
import { CommentComposer } from "@flanksource/clicky-ui/comments";
import { UiCamera, UiCursorClick } from "@flanksource/clicky-ui/icons";

export type PanelComposerProps = {
  config: CommentConfig;
  /** Posts a whole-page comment; a rejection keeps the draft in the composer and shows the error there. */
  onPost: (input: CommentCreateInput) => Promise<void>;
  /** Starts an element comment, drafted beside the element the user picks. */
  onPick: (options: { screenshot: boolean }) => void;
};

/** The panel's fixed footer: a whole-page composer, and the way into commenting on one element. */
export function PanelComposer({ config, onPost, onPick }: PanelComposerProps) {
  return (
    <footer data-testid="panel-composer" className="shrink-0 space-y-1.5 border-t border-border bg-background p-2">
      <div className="flex items-center gap-1">
        <span className="flex-1 text-xs text-muted-foreground">Whole page</span>
        <Button size="xs" variant="ghost" onClick={() => onPick({ screenshot: false })} title="Pick an element on the page to comment on">
          <UiCursorClick className="size-3.5" aria-hidden /> Element
        </Button>
        <Button size="xs" variant="ghost" onClick={() => onPick({ screenshot: true })} title="Pick an element and attach a screenshot of it">
          <UiCamera className="size-3.5" aria-hidden /> Screenshot
        </Button>
      </div>
      <CommentComposer
        config={config}
        anchor={DOCUMENT_ANCHOR}
        collapsible={false}
        placeholder="Comment on this page…"
        createActions={[{ id: "plain", label: "Post comment" }]}
        onCreate={input => onPost(input)}
      />
    </footer>
  );
}
