// The organization's connections as clicky-ui's connections page shows them, over the server's
// connection catalog: admins add, edit and delete them, members see them, its legend stacked.
import { useState } from 'react';
import { ConnectionsPage } from '@flanksource/clicky-ui/connections';
import { MonacoProvider } from '@flanksource/clicky-ui/monaco';
import { monacoWorker } from './monacoWorkers';

/** Where the server serves the connection catalog the page reads. */
export const CONNECTIONS_API_BASE = '/api/catalog';

/**
 * clicky-ui's health legend, the one element of the page with these classes; the panel stacks it
 * one state per line. The panel's class names repeat it literally, for Tailwind to find.
 */
export const LEGEND_SELECTOR = 'p.flex-wrap.gap-x-4';

export function ConnectionsPanel() {
  const [selected, setSelected] = useState<string>();
  return (
    <MonacoProvider getWorker={monacoWorker}>
      <div data-connections-panel className="[&_p.flex-wrap.gap-x-4]:flex-col [&_p.flex-wrap.gap-x-4]:items-start">
        <ConnectionsPage
          apiBase={CONNECTIONS_API_BASE}
          {...(selected ? { selected } : {})}
          onSelect={setSelected}
          connectionHref={id => `#connection=${encodeURIComponent(id)}`}
        />
      </div>
    </MonacoProvider>
  );
}
