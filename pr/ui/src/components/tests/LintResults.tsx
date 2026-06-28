import { GavelIcon } from '../GavelIcon';
import type { LinterResult, LintViolation } from './types';

export function LintResults({ lint }: { lint: LinterResult[] }) {
  if (lint.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">
        No lint findings in this run.
      </div>
    );
  }
  return (
    <div className="h-full space-y-4 overflow-auto p-3">
      {lint.map(result => (
        <LinterSection key={result.linter} result={result} />
      ))}
    </div>
  );
}

function LinterSection({ result }: { result: LinterResult }) {
  const count = result.violations.length;
  return (
    <div>
      <div className="mb-1 flex items-center gap-2 text-sm font-semibold">
        <GavelIcon
          name={count > 0 ? 'codicon:warning' : 'codicon:pass'}
          className={count > 0 ? 'text-amber-500' : 'text-green-500'}
        />
        {result.linter}
        <span className="text-xs font-normal text-muted-foreground">
          {count} violation{count !== 1 ? 's' : ''}
        </span>
      </div>
      <div className="divide-y divide-border rounded border border-border">
        {count === 0 ? (
          <div className="px-3 py-2 text-xs text-muted-foreground">No violations</div>
        ) : (
          result.violations.map((v, i) => <ViolationRow key={i} violation={v} />)
        )}
      </div>
    </div>
  );
}

function ViolationRow({ violation }: { violation: LintViolation }) {
  const rule = violation.code || violation.rule?.pattern || '';
  return (
    <div className="px-3 py-1.5 text-xs">
      <div className="flex items-center gap-2">
        <span className="font-mono text-muted-foreground">{location(violation)}</span>
        {rule && <span className="rounded bg-muted px-1 font-mono text-[10px]">{rule}</span>}
      </div>
      {violation.message && <div className="mt-0.5 whitespace-pre-wrap">{violation.message}</div>}
    </div>
  );
}

function location(v: LintViolation): string {
  if (!v.file) return '';
  let loc = v.file;
  if (v.line) loc += `:${v.line}`;
  if (v.column) loc += `:${v.column}`;
  return loc;
}
