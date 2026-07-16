import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import { UiBatteryChargingVertical, UiBatteryVerticalEmpty, UiBatteryVerticalFull, UiBatteryVerticalHigh, UiBatteryVerticalLow, UiBatteryVerticalMedium } from '@flanksource/clicky-ui/icons';
import type { TodoRunEffort } from '../../types';

type EffortPresentation = {
  icon: StaticIconComponent;
  label: string;
  className: string;
};

export function todoRunEffortPresentation(effort: TodoRunEffort | undefined): EffortPresentation {
  switch (effort) {
    case 'low':
      return { icon: UiBatteryVerticalLow, label: 'Low', className: 'border-sky-500/25 bg-sky-500/10 text-sky-600 [[data-theme=dark]_&]:text-sky-400' };
    case 'medium':
      return { icon: UiBatteryVerticalMedium, label: 'Medium', className: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-600 [[data-theme=dark]_&]:text-emerald-400' };
    case 'high':
      return { icon: UiBatteryVerticalHigh, label: 'High', className: 'border-amber-500/30 bg-amber-500/10 text-amber-600 [[data-theme=dark]_&]:text-amber-400' };
    case 'xhigh':
      return { icon: UiBatteryVerticalFull, label: 'XHigh', className: 'border-orange-500/30 bg-orange-500/10 text-orange-600 [[data-theme=dark]_&]:text-orange-400' };
    case 'max':
      return { icon: UiBatteryChargingVertical, label: 'Max', className: 'border-rose-500/30 bg-rose-500/10 text-rose-600 [[data-theme=dark]_&]:text-rose-400' };
    case 'ultra':
      return { icon: UiBatteryChargingVertical, label: 'Ultra', className: 'border-transparent bg-gradient-to-r from-fuchsia-500 via-cyan-500 to-amber-400 text-white shadow-sm' };
    default:
      return { icon: UiBatteryVerticalEmpty, label: 'Fixed', className: 'border-border bg-muted/60 text-muted-foreground' };
  }
}

export function TodoRunEffortBadge({ effort, className = '' }: { effort: TodoRunEffort | undefined; className?: string }) {
  const presentation = todoRunEffortPresentation(effort);
  const EffortIcon = presentation.icon;
  return (
    <span className={`inline-flex h-5 shrink-0 items-center gap-1 rounded-full border px-1.5 text-[10px] font-semibold leading-none ${presentation.className} ${className}`} title={`Effort: ${presentation.label}`}>
      <EffortIcon className="size-3" aria-hidden="true" />
      {presentation.label}
    </span>
  );
}
