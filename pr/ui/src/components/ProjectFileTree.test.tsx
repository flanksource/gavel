import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ProjectFileTree } from './ProjectFileTree';
import { projectStatusResponse } from './ProjectStatusView.fixture';

describe('ProjectFileTree line counts', () => {
  it('keeps assets selectable without displaying empty line counts', () => {
    const source = projectStatusResponse.files[0]!;
    render(
      <ProjectFileTree
        files={[
          source,
          { ...source, path: 'image.svg', adds: 0, dels: 0 },
          { ...source, path: 'report.pdf', adds: 0, dels: 0 },
        ]}
        search=""
        onSearchChange={vi.fn()}
        selected={new Set()}
        disabled={false}
        showResults={false}
        onToggleFile={vi.fn()}
        onToggleFiles={vi.fn()}
        onIgnore={vi.fn()}
      />,
    );
    const sourceRow = within(screen.getByRole('treeitem', { name: source.path }));
    expect(sourceRow.getByText(`+${source.adds}`)).toBeTruthy();
    expect(sourceRow.getByText(`−${source.dels}`)).toBeTruthy();
    for (const path of ['image.svg', 'report.pdf']) {
      const row = within(screen.getByRole('treeitem', { name: path }));
      expect(row.getByRole('checkbox', { name: `Select ${path}` })).toBeTruthy();
      expect(row.queryByText('+0')).toBeNull();
      expect(row.queryByText('−0')).toBeNull();
    }
  });
});
