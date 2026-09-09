import { useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { familiesFromRuntimeCatalog, PromptCatalogTable, PromptPage } from '@flanksource/clicky-ui/ai';
import { Combobox } from '@flanksource/clicky-ui/components';
import type { Project } from '../../types';
import { promptModelCatalog } from '../settings/models';
import { settingsPromptsQuery, settingsRunContextQuery } from '../settings/queries';
import { promptCatalogQuery, promptPageAdapter, scopeQueryFor, withDefaults } from './promptCatalog';

interface Props {
  projects: Project[];
  // scopeProject is the project whose config chain the catalog resolves against;
  // empty is the global scope (~/.gavel.yaml only).
  scopeProject: string;
  selectedId: string;
  onNavigate: (id: string, scopeProject: string) => void;
}

const GLOBAL_SCOPE = '__global__';

// PromptsView is the Prompts tab: every prompt gavel runs for a scope as a
// table, and a dedicated page per prompt for reviewing and tweaking it.
export function PromptsView({ projects, scopeProject, selectedId, onNavigate }: Props) {
  const client = useQueryClient();
  const scopeQuery = scopeQueryFor(scopeProject);
  const catalog = useQuery(promptCatalogQuery(scopeQuery));
  const descriptors = useQuery(settingsPromptsQuery());
  const runContext = useQuery(settingsRunContextQuery());

  const entries = useMemo(() => withDefaults(catalog.data ?? [], descriptors.data), [catalog.data, descriptors.data]);
  const models = useMemo(() => (runContext.data ? promptModelCatalog(runContext.data) : undefined), [runContext.data]);
  const families = useMemo(
    () => (runContext.data?.runtimes ? familiesFromRuntimeCatalog(runContext.data.runtimes) : undefined),
    [runContext.data],
  );
  const adapter = useMemo(() => promptPageAdapter(client, scopeQuery), [client, scopeQuery]);

  const scopeOptions = useMemo(
    () => [
      { value: GLOBAL_SCOPE, label: 'Global (~/.gavel.yaml)' },
      ...projects.map(project => ({ value: project.name, label: project.name, description: project.dir })),
    ],
    [projects],
  );

  const selected = selectedId ? entries.find(entry => entry.id === selectedId) : undefined;
  if (selectedId && selected) {
    return (
      <div className="h-full min-h-0 overflow-y-auto">
        <PromptPage
          key={`${selected.id}:${scopeQuery}`}
          entry={selected}
          adapter={adapter}
          models={models}
          families={families}
          onBack={() => onNavigate('', scopeProject)}
        />
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-2">
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Scope</span>
        <Combobox
          value={scopeProject || GLOBAL_SCOPE}
          onChange={value => onNavigate('', value === GLOBAL_SCOPE ? '' : value)}
          options={scopeOptions}
          className="w-72"
          placeholder="Choose a scope"
        />
        <span className="text-xs text-muted-foreground">
          Prompts resolve through every .gavel.yaml layer that applies to the selected scope.
        </span>
      </div>
      {selectedId && !selected && !catalog.isLoading && (
        <div role="alert" className="shrink-0 border-b border-amber-200 bg-amber-50 px-4 py-2 text-xs text-amber-800 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          No prompt {selectedId} in this scope.
        </div>
      )}
      <div className="min-h-0 flex-1 p-4">
        <PromptCatalogTable
          className="min-h-0 flex-1"
          entries={entries}
          loading={catalog.isLoading}
          error={catalog.error ? catalog.error.message : null}
          selectedId={selectedId || undefined}
          onSelect={entry => onNavigate(entry.id, scopeProject)}
          emptyMessage="No prompts resolve for this scope."
        />
      </div>
    </div>
  );
}
