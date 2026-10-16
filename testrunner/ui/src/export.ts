import { copyText } from '@flanksource/clicky-ui/components';
import { buildExportRoute, type ExportFormat, type RouteState } from './routes';

function exportFilename(state: RouteState, format: ExportFormat): string {
  const path = state.selectedPath ? state.selectedPath.replace(/\//g, '-') : state.tab;
  return `${path || state.tab}.${format}`;
}

export function downloadCurrentView(state: RouteState, format: Exclude<ExportFormat, 'pdf'>) {
  const link = document.createElement('a');
  link.href = buildExportRoute(state, format);
  link.download = exportFilename(state, format);
  document.body.appendChild(link);
  link.click();
  link.remove();
}

function promptHeader(state: RouteState): string {
  if (state.tab === 'lint') return 'Use this gavel lint report as context for your analysis or code changes.';
  if (state.tab === 'bench') return 'Use this gavel benchmark report as context for your analysis or code changes.';
  return 'Use this gavel test report as context for your analysis or code changes.';
}

/** Copies the view's Markdown export, framed as an AI prompt; rejects when the export or the clipboard fails. */
export async function copyCurrentViewForAgent(state: RouteState) {
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
    promptHeader(state),
    '',
    `Source URL: ${exportURL}`,
    '',
    '```md',
    markdown,
    '```',
  ].join('\n');

  await copyText(prompt);
}
