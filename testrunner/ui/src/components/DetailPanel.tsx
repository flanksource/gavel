import type { Test, FixtureContext, GinkgoContext, GoTestContext, Violation, LinterResult, RunMeta } from '../types';
import {
  statusIcon,
  statusColor,
  formatDuration,
  sum,
  frameworkIcon,
  relPath,
  lintNodeCount,
  formatRunTimestamp,
  formatRunDuration,
  hasTimeoutArgs,
  timeoutArgValue,
} from '../utils';
import { JsonView } from './JsonView';
import { AnsiHtml } from './AnsiHtml';
import { ProgressBar } from './ProgressBar';

export interface IgnoreRequest {
  source?: string;
  rule?: string;
  file?: string;
  work_dir?: string;
}

interface Props {
  test: Test | null;
  lint?: LinterResult[];
  onRerun?: (t: Test) => void;
  rerunBusy?: boolean;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
  runMeta?: RunMeta;
}

function taskMeta(t: Test): { duration?: string; status?: string; type?: string } | null {
  if (t.framework !== 'task' || !t.context || typeof t.context !== 'object') return null;
  const ctx = t.context as Record<string, any>;
  return {
    duration: typeof ctx.duration === 'string' ? ctx.duration : undefined,
    status: typeof ctx.status === 'string' ? ctx.status : undefined,
    type: typeof ctx.type === 'string' ? ctx.type : undefined,
  };
}

