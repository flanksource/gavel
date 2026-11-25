import { createContext, useContext } from 'react';

// The phone's ref bar offers a spot for a finished commit run's status chip so
// the run needs no row of its own; null where no ref bar hosts one.
export const CommitStatusSlotContext = createContext<HTMLElement | null>(null);

export const useCommitStatusSlot = () => useContext(CommitStatusSlotContext);
