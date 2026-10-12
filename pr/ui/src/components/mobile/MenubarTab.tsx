import { useEffect, useRef } from 'react';
import type { ComponentType } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { IconProps } from '@flanksource/clicky-ui/icons';

// MenubarTab is one segment of the menubar's Processes/PRs switcher. A status
// dot (process health) or an icon leads the label; count is the inline subtotal
// and badge is an attention-grabbing count (e.g. failing PRs).
export function MenubarTab({ label, icon: Icon, dot, count, badge, active, onClick }: {
  label: string;
  icon?: ComponentType<IconProps>;
  dot?: string;
  count?: number | string;
  badge?: number;
  active: boolean;
  onClick: () => void;
}) {
  // The strip scrolls sideways on narrow screens; keep the open tab in view
  // (e.g. Sessions, the last one, when a deep link opens it).
  const ref = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    // jsdom does not implement scrollIntoView.
    if (active) ref.current?.scrollIntoView?.({ block: 'nearest', inline: 'nearest' });
  }, [active]);
  return (
    <Button
      ref={ref}
      variant="ghost"
      type="button"
      onClick={onClick}
      className={`flex shrink-0 items-center gap-1.5 rounded-md px-2 py-1 text-xs transition h-auto justify-start ${
        active ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'
      }`}
    >
      {dot ? <span className={`inline-block h-2 w-2 rounded-full ${dot}`} /> : Icon ? <Icon /> : null}
      <span>{label}</span>
      {count !== undefined && count !== '' && (
        <span className="tabular-nums text-[10px] text-muted-foreground">{count}</span>
      )}
      {badge !== undefined && badge > 0 && (
        <span className="rounded-full bg-red-500/15 px-1 text-[10px] font-medium tabular-nums text-red-600 dark:text-red-400">{badge}</span>
      )}
    </Button>
  );
}
