// Connects a hosted organization to the gavel GitHub App: a dialog that links installations
// the user's GitHub account reaches or starts a new install, and the page GitHub returns to.
import { useEffect, useRef, useState } from 'react';
import { useAuth } from '@clerk/react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Modal } from '@flanksource/clicky-ui/components';
import { UiGitPr } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';

// Both requests carry a token fetched just before they're sent: the session cookie
// expires within a minute, and a visit to GitHub's install page can outlast it.

interface Installation {
  id: number;
  account: string;
  accountType: string;
  status: 'linked' | 'linked_elsewhere' | 'available';
}

const buttonClass = 'inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted disabled:opacity-60';

function errorText(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

export function ConnectGitHubButton({ navigate = url => window.location.assign(url) }: { navigate?: (url: string) => void }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        className={buttonClass}
        title="Link a gavel GitHub App installation to this organization, to see its pull requests"
      >
        <UiGitPr className="shrink-0 text-sm" />
        Connect GitHub
      </button>
      {open && <ConnectGitHubDialog navigate={navigate} onClose={() => setOpen(false)} />}
    </>
  );
}

function ConnectGitHubDialog({ navigate, onClose }: { navigate: (url: string) => void; onClose: () => void }) {
  const { getToken } = useAuth();
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const installations = useQuery({
    queryKey: queryKeys.githubInstallations(),
    queryFn: ({ signal }) => fetchJSON<Installation[]>({ url: '/api/github/installations', signal, context: 'List GitHub installations' }),
  });

  const run = async (action: (token: string | undefined) => Promise<void>) => {
    setBusy(true);
    setError('');
    try {
      await action((await getToken()) ?? undefined);
    } catch (cause) {
      setError(errorText(cause));
    } finally {
      setBusy(false);
    }
  };
  const link = (installation: Installation) => run(async token => {
    await mutationJSON({ url: `/api/github/installations/${installation.id}/link`, method: 'POST', token, context: `Link ${installation.account}` });
    await queryClient.invalidateQueries({ queryKey: queryKeys.githubInstallations() });
    // The org's PR poller reads the linked repositories on its next search.
    await mutationJSON({ url: '/api/prs/refresh', method: 'POST', token, context: 'Refresh pull requests' });
  });
  const install = () => run(async token => {
    const { url } = await mutationJSON<{ url: string }>({ url: '/api/github/install', method: 'POST', token, context: 'Install the GitHub App' });
    navigate(url);
  });

  return (
    <Modal
      open
      onClose={onClose}
      title="Connect GitHub"
      size="md"
      footer={(
        <button type="button" className={buttonClass} disabled={busy} onClick={() => void install()}>
          Install on another GitHub account
        </button>
      )}
    >
      <div className="space-y-3 text-sm">
        {installations.isPending && <p className="text-muted-foreground">Reading your GitHub installations…</p>}
        {installations.error && <p className="text-destructive">{errorText(installations.error)}</p>}
        {installations.data?.length === 0 && (
          <p className="text-muted-foreground">
            Your GitHub account can't reach any installation of the gavel App. Install it on a GitHub organization or account.
          </p>
        )}
        {installations.data && installations.data.length > 0 && (
          <ul className="divide-y divide-border rounded border border-border">
            {installations.data.map(installation => (
              <li key={installation.id} className="flex items-center justify-between gap-2 px-3 py-2">
                <span>
                  {installation.account}
                  <span className="ml-1 text-xs text-muted-foreground">{installation.accountType === 'User' ? 'user' : 'organization'}</span>
                </span>
                {installation.status === 'available' ? (
                  <button type="button" className={buttonClass} disabled={busy} aria-label={`Link ${installation.account}`} onClick={() => void link(installation)}>
                    Link
                  </button>
                ) : (
                  <span className="text-xs text-muted-foreground">
                    {installation.status === 'linked' ? 'Linked to this organization' : 'Linked to another gavel organization'}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
        {error && <p className="text-destructive">{error}</p>}
      </div>
    </Modal>
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

  const params = new URLSearchParams(search);
  const fromGitHub = params.has('installation_id') || params.has('state');

  useEffect(() => {
    if (!fromGitHub) return;
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
      (cause: unknown) => current && setState({ status: 'failed', error: errorText(cause) }),
    );
    return () => { current = false; };
  }, [fromGitHub, search]);

  if (!fromGitHub) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="w-full max-w-md space-y-3 text-sm">
          <p>GitHub returns here after the gavel App is installed. To connect GitHub, use Connect GitHub on the dashboard.</p>
          <a href="/" className="text-primary underline">Open the dashboard</a>
        </div>
      </div>
    );
  }

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
