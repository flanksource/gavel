import { useEffect, useMemo, useRef, useState, type ComponentType, type ReactNode } from 'react';
import {
  Modal,
  Field,
  Button,
  Select,
  SplitPane,
  ListMenu,
  ListMenuItem,
  ListMenuHeader,
  useListMenuSelection,
  type HostedModalProps,
} from '@flanksource/clicky-ui/components';
import type { AcceptanceCriterion, PRDetail, PRItem, Project, TodoPriority, TodoStatus } from '../../types';
import { ansiToHtml } from '../../ansi';
import { Markdown } from '../Markdown';
import { Avatar } from '../Avatar';
import { AddProjectDialog } from '../AddProjectDialog';
import { PRTodoWorkspaceField } from './PRTodoWorkspaceField';
import { usePRTodoProjectWorkspace } from './usePRTodoProjectWorkspace';
import { useIsMobile } from '../../useIsMobile';
import { UiArrowLeft, UiBeaker, UiComment, UiError, UiLinkExternal, UiWarningTriangle, type IconProps } from '@flanksource/clicky-ui/icons';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { TodoMutationError, useCreateTodoMutation, type CreateTodoResponse } from './todoMutations';
import { RunAfterCreation, useRunAfterCreation, type RunAfterCreationState } from './RunAfterCreation';
import { savePRTodoDraft, type PRTodoDraft } from './prTodoDraft';
import {
  buildPRTodoBody,
  buildPRTodoCandidates,
  buildPRTodoVerification,
  isPRTodoCriterion,
  type PRTodoCandidate,
  type PRTodoCandidateDetail,
  type PRTodoSourceGroup,
  type PRTodoVerificationPayload,
} from './PRTodoContent';

// GROUPS drives both the grouped checkbox lists and the one-click presets: each
// preset titles the todo and narrows the selection to that group ("Fix failing
// tests in repo#7", etc.).
const GROUPS: { id: PRTodoSourceGroup; label: string; icon: ComponentType<IconProps>; preset: string }[] = [
  { id: 'tests', label: 'Failing tests', icon: UiBeaker, preset: 'Fix failing tests' },
  { id: 'lint', label: 'Lint violations', icon: UiWarningTriangle, preset: 'Fix lint violations' },
  { id: 'checks', label: 'Failing checks', icon: UiError, preset: 'Fix failing checks' },
  { id: 'comments', label: 'Review comments', icon: UiComment, preset: 'Resolve review comments' },
];

// Section is one rendered group in the left list: tests/lint/checks each map to a
// single section; review comments are sub-grouped into one section per author.
interface Section {
  id: string;
  label: string;
  icon: ComponentType<IconProps>;
  avatarUrl?: string;
  isAuthor?: boolean;
  items: PRTodoCandidate[];
}

// buildSections orders the left list (tests, lint, checks, then comments) and
// splits the comments group into one section per author, preserving first-seen
// order so the list is stable as detail streams in.
function buildSections(candidates: PRTodoCandidate[]): Section[] {
  const out: Section[] = [];
  for (const g of GROUPS) {
    const items = candidates.filter(c => c.group === g.id);
    if (items.length === 0) continue;
    if (g.id !== 'comments') {
      out.push({ id: g.id, label: g.label, icon: g.icon, items });
      continue;
    }
    const byAuthor = new Map<string, PRTodoCandidate[]>();
    for (const c of items) {
      const author = c.author || 'unknown';
      const list = byAuthor.get(author);
      if (list) list.push(c);
      else byAuthor.set(author, [c]);
    }
    for (const [author, authorItems] of byAuthor) {
      out.push({
        id: `comments:${author}`,
        label: `@${author}`,
        icon: g.icon,
        avatarUrl: authorItems[0]?.avatarUrl,
        isAuthor: true,
        items: authorItems,
      });
    }
  }
  return out;
}

