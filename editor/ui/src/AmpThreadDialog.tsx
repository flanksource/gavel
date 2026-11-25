import { useState } from "react";
import { Button, Field, InputField, Modal } from "@flanksource/clicky-ui/components";

export function AmpThreadDialog({ send, close }: { send: (sessionId: string) => Promise<void>; close: () => void }) {
  const [value, setValue] = useState("");
  const [error, setError] = useState<string>();
  const [sending, setSending] = useState(false);
  const sessionId = value.trim().replace(/^https:\/\/ampcode\.com\/threads\//, "");
  const valid = sessionId.length <= 128 && /^T-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$/.test(sessionId);
  const submit = async () => {
    setSending(true);
    setError(undefined);
    try {
      await send(sessionId);
      close();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setSending(false);
    }
  };
  return <Modal open title="Send to Amp thread" onClose={() => { if (!sending) close(); }}
    footer={<div className="flex justify-end gap-2">
      <Button variant="ghost" disabled={sending} onClick={close}>Cancel</Button>
      <Button disabled={!valid || sending} onClick={() => void submit()}>{sending ? "Sending…" : "Send comment"}</Button>
    </div>}>
    <Field label="Amp thread ID or URL" htmlFor="amp-thread-reference" required
      {...(value.trim() && !valid ? { error: "Enter a T-… ID or an ampcode.com thread URL." } : {})}>
      <InputField id="amp-thread-reference" value={value} onChange={setValue} />
    </Field>
    {error && <p role="alert" className="mt-2 text-sm text-destructive">{error}</p>}
  </Modal>;
}
