import { useMemo } from 'react';
import type {
  DataTableSelectionAction,
  SelectionActionDisplay,
} from '@flanksource/clicky-ui/data';
import { useToast } from '@flanksource/clicky-ui/components';
import {
  UiCheck,
  UiClose,
  UiComment,
  UiListChecks,
  UiPlay,
  UiSeverityMedium,
  UiTag,
  UiTrash,
  type IconProps,
} from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import { priorityBadgeClass, priorityIcon, statusClass, statusIcon, statusLabel } from './format';
import {
  todoBulkActionLabel,
  todoBulkActionShortLabel,
  todoBulkResultMessage,
  useTodoBulkActionMutation,
  useTodoBulkActions,
  type TodoBulkAction,
  type TodoBulkResult,
} from './todoEntity';
import { selectionTargets, type SelectionState } from './todoSelection';
import { TodoLabelsMenu } from './TodoLabelsMenu';
import { buildTagIndex, todoVisibleLabels, type TagIndex } from './tagResolve';
import { normalizeTag } from './tagPalette';
import type { WorkspaceTodos } from './useWorkspaceTodos';
import type { TodoItem } from '../../types';

/**
 * Turns the server's action catalog into the descriptors a selection toolbar
 * renders.
 *
 * This module *maps* a registry; it does not define one. Every bulk action is
 * declared once in Go and published at `/api/entities` — so registering a new
 * one there makes it appear here, with its icon, its grouping, its confirmation
 * and its parameters, without touching this file. That is the whole point: the
 * dashboard used to hardcode three of the twenty-odd actions the CLI had, and
 * the two drifted.
 */

/** Icons the catalog names, resolved to components. Icon *names* render as
 *  broken glyphs, so the mapping is explicit and a name with no entry falls
 *  back rather than rendering a `?` box. */
const ACTION_ICONS: Record<string, ComponentType<IconProps>> = {
  'check-circle': UiCheck,
  flag: UiSeverityMedium,
  tag: UiTag,
  message: UiComment,
  trash: UiTrash,
  play: UiPlay,
};

function actionIcon(action: TodoBulkAction): ComponentType<IconProps> {
  return ACTION_ICONS[action.tool_hints?.icon ?? ''] ?? UiListChecks;
}

/**
 * Which actions get a named dropdown on the bar is derived from the catalog,
 * not from a list kept here: an action whose parameter is a closed set of
 * values — a status, a severity — is a field to edit, and a field belongs on
 * the bar as "Status ▾". Everything else is a command, and commands collapse.
 *
 * Deriving it is what keeps the bar honest. A bulk action added to
 * todos/entity/entity.go with an enum parameter appears as a dropdown without
 * anyone touching this file, which is the property the entity model exists for.
 */
function actionDisplay(
  action: TodoBulkAction,
  canEditSets: boolean,
): SelectionActionDisplay {
  if (enumParam(action)) return 'menu';
  // A set editor needs the label taxonomy to list. Without it the dropdown
  // would open on nothing, which is worse than one more click.
  if (arrayParams(action)) return canEditSets ? 'menu' : 'overflow';
  return 'overflow';
}

/**
 * Optional array parameters — a set of values toggled across the selection
 * rather than one value assigned to it. `labels` declares `add[]`/`remove[]`,
 * both optional, so `enumParam` (which wants exactly one *required* parameter)
 * correctly declines it and this picks it up instead.
 */
function arrayParams(action: TodoBulkAction): string[] | null {
  const schema = action.param_schema;
  if (!schema) return null;
  const names = Object.entries(schema.properties)
    .filter(([, field]) => field.type === 'array')
    .map(([name]) => name);
  // Every parameter must be one of them, or applying the menu would silently
  // omit whatever else the action needs.
  if (!names.length || names.length !== Object.keys(schema.properties).length) return null;
  return names;
}

/**
 * A parameter the action cannot run without and whose values are a closed set —
 * a status, a severity. Rendering it as a submenu of its own values is both
 * faster than a form and how the detail pane already presents the same choice.
 */
function enumParam(action: TodoBulkAction): { name: string; values: string[] } | null {
  const schema = action.param_schema;
  if (!schema) return null;
  const required = schema.required ?? [];
  if (required.length !== 1) return null;
  const [name] = required;
  const values = schema.properties[name]?.enum;
  if (!values?.length) return null;
  // Any other parameter must be optional, or a submenu would silently omit it.
  return { name, values };
}

/** Icon and badge colours for a value inside an enum submenu, so a status reads
 *  identically whether it is being set on one todo or forty. */
