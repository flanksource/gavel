import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { DetailPanel } from './DetailPanel';

describe('DetailPanel Ginkgo metadata', () => {
  it('shows cached status, inherited labels, retry count, and report entries', () => {
    render(<DetailPanel test={{
      name: 'reports metadata',
      framework: 'ginkgo',
      passed: true,
      cached: true,
      context: {
        labels: ['inherited', 'leaf'],
        num_attempts: 2,
        report_entries: [{ name: 'resource', value: { kind: 'Deployment' } }],
      },
    }} />);

    expect(screen.getByText('cached')).toBeTruthy();
    expect(screen.getByText('inherited')).toBeTruthy();
    expect(screen.getByText('leaf')).toBeTruthy();
    expect(screen.getByText('Ginkgo attempts')).toBeTruthy();
    expect(screen.getByText('resource')).toBeTruthy();
    expect(screen.getByText(/Deployment/)).toBeTruthy();
  });
});
