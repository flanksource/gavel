import { UiError, UiPass } from '@flanksource/clicky-ui/icons';
import { LogViewer } from '../LogViewer';
import type { VerificationChecklistItem, VerificationFixtureResult } from './verificationAttempts';

const PASS_STATUSES = new Set(['pass', 'passed', 'success', 'ok']);

export function stepPassed(step: VerificationFixtureResult): boolean {
  return PASS_STATUSES.has((step.status ?? '').toLowerCase());
}

/**
 * The Output tab of one verification attempt: every non-runner step (exec, bash,
 * ai) with its command, exit code and ANSI-rendered streams, plus the acceptance
 * criteria checklist the definition of done evaluated.
 */
export function VerificationStepOutput({
  steps,
  checklist,
}: {
  steps: VerificationFixtureResult[];
  checklist: VerificationChecklistItem[];
}) {
  if (steps.length === 0 && checklist.length === 0) {
    return <p className="px-3 py-4 text-xs text-muted-foreground">This attempt produced no command output or checklist.</p>;
  }
  return (
    <div className="space-y-2 p-3">
      {steps.map((step, index) => (
        <StepCard key={`${step.name ?? step.type ?? 'step'}:${index}`} step={step} />
      ))}
      {checklist.length > 0 && (
        <section className="rounded border border-border bg-card">
          <h4 className="border-b border-border px-2.5 py-1.5 text-[11px] font-semibold uppercase text-muted-foreground">Acceptance criteria</h4>
          <ul className="space-y-1 p-2.5">
            {checklist.map((item, index) => {
              const Icon = item.passed ? UiPass : UiError;
              return (
                <li key={`${item.item ?? 'criterion'}:${index}`} className="flex items-start gap-1.5 text-xs">
                  <Icon className={`mt-0.5 shrink-0 text-sm ${item.passed ? 'text-emerald-600' : 'text-red-600'}`} />
                  <span className="min-w-0">
                    {item.item}
                    {item.message ? <span className="block text-[11px] text-muted-foreground">{item.message}</span> : null}
                  </span>
                </li>
              );
            })}
          </ul>
        </section>
      )}
    </div>
  );
}

function StepCard({ step }: { step: VerificationFixtureResult }) {
  const passed = stepPassed(step);
  const Icon = passed ? UiPass : UiError;
  return (
    <section className="rounded border border-border bg-card">
      <header className="flex flex-wrap items-center gap-2 border-b border-border px-2.5 py-1.5">
        <Icon className={`shrink-0 text-sm ${passed ? 'text-emerald-600' : 'text-red-600'}`} />
        <span className="min-w-0 flex-1 truncate text-xs font-medium">{step.name || step.type || 'Verification step'}</span>
        {step.type && <span className="rounded border border-border px-1.5 py-0.5 text-[10px] uppercase text-muted-foreground">{step.type}</span>}
        {typeof step.duration === 'number' && step.duration > 0 && (
          <span className="text-[10px] tabular-nums text-muted-foreground">{stepDuration(step.duration)}</span>
        )}
        {typeof step.exit_code === 'number' && (
          <span className={`text-[10px] tabular-nums ${step.exit_code === 0 ? 'text-muted-foreground' : 'text-red-600'}`}>exit {step.exit_code}</span>
        )}
      </header>
      <div className="space-y-1.5 px-2.5 py-2 text-xs">
        {step.command && (
          <div className="min-w-0">
            <code className="block overflow-x-auto whitespace-pre rounded bg-muted px-1.5 py-1 text-[11px]">{step.command}</code>
            {step.cwd && <span className="mt-0.5 block text-[10px] text-muted-foreground">in {step.cwd}</span>}
          </div>
        )}
        {step.error && <div className="text-[11px] text-red-600">{step.error}</div>}
        {step.stdout && <StreamBlock label="stdout" text={step.stdout} />}
        {step.stderr && <StreamBlock label="stderr" text={step.stderr} />}
        {!passed && <CelDetails step={step} />}
      </div>
    </section>
  );
}

// Go marshals time.Duration as nanoseconds.
function stepDuration(nanos: number): string {
  const ms = nanos / 1e6;
  if (ms < 1) return '<1ms';
  if (ms < 1000) return `${Math.round(ms)}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

function StreamBlock({ label, text }: { label: string; text: string }) {
  return (
    <div>
      <span className="text-[10px] uppercase text-muted-foreground">{label}</span>
      <LogViewer logs={text} collapsedLines={8} />
    </div>
  );
}

// CEL detail is rendered only for a failing step: the expression and the values
// it saw are what explain the verdict.
function CelDetails({ step }: { step: VerificationFixtureResult }) {
  const vars = Object.entries(step.cel_vars ?? {});
  const expression = step.cel_trace || step.cel_expression;
  const expected = comparisonValue(step, step.expected);
  const actual = comparisonValue(step, step.actual);
  if (!expression && vars.length === 0 && expected === undefined && actual === undefined) return null;
  return (
    <div className="space-y-1 rounded border border-border bg-muted/30 px-2 py-1.5">
      {expression && (
        <div>
          <span className="text-[10px] uppercase text-muted-foreground">expression</span>
          <code className="block overflow-x-auto whitespace-pre text-[11px]">{expression}</code>
        </div>
      )}
      {vars.length > 0 && (
        <table className="w-full table-fixed text-[11px]">
          <tbody>
            {vars.map(([name, value]) => (
              <tr key={name} className="align-top">
                <td className="w-1/3 truncate pr-2 font-medium text-muted-foreground">{name}</td>
                <td className="break-words font-mono">{formatCelValue(value)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {expected !== undefined && (
        <div className="text-[11px]">
          <span className="text-muted-foreground">expected </span>
          <span className="font-mono">{formatCelValue(expected)}</span>
        </div>
      )}
      {actual !== undefined && (
        <div className="text-[11px]">
          <span className="text-muted-foreground">actual </span>
          <span className="font-mono">{formatCelValue(actual)}</span>
        </div>
      )}
    </div>
  );
}

// A command step with no CEL assertion stores the whole command result object
// in `actual` — command, exit code, stdout and duration are all rendered by the
// card already, so dumping it again as JSON is noise. Anything a CEL expression
// compared is shown as-is, structured values included.
function comparisonValue(step: VerificationFixtureResult, value: unknown): unknown {
  if (value === undefined) return undefined;
  if (step.cel_expression || step.cel_trace) return value;
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) return undefined;
  return value;
}

function formatCelValue(value: unknown): string {
  if (typeof value === 'string') return value;
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}
