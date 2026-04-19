import type { ReactNode } from "react";
import { Bot, GitPullRequest, ServerCog, ShieldCheck } from "lucide-react";
import NumberTicker from "@/magic/NumberTicker";
import TypingAnimation from "@/magic/TypingAnimation";
import { cn } from "@/lib/utils";

interface Feature {
  icon: ReactNode;
  title: string;
  hook: string;
  bullets: string[];
  span?: string;
}

const features: Feature[] = [
  {
    icon: <Bot size={18} />,
    title: "Testing + AI verify",
    hook: "Go & Ginkgo with AI-assisted review baked in",
    bullets: [
      "Structured JSON + HTML output with live UI",
      "Fixture-driven tests from markdown files",
      "Benchmark diffing across commits",
      "`gavel verify` surfaces AI-graded checks",
    ],
    span: "sm:col-span-2",
  },
  {
    icon: <ShieldCheck size={18} />,
    title: "Unified linting",
    hook: "Nine linters, one command, parallel by default",
    bullets: [
      "golangci-lint, ruff, pyright, eslint, tsc",
      "markdownlint, vale, jscpd, betterleaks",
      "Config-file auto-discovery per tool",
      "One HTML + JSON report for CI",
    ],
  },
  {
    icon: <GitPullRequest size={18} />,
    title: "PR workflow",
    hook: "From commit message to merge, automated",
    bullets: [
      "AI-generated conventional commits",
      "Pre-commit hooks that re-run verify",
      "`gavel pr fix` applies review suggestions",
      "Git history analysis for code review",
    ],
  },
  {
    icon: <ServerCog size={18} />,
    title: "CI integration",
    hook: "Built for GitHub Actions from day one",
    bullets: [
      "Ships as a first-class GitHub Action",
      "Markdown summary sized for PR comments",
      "JSON artifact for downstream tooling",
      "SSH server mode for local CI-like loops",
    ],
    span: "sm:col-span-2",
  },
];

function FeatureCard({ feature }: { feature: Feature }) {
  return (
    <div
      className={cn(
        "group flex flex-col rounded-xl border border-border bg-card p-6 transition-colors hover:border-brand-sky/40",
        feature.span,
      )}
    >
      <div className="flex items-center gap-3">
        <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-brand-sky/10 text-brand-sky">
          {feature.icon}
        </span>
        <h3 className="text-lg font-semibold">{feature.title}</h3>
      </div>
      <p className="mt-3 text-sm text-muted-foreground">{feature.hook}</p>
      <ul className="mt-4 space-y-2 text-sm text-foreground/80">
        {feature.bullets.map((b) => (
          <li key={b} className="flex gap-2 before:mt-2 before:h-1 before:w-1 before:shrink-0 before:rounded-full before:bg-brand-sky">
            <span>{b}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export default function FeatureGrid() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-16">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">What ships</p>
          <h2 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">
            Four pillars, one binary
          </h2>
        </div>
        <p className="max-w-sm text-sm text-muted-foreground">
          <NumberTicker value={9} className="text-3xl font-bold text-foreground" suffix="+" />
          <span className="ml-2">linters unified behind a single command</span>
        </p>
      </div>

      <div className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((f) => (
          <FeatureCard key={f.title} feature={f} />
        ))}
      </div>

      <div className="mt-12 rounded-xl border border-border bg-muted/30 p-6 font-mono text-sm">
        <TypingAnimation text="$ gavel test --lint --verify  →  tests + 9 linters + AI review in one run" />
      </div>
    </section>
  );
}
