import { useEffect, useState } from 'react';
import { CodeBlock } from '@flanksource/clicky-ui/data';
import { apiUrl } from '../config';
import type { Test } from '../types';

export type FixtureFileRef = { file: string; workDir: string; language: string; label: string };

const FILE_LANGUAGES: Record<string, { language: string; label: string }> = {
  yaml: { language: 'yaml', label: 'YAML' },
  yml: { language: 'yaml', label: 'YAML' },
  json: { language: 'json', label: 'JSON' },
  md: { language: 'markdown', label: 'Markdown' },
  mdx: { language: 'markdown', label: 'Markdown' },
};

// firstWorkDir is the work dir the fixture file's steps ran in: the file node
// itself carries none, every step under it carries the run's.
function firstWorkDir(node: Test): string | undefined {
  if (node.work_dir) return node.work_dir;
  for (const child of node.children ?? []) {
    const found = firstWorkDir(child);
    if (found) return found;
  }
  return undefined;
}

/**
 * The fixture file a root fixture node ran, when it is one: the execution
 * tree's `file` node with the file's path, the work dir its steps ran in and
 * a language its extension names. Undefined for every other node.
 */
export function fixtureFileRef(node: Test): FixtureFileRef | undefined {
  const detail = node.detail as { kind?: unknown } | undefined;
  if (node.framework !== 'fixture' || detail?.kind !== 'file' || !node.file) return undefined;
  const kind = FILE_LANGUAGES[node.file.split('.').pop()?.toLowerCase() ?? ''];
  const workDir = firstWorkDir(node);
  if (!kind || !workDir) return undefined;
  return { file: node.file, workDir, ...kind };
}

/** The whole fixture file, as `/api/file` serves it from the run's work dir. */
export function FixtureFileSource({ source, scope = {} }: { source: FixtureFileRef; scope?: { runId?: string } }) {
  const [state, setState] = useState<{ text?: string; error?: string }>({});
  useEffect(() => {
    let live = true;
    const params = new URLSearchParams({ work_dir: source.workDir, path: source.file });
    fetch(apiUrl(`/api/file?${params}`, scope))
      .then(async res => {
        const body = await res.text();
        if (!res.ok) throw new Error(`${res.status} ${body.trim() || res.statusText}`);
        return body;
      })
      .then(text => live && setState({ text }), (err: unknown) => live && setState({ error: err instanceof Error ? err.message : String(err) }));
    return () => {
      live = false;
    };
  }, [source.file, source.workDir, scope.runId]);
  if (state.error) return <div role="alert" className="p-density-4 text-xs text-red-600">{source.file}: {state.error}</div>;
  if (state.text === undefined) return <div className="p-density-4 text-xs text-muted-foreground">Loading {source.file}…</div>;
  return (
    <div className="min-w-0 p-density-4" data-fixture-file-source="">
      <CodeBlock language={source.language} source={state.text} copyable />
    </div>
  );
}
