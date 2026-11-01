Object.defineProperty(exports, Symbol.toStringTag, { value: "Module" });
const require_use_test_run = require("./hooks/use-test-run.cjs");
const require_use_process_trace = require("./hooks/use-process-trace.cjs");
exports.PROCESS_TRACE_ENDED = require_use_process_trace.PROCESS_TRACE_ENDED;
exports.SAMPLE_WINDOW = require_use_process_trace.SAMPLE_WINDOW;
exports.useProcessTrace = require_use_process_trace.useProcessTrace;
exports.useTestRun = require_use_test_run.useTestRun;
