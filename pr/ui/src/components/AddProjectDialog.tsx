import { Modal, Button } from '@flanksource/clicky-ui/components';
import { useProjectRegistration, ProjectFields } from './ProjectForm';

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
  /** Repos to offer in the picker (the currently-known repos). */
  repoOptions: string[];
}

// AddProjectDialog registers a new local workspace directory (optionally bound to
// repos). Editing and deleting an existing project moved to the settings dialog's
// Project tab; this dialog is add-only. Create POSTs to /api/projects.
export function AddProjectDialog({ open, onClose, onSaved, repoOptions }: Props) {
  const reg = useProjectRegistration(open, null);

  if (!open) return null;

  async function save() {
    if (await reg.save()) {
      onSaved();
      onClose();
    }
  }

  return (
    <Modal
      open
      onClose={onClose}
      title="Add local directory"
      size="md"
      footer={
        <div className="flex items-center justify-end gap-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={save} loading={reg.saving}>Save</Button>
        </div>
      }
    >
      <ProjectFields reg={reg} repoOptions={repoOptions} />
    </Modal>
  );
}
