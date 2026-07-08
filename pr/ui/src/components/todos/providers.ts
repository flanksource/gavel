import type { ChatModel, ToolMeta } from '@flanksource/clicky-ui/chat';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import type { SpecRuntimeFamily } from '@flanksource/clicky-ui/ai';
import { UiColumns, UiRobotAi, UiSparkles } from '@flanksource/clicky-ui/icons';
import type { TodoRunAgent, TodoRunDriver, TodoRunEffort } from '../../types';

// A run's driver is `<provider>-<mechanism>`. The advanced dialog selects the
// provider first (segmented control) and the mechanism + model second (pickers),
// then composes them into a TodoRunDriver. This module is the single catalog of
// what each provider offers, so the pickers show real per-provider choices
// instead of a free-text model field and a static effort list.

// RunProvider is the coding agent / vendor a run targets — the same axis as the
// driver's agent half (claude or codex).
export type RunProvider = TodoRunAgent;

// RunMechanism is how the agent is driven. cmux is the interactive TUI; the rest
// are the structured paths (headless stream-json, the SDK bridge, the raw API).
export type RunMechanism = 'cmux' | 'headless' | 'sdk' | 'api';

export interface ProviderCatalog {
  id: RunProvider;
  // Display label for the provider's segment ("OpenAI" for the codex agent).
  label: string;
  // clicky-ui Ui* icon component shown on the provider's segment.
  icon: StaticIconComponent;
  // Sentinel model meaning "let the agent pick its default"; the server maps it
  // back to the agent (cmux.ResolveAgent / headless newStreamer).
  defaultModel: string;
  mechanisms: Array<{ value: RunMechanism; label: string }>;
  // Suggested models for the shared clicky-ui AI model picker.
  models: ChatModel[];
  efforts: TodoRunEffort[];
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

const EFFORTS: TodoRunEffort[] = ['low', 'medium', 'high', 'xhigh'];

// FALLBACK_TOOLS mirrors gavel's drivers.DefaultTools so the picker still renders
// when the run-context fetch fails. Keep in sync with todoRunToolCatalog (Go).
const FALLBACK_TOOLS: ToolMeta[] = [
  { name: 'Read', label: 'Read', group: 'Files', defaultPermission: 'on' },
  { name: 'Edit', label: 'Edit', group: 'Files', defaultPermission: 'on' },
  { name: 'Write', label: 'Write', group: 'Files', defaultPermission: 'on' },
  { name: 'Bash', label: 'Bash', group: 'Shell', defaultPermission: 'ask' },
  { name: 'Glob', label: 'Glob', group: 'Search', defaultPermission: 'on' },
  { name: 'Grep', label: 'Grep', group: 'Search', defaultPermission: 'on' },
];

const CLAUDE: ProviderCatalog = {
  id: 'claude',
  label: 'Claude',
  icon: UiSparkles,
  defaultModel: 'claude',
  mechanisms: [
    { value: 'cmux', label: 'cmux (TUI)' },
    { value: 'headless', label: 'headless' },
    { value: 'sdk', label: 'SDK' },
    { value: 'api', label: 'API' },
  ],
  models: [
    { id: 'claude', provider: 'anthropic', label: 'Default', reasoning: true, configured: true },
    { id: 'opus', provider: 'anthropic', label: 'Opus', reasoning: true, configured: true },
    { id: 'sonnet', provider: 'anthropic', label: 'Sonnet', reasoning: true, configured: true },
    { id: 'haiku', provider: 'anthropic', label: 'Haiku', reasoning: true, configured: true },
  ],
  efforts: EFFORTS,
};

const CODEX: ProviderCatalog = {
  id: 'codex',
  label: 'OpenAI',
  icon: UiRobotAi,
  defaultModel: 'codex',
  mechanisms: [
    { value: 'cmux', label: 'cmux (TUI)' },
    { value: 'headless', label: 'headless' },
  ],
  models: [
    { id: 'codex', provider: 'openai', label: 'Default', reasoning: true, configured: true },
    { id: 'gpt-5-codex', provider: 'openai', label: 'GPT-5 Codex', reasoning: true, configured: true },
    { id: 'gpt-5', provider: 'openai', label: 'GPT-5', reasoning: true, configured: true },
    { id: 'o3', provider: 'openai', label: 'o3', reasoning: true, configured: true },
    { id: 'o4-mini', provider: 'openai', label: 'o4-mini', reasoning: true, configured: true },
  ],
  efforts: EFFORTS,
};

export const PROVIDERS: ProviderCatalog[] = [CLAUDE, CODEX];

export const FALLBACK_RUN_CONTEXT: RunContext = {
  defaultBackend: 'claude-agent',
  efforts: EFFORTS,
  tools: FALLBACK_TOOLS,
  backends: [
    {
      id: 'claude-agent',
      label: 'Claude Agent',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-agent-sonnet',
      driver: 'claude-headless',
      mechanisms: [{ value: 'headless', label: 'headless', driver: 'claude-headless' }],
      models: [
        { id: 'claude-agent-opus', provider: 'anthropic', label: 'Opus', reasoning: true, configured: true },
        { id: 'claude-agent-sonnet', provider: 'anthropic', label: 'Sonnet', reasoning: true, configured: true },
        { id: 'claude-agent-haiku', provider: 'anthropic', label: 'Haiku', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'claude-cli',
      label: 'Claude CLI',
      provider: 'anthropic',
      agent: 'claude',
      defaultModel: 'claude-agent-sonnet',
      driver: 'claude-headless',
      mechanisms: [{ value: 'headless', label: 'headless', driver: 'claude-headless' }],
      models: [
        { id: 'claude-agent-opus', provider: 'anthropic', label: 'Opus', reasoning: true, configured: true },
        { id: 'claude-agent-sonnet', provider: 'anthropic', label: 'Sonnet', reasoning: true, configured: true },
        { id: 'claude-agent-haiku', provider: 'anthropic', label: 'Haiku', reasoning: true, configured: true },
      ],
      configured: true,
    },
    {
      id: 'codex-cli',
      label: 'Codex CLI',
      provider: 'openai',
      agent: 'codex',
      defaultModel: 'gpt-5-codex',
      driver: 'codex-headless',
      mechanisms: [{ value: 'headless', label: 'headless', driver: 'codex-headless' }],
      models: [
        { id: 'gpt-5-codex', provider: 'openai', label: 'GPT-5 Codex', reasoning: true, configured: true },
      ],
      configured: true,
    },
  ],
};

export function providerCatalog(id: RunProvider): ProviderCatalog {
  return id === 'codex' ? CODEX : CLAUDE;
}

export function runContextWithFallback(context?: RunContext | null): RunContext {
  if (!context || context.backends.length === 0) return FALLBACK_RUN_CONTEXT;
  return {
    defaultBackend: context.defaultBackend || FALLBACK_RUN_CONTEXT.defaultBackend,
    efforts: context.efforts.length > 0 ? context.efforts : FALLBACK_RUN_CONTEXT.efforts,
    backends: context.backends.length > 0 ? context.backends : FALLBACK_RUN_CONTEXT.backends,
    tools: context.tools && context.tools.length > 0 ? context.tools : FALLBACK_TOOLS,
  };
}

export function backendsForAgent(context: RunContext, agent: RunProvider): RunBackendCatalog[] {
  return context.backends.filter(backend => backend.agent === agent);
}

export function backendCatalog(context: RunContext, id: string, agent: RunProvider): RunBackendCatalog {
  const byID = context.backends.find(backend => backend.id === id && backend.agent === agent);
  if (byID) return byID;
  return backendsForAgent(context, agent)[0] ?? FALLBACK_RUN_CONTEXT.backends.find(backend => backend.agent === agent)!;
}

export function defaultBackendForAgent(context: RunContext, agent: RunProvider): RunBackendCatalog {
  const preferred = context.defaultBackend
    ? context.backends.find(backend => backend.id === context.defaultBackend && backend.agent === agent)
    : undefined;
  return preferred ?? backendCatalog(context, '', agent);
}

// driverFor composes the TodoRunDriver from the two axes the dialog selects.
export function driverFor(provider: RunProvider, mechanism: RunMechanism): TodoRunDriver {
  return `${provider}-${mechanism}` as TodoRunDriver;
}

// buildRunFamilies maps a RunContext onto clicky's two-axis Family -> Mode
// picker (PromptRunEditor/RuntimeModePicker): one family per provider, with
// cmux as the TUI mode and every configured captain backend as a further
// mode. A family's `provider` tag is taken from its own model catalog (not
// hardcoded) so it always matches the real `ChatModel.provider` value
// ('anthropic'/'openai') the models list carries — required for the shared
// component's modelsForFamily/modelBelongsToFamily narrowing to work at all.
export function buildRunFamilies(context: RunContext): SpecRuntimeFamily[] {
  return PROVIDERS.map((provider) => {
    const cmuxMode = {
      id: 'cmux',
      label: 'cmux (TUI)',
      backend: driverFor(provider.id, 'cmux'),
      icon: UiColumns,
      title: `${provider.label} multiplexer`,
    };
    const backendModes = backendsForAgent(context, provider.id).map((item) => ({
      id: item.id,
      label: item.configured === false ? `${item.label} (not ready)` : item.label,
      backend: item.id,
      icon: provider.icon,
      title: item.label,
    }));
    return {
      id: provider.id,
      label: provider.label,
      provider: provider.models[0]?.provider ?? provider.id,
      modes: [cmuxMode, ...backendModes],
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
  return context.backends.find((item) => item.id === value)?.agent ?? 'claude';
}

// modelsForSelection/defaultModelForSelection return the model list and
// sentinel default model for whichever mode (cmux or a captain backend)
// `backend` currently selects.
export function modelsForSelection(context: RunContext, agent: RunProvider, backend: string | undefined): ChatModel[] {
  if (isCmuxBackend(agent, backend)) return providerCatalog(agent).models;
  return backendCatalog(context, backend ?? '', agent).models;
}

export function defaultModelForSelection(context: RunContext, agent: RunProvider, backend: string | undefined): string {
  if (isCmuxBackend(agent, backend)) return providerCatalog(agent).defaultModel;
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
  if (isCmuxBackend(agent, backend)) return { driver: driverFor(agent, 'cmux') };
  const cat = backendCatalog(context, backend ?? '', agent);
  return { driver: cat.driver, runBackend: cat.id };
}
