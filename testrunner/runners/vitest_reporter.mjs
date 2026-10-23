// Gavel live-progress reporter for Vitest 3+. Each hook writes one
// GAVEL_VITEST_V1:<json> line to stdout, decoded by parsers.VitestStreamWriter.
// The json reporter's --outputFile remains the authoritative result.
const PREFIX = "GAVEL_VITEST_V1:";

function emit(event) {
  process.stdout.write(PREFIX + JSON.stringify(event) + "\n");
}

function suitePath(testCase) {
  const names = [];
  for (let node = testCase.parent; node && node.type === "suite"; node = node.parent) {
    names.unshift(node.name);
  }
  return names;
}

function identify(testCase) {
  return {
    file: testCase.module.moduleId,
    name: testCase.name,
    suite: suitePath(testCase),
    line: testCase.location?.line ?? 0,
  };
}

function messages(errors) {
  return (errors ?? []).map((e) => e.stack || e.message || String(e));
}

export default class GavelReporter {
  onTestModuleStart(testModule) {
    emit({ event: "module-start", file: testModule.moduleId });
  }

  onTestCaseReady(testCase) {
    emit({ event: "test-ready", ...identify(testCase) });
  }

  onTestCaseResult(testCase) {
    const result = testCase.result();
    emit({
      event: "test-result",
      ...identify(testCase),
      state: result.state,
      duration: testCase.diagnostic()?.duration ?? 0,
      errors: messages(result.errors),
    });
  }

  onTestModuleEnd(testModule) {
    emit({
      event: "module-end",
      file: testModule.moduleId,
      state: testModule.state(),
      errors: messages(testModule.errors()),
    });
  }
}
