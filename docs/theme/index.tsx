import Callout from "../src/components/site/Callout";
import { getCustomMDXComponent as defaultComponents } from "@rspress/core/theme-original";

export * from "@rspress/core/theme-original";

export function getCustomMDXComponent() {
  return { ...defaultComponents(), Callout };
}
