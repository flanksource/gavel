import { useMemo } from 'react';
import { PromptForm, usePrompts, type Test } from '@flanksource/clicky-ui/data';
import { connectionLookupFetcher, promptsBase } from '../prompts';

export interface StepPendingFormProps {
  /** The fixture step waiting on its form; gavel labels the step's prompt `node` with this name. */
  node: Pick<Test, 'name'>;
  /** The run's prompt owner (metadata.prompt_owner). */
  owner: string | undefined;
  scope?: { runId?: string };
}

/**
 * The form a waiting fixture step (a `params:` or `form` step) asks, answerable
 * in the step's detail panel. It is the same pending prompt the header's
 * PromptIndicator lists, so answering here resolves it there too.
 */
export function StepPendingForm({ node, owner, scope = {} }: StepPendingFormProps) {
  const lookupFetcher = useMemo(() => connectionLookupFetcher(scope), [scope.runId]);
  const basePath = promptsBase(scope);
  const { pending } = usePrompts({ basePath, owner, labels: { node: node.name }, state: 'pending', enabled: !!owner });
  if (!owner) return <p className="text-xs text-muted-foreground">No pending form for this step: the run publishes no prompts</p>;
  if (pending.length === 0) return <p className="text-xs text-muted-foreground">No pending form for this step</p>;
  return (
    <div className="space-y-density-3">
      {pending.map(prompt => (
        <PromptForm key={prompt.id} prompt={prompt} basePath={basePath} lookupFetcher={lookupFetcher} idPrefix={`step-prompt-${prompt.id}`} />
      ))}
    </div>
  );
}
