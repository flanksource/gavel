import { APPROVAL_ICONS, SESSION_TONES } from '@flanksource/clicky-ui/ai';
import { Button, DropdownMenu, SplitButton, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { Icon } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import { UiCog, UiStop } from '@flanksource/clicky-ui/icons';
import type { TodoRunEffort, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import {
  PHASES,
  otherPhases,
  phase,
  phaseVerb,
  primaryAction,
  stepStatus,
  type PhaseId,
  type PhaseState,
} from './phaseMachine';
import { TodoRuntimeSummary } from './TodoRuntimeSummary';
import type { RunContext } from './providers';
import { runSpec } from './run';

/**
 * The todo header's single phase control.
 *
 * It replaces four controls per action (trigger + runtime combo + Advanced cog,
 * rendered twice for Plan and Run) with one split button: the primary does
 * whatever the machine suggests next, and the caret holds every *other* phase,
 * each with its own model and effort submenu, plus Advanced.
 *
 * The runtime rides inside the button rather than beside it, so one control says
 * both what will run and exactly what will run it. That is the combo's own
 * readout — provider, mode, model, effort in that order — not a new vocabulary.
 */

export type PhaseRunOptions = Partial<Record<PhaseId, TodoRunOptions>>;

export type TodoPhaseButtonProps = {
  state: PhaseState;
  context: RunContext | null;
  /** Effective run options per phase, already reconciled against the catalog. */
  options: PhaseRunOptions;
  disabled?: boolean;
  busy?: boolean;
  /** Enter a phase now. */
  onRunPhase: (id: PhaseId) => void;
  /** Stop the run in flight. Absent when nothing is stoppable. */
  onStop?: (() => void) | undefined;
  /** Scroll to / focus the review banner for a plan awaiting a decision. */
  onReview?: (() => void) | undefined;
  /** Change one phase's runtime without running it. */
  onOptionsChange: (id: PhaseId, options: TodoRunOptions) => void;
  /** Open the full spec editor for a phase. */
  onAdvanced: (id: PhaseId) => void;
};

/**
 * The menu rows for every non-suggested phase.
 *
 * Each phase is a *submenu*, not a leaf: running it, changing its model and
 * changing its effort are three different intents on the same phase, and
 * flattening them would put a dozen rows in one menu.
 */
function phaseItems({
  state,
  context,
  options,
  onRunPhase,
  onOptionsChange,
}: Pick<TodoPhaseButtonProps, 'state' | 'context' | 'options' | 'onRunPhase' | 'onOptionsChange'>): DropdownMenuItem[] {
  return otherPhases(state).map(entry => {
    const phaseOptions = options[entry.id];
    const spec = phaseOptions ? runSpec(phaseOptions) : {};
    const models = context?.backends.find(backend => backend.id === spec.backend)?.models
      ?? context?.backends.flatMap(backend => backend.models)
      ?? [];
    const children: DropdownMenuItem[] = [
      { label: `${phaseVerb(state, entry.id)} now`, icon: entry.icon, onSelect: () => onRunPhase(entry.id) },
    ];
    if (context && phaseOptions) {
      children.push(
        {
          label: 'Model',
          group: 'Runtime',
          // Required by the type even though a parent with `children` never
          // fires it — the submenu opens instead.
          onSelect: () => {},
          children: models.map(model => ({
            label: model.id === spec.model ? `✓ ${model.label}` : model.label,
            title: model.id,
            onSelect: () => onOptionsChange(entry.id, { ...phaseOptions, spec: { ...spec, model: model.id } }),
          })),
        },
        {
          label: 'Effort',
          group: 'Runtime',
          onSelect: () => {},
          children: context.efforts.map(value => ({
            label: value === spec.effort ? `✓ ${value}` : value,
            onSelect: () => onOptionsChange(entry.id, { ...phaseOptions, spec: { ...spec, effort: value as TodoRunEffort } }),
          })),
        },
      );
    }
    return {
      // The glyph rides in the label rather than the menu's icon slot, which
      // only takes a raw CSS colour and so cannot carry a dark-mode variant. A
      // phase is identified by its hue everywhere else; the menu that picks one
      // is the last place to drop it.
      label: (
        <span className="flex min-w-0 flex-1 items-center gap-2">
          <Icon icon={entry.icon} className={cn('size-4 shrink-0', SESSION_TONES[entry.tone].text)} />
          <span className="shrink-0">{phaseVerb(state, entry.id)}</span>
          {phaseOptions && context && (
            <TodoRuntimeSummary options={phaseOptions} context={context} className="ml-auto" />
          )}
        </span>
      ),
      group: 'Run any phase',
      title: entry.title,
      onSelect: () => onRunPhase(entry.id),
      children,
    };
  });
}

/**
 * One small glyph per phase, saying where the todo stands.
 *
 * Anonymous dots could only say "two of three done" — these say *which* phases,
 * because each keeps its own icon and colour. State rides on the container: a
 * tinted ring for the current phase, a faded glyph for one not reached yet.
 */
export function TodoPhaseTicks({ state, className }: { state: PhaseState; className?: string }) {
  return (
    <span
      className={cn('inline-flex shrink-0 items-center gap-0.5 rounded-md border border-border bg-muted/20 px-1 py-0.5', className)}
      role="img"
      aria-label={PHASES.map(entry => `${entry.label} ${stepStatus(state, entry.id)}`).join(', ')}
    >
      {PHASES.map(entry => {
        const status = stepStatus(state, entry.id);
        return (
          <span
            key={entry.id}
            title={`${entry.label}: ${status}`}
            className={cn(
              'inline-flex size-6 items-center justify-center rounded',
              status === 'current' && 'bg-background ring-1 ring-border',
            )}
          >
            <Icon
              icon={entry.icon}
              className={cn('size-[1.125rem]', status === 'todo' ? 'text-muted-foreground/40' : SESSION_TONES[entry.tone].text)}
            />
          </span>
        );
      })}
    </span>
  );
}

export function TodoPhaseButton({
  state,
  context,
  options,
  disabled,
  busy,
  onRunPhase,
  onStop,
  onReview,
  onOptionsChange,
  onAdvanced,
}: TodoPhaseButtonProps) {
  const action = primaryAction(state);
  const target = action.kind === 'phase' ? phase(action.phase) : undefined;
  const targetOptions = action.kind === 'phase' ? options[action.phase] : undefined;
  const items: DropdownMenuItem[] = [
    ...phaseItems({ state, context, options, onRunPhase, onOptionsChange }),
    {
      label: 'Advanced…',
      icon: UiCog,
      group: 'Configure',
      onSelect: () => onAdvanced(action.kind === 'phase' ? action.phase : 'run'),
    },
  ];

  const label = action.kind === 'phase'
    ? action.label
    : action.kind === 'stop'
      ? 'Stop'
      : action.kind === 'answer'
        ? 'Answer'
        : 'Review plan';

  function activate() {
    if (action.kind === 'phase') return onRunPhase(action.phase);
    if (action.kind === 'stop') return onStop?.();
    return onReview?.();
  }

  // Stop with nothing stoppable, and review with nowhere to go, are the two
  // cases where the primary genuinely has no action — better disabled and
  // honest than a control that looks live and does nothing.
  const inert = (action.kind === 'stop' && !onStop) || ((action.kind === 'review' || action.kind === 'answer') && !onReview);

  return (
    <SplitButton
      label={
        <span
          className={cn(
            'inline-flex min-w-0 items-center gap-1.5',
            // The neutral bordered shell is kept for Stop and only the label is
            // tinted, rather than switching to a solid destructive button.
            action.kind === 'stop' && 'text-red-600 [[data-theme=dark]_&]:text-red-400',
          )}
        >
          {busy ? (
            <Spinner className="shrink-0 text-xs" />
          ) : target ? (
            <Icon icon={target.icon} className={cn('size-4 shrink-0', SESSION_TONES[target.tone].text)} />
          ) : action.kind === 'stop' ? (
            <UiStop className="shrink-0 text-xs" />
          ) : (
            <Icon
              icon={APPROVAL_ICONS.question.icon}
              className={cn('size-4 shrink-0', SESSION_TONES[APPROVAL_ICONS.question.tone].text)}
            />
          )}
          <span className="shrink-0 font-medium">{label}</span>
          {targetOptions && context && (
            <>
              <span className="mx-0.5 h-3.5 w-px shrink-0 bg-border" aria-hidden="true" />
              <TodoRuntimeSummary options={targetOptions} context={context} />
            </>
          )}
        </span>
      }
      items={items}
      onClick={activate}
      disabled={disabled || inert}
      variant="outline"
      size="sm"
      align="right"
      className="h-8"
      title={target ? target.title : label}
    />
  );
}

/** The overflow `⋮` beside the phase button. */
export function TodoOverflowButton({ items, disabled }: { items: DropdownMenuItem[]; disabled?: boolean }) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Issue actions"
      menuClassName="w-80"
      items={items}
      trigger={
        <Button
          variant="ghost"
          size="icon"
          type="button"
          disabled={disabled}
          title="Issue actions"
          aria-label="Issue actions"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        >
          <span aria-hidden="true">⋮</span>
        </Button>
      }
    />
  );
}
