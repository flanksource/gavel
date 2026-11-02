import { useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { TodoRunOptions } from '../../types';
import { defaultRunOptionsForAction, TodoRunContextError, useTodoRunContext } from './run';
import { TodoRunRuntimeBar } from './TodoRunActionButton';

export const REVISE_SAME_AGENT_LABEL = 'Same agent as the plan';
export const REVISE_NEW_SESSION_HINT = 'Starts a new planning session with the current plan and your feedback.';

// ReviseRuntimePicker lets the reviewer optionally re-plan with a different
// agent than the one that made the plan. No value means "same agent": the
// request omits `options` and the server resumes the plan run's own runtime
// and session. A value is set only when the reviewer actually picks a runtime
// in the shared run-dialog RuntimeBar — opening the picker alone changes
// nothing — and the server decides whether the pick starts a fresh session.
//
// Open/closed is local state, so mount it with a key tied to the todo it
// belongs to: switching todos remounts it back to "same agent".
export function ReviseRuntimePicker({
  dir,
  value,
  onChange,
  disabled,
}: {
  dir: string;
  value: TodoRunOptions | undefined;
  onChange: (options: TodoRunOptions | undefined) => void;
  disabled?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const { context, loading, error } = useTodoRunContext({ dir, enabled: open });
  const unavailable = disabled || loading;

  if (!open) {
    return (
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <span>Agent: {REVISE_SAME_AGENT_LABEL}</span>
        <Button type="button" variant="ghost" size="sm" disabled={disabled} onClick={() => setOpen(true)} className="h-6 px-1.5 text-xs">
          Change
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
        <span>Agent:</span>
        {context && !error && (
          <TodoRunRuntimeBar
            action="plan"
            context={context}
            options={value ?? defaultRunOptionsForAction('plan', context)}
            disabled={unavailable}
            onChange={onChange}
          />
        )}
        {loading && !context && <span>Loading run providers…</span>}
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={disabled}
          onClick={() => {
            onChange(undefined);
            setOpen(false);
          }}
          className="h-6 px-1.5 text-xs"
        >
          Use {REVISE_SAME_AGENT_LABEL.toLowerCase()}
        </Button>
      </div>
      <TodoRunContextError error={error} />
      {value && <div className="text-[11px] text-muted-foreground">{REVISE_NEW_SESSION_HINT}</div>}
    </div>
  );
}
