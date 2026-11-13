import { describe, expect, it } from "vitest";
import { todoRelayTarget } from "./todo-relay-target";

const server = "http://localhost:9092";
const page = "https://app.example";
const form = (params: Record<string, string> = {}) => `${server}/todos/new?${new URLSearchParams({ embed: "1", title: "Card", body: "b", sourceUrl: `${page}/dash?x=1`, ...params })}`;
const options = (overrides: Partial<Parameters<typeof todoRelayTarget>[0]> = {}) => ({ target: form(), parentOrigin: page, ancestorOrigin: page, server, ...overrides });

describe("todoRelayTarget", () => {
  it("accepts the configured server's embed form for the page that frames the relay", () => {
    const { url, parentOrigin } = todoRelayTarget(options());
    expect([url.href, parentOrigin]).toEqual([form(), page]);
  });

  it("accepts a form without a source URL, and a file page, whose origin is null", () => {
    expect(() => todoRelayTarget(options({ target: `${server}/todos/new?embed=1` }))).not.toThrow();
    const file = todoRelayTarget(options({ target: form({ sourceUrl: "file:///Users/me/index.html" }), parentOrigin: "null", ancestorOrigin: "file://" }));
    expect(file.parentOrigin).toBe("null");
  });

  it.each([
    { name: "no target", overrides: { target: null }, error: /requires a target/ },
    { name: "no parent origin", overrides: { parentOrigin: null }, error: /requires the origin/ },
    { name: "an unframed relay", overrides: { ancestorOrigin: undefined }, error: /framed by a page/ },
    { name: "another server", overrides: { target: "http://evil.example/todos/new?embed=1" }, error: /not the configured gavel server/ },
    { name: "the same host on another port", overrides: { target: "http://localhost:9093/todos/new?embed=1" }, error: /not the configured gavel server/ },
    { name: "a different gavel page", overrides: { target: `${server}/api/todos/attachments?embed=1` }, error: /not the gavel new-todo form/ },
    { name: "a form that is not the embed form", overrides: { target: `${server}/todos/new` }, error: /not the gavel new-todo form/ },
    { name: "a target with credentials", overrides: { target: "http://user:pw@localhost:9092/todos/new?embed=1" }, error: /not the configured gavel server|credentials/ },
    { name: "a parent origin that is not an origin", overrides: { parentOrigin: `${page}/x`, ancestorOrigin: `${page}/x` }, error: /Invalid parent origin/ },
    { name: "a page that is not the one framing the relay", overrides: { ancestorOrigin: "https://bank.example" }, error: /not the framing page/ },
    { name: "a source URL from another site", overrides: { target: form({ sourceUrl: "https://bank.example/account" }) }, error: /source URL belongs to/ },
  ])("refuses $name", ({ overrides, error }) => {
    expect(() => todoRelayTarget(options(overrides as Partial<Parameters<typeof todoRelayTarget>[0]>))).toThrow(error);
  });
});
