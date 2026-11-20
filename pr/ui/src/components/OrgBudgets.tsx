// The organization's budgets: each with how much of it is used, in its window or held now, which
// every member sees and only an admin sets or removes.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiTrash } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';

type Resource = 'cpu' | 'memory' | 'tokens' | 'usd';
type Kind = 'cumulative' | 'concurrent';

// BudgetStatus is a budget and how much of it is used: since its window started for a cumulative
// budget, held by running runs now for a concurrent one.
export interface BudgetStatus {
  resource: Resource;
  kind: Kind;
  amount: number;
  window?: string;
  used: number;
  since?: string;
}

interface Budgets {
  budgets: BudgetStatus[];
  canEdit: boolean;
}

const resourceNames: Record<Resource, string> = { cpu: 'CPU', memory: 'Memory', tokens: 'Tokens', usd: 'Model cost' };

const windows = [
  { value: 'now/M', label: 'This month', period: 'this month' },
  { value: 'now-30d', label: 'Last 30 days', period: 'in the last 30 days' },
  { value: 'now/w', label: 'This week', period: 'this week' },
  { value: 'now/d', label: 'Today', period: 'today' },
];

const count = new Intl.NumberFormat('en-US');
const usd = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });
const inputClass = 'rounded border border-border bg-background px-2 py-1 text-sm';

// figure shows n with at most one decimal.
function figure(n: number): string {
  return String(Number(n.toFixed(1)));
}

function unit(resource: Resource, kind: Kind): string {
  if (resource === 'cpu') return kind === 'concurrent' ? 'cores' : 'core-hours';
  if (resource === 'memory') return kind === 'concurrent' ? 'GiB' : 'GiB-hours';
  return resource === 'tokens' ? 'tokens' : 'USD';
}

function budgetName(budget: BudgetStatus): string {
  const period = budget.kind === 'concurrent'
    ? 'at once'
    : windows.find(w => w.value === budget.window)?.period ?? `since ${budget.window}`;
  return `${resourceNames[budget.resource]} ${period}`;
}

function amounts(budget: BudgetStatus): string {
  if (budget.resource === 'usd') return `${usd.format(budget.used)} of ${usd.format(budget.amount)}`;
  if (budget.resource === 'tokens') return `${count.format(budget.used)} of ${count.format(budget.amount)} tokens`;
  return `${figure(budget.used)} of ${figure(budget.amount)} ${unit(budget.resource, budget.kind)}`;
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function OrgBudgets() {
  const queryClient = useQueryClient();
  const budgets = useQuery({
    queryKey: queryKeys.budgets(),
    queryFn: ({ signal }) => fetchJSON<Budgets>({ url: '/api/budgets', signal, context: 'load the organization\'s budgets' }),
  });
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.budgets() });
  const remove = useMutation({
    mutationFn: (budget: BudgetStatus) => mutationJSON<unknown>({
      url: `/api/budgets/${budget.resource}/${budget.kind}`, method: 'DELETE', context: `remove the ${budgetName(budget)} budget`,
    }),
    onSuccess: refresh,
  });
  const data = budgets.data;
  const error = budgets.error ?? remove.error;

  return (
    <section className="flex flex-col gap-2">
      <h3 className="font-medium">Budgets</h3>
      <p className="text-xs text-muted-foreground">
        A used-up budget refuses new runs and stops the ones running, saying why. CPU and memory budgets hold back every
        run; token and model cost budgets hold back agent runs only, as only they call models. A run that would take the
        organization over a budget at once waits its turn for room, and fails if none comes in time.
      </p>
      {data && data.budgets.length === 0 && <p>No budgets: this organization's runs and model calls are not capped.</p>}
      {data && data.budgets.map(budget => (
        <div key={`${budget.resource}/${budget.kind}`} className="flex items-center justify-between gap-2 rounded border border-border px-3 py-1">
          <span>{`${budgetName(budget)}: ${amounts(budget)}${budget.used >= budget.amount ? ' · used up' : ''}`}</span>
          {data.canEdit && (
            <Button
              variant="ghost"
              size="icon"
              type="button"
              aria-label={`Remove the ${budgetName(budget)} budget`}
              disabled={remove.isPending}
              onClick={() => remove.mutate(budget)}
            >
              <UiTrash />
            </Button>
          )}
        </div>
      ))}
      {data && !data.canEdit && <p className="text-muted-foreground">Only an organization admin can change budgets; ask one to set them.</p>}
      {data?.canEdit && <BudgetForm onSet={refresh} />}
      {error && <p role="alert" className="text-destructive">{errorText(error)}</p>}
    </section>
  );
}

function BudgetForm({ onSet }: { onSet: () => Promise<void> }) {
  const [resource, setResource] = useState<Resource>('cpu');
  const [chosenKind, setKind] = useState<Kind>('cumulative');
  const [amount, setAmount] = useState('');
  const [budgetWindow, setBudgetWindow] = useState(windows[0].value);
  const held = resource === 'cpu' || resource === 'memory';
  const kind: Kind = held ? chosenKind : 'cumulative';
  const set = useMutation({
    mutationFn: () => mutationJSON<BudgetStatus>({
      url: `/api/budgets/${resource}/${kind}`, method: 'PUT', context: 'set the budget',
      body: kind === 'cumulative' ? { amount: Number(amount), window: budgetWindow } : { amount: Number(amount) },
    }),
    onSuccess: () => {
      setAmount('');
      return onSet();
    },
  });

  return (
    <form
      className="flex flex-wrap items-end gap-3"
      onSubmit={event => {
        event.preventDefault();
        set.mutate();
      }}
    >
      <label className="flex flex-col gap-1">
        <span>Resource</span>
        <select className={inputClass} aria-label="Resource" value={resource} onChange={event => setResource(event.target.value as Resource)}>
          {(Object.keys(resourceNames) as Resource[]).map(r => <option key={r} value={r}>{resourceNames[r]}</option>)}
        </select>
      </label>
      {held && (
        <label className="flex flex-col gap-1">
          <span>Kind</span>
          <select className={inputClass} aria-label="Kind" value={kind} onChange={event => setKind(event.target.value as Kind)}>
            <option value="cumulative">Over a window</option>
            <option value="concurrent">At once</option>
          </select>
        </label>
      )}
      <label className="flex flex-col gap-1">
        <span>{`Amount (${unit(resource, kind)})`}</span>
        <input
          className={`${inputClass} w-32`}
          type="number"
          min="0"
          step="any"
          aria-label={`Amount (${unit(resource, kind)})`}
          value={amount}
          onChange={event => setAmount(event.target.value)}
        />
      </label>
      {kind === 'cumulative' && (
        <label className="flex flex-col gap-1">
          <span>Window</span>
          <select className={inputClass} aria-label="Budget window" value={budgetWindow} onChange={event => setBudgetWindow(event.target.value)}>
            {windows.map(w => <option key={w.value} value={w.value}>{w.label}</option>)}
          </select>
        </label>
      )}
      <Button type="submit" disabled={set.isPending || amount === ''}>Set budget</Button>
      {set.error && <p role="alert" className="w-full text-destructive">{errorText(set.error)}</p>}
    </form>
  );
}
