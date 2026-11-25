import { Children, isValidElement, type ReactNode } from "react";
import { Tabs } from "@rspress/core/theme";
import { Markdown } from "@flanksource/clicky-ui/data";

function codeText(node: ReactNode): string {
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (Array.isArray(node)) return node.map(codeText).join("");
  if (isValidElement<{ children?: ReactNode }>(node)) return codeText(node.props.children);
  if (node == null || typeof node === "boolean") return "";
  throw new Error("FixtureExample code must contain text or rendered code tokens");
}

export default function FixtureExample({ children }: { children: ReactNode }) {
  if (import.meta.env.SSG_MD) return <>{children}</>;
  const variants = Children.toArray(children).filter(isValidElement);
  if (variants.length !== 2) throw new Error("FixtureExample requires YAML and Markdown variants, in that order");
  const markdown = codeText(variants[1]).replace(/^---\r?\n[\s\S]*?\r?\n---(?:\r?\n|$)/, "").trim();
  if (!markdown) throw new Error("FixtureExample requires nonempty Markdown source for its preview");
  return <Tabs defaultValue="yaml" keepDOM={false} values={[
    { label: "YAML", value: "yaml" },
    { label: "Markdown", value: "markdown" },
    { label: "Preview", value: "preview" },
  ]}>
    {variants}
    <div className="fixture-preview min-w-0 p-4" aria-label="Markdown preview">
      <Markdown text={markdown} />
    </div>
  </Tabs>;
}
