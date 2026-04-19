import type { LucideIcon } from "lucide-react";
import {
  FileText,
  GaugeCircle,
  GitBranch,
  ListChecks,
  PlayCircle,
  Send,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface Feature {
  slug: string;
  title: string;
  taglines: [string, string, string];
  description: string;
  body: string;
  icon: LucideIcon;
}

const features: Feature[] = [
  {
    slug: "markdown-fixtures",
    title: "Markdown Fixture Tests",
    taglines: [
      "Tests that double as docs",
      "Tables in, CEL assertions out",
      "Write a smoke test in the README",
    ],
    description:
      "Declarative *.fixture.md files where each table row is a test case and assertions are CEL expressions over stdout or parsed JSON.",
    body: "Some tests are easier to explain as a table than as Go code. Fixture files are Markdown with a YAML front-matter command template and a table of rows; each row's columns become template variables, and a `CEL` column expresses the assertion. They run under `gavel fixtures` or fold into `gavel test --fixtures`, and because they're Markdown they render correctly on GitHub — so your API contract tests and your onboarding doc can be the same file.",
    icon: FileText,
  },
  {
    slug: "ui-cache",
    title: "Live UI with Replay Cache",
    taglines: [
      "A browser dashboard that survives the run",
      "Share a CI result as a local URL",
      "Replay any run. Rerun any test.",
    ],
    description:
      "`gavel test --ui` renders a live dashboard during the run; `gavel ui serve` replays any prior snapshot without re-executing a single test.",
    body: "Every run — local or CI — produces a JSON snapshot that the live `--ui` flag renders in real time: hook progress, test tree, lint violations, and failed-log tails, all in one browser tab. When the run ends, the snapshot doesn't disappear: `gavel ui serve run.json` replays it offline, and `--detach` hands the live dashboard off to a child server that keeps the URL alive after the parent process exits.",
    icon: PlayCircle,
  },
  {
    slug: "push-to-test",
    title: "Push-to-Test Remote",
    taglines: [
      "Push to test. No CI round-trip.",
      "Your laptop, your CI, one `git push` away",
      "Pre-flight checks at git-remote speed",
    ],
    description:
      "Add gavel as a git remote; every push runs `gavel test --lint` against the pushed ref and streams results back over the same connection.",
    body: "`git push gavel HEAD` becomes your fastest pre-CI. The gavel daemon accepts pushes over SSH, checks out the ref on a cached repo clone, runs whatever command you set in `.gavel.yaml`, and streams structured output back to the pushing terminal. Cached clones keep iterative pushes fast, `--changed` keeps each run short, and one gavel instance can be shared across a team so everyone runs the same check on the same cached repo state.",
    icon: Send,
  },
  {
    slug: "bench-compare",
    title: "Benchmark Regression Gate",
    taglines: [
      "Benchmarks, versioned",
      "Fail the PR when perf slips 15%",
      "`go test -bench` with a brain",
    ],
    description:
      "`gavel bench run` captures a benchmark snapshot to JSON; `gavel bench compare` fails the build when head regresses against base beyond a configurable threshold.",
    body: "Go benchmarks are powerful and almost always discarded the moment the terminal scrolls past them. Gavel treats a benchmark run as an artifact: run it on main, run it on your branch, compare the two JSON files, and get a pass/fail on a percent-regression threshold you choose. `--ui` renders the comparison as a browsable table with per-benchmark deltas, so a 12% regression on `BenchmarkEncode` stops being something you notice three weeks later.",
    icon: GaugeCircle,
  },
  {
    slug: "todos",
    title: "TODO Execution Loop",
    taglines: [
      "Failures in, fixes out",
      "Every broken test becomes a tracked task",
      "A queue your agent can work",
    ],
    description:
      "`--sync-todos` on test, lint, and PR commands writes each failure to a Markdown TODO file that `gavel todos run` executes interactively or in batch.",
    body: "Gavel closes the loop between finding problems and fixing them. Every failure-producing command (`test`, `lint`, `pr status`, `pr fix`) can materialize its findings as individual Markdown TODO files, grouped by file, package, or message. `gavel todos list` is your inbox; `gavel todos get` drills into one; `gavel todos run` dispatches them through the Claude Code integration, interactive or batch.",
    icon: ListChecks,
  },
  {
    slug: "git-analyze",
    title: "Semantic Git History",
    taglines: [
      "`git log`, but it understands your repo",
      "Scope · tech · severity · summary",
      "Ship notes without writing them",
    ],
    description:
      "`gavel git analyze` classifies commits by scope, detected technologies, and severity, then optionally summarizes a time window with AI.",
    body: "Once you've told gavel how your repo is laid out — via `arch.yaml` for scope/tech rules and `.gitanalyze.yaml` for include/exclude filters — `git analyze` becomes the answer to 'what happened this week?' Filter history by scope, by technology, or by author; fold a week or a month into a summary; exclude bots and merge commits so the signal survives. The same classification powers `gavel repomap`, so the taxonomy you set once keeps paying off across commands.",
    icon: GitBranch,
  },
];

function FeatureCard({ feature, className }: { feature: Feature; className?: string }) {
  const Icon = feature.icon;
  return (
    <article
      id={feature.slug}
      className={cn(
        "group flex flex-col rounded-xl border border-border bg-card p-6 transition-colors hover:border-brand-sky/40",
        className,
      )}
    >
      <div className="flex items-center gap-3">
        <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-brand-sky/10 text-brand-sky">
          <Icon size={18} />
        </span>
        <h3 className="text-base font-semibold leading-tight">{feature.title}</h3>
      </div>

      <ul className="mt-4 space-y-1 text-sm font-medium text-foreground/90">
        {feature.taglines.map((t) => (
          <li key={t} className="flex gap-2 before:mt-2 before:h-1 before:w-1 before:shrink-0 before:rounded-full before:bg-brand-sky">
            <span>{t}</span>
          </li>
        ))}
      </ul>

      <p className="mt-4 text-sm text-foreground/80">{feature.description}</p>
      <p className="mt-3 text-sm text-muted-foreground">{feature.body}</p>
    </article>
  );
}

export default function FeatureList() {
  return (
    <section className="mx-auto max-w-6xl px-6 pb-20">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">
            Under the hood
          </p>
          <h2 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">
            Six more reasons to install it
          </h2>
        </div>
        <p className="max-w-sm text-sm text-muted-foreground">
          Every entry is a CLI command you can run today — see MANUAL.md for flags.
        </p>
      </div>

      <div className="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {features.map((f) => (
          <FeatureCard key={f.slug} feature={f} />
        ))}
      </div>
    </section>
  );
}
