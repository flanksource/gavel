import { ChatWindowLayer } from '@flanksource/clicky-ui/ai';
import { chatContextTypeConfig, GavelContextPicker } from './chatContext';

// Pages that never get the assistant FAB: the native menubar webview and the
// focused new-todo window (standalone, or embedded in the React Grab dialog).
const chatlessPaths = new Set(['/menubar', '/todos/new']);

export function chatLayerEnabled(pathname: string): boolean {
  return !chatlessPaths.has(pathname.replace(/\/+$/, '') || '/');
}

export function ChatLayer() {
  return (
    <ChatWindowLayer
      title="Gavel assistant"
      sessionsApi="/api/chat/sessions"
      toolsApi="/api/chat/tools"
      runtimesApi="/api/chat/runtimes"
      defaultToolPolicy="auto"
      contextTypeConfig={chatContextTypeConfig}
      renderContextPicker={props => <GavelContextPicker {...props} />}
      chat={{
        api: '/api/chat',
        modelsApi: '/api/chat/models',
        placeholder: 'Ask about tracked work…',
        suggestions: [
          'List open TODOs',
          'List my projects',
          'Show high priority work',
          'Find TODOs ready to run',
        ],
      }}
    />
  );
}
