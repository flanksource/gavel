import type { FixtureSQLProfile } from '../types';

export function SQLProfileDetails({ profile }: { profile: FixtureSQLProfile }) {
  if (!profile.statements?.length) return null;
  return <details className="mt-2 text-left font-sans">
    <summary className="cursor-pointer text-xs text-blue-700">{profile.statements.length} {profile.statements.length === 1 ? 'query' : 'queries'}</summary>
    <div className="mt-2 max-h-80 min-w-80 space-y-2 overflow-auto">
      {profile.statements.map((statement, index) => <div key={index} className="rounded border border-gray-200 bg-gray-50 p-2 text-xs">
        <div className="mb-1 text-gray-500">{statement.duration_ms.toFixed(1)} ms · {statement.rows} {statement.rows === 1 ? 'row' : 'rows'}{statement.slow ? ' · slow' : ''}{statement.error ? ' · error' : ''}</div>
        <pre className="whitespace-pre-wrap break-all font-mono">{statement.sql}</pre>
        {statement.params.length > 0 && <div className="mt-1"><span className="text-gray-500">Params</span><ol className="list-decimal pl-5 font-mono">{statement.params.map((param, position) => <li key={position} className="break-all">{param}</li>)}</ol></div>}
      </div>)}
    </div>
  </details>;
}
