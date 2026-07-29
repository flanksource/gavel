import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoQuestion } from '../../types';
import type { RunContext } from './providers';
import {
  buildTodoAnswerInput,
  PlanApproveButtons,
  QuestionsPanel,
  TodoReviewBanner,
  type TodoQuestionSelections,
} from './planActions';

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
  SpecRuntimeEditor: () => <div>Advanced runtime editor</div>,
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
    UiLoader: Icon,
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

function mockRunContextFailure() {
  vi.stubGlobal('fetch', vi.fn(async () => ({
    ok: false,
    status: 503,
    json: async () => ({ error: 'load run providers from Captain: catalog unavailable' }),
  }) as Response));
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
  it('shows the Captain catalog error inline and disables approval execution', async () => {
    mockRunContextFailure();
    const onApprove = vi.fn();

    render(<PlanApproveButtons onApprove={onApprove} />);

    const alert = await screen.findByRole('alert');
    expect(alert.textContent).toContain('load run providers from Captain: catalog unavailable');
    const approveRun = screen.getByRole('button', { name: 'Approve & Run' });
    expect((approveRun as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(approveRun);
    expect(onApprove).not.toHaveBeenCalled();
  });

  it('approves and runs with the remembered run options', async () => {
    mockRunContext();
    localStorage.setItem(
      'gavel.pr-ui.todoRunChoices.v2',
      JSON.stringify({
        last: {
          run: {
            driver: 'claude-headless',
            runMode: 'run',
            spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' },
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
      runMode: 'run',
      spec: expect.objectContaining({
        backend: 'claude-agent',
        model: 'claude-opus-4-8',
        effort: 'high',
      }),
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
        runMode: 'run',
        spec: expect.objectContaining({
          backend: 'claude-agent',
          model: 'claude-opus-4-8',
          effort: 'high',
        }),
      })),
    );
    expect(JSON.parse(localStorage.getItem('gavel.pr-ui.promptRunChoices.v2') || '{}').approval.last).toMatchObject({
      runMode: 'run',
      spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' },
    });
  });

  it('offers approval-scoped recent configs and an advanced editor', async () => {
    mockRunContext();
    localStorage.setItem(
      'gavel.pr-ui.promptRunChoices.v2',
      JSON.stringify({
        approval: {
          last: { spec: { backend: 'codex-agent', model: 'gpt-5.5', effort: 'medium' } },
          recent: [
            { spec: { backend: 'claude-agent', model: 'claude-opus-4-8', effort: 'high' } },
          ],
        },
      }),
    );
    const onApprove = vi.fn();

    render(<PlanApproveButtons onApprove={onApprove} />);

    expect(await screen.findByText('Recent configs')).toBeTruthy();
    expect(screen.getByRole('button', { name: /Agent:opus-4\.8/ })).toBeTruthy();
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Advanced' }));
    });
    expect(screen.getByText('Advanced runtime editor')).toBeTruthy();
  });
});

const questions: TodoQuestion[] = [
  {
    text: 'Which database?',
    context: 'Choose the persistence layer.',
    options: ['Postgres', 'SQLite'],
  },
  {
    text: 'Which environment?',
    options: ['Development', 'Production'],
  },
];

function QuestionsHarness({ panels = 1, disabled = false }: { panels?: number; disabled?: boolean }) {
  const [selections, setSelections] = React.useState<TodoQuestionSelections>({});
  return (
    <>
      {Array.from({ length: panels }, (_, index) => (
        <QuestionsPanel
          key={index}
          questions={questions}
          selections={selections}
          disabled={disabled}
          onSelectionChange={(questionIndex, option) => {
            setSelections(previous => ({ ...previous, [questionIndex]: option }));
          }}
        />
      ))}
    </>
  );
}

function askTodo(ref = 'todo-1'): TodoItem {
  return {
    ref,
    title: `Question ${ref}`,
    status: 'ask',
    priority: 'medium',
    questions,
  };
}

describe('buildTodoAnswerInput', () => {
  it('builds mutually exclusive structured answers with normalized details', () => {
    expect(buildTodoAnswerInput(questions, { 0: ' Postgres ', 1: 'Development' }, '  Keep migrations reversible.  ')).toEqual({
      answers: {
        'Which database?': 'Postgres',
        'Which environment?': 'Development',
        'Additional details': 'Keep migrations reversible.',
      },
    });
  });

  it('keeps the legacy free-text answer when no option is selected', () => {
    expect(buildTodoAnswerInput(questions, {}, '  Use the existing database.  ')).toEqual({
      answer: 'Use the existing database.',
    });
  });

  it('disambiguates duplicate question text without dropping either selection', () => {
    const duplicateQuestions = [
      { text: 'Choose scope', options: ['Local'] },
      { text: 'Choose scope', options: ['Global'] },
    ];
    expect(buildTodoAnswerInput(duplicateQuestions, { 0: 'Local', 1: 'Global' }, '')).toEqual({
      answers: {
        'Choose scope (1)': 'Local',
        'Choose scope (2)': 'Global',
      },
    });
  });
});

