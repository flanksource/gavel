import type { ReactNode } from 'react';
import { Button, SplitPane } from '@flanksource/clicky-ui/components';
import { UiArrowLeft, UiChevronRight } from '@flanksource/clicky-ui/icons';
import { useIsMobile } from '../useIsMobile';
import type { ProjectFileStatus } from './ProjectStatusView';
import { ProjectLinkedDepsBadge } from './ProjectLinkedDepsBadge';
import { stageBlockReason } from './projectStage';

export interface ProjectChangesViewProps {
  files: ProjectFileStatus[];
  locked: Map<string, number>;
  disabled: boolean;
  diffPath: string;
  onDiffPathChange: (path: string) => void;
  // onToggleStaged stages the file, or unstages it when it is already fully staged.
  onToggleStaged: (file: ProjectFileStatus) => void;
  fileList: ReactNode;
  diff: ReactNode;
}

export function ProjectChangesView({ files, locked, disabled, diffPath, onDiffPathChange, onToggleStaged, fileList, diff }: ProjectChangesViewProps) {
  const mobile = useIsMobile();
  if (!mobile) return <SplitPane className="min-h-0 flex-1" defaultSplit={55} minLeft={35} minRight={30} leftClass="overflow-hidden" left={fileList} right={diff} />;
  if (!diffPath) return <div className="min-h-0 flex-1 overflow-hidden">{fileList}</div>;
  const ordered = [...files].sort((a, b) => a.path.localeCompare(b.path));
  const index = ordered.findIndex(file => file.path === diffPath);
  const file = ordered[index];
  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      <div className="shrink-0 border-b border-border px-3 py-1">
        <nav aria-label="Review changed files" className="flex items-center gap-1">
          <Button type="button" variant="ghost" aria-label="Back to files" onClick={() => onDiffPathChange('')} className="min-h-11 gap-1 px-2"><UiArrowLeft />Files</Button>
          <span role="status" className="ml-auto text-xs tabular-nums text-muted-foreground">{file ? `${index + 1} of ${ordered.length}` : 'Folder diff'}</span>
          <Button type="button" variant="ghost" size="icon" aria-label="Previous file" disabled={index <= 0} onClick={() => onDiffPathChange(ordered[index - 1].path)} className="min-h-11 min-w-11"><UiChevronRight className="rotate-180" /></Button>
          <Button type="button" variant="ghost" size="icon" aria-label="Next file" disabled={index >= ordered.length - 1} onClick={() => onDiffPathChange(ordered[index + 1].path)} className="min-h-11 min-w-11"><UiChevronRight /></Button>
        </nav>
        {file && <StageToggle file={file} blockReason={stageBlockReason({ file, lockedGroup: locked.get(file.path) })} disabled={disabled} onToggleStaged={onToggleStaged} />}
      </div>
      <div className="min-h-0 min-w-0 flex-1 overflow-hidden">{diff}</div>
    </div>
  );
}

function StageToggle({ file, blockReason, disabled, onToggleStaged }: {
  file: ProjectFileStatus;
  blockReason: string | null;
  disabled: boolean;
  onToggleStaged: (file: ProjectFileStatus) => void;
}) {
  return (
    <label className="flex min-h-11 items-center gap-2 text-sm">
      <input
        type="checkbox"
        aria-label={`Stage ${file.path}`}
        checked={file.state === 'staged'}
        ref={input => { if (input) input.indeterminate = file.state === 'both'; }}
        disabled={disabled || blockReason !== null}
        onChange={() => onToggleStaged(file)}
      />
      Staged
      <ProjectLinkedDepsBadge file={file} />
      {blockReason && <span className="text-xs text-muted-foreground">{blockReason}</span>}
    </label>
  );
}
