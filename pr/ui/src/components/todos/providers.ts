import { providerIcon, type ChatModel, type ToolMeta } from '@flanksource/clicky-ui/chat';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import { familiesFromRuntimeCatalog, type RuntimeCatalogFamily, type SpecRuntimeFamily } from '@flanksource/clicky-ui/ai';
import { UiRobotAi, UiSparkles } from '@flanksource/clicky-ui/icons';
import type { TodoRunAgent, TodoRunDriver, TodoRunEffort } from '../../types';

// Captain's run context owns every selectable backend and model. This module
// only supplies presentation metadata and projects the returned catalog into
// clicky-ui's runtime controls.

// RunProvider is the coding agent / vendor a run targets — the same axis as the
// driver's agent half (claude or codex).
export type RunProvider = TodoRunAgent;

// RunMechanism is the canonical authored backend and execution driver.
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

export interface RunBackendMechanism {
  value: RunMechanism;
  label: string;
  driver: TodoRunDriver;
}

export interface RunBackendCatalog {
  id: string;
  label: string;
  provider: string;
  agent: RunProvider;
  defaultModel: string;
  driver: TodoRunDriver;
  mechanisms: RunBackendMechanism[];
  models: ChatModel[];
  configured?: boolean;
  type?: string;
  authMethod?: string;
  authDetail?: string;
  binary?: string;
  binaryMissing?: string;
  modelError?: string;
}

export interface RunContext {
  backends: RunBackendCatalog[];
  runtimes?: RuntimeCatalogFamily[];
  models?: ChatModel[];
  efforts: TodoRunEffort[];
  defaultBackend?: string;
  defaultProvider?: string;
  // promptDefaults is the (backend, model) each named prompt resolves to,
  // keyed by prompt name — the server running todos/spec.Resolve, the same
  // resolution the run performs. Seeding a prompt's dialog from defaultBackend
  // instead sends an account-wide default as if the operator had chosen it,
  // which outranks the frontmatter that prompt pins.
  promptDefaults?: Record<string, { backend?: string; model?: string }>;
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

export function backendsForAgent(context: RunContext, agent: RunProvider): RunBackendCatalog[] {
  return context.backends.filter(backend => backend.agent === agent);
}

export function backendCatalog(context: RunContext, id: string, agent: RunProvider): RunBackendCatalog {
  const byID = context.backends.find(backend => backend.id === id && backend.agent === agent && backend.models.length > 0);
  if (byID) return byID;
  const first = backendsForAgent(context, agent).find(backend => backend.models.length > 0);
  if (first) return first;
  throw new Error(`Captain returned no models for ${agent}`);
}

export function defaultBackendForAgent(context: RunContext, agent: RunProvider): RunBackendCatalog {
  const preferred = context.defaultBackend
    ? context.backends.find(backend => backend.id === context.defaultBackend && backend.agent === agent && backend.models.length > 0)
    : undefined;
  return preferred ?? backendCatalog(context, '', agent);
}

// buildRunFamilies forwards Captain's runtime catalog through Clicky's display
// projection; Gavel never reconstructs provider/backend pairs.
export function buildRunFamilies(context: RunContext): SpecRuntimeFamily[] {
	if (!context.runtimes || context.runtimes.length === 0) {
		throw new Error('Captain returned no runtime catalog');
	}
  return familiesFromRuntimeCatalog(context.runtimes);
}

export function isCmuxBackend(backend: string | undefined): boolean {
  return backend === 'cmux';
}

// agentForRuntime selects the provider axis from the model catalog. Backend is
// intentionally insufficient because api/agent/cli/cmux are shared by families.
export function agentForRuntime(context: RunContext, backend: string | undefined, model: string | undefined): RunProvider {
  const value = backend ?? '';
  const byModel = context.backends.find(item =>
    item.id === value && item.models.some(entry => entry.id === model),
  )?.agent;
  if (byModel) return byModel;
  const defaultAgent = PROVIDERS.find(item => item.provider === context.defaultProvider)?.id;
  return defaultAgent
    ?? context.backends.find(item => item.id === context.defaultBackend && item.models.length > 0)?.agent
    ?? context.backends.find(item => item.models.length > 0)?.agent
    ?? context.backends[0]?.agent
    ?? (() => { throw new Error('Captain returned no run providers'); })();
}

// modelsForSelection/defaultModelForSelection return the model list and
// sentinel default model for whichever mode (cmux or a captain backend)
// `backend` currently selects.
export function modelsForSelection(context: RunContext, agent: RunProvider, backend: string | undefined): ChatModel[] {
  return backendCatalog(context, backend ?? '', agent).models;
}

export function defaultModelForSelection(context: RunContext, agent: RunProvider, backend: string | undefined): string {
  return backendCatalog(context, backend ?? '', agent).defaultModel;
}

// driverForSelection returns the canonical mechanism unchanged on both fields.
export function driverForSelection(
  context: RunContext,
  agent: RunProvider,
  backend: string | undefined,
): { driver: TodoRunDriver; runBackend?: string } {
  const cat = backendCatalog(context, backend ?? '', agent);
  return { driver: cat.driver, runBackend: cat.id };
}
