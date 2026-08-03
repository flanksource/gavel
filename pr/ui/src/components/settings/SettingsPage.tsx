import { useState, useEffect, useMemo, Suspense } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Button,
  JsonSchemaForm,
  Tabs,
} from "@flanksource/clicky-ui/components";
import type {
  JsonSchemaObject,
  JsonSchemaProperty,
  TabItem,
} from "@flanksource/clicky-ui/components";
import {
  UiFolder,
  UiGavel,
  UiChevronRight,
  UiClose,
  UiGitBranch,
} from "@flanksource/clicky-ui/icons";
import type { Project } from "../../types";
import { Spinner } from "../../icons/Spinner";
import { sectionIcon } from "../../icons/settings";
import { useProjectRegistration, ProjectFields } from "../ProjectForm";
import { buildRunFamilies } from "../todos/providers";
import {
  TABS,
  WORKSPACE_TAB,
  SECTION_TITLES,
  decorateSchema,
  usesPromptPicker,
} from "./schema";
import { promptModelCatalog } from "./models";
import { buildPre, buildPost } from "./extensions";
import { LayerSwitch } from "./LayerSwitch";
import { SettingsSectionCard } from "./SettingsSectionCard";
import { SaveBar } from "./SaveBar";
import type { GavelTrace, SettingsLayer } from "./provenance";
import {
  settingsConfigQuery,
  settingsPromptsQuery,
  settingsRunContextQuery,
  settingsSchemaQuery,
  settingsTraceQuery,
} from "./queries";
import { useSaveSettingsMutation } from "./mutations";

// Two-digit section index shown in each SectionCard header (01, 02, …).
const pad2 = (n: number) => String(n).padStart(2, "0");

// SettingsScope selects what the page targets: the user's global ~/.gavel.yaml
// (header gear) or one registered workspace (project gear). Project scope also
// unlocks the Project layer and the Workspace registration tab.
export type SettingsScope =
  | { kind: "global" }
  | { kind: "project"; project: Project };

interface Props {
  scope: SettingsScope;
  /** Repos offered in the Workspace tab's picker (project scope only). */
  repoOptions: string[];
  onClose: () => void;
  /** Called after the workspace registration is saved or deleted. */
  onSaved: () => void;
}

// scopeQueryFor maps the edited layer to the single-file settings endpoint's
// query: the user layer is always the global ~/.gavel.yaml; the project layer is
// the workspace's own .gavel.yaml.
function scopeQueryFor(layer: SettingsLayer, project: Project | null): string {
  return layer === "user" || !project
    ? "scope=global"
    : `project=${encodeURIComponent(project.name)}`;
}

