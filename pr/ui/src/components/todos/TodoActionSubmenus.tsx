import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronRight, UiFolder, UiGitPr, UiLink } from '@flanksource/clicky-ui/icons';
import type { Project, TodoItem } from '../../types';
import { inputClass } from './format';
import { todoPRLinksQueryOptions, useUnlinkTodoPR } from './todoLinkQueries';
import { todoMatchesQuery } from './todoFilter';

export function SetParentSubmenu({
  disabled,
  candidates,
  currentParentId,
  onSelect,
  onCloseParent,
}: {
  disabled?: boolean;
  candidates: TodoItem[];
  currentParentId?: string;
  onSelect: (ref: string) => void;
  onCloseParent: () => void;
}) {
  const [query, setQuery] = useState('');
  const matches = candidates.filter(candidate => todoMatchesQuery(candidate, query));
  return (
    <DropdownMenu
      align="right"
      menuLabel="Choose the parent todo"
      menuClassName="w-72 max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
          title="Make this todo a child of another"
        >
          <UiLink className="shrink-0 text-sm text-muted-foreground" />
          <span className="min-w-0 flex-1">
            <span className="block truncate font-medium text-foreground">Set parent…</span>
            <span className="block truncate text-[11px] text-muted-foreground">Choose a todo</span>
          </span>
          <UiChevronRight className="text-[11px] text-muted-foreground" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          <input
            className={`${inputClass} mb-1`}
            value={query}
            placeholder="Filter todos"
            aria-label="Filter parent todos"
            onChange={event => setQuery(event.currentTarget.value)}
          />
          <div className="max-h-64 overflow-y-auto">
            {matches.length === 0 && <div className="px-2 py-1.5 text-muted-foreground">No matching todos</div>}
            {matches.map(candidate => (
              <Button
                key={candidate.ref}
                variant="ghost"
                type="button"
                disabled={disabled || candidate.id === currentParentId}
                onClick={() => {
                  close();
                  onCloseParent();
                  onSelect(candidate.ref);
                }}
                className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium text-foreground">{candidate.title}</span>
                  {candidate.shortId && <span className="block truncate font-mono text-[10px] text-muted-foreground">{candidate.shortId}</span>}
                </span>
                {candidate.id === currentParentId && <UiCheck className="mt-0.5 text-xs text-primary" />}
              </Button>
            ))}
          </div>
        </div>
      )}
    </DropdownMenu>
  );
}

export function MoveSubmenu({
  disabled,
  targets,
  onSelect,
  onCloseParent,
}: {
  disabled?: boolean;
  targets: Project[];
  onSelect: (dir: string) => void;
  onCloseParent: () => void;
}) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Move todo to another project"
      menuClassName="w-72 max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
          title="Move todo to another project"
          aria-label="Move todo to another project"
        >
          <UiFolder className="shrink-0 text-sm text-muted-foreground" />
          <span className="min-w-0 flex-1">
            <span className="block truncate font-medium text-foreground">Move to</span>
            <span className="block truncate text-[11px] text-muted-foreground">Choose project</span>
          </span>
          <UiChevronRight className="text-[11px] text-muted-foreground" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {targets.map(target => (
            <Button
              key={target.dir}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                onCloseParent();
                onSelect(target.dir);
              }}
              className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <UiFolder className="mt-0.5 shrink-0 text-sm text-muted-foreground" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">{target.name || target.dir}</span>
                <span className="block truncate font-mono text-[10px] text-muted-foreground">{target.dir}</span>
              </span>
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}

// Lists the PRs currently linked to the todo and unlinks the chosen one. The
// submenu is absent while there is nothing to unlink, or while the links cannot
// be read: the failure is shown on the overview's links, not duplicated here.
export function UnlinkPRSubmenu({
  dir,
  todoRef,
  disabled,
  onCloseParent,
}: {
  dir: string;
  todoRef: string;
  disabled?: boolean;
  onCloseParent: () => void;
}) {
  const links = useQuery(todoPRLinksQueryOptions(dir, todoRef)).data ?? [];
  const unlink = useUnlinkTodoPR({ dir, todoRef });
  if (links.length === 0) return null;
  return (
    <DropdownMenu
      align="right"
      menuLabel="Unlink a pull request"
      menuClassName="w-72 max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled || unlink.isPending}
          className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
          title="Remove a pull request link from this todo"
        >
          <UiGitPr className="shrink-0 text-sm text-muted-foreground" />
          <span className="min-w-0 flex-1">
            <span className="block truncate font-medium text-foreground">Unlink PR</span>
            <span className="block truncate text-[11px] text-muted-foreground">{links.length === 1 ? 'Choose the linked PR' : `${links.length} linked PRs`}</span>
          </span>
          <UiChevronRight className="text-[11px] text-muted-foreground" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {unlink.error && <div className="px-2 py-1.5 text-red-600">{unlink.error.message}</div>}
          {links.map(link => (
            <Button
              key={link.url}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => unlink.mutate({ url: link.url }, { onSuccess: () => { close(); onCloseParent(); } })}
              className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <UiGitPr className="mt-0.5 shrink-0 text-sm text-muted-foreground" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">PR #{link.number}{link.title ? ` ${link.title}` : ''}</span>
                <span className="block truncate font-mono text-[10px] text-muted-foreground">{link.repo} · {link.relation}</span>
              </span>
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}
