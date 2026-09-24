import { useState } from 'react';
import type { TodoDetailView } from '../../routes';
import { TodoDetail, type TodoDetailProps } from './TodoDetail';

// StatefulTodoDetail stands in for the router in tests: it owns the detail
// view the way App's route state does, and reports each change with its
// history mode.
export function StatefulTodoDetail({
  initialView = {},
  onViewChange,
  ...props
}: Omit<TodoDetailProps, 'view' | 'onViewChange'> & {
  initialView?: TodoDetailView;
  onViewChange?: TodoDetailProps['onViewChange'];
}) {
  const [view, setView] = useState(initialView);
  return (
    <TodoDetail
      {...props}
      view={view}
      onViewChange={(next, mode) => {
        setView(next);
        onViewChange?.(next, mode);
      }}
    />
  );
}
