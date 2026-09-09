import { Field } from '@flanksource/clicky-ui/components';

export function TodoRunWarnings({ warnings }: { warnings?: string[] }) {
  if (!warnings?.length) return null;
  return (
    <Field label="Preflight warnings">
      <ul aria-label="Preflight warnings" className="list-disc space-y-1 pl-4 text-xs text-muted-foreground">
        {warnings.map((warning, index) => <li key={`${index}:${warning}`}>{warning}</li>)}
      </ul>
    </Field>
  );
}
