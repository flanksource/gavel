import { useState } from 'react';
import { Modal, Button } from '@flanksource/clicky-ui/components';
import { UiSearch } from '@flanksource/clicky-ui/icons';

// GrabberExtensionHelp points at the clicky-ui "Playground grabber" Chrome
// extension, which captures UI elements from any page into gavel todos by
// framing this dashboard's /todos/new?embed=1 form.
export function GrabberExtensionHelp() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button
        variant="ghost"
        type="button"
        onClick={() => setOpen(true)}
        title="Capture UI elements into todos"
        className="inline-flex h-8 items-center justify-start gap-1 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted"
      >
        <UiSearch className="text-xs" />
        Grabber
      </Button>
      {open && (
        <Modal
          open
          onClose={() => setOpen(false)}
          title="Capture UI elements into todos"
          size="md"
          footer={
            <div className="flex justify-end">
              <Button variant="outline" onClick={() => setOpen(false)}>Close</Button>
            </div>
          }
        >
          <ol className="list-decimal space-y-2 pl-5 text-sm">
            <li>In a clicky-ui checkout, build the <strong>Playground grabber</strong> Chrome extension: <code>pnpm --filter playground build:extension</code>.</li>
            <li>Open <code>chrome://extensions</code>, enable Developer mode and <strong>Load unpacked</strong> <code>apps/playground/extension/dist</code>.</li>
            <li>In the extension popup, set the backend to <strong>Gavel dashboard</strong> at <code>{window.location.origin}</code> and pick the project.</li>
          </ol>
        </Modal>
      )}
    </>
  );
}
