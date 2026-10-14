import { useEffect, useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { Button, Field, Modal, Select } from '@flanksource/clicky-ui/components';
import { UiCheck, UiSearch } from '@flanksource/clicky-ui/icons';
import type { TodoPRLinkRelation } from '../../types';
import { inputClass } from './format';
import { prSearchQueryOptions, useLinkTodoPR } from './todoLinkQueries';

const SEARCH_DEBOUNCE_MS = 250;

function useDebounced<T>(value: T, delayMs: number): T {
  const [settled, setSettled] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setSettled(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);
  return settled;
}

// Links a pull request to a todo: pick one from the persisted PR index, or paste
// its URL. The form mounts on open so every opening starts empty and nothing is
// searched while the dialog is closed.
export function LinkPRDialog({ open, ...props }: { open: boolean; dir: string; todoRef: string; onClose: () => void }) {
  if (!open) return null;
  return <LinkPRForm {...props} />;
}

function LinkPRForm({ dir, todoRef, onClose }: { dir: string; todoRef: string; onClose: () => void }) {
  const [search, setSearch] = useState('');
  const [url, setUrl] = useState('');
  const [relation, setRelation] = useState<TodoPRLinkRelation>('implements');
  const settledSearch = useDebounced(search.trim(), SEARCH_DEBOUNCE_MS);
  const results = useQuery({ ...prSearchQueryOptions({ query: settledSearch }), placeholderData: keepPreviousData });
  const link = useLinkTodoPR({ dir, todoRef });
  const target = url.trim();

  return (
    <Modal
      open
      onClose={onClose}
      title="Link pull request"
      size="lg"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button
            loading={link.isPending}
            disabled={!target}
            onClick={() => link.mutate({ url: target, relation }, { onSuccess: onClose })}
          >
            Link PR
          </Button>
        </div>
      }
    >
      <div className="space-y-3">
        {link.error && <div className="text-sm text-destructive">{link.error.message}</div>}
        <Field label="Find a pull request">
          <div className="relative">
            <UiSearch className="pointer-events-none absolute left-2.5 top-2.5 text-sm text-muted-foreground" />
            <input
              className={`${inputClass} pl-8`}
              value={search}
              placeholder="Title, repo or number"
              aria-label="Search pull requests"
              onChange={event => setSearch(event.currentTarget.value)}
              autoFocus
            />
          </div>
          <div className="mt-1 max-h-56 overflow-y-auto rounded-md border border-border">
            {results.isError && <p className="px-2 py-1.5 text-xs text-destructive">{results.error.message}</p>}
            {results.data?.length === 0 && <p className="px-2 py-1.5 text-xs text-muted-foreground">No matching pull requests</p>}
            {results.isPending && <p className="px-2 py-1.5 text-xs text-muted-foreground">Searching…</p>}
            {results.data?.map(pr => (
              <Button
                key={pr.url}
                variant="ghost"
                type="button"
                onClick={() => setUrl(pr.url)}
                className="flex h-auto w-full items-start justify-start gap-2 rounded-none px-2 py-1.5 text-left hover:bg-muted"
              >
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium text-foreground">{pr.title}</span>
                  <span className="block truncate font-mono text-[10px] text-muted-foreground">{pr.repo}#{pr.number} · {pr.state}{pr.isDraft ? ' · draft' : ''} · {pr.author}</span>
                </span>
                {pr.url === target && <UiCheck className="mt-0.5 text-xs text-primary" />}
              </Button>
            ))}
          </div>
        </Field>
        <Field label="Pull request URL">
          <input
            type="url"
            className={inputClass}
            value={url}
            placeholder="https://github.com/owner/repo/pull/123"
            aria-label="Pull request URL"
            onChange={event => setUrl(event.currentTarget.value)}
          />
        </Field>
        <Field label="Relation">
          <Select value={relation} onChange={event => setRelation(event.currentTarget.value as TodoPRLinkRelation)} className={inputClass} aria-label="Relation">
            <option value="implements">Implements this todo</option>
            <option value="related">Related</option>
          </Select>
        </Field>
      </div>
    </Modal>
  );
}
