import { LintNode, LinterResult } from '@flanksource/clicky-ui/data';
import { RerunRequest } from './types';
/**
 * The rerun request that re-lints what a lint tree node covers: its linter, the
 * files beneath it that still have findings, and the single work dir they share.
 * Without a node it reruns every linter. `work_dir` is empty when the scope
 * spans several lint roots, which the server cannot rerun as one.
 */
export declare function buildLintRerunBody(node: LintNode | null, lint: readonly LinterResult[]): RerunRequest;
//# sourceMappingURL=lintRerun.d.ts.map