function valueBadge(action: TodoBulkAction, value: string) {
  if (action.name === 'status') {
    return { Icon: statusIcon(value as never), badgeClass: statusClass(value as never), label: statusLabel(value as never) };
  }
  if (action.name === 'priority') {
    return { Icon: priorityIcon(value as never), badgeClass: priorityBadgeClass(value as never), label: value };
  }
  return { Icon: actionIcon(action), badgeClass: 'border-border text-muted-foreground', label: value };
}

/**
 * What the label menu needs about the current selection, derived once for both
 * layouts: the todos the checked ids resolved to, the taxonomy to offer, and
 * how much this project uses each label.
 *
 * A selection routinely spans workspaces — grouping by severity or age flattens
 * them — so the taxonomy is the union of the definitions of every workspace the
 * selection touches, and the counts are taken over those workspaces' todos.
 * Reading one workspace's index would paint another's overrides onto its rows.
 */
export function useTodoBulkContext(todos: WorkspaceTodos): {
  todos: TodoItem[];
  tags: TagIndex;
  labelCounts: Record<string, number>;
} {
  const { selection, byDir, tagsByDir } = todos;

  return useMemo(() => {
    const targets = selectionTargets(selection.selection);
    const dirs = new Set(targets.map(target => target.dir));
    const wanted = new Set(targets.map(target => target.ref));

    const selected: TodoItem[] = [];
    const counts: Record<string, number> = {};
    for (const dir of dirs) {
      for (const todo of byDir[dir]?.items ?? []) {
        if (wanted.has(todo.ref)) selected.push(todo);
        for (const label of new Set(todoVisibleLabels(todo).map(normalizeTag))) {
          counts[label] = (counts[label] ?? 0) + 1;
        }
      }
    }

    const defs = [...dirs].flatMap(dir => tagsByDir?.get(dir)?.defs ?? []);
    const byName = new Map(defs.map(def => [def.name, def]));
    return { todos: selected, tags: buildTagIndex([...byName.values()]), labelCounts: counts };
  }, [selection.selection, byDir, tagsByDir]);
}

export interface TodoSelectionActionsOptions {
  selection: SelectionState;
  /**
   * The todos the selection resolved to, so a menu can show what the selection
   * already holds — which labels are on all of it, on some, on none.
   *
   * The bar's own `selectedRows` cannot serve: it carries only the rows a table
   * has paged in, and the sidebar layout has no rows to hand over at all.
   */
  todos?: TodoItem[];
  /** Per-label todo counts, for ordering the label menu by what this project uses. */
  labelCounts?: Record<string, number>;
  /** Label definitions, for the chips inside the label menu. */
  tags?: TagIndex;
  /** Reports each finished batch so the host can surface it. */
  onResult?: (result: TodoBulkResult, action: TodoBulkAction) => void;
  onError?: (error: Error, action: TodoBulkAction) => void;
}

/**
 * The selection-toolbar descriptors for the current checked set.
 *
 * Selection keys carry `dir` so a row can be identified across workspaces, but
 * the entity id is the bare `ref` — a dir would not survive a URL path segment,
 * and the server resolves a ref globally anyway.
 */
