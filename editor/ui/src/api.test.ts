import { expect, it } from "vitest";
import { ApiError } from "./api";

it("shows recovery guidance for denied workspace access while retaining its structured body", () => {
  const body = { code: "workspace_access_denied", root: "/documents", error: "Access denied" };
  const error = new ApiError(403, JSON.stringify(body));
  expect(error.message).toBe("Workspace access is required for this folder. Your draft is preserved; restore access before saving.");
  expect(error.json()).toEqual(body);
});
