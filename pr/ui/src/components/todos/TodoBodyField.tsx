import { useRef } from 'react';
import { Field } from '@flanksource/clicky-ui/components';
import { inputClass } from './format';
import { insertTodoBodyPaste, normalizeTodoBodyPaste } from './todoBodyPaste';

export interface TodoBodyFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  disabled?: boolean;
}

export function TodoBodyField({ label, value, onChange, placeholder, disabled }: TodoBodyFieldProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  return (
    <Field
      label={label}
      helper="Pasted HTML becomes Markdown; multiline plain text becomes a code block."
    >
      <textarea
        ref={textareaRef}
        aria-label={label}
        className={`${inputClass} h-64 resize-y font-mono`}
        value={value}
        placeholder={placeholder}
        disabled={disabled}
        onChange={event => onChange(event.currentTarget.value)}
        onPaste={event => {
          const paste = normalizeTodoBodyPaste(event.clipboardData);
          if (!paste) return;
          event.preventDefault();
          const insertion = insertTodoBodyPaste(
            value,
            event.currentTarget.selectionStart,
            event.currentTarget.selectionEnd,
            paste,
          );
          onChange(insertion.value);
          queueMicrotask(() => textareaRef.current?.setSelectionRange(insertion.caret, insertion.caret));
        }}
      />
    </Field>
  );
}
