import { useEffect } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiArrowLeft, UiArrowRight } from '@flanksource/clicky-ui/icons';

export interface TodoNavigationControlsProps {
  position: number;
  total: number;
  canPrevious: boolean;
  canNext: boolean;
  onPrevious: () => void;
  onNext: () => void;
}

function isEditing(target: EventTarget | null): boolean {
  return target instanceof Element && !!target.closest('input, textarea, select, [contenteditable="true"], [role="textbox"]');
}

function hasOpenOverlay(): boolean {
  return !!document.querySelector([
    '[role="dialog"]:not([aria-hidden="true"])',
    '[role="menu"]:not([aria-hidden="true"])',
    '[role="listbox"]:not([aria-hidden="true"])',
    '[data-state="open"][data-radix-popper-content-wrapper]',
  ].join(','));
}

export function TodoNavigationControls({
  position,
  total,
  canPrevious,
  canNext,
  onPrevious,
  onNext,
}: TodoNavigationControlsProps) {
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.repeat || event.metaKey || event.ctrlKey || event.altKey || event.shiftKey) return;
      if (isEditing(event.target) || hasOpenOverlay()) return;
      const key = event.key.toLowerCase();
      if ((key === 'j' || key === 'n') && canNext) {
        event.preventDefault();
        onNext();
      } else if ((key === 'k' || key === 'p') && canPrevious) {
        event.preventDefault();
        onPrevious();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [canNext, canPrevious, onNext, onPrevious]);

  return (
    <div className="inline-flex shrink-0 items-center gap-0.5" aria-label={`Todo ${position} of ${total}`}>
      <Button
        variant="ghost"
        size="icon"
        type="button"
        onClick={onPrevious}
        disabled={!canPrevious}
        title="Previous todo (K or P)"
        aria-label="Previous todo"
        aria-keyshortcuts="K P"
        className="h-8 w-8 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-40"
      >
        <UiArrowLeft className="text-sm" />
      </Button>
      <span className="min-w-10 text-center text-xs tabular-nums text-muted-foreground" aria-hidden="true">
        {position}/{total}
      </span>
      <Button
        variant="ghost"
        size="icon"
        type="button"
        onClick={onNext}
        disabled={!canNext}
        title="Next todo (J or N)"
        aria-label="Next todo"
        aria-keyshortcuts="J N"
        className="h-8 w-8 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-40"
      >
        <UiArrowRight className="text-sm" />
      </Button>
    </div>
  );
}
