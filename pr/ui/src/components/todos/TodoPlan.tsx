import { lazy, Suspense, useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiSave } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import type { TodoItem } from '../../types';
import { fetchJSON } from '../../query';
import { inputClass, todoQuery } from './format';
import { setTodoCaches, todoMutationJSON } from './todoMutations';
import { todoQueryKeys } from './todoQueries';

// MdxEditorField lazily pulls in the heavy @mdxeditor/editor (the same markdown
// field the run dialog uses), so it is code-split and rendered under Suspense with a
// plain-textarea fallback.
const MdxEditorField = lazy(() =>
  import('@flanksource/clicky-ui/mdx-editor').then(m => ({ default: m.MdxEditorField })),
);

interface PlanResponse {
  found: boolean;
  path?: string;
  content?: string;
  onDisk?: boolean;
  slug?: string;
  ref?: string;
  version?: number;
  todo?: TodoItem;
}

// TodoPlan shows the latest immutable Captain revision selected on the issue.
// Human edits append another database revision; they never rewrite an agent's
// local plan file.
export function TodoPlan({
  dir,
  todo,
  active,
  onChanged,
}: {
  dir: string;
  todo: TodoItem;
  active: boolean;
  onChanged?: (todo: TodoItem) => void;
}) {
  const queryClient = useQueryClient();
  const [loaded, setLoaded] = useState(''); // server content — the save baseline
  const [draft, setDraft] = useState('');
  const queryKey = todoQueryKeys.plan(dir, todo.ref);
  const planQuery = useQuery({
    queryKey,
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams(todoQuery(dir));
      params.set('ref', todo.ref);
      return fetchJSON<PlanResponse>({
        url: `/api/todos/session/plan?${params.toString()}`,
        signal,
        context: 'Plan request failed',
      });
    },
    enabled: active && !!todo.ref,
    staleTime: 5_000,
  });
  const saveMutation = useMutation({
    mutationKey: ['todos', 'session', 'plan', 'save', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: (content: string) => todoMutationJSON<PlanResponse>(
      `/api/todos/session/plan?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, version: todo.version, content }),
      },
      'Plan save failed',
    ),
    onSuccess: async (data, content) => {
      const next = data.content ?? content;
      queryClient.setQueryData<PlanResponse>(queryKey, data);
      setLoaded(next);
      setDraft(next);
      if (data.todo) {
        await setTodoCaches(queryClient, dir, data.todo);
        onChanged?.(data.todo);
      }
    },
  });

  useEffect(() => {
    saveMutation.reset();
    const content = active ? planQuery.data?.content ?? '' : '';
    setLoaded(content);
    setDraft(content);
  }, [active, dir, planQuery.data, saveMutation.reset, todo.ref]);

  if (planQuery.isFetching && !planQuery.data) {
    return (
      <div className="flex items-center gap-2 px-4 py-3 text-sm text-muted-foreground">
        <Spinner />
        Loading plan
      </div>
    );
  }
  if (!planQuery.data?.found) {
    if (planQuery.error) return <div role="alert" className="px-4 py-3 text-sm text-red-600">{planQuery.error.message}</div>;
    return <PlanEmpty message="No plan yet. Run this todo in Plan mode to produce one." />;
  }

  const dirty = draft !== loaded;
  const saving = saveMutation.isPending;
  const path = planQuery.data.path ?? '';
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2 px-4 py-3">
      <div className="flex shrink-0 items-center justify-between gap-2">
        <div className="min-w-0 truncate text-xs text-muted-foreground" title={path}>
          {path || 'PostgreSQL plan revision'}
        </div>
        <Button size="sm" variant="outline" disabled={!dirty || saving} onClick={() => saveMutation.mutate(draft)}>
          {saving ? <Spinner /> : <UiSave />}
          Save
        </Button>
      </div>
      {saveMutation.error && <div className="shrink-0 text-xs text-red-600">{saveMutation.error.message}</div>}
      <div className="min-h-0 flex-1 overflow-auto">
        <Suspense
          fallback={
            <textarea
              className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`}
              value={draft}
              onChange={e => setDraft(e.currentTarget.value)}
            />
          }
        >
          <MdxEditorField value={draft} onChange={setDraft} className="min-h-[16rem]" />
        </Suspense>
      </div>
    </div>
  );
}

function PlanEmpty({ message }: { message: string }) {
  return <div className="px-4 py-3 text-sm text-muted-foreground">{message}</div>;
}
