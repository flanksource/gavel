import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ToastProvider } from '@flanksource/clicky-ui/components';

function testQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, refetchOnWindowFocus: false },
    },
  });
}

export function queryTestWrapper(client = testQueryClient()) {
  return function QueryTestWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

// For code that reports by toast (useTodoRun, the cmux menu section), as the
// app root does. Kept separate because many suites mock the components module.
export function queryToastTestWrapper(client = testQueryClient()) {
  return function QueryToastTestWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}><ToastProvider>{children}</ToastProvider></QueryClientProvider>;
  };
}