// DetailPane renders the focused candidate's full content in the right column:
// a markdown body for comments, an ANSI-rendered message/log for tests and
// checks, and a link back to the source on GitHub when available.
function DetailPane({ detail }: { detail?: PRTodoCandidateDetail }) {
  if (!detail) {
    return (
      <div className="flex h-full items-center justify-center p-4 text-center text-xs text-muted-foreground">
        Select a test, lint violation, check, or comment to see its details here.
      </div>
    );
  }
  return (
    <div className="space-y-2 p-3">
      <div>
        {detail.headingMarkdown ? (
          <Markdown inline text={detail.heading} className="break-words text-sm font-semibold text-foreground" />
        ) : (
          <div className="break-words text-sm font-semibold text-foreground">{detail.heading}</div>
        )}
        {(detail.meta || detail.location) && (
          <div className="mt-0.5 break-all font-mono text-[11px] text-muted-foreground">
            {[detail.meta, detail.location].filter(Boolean).join(' · ')}
          </div>
        )}
      </div>
      {detail.message && (
        <div
          className="whitespace-pre-wrap break-words text-xs text-foreground"
          dangerouslySetInnerHTML={{ __html: ansiToHtml(detail.message) }}
        />
      )}
      {detail.markdown && <Markdown text={detail.markdown} className="text-xs text-foreground" />}
      {detail.log && (
        <pre
          className="overflow-x-auto whitespace-pre-wrap rounded border border-border bg-black px-3 py-2 font-mono text-[11px] text-gray-100"
          dangerouslySetInnerHTML={{ __html: ansiToHtml(detail.log) }}
        />
      )}
      {detail.url && (
        <a href={detail.url} target="_blank" rel="noopener" className="inline-flex items-center gap-1 text-xs text-blue-600 hover:underline">
          <UiLinkExternal />
          View on GitHub
        </a>
      )}
    </div>
  );
}

// SignalPanes lays the signal list beside its focused detail, or on a phone
// pages between them: tapping a signal opens its detail, Back returns to the list.
function SignalPanes({ mobile, showDetail, onBack, list, detail }: {
  mobile: boolean;
  showDetail: boolean;
  onBack: () => void;
  list: ReactNode;
  detail: ReactNode;
}) {
  if (!mobile) return <SplitPane defaultSplit={42} minLeft={28} minRight={30} left={list} right={detail} />;
  if (!showDetail) return <div className="h-full overflow-y-auto">{list}</div>;
  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="shrink-0 border-b border-border px-1 py-0.5">
        <Button type="button" variant="ghost" aria-label="Back to signals" onClick={onBack} className="min-h-11 gap-1 px-2 text-xs">
          <UiArrowLeft />Signals
        </Button>
      </div>
      <div className="min-h-0 flex-1">{detail}</div>
    </div>
  );
}

