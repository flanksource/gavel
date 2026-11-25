import { describe, expect, it } from "vitest";
import gavelFill from "@iconify-icons/ph/gavel-fill";
import gavelLight from "@iconify-icons/ph/gavel-light";
import { EXTENSION_ICON_SIZES, GAVEL_FILLED_SVG, GAVEL_SVG, extensionIconSvg } from "./gavel-icon";

const parse = (markup: string) => new DOMParser().parseFromString(markup, "image/svg+xml");

describe("gavel icon markup", () => {
  it("is the Phosphor light and fill gavel as standalone currentColor SVG", () => {
    for (const [markup, icon] of [[GAVEL_SVG, gavelLight], [GAVEL_FILLED_SVG, gavelFill]] as const) {
      const svg = parse(markup).documentElement;
      expect([svg.tagName, svg.getAttribute("viewBox"), svg.getAttribute("xmlns")]).toEqual(["svg", "0 0 256 256", "http://www.w3.org/2000/svg"]);
      expect(svg.innerHTML.replace(/ xmlns="[^"]*"/g, "")).toBe(icon.body);
      expect(markup).toContain('fill="currentColor"');
    }
  });

  it("differs between the outline and the filled gavel", () => {
    expect(GAVEL_SVG).not.toBe(GAVEL_FILLED_SVG);
  });
});

describe("extensionIconSvg", () => {
  const document = parse(extensionIconSvg());

  it("is a valid square SVG with the gavel in white on a blue rounded square", () => {
    const svg = document.documentElement;
    expect([document.querySelector("parsererror"), svg.getAttribute("viewBox")]).toEqual([null, "0 0 128 128"]);
    expect(svg.querySelector("rect")?.getAttribute("fill")).toBe("#2563eb");
    expect(extensionIconSvg()).not.toContain("currentColor");
    expect(svg.querySelector("g path")?.getAttribute("fill")).toBe("#fff");
  });

  it("covers the sizes Chrome asks the manifest for", () => {
    expect(EXTENSION_ICON_SIZES).toEqual([16, 32, 48, 128]);
  });
});
