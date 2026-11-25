import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { GavelRail, type GavelRailProps } from "../src/GavelPanel";
import { GAVEL_COMMENT_CONFIG } from "../src/useGavelComments";
import type { DraftBox } from "./bridge-parse";

const sendToHost = vi.hoisted(() => vi.fn());
vi.mock("./useHostBridge", async importOriginal => ({ ...await importOriginal<typeof import("./useHostBridge")>(), sendToHost }));

beforeEach(() => vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} }));
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  sendToHost.mockReset();
  document.querySelectorAll("[data-floating-ui-portal]").forEach(node => node.remove());
});

const selection = { anchor: "#lifecycle", label: "Lifecycle actions", element: { source: "https://example.test/", html: "<details></details>", pageUrl: "https://example.test/review" } };
const cardBox: DraftBox = { left: 100, top: 96, width: 320 };

function renderDraft(draftBox?: DraftBox) {
  const props: GavelRailProps = {
    host: { send: vi.fn(), requestMode: vi.fn(), maximize: vi.fn() },
    state: { url: "https://example.test/review", title: "Review", selection },
    comments: {
      comments: [], loading: false, error: null,
      refresh: vi.fn(), reply: vi.fn(), edit: vi.fn(), editReply: vi.fn(), updateStatus: vi.fn(), acknowledge: vi.fn(),
      assign: vi.fn(), close: vi.fn(), remove: vi.fn(), applyProposal: vi.fn(),
      rejectProposal: vi.fn(), requestRevision: vi.fn(), loadProposalDiff: vi.fn(),
      deliver: vi.fn(), deliverAll: vi.fn(),
    },
    config: GAVEL_COMMENT_CONFIG,
    create: vi.fn(), floating: true, presentation: "draft", anchors: [], alerts: null,
    ...(draftBox ? { draftBox } : {}),
    renderRunDetails: () => null,
    session: { sessionId: undefined, session: undefined },
    sessionContent: null,
    settingsContent: null,
    extraAssignItems: () => [], assignAll: { label: "Assign all", onSelect: vi.fn() },
  };
  render(<GavelRail {...props} />);
  return screen.getByRole("dialog", { name: `New comment on ${selection.label}` });
}

const lastSize = () => sendToHost.mock.calls.map(([message]) => message).filter(message => message.type === "draft-size").at(-1);

it("asks for the viewport overlay while a menu is portalled and pins the card at the host-reported box", async () => {
  const portal = document.createElement("div");
  portal.setAttribute("data-floating-ui-portal", "");
  document.body.append(portal);
  const card = renderDraft(cardBox);
  await waitFor(() => expect(lastSize()?.overlay).toBe(true));
  // jsdom lays the card out at 0px, so the report is the 16px slack alone — no 480px floor for the open menu.
  expect({ size: lastSize(), position: card.style.position, left: card.style.left, top: card.style.top, width: card.style.width })
    .toEqual({ size: { type: "draft-size", height: 16, overlay: true }, position: "fixed", left: "100px", top: "96px", width: "320px" });
});

it("stays in flow when no menu is open, even with a stale host box", async () => {
  const card = renderDraft(cardBox);
  await waitFor(() => expect(lastSize()).toEqual({ type: "draft-size", height: 16, overlay: false }));
  expect(card.style.position).toBe("");
});

it("marks the frame body transparent only while the floating draft is shown", () => {
  renderDraft();
  const mounted = document.body.hasAttribute("data-floating-draft");
  cleanup();
  expect([mounted, document.body.hasAttribute("data-floating-draft")]).toEqual([true, false]);
});
