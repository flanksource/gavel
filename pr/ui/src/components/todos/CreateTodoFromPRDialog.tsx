import { useEffect, useMemo, useRef, useState } from 'react';
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
} from '@flanksource/clicky-ui/components';
import type { AcceptanceCriterion, PRDetail, PRItem, Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { extractCommentTitle, isDeploymentComment } from '../../utils';
import { ansiToHtml } from '../../ansi';
import { Markdown } from '../Markdown';
import { Avatar } from '../Avatar';
import { GavelIcon } from '../GavelIcon';
import { inputClass, priorities, statuses, statusLabel, todoQuery } from './format';

// SourceGroup is the kind of PR signal a candidate criterion came from.
type SourceGroup = 'tests' | 'lint' | 'checks' | 'comments';

// CandidateDetail is the full content shown in the right-hand pane when a row is
// focused: a heading + optional location/meta, a short message, a markdown body
// (comments), and a preformatted log (test stack traces, check logs).
interface CandidateDetail {
  heading: string;
  // headingMarkdown renders the heading as inline markdown (used for comment
  // titles, which may carry bold/code/link formatting).
  headingMarkdown?: boolean;
  location?: string;
  meta?: string;
  message?: string;
  markdown?: string;
  log?: string;
  url?: string;
}

// Candidate is one selectable acceptance criterion derived from a PR signal. The
// key is stable per signal so selection survives re-renders; text is the
// criterion stored on the todo; primary/secondary label the row; detail backs
// the right-hand pane.
interface Candidate {
  key: string;
  group: SourceGroup;
  text: string;
  primary: string;
  secondary?: string;
  // author/avatarUrl are set for review comments so they can be sub-grouped by
  // author in the list.
  author?: string;
  avatarUrl?: string;
  detail: CandidateDetail;
}

// GROUPS drives both the grouped checkbox lists and the one-click presets: each
// preset titles the todo and narrows the selection to that group ("Fix failing
// tests in repo#7", etc.).
const GROUPS: { id: SourceGroup; label: string; icon: string; preset: string }[] = [
  { id: 'tests', label: 'Failing tests', icon: 'codicon:beaker-stop', preset: 'Fix failing tests' },
  { id: 'lint', label: 'Lint violations', icon: 'codicon:warning', preset: 'Fix lint violations' },
  { id: 'checks', label: 'Failing checks', icon: 'codicon:error', preset: 'Fix failing checks' },
  { id: 'comments', label: 'Review comments', icon: 'codicon:comment-discussion', preset: 'Resolve review comments' },
];

function location(file?: string, line?: number): string {
  if (!file) return '';
  return line ? `${file}:${line}` : file;
}

// buildCandidates turns a PR's failing tests, lint violations, failed CI checks,
// and unresolved review comments into selectable acceptance criteria.
function buildCandidates(pr: PRItem, detail: PRDetail | null): Candidate[] {
  const out: Candidate[] = [];
  const shards = detail?.gavelResults ?? [];

  shards.forEach((s, si) => {
    (s.topFailures ?? []).forEach((f, i) => {
      const name = f.suite ? `${f.suite} › ${f.name}` : f.name;
      const loc = location(f.file, f.line);
      out.push({
        key: `test:${si}:${i}:${f.name}`,
        group: 'tests',
        text: `Test \`${name}\` passes`,
        primary: f.name,
        secondary: loc || f.suite,
        detail: { heading: name, location: loc, message: f.message, log: f.details },
      });
    });
  });

  shards.forEach((s, si) => {
    (s.topLintViolations ?? []).forEach((v, i) => {
      const loc = location(v.file, v.line);
      const rule = v.rule ? ` (${v.rule})` : '';
      out.push({
        key: `lint:${si}:${i}:${v.file ?? ''}:${v.line ?? ''}:${v.rule ?? ''}`,
        group: 'lint',
        text: `Resolve ${v.linter}${rule} violation${loc ? ` at ${loc}` : ''}`,
        primary: `${v.linter}${rule}`,
        secondary: loc || v.message,
        detail: { heading: `${v.linter}${rule}`, location: loc, message: v.message },
      });
    });
  });

  (pr.checkStatus?.failures ?? []).forEach((c, i) => {
    const steps = (c.failedSteps ?? []).join(', ');
    out.push({
      key: `check:${i}:${c.name}`,
      group: 'checks',
      text: `CI check "${c.name}" passes`,
      primary: c.name,
      secondary: steps || undefined,
      detail: { heading: c.name, meta: steps ? `Failed steps: ${steps}` : undefined, log: c.logTail, url: c.detailsUrl },
    });
  });

  (detail?.comments ?? [])
    .filter(c => !isDeploymentComment(c) && !c.isResolved && !c.isOutdated)
    .forEach((c, i) => {
      const title = extractCommentTitle(c.body);
      const loc = location(c.path, c.line);
      out.push({
        key: `comment:${c.id}:${i}`,
        group: 'comments',
        text: `Address @${c.author}'s comment${loc ? ` on ${loc}` : ''}: ${title}`,
        primary: title || `Comment by @${c.author}`,
        secondary: loc || undefined,
        author: c.author,
        avatarUrl: c.avatarUrl,
        detail: { heading: title || `Comment by @${c.author}`, headingMarkdown: true, meta: `@${c.author}`, location: loc, markdown: c.body, url: c.url },
      });
    });

  return out;
}

// Section is one rendered group in the left list: tests/lint/checks each map to a
// single section; review comments are sub-grouped into one section per author.
interface Section {
  id: string;
  label: string;
  icon: string;
  avatarUrl?: string;
  isAuthor?: boolean;
  items: Candidate[];
}

// buildSections orders the left list (tests, lint, checks, then comments) and
// splits the comments group into one section per author, preserving first-seen
// order so the list is stable as detail streams in.
function buildSections(candidates: Candidate[]): Section[] {
  const out: Section[] = [];
  for (const g of GROUPS) {
    const items = candidates.filter(c => c.group === g.id);
    if (items.length === 0) continue;
    if (g.id !== 'comments') {
      out.push({ id: g.id, label: g.label, icon: g.icon, items });
      continue;
    }
    const byAuthor = new Map<string, Candidate[]>();
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

// todoHref builds the /todos/{ref} deep link the same way routes.buildRoute does,
// encoding each path segment so a file-backed ref (a path) round-trips.
function todoHref(todo: TodoItem): string {
  return `/todos/${todo.ref.split('/').map(encodeURIComponent).join('/')}`;
}

// DetailPane renders the focused candidate's full content in the right column:
// a markdown body for comments, an ANSI-rendered message/log for tests and
// checks, and a link back to the source on GitHub when available.
function DetailPane({ detail }: { detail?: CandidateDetail }) {
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
          <GavelIcon name="codicon:link-external" />
          View on GitHub
        </a>
      )}
    </div>
  );
}

