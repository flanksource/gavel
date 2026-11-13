import { copyFile, mkdir, rm } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { build } from "vite";
import { developmentAliases } from "./dependencies";
import { writeExtensionIcons } from "./extension-icons";
import { reactDevtoolsHookScript } from "./react-devtools-hook";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const args = process.argv.slice(2);
if (args.some(arg => arg !== "--dev")) throw new Error("Usage: build.ts [--dev]");
const alias = args.includes("--dev") ? developmentAliases(root) : [];
const outDir = resolve(root, "dist");
await rm(outDir, { recursive: true, force: true });
await mkdir(outDir, { recursive: true });

await build({
  configFile: false, root, publicDir: false, logLevel: "warn",
  define: { "process.env.NODE_ENV": JSON.stringify("production") },
  resolve: { alias },
  build: {
    outDir, emptyOutDir: false,
    lib: { entry: resolve(root, "src/grabber.ts"), name: "GavelGrabber", formats: ["iife"], fileName: () => "grabber.js" },
    rolldownOptions: { output: { codeSplitting: false } },
  },
});
const hook = `(function () {\n${reactDevtoolsHookScript(root)}\n})`;
for (const entry of ["background", "bridge", "early"]) {
  await build({
    configFile: false, root, publicDir: false, logLevel: "warn",
    ...(entry === "early" ? { define: { __INSTALL_REACT_DEVTOOLS_HOOK__: hook } } : {}),
    build: { outDir, emptyOutDir: false, lib: { entry: resolve(root, `src/${entry}.ts`), name: `GavelGrabber_${entry}`, formats: ["iife"], fileName: () => `${entry}.js` } },
  });
}
await build({
  configFile: false, root: resolve(root, "src"), base: "./", publicDir: false, logLevel: "warn",
  plugins: [react(), tailwindcss()],
  define: { "process.env.NODE_ENV": JSON.stringify("production") },
  resolve: { dedupe: ["react", "react-dom"], alias },
  build: {
    outDir, emptyOutDir: false,
    rolldownOptions: { input: Object.fromEntries(["popup", "panel", "todo-relay"].map(page => [page, resolve(root, `src/${page}.html`)])) },
  },
});
await copyFile(resolve(root, "src/manifest.json"), resolve(outDir, "manifest.json"));
await writeExtensionIcons(outDir);
console.log(`Gavel extension built. Load unpacked from ${outDir}`);
