// src/utils.ts
var TEST_TASK_GROUP_NAME = "Running tests across packages";
var LINT_TASK_GROUP_NAME = "Running linters";
function isVirtualTaskGroup(t, groupName) {
	return t.framework === "task" && t.name === groupName;
}
function stripTestTaskGroup(tests) {
	return tests.filter((t) => !isVirtualTaskGroup(t, TEST_TASK_GROUP_NAME));
}
function taskStatus(t) {
	return t.context?.status;
}
function isLive(t) {
	const status = taskStatus(t);
	return status === "running" || status === "pending";
}
function hasLiveTestGroup(tests) {
	const testGroup = tests.find((t) => isVirtualTaskGroup(t, TEST_TASK_GROUP_NAME));
	return !!testGroup && isLive(testGroup);
}
function isLintOnlyPhase(tests, running) {
	return running && !hasLiveTestGroup(tests);
}
function isLintOnlyRun(tests, lintRun) {
	return lintRun && !hasLiveTestGroup(tests) && tests.every((t) => t.framework === "task");
}
function lintStatusCounts(lint) {
	const counts = {
		running: 0,
		done: 0,
		failed: 0
	};
	for (const r of lint ?? []) if (r.status) counts.running++;
	else if (!r.skipped && (r.error || r.timed_out || !r.success && !r.violations?.length)) counts.failed++;
	else counts.done++;
	return counts;
}
function liveTaskGroups(tests) {
	return tests.filter((t) => t.framework === "task" && t.name !== "Running tests across packages" && t.name !== "Running linters" && isLive(t)).map((t) => ({
		...t,
		children: (t.children ?? []).filter(isLive)
	}));
}
function timeoutArgValue(args, key) {
	if (!args) return null;
	const raw = args[key];
	if (typeof raw !== "string") return null;
	const trimmed = raw.trim();
	return trimmed === "" ? null : trimmed;
}
function processLinksByPid(tests) {
	const links = /* @__PURE__ */ new Map();
	const walk = (list) => {
		for (const test of list) {
			for (const pid of test.fixture?.pids || []) if (!links.has(pid) && test.route_path) links.set(pid, {
				pid,
				routePath: test.route_path,
				name: test.name
			});
			if (test.children) walk(test.children);
		}
	};
	walk(tests);
	return links;
}
exports.LINT_TASK_GROUP_NAME = LINT_TASK_GROUP_NAME;
exports.TEST_TASK_GROUP_NAME = TEST_TASK_GROUP_NAME;
exports.isLintOnlyPhase = isLintOnlyPhase;
exports.isLintOnlyRun = isLintOnlyRun;
exports.lintStatusCounts = lintStatusCounts;
exports.liveTaskGroups = liveTaskGroups;
exports.processLinksByPid = processLinksByPid;
exports.stripTestTaskGroup = stripTestTaskGroup;
exports.timeoutArgValue = timeoutArgValue;

//# sourceMappingURL=utils.cjs.map