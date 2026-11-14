import { useState } from 'react';
import { sum, type Test as ClickyTest } from '@flanksource/clicky-ui/data';
import { Button, ConfirmDialog, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiFile, UiRestart, UiSkipForward, UiTrash } from '@flanksource/clicky-ui/icons';
import { CopyAIPromptButton, DownloadMenu } from '../components/ExportControls';
import type { RouteState } from '../routes';
import type { Test, TestEditAction, TestEditScope } from '../types';

export interface NodeActionsConfig {
  /** The route state that exports one node (the view's filters with the node's path), if the node can be exported. */
  nodeRouteState: (node: Test) => RouteState | undefined;
  /** The same, narrowed to failing and timed-out tests, for the AI prompt. */
  failingOnlyRouteState: (node: Test) => RouteState | undefined;
  /** status.test_edit_supported: whether the server can edit test sources. */
  testEditSupported: boolean;
  onTestEdit: (node: Test, action: TestEditAction, scope: TestEditScope) => Promise<void>;
  testEditBusy: boolean;
  /** Reruns a fixture file from the given step, reusing the last run's earlier steps. */
  onRetryFromStep: (node: Test, stepId: string) => void;
  /** A rerun is already running: the server accepts one at a time. */
  rerunBusy: boolean;
}

const EDITABLE_FRAMEWORKS = new Set(['go test', 'ginkgo', 'vitest']);

/** The id of the typed fixture step a node is, or undefined for any other node. */
export function fixtureStepId(node: Test): string | undefined {
  if (node.framework !== 'fixture') return undefined;
  const id = (node.detail as { step?: { id?: unknown } } | undefined)?.step?.id;
  return typeof id === 'string' && id !== '' ? id : undefined;
}

export function canEditSource(node: Test, supported: boolean): boolean {
  return supported && node.framework !== 'task' && EDITABLE_FRAMEWORKS.has(node.framework ?? '') && !!node.file;
}

function hasFailingContent(node: Test): boolean {
  if (node.failed || node.timed_out) return true;
  if ((node.children?.length ?? 0) === 0) return false;
  const counts = sum(node as ClickyTest);
  return counts.failed > 0 || counts.timedout > 0;
}

type PendingEdit = { action: TestEditAction; scope: TestEditScope };

function SourceActions({ node, busy, onEdit }: { node: Test; busy: boolean; onEdit: NodeActionsConfig['onTestEdit'] }) {
  const [pending, setPending] = useState<PendingEdit | null>(null);
  const target = pending?.scope === 'file' ? (node.file || 'file') : (node.name || 'test');
  const verb = pending?.action === 'skip' ? 'Skip' : 'Delete';
  const scopeLabel = pending?.scope === 'file' ? 'file' : 'test';
  const choose = (action: TestEditAction, scope: TestEditScope) => () => setPending({ action, scope });
  return (
    <>
      <DropdownMenu
        label="Source"
        title="Edit the test source"
        items={[
          { label: 'Skip Test', icon: UiSkipForward, disabled: busy, onSelect: choose('skip', 'test') },
          { label: 'Delete Test', icon: UiTrash, disabled: busy, onSelect: choose('delete', 'test') },
          { label: 'Skip File', icon: UiFile, disabled: busy, onSelect: choose('skip', 'file') },
          { label: 'Delete File', icon: UiTrash, disabled: busy, onSelect: choose('delete', 'file') },
        ]}
      />
      <ConfirmDialog
        open={pending !== null}
        title={`${verb} ${scopeLabel} ${target}?`}
        message="This edits the source file on disk."
        confirmLabel={verb}
        variant={pending?.action === 'delete' ? 'destructive' : 'default'}
        onCancel={() => setPending(null)}
        onConfirm={() => {
          const edit = pending;
          setPending(null);
          if (edit) void onEdit(node, edit.action, edit.scope);
        }}
      />
    </>
  );
}

/**
 * The header actions of a test node: download and AI-prompt copy of its export,
 * and, for a Go, Ginkgo or vitest test with a source file, the Skip / Delete
 * source edits, and, for a fixture step, Retry from here.
 */
export function NodeActions({ node, config }: { node: Test; config: NodeActionsConfig }) {
  const nodeRouteState = node.framework === 'task' ? undefined : config.nodeRouteState(node);
  const failingOnly = config.failingOnlyRouteState(node);
  const failing = hasFailingContent(node);
  const stepId = fixtureStepId(node);
  return (
    <>
      {stepId && (
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={config.rerunBusy}
          onClick={() => config.onRetryFromStep(node, stepId)}
          title="Rerun this fixture from this step, reusing the earlier steps' results from the last run"
        >
          <UiRestart className="mr-1" />
          Retry from here
        </Button>
      )}
      {nodeRouteState && <DownloadMenu routeState={nodeRouteState} title="Download this node as JSON or Markdown" />}
      {nodeRouteState && failingOnly && (
        <CopyAIPromptButton
          routeState={failingOnly}
          disabled={!failing}
          disabledTitle="No failures to copy"
          title="Copy failing output and repro steps as AI prompt"
        />
      )}
      {canEditSource(node, config.testEditSupported) && (
        <SourceActions node={node} busy={config.testEditBusy} onEdit={config.onTestEdit} />
      )}
    </>
  );
}
