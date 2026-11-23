import { Button, DropdownMenu } from "@flanksource/clicky-ui";
import { UiChevronDown } from "@flanksource/clicky-ui/icons";

export type ScopeMenuProps = {
  host: string;
  /** What is stored for this site only (a project name, a dashboard host, a model). */
  valueLabel: string;
  /** What the other sites use; shown beside "Use the default here". */
  defaultLabel?: string | undefined;
  onReset: () => void;
  onPromote: () => void;
};

/** The pill on a row whose value is stored for this site only: use the default here again, or make the value the default everywhere. */
export function ScopeMenu({ host, valueLabel, defaultLabel, onReset, onPromote }: ScopeMenuProps) {
  return (
    <DropdownMenu
      align="right"
      menuLabel={`Where ${valueLabel} applies`}
      header={<p className="border-b border-border px-2 pb-2 pt-1.5 text-xs text-muted-foreground">{valueLabel} applies only to {host}.</p>}
      trigger={
        <Button size="sm" variant="ghost" className="h-6 shrink-0 gap-1 rounded-full bg-primary/10 px-2 text-xs font-semibold text-primary hover:bg-primary/20">
          This site
          <UiChevronDown className="size-3" aria-hidden />
        </Button>
      }
      items={[
        {
          label: (
            <span className="flex w-full items-center justify-between gap-3">
              <span>Use the default here</span>
              {defaultLabel && <span className="truncate text-xs text-muted-foreground">{defaultLabel}</span>}
            </span>
          ),
          onSelect: onReset,
        },
        { label: "Make this the default for all sites", onSelect: onPromote },
      ]}
    />
  );
}