describe('QuestionsPanel', () => {
  it('renders clickable native radio groups with independent question selections', () => {
    render(<QuestionsHarness />);

    expect(screen.getAllByRole('radio')).toHaveLength(4);
    fireEvent.click(screen.getByLabelText('Postgres'));
    fireEvent.click(screen.getByLabelText('Production'));

    expect((screen.getByLabelText('Postgres') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('SQLite') as HTMLInputElement).checked).toBe(false);
    expect((screen.getByLabelText('Production') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('Development') as HTMLInputElement).checked).toBe(false);

    fireEvent.click(screen.getByLabelText('SQLite'));
    expect((screen.getByLabelText('Postgres') as HTMLInputElement).checked).toBe(false);
    expect((screen.getByLabelText('SQLite') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('Production') as HTMLInputElement).checked).toBe(true);
  });

  it('uses distinct radio names for separate panel instances and disables controls', () => {
    render(<QuestionsHarness panels={2} disabled />);

    const databaseRadios = screen.getAllByRole('radio', { name: /Postgres|SQLite/ });
    expect(databaseRadios[0].getAttribute('name')).toBe(databaseRadios[1].getAttribute('name'));
    expect(databaseRadios[0].getAttribute('name')).not.toBe(databaseRadios[2].getAttribute('name'));
    for (const radio of screen.getAllByRole('radio')) expect((radio as HTMLInputElement).disabled).toBe(true);
  });
});

describe('TodoReviewBanner', () => {
  it('enables quick selection and submits structured answers without a flat answer', async () => {
    const todo = askTodo();
    const onChanged = vi.fn();
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ todo: { ...todo, status: 'in_progress' }, status: 'started' }),
    });
    vi.stubGlobal('fetch', fetchMock);
    render(<TodoReviewBanner todo={todo} dir="/workspace" onChanged={onChanged} />);

    const submit = screen.getByRole('button', { name: 'Send & resume' });
    expect((submit as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(screen.getByLabelText('Postgres'));
    fireEvent.click(screen.getByLabelText('Production'));
    expect((submit as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(submit);

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    const request = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      answers: {
        'Which database?': 'Postgres',
        'Which environment?': 'Production',
      },
    });
    expect((screen.getByLabelText('Postgres') as HTMLInputElement).checked).toBe(false);
    expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe('');
  });

  it('submits additional details inside answers and preserves drafts after failure', async () => {
    const todo = askTodo();
    const fetchMock = vi.fn().mockResolvedValue({ ok: false, json: async () => ({ error: 'resume failed' }) });
    vi.stubGlobal('fetch', fetchMock);
    render(<TodoReviewBanner todo={todo} dir="/workspace" onChanged={vi.fn()} />);

    fireEvent.click(screen.getByLabelText('SQLite'));
    fireEvent.change(screen.getByRole('textbox'), { target: { value: '  Keep the adapter boundary.  ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send & resume' }));

    await screen.findByText('resume failed');
    const request = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({
      ref: 'todo-1',
      answers: {
        'Which database?': 'SQLite',
        'Additional details': 'Keep the adapter boundary.',
      },
    });
    expect((screen.getByLabelText('SQLite') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe('  Keep the adapter boundary.  ');
  });

  it('keeps free-text-only submission and resets drafts when the todo changes', async () => {
    const first = askTodo();
    const second = askTodo('todo-2');
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ todo: { ...first, status: 'in_progress' }, status: 'started' }),
    });
    vi.stubGlobal('fetch', fetchMock);
    const onChanged = vi.fn();
    const { rerender } = render(<TodoReviewBanner todo={first} dir="/workspace" onChanged={onChanged} />);

    fireEvent.change(screen.getByRole('textbox'), { target: { value: '  Use the current default.  ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send & resume' }));
    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(JSON.parse(String((fetchMock.mock.calls[0][1] as RequestInit).body))).toEqual({
      ref: 'todo-1',
      answer: 'Use the current default.',
    });

    fireEvent.click(screen.getByLabelText('Production'));
    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'draft for first' } });
    rerender(<TodoReviewBanner todo={second} dir="/workspace" onChanged={vi.fn()} />);
    expect((screen.getByLabelText('Production') as HTMLInputElement).checked).toBe(false);
    expect((screen.getByRole('textbox') as HTMLTextAreaElement).value).toBe('');
  });

  it('disables radio controls while the answer request is in flight', async () => {
    let resolveFetch: ((value: unknown) => void) | undefined;
    vi.stubGlobal('fetch', vi.fn(() => new Promise(resolve => {
      resolveFetch = resolve;
    })));
    render(<TodoReviewBanner todo={askTodo()} dir="/workspace" onChanged={vi.fn()} />);

    fireEvent.click(screen.getByLabelText('Postgres'));
    fireEvent.click(screen.getByRole('button', { name: 'Send & resume' }));
    await waitFor(() => expect((screen.getByLabelText('Postgres') as HTMLInputElement).disabled).toBe(true));

    resolveFetch?.({ ok: true, json: async () => ({ todo: askTodo(), status: 'started' }) });
    await waitFor(() => expect((screen.getByLabelText('Postgres') as HTMLInputElement).disabled).toBe(false));
  });
});
