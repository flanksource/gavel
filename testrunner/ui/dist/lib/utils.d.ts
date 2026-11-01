import { LinterResult, Test } from './types';
export declare const TEST_TASK_GROUP_NAME = "Running tests across packages";
export declare const LINT_TASK_GROUP_NAME = "Running linters";
export declare function stripTestTaskGroup(tests: Test[]): Test[];
export declare function isLintOnlyPhase(tests: Test[], running: boolean): boolean;
export declare function isLintOnlyRun(tests: Test[], lintRun: boolean): boolean;
export interface LintStatusCounts {
    running: number;
    done: number;
    failed: number;
}
export declare function lintStatusCounts(lint: LinterResult[] | undefined): LintStatusCounts;
export declare function liveTaskGroups(tests: Test[]): Test[];
export declare function timeoutArgValue(args: Record<string, unknown> | undefined, key: string): string | null;
export interface ProcessStepLink {
    pid: number;
    routePath: string;
    name: string;
}
export declare function processLinksByPid(tests: Test[]): Map<number, ProcessStepLink>;
//# sourceMappingURL=utils.d.ts.map