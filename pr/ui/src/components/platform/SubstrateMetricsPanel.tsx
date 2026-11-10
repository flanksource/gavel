// The platform dashboard's substrate metrics panel: the panels of the Mission Control View that
// graphs substrate's Prometheus metrics, timeseries as line charts and single values as figures.
import { useQuery } from '@tanstack/react-query';
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { StatStrip } from '@flanksource/clicky-ui/data';
import { fetchJSON, queryKeys } from '../../query';

export interface MetricPoint {
  time: string;
  value: number;
}

export interface MetricSeries {
  labels: Record<string, string>;
  points: MetricPoint[];
}

export interface MetricPanel {
  name: string;
  description?: string;
  type: string;
  unit?: string;
  series?: MetricSeries[];
  value?: number;
  unsupported?: string;
}

// MetricsView is a Mission Control View's panels, as /api/platform/substrate/metrics answers them.
export interface MetricsView {
  missionControl: string;
  title?: string;
  refreshError?: string;
  lastRefreshedAt: string;
  panels: MetricPanel[];
}

// seriesSlots is how many series one chart draws: the categorical palette's slots in index.css.
const seriesSlots = 8;

// seriesLabel names a series by its labels' values, or its label pairs when it has several.
export function seriesLabel(labels: Record<string, string>, panel: string): string {
  const names = Object.keys(labels).sort();
  if (names.length === 0) return panel;
  if (names.length === 1) return labels[names[0]];
  return names.map(name => `${name}=${labels[name]}`).join(', ');
}

export type TimeRow = { time: string } & Record<string, number | string>;

// timeRows lines the series up by time, one row per sample time, each series' value under s<index>.
export function timeRows(series: MetricSeries[]): TimeRow[] {
  const rows = new Map<string, TimeRow>();
  series.forEach((line, index) => {
    for (const point of line.points) {
      const row = rows.get(point.time) ?? { time: point.time };
      row[`s${index}`] = point.value;
      rows.set(point.time, row);
    }
  });
  return [...rows.values()].sort((a, b) => Date.parse(a.time) - Date.parse(b.time));
}

function clock(time: string): string {
  return new Date(time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function figure(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(2);
}

// SeriesLines draws a timeseries panel's first seriesSlots series, each in its --series-* slot,
// with every series' numbers as a table: three of the colors are faint on the light surface.
function SeriesLines({ panel }: { panel: MetricPanel }) {
  const series = panel.series ?? [];
  const labels = series.map(line => seriesLabel(line.labels, panel.name));
  const rows = timeRows(series);
  const drawn = series.slice(0, seriesSlots);
  return (
    <figure className="flex flex-col gap-1">
      <figcaption className="text-xs text-muted-foreground">
        <span className="text-foreground">{panel.name}</span>
        {panel.description && <> · {panel.description}</>}
      </figcaption>
      {series.length === 0 ? (
        <p className="text-xs text-muted-foreground">No samples in the view's range.</p>
      ) : (
        <div className="h-48 text-muted-foreground">
          {/* initialDimension sizes the first render, before the container is measured. */}
          <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 480, height: 192 }}>
            <LineChart data={rows} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
              <CartesianGrid vertical={false} stroke="currentColor" strokeOpacity={0.15} />
              <XAxis dataKey="time" tickFormatter={clock} minTickGap={32} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
              <YAxis width={40} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
              <Tooltip labelFormatter={label => clock(String(label))} formatter={(value, key) => [figure(Number(value)), labels[Number(String(key).slice(1))]]} />
              {drawn.length > 1 && <Legend formatter={(key: string) => <span className="text-xs text-foreground">{labels[Number(key.slice(1))]}</span>} iconSize={8} />}
              {drawn.map((_, index) => (
                <Line key={index} dataKey={`s${index}`} stroke={`var(--series-${index + 1})`} strokeWidth={2} dot={false} activeDot={{ r: 4 }} connectNulls={false} isAnimationActive={false} />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
      {series.length > seriesSlots && (
        <p className="text-xs text-muted-foreground">The chart shows {seriesSlots} of {series.length} series; the table lists them all.</p>
      )}
      {series.length > 0 && (
        <details className="text-xs">
          <summary className="cursor-pointer text-muted-foreground">Show as a table</summary>
          <table aria-label={panel.name} className="mt-1 w-full border-collapse text-left">
            <thead className="text-muted-foreground">
              <tr>
                {['Time', ...labels].map((heading, index) => (
                  <th key={index} className="border-b border-border px-2 py-1 font-medium">{heading}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map(row => (
                <tr key={row.time} className="border-b border-border">
                  <td className="px-2 py-1 font-mono">{new Date(row.time).toLocaleString()}</td>
                  {series.map((_, index) => {
                    const value = row[`s${index}`];
                    return <td key={index} className="px-2 py-1">{typeof value === 'number' ? figure(value) : '—'}</td>;
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </details>
      )}
    </figure>
  );
}

export function SubstrateMetricsPanel() {
  const metrics = useQuery({
    queryKey: queryKeys.platform('substrate-metrics'),
    queryFn: ({ signal }) => fetchJSON<MetricsView>({ url: '/api/platform/substrate/metrics', signal, context: 'load substrate\'s metrics' }),
  });
  const view = metrics.data;
  const values = view?.panels.filter(panel => (panel.type === 'number' || panel.type === 'gauge') && !panel.unsupported) ?? [];
  const lines = view?.panels.filter(panel => panel.type === 'timeseries' && !panel.unsupported) ?? [];
  const unsupported = view?.panels.filter(panel => panel.unsupported) ?? [];

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-xs font-semibold uppercase text-muted-foreground">
        Substrate metrics
        {view && (
          <> · <a className="normal-case underline" href={view.missionControl.replace(/\/api\/?$/, '')} target="_blank" rel="noreferrer">{view.title || 'Mission Control'}</a>
            <span className="normal-case font-normal"> · refreshed {new Date(view.lastRefreshedAt).toLocaleString()}</span>
          </>
        )}
      </h2>
      {metrics.error && <p role="alert" className="text-red-600">{metrics.error.message}</p>}
      {view?.refreshError && <p role="status" className="text-amber-600">Mission Control couldn't refresh the view, so what it shows may be old or missing: {view.refreshError}</p>}
      {values.length > 0 && (
        <StatStrip items={values.map(panel => ({
          label: panel.name,
          value: panel.value === undefined ? 'no data' : `${figure(panel.value)}${panel.unit ? `${panel.unit === '%' ? '' : ' '}${panel.unit}` : ''}`,
        }))} />
      )}
      {lines.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          {lines.map(panel => <SeriesLines key={panel.name} panel={panel} />)}
        </div>
      )}
      {unsupported.length > 0 && (
        <ul className="text-xs text-muted-foreground">
          {unsupported.map(panel => <li key={panel.name}>{panel.name}: {panel.unsupported}</li>)}
        </ul>
      )}
    </section>
  );
}
