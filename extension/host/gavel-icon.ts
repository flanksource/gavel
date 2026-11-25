import gavelFill from "@iconify-icons/ph/gavel-fill";
import gavelLight from "@iconify-icons/ph/gavel-light";

/*
 * The grabber's gavel mark. The host is not React, so it takes the SVG markup of
 * the same Phosphor light/fill gavel clicky-ui selects (`UiGavel`/`UiGavelFilled`,
 * from icon-selections.json) from the offline icon data, with no runtime icon
 * name lookup. The extension build rasterizes `extensionIconSvg` for its icons.
 */

type IconData = { width?: number; height?: number; body: string };

function standaloneSvg(icon: IconData): string {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${icon.width ?? 256} ${icon.height ?? 256}">${icon.body}</svg>`;
}

/** The outline gavel: the toolbar button when the sidebar is closed. */
export const GAVEL_SVG = standaloneSvg(gavelLight);
/** The filled gavel: the FAB, and the toolbar button while the sidebar is open. */
export const GAVEL_FILLED_SVG = standaloneSvg(gavelFill);

/** Sizes of the manifest's `icons` and `action.default_icon`. */
export const EXTENSION_ICON_SIZES = [16, 32, 48, 128] as const;

const BRAND_BLUE = "#2563eb";

/** A white gavel on the grabber's blue, as a 128px square SVG the extension build rasterizes at each size. */
export function extensionIconSvg(): string {
  const scale = 88 / (gavelFill.width ?? 256);
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><rect width="128" height="128" rx="28" fill="${BRAND_BLUE}"/><g transform="translate(20 20) scale(${scale})">${gavelFill.body.replaceAll("currentColor", "#fff")}</g></svg>`;
}
