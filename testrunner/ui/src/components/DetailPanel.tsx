import type { Test, FixtureContext, GinkgoContext, GoTestContext } from '../types';
import { statusIcon, statusColor, formatDuration, sum, frameworkIcon } from '../utils';
import { JsonView } from './JsonView';
import { AnsiHtml } from './AnsiHtml';

interface Props {
  test: Test | null;
}

export function DetailPanel({ test: t }: Props) {
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

  return (
    <div class="p-5 space-y-4">
      {/* Header */}
      <div class="flex items-start gap-2">
        <iconify-icon icon={statusIcon(t)} class={`${statusColor(t)} text-2xl shrink-0 mt-0.5`} />
        <div class="min-w-0">
          <h2 class="text-lg font-bold text-gray-900 break-words">{t.name}</h2>
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
            ) : null}
            {t.file && (
              <span class="text-xs text-gray-500 font-mono">
                {t.file}{t.line ? `:${t.line}` : ''}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Summary for containers */}
      {s && s.total > 0 && (
        <div class="flex gap-4 text-sm border rounded-lg p-3 bg-gray-50">
          <Stat label="Total" value={s.total} color="text-gray-700" />
          <Stat label="Passed" value={s.passed} color="text-green-600" />
          <Stat label="Failed" value={s.failed} color="text-red-600" />
          {s.skipped > 0 && <Stat label="Skipped" value={s.skipped} color="text-yellow-600" />}
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
        <Section title="Tests">
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

function ChildRow({ test: t }: { test: Test }) {
  return (
    <div class="flex items-center gap-1.5 py-0.5 text-sm">
      <iconify-icon icon={statusIcon(t)} class={`${statusColor(t)} text-base`} />
      <span class={`truncate ${t.failed ? 'text-red-700' : 'text-gray-700'}`}>{t.name}</span>
      {t.duration ? <span class="text-xs text-gray-400 ml-auto">{formatDuration(t.duration)}</span> : null}
    </div>
  );
}
