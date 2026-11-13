import { expect, it } from "vitest";
import { commentPins } from "./comment-pins";

it("counts root comments and retains saved locators for moved-anchor recovery", () => {
  const locators = [{ kind: "css" as const, selector: "#action" }];
  expect(commentPins([
    { anchor: "#action" },
    { anchor: "#action", element: { locators } },
    { anchor: "#action", parentId: "thread" },
    { anchor: null },
    { anchor: "#legacy" },
  ])).toEqual([{ anchor: "#action", count: 2, locators }, { anchor: "#legacy", count: 1, locators: [] }]);
});
