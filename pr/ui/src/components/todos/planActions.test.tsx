import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { RunContext } from './providers';
import { PlanApproveButtons } from './planActions';

vi.mock('@flanksource/clicky-ui/components', () => {
  return {
    Button: ({
      children,
      variant: _variant,
      size: _size,
      loading: _loading,
      loadingLabel: _loadingLabel,
      asChild: _asChild,
      ...props
    }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string; loading?: boolean; loadingLabel?: React.ReactNode; asChild?: boolean }) => (
      // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
      <button type="button" {...props}>
        {children}
      </button>
    ),
    Combobox: () => null,
    DropdownMenu: ({
      trigger,
      children,
    }: {
      trigger: React.ReactNode;
      children: (close: () => void) => React.ReactNode;
    }) => (
      <div>
        {trigger}
        {children(() => {})}
      </div>
    ),
    Field: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
    ListMenuItem: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
    Modal: ({ children, open }: { children?: React.ReactNode; open?: boolean }) => (open === false ? null : <div>{children}</div>),
    SegmentedControl: <T extends string>({
      options,
      onChange,
    }: {
      options: Array<{ id: T; label: string; disabled?: boolean }>;
      onChange: (value: T) => void;
    }) => (
      <div>
        {options.map(option => (
          // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for SegmentedControl itself.
          <button key={option.id} type="button" disabled={option.disabled} onClick={() => onChange(option.id)}>
            {option.label}
          </button>
        ))}
      </div>
    ),
  };
});

vi.mock('@flanksource/clicky-ui/chat', () => ({
  ModelSelector: () => null,
  providerIcon: () => (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />,
  ProviderSelector: () => null,
}));

vi.mock('@flanksource/clicky-ui/ai', () => ({
  effortOptionsForModel: (_model: unknown, fallback: string[]) => fallback,
  PromptRunEditor: () => null,
  promptRuntimeValueToPayload: (value: unknown) => value,
  reconcileModelCapabilities: (value: unknown) => value,
}));

vi.mock('@flanksource/clicky-ui/icons', () => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    UiCancel: Icon,
    UiAdd: Icon,
    UiBatteryChargingVertical: Icon,
    UiBatteryVerticalEmpty: Icon,
    UiBatteryVerticalFull: Icon,
    UiBatteryVerticalHigh: Icon,
    UiBatteryVerticalLow: Icon,
    UiBatteryVerticalMedium: Icon,
    UiCheck: Icon,
    UiCheckFilled: Icon,
    UiChevronDown: Icon,
    UiChevronRight: Icon,
    UiClock: Icon,
    UiCloud: Icon,
    UiCog: Icon,
    UiColumns: Icon,
    UiComment: Icon,
    UiEdit: Icon,
    UiError: Icon,
    UiEye: Icon,
    UiFolder: Icon,
    UiGitGraph: Icon,
    UiHistory: Icon,
    UiLightbulb: Icon,
    UiListFlat: Icon,
    UiListDashes: Icon,
    UiPass: Icon,
    UiPlay: Icon,
    UiQuestion: Icon,
    UiRobotAi: Icon,
    UiRows: Icon,
    UiSparkles: Icon,
    UiTerminal: Icon,
    UiWarningTriangle: Icon,
  };
});

const context: RunContext = {
  defaultBackend: 'codex-agent',
  efforts: ['low', 'medium', 'high', 'xhigh'],
  tools: [],
  backends: [
    {
      id: 'codex-cmux',
      label: 'Codex cmux',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'codex-cmux',
      mechanisms: [{ value: 'cmux', label: 'cmux (TUI)', driver: 'codex-cmux' }],
      models: [
        { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'codex-agent',
      label: 'Codex Agent',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5.5',
      driver: 'codex-headless',
      mechanisms: [{ value: 'agent', label: 'Agent', driver: 'codex-headless' }],
      models: [
        { id: 'gpt-5.5', provider: 'openai', label: 'GPT-5.5', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'claude-agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-opus-4-8',
      driver: 'claude-headless',
      mechanisms: [{ value: 'agent', label: 'agent', driver: 'claude-headless' }],
      models: [
        { id: 'claude-opus-4-8', provider: 'anthropic', label: 'Claude Opus 4.8', reasoning: true, configured: true },
      ],
      configured: true,
    },
  ],
};

function mockRunContext() {
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => context }) as Response));
}

beforeEach(() => {
  const store: Record<string, string> = {};
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = String(value);
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      for (const key of Object.keys(store)) delete store[key];
    }),
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('PlanApproveButtons', () => {
  it('approves and runs with the remembered run options', async () => {
    mockRunContext();
    localStorage.setItem(
      'gavel.pr-ui.todoRunChoices.v1',
      JSON.stringify({
        last: {
          run: {
            driver: 'claude-headless',
            backend: 'claude-agent',
            model: 'claude-opus-4-8',
            effort: 'high',
            runMode: 'run',
          },
        },
        recentAdvanced: {},
      }),
    );
    const onApprove = vi.fn();

    render(<PlanApproveButtons onApprove={onApprove} />);

    const approveRun = await screen.findByRole('button', { name: /Approve & Run \(Agent:opus-4\.8\)/ });
    expect(approveRun.className).toContain('h-8');
    expect(approveRun.className).toContain('px-2');
    expect(approveRun.className).toContain('text-xs');
    expect(approveRun.className).toContain('font-medium');
    expect(approveRun.className).toContain('text-foreground');
    expect(approveRun.className).toContain('hover:bg-muted');
    expect(approveRun.className).not.toContain('bg-amber');

    fireEvent.click(approveRun);

    expect(onApprove).toHaveBeenCalledWith(true, expect.objectContaining({
      driver: 'claude-headless',
      backend: 'claude-agent',
      model: 'claude-opus-4-8',
      effort: 'high',
      runMode: 'run',
    }));
  });

  it('selects model and effort from the run dropdown before approving', async () => {
    mockRunContext();
    const onApprove = vi.fn();

    render(<PlanApproveButtons onApprove={onApprove} />);

    await screen.findByRole('button', { name: /Approve & Run \(Agent:gpt-5\.5\)/ });
    fireEvent.click(screen.getByRole('button', { name: 'Claude' }));
    fireEvent.change(screen.getByRole('slider', { name: 'Effort' }), { target: { value: '2' } });
    const opusButton = screen.getByRole('button', { name: 'Claude Opus 4.8' });
    expect(opusButton.textContent).toBe('Claude Opus 4.8');
    expect(opusButton.className).toContain('px-2');
    fireEvent.click(opusButton);

    await waitFor(() =>
      expect(onApprove).toHaveBeenCalledWith(true, expect.objectContaining({
        driver: 'claude-headless',
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
        runMode: 'run',
      })),
    );
    expect(JSON.parse(localStorage.getItem('gavel.pr-ui.todoRunChoices.v1') || '{}').last.run).toMatchObject({
      backend: 'claude-agent',
      model: 'claude-opus-4-8',
      effort: 'high',
      runMode: 'run',
    });
  });
});
