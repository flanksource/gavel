# @flanksource/gavel

React components and hooks for consuming [gavel](https://github.com/flanksource/gavel)'s
testrunner UI from your own React apps, fed by gavel's HTTP handlers for real-time test
execution progress.

This is the same UI gavel embeds in its binary (`gavel test --ui`), published as a library so you
can render the test tree, summary, and detail panel — and subscribe to live progress over
Server-Sent Events — inside any React application. It follows the same model as
[`@flanksource/clicky-ui`](https://www.npmjs.com/package/@flanksource/clicky-ui)'s `useTaskRun`
hook and `TaskProgress` component.

## Install

```sh
npm install @flanksource/gavel
```

`react`, `react-dom`, and `@flanksource/clicky-ui` are peer dependencies — install them in your
app if they aren't already present:

```sh
npm install react react-dom @flanksource/clicky-ui
```

## `useTestRun` — live test progress

The centerpiece is the `useTestRun` hook. Point it at a running gavel server and it subscribes to
`/api/tests/stream` (SSE), falling back to polling `/api/tests` where `EventSource` is
unavailable. It returns the latest full snapshot plus derived fields and the rerun/stop actions.

```tsx
import { useTestRun } from '@flanksource/gavel/testrunner/hooks';

function TestProgress() {
  const { tests, status, statusText, done, error, rerun, stop } = useTestRun({
    baseUrl: 'http://localhost:8080', // your gavel server origin (or '' if same-origin)
  });

  if (error) return <p>{error}</p>;
  return (
    <div>
      <p>{statusText}</p>
      <ul>{tests.map(t => <li key={t.name}>{t.name}</li>)}</ul>
      {status.running && <button onClick={() => stop()}>Stop</button>}
      {done && <button onClick={() => rerun()}>Rerun all</button>}
    </div>
  );
}
```

### Options

| Option | Default | Description |
| --- | --- | --- |
| `baseUrl` | `window.__gavelBasePath ?? ''` | API prefix, e.g. `''`, `'https://host'`, or `'/results/{repo}/{id}'`. |
| `basePath` | — | Alias for `baseUrl` (parity with `useTaskRun`). |
| `enabled` | `true` | Disable the subscription (e.g. before a target is known). |
| `pollMs` | `2000` | Polling-fallback interval in ms. |
| `forcePoll` | `false` | Force the polling transport even when `EventSource` exists. |

### Result

`{ snapshot, tests, lint, bench, status, statusText, runMeta, done, error, rerun, stop, refetch }`.
`rerun(body?)` POSTs to `/api/rerun`; `stop(taskId?)` POSTs to `/api/stop` (omit `taskId` for a
global stop).

## Components

The full UI is exported from `@flanksource/gavel/testrunner`:

```tsx
import { App } from '@flanksource/gavel/testrunner';
import '@flanksource/clicky-ui/styles.css';

// Mount the complete testrunner UI; it reads window.__gavelBasePath for its API origin.
export default function Page() {
  return <App />;
}
```

Lower-level pieces are also exported for composing your own layout: `Summary`, `TestNode`,
`DetailPanel`, `FilterBar`, `ProgressBar`, `LintView`, `BenchView`, `DiagnosticsView`, `SplitPane`,
`JsonView`, `AnsiHtml`.

## Types

All result and progress types (`Test`, `Snapshot`, `SnapshotStatus`, `LinterResult`, `Violation`,
`RunMeta`, `BenchComparison`, `ProcessNode`, `DiagnosticsSnapshot`, `RerunRequest`, …) are exported
from `@flanksource/gavel/testrunner/types`. They mirror the JSON gavel's handlers emit.

## Styling

The components are styled with Tailwind utility classes and use Iconify web components
(`<iconify-icon>`) for icons. Include Tailwind (or the clicky-ui stylesheet) and register the
Iconify element in your host page, exactly as gavel's embedded page does.
