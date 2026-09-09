import { APPROVAL_ICONS, SESSION_TONES, WORKFLOW_PHASES, type SessionTone } from '@flanksource/clicky-ui/ai';
import { Button, DropdownMenu, SplitButton, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { Icon } from '@flanksource/clicky-ui/data';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import { cn } from '@flanksource/clicky-ui/utils';
import { UiCheckFilled, UiCog, UiListChecks, UiStop } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoLifecycleStep, TodoRunEffort, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { TodoRuntimeSummary } from './TodoRuntimeSummary';
import type { RunContext } from './providers';
import { runSpec } from './run';

/**
 * The todo header's single lifecycle control.
 *
 * It replaces four controls per action (trigger + runtime combo + Advanced cog,
 * rendered twice for Plan and Run) with one split button: the primary does
 * whatever `todo.lifecycle.next` says, and the caret holds every other
 * *applicable* step, each with its own model and effort submenu, plus
 * Advanced.
 *
 * Where the todo stands in plan -> run -> verify used to be a client-side
 * machine derived from status + hasPlan + verification signals. That machine
 * is gone: the server now computes it and hands the header `todo.lifecycle`
 * (every step's status, which one is `next`, and `reason` it picked that one).
 * The only branches left on this side of the wire are the ones no server
 * verdict can replace — a session is running right now, or the todo is
 * parked on a human decision (review/ask).
 *
 * The runtime rides inside the button rather than beside it, so one control
 * says both what will run and exactly what will run it.
 */

export type PhaseRunOptions = Partial<Record<string, TodoRunOptions>>;

// primaryLifecycleAction picks the header's one action. A live session and the
// two human-decision statuses outrank the server's `next` — those are the
// only client-side decisions left; which step, whether it applies, and why is
// the server's call.
export type LifecyclePrimaryAction =
  | { kind: 'step'; name: string; label: string; reason: string }
  | { kind: 'stop' }
  | { kind: 'review' }
  | { kind: 'answer' }
  | { kind: 'none'; reason: string };

export function primaryLifecycleAction(todo: TodoItem, sessionInProgress: boolean): LifecyclePrimaryAction {
  if (sessionInProgress) return { kind: 'stop' };
  if (todo.status === 'review') return { kind: 'review' };
  if (todo.status === 'ask') return { kind: 'answer' };
  const lifecycle = todo.lifecycle;
  const next = lifecycle?.next ?? null;
  if (!next) return { kind: 'none', reason: lifecycle?.reason ?? 'Nothing to run.' };
  const step = lifecycle?.steps.find(entry => entry.name === next);
  return { kind: 'step', name: next, label: step?.label ?? next, reason: lifecycle?.reason ?? '' };
}

// otherLifecycleSteps is every applicable step besides the one already
// suggested — the caret's "run any step" menu. A step the server marked
// inapplicable is left out: offering it would just produce a rejected run.
export function otherLifecycleSteps(todo: TodoItem, primary: LifecyclePrimaryAction): TodoLifecycleStep[] {
  const steps = todo.lifecycle?.steps ?? [];
  const skip = primary.kind === 'step' ? primary.name : undefined;
  return steps.filter(entry => entry.applicable && entry.name !== skip);
}

type StepGlyph = { icon: StaticIconComponent; tone: SessionTone };

// Glyph and tone for the three pipeline steps come from the library's Agent
// Action Icons set, so a step looks the same in the header as it does in the
// session viewer. Triage and any other server-declared step name (a
// `.gavel.yaml` todos.prompts entry) fall back to a generic glyph — the
// server owns step identity now, so the client cannot enumerate every
// possible name up front.
const KNOWN_STEP_GLYPHS: Partial<Record<string, StepGlyph>> = {
  plan: { icon: WORKFLOW_PHASES.plan.icon, tone: WORKFLOW_PHASES.plan.tone },
  run: { icon: WORKFLOW_PHASES.run.icon, tone: WORKFLOW_PHASES.run.tone },
  verify: { icon: WORKFLOW_PHASES.verify.icon, tone: WORKFLOW_PHASES.verify.tone },
};

const DEFAULT_STEP_GLYPH: StepGlyph = { icon: UiListChecks, tone: 'slate' };

function stepGlyph(name: string): StepGlyph {
  return KNOWN_STEP_GLYPHS[name] ?? DEFAULT_STEP_GLYPH;
}

export type TodoPhaseButtonProps = {
  todo: TodoItem;
  sessionInProgress: boolean;
  context: RunContext | null;
  /** Effective run options per step, already reconciled against the catalog. */
  options: PhaseRunOptions;
  disabled?: boolean;
  busy?: boolean;
  /** Enter a step now. */
  onRunStep: (name: string) => void;
  /** Stop the run in flight. Absent when nothing is stoppable. */
  onStop?: (() => void) | undefined;
  /** Scroll to / focus the review banner for a plan awaiting a decision. */
  onReview?: (() => void) | undefined;
  /** Change one step's runtime without running it. */
  onOptionsChange: (name: string, options: TodoRunOptions) => void;
  /** Open the full spec editor for a step. */
  onAdvanced: (name: string) => void;
};

/**
 * The menu rows for every other applicable step.
 *
 * Each step is a *submenu*, not a leaf: running it, changing its model and
 * changing its effort are three different intents on the same step, and
 * flattening them would put a dozen rows in one menu.
 */
function stepItems({
  todo,
  primary,
  context,
  options,
  onRunStep,
  onOptionsChange,
}: Pick<TodoPhaseButtonProps, 'todo' | 'context' | 'options' | 'onRunStep' | 'onOptionsChange'> & { primary: LifecyclePrimaryAction }): DropdownMenuItem[] {
  return otherLifecycleSteps(todo, primary).map(entry => {
    const glyph = stepGlyph(entry.name);
    const stepOptions = options[entry.name];
    const spec = stepOptions ? runSpec(stepOptions) : {};
    const models = context?.modes.find(runtime => runtime.id === spec.mode)?.models
      ?? context?.modes.flatMap(runtime => runtime.models)
      ?? [];
    const children: DropdownMenuItem[] = [
      { label: `${entry.label} now`, icon: glyph.icon, onSelect: () => onRunStep(entry.name) },
    ];
    if (context && stepOptions) {
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
            onSelect: () => onOptionsChange(entry.name, { ...stepOptions, spec: { ...spec, model: model.id } }),
          })),
        },
        {
          label: 'Effort',
          group: 'Runtime',
          onSelect: () => {},
          children: (context.efforts ?? []).map(value => ({
            label: value === spec.effort ? `✓ ${value}` : value,
            onSelect: () => onOptionsChange(entry.name, { ...stepOptions, spec: { ...spec, effort: value as TodoRunEffort } }),
          })),
        },
      );
    }
    return {
      // The glyph rides in the label rather than the menu's icon slot, which
      // only takes a raw CSS colour and so cannot carry a dark-mode variant.
      label: (
        <span className="flex min-w-0 flex-1 items-center gap-2">
          <Icon icon={glyph.icon} className={cn('size-4 shrink-0', SESSION_TONES[glyph.tone].text)} />
          <span className="shrink-0">{entry.label}</span>
          {stepOptions && context && (
            <TodoRuntimeSummary options={stepOptions} context={context} className="ml-auto" />
          )}
        </span>
      ),
      group: 'Run any step',
      title: entry.label,
      onSelect: () => onRunStep(entry.name),
      children,
    };
  });
}

