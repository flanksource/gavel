// Stacked bars of a platform dashboard window's days, one segment per series in its fixed color,
// with the same numbers as a table: some of the colors are faint on the light surface.
import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { formatDay, formatDayShort } from './dates';

// DaySeries is one stacked segment: the key its value is under in each day's row, its label and color.
export interface DaySeries {
  key: string;
  label: string;
  color: string;
}

export type DayRow = { day: string } & Record<string, number | string>;

export function StackedDayBars({ title, data, series, unit }: { title: string; data: DayRow[]; series: DaySeries[]; unit: string }) {
  const labels = Object.fromEntries(series.map(s => [s.key, s.label]));
  const value = (row: DayRow, key: string) => Number(row[key] ?? 0);
  return (
    <figure className="flex flex-col gap-1">
      <figcaption className="text-xs text-muted-foreground">{title}</figcaption>
      <div className="h-48 text-muted-foreground">
        {/* initialDimension sizes the first render, before the container is measured. */}
        <ResponsiveContainer width="100%" height="100%" initialDimension={{ width: 480, height: 192 }}>
          <BarChart data={data} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
            <CartesianGrid vertical={false} stroke="currentColor" strokeOpacity={0.15} />
            <XAxis dataKey="day" tickFormatter={formatDayShort} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <YAxis allowDecimals={false} width={36} tick={{ fill: 'currentColor', fontSize: 11 }} tickLine={false} axisLine={false} />
            <Tooltip formatter={(v, key) => [`${Math.round(Number(v))} ${unit}`, labels[String(key)] ?? String(key)]} cursor={{ fillOpacity: 0.08 }} />
            <Legend formatter={(key: string) => <span className="text-xs text-foreground">{labels[key] ?? key}</span>} iconSize={8} />
            {series.map(s => (
              <Bar key={s.key} dataKey={s.key} stackId="series" fill={s.color} stroke="var(--background)" strokeWidth={1} maxBarSize={24} />
            ))}
          </BarChart>
        </ResponsiveContainer>
      </div>
      <details className="text-xs">
        <summary className="cursor-pointer text-muted-foreground">Show as a table</summary>
        <table aria-label={title} className="mt-1 w-full border-collapse text-left">
          <thead className="text-muted-foreground">
            <tr>
              {['Day', ...series.map(s => s.label), 'Total'].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map(row => (
              <tr key={row.day} className="border-b border-border">
                <td className="px-2 py-1 font-mono">{formatDay(row.day)}</td>
                {series.map(s => <td key={s.key} className="px-2 py-1">{Math.round(value(row, s.key))}</td>)}
                <td className="px-2 py-1">{Math.round(series.reduce((sum, s) => sum + value(row, s.key), 0))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </details>
    </figure>
  );
}
