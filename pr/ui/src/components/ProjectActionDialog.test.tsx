import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectActionDialog } from './ProjectActionDialog';

vi.mock('@flanksource/clicky-ui/components', () => ({
  // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for Clicky's Button itself.
  Button: ({ children, disabled }: { children: ReactNode; disabled?: boolean }) => <button disabled={disabled}>{children}</button>,
  JsonSchemaForm: ({ value }: { value: Record<string, unknown> }) => <div>options: {String(value.mode)}</div>,
  Modal: ({ children, open }: { children: ReactNode; open: boolean }) => open ? <div>{children}</div> : null,
}));

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ProjectActionDialog', () => {
  it('reuses a cached action schema after closing and reopening', async () => {
    vi.stubGlobal('localStorage', { getItem: vi.fn().mockReturnValue(null), setItem: vi.fn() });
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        schemaVersion: 1,
        action: 'lint',
        schema: { type: 'object', properties: {} },
        defaults: { mode: 'changed' },
      }),
    });
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const props = {
      projectName: 'acme/widget',
      selectedFiles: [],
      onClose: vi.fn(),
      onRun: vi.fn().mockResolvedValue(undefined),
    };
    const { rerender } = render(
      <QueryClientProvider client={queryClient}>
        <ProjectActionDialog {...props} action="lint" />
      </QueryClientProvider>,
    );
    expect(await screen.findByText('options: changed')).toBeTruthy();

    rerender(
      <QueryClientProvider client={queryClient}>
        <ProjectActionDialog {...props} action={null} />
      </QueryClientProvider>,
    );
    rerender(
      <QueryClientProvider client={queryClient}>
        <ProjectActionDialog {...props} action="lint" />
      </QueryClientProvider>,
    );
    expect(await screen.findByText('options: changed')).toBeTruthy();

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledWith('/api/projects/acme%2Fwidget/actions/schema?action=lint', {
      signal: expect.any(AbortSignal),
    });
  });
});
