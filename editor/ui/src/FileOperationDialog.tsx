import { useEffect, useState, type ReactNode } from "react";
import { Button, Field, InputField, Modal } from "@flanksource/clicky-ui/components";
import { request, type FileEntry } from "./api";
import { joinPath, retargetName, templateGroups, type Template } from "./file-operations";

/** The file operation in progress; paths are relative to the tree root. */
export type FileOperation =
  | { kind: "new"; folder: string }
  | { kind: "rename"; entry: FileEntry; path: string }
  | { kind: "delete"; entry: FileEntry; path: string };

export type FileOperationHandlers = {
  onCreate: (path: string, template: string) => Promise<void>;
  onRename: (entry: FileEntry, path: string) => Promise<void>;
  onDelete: (entry: FileEntry) => Promise<void>;
  onClose: () => void;
};

/** New file (from a template), rename/move, and delete-confirm, as one dialog per operation. */
export function FileOperationDialog({ operation, ...handlers }: { operation: FileOperation } & FileOperationHandlers) {
  const [error, setError] = useState<string>();
  const [busy, setBusy] = useState(false);
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    setError(undefined);
    try {
      await action();
    } catch (cause) {
      setError(String(cause));
    } finally {
      setBusy(false);
    }
  };
  const alert = error && (
    <p role="alert" className="text-sm text-destructive">
      {error}
    </p>
  );
  if (operation.kind === "new") return <NewFile folder={operation.folder} busy={busy} alert={alert} run={run} {...handlers} />;
  if (operation.kind === "rename") return <RenameFile operation={operation} busy={busy} alert={alert} run={run} {...handlers} />;
  return (
    <Modal
      open
      onClose={handlers.onClose}
      title="Delete file"
      size="sm"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={handlers.onClose}>
            Cancel
          </Button>
          <Button variant="destructive" disabled={busy} onClick={() => void run(() => handlers.onDelete(operation.entry))}>
            {busy ? "Deleting…" : "Delete"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-3 text-sm">
        <p>
          Delete <span className="font-mono">{operation.path}</span> from disk? Its comment threads stay on the old path.
        </p>
        {alert}
      </div>
    </Modal>
  );
}

type StepProps = FileOperationHandlers & {
  busy: boolean;
  alert: ReactNode;
  run: (action: () => Promise<void>) => Promise<void>;
};

function NewFile({ folder: initialFolder, busy, alert, run, onCreate, onClose }: StepProps & { folder: string }) {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loadError, setLoadError] = useState<string>();
  const [chosen, setChosen] = useState<Template>();
  const [folder, setFolder] = useState(initialFolder);
  const [name, setName] = useState("");
  useEffect(() => {
    void request<Template[]>("/api/templates")
      .then((next) => {
        setTemplates(next);
        if (next[0]) choose(next[0], undefined, "");
      })
      .catch((cause) => setLoadError(String(cause)));
  }, []);
  const choose = (next: Template, previous = chosen, current = name) => {
    setChosen(next);
    setName(retargetName(current, previous, next));
  };
  const valid = chosen && name.trim() !== "" && name.toLowerCase().endsWith(chosen.suffix);
  return (
    <Modal
      open
      onClose={onClose}
      title="New file"
      size="lg"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button disabled={!valid || busy} onClick={() => void run(() => onCreate(joinPath(folder.trim(), name.trim()), chosen!.id))}>
            {busy ? "Creating…" : "Create"}
          </Button>
        </div>
      }
    >
      <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div role="radiogroup" aria-label="Template" className="flex max-h-96 flex-col gap-3 overflow-auto">
          {loadError && (
            <p role="alert" className="text-sm text-destructive">
              Templates unavailable: {loadError}
            </p>
          )}
          {templateGroups(templates).map(([group, items]) => (
            <div key={group} className="flex flex-col gap-1">
              <p className="text-xs font-semibold uppercase text-muted-foreground">{group}</p>
              {items.map((template) => (
                <button
                  key={template.id}
                  type="button"
                  role="radio"
                  aria-checked={chosen?.id === template.id}
                  onClick={() => choose(template)}
                  className={`rounded-md border px-3 py-2 text-left text-sm ${chosen?.id === template.id ? "border-primary bg-primary/5" : "border-border hover:bg-muted"}`}
                >
                  <span className="block font-medium">{template.label}</span>
                  <span className="block text-xs text-muted-foreground">{template.description}</span>
                </button>
              ))}
            </div>
          ))}
        </div>
        <div className="flex min-w-0 flex-col gap-3">
          <Field label="Folder" htmlFor="new-file-folder" helper="Relative to the edited files; missing folders are created.">
            <InputField id="new-file-folder" value={folder} onChange={setFolder} />
          </Field>
          <Field label="Name" htmlFor="new-file-name" required helper={chosen ? `Must end in ${chosen.suffix}` : undefined}>
            <InputField id="new-file-name" value={name} onChange={setName} />
          </Field>
          {chosen && (
            <pre aria-label="Template preview" className="max-h-56 overflow-auto rounded-md bg-muted p-2 font-mono text-xs">
              {chosen.source}
            </pre>
          )}
          {alert}
        </div>
      </div>
    </Modal>
  );
}

function RenameFile({ operation, busy, alert, run, onRename, onClose }: StepProps & { operation: Extract<FileOperation, { kind: "rename" }> }) {
  const [path, setPath] = useState(operation.path);
  const changed = path.trim() !== "" && path.trim() !== operation.path;
  return (
    <Modal
      open
      onClose={onClose}
      title="Rename file"
      size="md"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button disabled={!changed || busy} onClick={() => void run(() => onRename(operation.entry, path.trim()))}>
            {busy ? "Renaming…" : "Rename"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        <Field label="Path" htmlFor="rename-file-path" required helper="Relative to the edited files; include folders to move it. Comment threads stay on the old path.">
          <InputField id="rename-file-path" value={path} onChange={setPath} />
        </Field>
        {alert}
      </div>
    </Modal>
  );
}
