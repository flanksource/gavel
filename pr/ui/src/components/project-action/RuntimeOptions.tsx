import { useQuery } from '@tanstack/react-query';
import { RuntimeBar, runtimeModeError } from '@flanksource/clicky-ui/ai';
import { Button } from '@flanksource/clicky-ui/components';
import { chatModelsQuery, chatRuntimesQuery } from '../oneShotQueries';
import { commitRuntimeValue, withCommitRuntime } from './runtime';

interface Props {
  value: Record<string, unknown>;
  onChange: (value: Record<string, unknown>) => void;
}

export function RuntimeOptions({ value, onChange }: Props) {
  const models = useQuery(chatModelsQuery());
  const families = useQuery(chatRuntimesQuery());
  const catalogError = models.error ?? families.error;
  if (catalogError) return (
    <div className="flex items-center gap-2 text-sm">
      <span role="alert">{catalogError.message}</span>
      <Button type="button" variant="ghost" onClick={() => {
        void models.refetch();
        void families.refetch();
      }}>Retry AI catalogs</Button>
    </div>
  );
  if (!models.data || !families.data) return <p className="text-sm text-muted-foreground">Loading AI options…</p>;

  let runtime;
  try {
    runtime = commitRuntimeValue(value);
    const error = runtimeModeError(families.data, runtime.mode);
    if (error) throw new Error(error);
  } catch (cause) {
    return (
      <div className="flex items-center gap-2 text-sm">
        <span role="alert">{cause instanceof Error ? cause.message : 'Invalid commit AI options'}</span>
        <Button type="button" variant="ghost" onClick={() => onChange(withCommitRuntime(value, {}))}>Reset AI options</Button>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      <RuntimeBar
        ariaLabel="Commit AI runtime"
        value={runtime}
        onChange={next => onChange(withCommitRuntime(value, next))}
        models={models.data}
        families={families.data}
      />
      <p className="text-xs text-muted-foreground">Unspecified options inherit the configured commit AI runtime.</p>
    </div>
  );
}
