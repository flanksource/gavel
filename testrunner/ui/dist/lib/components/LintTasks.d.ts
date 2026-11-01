import { Test } from '../types';
/**
 * The live clicky tasks behind a lint run, such as the per-review prompt tasks
 * under the `lint prompt` group, with each task's own status text (thinking,
 * tool: read_file). Renders nothing while no task is running or pending.
 */
export declare function LintTasks({ tests }: {
    tests: Test[];
}): import("react").JSX.Element | null;
//# sourceMappingURL=LintTasks.d.ts.map