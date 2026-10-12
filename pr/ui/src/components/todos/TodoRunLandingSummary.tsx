import { UiGitMerge, UiGitPr, UiLinkExternal } from '@flanksource/clicky-ui/icons';
import type { TodoRunLanding } from '../../types';
import { shortSha } from './runWorkspace';

export type TodoRunLandingFields = Pick<TodoRunLanding, 'via' | 'targetBranch' | 'landedSha' | 'prNumber' | 'prUrl'>;

/**
 * Reads a run_landed event payload (gavel's landingPayload) into landing
 * fields, or a string naming what is wrong with it, so a malformed payload is
 * shown as such rather than rendered as a partial landing.
 */
export function landingFromPayload(payload: Record<string, unknown> | undefined): TodoRunLandingFields | string {
  const { via, targetBranch, landedSha, prNumber, prUrl } = payload ?? {};
  if (via !== 'merge' && via !== 'pr') return `unknown via ${JSON.stringify(via)}`;
  if (typeof targetBranch !== 'string' || !targetBranch) return 'no target branch';
  if (typeof landedSha !== 'string' || !landedSha) return 'no landed sha';
  if (via === 'pr' && (typeof prNumber !== 'number' || typeof prUrl !== 'string')) return 'a pr landing without its pull request';
  return {
    via, targetBranch, landedSha,
    prNumber: typeof prNumber === 'number' ? prNumber : undefined,
    prUrl: typeof prUrl === 'string' ? prUrl : undefined,
  };
}

/** How the run's commits were landed: merged into a branch, or opened as a PR. */
export function TodoRunLandingSummary({ landing }: { landing: TodoRunLandingFields }) {
  const Icon = landing.via === 'pr' ? UiGitPr : UiGitMerge;
  return (
    <span className="inline-flex min-w-0 items-center gap-1.5 text-[11px] text-emerald-700 [[data-theme=dark]_&]:text-emerald-400">
      <Icon className="shrink-0 text-xs" aria-hidden="true" />
      {landing.via === 'pr' ? (
        <span className="inline-flex items-center gap-1">
          PR
          <a
            href={landing.prUrl}
            target="_blank"
            rel="noreferrer"
            title={`Open pull request #${landing.prNumber}`}
            className="inline-flex items-center gap-0.5 hover:underline"
          >
            #{landing.prNumber}
            <UiLinkExternal className="shrink-0 text-[10px]" aria-hidden="true" />
          </a>
          into
        </span>
      ) : (
        <span>Merged into</span>
      )}
      <span className="font-mono">{landing.targetBranch}</span>
      <span className="font-mono text-muted-foreground" title={landing.landedSha}>{shortSha(landing.landedSha)}</span>
    </span>
  );
}
