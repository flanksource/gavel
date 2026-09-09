import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck } from '@flanksource/clicky-ui/icons';

// SaveBar is the persistent sticky footer (design: SpecRuntimeEditor Footer):
// a status on the left, Discard / Save on the right. Backdrop-blurred so content
// scrolls under it.
interface Props {
  path: string;
  dirty: boolean;
  saving: boolean;
  onDiscard: () => void;
  onSave: () => void;
}

export function SaveBar({ path, dirty, saving, onDiscard, onSave }: Props) {
  return (
    <div className="sticky bottom-0 z-10 flex items-center justify-between gap-density-3 border-t border-border bg-background/90 px-density-4 py-density-2 backdrop-blur">
      {dirty ? (
        <span className="inline-flex items-center gap-2 text-xs text-foreground">
          <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
          Unsaved changes
          {path && <span className="font-mono text-[11px] text-muted-foreground">{path}</span>}
        </span>
      ) : (
        <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
          <UiCheck className="size-4" />
          Up to date
        </span>
      )}
      <div className="flex items-center gap-density-2">
        <Button variant="outline" size="sm" onClick={onDiscard} disabled={!dirty || saving}>
          Discard
        </Button>
        <Button size="sm" onClick={onSave} loading={saving} disabled={!dirty}>
          Save changes
        </Button>
      </div>
    </div>
  );
}
