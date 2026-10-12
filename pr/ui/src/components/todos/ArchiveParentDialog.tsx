import { Button, Modal } from '@flanksource/clicky-ui/components';
import type { TodoChildDisposition, TodoItem } from '../../types';

const LISTED_TITLES = 5;

// Archiving a parent must not close its open children behind the person's back,
// so the choice is theirs: archive them with it, or keep them as full todos.
// Neither is the default — Cancel is the only button that takes focus.
//
// `openChildren` is what the loaded list knows; `serverMessage` is the server's
// refusal when the list could not say (still loading, or failed), in which case
// there is nothing to list.
export function ArchiveParentDialog({ parentTitle, openChildren, serverMessage, onChoose, onCancel }: {
  parentTitle: string;
  openChildren: TodoItem[];
  serverMessage?: string;
  onChoose: (children: TodoChildDisposition) => void;
  onCancel: () => void;
}) {
  const listed = openChildren.slice(0, LISTED_TITLES);
  const more = openChildren.length - listed.length;
  return (
    <Modal
      open
      onClose={onCancel}
      title={`Archive “${parentTitle}”?`}
      size="md"
      footer={
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" autoFocus onClick={onCancel}>Cancel</Button>
          <Button variant="outline" onClick={() => onChoose('detach')}>Make them full todos</Button>
          <Button variant="destructive" onClick={() => onChoose('archive')}>Archive children too</Button>
        </div>
      }
    >
      <div className="space-y-2 text-sm">
        {openChildren.length > 0 && (
          <>
            <p>
              This todo has {openChildren.length} open {openChildren.length === 1 ? 'child' : 'children'}:
            </p>
            <ul className="list-disc space-y-0.5 pl-5">
              {listed.map(child => <li key={child.ref}>{child.title}</li>)}
              {more > 0 && <li className="list-none text-muted-foreground">and {more} more</li>}
            </ul>
          </>
        )}
        {serverMessage && <p className="text-muted-foreground">{serverMessage}</p>}
        <p>Archive them along with it, or keep them as full todos in the listing?</p>
      </div>
    </Modal>
  );
}