/**
 * One small glyph per lifecycle step, saying where the todo stands.
 *
 * Anonymous dots could only say "two of three done" — these say *which*
 * steps, because each keeps its own icon and colour. State rides on the
 * container: a tinted ring for `todo.lifecycle.next`, a faded glyph for a
 * step not done and not next.
 */
export function TodoPhaseTicks({ todo, className }: { todo: TodoItem; className?: string }) {
  const steps = todo.lifecycle?.steps ?? [];
  if (steps.length === 0) return null;
  const next = todo.lifecycle?.next ?? null;

  return (
    <span
      className={cn('inline-flex shrink-0 items-center gap-0.5 rounded-md border border-border bg-muted/20 px-1 py-0.5', className)}
      role="img"
      aria-label={steps.map(entry => `${entry.label} ${entry.name === next ? 'current' : entry.done ? 'done' : 'todo'}`).join(', ')}
    >
      {steps.map(entry => {
        const current = entry.name === next;
        const glyph = stepGlyph(entry.name);
        return (
          <span
            key={entry.name}
            title={`${entry.label}: ${current ? 'current' : entry.done ? 'done' : 'todo'}`}
            className={cn(
              'inline-flex size-6 items-center justify-center rounded',
              current && 'bg-background ring-1 ring-border',
            )}
          >
            <Icon
              icon={glyph.icon}
              className={cn('size-[1.125rem]', !current && !entry.done ? 'text-muted-foreground/40' : SESSION_TONES[glyph.tone].text)}
            />
          </span>
        );
      })}
    </span>
  );
}

