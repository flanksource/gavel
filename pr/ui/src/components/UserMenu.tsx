// Shows who is signed in: Clerk's org switcher (with model keys, and usage and budgets pages) and user
// button (with an SSH keys page) in clerk mode, or a menu on the local `gavel login` identity in none
// mode. Both user menus carry the Grabber help.
import { useState } from 'react';
import { OrganizationSwitcher, UserButton } from '@clerk/react';
import { useQuery } from '@tanstack/react-query';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiChartBar, UiKey, UiSearch, UiUser } from '@flanksource/clicky-ui/icons';
import { currentAuthMode } from '../query';
import { ConnectGitHubButton } from './GitHubConnect';
import { GrabberExtensionHelpModal, GRABBER_HELP_TITLE } from './GrabberExtensionHelp';
import { ModelKeysPanel } from './ModelKeysPanel';
import { OrgBudgets } from './OrgBudgets';
import { OwnMeasuredUsage } from './platform/MeasuredUsagePanel';
import { meQuery } from './oneShotQueries';
import { SSHKeysPanel } from './SSHKeysPanel';
import { UsagePanel } from './UsagePanel';

export function UserMenu() {
  const [grabberOpen, setGrabberOpen] = useState(false);
  const openGrabber = () => setGrabberOpen(true);
  return (
    <>
      {currentAuthMode() === 'clerk' ? <ClerkUserMenu onGrabber={openGrabber} /> : <LocalIdentity onGrabber={openGrabber} />}
      {grabberOpen && <GrabberExtensionHelpModal onClose={() => setGrabberOpen(false)} />}
    </>
  );
}

function ClerkUserMenu({ onGrabber }: { onGrabber: () => void }) {
  // Switching organizations reloads the page (Clerk navigates to "/" without a
  // router), which clears cached queries and reopens event streams for the new org.
  return (
    <div className="flex items-center gap-2">
      <ConnectGitHubButton />
      <OrganizationSwitcher hidePersonal afterSelectOrganizationUrl="/">
        <OrganizationSwitcher.OrganizationProfilePage label="Model keys" url="model-keys" labelIcon={<UiKey />}>
          <ModelKeysPanel />
        </OrganizationSwitcher.OrganizationProfilePage>
        <OrganizationSwitcher.OrganizationProfilePage label="Usage and budgets" url="usage" labelIcon={<UiChartBar />}>
          <div className="flex flex-col gap-6">
            <UsagePanel />
            <OwnMeasuredUsage />
            <OrgBudgets />
          </div>
        </OrganizationSwitcher.OrganizationProfilePage>
      </OrganizationSwitcher>
      <UserButton>
        <UserButton.MenuItems>
          <UserButton.Action label="Grabber" labelIcon={<UiSearch />} onClick={onGrabber} />
        </UserButton.MenuItems>
        <UserButton.UserProfilePage label="SSH keys" url="ssh-keys" labelIcon={<UiKey />}>
          <SSHKeysPanel />
        </UserButton.UserProfilePage>
      </UserButton>
    </div>
  );
}

function LocalIdentity({ onGrabber }: { onGrabber: () => void }) {
  const me = useQuery(meQuery());
  const name = me.data?.user?.name || 'local';
  return (
    <DropdownMenu
      align="right"
      menuLabel="User menu"
      trigger={
        <Button
          type="button"
          variant="ghost"
          aria-label={`User menu: ${name}`}
          className="inline-flex h-auto items-center gap-1 rounded border border-border px-2 py-1 text-xs font-normal text-muted-foreground hover:bg-muted"
          title={me.data?.user?.email ? `Logged in as ${me.data.user.email}` : 'Not logged in; run gavel login'}
        >
          <UiUser className="shrink-0 text-sm" />
          <span className="truncate">{name}</span>
        </Button>
      }
      items={[{ label: 'Grabber', icon: UiSearch, title: GRABBER_HELP_TITLE, onSelect: onGrabber }]}
    />
  );
}