// SettingsPage is the full-page settings surface: a schema-driven form over one
// .gavel.yaml layer, wrapped in the redesign's chrome (layer switch, section
// tabs, per-field provenance badges, sticky save bar). It edits a single layer
// (never the merged view) so saving the project file never bakes in user values.
export function SettingsPage({ scope, repoOptions, onClose, onSaved }: Props) {
  const project = scope.kind === "project" ? scope.project : null;
  const hasProject = project != null;

  const [value, setValue] = useState<Record<string, unknown>>({});
  const [path, setPath] = useState("");
  const [layer, setLayer] = useState<SettingsLayer>(
    hasProject ? "project" : "user"
  );
  const [tab, setTab] = useState(TABS[0].id);
  const [dirty, setDirty] = useState(false);
  const [error, setError] = useState("");

  const reg = useProjectRegistration({ open: hasProject, project });
  const isWorkspaceTab = hasProject && tab === WORKSPACE_TAB;
  const query = scopeQueryFor(layer, project);
  const traceScopeQuery = hasProject
    ? `project=${encodeURIComponent(project.name)}`
    : "scope=global";
  const schemaResult = useQuery(settingsSchemaQuery());
  const promptsResult = useQuery(settingsPromptsQuery());
  const runContextResult = useQuery(settingsRunContextQuery());
  const traceResult = useQuery(settingsTraceQuery(traceScopeQuery));
  const configResult = useQuery(settingsConfigQuery(query));
  const saveSettings = useSaveSettingsMutation(query);
  const schema = useMemo(
    () => schemaResult.data ? decorateSchema(schemaResult.data) : null,
    [schemaResult.data]
  );
  const registry = promptsResult.data ?? null;
  const runContext = runContextResult.data ?? null;
  const trace: GavelTrace | null = traceResult.data ?? null;

  const promptModels = useMemo(
    () => runContext ? promptModelCatalog(runContext) : [],
    [runContext]
  );
  const promptFamilies = useMemo(
    () => runContext ? buildRunFamilies(runContext) : [],
    [runContext]
  );
  const pre = useMemo(() => buildPre(), []);
  const post = useMemo(
    () =>
      registry
        ? buildPost(registry, query, promptModels, promptFamilies, trace)
        : [],
    [registry, query, promptModels, promptFamilies, trace]
  );
  const sectionTab = TABS.find((t) => t.id === tab) ?? TABS[0];
  // Sections present in the schema, in tab order.
  const activeSections = useMemo(
    () =>
      schema ? sectionTab.sections.filter((s) => schema.properties?.[s]) : [],
    [schema, sectionTab]
  );

  // Reset layer/tab whenever the target scope changes.
  useEffect(() => {
    setLayer(hasProject ? "project" : "user");
    setTab(TABS[0].id);
  }, [hasProject, project?.name]);

  useEffect(() => {
    setError("");
    setDirty(false);
  }, [query]);

  // Load the edited layer's single file. Re-runs when the layer switches.
  useEffect(() => {
    if (!configResult.data) return;
    setDirty(false);
    setValue(configResult.data.config ?? {});
    setPath(configResult.data.path || "");
  }, [query, configResult.data]);

  async function saveConfig() {
    setError("");
    try {
      const saved = await saveSettings.mutateAsync(value);
      setValue(saved.config ?? {});
      setPath(saved.path || "");
      setDirty(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save settings");
    }
  }

  async function onSave() {
    if (isWorkspaceTab) {
      if (await reg.save()) onSaved();
    } else {
      await saveConfig();
    }
  }

  async function onDiscard() {
    // Reload the current layer, dropping in-memory edits.
    setError("");
    try {
      const result = await configResult.refetch({ throwOnError: true });
      if (result.data) {
        setValue(result.data.config ?? {});
        setPath(result.data.path || "");
        setDirty(false);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to reload settings");
    }
  }

  // updateSection replaces one top-level section of the config and marks dirty.
  // Each SectionCard edits its own slice, so sections never clobber each other.
  function updateSection(section: string, next: unknown) {
    setValue((prev) => ({ ...prev, [section]: next }));
    setDirty(true);
  }

  // renderSectionForm renders one section's fields. Object sections use their
  // inner schema as the form root (no duplicate heading — the SectionCard owns
  // it); array/scalar sections (pre/post hooks) fall back to a titleless wrapper.
  function renderSectionForm(section: string) {
    const node = schema!.properties![section] as JsonSchemaProperty;
    const common = {
      pre,
      post,
      size: "sm" as const,
      showPreferencesMenu: false,
      persistPreferences: false,
    };
    if (node.properties && !usesPromptPicker(node)) {
      return (
        <JsonSchemaForm
          schema={node as JsonSchemaObject}
          value={(value[section] as Record<string, unknown>) ?? {}}
          onChange={(next) => updateSection(section, next)}
          {...common}
        />
      );
    }
    const wrapper: JsonSchemaObject = {
      type: "object",
      properties: { [section]: { ...node, title: "" } },
      "x-order": [section],
    };
    return (
      <JsonSchemaForm
        schema={wrapper}
        value={{ [section]: value[section] }}
        onChange={(next) =>
          updateSection(section, (next as Record<string, unknown>)[section])
        }
        {...common}
      />
    );
  }

  const configReady = schema && registry;
  const queryFailure = schemaResult.error
    ?? promptsResult.error
    ?? runContextResult.error
    ?? traceResult.error
    ?? configResult.error;
  const errorMessage = error || (queryFailure instanceof Error ? queryFailure.message : "");
  const loading = configResult.isPending || schemaResult.isPending || promptsResult.isPending;
  const layerLabel = layer === "user" ? "User layer" : "Project layer";
  // clicky-ui ships React 18 types; our icon components are typed against React
  // 19, so the structurally-identical icon prop needs a cast at this boundary.
  const sectionTabItems: TabItem[] = TABS.map((t) => ({
    id: t.id,
    label: t.label,
    icon: sectionIcon[t.sections[0]] as unknown as TabItem["icon"],
  }));
  const tabItems: TabItem[] = hasProject
    ? [
        {
          id: WORKSPACE_TAB,
          label: "Workspace",
          icon: UiFolder as unknown as TabItem["icon"],
        },
        ...sectionTabItems,
      ]
    : sectionTabItems;

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-background">
      <header className="flex-shrink-0 border-b border-border bg-card">
        <div className="flex h-[52px] items-center gap-3 px-5">
          <div className="flex items-center gap-2">
            <span className="inline-flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br from-primary to-indigo-500 text-white">
              <UiGavel className="text-[15px]" />
            </span>
            <span className="text-sm font-bold tracking-tight text-foreground">
              gavel
            </span>
          </div>
          <UiChevronRight className="text-[13px] text-muted-foreground" />
          <span className="text-sm font-semibold text-foreground">
            {hasProject ? `${project.name} settings` : "Settings"}
          </span>
          <div className="flex-1" />
          <LayerSwitch
            layer={layer}
            onChange={setLayer}
            hasProject={hasProject}
            path={path}
          />
          <Button
            variant="ghost"
            size="icon"
            onClick={onClose}
            aria-label="Close settings"
            className="text-muted-foreground hover:text-foreground"
          >
            <UiClose />
          </Button>
        </div>
        <Tabs
          tabs={tabItems}
          value={tab}
          onChange={setTab}
          className="flex-nowrap overflow-x-auto whitespace-nowrap border-t border-border px-3"
        />
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto @container">
        <div className="mx-auto max-w-[820px] px-density-4 py-density-4">
          {errorMessage && (
            <div role="alert" className="mb-3 text-sm text-destructive">{errorMessage}</div>
          )}

          {isWorkspaceTab ? (
            <section>
              <ProjectFields reg={reg} repoOptions={repoOptions} />
              <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={reg.remove}
                  loading={reg.deleting}
                >
                  Delete workspace
                </Button>
                <Button size="sm" onClick={onSave} loading={reg.saving}>
                  Save
                </Button>
              </div>
            </section>
          ) : loading || !configReady ? (
            <div className="flex items-center gap-2 py-10 text-sm text-muted-foreground">
              <Spinner /> Loading…
            </div>
          ) : (
            <>
              <header className="mb-density-4 border-b border-border pb-density-3">
                <p className="text-[10px] font-bold uppercase tracking-[0.12em] text-muted-foreground">
                  {layerLabel}
                </p>
                <div className="mt-1 flex flex-wrap items-center gap-density-3">
                  <h2 className="text-lg font-bold tracking-tight text-foreground">
                    {sectionTab.label}
                  </h2>
                  {path && (
                    <span className="inline-flex items-center gap-1.5 rounded-full border border-border bg-muted/40 px-density-2 py-0.5 font-mono text-[11px] text-muted-foreground">
                      <UiGitBranch className="size-3.5" />
                      {path}
                    </span>
                  )}
                </div>
              </header>
              <Suspense
                fallback={
                  <div className="py-8 text-sm text-muted-foreground">
                    Loading editor…
                  </div>
                }
              >
                {activeSections.map((sec, i) => {
                  const node = schema.properties?.[sec] as
                    | JsonSchemaProperty
                    | undefined;
                  const hint =
                    typeof node?.description === "string"
                      ? node.description
                      : undefined;
                  return (
                    <SettingsSectionCard
                      key={sec}
                      icon={sectionIcon[sec]}
                      title={SECTION_TITLES[sec] ?? sec}
                      number={pad2(i + 1)}
                      hint={hint}
                    >
                      {renderSectionForm(sec)}
                    </SettingsSectionCard>
                  );
                })}
              </Suspense>
            </>
          )}
        </div>
      </div>

      {!isWorkspaceTab && configReady && !loading && (
        <SaveBar
          path={path}
          dirty={dirty}
          saving={saveSettings.isPending}
          onDiscard={onDiscard}
          onSave={onSave}
        />
      )}
    </div>
  );
}
