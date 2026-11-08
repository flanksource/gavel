import { useMemo } from 'react';
import { PromptIndicator } from '@flanksource/clicky-ui/data';
import { connectionLookupFetcher, promptsBase } from '../prompts';

export interface RunPromptsProps {
  /** The run's prompt owner (metadata.prompt_owner); nothing shows without one. */
  owner?: string | undefined;
  scope?: { runId?: string };
}

/**
 * The run's fixture forms as one header indicator: "N pending" opens a panel
 * listing them, and a pending form opens in a dialog, where Submit answers with
 * the values and Cancel answers `{cancelled: true}`. A `format: connection`
 * field resolves against the server's connection catalog. Each form's dialog
 * opens on its own when the form becomes pending, so a run paused at a
 * `params:` or `form` step asks without a click.
 */
export function RunPrompts({ owner, scope = {} }: RunPromptsProps) {
  const lookupFetcher = useMemo(() => connectionLookupFetcher(scope), [scope.runId]);
  if (!owner) return null;
  return <PromptIndicator basePath={promptsBase(scope)} owner={owner} lookupFetcher={lookupFetcher} autoOpen />;
}
