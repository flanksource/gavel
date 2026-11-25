import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { listProjects } from "./gavel-client";
import { fetchRuntimeCatalog } from "./runtime-catalog";
import { hasServerAccess } from "./server-access";
import { useGavelChoices } from "./useGavelChoices";

const agentSessions = vi.hoisted(() => vi.fn());
vi.mock("./runtime-catalog", () => ({ fetchRuntimeCatalog: vi.fn() }));
vi.mock("./server-access", () => ({ hasServerAccess: vi.fn(), requestServerAccess: vi.fn() }));
vi.mock("./gavel-client", () => ({
  listProjects: vi.fn(),
  GavelClient: class {
    agentSessions = agentSessions;
  },
}));

const server = "http://localhost:9092";
const catalog = { models: [], families: [] };
const sessions = [{ id: "s-1", source: "claude", title: "Review changes", inCmux: true }];

beforeEach(() => {
  vi.mocked(hasServerAccess).mockResolvedValue(true);
  vi.mocked(listProjects).mockResolvedValue(["acme", "docs"]);
  vi.mocked(fetchRuntimeCatalog).mockResolvedValue(catalog);
  agentSessions.mockResolvedValue(sessions);
});
afterEach(() => vi.clearAllMocks());

type Dashboard = { server: string; project?: string };
const choicesFor = (initial: Dashboard) => renderHook(({ dashboard }: { dashboard: Dashboard }) => useGavelChoices(dashboard), { initialProps: { dashboard: initial } });

describe("error", () => {
  it("clears a failed session list once the project changes and its sessions load", async () => {
    agentSessions.mockRejectedValueOnce(new Error("sessions unavailable"));
    const { result, rerender } = choicesFor({ server, project: "acme" });
    await waitFor(() => expect(result.current.error).toBe("sessions unavailable"));
    rerender({ dashboard: { server, project: "docs" } });
    await waitFor(() => expect(result.current.agentSessions).toEqual(sessions));
    expect(result.current.error).toBe("");
  });

  it("clears a failed session list when the project is removed", async () => {
    agentSessions.mockRejectedValueOnce(new Error("sessions unavailable"));
    const { result, rerender } = choicesFor({ server, project: "acme" });
    await waitFor(() => expect(result.current.error).toBe("sessions unavailable"));
    rerender({ dashboard: { server } });
    await waitFor(() => expect(result.current.error).toBe(""));
  });

  it("keeps a failed runtime catalog while the sessions reload, and says it first when both failed", async () => {
    vi.mocked(fetchRuntimeCatalog).mockRejectedValue(new Error("catalog unavailable"));
    agentSessions.mockRejectedValueOnce(new Error("sessions unavailable"));
    const { result, rerender } = choicesFor({ server, project: "acme" });
    await waitFor(() => expect(result.current.error).toBe("catalog unavailable"));
    rerender({ dashboard: { server, project: "docs" } });
    await waitFor(() => expect(result.current.agentSessions).toEqual(sessions));
    expect(result.current.error).toBe("catalog unavailable");
  });

  it("clears a failed project load when the dashboard changes to one that answers", async () => {
    vi.mocked(listProjects).mockRejectedValueOnce(new Error("connection refused"));
    const { result, rerender } = choicesFor({ server });
    await waitFor(() => expect(result.current.error).toBe("connection refused"));
    rerender({ dashboard: { server: "http://localhost:9093" } });
    await waitFor(() => expect(result.current.projects).toEqual(["acme", "docs"]));
    expect(result.current.error).toBe("");
    expect(result.current.connectionError).toBe("");
  });
});

describe("connectionError", () => {
  it("is set when the project list cannot be loaded", async () => {
    vi.mocked(listProjects).mockRejectedValue(new Error("connection refused"));
    const { result } = choicesFor({ server });
    await waitFor(() => expect(result.current.connectionError).toBe("connection refused"));
    expect(result.current.error).toBe("connection refused");
  });

  it("is set for an invalid dashboard address, and the sessions load stays quiet about it", async () => {
    const { result } = choicesFor({ server: "localhost:9092", project: "acme" });
    await waitFor(() => expect(result.current.connectionError).toContain("must be an HTTP(S) URL"));
    expect(result.current.error).toBe(result.current.connectionError);
    expect(agentSessions).not.toHaveBeenCalled();
  });

  it("is not set for a failed session list, which only reaches the card's alert", async () => {
    agentSessions.mockRejectedValue(new Error("sessions unavailable"));
    const { result } = choicesFor({ server, project: "acme" });
    await waitFor(() => expect(result.current.error).toBe("sessions unavailable"));
    expect(result.current.connectionError).toBe("");
    expect(result.current.projects).toEqual(["acme", "docs"]);
  });

  it("is not set for a failed runtime catalog", async () => {
    vi.mocked(fetchRuntimeCatalog).mockRejectedValue(new Error("catalog unavailable"));
    const { result } = choicesFor({ server, project: "acme" });
    await waitFor(() => expect(result.current.error).toBe("catalog unavailable"));
    expect(result.current.connectionError).toBe("");
    await waitFor(() => expect(result.current.projects).toEqual(["acme", "docs"]));
  });
});
