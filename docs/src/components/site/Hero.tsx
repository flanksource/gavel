import { ArrowRight } from "lucide-react";
import AnimatedGradientText from "@/magic/AnimatedGradientText";
import ShimmerButton from "@/magic/ShimmerButton";
import CopyInstall from "./CopyInstall";

export default function Hero() {
  return (
    <section className="mx-auto max-w-6xl px-6 pb-16 pt-20 sm:pt-28">
      <p className="text-sm font-medium uppercase tracking-wider text-muted-foreground">
        Open source · Flanksource
      </p>
      <h1 className="mt-4 text-4xl font-bold leading-tight tracking-tight sm:text-6xl">
        One CLI for{" "}
        <AnimatedGradientText>tests, linting & AI code review</AnimatedGradientText>
      </h1>
      <p className="mt-6 max-w-2xl text-lg text-muted-foreground">
        Gavel runs your tests, orchestrates nine linters in parallel, generates
        conventional commits, and ships AI-powered PR review — all from a single
        CLI designed for local dev and CI alike.
      </p>

      <div className="mt-10 flex flex-col gap-4 sm:flex-row sm:items-center">
        <ShimmerButton onClick={() => document.getElementById("waitlist")?.scrollIntoView({ behavior: "smooth" })}>
          Join the waitlist
          <ArrowRight size={16} className="ml-2" />
        </ShimmerButton>
        <a
          href="/docs/getting-started"
          className="inline-flex h-11 items-center justify-center rounded-md border border-border px-6 text-sm font-medium text-foreground transition-colors hover:bg-muted"
        >
          Read the docs
        </a>
      </div>

      <CopyInstall className="mt-8 max-w-2xl" />
    </section>
  );
}
