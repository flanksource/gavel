import type { ComponentType } from 'react';
import { SESSION_TONES, effortIcon } from '@flanksource/clicky-ui/ai';
import { Icon } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import { UiRobotAi, type IconProps } from '@flanksource/clicky-ui/icons';
import type { TodoRunOptions } from '../../types';
import { buildRunFamilies, type RunContext } from './providers';
import { runSpec, todoRunButtonPresentation, todoRunModeLabel } from './run';

/**
 * Provider · mode · model · effort, in one control-height run.
 *
 * This is the runtime combo's own resting readout — same parts, same order — so
 * a reader who has learnt "Claude · cmux · sonnet-4.5 · high" left to right does
 * not have to re-learn it because it now lives inside the phase button instead
 * of beside it. It is display only; the caret's submenus do the changing.
 *
 * The effort glyph and its hue come from the library's Agent Action Icons ramp,
 * so a "high" here is the same battery the session viewer draws for that run.
 */
export function TodoRuntimeSummary({
  options,
  context,
  className,
  showModel = true,
}: {
  options: TodoRunOptions;
  context: RunContext;
  className?: string;
  /** Off for the tightest placements, where the glyphs alone must carry it. */
  showModel?: boolean;
}) {
  const presentation = todoRunButtonPresentation(options, context);
  const mode = todoRunModeLabel(options, context);
  const spec = runSpec(options);
  const family = buildRunFamilies(context).find(entry => entry.modes.some(item => item.id === spec.mode));
  const modeIcon = family?.modes.find(item => item.id === spec.mode)?.icon;
  // provider.icon may be a runtime icon name rather than a component, so fall
  // back to the generic agent glyph unless it can be rendered directly.
  const provider = presentation.provider?.icon;
  const ProviderIcon = provider && typeof provider !== 'string' ? (provider as ComponentType<IconProps>) : UiRobotAi;
  const effort = presentation.effort ? effortIcon(presentation.effort) : undefined;

  return (
    <span
      className={cn('inline-flex min-w-0 items-center gap-1', className)}
      title={`${presentation.provider?.label ?? 'Agent'} · ${mode} · ${spec.model ?? presentation.model}${presentation.effort ? ` · ${presentation.effort}` : ''}`}
    >
      <ProviderIcon
        className="size-4 shrink-0"
        {...(presentation.provider?.iconColor ? { style: { color: presentation.provider.iconColor } } : {})}
      />
      {/* `buildRunFamilies` only gives a mode its own glyph for cmux and falls
          back to the provider's for everything else, so rendering it
          unconditionally draws the same icon twice and says nothing the brand
          has not already said. The mode still reaches the reader in the title. */}
      {modeIcon && typeof modeIcon !== 'string' && modeIcon !== provider && (
        <Icon icon={modeIcon} className="size-4 shrink-0 text-muted-foreground" />
      )}
      {showModel && <span className="min-w-0 truncate text-[11px] text-muted-foreground">{presentation.model}</span>}
      {effort && <Icon icon={effort.icon} className={cn('size-4 shrink-0', SESSION_TONES[effort.tone].text)} />}
    </span>
  );
}
