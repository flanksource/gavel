import type { GavelAgentSession } from '@flanksource/clicky-ui/comments/gavel';
import type { TodoItem } from '../../types';
import { compareRecentTodos } from '../todos/todoNewParams';

export type CommentDestination =
  | { kind: 'session'; id: string; label: string; detail: string; session: GavelAgentSession }
  | { kind: 'todo'; id: string; label: string; detail: string; todo: TodoItem }
  | { kind: 'rule'; id: 'rule'; label: string; detail: string };

const RECENT_TODO_LIMIT = 8;

// commentDestinations orders where a new project comment can go: the scope's
// running agent sessions first, then the todos that have already run (most
// recent first), and last the project's own comment rule.
export function commentDestinations({ sessions, todos }: {
  sessions: readonly GavelAgentSession[];
  todos: readonly TodoItem[];
}): CommentDestination[] {
  const running = sessions.filter(session => session.status !== 'ended');
  const previous = todos.filter(todo => todo.lastRun || todo.sessionId).sort(compareRecentTodos).slice(0, RECENT_TODO_LIMIT);
  return [
    ...running.map((session): CommentDestination => ({
      kind: 'session',
      id: `session:${session.id}`,
      label: session.title || session.id,
      detail: [session.source, session.model, session.status].filter(Boolean).join(' · '),
      session,
    })),
    ...previous.map((todo): CommentDestination => ({
      kind: 'todo',
      id: `todo:${todo.ref}`,
      label: todo.title,
      detail: [todo.shortId || todo.ref, todo.status].join(' · '),
      todo,
    })),
    { kind: 'rule', id: 'rule', label: 'Project comment rule', detail: "Left as a thread, routed by the project's comment agent setting" },
  ];
}
