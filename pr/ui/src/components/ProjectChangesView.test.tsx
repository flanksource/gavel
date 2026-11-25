import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ProjectChangesView } from './ProjectChangesView';
import { projectStatusResponse } from './ProjectStatusView.fixture';
import type { ProjectFileStatus } from './ProjectStatusView';

const { mobile } = vi.hoisted(() => ({ mobile: vi.fn(() => true) }));
vi.mock('../useIsMobile', () => ({ useIsMobile: mobile }));
const files = projectStatusResponse.files as ProjectFileStatus[];

function Review({ initialPath = '', locked = new Map<string, number>(), onToggleStaged = vi.fn() }) {
  const [path, setPath] = useState(initialPath);
  return <ProjectChangesView
    files={files}
    locked={locked}
    disabled={false}
    diffPath={path}
    onDiffPathChange={setPath}
    onToggleStaged={onToggleStaged}
    fileList={<button onClick={() => setPath('one.go')}>Review one.go</button>}
    diff={<div>Diff for {path}</div>}
  />;
}

beforeEach(() => mobile.mockReturnValue(true));

describe('mobile project change review', () => {
  it('opens a diff and stages it from the review header, keeping the file list reachable', () => {
    const onToggleStaged = vi.fn();
    render(<Review onToggleStaged={onToggleStaged} />);
    expect(screen.queryByText('Diff for one.go')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Review one.go' }));
    expect(screen.queryByRole('button', { name: 'Review one.go' })).toBeNull();
    expect(screen.getByText('Diff for one.go')).toBeTruthy();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Stage one.go' }));
    expect(onToggleStaged).toHaveBeenCalledWith(files.find(file => file.path === 'one.go'));
    fireEvent.click(screen.getByRole('button', { name: 'Back to files' }));
    expect(screen.getByRole('button', { name: 'Review one.go' })).toBeTruthy();
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
    { path: 'conflicted.go', locked: new Map<string, number>(), reason: 'Resolve conflicts before staging' },
    { path: 'one.go', locked: new Map([['one.go', 1]]), reason: 'Queued in commit group 1' },
  ])('keeps $path reviewable while preventing staging: $reason', ({ path, locked, reason }) => {
    render(<Review initialPath={path} locked={locked} />);
    expect(screen.getByText(`Diff for ${path}`)).toBeTruthy();
    expect(screen.getByRole('checkbox', { name: `Stage ${path}` }).hasAttribute('disabled')).toBe(true);
    expect(screen.getByText(reason)).toBeTruthy();
  });

  it('shows the staged and partly staged states of the open file', () => {
    const staged = { ...files[0]!, state: 'staged' as const };
    const both = { ...files[0]!, state: 'both' as const };
    for (const [file, checked, indeterminate] of [[staged, true, false], [both, false, true]] as const) {
      const view = render(<ProjectChangesView files={[file]} locked={new Map()} disabled={false} diffPath={file.path} onDiffPathChange={vi.fn()} onToggleStaged={vi.fn()} fileList={null} diff={null} />);
      const box = screen.getByRole('checkbox', { name: `Stage ${file.path}` }) as HTMLInputElement;
      expect([box.checked, box.indeterminate]).toEqual([checked, indeterminate]);
      view.unmount();
    }
  });

  it('keeps the desktop split view without mobile navigation controls', () => {
    mobile.mockReturnValue(false);
    render(<Review initialPath="one.go" />);
    expect(screen.getByRole('button', { name: 'Review one.go' })).toBeTruthy();
    expect(screen.getByText('Diff for one.go')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Back to files' })).toBeNull();
  });
});