// CreateTodoFromPRDialog turns a PR's failing tests, lint violations, failed CI
// checks, and review comments into a new todo whose acceptance criteria are the
// selected signals. Selecting a row shows its full details on the right. It stays
// on the PR after creating and links to the new todo.
export function CreateTodoFromPRDialog({
  open,
  onClose,
  pr,
  detail,
  workspaces,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  pr: PRItem;
  detail: PRDetail | null;
  workspaces: Project[];
  // onCreated lets the host refresh todo counts after a todo is added.
  onCreated?: () => void;
}) {
  const candidates = useMemo(() => buildCandidates(pr, detail), [pr, detail]);
  // Prefer the workspace whose repos include this PR's repo, else the first.
  const defaultDir = useMemo(() => {
    const match = workspaces.find(w => (w.repos ?? []).includes(pr.repo));
    return match?.dir ?? workspaces[0]?.dir ?? '';
  }, [workspaces, pr.repo]);

  // The ListMenu multi-select owns which signals become acceptance criteria;
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

  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const touched = useRef(false);
  const selection = useListMenuSelection({
    keys,
    selectedKeys,
    onSelectionChange: next => {
      touched.current = true;
      setSelectedKeys(next);
    },
  });

  const [dir, setDir] = useState(defaultDir);
  const [title, setTitle] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [notes, setNotes] = useState('');
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [created, setCreated] = useState<TodoItem | null>(null);

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
      touched.current = false;
      setDir(defaultDir);
      setTitle(`Address feedback on ${pr.repo}#${pr.number}`);
      setPriority('medium');
      setStatus('pending');
      setNotes('');
      setError('');
      setBusy(false);
      setCreated(null);
    }
    if (!touched.current) {
      setSelectedKeys(defaultKeys);
      setActiveKey(prev =>
        prev && candidates.some(c => c.key === prev) ? prev : (defaultKeys[0] ?? candidates[0]?.key ?? null),
      );
    }
  }, [open, defaultDir, pr.repo, pr.number, candidates, defaultKeys]);

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
    if (!title.trim() || !dir || busy) return;
    setBusy(true);
    setError('');
    try {
      const provider = workspaces.find(w => w.dir === dir)?.todoProvider || 'auto';
      const sel = new Set(selectedKeys);
      const criteria: AcceptanceCriterion[] = candidates
        .filter(c => sel.has(c.key))
        .map(c => ({ text: c.text }));
      const bodyParts: string[] = [];
      if (notes.trim()) bodyParts.push(notes.trim());
      bodyParts.push(`_From [${pr.repo}#${pr.number}](${pr.url})._`);
      const res = await fetch(`/api/todos/new?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: title.trim(), body: bodyParts.join('\n\n'), priority, status, criteria }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Create failed');
      setCreated(data.todo as TodoItem);
      onCreated?.();
    } catch (err: any) {
      setError(err?.message || 'Create failed');
    } finally {
      setBusy(false);
    }
  }

  const selectedCount = selectedKeys.length;
  const presets = GROUPS.filter(g => candidates.some(c => c.group === g.id));
  const activeDetail = candidates.find(c => c.key === activeKey)?.detail;

  return (
    <Modal
      open
      onClose={onClose}
      title={created ? 'Todo created' : 'New todo from PR'}
      size="xl"
      className="max-w-6xl"
      footer={
        created ? (
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setCreated(null)}>Create another</Button>
            <Button onClick={onClose}>Done</Button>
          </div>
        ) : (
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs text-muted-foreground">
              {selectedCount} acceptance criteri{selectedCount === 1 ? 'on' : 'a'}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" onClick={onClose}>Cancel</Button>
              <Button onClick={submit} loading={busy} disabled={!title.trim() || !dir}>Add todo</Button>
            </div>
          </div>
        )
      }
    >
      {created ? (
        <div className="space-y-3">
          <div className="flex items-start gap-2 rounded-md border border-green-200 bg-green-50 p-3 text-sm text-green-800">
            <GavelIcon name="codicon:pass" className="mt-0.5 shrink-0 text-green-600" />
            <div className="min-w-0">
              <div className="font-medium">{created.title}</div>
              <div className="mt-0.5 text-xs text-green-700">
                {(created.criteria?.length ?? 0)} acceptance criteri{(created.criteria?.length ?? 0) === 1 ? 'on' : 'a'} added
              </div>
            </div>
          </div>
          <a
            href={todoHref(created)}
            className="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800 hover:underline"
          >
            <GavelIcon name="codicon:link-external" />
            Open todo
          </a>
        </div>
      ) : (
        <div className="space-y-3">
          {error && <div className="text-sm text-destructive">{error}</div>}

          {presets.length > 0 && (
            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Quick todo:</span>
              {presets.map(g => (
                <Button
                  key={g.id}
                  variant="outline"
                  size="sm"
                  className="h-7 gap-1 px-2 text-xs"
                  onClick={() => applyPreset(g)}
                >
                  <GavelIcon name={g.icon} className="text-xs" />
                  {g.preset}
                </Button>
              ))}
            </div>
          )}

          {workspaces.length > 1 && (
            <Field label="Workspace">
              <Select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace">
                {workspaces.map(w => <option key={w.dir} value={w.dir}>{w.name}</option>)}
              </Select>
            </Field>
          )}

          <Field label="Title">
            <input
              className={inputClass}
              value={title}
              placeholder="What needs doing?"
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

          <div>
            <div className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">Acceptance criteria</div>
            {candidates.length === 0 ? (
              <p className="rounded-md border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
                No failing tests, lint violations, checks, or open comments on this PR — the todo will be created from the title and notes alone.
              </p>
            ) : (
              <div className="h-[55vh] overflow-hidden rounded-md border border-border">
                <SplitPane
                  defaultSplit={42}
                  minLeft={28}
                  minRight={30}
                  left={
                    <ListMenu selection={selection}>
                      {sections.map(sec => {
                        const allOn = sec.items.every(c => selection.isSelected(c.key));
                        return (
                          <div key={sec.id}>
                            <ListMenuHeader>
                              {sec.isAuthor && sec.avatarUrl ? (
                                <Avatar src={sec.avatarUrl} alt={sec.label} size={16} />
                              ) : (
                                <GavelIcon name={sec.icon} className="text-sm text-muted-foreground" />
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
                                onClick={() => setActiveKey(c.key)}
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
                  right={
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
              onChange={e => setNotes(e.currentTarget.value)}
            />
          </Field>
        </div>
      )}
    </Modal>
  );
}
