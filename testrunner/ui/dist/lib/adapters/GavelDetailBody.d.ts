import { Test as ClickyTest } from '@flanksource/clicky-ui/data';
/**
 * The detail body of every node that is not a fixture step (fixture steps use
 * clicky's verification adapter): suite, package and command, the error, the
 * framework's own context (Go subtest parent, Ginkgo suite/labels/report
 * entries), the latest attempt's process metrics and stack, output and the
 * provider's structured detail.
 */
export declare function GavelDetailBody({ node: clickyNode }: {
    node: ClickyTest;
}): import("react").JSX.Element;
//# sourceMappingURL=GavelDetailBody.d.ts.map