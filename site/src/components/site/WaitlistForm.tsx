import { useState } from "react";
import { Check, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";

type Status = "idle" | "submitting" | "success" | "error";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function WaitlistForm({ className }: { className?: string }) {
  const [email, setEmail] = useState("");
  const [company, setCompany] = useState("");
  const [useCase, setUseCase] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [error, setError] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!EMAIL_RE.test(email)) {
      setError("Please enter a valid email address.");
      return;
    }

    setStatus("submitting");
    try {
      const res = await fetch("/api/waitlist", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ email, company: company || undefined, use_case: useCase || undefined }),
      });

      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as { error?: string };
        throw new Error(body.error ?? `Request failed with ${res.status}`);
      }

      setStatus("success");
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "Something went wrong. Please try again.");
    }
  };

  if (status === "success") {
    return (
      <div
        id="waitlist"
        className={cn(
          "flex flex-col items-center gap-3 rounded-xl border border-brand-sky/40 bg-brand-sky/5 p-10 text-center",
          className,
        )}
      >
        <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-sky text-white">
          <Check size={20} />
        </span>
        <h3 className="text-xl font-semibold">You're on the list</h3>
        <p className="text-sm text-muted-foreground">
          We sent a confirmation to <strong>{email}</strong>. We'll be in touch when the hosted tier opens.
        </p>
      </div>
    );
  }

  return (
    <form
      id="waitlist"
      onSubmit={submit}
      className={cn(
        "grid gap-4 rounded-xl border border-border bg-card p-8 sm:grid-cols-2",
        className,
      )}
      noValidate
    >
      <div className="sm:col-span-2">
        <h3 className="text-xl font-semibold">Join the waitlist</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Early access to the hosted tier, plus release notes for the OSS CLI.
        </p>
      </div>

      <label className="sm:col-span-2">
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Email</span>
        <input
          type="email"
          required
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="mt-1 block h-10 w-full rounded-md border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-brand-sky"
          placeholder="you@company.com"
        />
      </label>

      <label>
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Company (optional)</span>
        <input
          type="text"
          value={company}
          onChange={(e) => setCompany(e.target.value)}
          className="mt-1 block h-10 w-full rounded-md border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-brand-sky"
          placeholder="Flanksource"
        />
      </label>

      <label>
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">Use case (optional)</span>
        <input
          type="text"
          value={useCase}
          onChange={(e) => setUseCase(e.target.value)}
          className="mt-1 block h-10 w-full rounded-md border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-brand-sky"
          placeholder="Monorepo CI, OSS maintainer, …"
        />
      </label>

      {error && (
        <p role="alert" className="sm:col-span-2 text-sm text-destructive">
          {error}
        </p>
      )}

      <button
        type="submit"
        disabled={status === "submitting"}
        className="sm:col-span-2 inline-flex h-11 items-center justify-center gap-2 rounded-md bg-brand-dark px-6 text-sm font-medium text-white transition-colors hover:bg-brand-dark/90 disabled:opacity-60"
      >
        {status === "submitting" && <Loader2 size={16} className="animate-spin" />}
        {status === "submitting" ? "Sending…" : "Notify me"}
      </button>
    </form>
  );
}
