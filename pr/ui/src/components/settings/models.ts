import type { ChatModel } from '@flanksource/clicky-ui/chat';
import type { RunContext } from '../todos/providers';

// Captain already serves the prompt model catalog with provider and mode as
// separate axes. Gavel forwards it unchanged.
export function promptModelCatalog(context: RunContext): ChatModel[] {
	return context.models ?? [];
}
