import { Button } from '@flanksource/clicky-ui/components';
import { UiError } from '@flanksource/clicky-ui/icons';
import type { SessionStats } from '../../types';

export function VerificationFailureStatus({ stats, compact = false, onOpen }: {
  stats: SessionStats | undefined;
  compact?: boolean;
  onOpen?: (() => void) | undefined;
}) {
  if (stats?.state !== 'verification_failed') return null;
  if (!stats.verification?.ran || stats.verification.state !== 'failed') {
    throw new Error('Verification failure status requires a failed verification report');
  }
  return (
    <div
      role="status"
      title={stats.verification.reason}
      className={`flex shrink-0 items-center gap-2 border-red-500/30 bg-red-500/10 text-xs text-red-700 [[data-theme=dark]_&]:text-red-400 ${compact ? 'rounded border px-1.5 py-0.5' : 'border-b px-3 py-2'}`}
    >
      <UiError className="shrink-0 text-sm" />
      <div className="min-w-0">
        <span className="font-medium">Verification failed</span>
        {!compact && <p className="mt-0.5">{stats.verification.reason}</p>}
      </div>
      {onOpen && <Button variant="ghost" size="sm" type="button" onClick={onOpen} className="ml-auto shrink-0 text-xs">View verification</Button>}
    </div>
  );
}
