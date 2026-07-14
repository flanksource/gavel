import type { ChatModel } from '@flanksource/clicky-ui/chat';
import type { RunContext } from '../todos/providers';

// promptModelCatalog flattens a run context's per-backend model lists into one
// de-duplicated catalog for the prompt editor, unioning the backends that offer
// each model so a shared model keeps every backend it came from.
export function promptModelCatalog(context: RunContext): ChatModel[] {
  const models = new Map<string, ChatModel>();
  for (const backend of context.backends) {
    for (const model of backend.models) {
      const existing = models.get(model.id);
      const backends = new Set<string>(
        [...(existing?.backends ?? []), ...(model.backends ?? []), backend.id].filter(Boolean),
      );
      models.set(model.id, { ...existing, ...model, backends: [...backends] });
    }
  }
  return [...models.values()];
}
