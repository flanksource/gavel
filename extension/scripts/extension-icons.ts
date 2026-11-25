import { mkdir } from "node:fs/promises";
import { resolve } from "node:path";
import sharp from "sharp";
import { EXTENSION_ICON_SIZES, extensionIconSvg } from "../host/gavel-icon";

/**
 * Writes `icons/icon-<size>.png` into the extension build, rasterized from the
 * same gavel SVG the FAB and toolbar button use. The SVG is a 128px square, so
 * each size is rendered at its own density rather than downscaled, which keeps
 * the 16px toolbar icon crisp.
 */
export async function writeExtensionIcons(outDir: string): Promise<void> {
  const iconsDir = resolve(outDir, "icons");
  await mkdir(iconsDir, { recursive: true });
  const svg = Buffer.from(extensionIconSvg());
  for (const size of EXTENSION_ICON_SIZES) {
    await sharp(svg, { density: (72 * size) / 128 }).resize(size, size).png().toFile(resolve(iconsDir, `icon-${size}.png`));
  }
}
