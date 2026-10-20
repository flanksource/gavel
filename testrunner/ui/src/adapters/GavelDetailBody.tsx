import {
  AnsiHtml,
  JsonView,
  TestAttemptDetail,
  TestFailureDetail,
  TestOutput,
  sum,
  useTestRunner,
} from '@flanksource/clicky-ui/data';
import { Clicky, type ClickyDocument } from '@flanksource/clicky-ui/clicky';
import type { Test as ClickyTest } from '@flanksource/clicky-ui/data';
import { Field } from '../components/Field';
import type { GinkgoContext, GoTestContext, Test } from '../types';

function isLegacyClickyDocument(value: unknown): value is ClickyDocument {
  return typeof value === 'object' && value !== null && 'version' in value && 'node' in value;
}

function taskMeta(t: Test): { status?: string; type?: string } | null {
  if (t.framework !== 'task' || !t.context || typeof t.context !== 'object') return null;
  const ctx = t.context as Record<string, unknown>;
  return {
    status: typeof ctx.status === 'string' ? ctx.status : undefined,
    type: typeof ctx.type === 'string' ? ctx.type : undefined,
  };
}

function GoTestContextFields({ ctx }: { ctx: GoTestContext }) {
  return (
    <>
      {ctx.parent_test && <Field title="Parent Test"><span className="font-mono text-sm">{ctx.parent_test}</span></Field>}
      {ctx.import_path && <Field title="Import Path"><span className="font-mono text-sm">{ctx.import_path}</span></Field>}
    </>
  );
}

// Skip the Failure Location block when the structured failure already shows the
// same file:line, and the suite description when the Suite field shows it.
// suite_path (the filesystem location of the suite binary) is always useful.
function GinkgoContextFields({ ctx, failureLocation, suiteAbove }: { ctx: GinkgoContext; failureLocation?: string; suiteAbove?: string }) {
  const showLocation = ctx.failure_location && ctx.failure_location !== failureLocation;
  const showDescription = ctx.suite_description && ctx.suite_description !== suiteAbove;
  return (
    <>
      {(showDescription || ctx.suite_path) && (
        <Field title="Ginkgo suite">
          {showDescription && <div className="text-sm font-medium">{ctx.suite_description}</div>}
          {ctx.suite_path && <div className="font-mono text-xs text-muted-foreground">{ctx.suite_path}</div>}
        </Field>
      )}
      {showLocation && <Field title="Failure Location"><span className="font-mono text-sm text-red-600 dark:text-red-400">{ctx.failure_location}</span></Field>}
      {!!ctx.labels?.length && (
        <Field title="Labels">
          <div className="flex flex-wrap gap-1">
            {ctx.labels.map(label => <span key={label} className="rounded bg-muted px-2 py-0.5 text-xs">{label}</span>)}
          </div>
        </Field>
      )}
      {!!ctx.num_attempts && ctx.num_attempts > 1 && <Field title="Ginkgo attempts"><span className="text-sm">{ctx.num_attempts}</span></Field>}
      {!!ctx.report_entries?.length && (
        <Field title="Report entries">
          <div className="space-y-2">
            {ctx.report_entries.map((entry, index) => (
              <div key={`${entry.name}-${index}`} className="rounded border border-border bg-muted/50 p-2">
                <div className="text-sm font-medium">{entry.name}</div>
                {entry.file && <div className="font-mono text-xs text-muted-foreground">{entry.file}{entry.line ? `:${entry.line}` : ''}</div>}
                <div className="mt-1 break-words"><JsonView data={entry.value ?? entry.representation} /></div>
              </div>
            ))}
          </div>
        </Field>
      )}
    </>
  );
}

/**
 * The detail body of every node that is not a fixture step (fixture steps use
 * clicky's verification adapter): suite, package and command, the error, the
 * framework's own context (Go subtest parent, Ginkgo suite/labels/report
 * entries), the latest attempt's process metrics and stack, output and the
 * provider's structured detail.
 */
export function GavelDetailBody({ node: clickyNode }: { node: ClickyTest }) {
  const node = clickyNode as Test;
  const runner = useTestRunner();
  const attempt = node.attempts?.at(-1);
  const task = taskMeta(node);
  const counts = (node.children?.length ?? 0) > 0 ? sum(clickyNode) : null;
  const suite = node.suite?.join(' > ');
  const failure = node.failure_detail && node.failure_detail.kind && node.failure_detail.kind !== 'raw' ? node.failure_detail : undefined;
  const showFullMessage = !!node.message && !!failure && node.message !== failure.summary && node.message.includes('\n');

  return (
    <div className="space-y-4 p-density-4">
      {node.cached && <span className="inline-block rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">cached</span>}

      {suite && <Field title="Suite"><span className="text-sm">{suite}</span></Field>}
      {node.package_path && <Field title="Package"><span className="font-mono text-sm">{node.package_path}</span></Field>}

      {failure ? (
        <Field title="Error">
          <TestFailureDetail failure={failure} />
          {showFullMessage && (
            <details className="rounded bg-muted/50 p-2">
              <summary className="cursor-pointer text-xs font-semibold text-muted-foreground">Full message</summary>
              <AnsiHtml text={node.message!} className="mt-2 block max-h-64 overflow-y-auto whitespace-pre-wrap font-mono text-xs" />
            </details>
          )}
        </Field>
      ) : node.message ? (
        <Field title="Error">
          <AnsiHtml text={node.message} className="block max-h-64 overflow-y-auto whitespace-pre-wrap rounded bg-red-50 p-3 font-mono text-sm text-red-700 dark:bg-red-500/10 dark:text-red-300" />
        </Field>
      ) : null}

      {node.command && (
        <Field title="Command">
          <pre className="whitespace-pre-wrap rounded bg-muted p-2 font-mono text-sm"><span className="text-muted-foreground">$ </span>{node.command}</pre>
        </Field>
      )}

      {task && (task.status || task.type) && (
        <Field title="Task">
          <div className="flex gap-4 text-sm">
            {task.type && <span>type: <span className="font-mono">{task.type}</span></span>}
            {task.status && <span>status: <span className="font-mono">{task.status}</span></span>}
          </div>
        </Field>
      )}

      {node.framework === 'go test' && node.context && <GoTestContextFields ctx={node.context as GoTestContext} />}
      {node.framework === 'ginkgo' && node.context && (
        <GinkgoContextFields ctx={node.context as GinkgoContext} failureLocation={node.failure_detail?.location} suiteAbove={suite} />
      )}

      {attempt && (
        <Field title="Attempt">
          <TestAttemptDetail
            key={`${attempt.run_kind ?? ''}:${attempt.sequence}`}
            attempt={attempt}
            {...(runner.fetchProcessMetrics ? { fetchProcessMetrics: runner.fetchProcessMetrics } : {})}
          />
        </Field>
      )}

      <TestOutput node={clickyNode} />

      {counts && counts.total > 0 && (
        <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
          <span>{counts.total} total</span>
          {counts.passed > 0 && <span>{counts.passed} passed</span>}
          {counts.failed > 0 && <span>{counts.failed} failed</span>}
          {counts.warned > 0 && <span>{counts.warned} warned</span>}
          {counts.aborted > 0 && <span>{counts.aborted} aborted</span>}
          {counts.skipped > 0 && <span>{counts.skipped} skipped</span>}
          {counts.pending > 0 && <span>{counts.pending} pending</span>}
        </div>
      )}

      {node.detail != null && (
        <Field title="Detail">
          {isLegacyClickyDocument(node.detail) ? <Clicky data={node.detail} /> : <JsonView data={node.detail} />}
        </Field>
      )}
    </div>
  );
}
