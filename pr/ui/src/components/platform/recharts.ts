// recharts is loaded on demand, in a chunk of its own, the first time a platform dashboard chart
// renders; charts read it with use() under a Suspense boundary, so their tables don't wait for it.
type Recharts = typeof import('recharts');

let recharts: Promise<Recharts> | undefined;

export function loadRecharts(): Promise<Recharts> {
  return (recharts ??= import('recharts'));
}
