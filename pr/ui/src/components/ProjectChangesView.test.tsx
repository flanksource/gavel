import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ProjectChangesView } from './ProjectChangesView';
import { projectStatusResponse } from './ProjectStatusView.fixture';
import type { ProjectFileStatus } from './ProjectStatusView';

const { mobile } = vi.hoisted(() => ({ mobile: vi.fn(() => true) }));
vi.mock('../useIsMobile', () => ({ useIsMobile: mobile }));
const files = projectStatusResponse.files as ProjectFileStatus[];

function Review({ initialPath = '', locked = new Map<string, number>() }) {
  const [path, setPath] = useState(initialPath);
  const [selected, setSelected] = useState(new Set<string>());
  return <>
    <ProjectChangesView
      files={files}
      selected={selected}
      locked={locked}
      disabled={false}
      diffPath={path}
      onDiffPathChange={setPath}
      onToggleFile={file => setSelected(current => current.has(file) ? new Set([...current].filter(item => item !== file)) : new Set(current).add(file))}
      fileList={<button onClick={() => setPath('one.go')}>Review one.go</button>}
      diff={<div>Diff for {path}</div>}
    />
    <output aria-label="Selected files">{[...selected].join(',')}</output>
  </>;
}

beforeEach(() => mobile.mockReturnValue(true));

describe('mobile project change review', () => {
  it('opens a diff, selects it for commit, and retains selection on returning to files', () => {
    render(<Review />);
    expect(screen.queryByText('Diff for one.go')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Review one.go' }));
    expect(screen.queryByRole('button', { name: 'Review one.go' })).toBeNull();
    expect(screen.getByText('Diff for one.go')).toBeTruthy();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Include one.go in commit' }));
    fireEvent.click(screen.getByRole('button', { name: 'Back to files' }));
    expect(screen.getByRole('button', { name: 'Review one.go' })).toBeTruthy();
    expect(screen.getByLabelText('Selected files').textContent).toBe('one.go');
  });

  it('steps through changed files and stops at both ends', () => {
    render(<Review initialPath="conflicted.go" />);
    expect(screen.getByRole('button', { name: 'Previous file' }).hasAttribute('disabled')).toBe(true);
    expect(screen.getByText('1 of 4')).toBeTruthy();
    for (let index = 0; index < 3; index++) fireEvent.click(screen.getByRole('button', { name: 'Next file' }));
    expect(screen.getByText('Diff for one.go')).toBeTruthy();
    expect(screen.getByText('4 of 4')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Next file' }).hasAttribute('disabled')).toBe(true);
  });

  it.each([
    { path: 'conflicted.go', locked: new Map<string, number>() },
    { path: 'one.go', locked: new Map([['one.go', 1]]) },
  ])('keeps $path reviewable while preventing invalid selection', ({ path, locked }) => {
    render(<Review initialPath={path} locked={locked} />);
    expect(screen.getByText(`Diff for ${path}`)).toBeTruthy();
    expect(screen.getByRole('checkbox', { name: `Include ${path} in commit` }).hasAttribute('disabled')).toBe(true);
  });

  it('keeps the desktop split view without mobile navigation controls', () => {
    mobile.mockReturnValue(false);
    render(<Review initialPath="one.go" />);
    expect(screen.getByRole('button', { name: 'Review one.go' })).toBeTruthy();
    expect(screen.getByText('Diff for one.go')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Back to files' })).toBeNull();
  });
});
