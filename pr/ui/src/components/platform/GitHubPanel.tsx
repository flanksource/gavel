// The platform dashboard's GitHub panel: the GitHub App's installations with the
// organization each is linked to, its repositories, core rate limit and state.
import { useQuery } from '@tanstack/react-query';
import { fetchJSON, queryKeys } from '../../query';

export interface PlatformRateLimit {
  limit: number;
  remaining: number;
  used: number;
  reset: number;
  resource: string;
}

export interface PlatformInstallation {
  id: number;
  account: string;
  accountType?: string;
  createdAt?: string;
  suspendedAt?: string;
  org?: { id: string; slug: string; name: string };
  repos: number;
  uninstalled?: boolean;
  rateLimit?: PlatformRateLimit;
  rateLimitError?: string;
}

export interface PlatformGitHub {
  app: { id: number; slug: string };
  installations: PlatformInstallation[];
  githubError?: string;
}

function rateLimit(install: PlatformInstallation): string {
  if (install.rateLimitError) return install.rateLimitError;
  if (!install.rateLimit) return '—';
  return `${install.rateLimit.remaining} of ${install.rateLimit.limit} left`;
}

function state(install: PlatformInstallation): string {
  if (install.uninstalled) return 'Uninstalled from GitHub';
  if (install.suspendedAt) return `Suspended ${new Date(install.suspendedAt).toLocaleDateString()}`;
  return '';
}

export function GitHubPanel() {
  const github = useQuery({
    queryKey: queryKeys.platform('github'),
    queryFn: ({ signal }) => fetchJSON<PlatformGitHub>({ url: '/api/platform/github', signal, context: 'load the GitHub App' }),
  });
  const data = github.data;
  const linked = data?.installations.filter(install => install.org).length ?? 0;
  const suspended = data?.installations.filter(install => install.suspendedAt).length ?? 0;

  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-xs font-semibold uppercase text-muted-foreground">
        {data ? `GitHub App ${data.app.slug} · ${data.installations.length} installations, ${linked} linked, ${suspended} suspended` : 'GitHub App'}
      </h2>
      {github.error && <p role="alert" className="text-red-600">{github.error.message}</p>}
      {data?.githubError && <p role="alert" className="text-amber-600">{data.githubError}</p>}
      {!!data?.installations.length && (
        <table className="w-full border-collapse text-left text-xs">
          <thead className="text-muted-foreground">
            <tr>
              {['Account', 'Organization', 'Repositories', 'Rate limit', 'State', 'Installed'].map(heading => (
                <th key={heading} className="border-b border-border px-2 py-1.5 font-medium">{heading}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.installations.map(install => (
              <tr key={install.id} className="border-b border-border">
                <td className="px-2 py-1.5 font-medium" title={install.accountType}>{install.account}</td>
                <td className="px-2 py-1.5">{install.org ? install.org.name || install.org.slug : 'Not linked'}</td>
                <td className="px-2 py-1.5">{install.repos}</td>
                <td className={`px-2 py-1.5 ${install.rateLimitError ? 'text-red-600' : ''}`}>{rateLimit(install)}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{state(install)}</td>
                <td className="px-2 py-1.5 text-muted-foreground">{install.createdAt ? new Date(install.createdAt).toLocaleDateString() : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
