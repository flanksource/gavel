// The organization's connections as clicky-ui's connections page shows them, over the server's
// connection catalog: admins add, edit and delete them, members see them.
import { useState } from 'react';
import { ConnectionsPage } from '@flanksource/clicky-ui/connections';

/** Where the server serves the connection catalog the page reads. */
export const CONNECTIONS_API_BASE = '/api/catalog';

export function ConnectionsPanel() {
  const [selected, setSelected] = useState<string>();
  return (
    <ConnectionsPage
      apiBase={CONNECTIONS_API_BASE}
      {...(selected ? { selected } : {})}
      onSelect={setSelected}
      connectionHref={id => `#connection=${encodeURIComponent(id)}`}
    />
  );
}
