import type React from "react";
import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ProcessTable } from "./ProcessTable";

vi.mock("@flanksource/clicky-ui/components", () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  Modal: ({ open, children }: { open: boolean; children: ReactNode }) => open ? <section>{children}</section> : null,
  Select: (props: React.SelectHTMLAttributes<HTMLSelectElement>) => <select {...props} />,
}));

vi.mock("@flanksource/clicky-ui/data", () => ({
  AnsiHtml: ({ text }: { text: string }) => <pre>{text}</pre>,
  TimeseriesCoreBars: () => <span data-testid="metric" />,
}));

afterEach(() => vi.restoreAllMocks());

describe("ProcessTable", () => {
  it("shows the extended descendant tree and links its supervised task generation", async () => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.startsWith("/api/proc/status")) {
        return {
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
      return { text: async () => "ready" } as Response;
    }));

    render(<ProcessTable
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
    />);

    fireEvent.click(screen.getByTitle("Toggle logs"));

    expect((await screen.findByRole("link", { name: "Task details" })).getAttribute("href")).toBe("/tasks/generation-123");
    expect(await screen.findByText("worker --serve")).toBeTruthy();
    expect(screen.getByText("VMS")).toBeTruthy();
    expect(screen.getByText("root")).toBeTruthy();
    expect(screen.getByText("sleeping")).toBeTruthy();
    expect(screen.getByText("7")).toBeTruthy();
  });
});
