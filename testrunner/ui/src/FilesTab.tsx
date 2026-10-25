import { FileFindingsView, type FileFindingsViewProps } from '@flanksource/clicky-ui/file-findings';
import { MonacoProvider } from '@flanksource/clicky-ui/monaco';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker&inline';

// The one editor worker serves every language: the Files tab only reads source,
// so the per-language services (ts, json, css, html) are not needed.
const getEditorWorker = () => new EditorWorker();

// FilesTab is the Files tab body. App loads it with React.lazy so Monaco and its
// inline worker stay out of the entry bundle until the tab is opened.
export default function FilesTab(props: FileFindingsViewProps) {
  return (
    <MonacoProvider getWorker={getEditorWorker}>
      <FileFindingsView {...props} />
    </MonacoProvider>
  );
}
