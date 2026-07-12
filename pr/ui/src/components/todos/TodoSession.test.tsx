import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ApprovalBanner, formatQuestionApprovalMessage } from './TodoSession';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionViewer: () => <div data-testid="session-viewer" />,
  questionsFromToolInput: (input?: Record<string, unknown>) => {
    if (!input) return [];
    const raw = input.questions;
    if (Array.isArray(raw)) {
      return raw.map((q, index) => {
        const record = q as Record<string, unknown>;
        const text = String(record.question || record.text || record.prompt || '').trim();
        if (!text) return null;
        return {
          id: String(record.id || index + 1),
          text,
          context: typeof record.header === 'string' ? record.header : undefined,
          multiSelect: record.multiSelect === true,
          options: Array.isArray(record.options)
            ? record.options.map(option => {
                if (typeof option === 'string') return { value: option, label: option };
                const optionRecord = option as Record<string, unknown>;
                return {
                  value: String(optionRecord.value || optionRecord.id || optionRecord.label),
                  label: String(optionRecord.label),
                  description: typeof optionRecord.description === 'string' ? optionRecord.description : undefined,
                };
              })
            : [],
        };
      }).filter(Boolean);
    }
    const text = String(input.question || input.prompt || input.text || '').trim();
    if (!text) return [];
    return [{
      id: '1',
      text,
      context: typeof input.header === 'string' ? input.header : undefined,
      multiSelect: false,
      options: [],
    }];
  },
}));

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({
    children,
    variant: _variant,
    size: _size,
    ...props
  }: React.ButtonHTMLAttributes<HTMLButtonElement> & { variant?: string; size?: string }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button" {...props}>
      {children}
    </button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => {
  const Icon = (props: React.SVGProps<SVGSVGElement>) => <svg {...props} />;
  return {
    ...(await importOriginal<object>()),
    UiCancel: Icon,
    UiCheck: Icon,
    UiCircleFilled: Icon,
    UiComment: Icon,
    UiError: Icon,
    UiLightbulb: Icon,
    UiPass: Icon,
    UiShield: Icon,
  };
});

describe('ApprovalBanner', () => {
  it('shows AskUserQuestion options and submits the selected answer', () => {
    const onDecide = vi.fn();
    render(
      <ApprovalBanner
        approval={{
          sessionId: 'session-1',
          toolUseId: 'tool-1',
          tool: 'AskUserQuestion',
          input: {
            questions: [{
              question: 'Which scope should the rule use?',
              header: 'Scope',
              options: [{ label: 'Project', description: 'Current workspace' }, { label: 'Global' }],
            }],
          },
        }}
        busy={false}
        onDecide={onDecide}
      />,
    );

    expect(screen.getByText('Which scope should the rule use?')).toBeTruthy();
    expect(screen.getByText('Scope')).toBeTruthy();
    expect(screen.getByText('Project')).toBeTruthy();
    expect(screen.getByText('Current workspace')).toBeTruthy();

    fireEvent.click(screen.getByLabelText(/Project/));
    fireEvent.change(screen.getByPlaceholderText(/Additional details/), { target: { value: 'Use the project scope' } });
    fireEvent.click(screen.getByRole('button', { name: 'Send answer' }));

    expect(onDecide).toHaveBeenCalledWith(true, 'Scope: Project\nAdditional details: Use the project scope');
  });

  it('shows normal permission approval details and keeps the allow action', () => {
    const onDecide = vi.fn();
    render(
      <ApprovalBanner
        approval={{
          sessionId: 'session-1',
          tool: 'Bash',
          input: { command: 'npm test' },
        }}
        busy={false}
        onDecide={onDecide}
      />,
    );

    expect(screen.getAllByText('npm test').length).toBeGreaterThanOrEqual(1);
    fireEvent.click(screen.getByRole('button', { name: 'Allow' }));
    expect(onDecide).toHaveBeenCalledWith(true);
  });

  it('submits reject comments for permission approvals', () => {
    const onDecide = vi.fn();
    render(
      <ApprovalBanner
        approval={{
          sessionId: 'session-1',
          tool: 'Bash',
          input: { command: 'npm test' },
        }}
        busy={false}
        onDecide={onDecide}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Reject with comment' }));
    fireEvent.change(screen.getByPlaceholderText(/why this request is rejected/), {
      target: { value: 'Tests are already running' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Reject with comment' }));

    expect(onDecide).toHaveBeenCalledWith(false, 'Tests are already running');
  });

  it('keeps a free-text answer fallback for unstructured question payloads', () => {
    const onDecide = vi.fn();
    render(
      <ApprovalBanner
        approval={{
          sessionId: 'session-1',
          toolUseId: 'tool-1',
          tool: 'AskUserQuestion',
          input: { questions: [{ options: ['Yes', 'No'] }] },
        }}
        busy={false}
        onDecide={onDecide}
      />,
    );

    fireEvent.change(screen.getByPlaceholderText("Answer the agent's question..."), {
      target: { value: 'Use the project scope' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Send answer' }));

    expect(onDecide).toHaveBeenCalledWith(true, 'Use the project scope');
  });
});

describe('formatQuestionApprovalMessage', () => {
  it('formats selected options, free text and additional details', () => {
    expect(formatQuestionApprovalMessage([
      {
        id: 'scope',
        text: 'Which scope?',
        context: 'Scope',
        options: [
          { value: 'project', label: 'Project' },
          { value: 'global', label: 'Global' },
        ],
      },
      {
        id: 'reason',
        text: 'Why?',
        options: [],
      },
    ], {
      scope: 'project',
      reason: 'It matches the current workspace',
    }, 'Keep it narrow')).toBe([
      'Scope: Project',
      'reason: It matches the current workspace',
      'Additional details: Keep it narrow',
    ].join('\n'));
  });

  it('formats fallback answers without a label when no question parsed', () => {
    expect(formatQuestionApprovalMessage([], {}, 'Use the project scope')).toBe('Use the project scope');
  });
});
