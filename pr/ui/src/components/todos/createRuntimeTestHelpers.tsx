import { Button } from '@flanksource/clicky-ui/components';
import type { TodoRunOptions } from '../../types';
import type { RunContext } from './providers';

export const CREATE_MODEL = 'acme-model';
export const CREATE_RUNTIME = { model: CREATE_MODEL, mode: 'api', effort: 'high' };
export const CREATE_RUNTIME_CONTEXT: RunContext = {
  modes: [{ id: 'api', label: 'API', provider: 'acme', agent: 'codex', defaultModel: CREATE_MODEL, driver: 'api', mechanisms: [], models: [{ id: CREATE_MODEL, label: 'Acme model', provider: 'acme', reasoning: true }] }],
  runtimes: [], models: [{ id: CREATE_MODEL, label: 'Acme model', provider: 'acme', reasoning: true }], efforts: ['low', 'high'], tools: [],
  lifecycle: { steps: [] },
};

export function CreateRuntimeBarStub({ action, options, onChange, disabled }: {
  action: string;
  options: TodoRunOptions;
  onChange: (value: TodoRunOptions) => void;
  disabled?: boolean;
}) {
  return <Button type="button" disabled={disabled} aria-label={`${action} runtime`} onClick={() => onChange({ ...options, spec: { ...options.spec, ...CREATE_RUNTIME } })}>Runtime</Button>;
}
