import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface Tier {
  name: string;
  price: string;
  tagline: string;
  features: string[];
  cta: { label: string; href: string };
  highlight?: boolean;
  badge?: string;
}

const tiers: Tier[] = [
  {
    name: "Open source",
    price: "Free",
    tagline: "The full CLI, self-hosted, forever.",
    features: [
      "Every linter, test runner, and AI verify",
      "GitHub Action + SSH server mode",
      "HTML + JSON reports",
      "MIT-licensed, contribute on GitHub",
    ],
    cta: { label: "Install on GitHub", href: "https://github.com/flanksource/gavel" },
  },
  {
    name: "Team / Hosted",
    price: "Coming soon",
    tagline: "Gavel as a managed service for your team.",
    features: [
      "Hosted CI dashboard across repos and branches",
      "Managed git-push endpoints for rapid incremental test/lint",
      "Shared AI review cache + policy controls",
      "Priority support",
    ],
    cta: { label: "Join the waitlist", href: "#waitlist" },
    highlight: true,
    badge: "Waitlist",
  },
];

export default function PricingSection() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-20">
      <div className="text-center">
        <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">Pricing</p>
        <h2 className="mt-2 text-3xl font-bold tracking-tight sm:text-4xl">Simple and honest</h2>
        <p className="mt-4 text-muted-foreground">
          The OSS CLI stays free. A managed tier is in private preview — join the waitlist below.
        </p>
      </div>

      <div className="mt-12 grid gap-6 sm:grid-cols-2">
        {tiers.map((tier) => (
          <div
            key={tier.name}
            className={cn(
              "flex flex-col rounded-xl border border-border bg-card p-8 transition-colors",
              tier.highlight && "border-brand-sky/60 ring-1 ring-brand-sky/30",
            )}
          >
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold">{tier.name}</h3>
              {tier.badge && (
                <span className="rounded-full border border-brand-sky/40 bg-brand-sky/10 px-3 py-1 text-xs font-medium text-brand-sky">
                  {tier.badge}
                </span>
              )}
            </div>
            <p className="mt-2 text-sm text-muted-foreground">{tier.tagline}</p>
            <p className="mt-6 text-3xl font-bold tracking-tight">{tier.price}</p>

            <ul className="mt-6 flex-1 space-y-3 text-sm">
              {tier.features.map((f) => (
                <li key={f} className="flex gap-3">
                  <Check size={16} className="mt-0.5 shrink-0 text-brand-sky" />
                  <span>{f}</span>
                </li>
              ))}
            </ul>

            <a
              href={tier.cta.href}
              className={cn(
                "mt-8 inline-flex h-11 items-center justify-center rounded-md px-6 text-sm font-medium transition-colors",
                tier.highlight
                  ? "bg-brand-dark text-white hover:bg-brand-dark/90"
                  : "border border-border hover:bg-muted",
              )}
              target={tier.cta.href.startsWith("http") ? "_blank" : undefined}
              rel={tier.cta.href.startsWith("http") ? "noreferrer" : undefined}
            >
              {tier.cta.label}
            </a>
          </div>
        ))}
      </div>
    </section>
  );
}
