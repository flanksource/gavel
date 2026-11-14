import { existsSync } from "node:fs";
import { describe, expect, it } from "vitest";
import config from "../../rspress.config";
import { buildSidebar, readDocs } from "../../config/content";
import type { SidebarGroup, SidebarItem } from "@rspress/core";

describe("Rspress application", () => {
  it("publishes every guide once in the sidebar and keeps Getting started first", () => {
    function links(items: (SidebarGroup | SidebarItem)[]): string[] {
      return items.flatMap((item) => [item.link!, ...("items" in item ? links(item.items as (SidebarGroup | SidebarItem)[]) : [])]);
    }
    const sidebarLinks = links(buildSidebar());
    expect(sidebarLinks[0]).toBe("/docs/getting-started");
    expect(sidebarLinks.sort()).toEqual(readDocs().map((doc) => `/docs/${doc.slug}`).sort());
  });

  it("registers nested service guides at their existing public URLs", async () => {
    const addPages = config.plugins?.find((plugin) => plugin.name === "gavel-documentation-routes")?.addPages;
    if (!addPages) throw new Error("Service route registration is missing");
    const pages = await addPages(config, true);
    expect(pages.map((page) => page.routePath).sort()).toEqual([
      "/docs/service-dbtest", "/docs/service-deps", "/docs/service-exec", "/docs/service-git",
    ]);
    for (const page of pages) expect(existsSync(page.filepath!), page.filepath).toBe(true);
  });

  it("keeps the services hierarchy in the Rspress sidebar", () => {
    expect(buildSidebar([
      { slug: "fixtures", frontmatter: { title: "Fixtures" } },
      { slug: "services", frontmatter: { title: "Services", parent: "fixtures" } },
      { slug: "service-dbtest", frontmatter: { title: "PostgreSQL", parent: "services" } },
    ])).toEqual([{
      text: "Fixtures", link: "/docs/fixtures", collapsible: true, collapsed: true,
      items: [{ text: "Services", link: "/docs/services", collapsible: true, collapsed: true,
        items: [{ text: "PostgreSQL", link: "/docs/service-dbtest" }] }],
    }]);
  });
});
