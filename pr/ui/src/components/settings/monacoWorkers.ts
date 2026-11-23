// The web workers Monaco runs a language's services in, for the connection editor's YAML and JSON
// tabs: Vite emits each as its own chunk, loaded only when an editor opens.
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import JsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import YamlWorker from 'monaco-yaml/yaml.worker?worker';

/** The worker for label, the language Monaco asks one for; the editor's own for any other. */
export function monacoWorker(label: string): Worker {
  if (label === 'yaml') return new YamlWorker();
  if (label === 'json') return new JsonWorker();
  return new EditorWorker();
}
