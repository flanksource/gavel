"use strict";

const path = require("path");

const sys = {
  fileExists: () => true,
  readFile: () => "{}",
  writeFile: (fileName) => require("fs").writeFileSync(fileName, "output"),
};

module.exports = {
  sys,
  DiagnosticCategory: { 1: "Error" },
  ExitStatus: { Success: 0 },
  findConfigFile: () => "tsconfig.json",
  readConfigFile: () => ({ config: {} }),
  parseJsonConfigFileContent: () => ({
    fileNames: [],
    options: {},
    projectReferences: [{ path: "packages/app" }],
    errors: [],
  }),
  createProgram: () => ({}),
  getPreEmitDiagnostics: () => [],
  createSolutionBuilderHost: (system, _createProgram, reportDiagnostic) => ({
    writeFile: system.writeFile,
    reportDiagnostic,
  }),
  createSolutionBuilder: (host, roots, options) => ({
    build: () => {
      if (roots.length !== 1 || roots[0] !== "tsconfig.json") {
        throw new Error(`unexpected roots: ${JSON.stringify(roots)}`);
      }
      if (!options.noEmit || !options.force) {
        throw new Error(`expected noEmit and force, got ${JSON.stringify(options)}`);
      }
      host.writeFile(path.join("packages", "app", "dist", "index.js"), "output");
      host.reportDiagnostic({
        file: {
          fileName: "packages/app/src/index.ts",
          getLineAndCharacterOfPosition: () => ({ line: 0, character: 0 }),
        },
        start: 0,
        code: 2322,
        category: 1,
        messageText: "Type 'string' is not assignable to type 'number'.",
      });
      return 1;
    },
  }),
  flattenDiagnosticMessageText: (message) => message,
};
