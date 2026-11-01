import { FixturePerformance, Test } from '../types';
interface Props {
    performance: FixturePerformance;
    tests: Test[];
    error?: string;
    sampleIndex: number | undefined;
    onSampleChange: (index: number) => void;
}
/** A fixture benchmark or profile run: per-step timings against a baseline, and the sampled process tree. */
export declare function FixtureBenchView({ performance, tests, error, sampleIndex, onSampleChange }: Props): import("react").JSX.Element;
export {};
//# sourceMappingURL=FixtureBenchView.d.ts.map