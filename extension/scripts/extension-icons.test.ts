import { mkdir, mkdtemp, readFile, readdir, rm } from "node:fs/promises";
import { join } from "node:path";
import { afterEach, expect, it } from "vitest";
import { EXTENSION_ICON_SIZES } from "../host/gavel-icon";
import { writeExtensionIcons } from "./extension-icons";

let dir: string | undefined;
afterEach(async () => {
  if (dir) await rm(dir, { recursive: true, force: true });
  dir = undefined;
});

it("writes nonempty square PNGs at every manifest icon size", async () => {
  await mkdir(".tmp", { recursive: true });
  dir = await mkdtemp(".tmp/icons-");
  await writeExtensionIcons(dir);
  expect((await readdir(join(dir, "icons"))).sort()).toEqual(EXTENSION_ICON_SIZES.map(size => `icon-${size}.png`).sort());
  const sharp = (await import("sharp")).default;
  for (const size of EXTENSION_ICON_SIZES) {
    const bytes = await readFile(join(dir, "icons", `icon-${size}.png`));
    expect({ width: bytes.readUInt32BE(16), height: bytes.readUInt32BE(20) }).toEqual({ width: size, height: size });
    const { data, info } = await sharp(bytes).raw().toBuffer({ resolveWithObject: true });
    const colors = new Set<string>();
    for (let offset = 0; offset < data.length; offset += info.channels) colors.add(`${data[offset]},${data[offset + 1]},${data[offset + 2]}`);
    expect(colors.size).toBeGreaterThan(1);
  }
});
