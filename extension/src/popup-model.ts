import type { AgentSessionGrouping, AgentSessionRuntime } from "../contracts/agent-session-store";
import type { SiteSettings } from "../host/site-settings";
import { DEFAULT_GROUPING, sameJson, serverOrigin, type GlobalDefaults } from "./settings";

/** What the extension has stored: its defaults and this site's overrides. */
export type Stored = { global: GlobalDefaults; site: SiteSettings };
export type ScopedField = "server" | "project" | "runtime" | "grouping";
export type FieldValues = { server: string; project: string; runtime: AgentSessionRuntime; grouping: AgentSessionGrouping };

function omit<T extends object>(value: T, keys: readonly (keyof T)[]): T {
  return Object.fromEntries(Object.entries(value).filter(([key]) => !keys.includes(key as keyof T))) as T;
}

/** The server and the project of its dashboard go together; every other field stands alone. */
function ownKeys(field: ScopedField): readonly (keyof SiteSettings)[] {
  return field === "server" ? ["server", "project"] : [field];
}

/** Whether this site stores its own value for the field. */
export function isSiteOnly(stored: Stored, field: ScopedField): boolean {
  return stored.site[field] !== undefined;
}

function editServer(stored: Stored, input: string): Stored {
  const { global, site } = stored;
  const server = serverOrigin(input);
  if (server === (site.server ?? global.server)) return { global: { ...global }, site: { ...site } };
  const rest = omit(site, ownKeys("server"));
  return { global: { ...global }, site: server === global.server ? rest : { ...rest, server } };
}

function editInherited(stored: Stored, field: Exclude<ScopedField, "server">, value: FieldValues[typeof field]): Stored {
  const { global, site } = stored;
  const dashboardOwn = field === "project" && site.server !== undefined;
  const fallback = field === "grouping" ? (global.grouping ?? DEFAULT_GROUPING) : global[field];
  if (!dashboardOwn && sameJson(value, fallback)) return { global: { ...global }, site: omit(site, [field]) };
  if (!dashboardOwn && fallback === undefined) return { global: { ...global, [field]: value }, site: omit(site, [field]) };
  return { global: { ...global }, site: { ...site, [field]: value } };
}

/** An edit made in the popup. Returns the new stored state; never mutates its input. */
export function editField<F extends ScopedField>(stored: Stored, field: F, value: FieldValues[F]): Stored {
  if (field === "server") return editServer(stored, value as FieldValues["server"]);
  if (field === "project" && value === "") throw new Error("A project must be a non-empty string");
  return editInherited(stored, field as Exclude<ScopedField, "server">, value);
}

/** "Use the default here": drops the site's own value. */
export function resetField(stored: Stored, field: ScopedField): Stored {
  if (field === "project" && stored.site.server !== undefined) throw new Error("Cannot reset the project to the default: this site uses its own dashboard, and a project belongs to its dashboard");
  return { global: { ...stored.global }, site: omit(stored.site, ownKeys(field)) };
}

/** "Make this the default for all sites": moves the site's value into the defaults and drops the override. */
export function promoteField(stored: Stored, field: ScopedField): Stored {
  const { global, site } = stored;
  if (field === "project" && site.server !== undefined) throw new Error("Cannot make the project the default for all sites: this site uses its own dashboard, and the dashboard is what gets promoted");
  if (field === "server") {
    if (site.server === undefined) throw new Error("Cannot make the server the default for all sites: this site has no server of its own");
    const project = site.project === undefined ? {} : { project: site.project };
    return { global: { ...omit(global, ["project"]), server: site.server, ...project }, site: omit(site, ownKeys("server")) };
  }
  const value = site[field];
  if (value === undefined) throw new Error(`Cannot make ${field} the default for all sites: this site has no ${field} of its own`);
  return { global: { ...global, [field]: value }, site: omit(site, [field]) };
}
