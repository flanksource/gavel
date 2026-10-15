import { DataTable, type DataTableColumnInput } from '@flanksource/clicky-ui/data';
import { useProcessFacet, type ProcessFacet } from '../hooks/use-process-facet';
import type { ProcessDescriptor } from '../types';

type EnvRow = { name: string; value: string };

const descriptorColumns: Record<Exclude<ProcessFacet, 'env'>, Array<DataTableColumnInput<ProcessDescriptor>>> = {
  files: [
    { key: 'fd', label: 'FD', sortable: true },
    { key: 'type', label: 'Type', sortable: true, filterable: true },
    { key: 'name', label: 'Path', sortable: true, grow: true },
  ],
  sockets: [
    { key: 'fd', label: 'FD', sortable: true },
    { key: 'protocol', label: 'Protocol', sortable: true, filterable: true, accessor: d => d.protocol || d.type },
    { key: 'state', label: 'State', sortable: true, filterable: true },
    { key: 'localAddr', label: 'Local', sortable: true, grow: true, accessor: d => d.localAddr || d.name },
    { key: 'remoteAddr', label: 'Remote', sortable: true },
  ],
  ports: [
    { key: 'port', label: 'Port', sortable: true },
    { key: 'protocol', label: 'Protocol', sortable: true, filterable: true },
    { key: 'localAddr', label: 'Address', sortable: true, grow: true },
    { key: 'fd', label: 'FD', sortable: true },
  ],
};

const envColumns: Array<DataTableColumnInput<EnvRow>> = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'value', label: 'Value', grow: true },
];

const emptyMessages: Record<ProcessFacet, string> = {
  env: 'The process has no environment variables',
  files: 'No open files',
  sockets: 'No open sockets',
  ports: 'Not listening on any port',
};

// ProcessFacetView reads one facet of a live process when it is shown.
export function ProcessFacetView({ facet, pid, exited }: { facet: ProcessFacet; pid: number; exited: boolean }) {
  const state = useProcessFacet(exited ? null : facet, pid);
  if (exited) {
    return <div className="text-sm text-gray-500">The process has exited, so its environment, files and sockets can no longer be read.</div>;
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-3 text-xs text-gray-500">
        <span>{state.loadedAt ? `Read ${state.loadedAt.toLocaleTimeString()}` : state.loading ? 'Reading...' : ''}</span>
        <button
          className="shrink-0 text-[11px] px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
          onClick={state.reload}
          disabled={state.loading}
        >
          <iconify-icon icon={state.loading ? 'svg-spinners:ring-resize' : 'codicon:refresh'} />
          Refresh
        </button>
      </div>
      {state.error && <div className="text-xs text-red-700 whitespace-pre-wrap">{state.error}</div>}
      {state.data !== undefined && (facet === 'env'
        ? <DataTable<EnvRow>
            data={Object.entries(state.data as Record<string, string>)
              .map(([name, value]) => ({ name, value }))
              .sort((a, b) => a.name.localeCompare(b.name))}
            columns={envColumns}
            getRowId={row => row.name}
            emptyMessage={emptyMessages.env}
            autoFilter
            showThemeControl={false}
            showFullscreenControl={false}
          />
        : <DataTable<ProcessDescriptor>
            data={state.data as ProcessDescriptor[]}
            columns={descriptorColumns[facet]}
            getRowId={(row, index) => `${row.fd}:${index}`}
            emptyMessage={emptyMessages[facet]}
            autoFilter
            showThemeControl={false}
            showFullscreenControl={false}
          />)}
    </div>
  );
}
