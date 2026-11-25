import type { ReactNode } from 'react';
import { ModalProvider } from '@flanksource/clicky-ui/components';
import { prActionForms } from './PRActionDialog';
import { todoForms } from './todos/CreateTodoDialog';
import { prTodoForms } from './todos/HostedPRTodoDialog';

const forms = { ...prActionForms, ...prTodoForms, ...todoForms };

export function PRDialogProvider({ scope, children }: { scope: string; children: ReactNode }) {
  return <ModalProvider scope={`gavel:${scope}`} forms={forms}>{children}</ModalProvider>;
}
