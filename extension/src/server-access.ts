import { serverOrigin } from "./settings";

/**
 * Extension pages skip CORS only for hosts Chrome actually grants. When the
 * user (or Chrome's site-access default) withholds `<all_urls>`, requests to
 * the server fail as opaque CORS errors, so check the grant first and say so.
 */
export function serverAccessPattern(server: string): string {
  return `${serverOrigin(server)}/*`;
}

export function hasServerAccess(server: string): Promise<boolean> {
  return chrome.permissions.contains({ origins: [serverAccessPattern(server)] });
}

/** Must be called from a user gesture (a click in the popup). */
export function requestServerAccess(server: string): Promise<boolean> {
  return chrome.permissions.request({ origins: [serverAccessPattern(server)] });
}

export async function assertServerAccess(server: string): Promise<void> {
  if (await hasServerAccess(server)) return;
  throw new Error(`Chrome is withholding this extension's access to ${serverOrigin(server)}, so its requests fail with CORS errors. Click the extension icon and choose "Grant access", or set the extension's Site access to "On all sites" in chrome://extensions.`);
}
