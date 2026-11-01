const require_utils = require("./utils.cjs");
const require_use_test_run = require("./hooks/use-test-run.cjs");
const require_config = require("./config.cjs");
const require_traces = require("./traces.cjs");
const require_use_process_trace = require("./hooks/use-process-trace.cjs");
const require_processMetrics = require("./processMetrics.cjs");
const require_routes = require("./routes.cjs");
const require_index = require("./adapters/index.cjs");
const require_RunProcesses = require("./components/RunProcesses.cjs");
const require_BenchView = require("./components/BenchView.cjs");
const require_FixtureBenchView = require("./components/FixtureBenchView.cjs");
const require_RerunDialog = require("./components/RerunDialog.cjs");
const require_RunActions = require("./components/RunActions.cjs");
const require_lintRerun = require("./lintRerun.cjs");
const require_LintTasks = require("./components/LintTasks.cjs");
const require_RunHistory = require("./components/RunHistory.cjs");
const require_RunSwitcher = require("./components/RunSwitcher.cjs");
const require_use_run_history = require("./hooks/use-run-history.cjs");
const require_exportPaths = require("./exportPaths.cjs");
let react = require("react");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let react_jsx_runtime = require("react/jsx-runtime");
let _flanksource_clicky_ui_components = require("@flanksource/clicky-ui/components");
let _flanksource_clicky_ui_icons = require("@flanksource/clicky-ui/icons");
let _flanksource_clicky_ui_file_findings = require("@flanksource/clicky-ui/file-findings");
// src/App.tsx
function countFixtureWork(tests) {
	return tests.reduce((count, test) => count + (test.fixture ? 1 : 0) + countFixtureWork(test.children || []), 0);
}
var asGavel = (tests) => tests;
var FilesTab = (0, react.lazy)(() => Promise.resolve().then(() => require("./FilesTab.cjs")));
var PromptRunView = (0, react.lazy)(() => Promise.resolve().then(() => require("./PromptRunView.cjs")));
function renderPromptRun(result) {
	if (!result.run_id) return null;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)(react.Suspense, {
		fallback: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("p", {
			role: "status",
			className: "text-sm text-muted-foreground",
			children: "Loading the prompt run…"
		}),
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(PromptRunView, {
			result,
			baseUrl: require_config.apiBase()
		})
	});
}
function FilesLoading() {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Icon, {
			icon: _flanksource_clicky_ui_icons.UiLoader,
			className: "mr-2 animate-spin text-blue-500"
		}), "Loading files..."]
	});
}
function withWorkDir(tests, root) {
	if (!root) return tests;
	return tests.map((test) => ({
		...test,
		work_dir: test.work_dir || root,
		...test.children ? { children: withWorkDir(test.children, root) } : {}
	}));
}
function findingKey(finding) {
	return [
		finding.workDir,
		finding.file,
		finding.line,
		finding.rule,
		finding.message
	].join("\0");
}
function sourceUrl(url, line) {
	if (!url) return void 0;
	if (!line) return url;
	const parsed = new URL(url);
	if (!parsed.searchParams.has("line")) parsed.searchParams.set("line", String(line));
	return parsed.toString();
}
async function fetchFile(ref, scope) {
	const params = new URLSearchParams();
	if (ref.workDir) params.set("work_dir", ref.workDir);
	params.set("path", ref.file);
	const res = await fetch(require_config.apiUrl(`/api/file?${params}`, scope));
	const body = await res.text();
	if (!res.ok) throw new Error(`${res.status} ${body.trim() || res.statusText}`);
	return body;
}
function withOwnNames(tests) {
	return tests.map((test) => ({
		...test,
		own_name: test.name,
		...test.children ? { children: withOwnNames(test.children) } : {}
	}));
}
function rerunBody(node) {
	const paths = /* @__PURE__ */ new Set();
	const workDirs = /* @__PURE__ */ new Set();
	const collect = (n) => {
		if (n.package_path) paths.add(n.package_path);
		if (n.work_dir) workDirs.add(n.work_dir);
		(n.children || []).forEach(collect);
	};
	collect(node);
	if (workDirs.size > 1) return { error: "Rerun across multiple module roots is not supported" };
	const isLeaf = !node.children || node.children.length === 0;
	return { body: {
		package_paths: Array.from(paths),
		work_dir: workDirs.size === 1 ? Array.from(workDirs)[0] : "",
		test_name: isLeaf ? node.own_name ?? node.name : "",
		suite: node.suite || [],
		framework: node.framework || ""
	} };
}
function PageHeader({ icon, title, status, actions }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
		"data-page-header": "",
		children: [
			/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("h1", {
				className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
				children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Icon, {
					icon,
					className: "text-primary"
				}), title]
			}),
			status && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
				className: "truncate text-sm text-muted-foreground",
				children: status
			}),
			actions && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: actions
			})
		]
	});
}
function Waiting({ done }) {
	return /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground",
		children: done ? "No tests were run" : /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", { children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.Icon, {
			icon: _flanksource_clicky_ui_icons.UiLoader,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("p", {
			className: "mt-2",
			children: "Waiting for test results..."
		})] })
	});
}
var NOTICE_MS = 8e3;
var GLOBAL_STOP_KEY = "__global__";
var loadPage = (href) => window.location.assign(href);
function App({ runId, navigate = loadPage } = {}) {
	const scope = (0, react.useMemo)(() => runId ? { runId } : {}, [runId]);
	const base = require_config.apiBase(scope);
	const run = require_use_test_run.useTestRun(scope.runId ? { runId: scope.runId } : {});
	const { snapshot, tests, lint, status: snapshotStatus, runMeta, done, statusText, error: streamError } = run;
	const performance = snapshot?.performance;
	const runError = snapshot?.error;
	const lintRun = !!snapshotStatus.lint_run;
	const history = require_use_run_history.useRunHistory();
	const [route, setRoute] = (0, react.useState)(() => require_routes.parseRoute(window.location));
	const [expandAll, setExpandAll] = (0, react.useState)(null);
	const [rerunBusy, setRerunBusy] = (0, react.useState)(false);
	const [rerunDialogOpen, setRerunDialogOpen] = (0, react.useState)(false);
	const [ignoreBusy, setIgnoreBusy] = (0, react.useState)(false);
	const [testEditBusy, setTestEditBusy] = (0, react.useState)(false);
	const [stopBusyKey, setStopBusyKey] = (0, react.useState)(null);
	const [notice, setNotice] = (0, react.useState)(void 0);
	const noticeTimer = (0, react.useRef)(null);
	const firstSeen = (0, react.useRef)(Date.now());
	const [now, setNow] = (0, react.useState)(Date.now());
	const say = (0, react.useCallback)((message) => {
		setNotice(message);
		if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current);
		noticeTimer.current = window.setTimeout(() => setNotice(void 0), NOTICE_MS);
	}, []);
	(0, react.useEffect)(() => () => {
		if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current);
	}, []);
	(0, react.useEffect)(() => {
		if (done) return;
		const timer = window.setInterval(() => setNow(Date.now()), 1e3);
		return () => window.clearInterval(timer);
	}, [done]);
	const startTime = runMeta?.started ? Date.parse(runMeta.started) : firstSeen.current;
	const endTime = runMeta?.ended ? Date.parse(runMeta.ended) : null;
	const commitRoute = (0, react.useCallback)((next, mode = "push") => {
		setRoute(next);
		const url = require_routes.buildRoute(next);
		if (url === `${window.location.pathname}${window.location.search}`) return;
		if (mode === "replace") window.history.replaceState({}, "", url);
		else window.history.pushState({}, "", url);
	}, []);
	(0, react.useEffect)(() => {
		const onPopState = () => setRoute(require_routes.parseRoute(window.location));
		window.addEventListener("popstate", onPopState);
		return () => window.removeEventListener("popstate", onPopState);
	}, []);
	(0, react.useEffect)(() => {
		if (performance && route.tab === "tests" && window.location.pathname === `${require_config.apiBase()}/`) commitRoute({
			...route,
			tab: "bench",
			selectedPath: ""
		}, "replace");
	}, [
		performance,
		route,
		commitRoute
	]);
	const displayedTests = (0, react.useMemo)(() => require_utils.isLintOnlyPhase(tests, !!snapshotStatus.running) ? require_utils.stripTestTaskGroup(tests) : tests, [tests, snapshotStatus.running]);
	const totals = (0, react.useMemo)(() => (0, _flanksource_clicky_ui_data.aggregateStatusCounts)(displayedTests), [displayedTests]);
	const filteredTests = (0, react.useMemo)(() => (0, _flanksource_clicky_ui_data.annotateRoutePaths)((0, _flanksource_clicky_ui_data.collapseSingleChildChains)(require_exportPaths.withExportPaths((0, _flanksource_clicky_ui_data.filterTests)(withOwnNames(displayedTests), route.filters.status, route.filters.framework)))), [displayedTests, route.filters]);
	const selectedTest = (0, react.useMemo)(() => route.tab === "tests" && route.selectedPath ? (0, _flanksource_clicky_ui_data.findNodeByRoutePath)(filteredTests, route.selectedPath) : null, [
		route.tab,
		route.selectedPath,
		filteredTests
	]);
	const lintTree = (0, react.useMemo)(() => (0, _flanksource_clicky_ui_data.buildLintTree)({
		results: lint ?? [],
		grouping: route.lintGrouping,
		filters: route.lintFilters
	}), [
		lint,
		route.lintGrouping,
		route.lintFilters
	]);
	const selectedLint = route.tab === "lint" && route.selectedPath ? (0, _flanksource_clicky_ui_data.findLintNode)(lintTree, route.selectedPath) : null;
	const lintTotal = (0, react.useMemo)(() => (0, _flanksource_clicky_ui_data.totalLintViolations)(lint ?? []), [lint]);
	const gitRoot = snapshot?.git?.root;
	const { findings, linterByFinding, sourceUrlByFinding } = (0, react.useMemo)(() => {
		const linterByFinding = /* @__PURE__ */ new Map();
		const sourceUrlByFinding = /* @__PURE__ */ new Map();
		return {
			findings: [...(lint ?? []).flatMap((result) => {
				const workDir = result.work_dir || gitRoot;
				return (0, _flanksource_clicky_ui_file_findings.lintFindings)([{
					...result,
					...workDir ? { work_dir: workDir } : {}
				}]).map((finding) => {
					linterByFinding.set(findingKey(finding), result.linter);
					const url = sourceUrl(result.sources?.[finding.file]?.url, finding.line);
					if (url) sourceUrlByFinding.set(findingKey(finding), url);
					return finding;
				});
			}), ...(0, _flanksource_clicky_ui_file_findings.testFindings)(withWorkDir(tests, gitRoot))],
			linterByFinding,
			sourceUrlByFinding
		};
	}, [
		lint,
		tests,
		gitRoot
	]);
	const selectedFile = (0, react.useMemo)(() => route.tab === "files" && route.selectedPath ? {
		file: route.selectedPath,
		...route.workDir ? { workDir: route.workDir } : {}
	} : null, [
		route.tab,
		route.selectedPath,
		route.workDir
	]);
	const traces = (0, react.useMemo)(() => require_traces.rebaseTraces(snapshotStatus.traces, base), [snapshotStatus.traces, base]);
	const processTraceAvailable = !!require_traces.processMount(snapshotStatus.traces);
	const trace = require_use_process_trace.useProcessTrace({
		traces: snapshotStatus.traces,
		done,
		base,
		...route.tab === "processes" && route.selectedPath ? { selectedKey: route.selectedPath } : {}
	});
	const processCount = (0, react.useMemo)(() => trace.nodes.filter((node) => !(0, _flanksource_clicky_ui_data.isProcessExited)(node)).length, [trace.nodes]);
	const rootProcess = (0, react.useMemo)(() => trace.nodes.find((node) => node.isRoot) ?? null, [trace.nodes]);
	const selectedProcess = route.tab === "processes" ? trace.nodes.find((node) => node.key === route.selectedPath) ?? null : null;
	const processLinks = (0, react.useMemo)(() => processTraceAvailable ? require_utils.processLinksByPid(asGavel((0, _flanksource_clicky_ui_data.annotateRoutePaths)((0, _flanksource_clicky_ui_data.collapseSingleChildChains)(displayedTests)))) : /* @__PURE__ */ new Map(), [processTraceAvailable, displayedTests]);
	(0, react.useEffect)(() => {
		if (route.tab !== "processes" || route.selectedPath || !rootProcess) return;
		commitRoute({
			...route,
			selectedPath: rootProcess.key
		}, "replace");
	}, [
		route,
		rootProcess,
		commitRoute
	]);
	(0, react.useEffect)(() => {
		if (!route.selectedPath) return;
		const ready = route.tab === "tests" ? tests.length > 0 || done : route.tab === "lint" ? lintRun || done || !!lint : route.tab === "processes" ? trace.nodes.length > 0 : true;
		const found = route.tab === "tests" ? selectedTest : route.tab === "lint" ? selectedLint : route.tab === "processes" ? selectedProcess : true;
		if (!ready || found) return;
		commitRoute({
			...route,
			selectedPath: route.tab === "processes" && rootProcess ? rootProcess.key : ""
		}, "replace");
	}, [
		route,
		selectedTest,
		selectedLint,
		selectedProcess,
		tests.length,
		done,
		lintRun,
		lint,
		trace.nodes.length,
		rootProcess,
		commitRoute
	]);
	const onTabChange = (0, react.useCallback)((tab) => {
		const { workDir: _workDir, line: _line, ...rest } = route;
		commitRoute({
			...rest,
			tab,
			selectedPath: ""
		});
	}, [route, commitRoute]);
	const onSelectTest = (0, react.useCallback)((node) => commitRoute({
		...route,
		selectedPath: node?.route_path ?? ""
	}), [route, commitRoute]);
	const onFiltersChange = (0, react.useCallback)((filters) => commitRoute({
		...route,
		tab: "tests",
		filters,
		selectedPath: ""
	}), [route, commitRoute]);
	const onLintFiltersChange = (0, react.useCallback)((lintFilters) => commitRoute({
		...route,
		tab: "lint",
		lintFilters,
		selectedPath: ""
	}), [route, commitRoute]);
	const onLintGroupingChange = (0, react.useCallback)((lintGrouping) => commitRoute({
		...route,
		tab: "lint",
		lintGrouping,
		selectedPath: ""
	}), [route, commitRoute]);
	const onSelectLint = (0, react.useCallback)((selectedPath) => commitRoute({
		...route,
		selectedPath: selectedPath ?? ""
	}), [route, commitRoute]);
	const onSelectProcess = (0, react.useCallback)((key) => commitRoute({
		...route,
		tab: "processes",
		selectedPath: key
	}), [route, commitRoute]);
	const onHideExitedChange = (0, react.useCallback)((hideExited) => commitRoute({
		...route,
		tab: "processes",
		hideExited
	}, "replace"), [route, commitRoute]);
	const openFile = (0, react.useCallback)((ref, line) => {
		const { workDir: _workDir, line: _line, ...rest } = route;
		commitRoute({
			...rest,
			tab: "files",
			selectedPath: ref.file,
			...ref.workDir ? { workDir: ref.workDir } : {},
			...line ? { line } : {}
		});
	}, [route, commitRoute]);
	const onOpenFile = (0, react.useCallback)((target) => openFile({
		file: target.file,
		...target.workDir || gitRoot ? { workDir: target.workDir || gitRoot } : {}
	}, target.line), [openFile, gitRoot]);
	const onSelectFile = (0, react.useCallback)((ref) => openFile(ref), [openFile]);
	const onLineChange = (0, react.useCallback)((line) => {
		const { line: _line, ...rest } = route;
		commitRoute({
			...rest,
			...line ? { line } : {}
		}, "replace");
	}, [route, commitRoute]);
	const onViewChange = (0, react.useCallback)((view) => commitRoute({
		...route,
		view
	}, "replace"), [route, commitRoute]);
	const loadFile = (0, react.useCallback)((ref) => fetchFile(ref, scope), [scope]);
	const onSampleChange = (0, react.useCallback)((sampleIndex) => commitRoute({
		...route,
		sampleIndex
	}), [route, commitRoute]);
	const onOpenStep = (0, react.useCallback)((link) => {
		commitRoute({
			...route,
			tab: "tests",
			filters: {
				status: /* @__PURE__ */ new Map(),
				framework: /* @__PURE__ */ new Map()
			},
			selectedPath: link.routePath
		});
	}, [route, commitRoute]);
	const startRerun = (0, react.useCallback)(async (body, label) => {
		if (rerunBusy) return;
		setRerunBusy(true);
		setRerunDialogOpen(true);
		say(`Rerunning ${label}...`);
		try {
			const result = await run.rerun(body);
			if (result.conflict) say("Another rerun is in progress");
			else if (!result.ok) say(`Rerun failed: ${result.message ?? result.status}`);
		} catch (e) {
			say(`Rerun error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			setRerunBusy(false);
		}
	}, [
		rerunBusy,
		run,
		say
	]);
	const onRerun = (0, react.useCallback)((node) => {
		if (node.framework === "task") return;
		const request = rerunBody(node);
		if ("error" in request) say(request.error);
		else startRerun(request.body, node.name);
	}, [startRerun, say]);
	const onRerunLint = (0, react.useCallback)(() => {
		const body = require_lintRerun.buildLintRerunBody(selectedLint, lint ?? []);
		if (!body.work_dir) say("Rerun across multiple lint roots is not supported");
		else startRerun(body, selectedLint?.name ?? "lint");
	}, [
		selectedLint,
		lint,
		startRerun,
		say
	]);
	const stop = (0, react.useCallback)(async ({ taskId, label }) => {
		if (stopBusyKey) return;
		setStopBusyKey(taskId ?? GLOBAL_STOP_KEY);
		say(`Stopping ${label}...`);
		try {
			const result = await run.stop(taskId);
			if (!result.ok) say(`Stop failed: ${result.message ?? result.status}`);
		} catch (e) {
			say(`Stop error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			setStopBusyKey(null);
		}
	}, [
		stopBusyKey,
		run,
		say
	]);
	const onStop = (0, react.useCallback)((node) => {
		if (node.task_id) stop({
			taskId: node.task_id,
			label: node.name
		});
	}, [stop]);
	const onStopAll = (0, react.useCallback)(() => {
		stop({ label: "the run" });
	}, [stop]);
	const onIgnore = (0, react.useCallback)(async (action) => {
		if (ignoreBusy) return;
		setIgnoreBusy(true);
		try {
			const res = await fetch(require_config.apiUrl("/api/lint/ignore", scope), {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(action.req)
			});
			if (!res.ok) {
				say(`Ignore failed: ${(await res.text()).trim()}`);
				return;
			}
			say("Ignore rule saved to .gavel.yaml");
			if (route.tab !== "files") commitRoute({
				...route,
				selectedPath: ""
			}, "replace");
		} catch (e) {
			say(`Ignore error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			setIgnoreBusy(false);
		}
	}, [
		ignoreBusy,
		scope,
		route,
		commitRoute,
		say
	]);
	const renderFindingActions = (0, react.useCallback)((finding) => {
		if (finding.kind !== "lint") return null;
		const linter = linterByFinding.get(findingKey(finding));
		if (!linter) throw new Error(`no linter recorded for the finding at ${finding.file}:${finding.line}`);
		const violation = {
			...finding.rule ? { rule: { method: finding.rule } } : {},
			...finding.line ? { line: finding.line } : {}
		};
		const ignores = (0, _flanksource_clicky_ui_data.violationIgnoreActions)({
			violation,
			linter,
			file: finding.file,
			workDir: finding.workDir ?? ""
		});
		const source = sourceUrlByFinding.get(findingKey(finding));
		if (ignores.length === 0 && !source) return null;
		return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [source && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.Button, {
			asChild: true,
			size: "sm",
			variant: "outline",
			className: "h-6 gap-1 px-2 text-xs text-muted-foreground",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("a", {
				href: source,
				target: "_blank",
				rel: "noreferrer",
				title: `Open ${finding.file}:${finding.line ?? ""} at its source`,
				children: "Open source"
			})
		}), ignores.length > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.DropdownMenu, {
			label: "Ignore",
			icon: _flanksource_clicky_ui_icons.UiEyeClosed,
			variant: "outline",
			size: "sm",
			align: "right",
			title: "Add an ignore rule to .gavel.yaml",
			menuLabel: "Ignore scope",
			items: ignores.map((action) => ({
				label: action.label,
				title: action.title,
				disabled: ignoreBusy,
				onSelect: () => void onIgnore(action)
			}))
		})] });
	}, [
		linterByFinding,
		sourceUrlByFinding,
		ignoreBusy,
		onIgnore
	]);
	const onTestEdit = (0, react.useCallback)(async (node, action, editScope) => {
		if (testEditBusy || snapshotStatus.test_edit_supported !== true) return;
		const target = editScope === "file" ? node.file || "file" : node.name || "test";
		setTestEditBusy(true);
		say(action === "skip" ? `Skipping ${target}...` : `Deleting ${target}...`);
		try {
			const result = await run.editTest({
				action,
				scope: editScope,
				framework: node.framework || "",
				work_dir: node.work_dir || "",
				package_path: node.package_path || "",
				file: node.file || "",
				line: node.line || 0,
				test_name: node.name || "",
				suite: node.suite || []
			});
			if (!result.ok) {
				say(`Test edit failed: ${result.message ?? result.status}`);
				return;
			}
			say(result.message || `${action === "skip" ? "Skip" : "Delete"} ${editScope === "file" ? "file" : "test"} saved`);
			if (action === "delete") commitRoute({
				...route,
				selectedPath: ""
			}, "replace");
		} catch (e) {
			say(`Test edit error: ${e instanceof Error ? e.message : String(e)}`);
		} finally {
			setTestEditBusy(false);
		}
	}, [
		testEditBusy,
		snapshotStatus.test_edit_supported,
		run,
		route,
		commitRoute,
		say
	]);
	const wholeResultRoute = (0, react.useMemo)(() => ({
		...route,
		selectedPath: ""
	}), [route]);
	const actions = (0, react.useMemo)(() => ({
		nodeRouteState: (node) => node.export_path ? {
			...route,
			tab: "tests",
			selectedPath: node.export_path
		} : void 0,
		failingOnlyRouteState: (node) => node.export_path ? {
			...route,
			tab: "tests",
			selectedPath: node.export_path,
			filters: {
				...route.filters,
				status: /* @__PURE__ */ new Map([["failed", "include"], ["timedout", "include"]])
			}
		} : void 0,
		testEditSupported: snapshotStatus.test_edit_supported === true,
		onTestEdit,
		testEditBusy
	}), [
		route,
		snapshotStatus.test_edit_supported,
		onTestEdit,
		testEditBusy
	]);
	const registry = (0, react.useMemo)(() => require_index.gavelRegistry({
		traces,
		processes: trace.nodes,
		onOpenProcess: onSelectProcess,
		actions
	}), [
		traces,
		trace.nodes,
		onSelectProcess,
		actions
	]);
	const fetchProcessMetrics = (0, react.useMemo)(() => require_processMetrics.fetchProcessMetricsFromSamples({
		samples: trace.samples,
		nodes: trace.nodes
	}), [trace.samples, trace.nodes]);
	const showLintTab = lintRun;
	const showFilesTab = findings.length > 0 || route.tab === "files";
	const showBenchTab = !!run.bench || !!performance;
	const showProcessesTab = processTraceAvailable;
	const benchRegressions = run.bench?.deltas?.filter((d) => d.significant && d.delta_pct > run.bench.threshold).length || 0;
	const canExport = route.tab === "tests" && displayedTests.length > 0 || route.tab === "lint" && lintRun || route.tab === "bench" && showBenchTab;
	const canGlobalStop = snapshotStatus.running && !!snapshotStatus.stop_supported;
	const runActions = /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_RunActions.RunActions, {
		runMeta,
		canStop: canGlobalStop,
		stopping: stopBusyKey === GLOBAL_STOP_KEY,
		onStop: onStopAll,
		exportState: canExport ? wholeResultRoute : void 0
	});
	const status = notice ?? streamError ?? statusText;
	const backTo = window.__gavelBackTo;
	const tabs = [
		{
			id: "tests",
			label: "Tests",
			icon: _flanksource_clicky_ui_icons.UiBeaker,
			count: totals.failed > 0 ? totals.failed : totals.total,
			countColor: totals.failed > 0 ? "bg-red-500" : "bg-muted-foreground"
		},
		...showLintTab ? [{
			id: "lint",
			label: "Lint",
			icon: _flanksource_clicky_ui_icons.UiListChecks,
			count: lintTotal,
			countColor: lintTotal > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...showFilesTab ? [{
			id: "files",
			label: "Files",
			icon: _flanksource_clicky_ui_icons.UiFileCode,
			count: findings.length,
			countColor: findings.length > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...showBenchTab ? [{
			id: "bench",
			label: "Bench",
			icon: _flanksource_clicky_ui_icons.UiChartLineUp,
			count: performance ? countFixtureWork(tests) : benchRegressions > 0 ? benchRegressions : run.bench?.deltas?.length || 0,
			countColor: benchRegressions > 0 ? "bg-red-500" : "bg-muted-foreground"
		}] : [],
		...showProcessesTab ? [{
			id: "processes",
			label: "Processes",
			icon: _flanksource_clicky_ui_icons.UiServerProcess,
			count: processCount,
			countColor: "bg-blue-500"
		}] : [],
		...history.available ? [{
			id: "history",
			label: "History",
			icon: _flanksource_clicky_ui_icons.UiHistory
		}] : []
	];
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "flex h-screen flex-col bg-background text-foreground",
		children: [
			backTo && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "flex shrink-0 items-center gap-2 border-b border-border bg-card px-4 py-1.5 text-sm",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("a", {
					href: backTo,
					className: "inline-flex items-center gap-1 text-primary no-underline hover:underline",
					children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_icons.UiArrowLeft, {}), "Back to PR"]
				})
			}),
			(tabs.length > 1 || history.available) && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "flex shrink-0 items-center gap-2 pr-density-4",
				children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_components.Tabs, {
					tabs,
					value: route.tab,
					onChange: onTabChange,
					className: "min-w-0 flex-1 px-density-4"
				}), history.available && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_RunSwitcher.RunSwitcher, {
					history,
					navigate
				})]
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "min-h-0 flex-1",
				children: [
					route.tab === "tests" && (displayedTests.length === 0 ? /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(PageHeader, {
							icon: _flanksource_clicky_ui_icons.UiBeaker,
							title: "Test Results",
							status,
							actions: runActions
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(Waiting, { done })
						})]
					}) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.TestRunner, {
						tests: filteredTests,
						allTests: displayedTests,
						selected: selectedTest,
						filters: route.filters,
						expandAll,
						done,
						...runMeta ? { runMeta } : {},
						status: snapshotStatus,
						statusText: filteredTests.length === 0 ? `${status} · No tests match the current filters` : status,
						now,
						startTime,
						endTime,
						busy: {
							rerun: rerunBusy,
							stop: stopBusyKey !== null
						},
						onSelect: onSelectTest,
						onFiltersChange,
						onExpandAllChange: setExpandAll,
						onRerun,
						onStop,
						fetchProcessMetrics,
						adapters: registry,
						onOpenFile,
						headerSlot: runActions
					})),
					route.tab === "lint" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_LintTasks.LintTasks, { tests }), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.LintResults, {
								results: lint ?? [],
								grouping: route.lintGrouping,
								onGroupingChange: onLintGroupingChange,
								filters: route.lintFilters,
								onFiltersChange: onLintFiltersChange,
								selected: selectedLint ? route.selectedPath : null,
								onSelect: onSelectLint,
								expandAll,
								onExpandAllChange: setExpandAll,
								onIgnore,
								ignoreBusy,
								onOpenFile,
								renderRun: renderPromptRun,
								headerSlot: /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [
									/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
										className: "truncate text-sm text-muted-foreground",
										children: status
									}),
									/* @__PURE__ */ (0, react_jsx_runtime.jsxs)(_flanksource_clicky_ui_components.Button, {
										type: "button",
										size: "sm",
										disabled: rerunBusy,
										onClick: onRerunLint,
										title: selectedLint ? "Rerun the linters for this selection" : "Rerun the linters",
										children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_icons.UiRefresh, { className: "mr-1" }), rerunBusy ? "Running..." : "Rerun"]
									}),
									runActions
								] })
							})
						})]
					}),
					route.tab === "files" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(PageHeader, {
							icon: _flanksource_clicky_ui_icons.UiFileCode,
							title: "Files",
							status,
							actions: runActions
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(react.Suspense, {
								fallback: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(FilesLoading, {}),
								children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(FilesTab, {
									findings,
									selectedFile,
									onSelectFile,
									line: route.line,
									onLineChange,
									view: route.view ?? "preview",
									onViewChange,
									loadFile,
									renderFindingActions
								})
							})
						})]
					}),
					route.tab === "bench" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(PageHeader, {
							icon: _flanksource_clicky_ui_icons.UiChartLineUp,
							title: performance ? "Fixture Benchmark" : "Benchmark Comparison",
							status,
							actions: runActions
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "min-h-0 flex-1 overflow-auto",
							children: performance ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_FixtureBenchView.FixtureBenchView, {
								performance,
								tests,
								error: runError,
								sampleIndex: route.sampleIndex,
								onSampleChange
							}) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_BenchView.BenchView, { bench: run.bench })
						})]
					}),
					route.tab === "history" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(PageHeader, {
							icon: _flanksource_clicky_ui_icons.UiHistory,
							title: "Run History",
							status: history.current ? `Viewing ${history.current}` : "Viewing the latest results"
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "flex min-h-0 flex-1 flex-col p-density-2",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_RunHistory.RunHistory, {
								history,
								navigate
							})
						})]
					}),
					route.tab === "processes" && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(PageHeader, {
							icon: _flanksource_clicky_ui_icons.UiServerProcess,
							title: "Processes",
							status,
							actions: runActions
						}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_RunProcesses.RunProcesses, {
								trace,
								selectedKey: selectedProcess?.key ?? null,
								onSelect: onSelectProcess,
								hideExited: !!route.hideExited,
								onHideExitedChange,
								expandAll,
								onExpandAllChange: setExpandAll,
								stepLinks: processLinks,
								onOpenStep
							})
						})]
					})
				]
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_RerunDialog.RerunDialog, {
				open: rerunDialogOpen,
				onClose: () => setRerunDialogOpen(false),
				streamUrl: require_config.apiUrl("/api/rerun/stream", scope)
			})
		]
	});
}
exports.App = App;
exports.sourceUrl = sourceUrl;

//# sourceMappingURL=App.cjs.map