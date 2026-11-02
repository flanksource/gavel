import { FileFindingsView, type FileFindingsViewProps } from '@flanksource/clicky-ui/file-findings';

// FilesTab is the Files tab body. App loads it with React.lazy so the file
// viewer and its syntax highlighter stay out of the entry bundle until the tab is opened.
export default function FilesTab(props: FileFindingsViewProps) {
  return <FileFindingsView {...props} />;
}
