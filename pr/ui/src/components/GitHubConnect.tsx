// Connects a hosted organization to the gavel GitHub App: the button that starts GitHub's
// install flow, and the page GitHub returns to, which links the installation made there.
import { useEffect, useRef, useState } from 'react';
import { useAuth } from '@clerk/react';
import { UiGitPr } from '@flanksource/clicky-ui/icons';
import { mutationJSON } from '../query';

// Both requests carry a token fetched just before they're sent: the session cookie
// expires within a minute, and a visit to GitHub's install page can outlast it.

export function ConnectGitHubButton({ navigate = url => window.location.assign(url) }: { navigate?: (url: string) => void }) {
  const { getToken } = useAuth();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const connect = async () => {
    setBusy(true);
    setError('');
    try {
      const token = (await getToken()) ?? undefined;
      const { url } = await mutationJSON<{ url: string }>({ url: '/api/github/install', method: 'POST', token, context: 'Connect GitHub' });
      navigate(url);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
      setBusy(false);
    }
  };

  return (
    <span className="inline-flex items-center gap-2">
      <button
        type="button"
        onClick={() => void connect()}
        disabled={busy}
        className="inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted disabled:opacity-60"
        title="Install the gavel GitHub App on a GitHub organization or account, to see its pull requests"
      >
        <UiGitPr className="shrink-0 text-sm" />
        Connect GitHub
      </button>
      {error && <span className="max-w-xs truncate text-xs text-destructive" title={error}>{error}</span>}
    </span>
  );
}

type SetupState =
  | { status: 'linking' }
  | { status: 'linked'; repositories: string[] }
  | { status: 'failed'; error: string };

// GitHubSetupPage is where GitHub sends the user after they install the App; search is
// the query GitHub added (installation_id, setup_action, state).
export function GitHubSetupPage({ search = window.location.search }: { search?: string }) {
  const { getToken } = useAuth();
  // Read through a ref: getToken's identity can change between renders, and the link
  // must be requested once per visit.
  const getTokenRef = useRef(getToken);
  getTokenRef.current = getToken;
  const [state, setState] = useState<SetupState>({ status: 'linking' });

  useEffect(() => {
    let current = true;
    const params = new URLSearchParams(search);
    const link = async () => {
      const token = (await getTokenRef.current()) ?? undefined;
      return mutationJSON<{ repositories: string[] }>({
        url: '/api/github/setup',
        method: 'POST',
        token,
        body: { installation_id: params.get('installation_id') ?? '', state: params.get('state') ?? '' },
        context: 'Connect GitHub',
      });
    };
    link().then(
      ({ repositories }) => current && setState({ status: 'linked', repositories: repositories ?? [] }),
      (cause: unknown) => current && setState({ status: 'failed', error: cause instanceof Error ? cause.message : String(cause) }),
    );
    return () => { current = false; };
  }, [search]);

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md space-y-3 text-sm">
        {state.status === 'linking' && <p className="text-muted-foreground">Connecting GitHub…</p>}
        {state.status === 'linked' && (
          <>
            <h1 className="text-base font-semibold">GitHub connected</h1>
            {state.repositories.length === 0 ? (
              <p className="text-muted-foreground">
                The installation shares no repositories with gavel yet. Choose repositories in the App's settings on GitHub.
              </p>
            ) : (
              <ul className="list-inside list-disc">
                {state.repositories.map(repo => <li key={repo}>{repo}</li>)}
              </ul>
            )}
            <a href="/" className="text-primary underline">Open the dashboard</a>
          </>
        )}
        {state.status === 'failed' && (
          <>
            <h1 className="text-base font-semibold">GitHub wasn't connected</h1>
            <p className="text-destructive">{state.error}</p>
            <a href="/" className="text-primary underline">Back to the dashboard</a>
          </>
        )}
      </div>
    </div>
  );
}
