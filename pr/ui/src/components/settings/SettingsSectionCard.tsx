import { useState, type ComponentType, type ReactNode } from 'react';
import { UiChevronDown, UiChevronRight } from '@flanksource/clicky-ui/icons';

// SettingsSectionCard mirrors clicky-ui's SpecRuntimeEditor SectionCard: a
// numbered, collapsible config section separated by a top rule (no boxed card).
// The heading itself is the toggle (WAI-ARIA accordion: heading > button).
interface Props {
  icon?: ComponentType<{ className?: string }>;
  title: string;
  number: string;
  hint?: ReactNode;
  defaultCollapsed?: boolean;
  children: ReactNode;
}

export function SettingsSectionCard({ icon: Glyph, title, number, hint, defaultCollapsed = false, children }: Props) {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const Chevron = collapsed ? UiChevronRight : UiChevronDown;
  return (
    <section className="scroll-mt-4 border-t border-border py-density-4 first:border-t-0 first:pt-0">
      <header>
        <h3 className="text-base font-bold tracking-tight">
          <button
            type="button"
            aria-expanded={!collapsed}
            onClick={() => setCollapsed(c => !c)}
            className="group flex w-full items-center gap-density-2 text-left"
          >
            {Glyph && <Glyph className="size-5 shrink-0 text-muted-foreground" />}
            <span>{title}</span>
            <span className="text-[10px] font-bold tabular-nums text-muted-foreground/70">{number}</span>
            <Chevron className="ml-auto size-4 shrink-0 text-muted-foreground/70 group-hover:text-foreground" />
          </button>
        </h3>
        {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
      </header>
      {!collapsed && <div className="mt-density-3">{children}</div>}
    </section>
  );
}
