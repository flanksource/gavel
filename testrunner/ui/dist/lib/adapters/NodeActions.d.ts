import { RouteState } from '../routes';
import { Test, TestEditAction, TestEditScope } from '../types';
export interface NodeActionsConfig {
    /** The route state that exports one node (the view's filters with the node's path), if the node can be exported. */
    nodeRouteState: (node: Test) => RouteState | undefined;
    /** The same, narrowed to failing and timed-out tests, for the AI prompt. */
    failingOnlyRouteState: (node: Test) => RouteState | undefined;
    /** status.test_edit_supported: whether the server can edit test sources. */
    testEditSupported: boolean;
    onTestEdit: (node: Test, action: TestEditAction, scope: TestEditScope) => Promise<void>;
    testEditBusy: boolean;
}
export declare function canEditSource(node: Test, supported: boolean): boolean;
/**
 * The header actions of a test node: download and AI-prompt copy of its export,
 * and, for a Go, Ginkgo or vitest test with a source file, the Skip / Delete
 * source edits.
 */
export declare function NodeActions({ node, config }: {
    node: Test;
    config: NodeActionsConfig;
}): import("react").JSX.Element;
//# sourceMappingURL=NodeActions.d.ts.map