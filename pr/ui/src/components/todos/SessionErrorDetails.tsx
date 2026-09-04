import { useEffect, useMemo, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown, UiChevronRight, UiCopy, UiError } from '@flanksource/clicky-ui/icons';
import { copyText } from '../../clipboard';

export interface SessionError {
  source: string;
  message: string;
}

export async function sessionResponseError(response: Response, label: string): Promise<string> {
  const status = `HTTP ${response.status}${response.statusText ? ` ${response.statusText}` : ''}`;
  const raw = (await response.text()).trim();
  if (!raw) return `${label} (${status})`;

  try {
    const payload = JSON.parse(raw) as unknown;
    if (payload && typeof payload === 'object' && !Array.isArray(payload)) {
      const record = payload as Record<string, unknown>;
      const message = typeof record.error === 'string' && record.error.trim()
        ? record.error.trim()
        : raw;
      if (Object.keys(record).length === 1 && typeof record.error === 'string') {
        return `${label} (${status})\n${message}`;
      }
      return `${label} (${status})\n${message}\n\nResponse body:\n${JSON.stringify(payload, null, 2)}`;
    }
  } catch {
    return `${label} (${status})\n${raw}`;
  }
  return `${label} (${status})\n${raw}`;
}

function errorText(errors: SessionError[]): string {
  return errors.flatMap(({ source, message }, index) => [
    ...(index > 0 ? [''] : []),
    source,
    message,
  ]).join('\n');
}

export function SessionErrorDetails({ errors }: { errors: SessionError[] }) {
  const [expanded, setExpanded] = useState(false);
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>('idle');
  const details = useMemo(() => errorText(errors), [errors]);
  const ChevronIcon = expanded ? UiChevronDown : UiChevronRight;

  useEffect(() => {
    if (copyState === 'idle') return;
    const timeout = window.setTimeout(() => setCopyState('idle'), 1800);
    return () => window.clearTimeout(timeout);
  }, [copyState]);

  if (errors.length === 0) return null;

  async function copyDetails() {
    try {
      await copyText(details);
      setCopyState('copied');
    } catch {
      setCopyState('error');
    }
  }

  return (
    <div role="alert" className="shrink-0 border-b border-red-500/25 bg-red-500/10 px-3 py-2 text-xs text-red-700 dark:text-red-300">
      <div className="flex min-w-0 items-center gap-2">
        <UiError className="shrink-0 text-sm" />
        <span className="min-w-0 flex-1 font-medium">
          {errors.length === 1 ? 'Session error' : `${errors.length} session errors`}
        </span>
        <Button
          variant="ghost"
          type="button"
          onClick={() => setExpanded(value => !value)}
          aria-expanded={expanded}
          aria-label={expanded ? 'Hide details' : 'Show details'}
          className="inline-flex h-8 items-center gap-1 px-2 text-[11px] text-red-700 hover:bg-red-500/10 dark:text-red-300"
        >
          <ChevronIcon className="text-[10px]" />
          {expanded ? 'Hide details' : 'Show details'}
        </Button>
      </div>
      {expanded && (
        <div className="mt-2 rounded border border-red-500/20 bg-background/80 p-2 text-foreground">
          <div className="mb-1.5 flex items-center justify-end">
            <Button
              variant="ghost"
              type="button"
              onClick={copyDetails}
              aria-label="Copy error details"
              className="inline-flex h-8 items-center gap-1.5 px-2 text-[11px]"
            >
              {copyState === 'copied' ? <UiCheck className="text-emerald-600" /> : <UiCopy />}
              {copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy details'}
            </Button>
          </div>
          <pre className="max-h-56 overflow-auto whitespace-pre-wrap break-words font-mono text-[11px] leading-relaxed">{details}</pre>
        </div>
      )}
    </div>
  );
}
