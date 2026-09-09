import type { ComponentType } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiComment, UiListDashes, UiListFlat } from '@flanksource/clicky-ui/icons';

export type TodoDetailTabKey = 'overview' | 'verification' | 'session' | 'plan';

export interface TodoDetailBadge {
  count: number;
  /** Tints the badge red — the last verification attempt did not pass. */
  failing: boolean;
  title: string;
}

export function TodoDetailTabs({
  tab,
  onSelect,
  verification,
}: {
  tab: TodoDetailTabKey;
  onSelect: (tab: TodoDetailTabKey) => void;
  verification: TodoDetailBadge;
}) {
  return (
    <div className="flex shrink-0 flex-nowrap gap-1 overflow-x-auto border-b border-border bg-background px-4 pt-2">
      <DetailTab active={tab === 'overview'} onClick={() => onSelect('overview')} icon={UiListFlat} label="Overview" />
      <DetailTab
        active={tab === 'verification'}
        onClick={() => onSelect('verification')}
        icon={UiListDashes}
        label="Verification"
        count={verification.count}
        tone={verification.failing ? 'danger' : 'default'}
        title={verification.title}
      />
      <DetailTab active={tab === 'session'} onClick={() => onSelect('session')} icon={UiComment} label="Session" />
      <DetailTab active={tab === 'plan'} onClick={() => onSelect('plan')} icon={UiListDashes} label="Plan" />
    </div>
  );
}

export function DetailTab({
  active,
  onClick,
  icon: Icon,
  label,
  count,
  tone = 'default',
  title,
}: {
  active: boolean;
  onClick: () => void;
  icon: ComponentType<IconProps>;
  label: string;
  count?: number;
  tone?: 'default' | 'danger';
  title?: string;
}) {
  return (
    <Button
      variant="ghost"
      type="button"
      onClick={onClick}
      aria-pressed={active}
      title={title}
      className={`-mb-px inline-flex h-auto shrink-0 items-center gap-1.5 border-b-2 px-2.5 py-1.5 text-xs font-medium transition-colors ${
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      }`}
    >
      <Icon className="text-sm" />
      {label}
      {typeof count === 'number' && count > 0 && (
        <span
          className={`ml-0.5 inline-flex min-w-[1.25rem] items-center justify-center rounded-full border px-1.5 py-0.5 text-[10px] tabular-nums ${
            tone === 'danger'
              ? 'border-red-500/25 bg-red-500/10 text-red-700 dark:text-red-400'
              : 'border-border bg-background text-muted-foreground'
          }`}
        >
          {count}
        </span>
      )}
    </Button>
  );
}
