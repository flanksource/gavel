import { Component, type ReactNode } from 'react';
import { UiError } from '@flanksource/clicky-ui/icons';

// React error boundaries must be class components; there is no hook equivalent.
export class ErrorBoundary extends Component<{
  children: ReactNode;
  fallback?: (error: Error) => ReactNode;
}, { error: Error | null }> {
  state = { error: null as Error | null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  render() {
    if (!this.state.error) return this.props.children;
    if (this.props.fallback) return this.props.fallback(this.state.error);
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 p-6 text-sm">
        <UiError className="text-2xl text-red-500" />
        <p className="font-medium">Something went wrong rendering this view.</p>
        <pre className="max-w-full overflow-auto whitespace-pre-wrap rounded bg-muted p-2 font-mono text-xs text-muted-foreground">
          {this.state.error.message}
        </pre>
      </div>
    );
  }
}
