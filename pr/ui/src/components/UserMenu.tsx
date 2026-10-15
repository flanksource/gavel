// Shows who is signed in: Clerk's org switcher and user button (with an SSH keys page)
// in clerk mode, or a badge with the local `gavel login` identity in none mode.
import { OrganizationSwitcher, UserButton } from '@clerk/react';
import { useQuery } from '@tanstack/react-query';
import { UiGitPr, UiKey, UiUser } from '@flanksource/clicky-ui/icons';
import { currentAuthMode } from '../query';
import { meQuery } from './oneShotQueries';
import { SSHKeysPanel } from './SSHKeysPanel';

export function UserMenu() {
  if (currentAuthMode() === 'clerk') {
    // Switching organizations reloads the page (Clerk navigates to "/" without a
    // router), which clears cached queries and reopens event streams for the new org.
    return (
      <div className="flex items-center gap-2">
        <a
          href="/api/github/install"
          className="inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs text-muted-foreground hover:bg-muted"
          title="Install the gavel GitHub App on a GitHub organization or account, to see its pull requests"
        >
          <UiGitPr className="shrink-0 text-sm" />
          Connect GitHub
        </a>
        <OrganizationSwitcher hidePersonal afterSelectOrganizationUrl="/" />
        <UserButton>
          <UserButton.UserProfilePage label="SSH keys" url="ssh-keys" labelIcon={<UiKey />}>
            <SSHKeysPanel />
          </UserButton.UserProfilePage>
        </UserButton>
      </div>
    );
  }
  return <LocalIdentity />;
}

function LocalIdentity() {
  const me = useQuery(meQuery());
  const name = me.data?.user?.name || 'local';
  return (
    <span
      className="inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs text-muted-foreground"
      title={me.data?.user?.email ? `Logged in as ${me.data.user.email}` : 'Not logged in; run gavel login'}
    >
      <UiUser className="shrink-0 text-sm" />
      <span className="truncate">{name}</span>
    </span>
  );
}
