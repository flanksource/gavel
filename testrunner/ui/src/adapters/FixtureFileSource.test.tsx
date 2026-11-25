import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { FixtureFileSource, fixtureFileRef } from './FixtureFileSource';
import type { Test } from '../types';

const WORK_DIR = '/repo/fixtures';
const FILE = 'purge.fixture.yaml';
const SOURCE = 'kind: Fixture\nsteps:\n  - id: inspect\n    kind: query\n';

// The execution tree's file node: the file's path, no work dir of its own,
// steps that ran in the work dir.
const fileNode = (file: string): Test => ({
  name: file,
  framework: 'fixture',
  file,
  detail: { name: file, kind: 'file', state: 'passed' },
  children: [{ name: 'inspect', framework: 'fixture', file, work_dir: WORK_DIR, passed: true }],
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('fixtureFileRef', () => {
  it.each([
    [FILE, 'yaml', 'YAML'],
    ['checks.fixture.md', 'markdown', 'Markdown'],
  ])('names %s a %s file shown as %s, read from its steps\' work dir', (file, language, label) => {
    expect(fixtureFileRef(fileNode(file))).toEqual({ file, workDir: WORK_DIR, language, label });
  });

  it('is undefined for a step, and for a file with no step that ran yet', () => {
    expect(fixtureFileRef(fileNode(FILE).children![0]!)).toBeUndefined();
    expect(fixtureFileRef({ ...fileNode(FILE), children: [] })).toBeUndefined();
  });
});

describe('FixtureFileSource', () => {
  it('shows the whole file /api/file serves from the work dir', async () => {
    const fetchMock = vi.fn(async () => new Response(SOURCE, { status: 200 }));
    vi.stubGlobal('fetch', fetchMock);
    render(<FixtureFileSource source={fixtureFileRef(fileNode(FILE))!} />);
    expect((await screen.findByText(/kind: query/)).textContent).toContain('kind: query');
    expect(fetchMock).toHaveBeenCalledWith(`/api/file?${new URLSearchParams({ work_dir: WORK_DIR, path: FILE })}`);
  });

  it('says why the file could not be read', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('file not found', { status: 404 })));
    render(<FixtureFileSource source={fixtureFileRef(fileNode(FILE))!} />);
    expect((await screen.findByRole('alert')).textContent).toBe(`${FILE}: 404 file not found`);
  });
});
