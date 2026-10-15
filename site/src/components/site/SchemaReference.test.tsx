import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import SchemaReference from "./SchemaReference";

describe("SchemaReference", () => {
  it("documents nested arrays, maps, alternatives, constraints and referenced types", () => {
    render(<SchemaReference id="config" schema={{
      type: "object",
      required: ["steps"],
      properties: {
        enabled: { type: "boolean", default: false, description: "Enable checks." },
        steps: { type: "array", minItems: 1, items: { oneOf: [
          { title: "exec", type: "object", properties: {
            kind: { const: "exec" }, timeout: { type: "string", format: "duration" },
          } },
          { title: "test", type: "object", properties: { kind: { const: "test" } } },
        ] } },
        runtimes: { type: "object", additionalProperties: {
          type: "object", properties: { cpu: { type: "number", minimum: 0 } },
        } },
        model: { anyOf: [{ type: "string" }, { $ref: "#/$defs/Model" }] },
      },
      $defs: { Model: { type: "object", properties: {
        mode: { type: "string", enum: ["api", "agent"] },
        next: { $ref: "#/$defs/Model" },
      } } },
    }} />);

    expect(screen.getByText("Enable checks.")).toBeInTheDocument();
    expect(screen.getByText("default: false")).toBeInTheDocument();
    expect(screen.getByText("runtimes.<name>.cpu")).toBeInTheDocument();
    expect(screen.getByText("steps[] (exec).timeout")).toBeInTheDocument();
    expect(screen.getByText("steps[] (test).kind")).toBeInTheDocument();
    expect(screen.getByText("minimum: 0")).toBeInTheDocument();
    expect(screen.getByText("format: duration")).toBeInTheDocument();
    expect(within(screen.getByText("steps").closest("summary")!).getByText("required")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "Model" })[0]).toHaveAttribute("href", "#config-Model");
    expect(document.getElementById("config-Model")).toHaveTextContent("api");
    fireEvent.click(screen.getAllByRole("link", { name: "Model" })[0]!);
    expect(document.getElementById("config-Model")).toHaveAttribute("open");
    expect(screen.getByText("runtimes.<name>.cpu").closest("details")).toHaveAttribute("data-path", "runtimes.<name>");
    expect(screen.getByText("runtimes.<name>.cpu").closest("details")!.parentElement?.closest("details")).toHaveAttribute("data-path", "runtimes");
  });
});
