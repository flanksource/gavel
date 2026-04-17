// gavel tsc linter wrapper.
// Invoked as `node <path>` from the project root. Uses the TypeScript
// compiler API (resolved from the project's own node_modules) to type-check
// the project and emits diagnostics as a JSON array on stdout.
//
// Exit codes: 0 on success (even when diagnostics are present), 1 on
// internal failure. Diagnostic count is encoded in the JSON payload, not the
// exit code, so gavel can distinguish "ran and found things" from "failed to
// run."

"use strict";

let ts;
try {
  ts = require("typescript");
} catch (err) {
  process.stderr.write(
    "gavel-tsc: cannot load 'typescript' from " +
      process.cwd() +
      "/node_modules. Install it with your package manager (e.g. `npm i -D typescript`).\n"
  );
  process.exit(1);
}

const configPath = ts.findConfigFile("./", ts.sys.fileExists, "tsconfig.json");
if (!configPath) {
  process.stderr.write("gavel-tsc: no tsconfig.json found in " + process.cwd() + "\n");
  process.exit(1);
}

const readResult = ts.readConfigFile(configPath, ts.sys.readFile);
if (readResult.error) {
  process.stderr.write(
    "gavel-tsc: failed to read " +
      configPath +
      ": " +
      ts.flattenDiagnosticMessageText(readResult.error.messageText, "\n") +
      "\n"
  );
  process.exit(1);
}

const parsed = ts.parseJsonConfigFileContent(
  readResult.config,
  ts.sys,
  "./",
  undefined,
  configPath
);

// Force no-emit: we only want diagnostics.
parsed.options.noEmit = true;

const program = ts.createProgram({
  rootNames: parsed.fileNames,
  options: parsed.options,
  projectReferences: parsed.projectReferences,
});

const diagnostics = ts
  .getPreEmitDiagnostics(program)
  .concat(parsed.errors || []);

const output = diagnostics.map((d) => {
  let line = 0;
  let character = 0;
  if (d.file && typeof d.start === "number") {
    const pos = d.file.getLineAndCharacterOfPosition(d.start);
    line = pos.line;
    character = pos.character;
  }
  return {
    file: d.file ? d.file.fileName : "",
    line: line + 1,
    column: character + 1,
    code: d.code,
    category: ts.DiagnosticCategory[d.category],
    message: ts.flattenDiagnosticMessageText(d.messageText, "\n"),
  };
});

process.stdout.write(JSON.stringify(output));
