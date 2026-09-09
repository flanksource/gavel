import { Modal, Button } from '@flanksource/clicky-ui/components';
import { useProjectRegistration, ProjectFields } from './ProjectForm';
import type { Project } from '../types';

interface Props {
  open: boolean;
  onClose: () => void;
  onSaved: (project: Project) => void;
  /** Repos to offer in the picker (the currently-known repos). */
  repoOptions: string[];
  defaults?: Partial<Pick<Project, 'name' | 'dir' | 'repos'>>;
}

// AddProjectDialog registers a new local workspace directory (optionally bound to
// repos). Editing and deleting an existing project moved to the settings dialog's
// Project tab; this dialog is add-only. Create POSTs to /api/projects.
export function AddProjectDialog({ open, onClose, onSaved, repoOptions, defaults }: Props) {
  const reg = useProjectRegistration({ open, project: null, defaults });

  if (!open) return null;

  async function save() {
    if (await reg.save()) {
      onSaved({
        name: reg.name.trim(),
        dir: reg.dir.trim(),
        repos: reg.repos,
      });
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