export function TodoPhaseButton({
  todo,
  sessionInProgress,
  context,
  options,
  disabled,
  busy,
  onRunStep,
  onStop,
  onReview,
  onOptionsChange,
  onAdvanced,
}: TodoPhaseButtonProps) {
  const primary = primaryLifecycleAction(todo, sessionInProgress);
  const targetOptions = primary.kind === 'step' ? options[primary.name] : undefined;
  const targetGlyph = primary.kind === 'step' ? stepGlyph(primary.name) : undefined;
  const items: DropdownMenuItem[] = [
    ...stepItems({ todo, primary, context, options, onRunStep, onOptionsChange }),
    {
      label: 'Advanced…',
      icon: UiCog,
      group: 'Configure',
      onSelect: () => onAdvanced(primary.kind === 'step' ? primary.name : 'run'),
    },
  ];

  const label = primary.kind === 'step'
    ? primary.label
    : primary.kind === 'stop'
      ? 'Stop'
      : primary.kind === 'answer'
        ? 'Answer'
        : primary.kind === 'review'
          ? 'Review plan'
          : 'Done';

  function activate() {
    if (primary.kind === 'step') return onRunStep(primary.name);
    if (primary.kind === 'stop') return onStop?.();
    if (primary.kind === 'none') return;
    return onReview?.();
  }

  // Stop with nothing stoppable, review/answer with nowhere to go, and "no
  // step applies" are the cases where the primary genuinely has no action —
  // better disabled and honest than a control that looks live and does
  // nothing.
  const inert = primary.kind === 'none'
    || (primary.kind === 'stop' && !onStop)
    || ((primary.kind === 'review' || primary.kind === 'answer') && !onReview);

  const tooltip = primary.kind === 'step' ? (primary.reason || primary.label)
    : primary.kind === 'none' ? primary.reason
      : label;

  return (
    <SplitButton
      label={
        <span
          className={cn(
            'inline-flex min-w-0 items-center gap-1.5',
            // The neutral bordered shell is kept for Stop and only the label is
            // tinted, rather than switching to a solid destructive button.
            primary.kind === 'stop' && 'text-red-600 [[data-theme=dark]_&]:text-red-400',
          )}
        >
          {busy ? (
            <Spinner className="shrink-0 text-xs" />
          ) : targetGlyph ? (
            <Icon icon={targetGlyph.icon} className={cn('size-4 shrink-0', SESSION_TONES[targetGlyph.tone].text)} />
          ) : primary.kind === 'stop' ? (
            <UiStop className="shrink-0 text-xs" />
          ) : primary.kind === 'none' ? (
            <UiCheckFilled className="shrink-0 text-xs text-emerald-600 dark:text-emerald-400" />
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
      title={tooltip}
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
