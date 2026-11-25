import { useEffect, useRef, useState } from 'react';
import { PromptDialog } from '@flanksource/clicky-ui/data';
import { answerPrompt, usePrompts, type PromptSnapshot } from '@flanksource/clicky-ui/hooks';

const basePath = '/api/todos';

export function useProjectGitLockRequest() {
  const [owner] = useState(() => `project-git-lock-${crypto.getRandomValues(new Uint32Array(4)).join('-')}`);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => {
    const active = new AbortController();
    controller.current = active;
    return () => active.abort();
  }, []);

  return {
    owner,
    request: () => ({ headers: { 'X-Gavel-Prompt-Owner': owner }, ...(controller.current ? { signal: controller.current.signal } : {}) }),
  };
}

export function ProjectGitLockPrompt({ owner, enabled }: { owner: string; enabled: boolean }) {
  const { pending } = usePrompts({ basePath, owner, enabled });
  const prompt = pending[0];
  return enabled && prompt ? <GitLockDialog key={prompt.id} prompt={prompt} /> : null;
}

function GitLockDialog({ prompt }: { prompt: PromptSnapshot }) {
  const resolved = useRef(false);
  const [error, setError] = useState('');
  return <>
    <PromptDialog
      prompt={prompt}
      basePath={basePath}
      open
      onResolved={() => { resolved.current = true; }}
      onClose={() => {
        if (!resolved.current) {
          void answerPrompt(basePath, prompt.id, { cancelled: true }).catch(cause => setError(String(cause)));
        }
      }}
    />
    {error && <div role="alert">{error}</div>}
  </>;
}
