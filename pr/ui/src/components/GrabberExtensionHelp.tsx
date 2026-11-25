import { Modal, Button } from '@flanksource/clicky-ui/components';

export const GRABBER_HELP_TITLE = 'Capture UI elements into todos';

// GrabberExtensionHelpModal points at the clicky-ui "Playground grabber" Chrome
// extension, which captures UI elements from any page into gavel todos by
// framing this dashboard's /todos/new?embed=1 form. The user menu opens it.
export function GrabberExtensionHelpModal({ onClose }: { onClose: () => void }) {
  return (
    <Modal
      open
      onClose={onClose}
      title={GRABBER_HELP_TITLE}
      size="md"
      footer={
        <div className="flex justify-end">
          <Button variant="outline" onClick={onClose}>Close</Button>
        </div>
      }
    >
      <ol className="list-decimal space-y-2 pl-5 text-sm">
        <li>In a clicky-ui checkout, build the <strong>Playground grabber</strong> Chrome extension: <code>pnpm --filter playground build:extension</code>.</li>
        <li>Open <code>chrome://extensions</code>, enable Developer mode and <strong>Load unpacked</strong> <code>apps/playground/extension/dist</code>.</li>
        <li>In the extension popup, set the backend to <strong>Gavel dashboard</strong> at <code>{window.location.origin}</code> and pick the project.</li>
      </ol>
    </Modal>
  );
}
