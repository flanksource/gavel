// The days of a platform dashboard window, so a chart's axis shows the days without data too.

// windowDays is each UTC day, as YYYY-MM-DD, from days days before now through now's day, oldest first.
export function windowDays(days: number, now: Date): string[] {
  const today = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
  const all: string[] = [];
  for (let back = days; back >= 0; back--) all.push(new Date(today - back * 86_400_000).toISOString().slice(0, 10));
  return all;
}
