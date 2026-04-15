import { buildExportRoute, type ExportFormat, type RouteState } from './routes';

function exportFilename(state: RouteState, format: ExportFormat): string {
  const base = state.selectedPath ? state.selectedPath.replace(/\//g, '-') : 'prs';
  return `${base}.${format}`;
}

export function downloadCurrentView(state: RouteState, format: ExportFormat) {
  const link = document.createElement('a');
  link.href = buildExportRoute(state, format);
  link.download = exportFilename(state, format);
  document.body.appendChild(link);
  link.click();
  link.remove();
}

export async function copyCurrentViewForAgent(state: RouteState) {
  if (!navigator.clipboard?.writeText) {
    throw new Error('Clipboard access is unavailable in this browser');
  }

  const exportPath = buildExportRoute(state, 'md');
  const exportURL = new URL(exportPath, window.location.origin).toString();
  const response = await fetch(exportPath, {
    headers: { Accept: 'text/markdown' },
  });
  if (!response.ok) {
    throw new Error(`Markdown export failed (${response.status})`);
  }

  const markdown = (await response.text()).trimEnd();
  const prompt = [
    'Use this gavel PR dashboard report as context for your analysis or code changes.',
    '',
    `Source URL: ${exportURL}`,
    '',
    '```md',
    markdown,
    '```',
  ].join('\n');

  await navigator.clipboard.writeText(prompt);
}
