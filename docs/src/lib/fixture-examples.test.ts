import { execFileSync } from "node:child_process";
import { mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, it } from "vitest";

it("provides YAML-first fixture examples whose two variants parse into runnable steps", () => {
  mkdirSync(resolve(".tmp"), { recursive: true });
  const directory = mkdtempSync(resolve(".tmp/fixture-examples-"));
  let examples = 0;
  try {
    for (const file of readdirSync("content/docs", { recursive: true, encoding: "utf8" })) {
      if (!file.endsWith(".mdx")) continue;
      const source = readFileSync(resolve("content/docs", file), "utf8");
      for (const [index, match] of [...source.matchAll(/<FixtureExample>([\s\S]*?)<\/FixtureExample>/g)].entries()) {
        const variants = [...match[1]!.matchAll(/^(`{3,})(yaml|markdown)\n([\s\S]*?)^\1$/gm)];
        expect(variants.map((variant) => variant[2]), `${file} example ${index + 1}`).toEqual(["yaml", "markdown"]);
        const stepCounts: number[] = [];
        for (const variant of variants) {
          const path = resolve(directory, `example.fixture.${variant[2] === "yaml" ? "yaml" : "md"}`);
          const content = variant[3]!.replace(/(?:[ \t]*(?:#|\/\/|<!--)\s*)?\[!code[^\]]+\](?:\s*-->)?/g, "");
          writeFileSync(path, content);
          const outline = JSON.parse(execFileSync("gavel", ["fixtures", "outline", path, "--json"], { encoding: "utf8", stdio: "pipe" }));
          expect(outline.fixtures, `${file} example ${index + 1} ${variant[2]}`).toBeGreaterThan(0);
          stepCounts.push(outline.fixtures);
        }
        expect(stepCounts[0], `${file} example ${index + 1} preserves its checks`).toBe(stepCounts[1]);
        examples++;
      }
    }
    expect(examples).toBeGreaterThan(0);
  } finally {
    rmSync(directory, { recursive: true });
  }
}, 30000);
