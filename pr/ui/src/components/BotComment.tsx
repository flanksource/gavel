import { useState } from 'react';
import type { PRComment } from '../types';
import { Markdown } from './Markdown';
import { GavelIcon } from './GavelIcon';

interface Props {
  comment: PRComment;
}

export function BotCommentBody({ comment }: Props) {
  switch (comment.botType) {
    case 'vercel': return <VercelComment comment={comment} />;
    case 'copilot': return <CopilotComment comment={comment} />;
    case 'coderabbit': return <CodeRabbitComment comment={comment} />;
    case 'gavel': return <GavelComment comment={comment} />;
    default: return <Markdown text={comment.body} />;
  }
}

// --- Vercel ---

const vercelUrlRegex = /https:\/\/[^\s|)]+\.vercel\.app[^\s|)"]*/;

function extractVercelPreviewUrl(body: string): string | null {
  const m = body.match(vercelUrlRegex);
  return m ? m[0] : null;
}

function stripVercelNoise(body: string): string {
  return body
    .replace(/\*\*The latest updates on your projects\*\*.*$/s, '')
    .trim();
}

function VercelComment({ comment }: Props) {
  const previewUrl = extractVercelPreviewUrl(comment.body);
  const [showFull, setShowFull] = useState(false);

  return (
    <div className="text-xs">
      {previewUrl && (
        <a
          href={previewUrl}
          target="_blank"
          rel="noopener"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-md text-blue-700 hover:bg-blue-100 transition-colors mb-2"
        >
          <GavelIcon name="codicon:link-external" className="text-xs" />
          <span className="font-medium">Preview Deployment</span>
        </a>
      )}
      <button
        className="text-[11px] text-gray-400 hover:text-gray-600 flex items-center gap-1"
        onClick={() => setShowFull(!showFull)}
      >
        <GavelIcon name={showFull ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-[9px]" />
        {showFull ? 'Hide details' : 'Show details'}
      </button>
      {showFull && <Markdown text={comment.body} className="text-xs text-gray-600 mt-1" />}
    </div>
  );
}

// --- Copilot ---

const suggestionFenceRegex = /```suggestion\n([\s\S]*?)```/g;

function CopilotComment({ comment }: Props) {
  const hasSuggestion = suggestionFenceRegex.test(comment.body);
  suggestionFenceRegex.lastIndex = 0;

  if (!hasSuggestion) {
    return <Markdown text={comment.body} className="text-xs text-gray-700" />;
  }

  const parts: { type: 'text' | 'suggestion'; content: string }[] = [];
  let lastIdx = 0;
  let match: RegExpExecArray | null;
  while ((match = suggestionFenceRegex.exec(comment.body)) !== null) {
    if (match.index > lastIdx) {
      parts.push({ type: 'text', content: comment.body.slice(lastIdx, match.index) });
    }
    parts.push({ type: 'suggestion', content: match[1] });
    lastIdx = match.index + match[0].length;
  }
  if (lastIdx < comment.body.length) {
    parts.push({ type: 'text', content: comment.body.slice(lastIdx) });
  }

  return (
    <div className="text-xs">
      {parts.map((part, i) =>
        part.type === 'suggestion' ? (
          <div key={i} className="my-1.5">
            <div className="text-[10px] text-green-700 font-medium mb-0.5 flex items-center gap-1">
              <GavelIcon name="codicon:lightbulb" />
              Suggested change
            </div>
            <pre className="bg-green-50 border border-green-200 rounded p-2 text-xs overflow-x-auto">
              <code>{part.content}</code>
            </pre>
          </div>
        ) : (
          <Markdown key={i} text={part.content.trim()} className="text-xs text-gray-700" />
        )
      )}
    </div>
  );
}

// --- CodeRabbit ---

const coderabbitSections = [
  { prefix: '📝 Walkthrough', label: 'Walkthrough' },
  { prefix: '📋 Walkthrough', label: 'Walkthrough' },
  { prefix: 'Walkthrough', label: 'Walkthrough' },
  { prefix: '## Changes', label: 'Changes' },
];

function CodeRabbitComment({ comment }: Props) {
  if (comment.severity === 'nitpick' || comment.path) {
    return <Markdown text={comment.body} className="text-xs text-gray-700" />;
  }

  const actionableMatch = comment.body.match(/\*\*Actionable comments posted:\s*(\d+)\*\*/);
  const actionableCount = actionableMatch ? parseInt(actionableMatch[1], 10) : null;

  const [showFull, setShowFull] = useState(false);

  return (
    <div className="text-xs">
      {actionableCount !== null && (
        <div className="flex items-center gap-1.5 text-sm font-medium text-gray-700 mb-1.5">
          <GavelIcon name="codicon:comment-discussion" className="text-purple-500" />
          {actionableCount} actionable comment{actionableCount !== 1 ? 's' : ''} posted
        </div>
      )}
      <button
        className="text-[11px] text-gray-400 hover:text-gray-600 flex items-center gap-1"
        onClick={() => setShowFull(!showFull)}
      >
        <GavelIcon name={showFull ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-[9px]" />
        {showFull ? 'Hide full review' : 'Show full review'}
      </button>
      {showFull && <Markdown text={comment.body} className="text-xs text-gray-600 mt-1" />}
    </div>
  );
}

// --- Gavel ---

const gavelSummaryRegex = /(\d+)\s+passed.*?(\d+)\s+failed/i;
const gavelLintRegex = /(\d+)\s+violation/i;

function GavelComment({ comment }: Props) {
  const body = comment.body
    .replace(/<!--[^>]*-->/g, '')
    .trim();

  const testMatch = body.match(gavelSummaryRegex);
  const lintMatch = body.match(gavelLintRegex);
  const [showFull, setShowFull] = useState(false);

  return (
    <div className="text-xs">
      <div className="flex items-center gap-2 mb-1.5">
        {testMatch && (
          <span className="text-gray-700">
            Tests: <span className="text-green-600 font-medium">{testMatch[1]} passed</span>
            {parseInt(testMatch[2], 10) > 0 && (
              <span className="text-red-600 font-medium ml-1">{testMatch[2]} failed</span>
            )}
          </span>
        )}
        {lintMatch && (
          <span className="text-gray-700">
            Lint: <span className={parseInt(lintMatch[1], 10) > 0 ? 'text-yellow-600 font-medium' : 'text-green-600 font-medium'}>
              {lintMatch[1]} violation{parseInt(lintMatch[1], 10) !== 1 ? 's' : ''}
            </span>
          </span>
        )}
      </div>
      <button
        className="text-[11px] text-gray-400 hover:text-gray-600 flex items-center gap-1"
        onClick={() => setShowFull(!showFull)}
      >
        <GavelIcon name={showFull ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="text-[9px]" />
        {showFull ? 'Hide details' : 'Show details'}
      </button>
      {showFull && <Markdown text={body} className="text-xs text-gray-600 mt-1" />}
    </div>
  );
}

// --- Bot badge ---

const botLabels: Record<string, { icon: string; label: string; color: string }> = {
  vercel: { icon: 'simple-icons:vercel', label: 'Vercel', color: 'text-gray-700' },
  copilot: { icon: 'simple-icons:githubcopilot', label: 'Copilot', color: 'text-purple-600' },
  coderabbit: { icon: 'codicon:hubot', label: 'CodeRabbit', color: 'text-orange-600' },
  gavel: { icon: 'codicon:beaker', label: 'Gavel', color: 'text-blue-600' },
};

export function BotBadge({ botType }: { botType: string }) {
  const info = botLabels[botType];
  if (!info) return null;
  return (
    <span className={`inline-flex items-center gap-0.5 text-[10px] ${info.color} opacity-70`}>
      <GavelIcon name={info.icon} className="text-[10px]" />
      {info.label}
    </span>
  );
}
