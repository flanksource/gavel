import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";

const INSTALL_COMMAND = "go install github.com/flanksource/gavel/cmd/gavel@latest";

export default function CopyInstall({ className }: { className?: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL_COMMAND);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard API unavailable (http context, permissions) — leave state unchanged
    }
  };

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-md border border-border bg-muted/40 px-4 py-3 font-mono text-sm",
        className,
      )}
    >
      <span className="select-none text-muted-foreground">$</span>
      <code className="flex-1 overflow-x-auto whitespace-nowrap">{INSTALL_COMMAND}</code>
      <button
        type="button"
        onClick={copy}
        aria-label="Copy install command"
        className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground transition-colors hover:text-foreground"
      >
        {copied ? <Check size={14} className="text-brand-sky" /> : <Copy size={14} />}
      </button>
    </div>
  );
}
