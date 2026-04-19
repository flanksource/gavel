import type { LucideIcon } from "lucide-react";
import {
  Boxes,
  CircleDot,
  Cloud,
  FileSearch,
  GaugeCircle,
  GitCommitHorizontal,
  GitPullRequest,
  KeyRound,
  MonitorDot,
  ScanSearch,
  Settings2,
  Sparkles,
  Type as TypeIcon,
  Zap,
} from "lucide-react";
import AnimatedBeam from "@/magic/AnimatedBeam";
import IconOrbit from "@/magic/IconOrbit";
import PulseNodes from "@/magic/PulseNodes";
import TypingAnimation from "@/magic/TypingAnimation";
import { cn } from "@/lib/utils";

type AnimationKind = "typing" | "icon-orbit" | "pulse-nodes" | "beam-connector";

interface Feature {
  slug: string;
  pillar_title: string;
  hook: string;
  body: string;
  bullets: string[];
  uniqueness: string;
  animation_kind: AnimationKind;
  primary_icon: LucideIcon;
}

const features: Feature[] = [
  {
    slug: "incremental",
    pillar_title: "Incremental Test & Lint",
    hook: "Run only what your branch touched",
    body: "Gavel resolves your working tree against origin/main, walks the package graph, and runs only affected packages. Content fingerprints skip anything that already passed.",
    bullets: [
      "--changed against any ref",
      "--cache on green fingerprint",
      "Same engine for lint",
      "Drops to full suite in CI",
    ],
    uniqueness:
      "`gavel test --changed --cache` gives sub-second feedback on monorepos without sacrificing the full CI run.",
    animation_kind: "typing",
    primary_icon: Zap,
  },
  {
    slug: "one-config",
    pillar_title: "Unified Test · Lint · Ignore",
    hook: "One YAML. Every language. Every linter.",
    body: "`.gavel.yaml` is the single policy layer across Go, TypeScript, Python, and Markdown. Native linter configs are still honored — gavel wraps them, it doesn't reinvent them.",
    bullets: [
      "Ignore by source + rule + file",
      "Enable or disable linters",
      "Pre/post hooks in one place",
      "Inherits ~/.gavel.yaml",
    ],
    uniqueness:
      "One ignore list matches across golangci, eslint, ruff, and markdownlint — no per-language config sprawl.",
    animation_kind: "icon-orbit",
    primary_icon: Settings2,
  },
  {
    slug: "pr-watcher",
    pillar_title: "PR Watcher",
    hook: "Your PR queue, in one tab",
    body: "A live dashboard that polls PRs across repos, surfaces GitHub Actions check status, and tails failed-job logs inline — no more F5-ing.",
    bullets: [
      "Filter by author, org, state",
      "Failed-log tails inline",
      "Watch whole orgs in parallel",
      "Terminal or browser surface",
    ],
    uniqueness:
      "`gavel pr list --ui` is the only PR dashboard that folds check status, failed logs, and TODO sync into one view.",
    animation_kind: "pulse-nodes",
    primary_icon: GitPullRequest,
  },
  {
    slug: "menu-bar",
    pillar_title: "macOS Menu Bar",
    hook: "Ambient CI, always on",
    body: "A native macOS menu-bar indicator color-coded by the worst current state across your PRs — green passing, amber pending, red broken. Click for the full dashboard.",
    bullets: [
      "Persists across sessions",
      "Survives reboots as a service",
      "Single-click to full UI",
      "No dock icon required",
    ],
    uniqueness:
      "`gavel pr list --menu-bar` installs as a launchd agent — a GitHub checks light permanently wired into your desktop.",
    animation_kind: "beam-connector",
    primary_icon: MonitorDot,
  },
];

const LINTER_ICONS: LucideIcon[] = [
  Boxes,
  CircleDot,
  ScanSearch,
  Sparkles,
  TypeIcon,
  FileSearch,
  Cloud,
  GitCommitHorizontal,
  KeyRound,
];

function AnimationVisual({ feature }: { feature: Feature }) {
  switch (feature.animation_kind) {
    case "typing":
      return (
        <div className="flex h-full w-full items-center justify-center p-3">
          <div className="w-full max-w-sm overflow-hidden rounded-md border border-border bg-background shadow-sm">
            <div className="flex items-center gap-1.5 border-b border-border bg-muted/60 px-3 py-1.5">
              <span className="h-2 w-2 rounded-full bg-red-400/70" />
              <span className="h-2 w-2 rounded-full bg-amber-400/70" />
              <span className="h-2 w-2 rounded-full bg-emerald-400/70" />
              <span className="ml-2 font-mono text-[0.65rem] text-muted-foreground">~/repo</span>
            </div>
            <div className="px-3 py-3 font-mono text-xs">
              <TypingAnimation text="$ gavel test --changed --cache" />
            </div>
          </div>
        </div>
      );
    case "icon-orbit":
      return <IconOrbit icons={LINTER_ICONS} />;
    case "pulse-nodes":
      return <PulseNodes count={5} labels={["#12", "#13", "#14", "#15", "✓"]} />;
    case "beam-connector":
      return (
        <AnimatedBeam
          fromLabel="github"
          toLabel="menu bar"
          fromIcon={<GitPullRequest size={16} />}
          toIcon={<GaugeCircle size={16} />}
        />
      );
  }
}

function BentoCard({ feature, className }: { feature: Feature; className?: string }) {
  const Icon = feature.primary_icon;
  return (
    <article
      id={feature.slug}
      className={cn(
        "group flex flex-col overflow-hidden rounded-xl border border-border bg-card transition-colors hover:border-brand-sky/40",
        className,
      )}
    >
      <div className="relative h-44 border-b border-border bg-gradient-to-br from-muted/40 via-background to-background">
        <AnimationVisual feature={feature} />
      </div>
      <div className="flex flex-1 flex-col p-6">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-brand-sky/10 text-brand-sky">
            <Icon size={18} />
          </span>
          <h3 className="text-lg font-semibold">{feature.pillar_title}</h3>
        </div>
        <p className="mt-2 text-sm font-medium text-foreground/90">{feature.hook}</p>
        <p className="mt-3 text-sm text-muted-foreground">{feature.body}</p>

        <ul className="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-sm text-foreground/80">
          {feature.bullets.map((b) => (
            <li key={b} className="flex gap-2 before:mt-2 before:h-1 before:w-1 before:shrink-0 before:rounded-full before:bg-brand-sky">
              <span>{b}</span>
            </li>
          ))}
        </ul>

        <p className="mt-auto pt-5 text-xs italic text-muted-foreground">
          <span className="font-medium not-italic text-brand-sky">Unique:</span>{" "}
          {feature.uniqueness}
        </p>
      </div>
    </article>
  );
}

export default function BentoGrid() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-20">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">What ships</p>
          <h2 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">
            Four pillars, one binary
          </h2>
        </div>
        <p className="max-w-sm text-sm text-muted-foreground">
          Every card below is a capability MANUAL.md documents today — not a roadmap.
        </p>
      </div>

      <div className="mt-10 grid gap-4 lg:grid-cols-2">
        {features.map((f) => (
          <BentoCard key={f.slug} feature={f} />
        ))}
      </div>
    </section>
  );
}
