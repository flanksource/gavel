import { useMemo, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Field, Modal } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { projectScope } from '@flanksource/clicky-ui/comments/gavel';
import { randomUUID } from '@flanksource/clicky-ui/utils';
import type { Project } from '../../types';
import { inputClass } from '../todos/format';
import { useUpdateTodoMutation } from '../todos/todoMutations';
import { todoListQueryOptions } from '../todos/todoQueries';
import { commentDestinations, type CommentDestination } from './commentDestinations';
import { commentClient, commentListKey, commentSessionsQuery } from './queries';

const groups: { kind: CommentDestination['kind']; title: string }[] = [
  { kind: 'session', title: 'Running sessions' },
  { kind: 'todo', title: 'Previous todos' },
  { kind: 'rule', title: 'Other' },
];

// ProjectCommentDialog comments on a file or folder of a project and hands it to
// a running agent session, appends it to a todo that has already run, or leaves
// it as a thread for the project's comment rule — in that order of preference.
export function ProjectCommentDialog({ project, path, folder, onClose }: {
  project: Project;
  path: string;
  folder: boolean;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const scope = projectScope(project.name);
  const sessions = useQuery(commentSessionsQuery(scope));
  const todos = useQuery(todoListQueryOptions(project.dir));
  const destinations = useMemo(
    () => commentDestinations({ sessions: sessions.data ?? [], todos: todos.data?.items ?? [] }),
    [sessions.data, todos.data],
  );
  const loading = sessions.isPending || todos.isPending;
  const [selectedId, setSelectedId] = useState<string>();
  const selected = destinations.find(destination => destination.id === selectedId) ?? (loading ? undefined : destinations[0]);
  const [body, setBody] = useState('');
  const requestId = useRef(randomUUID());
  const label = `${path}${folder ? '/' : ''}`;
  const updateTodo = useUpdateTodoMutation(project.dir, `Failed to add the comment on ${label} to the todo`);
  const send = useMutation({
    mutationFn: async (destination: CommentDestination) => {
      const text = body.trim();
      if (destination.kind === 'todo') {
        const { ref } = destination.todo;
        await updateTodo.mutateAsync({
          ref,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ref, comment: `On \`${label}\`:\n\n${text}` }),
        });
        return;
      }
      await commentClient(scope).create({ request: {
        target: { uri: `${scope}/files/${path}`, scopeUri: scope, kind: folder ? 'folder' : 'file', label },
        bodyMarkdown: text,
        requestId: requestId.current,
        ...(destination.kind === 'session' ? { deliver: { kind: 'session' as const, sessionId: destination.session.id } } : {}),
      } });
      await queryClient.invalidateQueries({ queryKey: commentListKey(scope) });
    },
    onSuccess: onClose,
  });
  const canSend = !!selected && body.trim() !== '' && !send.isPending;

  return <Modal open title={`Comment on ${label}`} onClose={() => { if (!send.isPending) onClose(); }}
    footer={<div className="flex justify-end gap-2">
      <Button variant="outline" disabled={send.isPending} onClick={onClose}>Cancel</Button>
      <Button disabled={!canSend} loading={send.isPending} onClick={() => { if (selected && canSend) send.mutate(selected); }}>
        {selected?.kind === 'todo' ? 'Add to todo' : selected?.kind === 'session' ? 'Send to session' : 'Comment'}
      </Button>
    </div>}>
    <div className="space-y-4 text-sm">
      <Field label="Comment">
        <textarea aria-label="Comment" autoFocus className={`${inputClass} min-h-24`} value={body} disabled={send.isPending} onChange={event => setBody(event.currentTarget.value)} />
      </Field>
      <fieldset disabled={send.isPending} className="space-y-3">
        <legend className="font-medium">Send to</legend>
        {loading && <p role="status" className="text-muted-foreground">Loading sessions and todos…</p>}
        {sessions.error && <ActionError error={sessions.error} title="Could not load agent sessions" />}
        {todos.error && <ActionError error={todos.error} title="Could not load todos" />}
        {!loading && groups.map(group => {
          const items = destinations.filter(destination => destination.kind === group.kind);
          if (items.length === 0) return null;
          return <div key={group.kind} role="group" aria-label={group.title} className="space-y-1">
            <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{group.title}</h3>
            {items.map(destination => (
              <label key={destination.id} className="flex items-start gap-2 rounded px-2 py-1 hover:bg-muted/40">
                <input type="radio" name="comment-destination" className="mt-1" checked={selected?.id === destination.id} onChange={() => setSelectedId(destination.id)} />
                <span className="min-w-0">
                  <span className="block truncate">{destination.label}</span>
                  <span className="block truncate text-xs text-muted-foreground">{destination.detail}</span>
                </span>
              </label>
            ))}
          </div>;
        })}
      </fieldset>
      {send.error && <ActionError error={send.error} />}
    </div>
  </Modal>;
}
