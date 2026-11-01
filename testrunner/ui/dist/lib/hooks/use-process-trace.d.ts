import { GoProfileRecord, StackRecord } from '@flanksource/clicky-ui/data';
import { ProcessSample, ProcessTraceNode, TracesStatus } from '../types';
export declare const SAMPLE_WINDOW = 2000;
export declare const PROCESS_TRACE_ENDED = "The run has ended, so its process trace can no longer be checkpointed";
export interface UseProcessTraceOptions {
    traces?: TracesStatus;
    done: boolean;
    selectedKey?: string;
    base?: string;
}
export interface ProcessTraceState {
    nodes: ProcessTraceNode[];
    samples: ProcessSample[];
    stacks: StackRecord[];
    goprof: GoProfileRecord[];
    goprofExportUrl?: string;
    loading: boolean;
    error?: string;
    canCollect: boolean;
    collecting: boolean;
    collectError?: string;
    collectStack: (key: string, transport: string) => Promise<void>;
    refreshing: boolean;
    refreshError?: string;
    refreshFacets: (key: string) => Promise<void>;
}
export declare function useProcessTrace({ traces: reported, done, selectedKey, base }: UseProcessTraceOptions): ProcessTraceState;
//# sourceMappingURL=use-process-trace.d.ts.map