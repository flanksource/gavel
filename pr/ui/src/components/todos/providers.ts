import type { ComboboxOption } from '@flanksource/clicky-ui/components';
import type { StaticIconComponent } from '@flanksource/clicky-ui/data';
import { UiRobotAi, UiSparkles } from '@flanksource/clicky-ui/icons';
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
  // Suggested models for the picker. The picker allows a custom value too, so a
  // pinned model id can still be typed in.
  models: ComboboxOption[];
  efforts: TodoRunEffort[];
}

const EFFORTS: TodoRunEffort[] = ['low', 'medium', 'high'];

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
    { value: 'claude', label: 'Default' },
    { value: 'opus', label: 'Opus' },
    { value: 'sonnet', label: 'Sonnet' },
    { value: 'haiku', label: 'Haiku' },
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
    { value: 'codex', label: 'Default' },
    { value: 'gpt-5-codex', label: 'GPT-5 Codex' },
    { value: 'gpt-5', label: 'GPT-5' },
    { value: 'o3', label: 'o3' },
    { value: 'o4-mini', label: 'o4-mini' },
  ],
  efforts: EFFORTS,
};

export const PROVIDERS: ProviderCatalog[] = [CLAUDE, CODEX];

export function providerCatalog(id: RunProvider): ProviderCatalog {
  return id === 'codex' ? CODEX : CLAUDE;
}

// driverFor composes the TodoRunDriver from the two axes the dialog selects.
export function driverFor(provider: RunProvider, mechanism: RunMechanism): TodoRunDriver {
  return `${provider}-${mechanism}` as TodoRunDriver;
}
