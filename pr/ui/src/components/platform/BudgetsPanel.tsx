// The platform dashboard's budgets panel: Captain's budget rules, each group's completed spend in
// the rule's current window against its amount, and why a rule or group has no total.
import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from '../../query';
import { PanelHeading } from './PanelHeading';
import { formatDay } from './dates';

export interface BudgetGroup {
  values: Record<string, string>;
  spendUSD: number;
  calls: number;
  error?: string;
}

// Budget is a live budget rule; windowError says why it has no windowStart and no groups.
export interface Budget {
  id: string;
  name: string;
  match: { dimensions?: Record<string, string>; models?: string[] };
  groupBy: string[];
  amountUSD: number;
  window: string;
  windowStart?: string;
  windowError?: string;
  groups: BudgetGroup[];
}

const usd = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

function groupName(values: Record<string, string>): string {
  return Object.keys(values).sort().map(key => `${key}=${values[key]}`).join(', ');
}

// used is the share of the amount spent, saying so in words when the group is over budget.
function used(spend: number, amount: number): string {
  const percent = Math.round((spend / amount) * 100);
  return spend > amount ? `${percent}% · over budget` : `${percent}%`;
}

interface Row {
  key: string;
  cells: string[];
}

// budgetRows is one row per group of each rule, or one saying why the rule has none.
function budgetRows(budgets: Budget[]): Row[] {
  return budgets.flatMap(budget => {
    const window = budget.windowStart ? `${budget.window} since ${formatDay(budget.windowStart.slice(0, 10))}` : budget.window;
    const amount = usd.format(budget.amountUSD);
    if (budget.windowError) return [{ key: budget.id, cells: [budget.name, window, budget.windowError, '—', amount, '—'] }];
    if (budget.groups.length === 0) return [{ key: budget.id, cells: [budget.name, window, 'No spend in this window', usd.format(0), amount, used(0, budget.amountUSD)] }];
    return budget.groups.map(group => {
      const name = groupName(group.values);
      const key = `${budget.id}/${name}`;
      if (group.error) return { key, cells: [budget.name, window, name, group.error, amount, '—'] };
      return { key, cells: [budget.name, window, name, usd.format(group.spendUSD), amount, used(group.spendUSD, budget.amountUSD)] };
    });
  });
}

export function BudgetsPanel() {
  const budgets = useQuery({
    queryKey: queryKeys.platform('budgets'),
    queryFn: ({ signal }) => fetchJSON<{ budgets: Budget[] }>({ url: '/api/platform/budgets', signal, context: 'load budgets' }),
  });
  const data = budgets.data?.budgets;

  return (
    <section className="flex flex-col gap-3">
      <PanelHeading>Budgets</PanelHeading>
      {budgets.error && <p role="alert" className="text-red-600">{budgets.error.message}</p>}
      {data && data.length === 0 && (
        <p className="text-xs text-muted-foreground">No budget rules. Add one with captain budget-rule create, or on captain serve's Budgets page.</p>
      )}
      {data && data.length > 0 && (
        <table aria-label="Budgets" className="w-full border-collapse text-left text-xs">
          <thead className="text-muted-foreground">
            <tr>
              {['Rule', 'Window', 'Group', 'Spent', 'Budget', 'Used'].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {budgetRows(data).map(row => (
              <tr key={row.key} className="border-b border-border">
                {row.cells.map((cell, i) => <td key={i} className={`px-2 py-1.5${i === 0 ? ' font-medium' : ''}`}>{cell}</td>)}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
