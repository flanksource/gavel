import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCheck, UiCopy, UiDownload, UiJson, UiLoader, UiMarkdown, UiWarningCircle } from '@flanksource/clicky-ui/icons';
import { copyCurrentViewForAgent, downloadCurrentView } from '../export';
import { useCopyAction } from '../hooks/use-copy-feedback';
import { buildExportRoute, type RouteState } from '../routes';

/** Download of the view (or one node of it) as JSON or Markdown, served by the export routes. */
export function DownloadMenu({ routeState, title }: { routeState: RouteState; title?: string }) {
  return (
    <DropdownMenu
      label="Download"
      icon={UiDownload}
      variant="outline"
      size="sm"
      items={[
        { label: 'JSON', icon: UiJson, title: buildExportRoute(routeState, 'json'), onSelect: () => downloadCurrentView(routeState, 'json') },
        { label: 'Markdown', icon: UiMarkdown, title: buildExportRoute(routeState, 'md'), onSelect: () => downloadCurrentView(routeState, 'md') },
      ]}
      {...(title ? { title } : {})}
    />
  );
}

const COPY_LABEL = { idle: 'Copy AI Prompt', copying: 'Copying...', copied: 'Copied', error: 'Copy failed' } as const;
const COPY_ICON = { idle: UiCopy, copying: UiLoader, copied: UiCheck, error: UiWarningCircle } as const;

/** Copies the Markdown export of `routeState` framed as an AI prompt. */
export function CopyAIPromptButton({ routeState, disabled, disabledTitle, title }: { routeState: RouteState; disabled?: boolean; disabledTitle?: string; title?: string }) {
  const { copyState, copyError, run } = useCopyAction(() => copyCurrentViewForAgent(routeState));
  const Icon = COPY_ICON[copyState];
  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      disabled={disabled || copyState === 'copying'}
      onClick={() => { void run(); }}
      title={disabled ? disabledTitle : copyError || title || buildExportRoute(routeState, 'md')}
    >
      <Icon className={copyState === 'copying' ? 'mr-1 animate-spin' : 'mr-1'} />
      {COPY_LABEL[copyState]}
    </Button>
  );
}
