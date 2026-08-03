import type React from "react";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ProcStatus, Project } from "../types";
import { queryKeys } from "../query";
import { ProcessTable, WorkspaceGroup } from "./ProcessTable";

vi.mock("@flanksource/clicky-ui/components", () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  Modal: ({ open, children }: { open: boolean; children: ReactNode }) => open ? <section>{children}</section> : null,
  Select: (props: React.SelectHTMLAttributes<HTMLSelectElement>) => <select {...props} />,
}));

vi.mock("@flanksource/clicky-ui/data", () => ({
  AnsiHtml: ({ text }: { text: string }) => <pre>{text}</pre>,
  TimeseriesCoreBars: () => <span data-testid="metric" />,
}));

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

function queryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderWithClient(client: QueryClient, children: ReactNode) {
  return render(<QueryClientProvider client={client}>{children}</QueryClientProvider>);
}

const project: Project = { name: "gavel", dir: "/work/gavel", repos: ["flanksource/gavel"] };

const stoppedStatus: ProcStatus = {
  hasProcfile: true,
  running: false,
  processes: [{
    name: "worker",
    command: "worker --serve",
    status: "stopped",
    restarts: 0,
    logFile: "/work/gavel/.gavel/worker.log",
  }],
  profiles: ["web"],
};

describe("ProcessTable", () => {
  it("shows the extended descendant tree and links its supervised task generation", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith("/api/proc/status")) {
        return {
          ok: true,
          json: async () => ({
            hasProcfile: true,
            running: true,
            processes: [{
              name: "worker",
              status: "running",
              tree: [{
                pid: 41,
                ppid: 1,
                command: "worker --serve",
                root: true,
                status: "sleeping",
                cpuPercent: 12,
                memoryRss: 1024,
                memoryVms: 2048,
                openFiles: 7,
              }],
            }],
          }),
        } as Response;
      }
      return { ok: true, text: async () => "ready" } as Response;
    }));

    renderWithClient(
      queryClient(),
      <ProcessTable
        procs={[{
          project: { name: "gavel", dir: "/work/gavel", repos: [] },
          proc: {
            name: "worker",
            command: "worker --serve",
            status: "running",
            restarts: 0,
            logFile: "/work/gavel/.gavel/worker.log",
            taskRunId: "generation-123",
          },
        }]}
        onChanged={() => undefined}
        showWorkspace={false}
      />,
    );

    fireEvent.click(screen.getByTitle("Toggle logs"));

    expect((await screen.findByRole("link", { name: "Task details" })).getAttribute("href")).toBe("/tasks/generation-123");
    expect(await screen.findByText("worker --serve")).toBeTruthy();
    expect(screen.getByText("VMS")).toBeTruthy();
    expect(screen.getByText("root")).toBeTruthy();
    expect(screen.getByText("sleeping")).toBeTruthy();
    expect(screen.getByText("7")).toBeTruthy();
  });

  it("deduplicates a process control request and updates the exact process caches", async () => {
    const client = queryClient();
    const oldStatus = { ...stoppedStatus, processes: [] };
    client.setQueryData(queryKeys.processStatuses(), {
      gavel: oldStatus,
      "flanksource/gavel": oldStatus,
      clicky: oldStatus,
    });
    client.setQueryData(queryKeys.processStatus(project.name), oldStatus);
    client.setQueryData(queryKeys.projectStatus(project.name), { files: [] });
    const invalidateQueries = vi.spyOn(client, "invalidateQueries");
    let resolveControl: (response: Response) => void = () => undefined;
    const fetchMock = vi.fn(() => new Promise<Response>((resolve) => {
      resolveControl = resolve;
    }));
    vi.stubGlobal("fetch", fetchMock);
    const onChanged = vi.fn();

    renderWithClient(
      client,
      <ProcessTable
        procs={[{ project, proc: stoppedStatus.processes![0] }]}
        onChanged={onChanged}
        showWorkspace={false}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Start" }));
    fireEvent.click(screen.getByRole("button", { name: "Start" }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(fetchMock).toHaveBeenCalledWith("/api/proc/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ project: "gavel", names: ["worker"] }),
    });

    const runningStatus: ProcStatus = {
      ...stoppedStatus,
      running: true,
      processes: [{ ...stoppedStatus.processes![0], status: "running" }],
    };
    resolveControl({
      ok: true,
      status: 200,
      text: async () => JSON.stringify(runningStatus),
    } as Response);

    await waitFor(() => expect(onChanged).toHaveBeenCalledTimes(1));
    expect(client.getQueryData(queryKeys.processStatus("gavel"))).toEqual(runningStatus);
    expect(client.getQueryData<Record<string, ProcStatus>>(queryKeys.processStatuses())).toEqual({
      gavel: runningStatus,
      "flanksource/gavel": runningStatus,
      clicky: oldStatus,
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope("gavel") });
  });

  it("surfaces a workspace control error without reporting a successful change", async () => {
    const fetchMock = vi.fn(async () => ({
      ok: false,
      status: 409,
      text: async () => JSON.stringify({ error: "worker is already running" }),
    } as Response));
    vi.stubGlobal("fetch", fetchMock);
    const onChanged = vi.fn();

    renderWithClient(
      queryClient(),
      <WorkspaceGroup project={project} status={stoppedStatus} onChanged={onChanged} />,
    );

    fireEvent.change(screen.getByTitle("Profile to start").querySelector("select")!, { target: { value: "web" } });
    fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect((await screen.findByRole("alert")).textContent).toContain("worker is already running");
    expect(fetchMock).toHaveBeenCalledWith("/api/proc/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ project: "gavel", profile: "web" }),
    });
    expect(onChanged).not.toHaveBeenCalled();
    expect((screen.getByRole("button", { name: "Start" }) as HTMLButtonElement).disabled).toBe(false);
  });
});