export function useTodoSelectionActions({
  selection,
  todos,
  labelCounts,
  tags,
  onResult,
  onError,
}: TodoSelectionActionsOptions): DataTableSelectionAction[] {
  const { data: catalog } = useTodoBulkActions();
  const runAction = useTodoBulkActionMutation();

  return useMemo(() => {
    if (!catalog?.length) return [];
    const refs = selectionTargets(selection).map(target => target.ref);

    const dispatch = (action: TodoBulkAction, params?: Record<string, string>) =>
      new Promise<void>(resolve => {
        runAction.mutate({ action, refs, params }, {
          onSuccess: result => { onResult?.(result, action); resolve(); },
          onError: error => { onError?.(error as Error, action); resolve(); },
        });
      });

    const descriptors = catalog.map<DataTableSelectionAction>(action => {
      const Icon = actionIcon(action);
      const destructive = action.tool_hints?.destructiveHint === true;
      const label = action.short ?? todoBulkActionLabel(action);
      const base: DataTableSelectionAction = {
        id: action.name,
        label: todoBulkActionShortLabel(action),
        // The icon is passed as a component, never as an element or a name —
        // a name renders as a broken glyph.
        icon: Icon,
        section: action.tool_hints?.group,
        description: label,
        display: actionDisplay(action, !!tags),
        variant: destructive ? 'destructive' : 'outline',
        disabled: refs.length === 0,
        onSelect: () => dispatch(action),
      };

      // A destructive action can act on a selection the caller never
      // enumerated, so it asks first — and it asks with the count, which is the
      // one number that makes the prompt worth stopping for.
      if (destructive) {
        base.confirm = {
          message: context =>
            `This permanently deletes ${context.selectedRowIds.length} todo${context.selectedRowIds.length === 1 ? '' : 's'}.`,
          confirmLabel: 'Delete',
        };
        // The server refuses without it; a UI confirmation is the caller saying
        // it out loud.
        base.onSelect = () => dispatch(action, { confirm: 'true' });
      }

      const choices = enumParam(action);
      if (choices) {
        base.children = choices.values.map(value => {
          const { Icon: ValueIcon, badgeClass, label: valueLabel } = valueBadge(action, value);
          return {
            id: `${action.name}:${value}`,
            label: (
              <span className="flex items-center gap-2">
                <span className={`inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full border ${badgeClass}`}>
                  <ValueIcon className="text-[10px]" />
                </span>
                <span className="capitalize">{valueLabel}</span>
              </span>
            ),
            disabled: refs.length === 0,
            onSelect: () => dispatch(action, { [choices.name]: value }),
          };
        });
        // The parent only opens the submenu; picking a value is the write.
        base.onSelect = () => {};
      }

      // A set toggled across the selection rather than one value assigned to
      // it. The menu is tri-state because the selection can disagree, and it
      // applies once — a write per click would fire a bulk call per keystroke
      // of intent and leave the list fighting its own refetches.
      const arrays = arrayParams(action);
      if (arrays && tags) {
        const [addParam, removeParam] = arrays;
        base.menu = ({ close, busy, run }) => (
          <TodoLabelsMenu
            index={tags}
            counts={labelCounts ?? {}}
            selected={todos ?? []}
            // Every selected todo has to be in hand before "on all of them"
            // means anything.
            complete={!!todos && todos.length === refs.length}
            busy={busy}
            onApply={edit => {
              run({
                onSelect: () =>
                  dispatch(action, {
                    ...(edit.add.length ? { [addParam]: edit.add.join(',') } : {}),
                    ...(edit.remove.length && removeParam
                      ? { [removeParam]: edit.remove.join(',') }
                      : {}),
                  }),
              });
              close();
            }}
          />
        );
        base.onSelect = () => {};
      }

      return base;
    });

    // Danger last, so the destructive action is never the item directly under
    // the cursor that just opened the menu. The bar groups by section in
    // first-appearance order, so this ordering is the one the menu shows.
    return descriptors.sort((a, b) =>
      Number(a.section === 'Danger') - Number(b.section === 'Danger'));
  }, [catalog, selection, runAction, onResult, onError]);
}

/**
 * The descriptors both layouts render, with each batch reported as a toast.
 *
 * The outcome moved out of the action bar's body: a partial failure names every
 * todo that failed and why, which does not fit in a toolbar row — and inside
 * DataTable's flex toolbar there is nowhere to put it at all.
 */
export function useTodoBulkToolbar({
  selection,
  todos,
  labelCounts,
  tags,
  onApplied,
}: Omit<TodoSelectionActionsOptions, 'onResult' | 'onError'> & {
  onApplied?: () => void;
}): DataTableSelectionAction[] {
  const { toast } = useToast();
  return useTodoSelectionActions({
    selection,
    ...(todos ? { todos } : {}),
    ...(labelCounts ? { labelCounts } : {}),
    ...(tags ? { tags } : {}),
    onResult: (result, action) => {
      // "Started", not "ran": a run-shaped action dispatches agent sessions that
      // land their edits later, so claiming the work is done would be a lie.
      const verb = RUN_SHAPED_ACTIONS.has(action.name) ? 'Started' : 'Updated';
      toast({
        message: todoBulkResultMessage(result, verb),
        tone: result.failed > 0 ? 'warning' : 'success',
        // A partial failure names each todo and why; that needs reading time.
        durationMs: result.failed > 0 ? 0 : undefined,
      });
      onApplied?.();
    },
    onError: (error, action) => {
      toast({ message: `${todoBulkActionLabel(action)} failed: ${error.message}`, tone: 'danger', durationMs: 0 });
    },
  });
}

const RUN_SHAPED_ACTIONS = new Set(['run', 'plan', 'triage']);

/** Re-exported so a host can report a batch without importing the data layer. */
export { todoBulkResultMessage };
export const TodoSelectionClearIcon = UiClose;
