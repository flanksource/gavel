import { useState } from 'react';
import { Modal, Button } from '@flanksource/clicky-ui/components';
import { GavelIcon } from './GavelIcon';

// ReactGrabHelp surfaces how to capture UI elements into gavel todos with React
// Grab. The bookmarklet and console snippet both load this gavel server's plugin
// (window.location.origin = the dashboard's own origin), so they keep targeting
// gavel even when injected into another app's dev page.
export function ReactGrabHelp() {
  const [open, setOpen] = useState(false);
  const src = `${window.location.origin}/react-grab-plugin.js`;
  const bookmarklet = `javascript:(function(){var s=document.createElement('script');s.src='${src}';document.body.appendChild(s);})();`;
  const snippet = `var s=document.createElement('script');s.src='${src}';document.body.appendChild(s);`;

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        title="Install React Grab → todo"
        className="inline-flex h-8 items-center gap-1 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted"
      >
        <GavelIcon name="codicon:inspect" className="text-xs" />
        React Grab
      </button>
      {open && (
        <Modal
          open
          onClose={() => setOpen(false)}
          title="React Grab → todo"
          size="md"
          footer={
            <div className="flex justify-end">
              <Button variant="outline" onClick={() => setOpen(false)}>Close</Button>
            </div>
          }
        >
          <div className="space-y-4 text-sm">
            <p className="text-muted-foreground">
              Capture any UI element into a gavel todo. Grab an element with{' '}
              <a className="underline" href="https://github.com/aidenybai/react-grab" target="_blank" rel="noreferrer">
                React Grab
              </a>{' '}
              and run the <strong>Add to gavel todo</strong> action — a dialog opens with this gavel's
              new-todo form prefilled from the component and its source.
            </p>

            <div>
              <div className="mb-1 font-medium">Bookmarklet</div>
              <p className="mb-2 text-muted-foreground">Drag to your bookmarks bar, then click it on any running dev page:</p>
              {/* React blocks `javascript:` in the href prop, so set it on the
                  DOM node directly — the browser still reads it when the link is
                  dragged to the bookmarks bar. */}
              <a
                ref={node => node?.setAttribute('href', bookmarklet)}
                draggable
                onClick={e => e.preventDefault()}
                className="inline-flex items-center gap-1 rounded-md bg-foreground px-3 py-1.5 font-medium text-background"
              >
                <GavelIcon name="codicon:inspect" className="text-xs" />
                Add to gavel todo
              </a>
            </div>

            <div>
              <div className="mb-1 font-medium">Or paste in the console</div>
              <pre className="overflow-x-auto rounded-md bg-black p-3 text-xs text-green-200">
                <code>{snippet}</code>
              </pre>
            </div>

            <p className="text-muted-foreground">
              Developing with <code>gavel pr list --dev</code>? It's already loaded — just grab and run the action.
            </p>
          </div>
        </Modal>
      )}
    </>
  );
}
