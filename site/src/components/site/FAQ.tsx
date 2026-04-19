import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

const faqs = [
  {
    q: "What license does gavel ship under?",
    a: "MIT. Fork it, vendor it, embed it in your own tooling — no strings.",
  },
  {
    q: "Do I have to self-host?",
    a: "The CLI is self-hosted by design and runs anywhere Go binaries do. A managed tier is on the roadmap (join the waitlist).",
  },
  {
    q: "Which languages are supported?",
    a: "First-class support for Go, TypeScript, Python, and Markdown today. Linters auto-activate when their native config file is discovered — so eslint only runs if you have an ESLint config, and so on.",
  },
  {
    q: "How is this different from mega-linter or reviewdog?",
    a: "Gavel bundles a native test runner, AI-powered verify pass, PR/commit automation, and a live UI — not just a linter fan-out. Each piece can stand alone, but they share a single config, cache, and report format.",
  },
  {
    q: "Does gavel require sending my code to an LLM?",
    a: "Only the `verify`, `commit`, and `pr fix` flows call an LLM, and they are opt-in. Tests and linting run entirely locally.",
  },
];

export default function FAQ() {
  return (
    <section className="mx-auto max-w-3xl px-6 py-20">
      <div className="text-center">
        <p className="text-sm font-medium uppercase tracking-wider text-brand-sky">FAQ</p>
        <h2 className="mt-2 text-3xl font-bold tracking-tight">Frequently asked</h2>
      </div>

      <Accordion type="single" collapsible className="mt-10">
        {faqs.map((f, i) => (
          <AccordionItem key={f.q} value={`item-${i}`}>
            <AccordionTrigger>{f.q}</AccordionTrigger>
            <AccordionContent>{f.a}</AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>
    </section>
  );
}