// CreateTodoFromPRDialog turns a PR's failing tests, lint violations, failed CI
// checks, and review comments into a new todo. Tests/lint contribute body
// details and acceptance criteria; selected comments add body context, while
// checks/comments remain verification gates.
export function CreateTodoFromPRDialog({
  pr,
  detail,
  workspaces,
  onProjectsChanged,
  hosted,
  recovered,
  onResult,
}: {
  pr: PRItem;
  detail: PRDetail | null;
  workspaces: Project[];
  // onProjectsChanged lets the host refresh the project catalog after one is registered.
  onProjectsChanged?: () => void;
  hosted: HostedModalProps;
  recovered?: PRTodoDraft;
  onResult: (result: CreateTodoResponse, run: RunAfterCreationState['submitted']) => void;
}) {
  const { modalProps, submit: submitHosted } = hosted;
  const { open, onClose } = modalProps;
  const candidates = useMemo(() => recovered?.candidates ?? buildPRTodoCandidates(pr, detail), [pr, detail, recovered]);
  // Prefer the workspace whose repos include this PR's repo, else the first.
  const defaultDir = useMemo(() => {
    const match = workspaces.find(w => (w.repos ?? []).includes(pr.repo));
    return match?.dir ?? workspaces[0]?.dir ?? '';
  }, [workspaces, pr.repo]);

  // The ListMenu multi-select owns which signals become criteria or verification gates;
  // activeKey is the separate detail-focus row shown in the right pane. Selection
  // is controlled (we own selectedKeys) so checkbox writes and the submit read
  // share one source of truth, and so programmatic defaults (which call
  // setSelectedKeys directly) are distinguishable from user edits (which route
  // through onSelectionChange and flip `touched`).
  const keys = useMemo(() => candidates.map(c => c.key), [candidates]);
  const sections = useMemo(() => buildSections(candidates), [candidates]);
  // Default pre-selection is the CI failures (tests/lint/checks); review comments
  // start unchecked so a long nitpick list isn't dumped in.
  const defaultKeys = useMemo(() => candidates.filter(c => c.group !== 'comments').map(c => c.key), [candidates]);

  const [selectedKeys, setSelectedKeys] = useState<string[]>(recovered?.selectedKeys ?? []);
  const touched = useRef(!!recovered);
  const selection = useListMenuSelection({
    keys,
    selectedKeys,
    onSelectionChange: next => {
      touched.current = true;
      setSelectedKeys(next);
    },
  });

  const [dir, setDir] = useState(recovered?.dir ?? defaultDir);
  const [title, setTitle] = useState(recovered?.title ?? '');
  const [priority, setPriority] = useState<TodoPriority>(recovered?.priority ?? 'medium');
  const [status, setStatus] = useState<TodoStatus>(recovered?.status ?? 'pending');
  const [notes, setNotes] = useState(recovered?.notes ?? '');
  const [activeKey, setActiveKey] = useState<string | null>(null);
  // A phone shows the signal list or one signal's detail, not both side by side.
  const mobile = useIsMobile();
  const [detailOpen, setDetailOpen] = useState(false);
  const [error, setError] = useState('');
  const runAfter = useRunAfterCreation({ dir, status, enabled: open, initialStep: recovered?.step });
  const createTodo = useCreateTodoMutation(dir);
  const busy = createTodo.isPending;
  const projectWorkspace = usePRTodoProjectWorkspace({
    open, prRepo: pr.repo, workspaces, onDirChange: setDir, onProjectsChanged,
  });

  // Reset the form on the closed→open transition, then keep re-applying the
  // default selection as the PR detail streams in (comments first, test/lint
  // results later) until the user edits the selection. Without the re-apply,
  // opening the dialog before gavelResults arrive would leave the failing tests
  // unchecked, so "Add todo" created a todo with no acceptance criteria. Once the
  // user touches the selection (`touched`), streaming stops overwriting it.
  const wasOpen = useRef(false);
  useEffect(() => {
    if (!open) {
      wasOpen.current = false;
      return;
    }
    if (!wasOpen.current) {
      wasOpen.current = true;
      touched.current = !!recovered;
      setDir(recovered?.dir ?? defaultDir);
      setTitle(recovered?.title ?? `Address feedback on ${pr.repo}#${pr.number}`);
      setPriority(recovered?.priority ?? 'medium');
      setStatus(recovered?.status ?? 'pending');
      setNotes(recovered?.notes ?? '');
      setError('');
      createTodo.reset();
    }
    if (!touched.current) {
      setSelectedKeys(defaultKeys);
      setActiveKey(prev =>
        prev && candidates.some(c => c.key === prev) ? prev : (defaultKeys[0] ?? candidates[0]?.key ?? null),
      );
    }
  }, [open, defaultDir, pr.repo, pr.number, candidates, defaultKeys, createTodo.reset, recovered]);

  if (!open) return null;

  function applyPreset(g: (typeof GROUPS)[number]) {
    touched.current = true;
    setSelectedKeys(candidates.filter(c => c.group === g.id).map(c => c.key));
    setActiveKey(candidates.find(c => c.group === g.id)?.key ?? null);
    setTitle(`${g.preset} in ${pr.repo}#${pr.number}`);
  }

  // toggleKeys flips a section's All/None: deselect them all when every key is
  // already selected, otherwise add the missing ones.
  function toggleKeys(sectionKeys: string[]) {
    touched.current = true;
    const allOn = sectionKeys.length > 0 && sectionKeys.every(k => selectedKeys.includes(k));
    setSelectedKeys(
      allOn
        ? selectedKeys.filter(k => !sectionKeys.includes(k))
        : Array.from(new Set([...selectedKeys, ...sectionKeys])),
    );
  }

  async function submit() {
    if (!title.trim() || !dir || busy || runAfter.blocked) return;
    setError('');
    try {
      const sel = new Set(selectedKeys);
      const selected = candidates.filter(c => sel.has(c.key));
      const criteria: AcceptanceCriterion[] = selected
        .filter(isPRTodoCriterion)
        .map(c => ({ text: c.text }));
      const prVerification = buildPRTodoVerification(pr, candidates, selected);
      const payload: {
        title: string;
        body: string;
        priority: TodoPriority;
        status: TodoStatus;
        criteria: AcceptanceCriterion[];
        prVerification?: PRTodoVerificationPayload;
        pr: { number: number; url: string };
      } = { title: title.trim(), body: buildPRTodoBody(pr, notes, selected), priority, status, criteria, pr: { number: pr.number, url: pr.url } };
      if (prVerification) payload.prVerification = prVerification;
      const runFields = runAfter.capture();
      const mutation = {
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...payload, ...runFields }),
      };
      await submitHosted({
        draft: savePRTodoDraft(pr, { dir, title, priority, status, notes, selectedKeys: selected.map(candidate => candidate.key), candidates, step: runAfter.step }),
        request: async () => {
          try {
            return { status: 'success', value: await createTodo.mutateAsync(mutation) };
          } catch (cause) {
            if (cause instanceof TodoMutationError && cause.status < 500) return { status: 'rejected', error: cause.message };
            return { status: 'unknown', error: cause instanceof Error ? cause.message : 'Could not create todo' };
          }
        },
        onSuccess: data => onResult(data, runFields.step ? { dir, ...runFields, step: runFields.step } : null),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    }
  }

  const selectedCandidates = candidates.filter(c => selectedKeys.includes(c.key));
  const selectedCriteriaCount = selectedCandidates.filter(isPRTodoCriterion).length;
  const selectedGateCount = selectedCandidates.filter(c => c.actionPattern || c.commentId).length;
  const presets = GROUPS.filter(g => candidates.some(c => c.group === g.id));
  const activeDetail = candidates.find(c => c.key === activeKey)?.detail;

  return (
    <>
      <Modal
        {...modalProps}
        error={error || modalProps.error}
        size="xl"
        className="max-w-6xl"
        footer={
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="min-w-0 text-xs text-muted-foreground">
              {selectedCriteriaCount} acceptance criteri{selectedCriteriaCount === 1 ? 'on' : 'a'}
              {selectedGateCount > 0 && ` · ${selectedGateCount} verification gate${selectedGateCount === 1 ? '' : 's'}`}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" onClick={onClose}>Cancel</Button>
              <Button onClick={submit} loading={busy} disabled={!title.trim() || !dir || runAfter.blocked}>Add todo</Button>
            </div>
          </div>
        }
      >
        <div className="space-y-3">
          {presets.length > 0 && (
            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Quick todo:</span>
              {presets.map(g => {
                const Icon = g.icon;
                return (
                  <Button
                    key={g.id}
                    variant="outline"
                    size="sm"
                    className="h-7 gap-1 px-2 text-xs"
                    onClick={() => applyPreset(g)}
                  >
                    <Icon className="text-xs" />
                    {g.preset}
                  </Button>
                );
              })}
            </div>
          )}

          <PRTodoWorkspaceField
            choices={projectWorkspace.choices} value={dir} busy={busy}
            onChange={setDir} onNewProject={projectWorkspace.openAddProject}
          />

          <Field label="Title">
            <input
              className={inputClass}
              value={title}
              placeholder="What needs doing?"
              aria-label="Title"
              onChange={e => setTitle(e.currentTarget.value)}
              autoFocus
            />
          </Field>

          <div className="flex gap-3">
            <div className="flex-1">
              <Field label="Priority">
                <Select value={priority} onChange={e => setPriority(e.currentTarget.value as TodoPriority)} className={inputClass} aria-label="Priority">
                  {priorities.map(p => <option key={p} value={p}>{p}</option>)}
                </Select>
              </Field>
            </div>
            <div className="flex-1">
              <Field label="Status">
                <Select value={status} onChange={e => setStatus(e.currentTarget.value as TodoStatus)} className={inputClass} aria-label="Status">
                  {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
                </Select>
              </Field>
            </div>
          </div>

          <RunAfterCreation state={runAfter} disabled={busy} />
          <div>
            <div className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">PR signals</div>
            {candidates.length === 0 ? (
              <p className="rounded-md border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
                No failing tests, lint violations, checks, or open comments on this PR — the todo will be created from the title and notes alone.
              </p>
            ) : (
              <div className="h-[55vh] overflow-hidden rounded-md border border-border">
                <SignalPanes
                  mobile={mobile}
                  showDetail={detailOpen && activeDetail !== undefined}
                  onBack={() => setDetailOpen(false)}
                  list={
                    <ListMenu selection={selection}>
                      {sections.map(sec => {
                        const allOn = sec.items.every(c => selection.isSelected(c.key));
                        const Icon = sec.icon;
                        return (
                          <div key={sec.id}>
                            <ListMenuHeader>
                              {sec.isAuthor && sec.avatarUrl ? (
                                <Avatar src={sec.avatarUrl} alt={sec.label} size={16} />
                              ) : (
                                <Icon className="text-sm text-muted-foreground" />
                              )}
                              <span className={`flex-1 truncate text-xs font-semibold tracking-wide text-muted-foreground ${sec.isAuthor ? '' : 'uppercase'}`}>{sec.label}</span>
                              <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">{sec.items.length}</span>
                              <Button
                                type="button"
                                size="sm"
                                variant="ghost"
                                className="h-6 px-1.5 text-[11px]"
                                onClick={() => toggleKeys(sec.items.map(c => c.key))}
                              >
                                {allOn ? 'None' : 'All'}
                              </Button>
                            </ListMenuHeader>
                            {sec.items.map(c => (
                              <ListMenuItem
                                key={c.key}
                                itemKey={c.key}
                                active={activeKey === c.key}
                                className="px-2.5 py-1.5"
                                onClick={() => { setActiveKey(c.key); setDetailOpen(true); }}
                              >
                                {c.group === 'comments' ? (
                                  <Markdown inline text={c.primary} className="block truncate text-sm text-foreground" />
                                ) : (
                                  <span className="block truncate text-sm text-foreground">{c.primary}</span>
                                )}
                                {c.secondary && <span className="block truncate font-mono text-[11px] text-muted-foreground">{c.secondary}</span>}
                              </ListMenuItem>
                            ))}
                          </div>
                        );
                      })}
                    </ListMenu>
                  }
                  detail={
                    <div className="h-full overflow-y-auto bg-muted/20">
                      <DetailPane detail={activeDetail} />
                    </div>
                  }
                />
              </div>
            )}
          </div>

          <Field label="Notes">
            <textarea
              className={`${inputClass} h-20 resize-none`}
              value={notes}
              placeholder="Extra context (optional)"
              aria-label="Notes"
              onChange={e => setNotes(e.currentTarget.value)}
            />
          </Field>
        </div>
      </Modal>
      <AddProjectDialog
        open={projectWorkspace.addProjectOpen && !modalProps.minimized}
        onClose={projectWorkspace.closeAddProject}
        onSaved={projectWorkspace.projectSaved}
        repoOptions={projectWorkspace.repoOptions}
        defaults={projectWorkspace.defaults}
      />
    </>
  );
}
