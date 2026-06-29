import { useEffect, useMemo, useRef, useState } from 'react';
import { Modal, Field, Button, Select } from '@flanksource/clicky-ui/components';
import type { AcceptanceCriterion, PRDetail, PRItem, Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { extractCommentTitle, isDeploymentComment } from '../../utils';
import { GavelIcon } from '../GavelIcon';
import { inputClass, priorities, statuses, statusLabel, todoQuery } from './format';

// SourceGroup is the kind of PR signal a candidate criterion came from.
type SourceGroup = 'tests' | 'lint' | 'checks' | 'comments';

// Candidate is one selectable acceptance criterion derived from a PR signal. The
// key is stable per signal so selection survives re-renders; text is the
// criterion stored on the todo; primary/secondary label the row.
interface Candidate {
  key: string;
  group: SourceGroup;
  text: string;
  primary: string;
  secondary?: string;
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
      out.push({
        key: `test:${si}:${i}:${f.name}`,
        group: 'tests',
        text: `Test \`${name}\` passes`,
        primary: f.name,
        secondary: location(f.file, f.line) || f.suite,
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
      });
    });
  });

  (pr.checkStatus?.failures ?? []).forEach((c, i) => {
    out.push({
      key: `check:${i}:${c.name}`,
      group: 'checks',
      text: `CI check "${c.name}" passes`,
      primary: c.name,
      secondary: (c.failedSteps ?? []).join(', ') || undefined,
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
        secondary: `@${c.author}${loc ? ` · ${loc}` : ''}`,
      });
    });

  return out;
}

// todoHref builds the /todos/{ref} deep link the same way routes.buildRoute does,
// encoding each path segment so a file-backed ref (a path) round-trips.
function todoHref(todo: TodoItem): string {
  return `/todos/${todo.ref.split('/').map(encodeURIComponent).join('/')}`;
}

// CreateTodoFromPRDialog turns a PR's failing tests, lint violations, failed CI
// checks, and review comments into a new todo whose acceptance criteria are the
// selected signals. It stays on the PR after creating and links to the new todo.
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

  const [dir, setDir] = useState(defaultDir);
  const [title, setTitle] = useState('');
  const [priority, setPriority] = useState<TodoPriority>('medium');
  const [status, setStatus] = useState<TodoStatus>('pending');
  const [notes, setNotes] = useState('');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [created, setCreated] = useState<TodoItem | null>(null);

  // Reset the form on the closed→open transition only — not on later renders
  // when the PR detail streams in (comments first, test/lint results later), so
  // an in-flight update never wipes a title the user has typed or items they
  // toggled. The CI failures (tests, lint, checks) are pre-selected since "fix
  // the failures" is the common intent; review comments start unselected so a
  // long nitpick list isn't dumped in.
  const wasOpen = useRef(false);
  useEffect(() => {
    if (open && !wasOpen.current) {
      setDir(defaultDir);
      setTitle(`Address feedback on ${pr.repo}#${pr.number}`);
      setPriority('medium');
      setStatus('pending');
      setNotes('');
      setSelected(new Set(candidates.filter(c => c.group !== 'comments').map(c => c.key)));
      setError('');
      setBusy(false);
      setCreated(null);
    }
    wasOpen.current = open;
  }, [open, defaultDir, pr.repo, pr.number, candidates]);

  if (!open) return null;

  function toggle(key: string) {
    setSelected(prev => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }

  function applyPreset(g: (typeof GROUPS)[number]) {
    setSelected(new Set(candidates.filter(c => c.group === g.id).map(c => c.key)));
    setTitle(`${g.preset} in ${pr.repo}#${pr.number}`);
  }

  function toggleGroup(id: SourceGroup) {
    const keys = candidates.filter(c => c.group === id).map(c => c.key);
    const allOn = keys.length > 0 && keys.every(k => selected.has(k));
    setSelected(prev => {
      const next = new Set(prev);
      keys.forEach(k => (allOn ? next.delete(k) : next.add(k)));
      return next;
    });
  }

  async function submit() {
    if (!title.trim() || !dir || busy) return;
    setBusy(true);
    setError('');
    try {
      const provider = workspaces.find(w => w.dir === dir)?.todoProvider || 'auto';
      const criteria: AcceptanceCriterion[] = candidates
        .filter(c => selected.has(c.key))
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

  const selectedCount = selected.size;
  const presets = GROUPS.filter(g => candidates.some(c => c.group === g.id));

  return (
    <Modal
      open
      onClose={onClose}
      title={created ? 'Todo created' : 'New todo from PR'}
      size="lg"
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
              <div className="space-y-2">
                {GROUPS.map(g => {
                  const items = candidates.filter(c => c.group === g.id);
                  if (items.length === 0) return null;
                  const allOn = items.every(c => selected.has(c.key));
                  return (
                    <div key={g.id} className="overflow-hidden rounded-md border border-border">
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => toggleGroup(g.id)}
                        className="flex h-auto w-full items-center justify-start gap-2 rounded-none border-b border-border bg-muted/40 px-2.5 py-1.5 text-left text-xs font-semibold text-muted-foreground hover:bg-muted"
                      >
                        <GavelIcon name={allOn ? 'codicon:check' : 'codicon:checklist'} className="text-sm" />
                        <span className="flex-1">{g.label}</span>
                        <span className="rounded-full border border-border bg-background px-1.5 py-0.5 tabular-nums">{items.length}</span>
                      </Button>
                      <ul className="divide-y divide-border">
                        {items.map(c => (
                          <li key={c.key}>
                            <label className="flex cursor-pointer items-start gap-2 px-2.5 py-1.5 hover:bg-muted/40">
                              <input
                                type="checkbox"
                                checked={selected.has(c.key)}
                                onChange={() => toggle(c.key)}
                                className="mt-1 h-4 w-4 shrink-0 accent-primary"
                              />
                              <span className="min-w-0 flex-1">
                                <span className="block truncate text-sm text-foreground">{c.primary}</span>
                                {c.secondary && <span className="block truncate font-mono text-[11px] text-muted-foreground">{c.secondary}</span>}
                              </span>
                            </label>
                          </li>
                        ))}
                      </ul>
                    </div>
                  );
                })}
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
