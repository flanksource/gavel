import type { ReactNode } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiArrowLeft } from '@flanksource/clicky-ui/icons';

export function MobileProjectsView({ selected, reviewing, onBack, sidebar, children }: {
  selected: string;
  reviewing: boolean;
  onBack: () => void;
  sidebar: ReactNode;
  children: ReactNode;
}) {
  if (!selected) return <div className="h-full min-w-0">{sidebar}</div>;
  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col">
      {!reviewing && <div className="shrink-0 border-b border-border px-2">
        <Button type="button" variant="ghost" aria-label="Back to projects" onClick={onBack} className="min-h-11 gap-2 px-2"><UiArrowLeft />Projects</Button>
      </div>}
      <div className="min-h-0 min-w-0 flex-1">{children}</div>
    </div>
  );
}
