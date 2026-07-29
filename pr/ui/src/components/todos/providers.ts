import { providerIcon, type ChatModel, type ToolMeta } from '@flanksource/clicky-ui/chat';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import type { SpecRuntimeFamily } from '@flanksource/clicky-ui/ai';
import { UiColumns, UiRobotAi, UiSparkles } from '@flanksource/clicky-ui/icons';
import type { TodoRunAgent, TodoRunDriver, TodoRunEffort } from '../../types';

// Captain's run context owns every selectable backend and model. This module
// only supplies presentation metadata and projects the returned catalog into
// clicky-ui's runtime controls.

// RunProvider is the coding agent / vendor a run targets — the same axis as the
// driver's agent half (claude or codex).
export type RunProvider = TodoRunAgent;

// RunMechanism is the user-facing runtime mode. The driver string still carries
// implementation detail such as claude-headless; the picker presents agent/cli/API.
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
  efforts: TodoRunEffort[];
  defaultBackend?: string;
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

// The compact picker is intentionally Agent-first. Advanced remembers whichever
// backend the user chose, but switching provider in the compact flow always
// lands on that provider's headless Agent backend.
export function agentBackendForAgent(context: RunContext, agent: RunProvider): RunBackendCatalog {
  const agentBackend = backendsForAgent(context, agent).find(
    backend => backend.models.length > 0 && (backend.mechanisms.some(mechanism => mechanism.value === 'agent') || backend.id.endsWith('-agent')),
  );
  return agentBackend ?? defaultBackendForAgent(context, agent);
}

// driverFor composes the TodoRunDriver from the two axes the dialog selects.
export function driverFor(provider: RunProvider, mechanism: RunMechanism): TodoRunDriver {
  return `${provider}-${mechanism}` as TodoRunDriver;
}

// buildRunFamilies maps the whoami-backed RunContext onto clicky's two-axis
// Family -> Mode picker: one family per provider, with every mode coming from a
// backend row served by /api/todos/run/context.
export function buildRunFamilies(context: RunContext): SpecRuntimeFamily[] {
  const agents = new Set(context.backends.map(backend => backend.agent));
  return PROVIDERS.filter(provider => agents.has(provider.id)).map((provider) => {
    const backendModes = backendsForAgent(context, provider.id).filter(item => item.models.length > 0).map((item) => ({
      id: item.id,
      label: item.configured === false ? `${item.label} (not ready)` : item.label,
      backend: item.id,
      icon: item.driver.endsWith('cmux') ? UiColumns : provider.icon,
      title: item.label,
    }));
    return {
      id: provider.id,
      label: provider.label,
      provider: provider.provider,
      modes: backendModes,
    };
  });
}

// isCmuxBackend reports whether `backend` is the provider's cmux mode, as
// opposed to one of its captain (headless) backends.
export function isCmuxBackend(agent: RunProvider, backend: string | undefined): boolean {
  return (backend ?? '') === driverFor(agent, 'cmux');
}

// agentForBackend recovers which provider a `spec.backend` value belongs to,
// so the dialog can derive provider-scoped state (models, driver) from the
// single backend string the two-axis picker writes.
export function agentForBackend(context: RunContext, backend: string | undefined): RunProvider {
  const value = backend ?? '';
  for (const provider of PROVIDERS) {
    if (value === driverFor(provider.id, 'cmux')) return provider.id;
  }
  return context.backends.find((item) => item.id === value && item.models.length > 0)?.agent
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

// driverForSelection composes the TodoRunDriver + optional captain backend id
// the run/preview/submit payloads need from the two-axis picker's single
// `backend` selection.
export function driverForSelection(
  context: RunContext,
  agent: RunProvider,
  backend: string | undefined,
): { driver: TodoRunDriver; runBackend?: string } {
  const cat = backendCatalog(context, backend ?? '', agent);
  if (isCmuxBackend(agent, cat.id)) return { driver: cat.driver };
  return { driver: cat.driver, runBackend: cat.id };
}
