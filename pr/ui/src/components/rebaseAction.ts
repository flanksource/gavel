import { createContext, useContext } from 'react';
import type { DropdownMenuItem } from '@flanksource/clicky-ui/components';

// The project ref pane's "Rebase…" entry, offered as an option on the PR split
// buttons below it; null where no ref pane hosts a rebase.
export const RebaseActionContext = createContext<DropdownMenuItem | null>(null);

export const useRebaseAction = () => useContext(RebaseActionContext);
