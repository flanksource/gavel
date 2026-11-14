import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { RunHistory } from './RunHistory';
import { RunSwitcher } from './RunSwitcher';
import type { RunHistory as HistoryState } from '../hooks/use-run-history';

describe('fixture run history', () => {
  it('names fixtures in the table and switcher and opens the recorded run', () => {
    const navigate = vi.fn();
    const history: HistoryState = {
      available: true, latest: [], latestHref: '/', hasMore: false, loading: false, loadMore: vi.fn(),
      runs: [{
        runId: 'run-example', path: '/repo/.gavel/run-example.json', href: '/runs/run-example/',
        kind: 'fixture', frameworks: ['fixture'], fixtureFiles: ['hack/purge.fixture.yaml'],
        passed: 3, failed: 1, total: 4, skipped: 0, warned: 0, lintViolations: 0, lintLinters: 0,
      }],
    };
    render(<><RunHistory history={history} navigate={navigate} /><RunSwitcher history={history} navigate={navigate} /></>);
    expect(screen.getByRole('columnheader', { name: /Fixture/ })).toBeTruthy();
    const file = screen.getByText('hack/purge.fixture.yaml');
    expect(document.querySelector('[data-run-switcher]')?.textContent).toContain('hack/purge.fixture.yaml');
    expect(screen.queryByText('0 lint')).toBeNull();
    fireEvent.click(file.closest('tr')!);
    expect(navigate).toHaveBeenCalledWith('/runs/run-example/');
  });
});
