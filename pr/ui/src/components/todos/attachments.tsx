import { useCallback, useEffect, useRef, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { GavelIcon } from '../GavelIcon';

// Attachment is one image picked for a new todo: the raw bytes plus a filename
// the server stores and embeds into the todo body.
export interface Attachment {
  blob: Blob;
  name: string;
}

export interface AttachmentPreview {
  name: string;
  url: string;
}

// defaultAttachmentName names a pasted/dropped image that arrives without a
// filename (clipboard images usually do), deriving the extension from its MIME
// type so the stored file keeps a sensible suffix.
function defaultAttachmentName(type: string): string {
  const ext = type.split('/')[1] || 'png';
  return `screenshot.${ext}`;
}

// imageAttachmentsFromDataTransfer extracts image attachments from a clipboard or
// drag-and-drop payload. It prefers `files` (drops and modern clipboard pastes)
// and falls back to `items` for browsers that only expose pasted images there;
// non-image payloads (e.g. pasted text) yield nothing so the caller can let the
// default paste/drop happen.
export function imageAttachmentsFromDataTransfer(data: DataTransfer | null): Attachment[] {
  if (!data) return [];
  const fromFiles: Attachment[] = [];
  for (const file of Array.from(data.files)) {
    if (file.type.startsWith('image/')) {
      fromFiles.push({ blob: file, name: file.name || defaultAttachmentName(file.type) });
    }
  }
  if (fromFiles.length) return fromFiles;
  const fromItems: Attachment[] = [];
  for (const item of Array.from(data.items)) {
    if (item.kind === 'file' && item.type.startsWith('image/')) {
      const file = item.getAsFile();
      if (file) fromItems.push({ blob: file, name: file.name || defaultAttachmentName(file.type) });
    }
  }
  return fromItems;
}

// imageAttachmentsFromFiles keeps only the images from a file-input selection.
export function imageAttachmentsFromFiles(files: FileList | null): Attachment[] {
  if (!files) return [];
  const out: Attachment[] = [];
  for (const file of Array.from(files)) {
    if (file.type.startsWith('image/')) {
      out.push({ blob: file, name: file.name || defaultAttachmentName(file.type) });
    }
  }
  return out;
}

// todoFormData builds the multipart body a create-with-attachments POST sends to
// /api/todos/new. Image bytes ride along under the `attachment` field, which the
// server persists and embeds into the todo body. The browser sets the multipart
// Content-Type (with boundary) for a FormData body.
export function todoFormData(
  fields: { title: string; body: string; priority: string; status: string; autoSave?: boolean },
  attachments: Attachment[],
): FormData {
  const form = new FormData();
  form.append('title', fields.title);
  form.append('body', fields.body);
  form.append('priority', fields.priority);
  form.append('status', fields.status);
  if (fields.autoSave !== undefined) form.append('autoSave', String(fields.autoSave));
  attachments.forEach(a => form.append('attachment', a.blob, a.name));
  return form;
}

export interface UseAttachments {
  attachments: Attachment[];
  previews: AttachmentPreview[];
  add: (items: Attachment[]) => void;
  remove: (index: number) => void;
  clear: () => void;
}

// useAttachments owns the new-todo screenshot set: it derives revocable object
// URLs for thumbnails and, when `pasteAnywhere` is set, captures an image pasted
// anywhere on the form. Only image payloads are captured, so pasting text into a
// field still pastes normally. `pasteAnywhere` is gated (e.g. on a modal's open
// flag) so an always-mounted form doesn't hijack paste while it's hidden.
export function useAttachments({ pasteAnywhere = false }: { pasteAnywhere?: boolean } = {}): UseAttachments {
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [previews, setPreviews] = useState<AttachmentPreview[]>([]);

  useEffect(() => {
    const urls = attachments.map(a => ({ name: a.name, url: URL.createObjectURL(a.blob) }));
    setPreviews(urls);
    return () => urls.forEach(u => URL.revokeObjectURL(u.url));
  }, [attachments]);

  const add = useCallback((items: Attachment[]) => {
    if (items.length) setAttachments(prev => [...prev, ...items]);
  }, []);
  const remove = useCallback((index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index));
  }, []);
  const clear = useCallback(() => setAttachments([]), []);

  useEffect(() => {
    if (!pasteAnywhere) return;
    function onPaste(e: ClipboardEvent) {
      const items = imageAttachmentsFromDataTransfer(e.clipboardData);
      if (items.length) {
        e.preventDefault();
        add(items);
      }
    }
    window.addEventListener('paste', onPaste);
    return () => window.removeEventListener('paste', onPaste);
  }, [add, pasteAnywhere]);

  return { attachments, previews, add, remove, clear };
}

// ScreenshotPicker is the new-todo form's image attachment control: a dashed drop
// zone that also opens a file picker on click, plus a thumbnail strip with a
// per-image remove button. Pasting is handled form-wide by useAttachments.
export function ScreenshotPicker({ previews, onAdd, onRemove, disabled }: {
  previews: AttachmentPreview[];
  onAdd: (items: Attachment[]) => void;
  onRemove: (index: number) => void;
  disabled?: boolean;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);

  return (
    <div className="space-y-2">
      <div
        role="button"
        tabIndex={0}
        aria-label="Attach screenshot"
        onClick={() => inputRef.current?.click()}
        onKeyDown={e => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            inputRef.current?.click();
          }
        }}
        onDragOver={e => {
          e.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={e => {
          e.preventDefault();
          setDragging(false);
          onAdd(imageAttachmentsFromDataTransfer(e.dataTransfer));
        }}
        className={`flex cursor-pointer flex-col items-center justify-center gap-1 rounded-md border border-dashed px-3 py-4 text-center text-xs text-muted-foreground transition-colors hover:bg-muted ${
          dragging ? 'border-primary bg-primary/5' : 'border-border'
        }`}
      >
        <GavelIcon name="codicon:device-camera" className="text-base" />
        <span>
          <span className="font-medium text-primary underline">Choose an image</span>, or drag, drop, or paste a screenshot
        </span>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          multiple
          disabled={disabled}
          className="hidden"
          onChange={e => {
            onAdd(imageAttachmentsFromFiles(e.currentTarget.files));
            e.currentTarget.value = '';
          }}
        />
      </div>
      {previews.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {previews.map((p, i) => (
            <div key={p.url} className="relative">
              <img src={p.url} alt={p.name} className="max-h-32 rounded border border-border" />
              <Button
                variant="ghost"
                size="icon"
                type="button"
                onClick={() => onRemove(i)}
                title="Remove screenshot"
                aria-label={`Remove ${p.name}`}
                className="absolute -right-2 -top-2 h-5 w-5 rounded-full border border-border bg-background text-muted-foreground shadow hover:text-destructive"
              >
                <GavelIcon name="codicon:close" className="text-[11px]" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
