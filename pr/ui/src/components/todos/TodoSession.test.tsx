import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ApprovalBanner, approvalQuestions } from './TodoSession';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionViewer: () => <div data-testid="session-viewer" />,
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
    UiCircleFilled: Icon,
    UiComment: Icon,
    UiError: Icon,
    UiLightbulb: Icon,
    UiPass: Icon,
    UiShield: Icon,
  };
});

vi.mock('./planActions', () => ({
  AnswerBox: ({
    value,
    onChange,
    onSubmit,
    busy,
    label,
    placeholder,
  }: {
    value: string;
    onChange: (value: string) => void;
    onSubmit: () => void;
    busy?: boolean;
    label?: string;
    placeholder?: string;
  }) => (
    <div>
      <textarea
        placeholder={placeholder}
        value={value}
        onChange={event => onChange(event.currentTarget.value)}
      />
      <button type="button" disabled={busy || !value.trim()} onClick={onSubmit}>
        {label}
      </button>
    </div>
  ),
  QuestionsPanel: ({
    questions,
  }: {
    questions: Array<{ text: string; context?: string; options?: string[] }>;
  }) => (
    <div>
      {questions.map(q => (
        <div key={q.text}>
          <div>{q.text}</div>
          {q.context && <div>{q.context}</div>}
          {q.options?.map(option => <span key={option}>{option}</span>)}
        </div>
      ))}
    </div>
  ),
}));

describe('ApprovalBanner', () => {
  it('shows AskUserQuestion content and submits the typed answer', () => {
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
    expect(screen.getByText('Project - Current workspace')).toBeTruthy();

    fireEvent.change(screen.getByPlaceholderText(/Answer the agent/), {
      target: { value: 'Use the project scope' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Send answer' }));

    expect(onDecide).toHaveBeenCalledWith(true, 'Use the project scope');
  });

  it('keeps the normal allow action for permission approvals', () => {
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

    expect(screen.getByText('npm test')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Allow' }));
    expect(onDecide).toHaveBeenCalledWith(true);
  });
});

describe('approvalQuestions', () => {
  it('normalizes single question and options payloads', () => {
    expect(approvalQuestions({
      question: 'Run the migration?',
      header: 'Database',
      options: ['Yes', { label: 'No', description: 'Skip it' }],
    })).toEqual([{
      text: 'Run the migration?',
      context: 'Database',
      options: ['Yes', 'No - Skip it'],
    }]);
  });
});
