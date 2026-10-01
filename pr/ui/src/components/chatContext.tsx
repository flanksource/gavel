import type { ChatContextPickerRenderProps, ContextTypeConfig } from '@flanksource/clicky-ui/ai';
import { UiCheck, UiFolderGit } from '@flanksource/clicky-ui/icons';
import {
  createOperationsApiClient,
  OperationEntityContextPicker,
  type EntityContextSurfaceFilter,
  type EntityContextSurfaceIcon,
  type EntityContextSurfaceText,
  type OperationsApiClient,
} from '@flanksource/clicky-ui/rpc';

// The generated entity document (pr/ui/entity_routes.go). It is served beside
// the hand-built /api/openapi.json, which describes the /api/projects CRUD
// routes rather than the entity operations the picker lists from.
export const GAVEL_ENTITY_OPENAPI_PATH = '/api/v1/openapi.json';

export function createGavelEntityClient(fetchImpl?: typeof fetch) {
  return createOperationsApiClient({
    openApiPath: GAVEL_ENTITY_OPENAPI_PATH,
    ...(fetchImpl ? { fetch: fetchImpl } : {}),
  });
}

const gavelEntityClient = createGavelEntityClient();

// The entity surfaces a chat can attach, keyed by surface key — which is also
// the `type` of the context item a picked row becomes.
export const CHAT_CONTEXT_SURFACES = {
  todo: { label: 'TODOs', icon: UiCheck },
  project: { label: 'Projects', icon: UiFolderGit },
} as const;

type ChatContextSurfaceKey = keyof typeof CHAT_CONTEXT_SURFACES;

function chatContextSurface(key: string) {
  return Object.hasOwn(CHAT_CONTEXT_SURFACES, key)
    ? CHAT_CONTEXT_SURFACES[key as ChatContextSurfaceKey]
    : undefined;
}

export const chatContextSurfaceFilter: EntityContextSurfaceFilter = surface =>
  chatContextSurface(surface.key) !== undefined;

const chatContextSurfaceLabel: EntityContextSurfaceText = surface =>
  chatContextSurface(surface.key)?.label ?? surface.title;

const chatContextSurfaceIcon: EntityContextSurfaceIcon = surface =>
  chatContextSurface(surface.key)?.icon;

export const chatContextTypeConfig: ContextTypeConfig = Object.fromEntries(
  Object.entries(CHAT_CONTEXT_SURFACES).map(([key, { icon }]) => [key, { icon }]),
);

export type GavelContextPickerProps = ChatContextPickerRenderProps & {
  client?: OperationsApiClient;
};

/** The chat window's "Add context" menu: attach TODOs and projects to a thread. */
export function GavelContextPicker({ client = gavelEntityClient, items, onAdd, onAddMany }: GavelContextPickerProps) {
  return (
    <OperationEntityContextPicker
      client={client}
      items={items}
      onAdd={onAdd}
      onAddMany={onAddMany}
      surfaceFilter={chatContextSurfaceFilter}
      surfaceLabel={chatContextSurfaceLabel}
      surfaceIcon={chatContextSurfaceIcon}
    />
  );
}
