import { useRef, useState } from 'react';
import { SegmentedControl } from '@flanksource/clicky-ui/components';

// A prompt override is either inline template text or a path to a .prompt file
// (the union the Go PromptOverride marshals). An unset/empty override means the
// built-in default is used.
export type PromptOverrideValue = string | { inline?: string; file?: string };

interface Props {
  value: PromptOverrideValue | undefined;
  onChange: (next: PromptOverrideValue | undefined) => void;
  /** Embedded built-in template — the prompt in use when there is no override. */
  defaultText: string;
  description?: string;
}

function normalize(v: PromptOverrideValue | undefined): { inline: string; file: string } {
  if (typeof v === 'string') return { inline: v, file: '' };
  if (v && typeof v === 'object') return { inline: v.inline ?? '', file: v.file ?? '' };
  return { inline: '', file: '' };
}

const SOURCES = [
  { id: 'inline', label: 'Inline' },
  { id: 'file', label: 'File' },
] as const;

// PromptOverrideField edits one prompt override with a segmented Inline / File
// source toggle. In Inline mode the textarea is pre-filled with the prompt
// actually in use — the override when set, otherwise the built-in default — so
// the form shows the effective value and you edit it directly. Editing it back to
// the default (or clearing it) emits `undefined`, so the default is never
// redundantly written into the .gavel.yaml file. The inactive source's text is
// stashed locally so toggling between Inline and File does not lose edits.
export function PromptOverrideField({ value, onChange, defaultText, description }: Props) {
  const { inline, file } = normalize(value);
  const [mode, setMode] = useState<'inline' | 'file'>(file ? 'file' : 'inline');
  const stash = useRef({ inline, file });
  if (mode === 'inline') stash.current.inline = inline;
  else stash.current.file = file;

  // The textarea shows the effective inline prompt: the override when set,
  // otherwise the built-in default that is in use.
  const inlineText = inline !== '' ? inline : defaultText;
  const isDefaulted = file === '' && (inline === '' || inline === defaultText);

  // commitInline persists an override only when the text differs from the
  // default; equal-to-default or empty clears it back to the built-in.
  function commitInline(next: string) {
    stash.current.inline = next;
    if (next.trim() === '' || next === defaultText) onChange(undefined);
    else onChange({ inline: next });
  }

  function pickSource(next: 'inline' | 'file') {
    setMode(next);
    if (next === 'inline') {
      const v = stash.current.inline;
      onChange(v && v !== defaultText ? { inline: v } : undefined);
    } else {
      onChange(stash.current.file ? { file: stash.current.file } : undefined);
    }
  }

  return (
    <div className="space-y-2">
      {description && <p className="text-xs text-muted-foreground">{description}</p>}

      <div className="flex items-center gap-2">
        <SegmentedControl
          size="sm"
          value={mode}
          options={SOURCES.map(s => ({ id: s.id, label: s.label }))}
          onChange={id => pickSource(id as 'inline' | 'file')}
          aria-label="Prompt source"
        />
        {isDefaulted ? (
          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
            Default
          </span>
        ) : (
          <button
            type="button"
            onClick={() => onChange(undefined)}
            className="text-xs text-muted-foreground underline hover:text-foreground"
          >
            Reset to default
          </button>
        )}
      </div>

      {mode === 'inline' ? (
        <textarea
          className="min-h-[10rem] w-full rounded-md border border-border bg-background p-2 font-mono text-xs"
          value={inlineText}
          spellCheck={false}
          onChange={e => commitInline(e.target.value)}
        />
      ) : (
        <div className="space-y-1">
          <input
            type="text"
            className="w-full rounded-md border border-border bg-background p-2 font-mono text-xs"
            value={file}
            placeholder="./prompts/my-prompt.prompt"
            spellCheck={false}
            onChange={e => {
              const next = e.target.value;
              stash.current.file = next;
              onChange(next ? { file: next } : undefined);
            }}
          />
          <p className="text-[11px] text-muted-foreground">
            Relative paths resolve against the .gavel.yaml directory. Leave empty to use the built-in default.
          </p>
        </div>
      )}
    </div>
  );
}
