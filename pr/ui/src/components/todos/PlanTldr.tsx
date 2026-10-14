import { useQuery } from '@tanstack/react-query';
import { Markdown } from '@flanksource/clicky-ui/data';
import type { TodoItem } from '../../types';
import { fetchJSON } from '../../query';
import { todoQuery } from './format';
import { todoQueryKeys } from './todoQueries';

// TLDR_HOOK is the captain plan.revision hook prompt (plan-tldr.prompt) whose
// block the server lifts out of the plan into `blocks`.
export const TLDR_HOOK = 'plan-tldr';

export interface PlanResponse {
  found: boolean;
  path?: string;
  // content is the authored plan body, without captain hook blocks.
  content?: string;
  // blocks holds the captain hook blocks keyed by hook name, e.g. plan-tldr.
  blocks?: Record<string, string>;
  onDisk?: boolean;
  slug?: string;
  ref?: string;
  version?: number;
  todo?: TodoItem;
}

export function usePlanQuery(dir: string, ref: string, enabled: boolean) {
  return useQuery({
    queryKey: todoQueryKeys.plan(dir, ref),
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams(todoQuery(dir));
      params.set('ref', ref);
      return fetchJSON<PlanResponse>({
        url: `/api/todos/session/plan?${params.toString()}`,
        signal,
        context: 'Plan request failed',
      });
    },
    enabled: enabled && !!ref,
    staleTime: 5_000,
  });
}

// PlanTldrCallout renders the plan's TL;DR hook block, or nothing when the plan
// has none.
export function PlanTldrCallout({ plan }: { plan?: PlanResponse }) {
  const tldr = plan?.blocks?.[TLDR_HOOK];
  if (!tldr) return null;
  return (
    <div role="note" aria-label="Plan TL;DR" className="shrink-0 rounded-md border border-sky-500/30 bg-sky-500/5 px-3 py-2">
      <Markdown text={tldr} className="text-sm" />
    </div>
  );
}

// PlanTldr fetches the plan for one todo and renders its TL;DR callout.
export function PlanTldr({ dir, refId }: { dir: string; refId: string }) {
  const planQuery = usePlanQuery(dir, refId, true);
  return <PlanTldrCallout plan={planQuery.data} />;
}
