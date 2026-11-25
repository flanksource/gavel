import { useId, useState } from "react";
import { Button, Field, InputField } from "@flanksource/clicky-ui/components";
import { ampThreadId } from "./session-urls";

export function AmpThreadField({ onChoose, label }: { onChoose: (id: string) => void; label: string }) {
  const id = useId();
  const [value, setValue] = useState("");
  const thread = ampThreadId(value);
  const invalid = value.trim() !== "" && !thread;
  return <div className="space-y-2">
    <Field label="Amp thread ID or URL" htmlFor={id} required {...(invalid ? { error: "Enter a T-… ID or an ampcode.com thread URL." } : {})}>
      <InputField id={id} value={value} onChange={setValue} />
    </Field>
    <Button type="button" size="sm" variant="outline" disabled={!thread} onClick={() => { if (thread) onChoose(thread); }}>{label}</Button>
  </div>;
}
