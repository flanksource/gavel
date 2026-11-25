import { useState } from "react";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { TodoRunControls } from "./TodoRunControls";
import { runCaptureError, type TodoRun } from "./runtime";
import { context, steps } from "./testing/runtime";

afterEach(cleanup);

function Harness({ remembered }: { remembered?: TodoRun }) {
  const [run, setRun] = useState<TodoRun>();
  return <>
    <TodoRunControls run={run} context={context} steps={steps} remembered={remembered} ready disabled={false} error={runCaptureError(run, context, steps)} onChange={setRun} />
    <output aria-label="Queued runtime">{JSON.stringify(run ?? null)}</output>
  </>;
}

describe("offline todo run controls", () => {
  it("starts with the workspace lifecycle defaults and lets the shared runtime bar choose a model", () => {
    render(<Harness />);
    fireEvent.click(screen.getByRole("switch", { name: "Start a new run" }));
    expect(JSON.parse(screen.getByLabelText("Queued runtime").textContent ?? "")).toEqual({ step: "run", spec: {} });
    fireEvent.click(screen.getByTitle("Model — unspecified"));
    fireEvent.mouseDown(screen.getByRole("option", { name: /^Acme model/ }));
    expect(JSON.parse(screen.getByLabelText("Queued runtime").textContent ?? "")).toMatchObject({ step: "run", spec: { model: "acme-model", mode: "agent" } });
  });

  it("restores cached run choices without replacing their spec or restarting a disabled run", () => {
    const remembered = { step: "plan", spec: { model: "acme-model", mode: "agent", effort: "high", budget: { cost: 3 } } };
    render(<Harness remembered={remembered} />);
    expect(JSON.parse(screen.getByLabelText("Queued runtime").textContent ?? "")).toEqual(remembered);
    fireEvent.click(screen.getByRole("switch", { name: "Start a new run" }));
    expect(screen.getByLabelText("Queued runtime").textContent).toBe("null");
  });
});
