import { useMemo, useState } from 'react';
import { Button, IconMenuPicker, Modal, type IconMenuOption } from '@flanksource/clicky-ui/components';
import { UiAdd, UiClose, UiTrash } from '@flanksource/clicky-ui/icons';

import type { TodoTagDef } from '../../../types';
import { TodoTag } from '../../todos/TodoTag';
import { TagPicker } from '../../todos/TagPicker';
import { TAG_PALETTE, tagClasses, tagHash, normalizeTag } from '../../todos/tagPalette';
import { useDeleteTagMutation, useSaveTagMutation, useTodoTagCounts } from '../../todos/tagQueries';
import { useTodoTagIndex } from '../../todos/tagQueries';
import { tagCandidates } from '../../todos/tagResolve';
import { boundTagGlyph, hasTagGlyph, TAG_GLYPH_OPTIONS } from '../../../icons/tags';

// TagDefinitionsPanel is the Tags settings tab: the editable taxonomy behind
// every tag chip in the app.
//
// It is not part of the schema-driven .gavel.yaml form because definitions live
// in the database, not in config — the same reason the Workspace tab is
// special-cased. Each row saves itself, so there is no page-level dirty state or
// Save bar to reconcile.
export function TagDefinitionsPanel({ dir }: { dir: string }) {
  const index = useTodoTagIndex(dir);
  const counts = useTodoTagCounts(dir);
  const save = useSaveTagMutation(dir);
  const remove = useDeleteTagMutation(dir);
  const [adding, setAdding] = useState(false);
  const [picking, setPicking] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  /** The tag awaiting removal confirmation, or null when nothing is pending. */
  const [pendingDrop, setPendingDrop] = useState<TodoTagDef | null>(null);

  // Common tags are the ones that resolve here but are not yet this project's
  // own: the built-in defaults, anything defined globally, and any label the
  // backlog already uses that nothing has defined. Picking one stores it for
  // this project, which is what makes it editable per project — the same
  // adoption the label backfill performs, for labels created since it ran.
  const common = useMemo(
    () => tagCandidates(index, counts).filter(def => def.scope !== 'workspace'),
    [index, counts],
  );

  // Tags in use first, then alphabetical — the same order the filter facet uses,
  // so the two surfaces agree on what matters in this workspace.
  const rows = useMemo(() => [...index.defs].sort((a, b) => {
    const used = (counts[b.name] ?? 0) - (counts[a.name] ?? 0);
    return used || a.name.localeCompare(b.name);
  }), [index.defs, counts]);

  async function persist(input: { name: string; color?: string; icon?: string; description?: string; global?: boolean }) {
    setError('');
    try {
      await save.mutateAsync(input);
      setAdding(false);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    }
  }

  // Removing a project tag also strips it from every todo carrying it, so it is
  // confirmed first, with the count. A global tag is presentation only and never
  // touches todo content, so it goes straight through.
  function requestDrop(def: TodoTagDef) {
    setError('');
    if (def.scope !== 'global' && (counts[def.name] ?? 0) > 0) {
      setPendingDrop(def);
      return;
    }
    void drop(def);
  }

  async function drop(def: TodoTagDef) {
    setError('');
    try {
      const removed = await remove.mutateAsync({ name: def.name, global: def.scope === 'global' });
      setPendingDrop(null);
      setNotice(removed?.removed?.todos
        ? `Removed "${def.name}" from ${removed.removed.todos} todo${removed.removed.todos === 1 ? '' : 's'}.`
        : '');
    } catch (cause) {
      setPendingDrop(null);
      setError(cause instanceof Error ? cause.message : String(cause));
    }
  }

  // Adopting a common tag is one click: it is stored for this project with the
  // colour and glyph it already renders with, so picking it never repaints it.
  async function adopt(name: string) {
    const def = common.find(candidate => candidate.name === name);
    setPicking(false);
    await persist({
      name,
      color: def?.color || tagHash(name),
      icon: def?.icon,
      description: def?.description,
    });
  }

  return (
    <section className="space-y-3">
      <header className="space-y-1">
        <h2 className="text-sm font-semibold text-foreground">Tags</h2>
        <p className="text-xs text-muted-foreground">
          How each label renders across the dashboard and the terminal. A tag with no definition
          still gets a stable colour derived from its name, so this is always an override —
          never the difference between coloured and colourless.
        </p>
      </header>

      {error && <div role="alert" className="text-sm text-destructive">{error}</div>}
      {notice && <div role="status" className="text-xs text-muted-foreground">{notice}</div>}

      <div className="divide-y divide-border rounded-md border border-border">
        {rows.map(def => (
          <TagRow
            key={`${def.scope}:${def.name}`}
            def={def}
            usage={counts[def.name] ?? 0}
            busy={save.isPending || remove.isPending}
            onSave={persist}
            onDelete={() => requestDrop(def)}
          />
        ))}
        {rows.length === 0 && (
          <p className="px-3 py-6 text-center text-xs text-muted-foreground">
            No tags defined yet.
          </p>
        )}
      </div>

      {adding ? (
        <TagRow
          def={{ name: '', color: 'slate', scope: 'workspace' }}
          usage={0}
          busy={save.isPending}
          draft
          onSave={persist}
          onDelete={() => setAdding(false)}
        />
      ) : (
        <div className="flex flex-wrap items-center gap-1">
          <span className="relative inline-flex">
            <Button
              size="sm"
              variant="ghost"
              disabled={common.length === 0}
              aria-expanded={picking}
              onClick={() => setPicking(open => !open)}
              className="gap-1"
            >
              <UiAdd className="text-xs" /> Add common tag
            </Button>
            {picking && (
              <TagPicker
                candidates={common}
                counts={counts}
                index={index}
                allowCreate={false}
                onPick={name => void adopt(name)}
                onClose={() => setPicking(false)}
                placeholder="Filter common tags…"
                emptyLabel="Every common tag is already defined here."
              />
            )}
          </span>
          <Button size="sm" variant="ghost" onClick={() => setAdding(true)} className="gap-1">
            <UiAdd className="text-xs" /> New tag
          </Button>
        </div>
      )}

      {pendingDrop && (
        <Modal
          open
          onClose={() => setPendingDrop(null)}
          title={`Remove "${pendingDrop.name}" from this project?`}
          size="sm"
          footer={
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setPendingDrop(null)} disabled={remove.isPending}>
                Cancel
              </Button>
              <Button variant="destructive" loading={remove.isPending} onClick={() => void drop(pendingDrop)}>
                Remove
              </Button>
            </div>
          }
        >
          <div className="space-y-3 text-sm text-foreground">
            <p>
              This deletes the tag definition and removes the tag from{' '}
              <strong>{counts[pendingDrop.name] ?? 0}</strong>{' '}
              {(counts[pendingDrop.name] ?? 0) === 1 ? 'todo' : 'todos'}. It cannot be undone.
            </p>
            <TodoTag tag={index.resolve(pendingDrop.name)} />
          </div>
        </Modal>
      )}
    </section>
  );
}

