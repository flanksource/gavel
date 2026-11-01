import { LinterResult } from '@flanksource/clicky-ui/data';
/**
 * Follows the Captain prompt run behind a prompt lint result and shows its
 * transcript as it streams, under a header with the run's status, model,
 * ids and, once it finishes, its duration. The transcript is the run's own
 * `entry` frames, so it shows whether or not Captain records a session. The
 * run's findings reach the lint tree through the run snapshot.
 */
export default function PromptRunView({ result, baseUrl }: {
    result: LinterResult;
    baseUrl: string;
}): import("react").JSX.Element;
//# sourceMappingURL=PromptRunView.d.ts.map