import { providerIcon, type ChatModel, type ToolMeta } from '@flanksource/clicky-ui/chat';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import { familiesFromRuntimeCatalog, type RuntimeCatalogFamily, type SpecRuntimeFamily } from '@flanksource/clicky-ui/ai';
import { UiRobotAi, UiSparkles } from '@flanksource/clicky-ui/icons';
import type { TodoRunAgent, TodoRunDriver, TodoRunEffort } from '../../types';

// Captain's run context owns every selectable runtime and model. This module
// only supplies presentation metadata and projects the returned catalog into
// clicky-ui's runtime controls.

// RunProvider is the coding agent / vendor a run targets — the same axis as the
// driver's agent half (claude or codex).
export type RunProvider = TodoRunAgent;

// RunMechanism is the canonical authored runtime mode and execution driver.
export type RunMechanism = 'cmux' | 'agent' | 'cli' | 'api';

export interface ProviderCatalog {
  id: RunProvider;
  // Display label for the provider's segment ("OpenAI" for the codex agent).
  label: string;
  // Model provider key used by clicky-ui's family filter.
  provider: string;
  // Brand/provider glyph shown on provider segments and run dropdown headers.
  icon: StaticIconComponent;
  iconColor?: string;
}

export interface RunModeMechanism {
  value: RunMechanism;
  label: string;
  driver: TodoRunDriver;
}

// RunModeCatalog is one (provider, mode) runtime the operator can pick, with the
// models, auth state and binary that runtime actually has.
export interface RunModeCatalog {
  id: string;
  label: string;
  provider: string;
  agent: RunProvider;
  defaultModel: string;
  driver: TodoRunDriver;
  mechanisms: RunModeMechanism[];
  models: ChatModel[];
  configured?: boolean;
  type?: string;
  authMethod?: string;
  authDetail?: string;
  binary?: string;
  binaryMissing?: string;
  modelError?: string;
}

// RunContextLifecycleStep is one step of the run dialog's lifecycle catalog:
// its display label, which prompt template it runs, and whether it is
// read-only (offered for inspection — e.g. from the history — but not a step
// the dialog's picker dispatches directly).
export interface RunContextLifecycleStep {
  name: string;
  label: string;
  prompt: string;
  readOnly: boolean;
}

export interface RunContext {
  modes: RunModeCatalog[];
  runtimes: RuntimeCatalogFamily[];
  models: ChatModel[];
  efforts: TodoRunEffort[];
  defaultMode?: string;
  defaultProvider?: string;
  // promptDefaults is the (mode, model) each lifecycle step resolves to,
  // keyed by step name — the server running todos/spec.Resolve, the same
  // resolution the run performs. Seeding a step's dialog from defaultMode
  // instead sends an account-wide default as if the operator had chosen it,
  // which outranks the frontmatter that step's prompt pins.
  promptDefaults?: Record<string, { mode?: string; model?: string }>;
  // lifecycle is the run dialog's step catalog — every step the operator can
  // pick from, in the order the pipeline runs them.
  lifecycle: { steps: RunContextLifecycleStep[] };
  // tools is the agent tool catalog the run dialog's tool-permissions control
  // renders; served by /api/todos/run/context (gavel drivers.DefaultTools).
  tools: ToolMeta[];
}

const CLAUDE: ProviderCatalog = {
  id: 'claude',
  label: 'Claude',
  provider: 'anthropic',
  icon: providerIcon('anthropic') ?? UiSparkles,
  iconColor: '#D97757',
};

const CODEX: ProviderCatalog = {
  id: 'codex',
  label: 'OpenAI',
  provider: 'openai',
  icon: providerIcon('openai') ?? UiRobotAi,
  iconColor: '#10A37F',
};

export const PROVIDERS: ProviderCatalog[] = [CLAUDE, CODEX];

export function modesForAgent(context: RunContext, agent: RunProvider): RunModeCatalog[] {
  return context.modes.filter(mode => mode.agent === agent);
}

export function modeCatalog(context: RunContext, id: string, agent: RunProvider): RunModeCatalog {
  const byID = context.modes.find(mode => mode.id === id && mode.agent === agent && mode.models.length > 0);
  if (byID) return byID;
  const first = modesForAgent(context, agent).find(mode => mode.models.length > 0);
  if (first) return first;
  throw new Error(`Captain returned no models for ${agent}`);
}

export function defaultModeForAgent(context: RunContext, agent: RunProvider): RunModeCatalog {
  const preferred = context.defaultMode
    ? context.modes.find(mode => mode.id === context.defaultMode && mode.agent === agent && mode.models.length > 0)
    : undefined;
  return preferred ?? modeCatalog(context, '', agent);
}

// buildRunFamilies forwards Captain's runtime catalog through Clicky's display
// projection; Gavel never reconstructs (provider, mode) pairs.
export function buildRunFamilies(context: RunContext): SpecRuntimeFamily[] {
  if (context.runtimes.length === 0) {
    throw new Error('Captain returned no runtime catalog');
  }
  return familiesFromRuntimeCatalog(context.runtimes);
}

export function isCmuxMode(mode: string | undefined): boolean {
  return mode === 'cmux';
}

// agentForRuntime selects the provider axis from the model catalog. The mode is
// intentionally insufficient because api/agent/cli/cmux are shared by families.
export function agentForRuntime(context: RunContext, mode: string | undefined, model: string | undefined): RunProvider {
  const value = mode ?? '';
  const byModel = context.modes.find(item =>
    item.id === value && item.models.some(entry => entry.id === model),
  )?.agent;
  if (byModel) return byModel;
  const defaultAgent = PROVIDERS.find(item => item.provider === context.defaultProvider)?.id;
  return defaultAgent
    ?? context.modes.find(item => item.id === context.defaultMode && item.models.length > 0)?.agent
    ?? context.modes.find(item => item.models.length > 0)?.agent
    ?? context.modes[0]?.agent
    ?? (() => { throw new Error('Captain returned no run providers'); })();
}

// modelsForSelection/defaultModelForSelection return the model list and
// sentinel default model for whichever runtime `mode` currently selects.
export function modelsForSelection(context: RunContext, agent: RunProvider, mode: string | undefined): ChatModel[] {
  return modeCatalog(context, mode ?? '', agent).models;
}

export function defaultModelForSelection(context: RunContext, agent: RunProvider, mode: string | undefined): string {
  return modeCatalog(context, mode ?? '', agent).defaultModel;
}

// driverForSelection returns the canonical mechanism unchanged on both fields.
export function driverForSelection(
  context: RunContext,
  agent: RunProvider,
  mode: string | undefined,
): { driver: TodoRunDriver; runMode?: string } {
  const cat = modeCatalog(context, mode ?? '', agent);
  return { driver: cat.driver, runMode: cat.id };
}