// TagRow edits one definition. A built-in row is editable too: saving it stores
// a real row that shadows the default, and deleting that row restores it.
function TagRow({ def, usage, busy, draft = false, onSave, onDelete }: {
  def: TodoTagDef;
  usage: number;
  busy: boolean;
  draft?: boolean;
  onSave: (input: { name: string; color?: string; icon?: string; description?: string; global?: boolean }) => void;
  onDelete: () => void;
}) {
  const [name, setName] = useState(def.name);
  const [color, setColor] = useState(def.color);
  const [icon, setIcon] = useState(def.icon ?? '');
  const [description, setDescription] = useState(def.description ?? '');
  const [global, setGlobal] = useState(def.scope === 'global');

  const normalized = normalizeTag(name);
  const iconOptions = useMemo<IconMenuOption<string>[]>(() => {
    if (!icon || hasTagGlyph(icon)) return TAG_GLYPH_OPTIONS;
    return [
      { value: icon, label: icon, icon: boundTagGlyph(icon, def.iconify) },
      ...TAG_GLYPH_OPTIONS,
    ];
  }, [def.iconify, icon]);
  // The preview is the authoring-time check that a colour and glyph actually
  // render — the reason an arbitrary Iconify name is safe to allow at all.
  const preview = {
    token: normalized || 'example',
    value: normalized || 'example',
    color: color || tagHash(normalized || 'example'),
    icon: icon || undefined,
    iconify: icon === def.icon ? def.iconify : undefined,
    description,
    defined: true,
  };

  const dirty = draft
    || color !== def.color
    || icon !== (def.icon ?? '')
    || description !== (def.description ?? '')
    || global !== (def.scope === 'global');

  return (
    <div className="space-y-2 px-3 py-2.5">
      <div className="flex flex-wrap items-center gap-2">
        {draft ? (
          <input
            value={name}
            onChange={event => setName(event.target.value)}
            placeholder="tag name"
            aria-label="Tag name"
            className="h-7 w-36 rounded border border-border bg-background px-2 text-xs text-foreground"
          />
        ) : (
          <TodoTag tag={preview} />
        )}

        <span className="text-[11px] text-muted-foreground">
          {def.scope === 'builtin' ? 'built-in default' : def.scope}
          {usage > 0 && ` · ${usage} todo${usage === 1 ? '' : 's'}`}
        </span>

        <div className="ml-auto flex items-center gap-1">
          {dirty && (
            <Button
              size="sm"
              disabled={busy || !normalized}
              onClick={() => onSave({ name: normalized, color, icon, description, global })}
            >
              Save
            </Button>
          )}
          <Button
            size="icon"
            variant="ghost"
            disabled={busy}
            title={draft ? 'Discard' : 'Remove this definition'}
            aria-label={draft ? 'Discard new tag' : `Remove ${def.name}`}
            onClick={onDelete}
            className="h-7 w-7 text-muted-foreground hover:text-destructive"
          >
            {draft ? <UiClose className="text-xs" /> : <UiTrash className="text-xs" />}
          </Button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-1">
        {TAG_PALETTE.map(hue => (
          <button
            key={hue}
            type="button"
            title={hue}
            aria-label={`Use ${hue}`}
            aria-pressed={color === hue}
            onClick={() => setColor(hue)}
            className={`h-5 w-5 rounded border ${tagClasses(hue)} ${color === hue ? 'ring-2 ring-ring' : 'border-transparent'}`}
          />
        ))}
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <IconMenuPicker
          value={icon}
          onChange={setIcon}
          options={iconOptions}
          ariaLabel={`Tag icon for ${normalized || 'new tag'}`}
          triggerClassName="h-7 w-7 text-muted-foreground"
          menuClassName="max-h-[min(24rem,calc(100vh-2rem))] overflow-y-auto"
        />
        <input
          value={description}
          onChange={event => setDescription(event.target.value)}
          placeholder="What this tag means"
          aria-label="Tag description"
          className="h-7 min-w-0 flex-1 rounded border border-border bg-background px-2 text-xs text-foreground"
        />
        <label className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
          <input
            type="checkbox"
            checked={global}
            onChange={event => setGlobal(event.target.checked)}
            className="h-3 w-3 accent-primary"
          />
          All workspaces
        </label>
      </div>
    </div>
  );
}