export function DetailPanel({ test: t, lint, onRerun, rerunBusy, onIgnore, ignoreBusy, runMeta }: Props) {
  if (!t) {
    return (
      <div class="flex items-center justify-center h-full text-gray-400 text-sm">
        <div class="text-center">
          <iconify-icon icon="codicon:list-tree" class="text-4xl mb-2 block" />
          Select a test to view details
        </div>
      </div>
    );
  }

  const hasChildren = (t.children?.length ?? 0) > 0;
  const s = hasChildren ? sum(t) : null;
  const fw = t.framework;
  const fwIcon = frameworkIcon(fw);
  const task = taskMeta(t);
  const isLint = t.kind === 'lint-root' || t.kind === 'lint-folder' || t.kind === 'linter'
    || t.kind === 'violation' || t.kind === 'lint-file' || t.kind === 'lint-rule' || t.kind === 'lint-rule-group';
  const canRerun = !!onRerun && t.kind !== 'violation' && t.framework !== 'task';

  return (
    <div class="h-full overflow-y-auto p-5 space-y-4">
      {/* Header */}
      <div class="flex items-start gap-2">
        <iconify-icon icon={statusIcon(t)} class={`${statusColor(t)} text-2xl shrink-0 mt-0.5`} />
        <div class="min-w-0 flex-1">
          <div class="flex items-start justify-between gap-2">
            <h2 class="text-lg font-bold text-gray-900 break-words">{t.name}</h2>
            {canRerun && (
              <button
                class="shrink-0 text-xs px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                onClick={() => onRerun!(t)}
                disabled={rerunBusy}
                title="Rerun this test or subtree"
              >
                <iconify-icon icon="codicon:refresh" />
                {rerunBusy ? 'Running...' : 'Rerun'}
              </button>
            )}
          </div>
          <div class="flex items-center gap-2 mt-1 flex-wrap">
            {fwIcon && (
              <span class="inline-flex items-center gap-1 text-xs bg-gray-100 rounded px-1.5 py-0.5 text-gray-600">
                <iconify-icon icon={fwIcon} class="text-sm" />
                {fw}
              </span>
            )}
            {t.duration ? (
              <span class="text-xs text-gray-500">
                <iconify-icon icon="codicon:clock" class="mr-0.5" />
                {formatDuration(t.duration)}
              </span>
            ) : task?.duration ? (
              <span class="text-xs text-gray-500">
                <iconify-icon icon="codicon:clock" class="mr-0.5" />
                {task.duration}
              </span>
            ) : null}
            {t.file && (
              <span class="text-xs text-gray-500 font-mono">
                {t.file}{t.line ? `:${t.line}` : ''}
              </span>
            )}
          </div>
        </div>
      </div>

      {t.kind === 'violation' && t.violation && <ViolationDetail v={t.violation} />}
      {t.kind === 'linter' && <LinterDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />}
      {t.kind === 'lint-folder' && (
        <FolderLintDetail t={t} lint={lint} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}
      {t.kind === 'lint-file' && (
        <FileViolationsDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}
      {(t.kind === 'lint-rule' || t.kind === 'lint-rule-group') && (
        <RuleViolationsDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}

      {runMeta && (
        <Section title="Run">
          <div class="grid grid-cols-3 gap-3 text-sm">
            <MetaCard
              label={runMeta.kind === 'rerun' ? `Rerun #${runMeta.sequence}` : 'Initial run'}
              value={runMeta.started ? formatRunTimestamp(runMeta.started) : 'Unavailable'}
            />
            <MetaCard
              label="Finished"
              value={runMeta.ended ? formatRunTimestamp(runMeta.ended) : 'In progress'}
            />
            <MetaCard label="Duration" value={formatRunDuration(runMeta.started, runMeta.ended)} />
          </div>
          {hasTimeoutArgs(runMeta.args) && (
            <div class="mt-3 grid grid-cols-3 gap-3 text-sm">
              {timeoutArgValue(runMeta.args, 'timeout') && (
                <MetaCard label="Global timeout" value={timeoutArgValue(runMeta.args, 'timeout')!} />
              )}
              {timeoutArgValue(runMeta.args, 'test_timeout') && (
                <MetaCard label="Per-test" value={timeoutArgValue(runMeta.args, 'test_timeout')!} />
              )}
              {timeoutArgValue(runMeta.args, 'lint_timeout') && (
                <MetaCard label="Per-linter" value={timeoutArgValue(runMeta.args, 'lint_timeout')!} />
              )}
            </div>
          )}
        </Section>
      )}

      {/* Summary for containers */}
      {s && s.total > 0 && (
        <div class="space-y-3 border rounded-lg p-3 bg-gray-50">
          <div class="flex gap-4 text-sm">
            <Stat label="Total" value={s.total} color="text-gray-700" />
            <Stat label="Passed" value={s.passed} color="text-green-600" />
            <Stat label="Failed" value={s.failed} color="text-red-600" />
            {s.skipped > 0 && <Stat label="Skipped" value={s.skipped} color="text-yellow-600" />}
            {s.pending > 0 && <Stat label="Pending" value={s.pending} color="text-blue-600" />}
          </div>
          <ProgressBar
            segments={[
              { count: s.passed, color: 'bg-green-500', label: 'passed' },
              { count: s.skipped, color: 'bg-yellow-400', label: 'skipped' },
              { count: s.failed, color: 'bg-red-500', label: 'failed' },
              { count: s.pending, color: 'bg-blue-300', label: 'pending' },
            ]}
            total={s.total}
            height="h-2.5"
          />
        </div>
      )}

      {/* Suite path */}
      {t.suite && t.suite.length > 0 && (
        <Section title="Suite">
          <span class="text-sm text-gray-700">{t.suite.join(' > ')}</span>
        </Section>
      )}

      {/* Package */}
      {t.package_path && (
        <Section title="Package">
          <span class="text-sm text-gray-700 font-mono">{t.package_path}</span>
        </Section>
      )}

      {/* Error message */}
      {t.message && (
        <Section title="Error">
          <pre class="text-sm text-red-700 whitespace-pre-wrap font-mono bg-red-50 rounded p-3 max-h-64 overflow-y-auto">
            {t.message}
          </pre>
        </Section>
      )}

      {/* Command */}
      {t.command && (
        <Section title="Command">
          <pre class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-3">
            <span class="text-gray-400">$ </span>{t.command}
          </pre>
        </Section>
      )}

      {task && (task.status || task.type) && (
        <Section title="Task">
          <div class="flex gap-4 text-sm text-gray-700">
            {task.type && <span>type: <span class="font-mono">{task.type}</span></span>}
            {task.status && <span>status: <span class="font-mono">{task.status}</span></span>}
          </div>
        </Section>
      )}

      {/* Framework-specific context */}
      {fw === 'fixture' && t.context && <FixtureDetail ctx={t.context as FixtureContext} />}
      {fw === 'ginkgo' && t.context && <GinkgoDetail ctx={t.context as GinkgoContext} />}
      {fw === 'go test' && t.context && <GoTestDetail ctx={t.context as GoTestContext} />}

      {t.stdout && (
        <Section title="stdout">
          <AnsiHtml text={t.stdout} class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto" />
        </Section>
      )}

      {t.stderr && (
        <Section title="stderr">
          <AnsiHtml text={t.stderr} class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto" />
        </Section>
      )}

      {/* Children list for containers */}
      {hasChildren && (
        <Section title={isLint ? 'Tree' : 'Tests'}>
          <div class="space-y-0.5">
            {t.children!.map((child, i) => (
              <ChildRow key={i} test={child} />
            ))}
          </div>
        </Section>
      )}
    </div>
  );
}

const CEL_VAR_HIDDEN = new Set(['stdout', 'stderr', 'Stdout', 'Stderr']);

function filterCelVars(vars: Record<string, any>): Record<string, any> {
  const filtered: Record<string, any> = {};
  for (const [k, v] of Object.entries(vars)) {
    if (CEL_VAR_HIDDEN.has(k)) {
      filtered[k] = '(shown above)';
    } else {
      filtered[k] = v;
    }
  }
  return filtered;
}

function FixtureDetail({ ctx }: { ctx: FixtureContext }) {
  const hasComparison = ctx.expected !== undefined && ctx.actual !== undefined;

  return (
    <>
      {ctx.command && (
        <Section title="Command">
          <pre class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-3">
            <span class="text-gray-400">$ </span>{ctx.command}
          </pre>
          <div class="flex gap-4 mt-1 text-xs text-gray-500">
            {ctx.cwd && <span><iconify-icon icon="codicon:folder" class="mr-0.5" />{ctx.cwd}</span>}
            {ctx.exit_code !== undefined && ctx.exit_code !== 0 && (
              <span class="text-red-500">exit code: {ctx.exit_code}</span>
            )}
          </div>
        </Section>
      )}

      {ctx.cel_expression && (
        <Section title="CEL Expression">
          <pre class="text-sm whitespace-pre-wrap font-mono bg-purple-50 text-purple-800 rounded p-3">
            {ctx.cel_expression}
          </pre>
        </Section>
      )}

      {ctx.cel_vars && Object.keys(ctx.cel_vars).length > 0 && (
        <Section title="CEL Variables">
          <div class="bg-gray-50 rounded p-2 max-h-80 overflow-y-auto">
            <JsonView data={filterCelVars(ctx.cel_vars)} />
          </div>
        </Section>
      )}

      {hasComparison && (
        <Section title="Expected vs Actual">
          <div class="grid grid-cols-2 gap-2">
            <div>
              <div class="text-xs font-semibold text-gray-500 mb-1">Expected</div>
              <pre class="text-sm font-mono bg-green-50 text-green-800 rounded p-2 whitespace-pre-wrap max-h-40 overflow-y-auto">
                {typeof ctx.expected === 'object' ? JSON.stringify(ctx.expected, null, 2) : String(ctx.expected)}
              </pre>
            </div>
            <div>
              <div class="text-xs font-semibold text-gray-500 mb-1">Actual</div>
              <pre class="text-sm font-mono bg-red-50 text-red-800 rounded p-2 whitespace-pre-wrap max-h-40 overflow-y-auto">
                {typeof ctx.actual === 'object' ? JSON.stringify(ctx.actual, null, 2) : String(ctx.actual)}
              </pre>
            </div>
          </div>
        </Section>
      )}
    </>
  );
}

function GinkgoDetail({ ctx }: { ctx: GinkgoContext }) {
  return (
    <>
      {(ctx.suite_description || ctx.suite_path) && (
        <Section title="Suite">
          {ctx.suite_description && <div class="text-sm text-gray-700 font-medium">{ctx.suite_description}</div>}
          {ctx.suite_path && <div class="text-xs text-gray-500 font-mono mt-0.5">{ctx.suite_path}</div>}
        </Section>
      )}

      {ctx.failure_location && (
        <Section title="Failure Location">
          <span class="text-sm text-red-600 font-mono">{ctx.failure_location}</span>
        </Section>
      )}
    </>
  );
}

function GoTestDetail({ ctx }: { ctx: GoTestContext }) {
  return (
    <>
      {ctx.parent_test && (
        <Section title="Parent Test">
          <span class="text-sm text-gray-700 font-mono">{ctx.parent_test}</span>
        </Section>
      )}
      {ctx.import_path && (
        <Section title="Import Path">
          <span class="text-sm text-gray-700 font-mono">{ctx.import_path}</span>
        </Section>
      )}
    </>
  );
}

function Section({ title, children }: { title: string; children: any }) {
  return (
    <div>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">{title}</h3>
      {children}
    </div>
  );
}

function Stat({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div class="text-center">
      <div class={`text-lg font-bold ${color}`}>{value}</div>
      <div class="text-xs text-gray-500">{label}</div>
    </div>
  );
}

function MetaCard({ label, value }: { label: string; value: string }) {
  return (
    <div class="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2">
      <div class="text-xs uppercase tracking-wide text-gray-500">{label}</div>
      <div class="mt-1 text-sm font-medium text-gray-800">{value}</div>
    </div>
  );
}

function ViolationDetail({ v }: { v: Violation }) {
  const sevColor = v.severity === 'error' ? 'bg-red-100 text-red-800'
    : v.severity === 'warning' ? 'bg-yellow-100 text-yellow-800'
    : 'bg-blue-100 text-blue-800';
  return (
    <>
      <Section title="Severity">
        <span class={`inline-block text-xs font-semibold uppercase rounded px-2 py-0.5 ${sevColor}`}>
          {v.severity || 'info'}
        </span>
      </Section>
      {v.file && (
        <Section title="Location">
          <span class="text-sm font-mono text-gray-700">
            {v.file}{v.line ? `:${v.line}` : ''}{v.column ? `:${v.column}` : ''}
          </span>
        </Section>
      )}
      {v.rule?.method && (
        <Section title="Rule">
          <span class="text-sm font-mono text-gray-700">{v.rule.method}</span>
          {v.rule.description && <div class="text-xs text-gray-500 mt-0.5">{v.rule.description}</div>}
        </Section>
      )}
      {v.message && (
        <Section title="Message">
          <pre class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3">{v.message}</pre>
        </Section>
      )}
      {v.code && (
        <Section title="Code">
          <pre class="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 overflow-x-auto">{v.code}</pre>
        </Section>
      )}
    </>
  );
}

function groupViolationsByRule(violations: Violation[]): Array<{ rule: string; violations: Violation[] }> {
  const grouped = new Map<string, Violation[]>();
  for (const v of violations) {
    const key = v.rule?.method || '(no rule)';
    let bucket = grouped.get(key);
    if (!bucket) {
      bucket = [];
      grouped.set(key, bucket);
    }
    bucket.push(v);
  }
  return Array.from(grouped.keys()).sort().map(rule => ({
    rule,
    violations: grouped.get(rule)!.slice().sort((a, b) => (a.line || 0) - (b.line || 0)),
  }));
}

function LinterDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const lr = t.linter;
  const noFileViolations = t.noFileViolations || [];
  const linter = t.linterName || lr?.linter || '';
  const workDir = t.work_dir || lr?.work_dir || '';
  return (
    <>
      {lr && (
        <>
          <Section title="Status">
            <div class="flex gap-3 text-sm">
              <span class={lr.skipped ? 'text-gray-500' : lr.timed_out ? 'text-amber-600' : lr.success ? 'text-green-600' : 'text-red-600'}>
                {lr.skipped ? 'skipped' : lr.timed_out ? 'timed out' : lr.success ? 'success' : 'failed'}
              </span>
              <span class="text-gray-500">{(lr.violations || []).length} violations</span>
              {lr.file_count !== undefined && <span class="text-gray-500">{lr.file_count} files</span>}
              {lr.rule_count !== undefined && <span class="text-gray-500">{lr.rule_count} rules</span>}
            </div>
          </Section>
          {lr.command && (
            <Section title="Command">
              <pre class="text-xs text-gray-800 whitespace-pre-wrap break-all font-mono bg-gray-50 rounded p-3">{formatLinterCommand(lr)}</pre>
              {workDir && (
                <div class="mt-1 text-xs text-gray-500 font-mono">cwd: {workDir}</div>
              )}
            </Section>
          )}
          {linter && (
            <Section title="Actions">
              <div class="flex flex-wrap gap-1.5">
                <IgnoreButton
                  label={`Disable ${linter} entirely`}
                  title="Add {source} to .gavel.yaml"
                  req={{ source: linter, work_dir: workDir }}
                  onIgnore={onIgnore}
                  disabled={ignoreBusy}
                />
              </div>
            </Section>
          )}
          {lr.error && (
            <Section title="Error">
              <pre class="text-sm text-red-700 whitespace-pre-wrap font-mono bg-red-50 rounded p-3">{lr.error}</pre>
            </Section>
          )}
          {lr.raw_output && (
            <Section title="Raw output">
              <pre class="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto">{lr.raw_output}</pre>
            </Section>
          )}
        </>
      )}
      {noFileViolations.length > 0 && (
        <Section title="Violations Without File">
          <div class="space-y-3">
            {groupViolationsByRule(noFileViolations).map(({ rule, violations }) => (
              <div key={rule} class="border border-gray-200 rounded p-3 bg-gray-50">
                <div class="text-sm font-mono text-gray-700 mb-2">
                  {rule} <span class="text-xs text-gray-400">({violations.length})</span>
                </div>
                <div class="space-y-2">
                  {violations.map((v, i) => (
                    <ViolationRow
                      key={i}
                      v={v}
                      linter={linter}
                      file=""
                      workDir={workDir}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

// formatLinterCommand returns the shell-quoted argv that was executed. Args
// containing whitespace or shell metacharacters are single-quoted so the line
// can be copy-pasted into a terminal safely.
function formatLinterCommand(lr: LinterResult): string {
  if (!lr.command) return '';
  const parts = [shellQuote(lr.command), ...(lr.args || []).map(shellQuote)];
  return parts.join(' ');
}

function shellQuote(s: string): string {
  if (s === '') return "''";
  if (/^[A-Za-z0-9_./:=@%+,-]+$/.test(s)) return s;
  return `'${s.replace(/'/g, `'\\''`)}'`;
}

function folderPattern(path: string | undefined): string {
  if (!path) return '**';
  const trimmed = path.replace(/\/+$/, '');
  return trimmed ? `${trimmed}/**` : '**';
}

function collectFolderLintStats(folder: Test, lint: LinterResult[] | undefined): Array<{ linter: string; count: number; workDir?: string }> {
  const targetPath = folder.target_path || '';
  const counts = new Map<string, { linter: string; count: number; workDir?: string }>();
  for (const lr of lint || []) {
    if (folder.work_dir && lr.work_dir && lr.work_dir !== folder.work_dir) continue;
    for (const violation of lr.violations || []) {
      const rawFile = relPath(violation.file, lr.work_dir);
      if (!rawFile) continue;
      const matches = targetPath === '' ? true : rawFile === targetPath || rawFile.startsWith(`${targetPath}/`);
      if (!matches) continue;
      const current = counts.get(lr.linter);
      if (current) {
        current.count += 1;
      } else {
        counts.set(lr.linter, { linter: lr.linter, count: 1, workDir: lr.work_dir });
      }
    }
  }
  return Array.from(counts.values()).sort((a, b) => b.count - a.count || a.linter.localeCompare(b.linter));
}

function FolderLintDetail({ t, lint, onIgnore, ignoreBusy }: LintDetailProps & { lint?: LinterResult[] }) {
  const folderDisplay = t.file || t.name;
  const pattern = folderPattern(t.target_path);
  const scopedRule = t.ruleName || '';
  const scopedLinter = t.linterName || '';
  const linters = scopedRule ? [] : collectFolderLintStats(t, lint);
  const workDir = t.work_dir || (linters.length === 1 ? linters[0].workDir : '');
  return (
    <>
      <Section title="Folder">
        <div class="flex items-center gap-2">
          <span class="text-sm font-mono text-gray-700">{folderDisplay}</span>
          {t.target_path !== undefined && (
            <span class="text-xs text-gray-400">{pattern}</span>
          )}
        </div>
      </Section>
      <Section title="Actions">
        <div class="flex flex-wrap gap-1.5">
          {scopedRule ? (
            <>
              <IgnoreButton
                label={`Ignore ${scopedRule} in this path`}
                title="Add {source, rule, file} to .gavel.yaml"
                req={{ source: scopedLinter, rule: scopedRule, file: pattern, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy || !workDir}
              />
              <IgnoreButton
                label={`Disable ${scopedLinter} entirely`}
                title="Add {source} to .gavel.yaml"
                req={{ source: scopedLinter, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy}
                variant="subtle"
              />
            </>
          ) : (
            <>
              <IgnoreButton
                label="Ignore everything in this folder"
                title="Add {file} to .gavel.yaml"
                req={{ file: pattern, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy || !workDir}
              />
              {linters.map(({ linter }) => (
                <IgnoreButton
                  key={linter}
                  label={`Ignore ${linter} in this folder`}
                  title="Add {source, file} to .gavel.yaml"
                  req={{ source: linter, file: pattern, work_dir: workDir }}
                  onIgnore={onIgnore}
                  disabled={ignoreBusy || !workDir}
                  variant="subtle"
                />
              ))}
            </>
          )}
        </div>
      </Section>
      {scopedRule ? (
        <Section title="Scope">
          <div class="space-y-1 text-sm text-gray-700">
            <div><span class="text-gray-500">Linter:</span> <span class="font-mono">{scopedLinter}</span></div>
            <div><span class="text-gray-500">Rule:</span> <span class="font-mono">{scopedRule}</span></div>
          </div>
        </Section>
      ) : (
        <Section title="Linters">
          <div class="space-y-1">
            {linters.map(({ linter, count }) => (
              <div key={linter} class="flex items-center justify-between text-sm text-gray-700">
                <span class="font-mono">{linter}</span>
                <span class="text-xs text-gray-400">{count} violations</span>
              </div>
            ))}
            {linters.length === 0 && (
              <div class="text-sm text-gray-400">No violations found under this folder.</div>
            )}
          </div>
        </Section>
      )}
    </>
  );
}

interface LintDetailProps {
  t: Test;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
}

function IgnoreButton({
  label,
  title,
  req,
  onIgnore,
  disabled,
  variant,
}: {
  label: string;
  title: string;
  req: IgnoreRequest;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  disabled?: boolean;
  variant?: 'primary' | 'subtle';
}) {
  if (!onIgnore) return null;
  const base = 'text-xs px-2 py-0.5 rounded border transition-colors disabled:opacity-40 disabled:cursor-not-allowed inline-flex items-center gap-1';
  const cls = variant === 'subtle'
    ? `${base} border-gray-200 text-gray-500 hover:bg-gray-50`
    : `${base} border-yellow-300 bg-yellow-50 text-yellow-800 hover:bg-yellow-100`;
  return (
    <button
      class={cls}
      title={title}
      disabled={disabled}
      onClick={(e) => { e.stopPropagation(); void onIgnore(req); }}
    >
      <iconify-icon icon="codicon:eye-closed" class="text-xs" />
      {label}
    </button>
  );
}

function FileViolationsDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const linter = t.linterName || '';
  const file = t.file || '';
  const targetPath = t.target_path || file;
  const workDir = t.work_dir || '';
  const vs = t.violations || [];
  return (
    <>
      <Section title="File">
        <div class="flex items-center gap-2">
          <span class="text-sm font-mono text-gray-700">{file}</span>
          {vs.length > 0 && <span class="text-xs text-gray-400">({vs.length} violations)</span>}
        </div>
      </Section>
      {vs.length > 0 && (
        <>
          <Section title="Actions">
            <div class="flex flex-wrap gap-1.5">
              <IgnoreButton
                label={`Ignore all ${linter} in this file`}
                title="Add {source, file} to .gavel.yaml"
                req={{ source: linter, file: targetPath, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy}
              />
              <IgnoreButton
                label={`Disable ${linter} entirely`}
                title="Add {source} to .gavel.yaml"
                req={{ source: linter, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy}
                variant="subtle"
              />
            </div>
          </Section>
          <Section title="Violations">
            <div class="space-y-2">
              {vs.map((v, i) => (
                <ViolationRow
                  key={i}
                  v={v}
                  linter={linter}
                  file={targetPath}
                  workDir={workDir}
                  onIgnore={onIgnore}
                  ignoreBusy={ignoreBusy}
                />
              ))}
            </div>
          </Section>
        </>
      )}
    </>
  );
}

function RuleViolationsDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const linter = t.linterName || '';
  const file = t.file || '';
  const targetPath = t.target_path || file;
  const rule = t.ruleName || '';
  const workDir = t.work_dir || '';
  const vs = t.violations || [];
  const noFileViolations = t.noFileViolations || [];
  const total = lintNodeCount(t);
  return (
    <>
      <Section title="Rule">
        <span class="text-sm font-mono text-gray-700">{rule}</span>
        <span class="text-xs text-gray-400 ml-2">({total} violations)</span>
      </Section>
      {file && (
        <Section title="File">
          <span class="text-sm font-mono text-gray-700">{file}</span>
        </Section>
      )}
      <Section title="Actions">
        <div class="flex flex-wrap gap-1.5">
          {file && (
            <IgnoreButton
              label={`Ignore ${rule} in this file`}
              title="Add {source, rule, file} to .gavel.yaml"
              req={{ source: linter, rule, file: targetPath, work_dir: workDir }}
              onIgnore={onIgnore}
              disabled={ignoreBusy}
            />
          )}
          <IgnoreButton
            label={`Ignore rule ${rule} everywhere`}
            title="Add {source, rule} to .gavel.yaml"
            req={{ source: linter, rule, work_dir: workDir }}
            onIgnore={onIgnore}
            disabled={ignoreBusy}
          />
          <IgnoreButton
            label={`Disable ${linter} entirely`}
            title="Add {source} to .gavel.yaml"
            req={{ source: linter, work_dir: workDir }}
            onIgnore={onIgnore}
            disabled={ignoreBusy}
            variant="subtle"
          />
        </div>
      </Section>
      {vs.length > 0 && (
        <Section title="Violations">
          <div class="space-y-2">
            {vs.map((v, i) => (
              <ViolationRow
                key={i}
                v={v}
                linter={linter}
                file={v.raw_file || targetPath}
                workDir={workDir}
                onIgnore={onIgnore}
                ignoreBusy={ignoreBusy}
                showFile={!file}
              />
            ))}
          </div>
        </Section>
      )}
      {noFileViolations.length > 0 && (
        <Section title="Violations Without File">
          <div class="space-y-2">
            {noFileViolations.map((v, i) => (
              <ViolationRow
                key={i}
                v={v}
                linter={linter}
                file=""
                workDir={workDir}
              />
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

function ViolationRow({
  v, linter, file, workDir, onIgnore, ignoreBusy, showFile,
}: {
  v: Violation;
  linter: string;
  file: string;
  workDir: string;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
  showFile?: boolean;
}) {
  const sev = v.severity || 'error';
  const sevColor = sev === 'error' ? 'text-red-600'
    : sev === 'warning' ? 'text-yellow-600' : 'text-blue-500';
  const sevIcon = sev === 'error' ? 'codicon:error'
    : sev === 'warning' ? 'codicon:warning' : 'codicon:info';
  const rule = v.rule?.method || '';
  return (
    <div class="border border-gray-200 rounded p-2 bg-white">
      <div class="flex items-start gap-2">
        <iconify-icon icon={sevIcon} class={`${sevColor} text-base shrink-0 mt-0.5`} />
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2 flex-wrap text-xs text-gray-500 font-mono">
            {showFile && v.file && <span>{v.file}</span>}
            <span>
              :{v.line || 0}{v.column ? `:${v.column}` : ''}
            </span>
            {rule && <span class="text-gray-700">[{rule}]</span>}
          </div>
          {v.message && (
            <div class="text-sm text-gray-800 whitespace-pre-wrap mt-0.5">{v.message}</div>
          )}
          {v.code && (
            <pre class="text-xs font-mono bg-gray-50 rounded p-2 mt-1 overflow-x-auto">{v.code}</pre>
          )}
          {rule && file && (
            <div class="mt-1.5">
              <IgnoreButton
                label="Ignore this violation"
                title="Add {source, rule, file} to .gavel.yaml"
                req={{ source: linter, rule, file, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy}
                variant="subtle"
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ChildRow({ test: t }: { test: Test }) {
  return (
    <div class="flex items-center gap-1.5 py-0.5 text-sm">
      <iconify-icon icon={statusIcon(t)} class={`${statusColor(t)} text-base`} />
      <span class={`truncate ${t.failed ? 'text-red-700' : 'text-gray-700'}`}>{t.name}</span>
      {t.duration ? <span class="text-xs text-gray-400 ml-auto">{formatDuration(t.duration)}</span> : null}
    </div>
  );
}
