import { execFileSync } from "node:child_process";
import { mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { basename, resolve } from "node:path";
import { expect, it } from "vitest";

interface Example {
  label: string;
  files: Record<"yaml" | "markdown", string>;
}

interface Outline {
  tree: { file: string; fixtures: number }[];
}

it("provides YAML-first fixture examples whose two variants parse into runnable steps", () => {
  mkdirSync(resolve(".tmp"), { recursive: true });
  const directory = mkdtempSync(resolve(".tmp/fixture-examples-"));
  try {
    const examples: Example[] = [];
    for (const file of readdirSync("content/docs", { recursive: true, encoding: "utf8" })) {
      if (!file.endsWith(".mdx")) continue;
      const source = readFileSync(resolve("content/docs", file), "utf8");
      for (const [index, match] of [...source.matchAll(/<FixtureExample>([\s\S]*?)<\/FixtureExample>/g)].entries()) {
        const label = `${file} example ${index + 1}`;
        const variants = [...match[1]!.matchAll(/^(`{3,})(yaml|markdown)\n([\s\S]*?)^\1$/gm)];
        expect(variants.map((variant) => variant[2]), label).toEqual(["yaml", "markdown"]);
        const files = {} as Example["files"];
        for (const variant of variants) {
          const kind = variant[2] as "yaml" | "markdown";
          const name = `example-${examples.length}.fixture.${kind === "yaml" ? "yaml" : "md"}`;
          writeFileSync(resolve(directory, name), variant[3]!.replace(/(?:[ \t]*(?:#|\/\/|<!--)\s*)?\[!code[^\]]+\](?:\s*-->)?/g, ""));
          files[kind] = name;
        }
        examples.push({ label, files });
      }
    }
    expect(examples.length).toBeGreaterThan(0);

    // One outline over every variant: a gavel process per file made this test
    // spend its whole budget on process start-up under a loaded CI runner.
    const paths = examples.flatMap((example) => [example.files.yaml, example.files.markdown].map((name) => resolve(directory, name)));
    const outline: Outline = JSON.parse(execFileSync("gavel", ["fixtures", "outline", ...paths, "--json"], { encoding: "utf8", stdio: "pipe" }));
    const fixtures = new Map(outline.tree.map((node) => [basename(node.file), node.fixtures]));
    for (const example of examples) {
      const yaml = fixtures.get(example.files.yaml);
      const markdown = fixtures.get(example.files.markdown);
      expect(yaml, `${example.label} yaml`).toBeGreaterThan(0);
      expect(markdown, `${example.label} markdown`).toBeGreaterThan(0);
      expect(yaml, `${example.label} preserves its checks`).toBe(markdown);
    }
  } finally {
    rmSync(directory, { recursive: true });
  }
}, 30000);
