import { isLintOnlyPhase, processLinksByPid, stripTestTaskGroup } from "./utils.js";
import { useTestRun } from "./hooks/use-test-run.js";
import { apiBase, apiUrl } from "./config.js";
import { processMount, rebaseTraces } from "./traces.js";
import { useProcessTrace } from "./hooks/use-process-trace.js";
import { fetchProcessMetricsFromSamples } from "./processMetrics.js";
import { buildRoute, parseRoute } from "./routes.js";
import { gavelRegistry } from "./adapters/index.js";
import { RunProcesses } from "./components/RunProcesses.js";
import { BenchView } from "./components/BenchView.js";
import { FixtureBenchView } from "./components/FixtureBenchView.js";
import { RerunDialog } from "./components/RerunDialog.js";
import { RunActions } from "./components/RunActions.js";
import { buildLintRerunBody } from "./lintRerun.js";
import { LintTasks } from "./components/LintTasks.js";
import { RunHistory } from "./components/RunHistory.js";
import { RunSwitcher } from "./components/RunSwitcher.js";
import { useRunHistory } from "./hooks/use-run-history.js";
import { withExportPaths } from "./exportPaths.js";
import { Suspense, lazy, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Icon, LintResults, TestRunner, aggregateStatusCounts, annotateRoutePaths, buildLintTree, collapseSingleChildChains, filterTests, findLintNode, findNodeByRoutePath, isProcessExited, totalLintViolations, violationIgnoreActions } from "@flanksource/clicky-ui/data";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";
import { Button, DropdownMenu, Tabs } from "@flanksource/clicky-ui/components";
import { UiArrowLeft, UiBeaker, UiChartLineUp, UiEyeClosed, UiFileCode, UiHistory, UiListChecks, UiLoader, UiRefresh, UiServerProcess } from "@flanksource/clicky-ui/icons";
import { lintFindings, testFindings } from "@flanksource/clicky-ui/file-findings";
// src/App.tsx
function countFixtureWork(tests) {
	return tests.reduce((count, test) => count + (test.fixture ? 1 : 0) + countFixtureWork(test.children || []), 0);
}
var asGavel = (tests) => tests;
var FilesTab = lazy(() => import("./FilesTab.js"));
var PromptRunView = lazy(() => import("./PromptRunView.js"));
function renderPromptRun(result) {
	if (!result.run_id) return null;
	return /* @__PURE__ */ jsx(Suspense, {
		fallback: /* @__PURE__ */ jsx("p", {
			role: "status",
			className: "text-sm text-muted-foreground",
			children: "Loading the prompt run…"
		}),
		children: /* @__PURE__ */ jsx(PromptRunView, {
			result,
			baseUrl: apiBase()
		})
	});
}
function FilesLoading() {
	return /* @__PURE__ */ jsxs("div", {
		className: "flex h-full items-center justify-center text-sm text-muted-foreground",
		role: "status",
		children: [/* @__PURE__ */ jsx(Icon, {
			icon: UiLoader,
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
	const res = await fetch(apiUrl(`/api/file?${params}`, scope));
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
	return /* @__PURE__ */ jsxs("div", {
		className: "flex min-w-0 items-center gap-3 border-b border-border bg-card px-density-4 py-density-2",
		"data-page-header": "",
		children: [
			/* @__PURE__ */ jsxs("h1", {
				className: "inline-flex shrink-0 items-center gap-1.5 text-base font-bold",
				children: [/* @__PURE__ */ jsx(Icon, {
					icon,
					className: "text-primary"
				}), title]
			}),
			status && /* @__PURE__ */ jsx("span", {
				className: "truncate text-sm text-muted-foreground",
				children: status
			}),
			actions && /* @__PURE__ */ jsx("div", {
				className: "ml-auto flex shrink-0 items-center gap-1",
				children: actions
			})
		]
	});
}
function Waiting({ done }) {
	return /* @__PURE__ */ jsx("div", {
		className: "flex h-full items-center justify-center p-8 text-center text-sm text-muted-foreground",
		children: done ? "No tests were run" : /* @__PURE__ */ jsxs("div", { children: [/* @__PURE__ */ jsx(Icon, {
			icon: UiLoader,
			className: "animate-spin text-3xl text-blue-500"
		}), /* @__PURE__ */ jsx("p", {
			className: "mt-2",
			children: "Waiting for test results..."
		})] })
	});
}
var NOTICE_MS = 8e3;
var GLOBAL_STOP_KEY = "__global__";
var loadPage = (href) => window.location.assign(href);
function App({ runId, navigate = loadPage } = {}) {
	const scope = useMemo(() => runId ? { runId } : {}, [runId]);
	const base = apiBase(scope);
	const run = useTestRun(scope.runId ? { runId: scope.runId } : {});
	const { snapshot, tests, lint, status: snapshotStatus, runMeta, done, statusText, error: streamError } = run;
	const performance = snapshot?.performance;
	const runError = snapshot?.error;
	const lintRun = !!snapshotStatus.lint_run;
	const history = useRunHistory();
	const [route, setRoute] = useState(() => parseRoute(window.location));
	const [expandAll, setExpandAll] = useState(null);
	const [rerunBusy, setRerunBusy] = useState(false);
	const [rerunDialogOpen, setRerunDialogOpen] = useState(false);
	const [ignoreBusy, setIgnoreBusy] = useState(false);
	const [testEditBusy, setTestEditBusy] = useState(false);
	const [stopBusyKey, setStopBusyKey] = useState(null);
	const [notice, setNotice] = useState(void 0);
	const noticeTimer = useRef(null);
	const firstSeen = useRef(Date.now());
	const [now, setNow] = useState(Date.now());
	const say = useCallback((message) => {
		setNotice(message);
		if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current);
		noticeTimer.current = window.setTimeout(() => setNotice(void 0), NOTICE_MS);
	}, []);
	useEffect(() => () => {
		if (noticeTimer.current !== null) window.clearTimeout(noticeTimer.current);
	}, []);
	useEffect(() => {
		if (done) return;
		const timer = window.setInterval(() => setNow(Date.now()), 1e3);
		return () => window.clearInterval(timer);
	}, [done]);
	const startTime = runMeta?.started ? Date.parse(runMeta.started) : firstSeen.current;
	const endTime = runMeta?.ended ? Date.parse(runMeta.ended) : null;
	const commitRoute = useCallback((next, mode = "push") => {
		setRoute(next);
		const url = buildRoute(next);
		if (url === `${window.location.pathname}${window.location.search}`) return;
		if (mode === "replace") window.history.replaceState({}, "", url);
		else window.history.pushState({}, "", url);
	}, []);
	useEffect(() => {
		const onPopState = () => setRoute(parseRoute(window.location));
		window.addEventListener("popstate", onPopState);
		return () => window.removeEventListener("popstate", onPopState);
	}, []);
	useEffect(() => {
		if (performance && route.tab === "tests" && window.location.pathname === `${apiBase()}/`) commitRoute({
			...route,
			tab: "bench",
			selectedPath: ""
		}, "replace");
	}, [
		performance,
		route,
		commitRoute
	]);
	const displayedTests = useMemo(() => isLintOnlyPhase(tests, !!snapshotStatus.running) ? stripTestTaskGroup(tests) : tests, [tests, snapshotStatus.running]);
	const totals = useMemo(() => aggregateStatusCounts(displayedTests), [displayedTests]);
	const filteredTests = useMemo(() => annotateRoutePaths(collapseSingleChildChains(withExportPaths(filterTests(withOwnNames(displayedTests), route.filters.status, route.filters.framework)))), [displayedTests, route.filters]);
	const selectedTest = useMemo(() => route.tab === "tests" && route.selectedPath ? findNodeByRoutePath(filteredTests, route.selectedPath) : null, [
		route.tab,
		route.selectedPath,
		filteredTests
	]);
	const lintTree = useMemo(() => buildLintTree({
		results: lint ?? [],
		grouping: route.lintGrouping,
		filters: route.lintFilters
	}), [
		lint,
		route.lintGrouping,
		route.lintFilters
	]);
	const selectedLint = route.tab === "lint" && route.selectedPath ? findLintNode(lintTree, route.selectedPath) : null;
	const lintTotal = useMemo(() => totalLintViolations(lint ?? []), [lint]);
	const gitRoot = snapshot?.git?.root;
	const { findings, linterByFinding, sourceUrlByFinding } = useMemo(() => {
		const linterByFinding = /* @__PURE__ */ new Map();
		const sourceUrlByFinding = /* @__PURE__ */ new Map();
		return {
			findings: [...(lint ?? []).flatMap((result) => {
				const workDir = result.work_dir || gitRoot;
				return lintFindings([{
					...result,
					...workDir ? { work_dir: workDir } : {}
				}]).map((finding) => {
					linterByFinding.set(findingKey(finding), result.linter);
					const url = sourceUrl(result.sources?.[finding.file]?.url, finding.line);
					if (url) sourceUrlByFinding.set(findingKey(finding), url);
					return finding;
				});
			}), ...testFindings(withWorkDir(tests, gitRoot))],
			linterByFinding,
			sourceUrlByFinding
		};
	}, [
		lint,
		tests,
		gitRoot
	]);
	const selectedFile = useMemo(() => route.tab === "files" && route.selectedPath ? {
		file: route.selectedPath,
		...route.workDir ? { workDir: route.workDir } : {}
	} : null, [
		route.tab,
		route.selectedPath,
		route.workDir
	]);
	const traces = useMemo(() => rebaseTraces(snapshotStatus.traces, base), [snapshotStatus.traces, base]);
	const processTraceAvailable = !!processMount(snapshotStatus.traces);
	const trace = useProcessTrace({
		traces: snapshotStatus.traces,
		done,
		base,
		...route.tab === "processes" && route.selectedPath ? { selectedKey: route.selectedPath } : {}
	});
	const processCount = useMemo(() => trace.nodes.filter((node) => !isProcessExited(node)).length, [trace.nodes]);
	const rootProcess = useMemo(() => trace.nodes.find((node) => node.isRoot) ?? null, [trace.nodes]);
	const selectedProcess = route.tab === "processes" ? trace.nodes.find((node) => node.key === route.selectedPath) ?? null : null;
	const processLinks = useMemo(() => processTraceAvailable ? processLinksByPid(asGavel(annotateRoutePaths(collapseSingleChildChains(displayedTests)))) : /* @__PURE__ */ new Map(), [processTraceAvailable, displayedTests]);
	useEffect(() => {
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
	useEffect(() => {
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
	const onTabChange = useCallback((tab) => {
		const { workDir: _workDir, line: _line, ...rest } = route;
		commitRoute({
			...rest,
			tab,
			selectedPath: ""
		});
	}, [route, commitRoute]);
	const onSelectTest = useCallback((node) => commitRoute({
		...route,
		selectedPath: node?.route_path ?? ""
	}), [route, commitRoute]);
	const onFiltersChange = useCallback((filters) => commitRoute({
		...route,
		tab: "tests",
		filters,
		selectedPath: ""
	}), [route, commitRoute]);
	const onLintFiltersChange = useCallback((lintFilters) => commitRoute({
		...route,
		tab: "lint",
		lintFilters,
		selectedPath: ""
	}), [route, commitRoute]);
	const onLintGroupingChange = useCallback((lintGrouping) => commitRoute({
		...route,
		tab: "lint",
		lintGrouping,
		selectedPath: ""
	}), [route, commitRoute]);
	const onSelectLint = useCallback((selectedPath) => commitRoute({
		...route,
		selectedPath: selectedPath ?? ""
	}), [route, commitRoute]);
	const onSelectProcess = useCallback((key) => commitRoute({
		...route,
		tab: "processes",
		selectedPath: key
	}), [route, commitRoute]);
	const onHideExitedChange = useCallback((hideExited) => commitRoute({
		...route,
		tab: "processes",
		hideExited
	}, "replace"), [route, commitRoute]);
	const openFile = useCallback((ref, line) => {
		const { workDir: _workDir, line: _line, ...rest } = route;
		commitRoute({
			...rest,
			tab: "files",
			selectedPath: ref.file,
			...ref.workDir ? { workDir: ref.workDir } : {},
			...line ? { line } : {}
		});
	}, [route, commitRoute]);
	const onOpenFile = useCallback((target) => openFile({
		file: target.file,
		...target.workDir || gitRoot ? { workDir: target.workDir || gitRoot } : {}
	}, target.line), [openFile, gitRoot]);
	const onSelectFile = useCallback((ref) => openFile(ref), [openFile]);
	const onLineChange = useCallback((line) => {
		const { line: _line, ...rest } = route;
		commitRoute({
			...rest,
			...line ? { line } : {}
		}, "replace");
	}, [route, commitRoute]);
	const onViewChange = useCallback((view) => commitRoute({
		...route,
		view
	}, "replace"), [route, commitRoute]);
	const loadFile = useCallback((ref) => fetchFile(ref, scope), [scope]);
	const onSampleChange = useCallback((sampleIndex) => commitRoute({
		...route,
		sampleIndex
	}), [route, commitRoute]);
	const onOpenStep = useCallback((link) => {
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
	const startRerun = useCallback(async (body, label) => {
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
	const onRerun = useCallback((node) => {
		if (node.framework === "task") return;
		const request = rerunBody(node);
		if ("error" in request) say(request.error);
		else startRerun(request.body, node.name);
	}, [startRerun, say]);
	const onRerunLint = useCallback(() => {
		const body = buildLintRerunBody(selectedLint, lint ?? []);
		if (!body.work_dir) say("Rerun across multiple lint roots is not supported");
		else startRerun(body, selectedLint?.name ?? "lint");
	}, [
		selectedLint,
		lint,
		startRerun,
		say
	]);
	const stop = useCallback(async ({ taskId, label }) => {
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
	const onStop = useCallback((node) => {
		if (node.task_id) stop({
			taskId: node.task_id,
			label: node.name
		});
	}, [stop]);
	const onStopAll = useCallback(() => {
		stop({ label: "the run" });
	}, [stop]);
	const onIgnore = useCallback(async (action) => {
		if (ignoreBusy) return;
		setIgnoreBusy(true);
		try {
			const res = await fetch(apiUrl("/api/lint/ignore", scope), {
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
	const renderFindingActions = useCallback((finding) => {
		if (finding.kind !== "lint") return null;
		const linter = linterByFinding.get(findingKey(finding));
		if (!linter) throw new Error(`no linter recorded for the finding at ${finding.file}:${finding.line}`);
		const violation = {
			...finding.rule ? { rule: { method: finding.rule } } : {},
			...finding.line ? { line: finding.line } : {}
		};
		const ignores = violationIgnoreActions({
			violation,
			linter,
			file: finding.file,
			workDir: finding.workDir ?? ""
		});
		const source = sourceUrlByFinding.get(findingKey(finding));
		if (ignores.length === 0 && !source) return null;
		return /* @__PURE__ */ jsxs(Fragment, { children: [source && /* @__PURE__ */ jsx(Button, {
			asChild: true,
			size: "sm",
			variant: "outline",
			className: "h-6 gap-1 px-2 text-xs text-muted-foreground",
			children: /* @__PURE__ */ jsx("a", {
				href: source,
				target: "_blank",
				rel: "noreferrer",
				title: `Open ${finding.file}:${finding.line ?? ""} at its source`,
				children: "Open source"
			})
		}), ignores.length > 0 && /* @__PURE__ */ jsx(DropdownMenu, {
			label: "Ignore",
			icon: UiEyeClosed,
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
	const onTestEdit = useCallback(async (node, action, editScope) => {
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
	const wholeResultRoute = useMemo(() => ({
		...route,
		selectedPath: ""
	}), [route]);
	const actions = useMemo(() => ({
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
	const registry = useMemo(() => gavelRegistry({
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
	const fetchProcessMetrics = useMemo(() => fetchProcessMetricsFromSamples({
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
	const runActions = /* @__PURE__ */ jsx(RunActions, {
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
			icon: UiBeaker,
			count: totals.failed > 0 ? totals.failed : totals.total,
			countColor: totals.failed > 0 ? "bg-red-500" : "bg-muted-foreground"
		},
		...showLintTab ? [{
			id: "lint",
			label: "Lint",
			icon: UiListChecks,
			count: lintTotal,
			countColor: lintTotal > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...showFilesTab ? [{
			id: "files",
			label: "Files",
			icon: UiFileCode,
			count: findings.length,
			countColor: findings.length > 0 ? "bg-yellow-500" : "bg-muted-foreground"
		}] : [],
		...showBenchTab ? [{
			id: "bench",
			label: "Bench",
			icon: UiChartLineUp,
			count: performance ? countFixtureWork(tests) : benchRegressions > 0 ? benchRegressions : run.bench?.deltas?.length || 0,
			countColor: benchRegressions > 0 ? "bg-red-500" : "bg-muted-foreground"
		}] : [],
		...showProcessesTab ? [{
			id: "processes",
			label: "Processes",
			icon: UiServerProcess,
			count: processCount,
			countColor: "bg-blue-500"
		}] : [],
		...history.available ? [{
			id: "history",
			label: "History",
			icon: UiHistory
		}] : []
	];
	return /* @__PURE__ */ jsxs("div", {
		className: "flex h-screen flex-col bg-background text-foreground",
		children: [
			backTo && /* @__PURE__ */ jsx("div", {
				className: "flex shrink-0 items-center gap-2 border-b border-border bg-card px-4 py-1.5 text-sm",
				children: /* @__PURE__ */ jsxs("a", {
					href: backTo,
					className: "inline-flex items-center gap-1 text-primary no-underline hover:underline",
					children: [/* @__PURE__ */ jsx(UiArrowLeft, {}), "Back to PR"]
				})
			}),
			(tabs.length > 1 || history.available) && /* @__PURE__ */ jsxs("div", {
				className: "flex shrink-0 items-center gap-2 pr-density-4",
				children: [/* @__PURE__ */ jsx(Tabs, {
					tabs,
					value: route.tab,
					onChange: onTabChange,
					className: "min-w-0 flex-1 px-density-4"
				}), history.available && /* @__PURE__ */ jsx(RunSwitcher, {
					history,
					navigate
				})]
			}),
			/* @__PURE__ */ jsxs("div", {
				className: "min-h-0 flex-1",
				children: [
					route.tab === "tests" && (displayedTests.length === 0 ? /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(PageHeader, {
							icon: UiBeaker,
							title: "Test Results",
							status,
							actions: runActions
						}), /* @__PURE__ */ jsx("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ jsx(Waiting, { done })
						})]
					}) : /* @__PURE__ */ jsx(TestRunner, {
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
					route.tab === "lint" && /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(LintTasks, { tests }), /* @__PURE__ */ jsx("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ jsx(LintResults, {
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
								headerSlot: /* @__PURE__ */ jsxs(Fragment, { children: [
									/* @__PURE__ */ jsx("span", {
										className: "truncate text-sm text-muted-foreground",
										children: status
									}),
									/* @__PURE__ */ jsxs(Button, {
										type: "button",
										size: "sm",
										disabled: rerunBusy,
										onClick: onRerunLint,
										title: selectedLint ? "Rerun the linters for this selection" : "Rerun the linters",
										children: [/* @__PURE__ */ jsx(UiRefresh, { className: "mr-1" }), rerunBusy ? "Running..." : "Rerun"]
									}),
									runActions
								] })
							})
						})]
					}),
					route.tab === "files" && /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(PageHeader, {
							icon: UiFileCode,
							title: "Files",
							status,
							actions: runActions
						}), /* @__PURE__ */ jsx("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ jsx(Suspense, {
								fallback: /* @__PURE__ */ jsx(FilesLoading, {}),
								children: /* @__PURE__ */ jsx(FilesTab, {
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
					route.tab === "bench" && /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(PageHeader, {
							icon: UiChartLineUp,
							title: performance ? "Fixture Benchmark" : "Benchmark Comparison",
							status,
							actions: runActions
						}), /* @__PURE__ */ jsx("div", {
							className: "min-h-0 flex-1 overflow-auto",
							children: performance ? /* @__PURE__ */ jsx(FixtureBenchView, {
								performance,
								tests,
								error: runError,
								sampleIndex: route.sampleIndex,
								onSampleChange
							}) : /* @__PURE__ */ jsx(BenchView, { bench: run.bench })
						})]
					}),
					route.tab === "history" && /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(PageHeader, {
							icon: UiHistory,
							title: "Run History",
							status: history.current ? `Viewing ${history.current}` : "Viewing the latest results"
						}), /* @__PURE__ */ jsx("div", {
							className: "flex min-h-0 flex-1 flex-col p-density-2",
							children: /* @__PURE__ */ jsx(RunHistory, {
								history,
								navigate
							})
						})]
					}),
					route.tab === "processes" && /* @__PURE__ */ jsxs("div", {
						className: "flex h-full flex-col",
						children: [/* @__PURE__ */ jsx(PageHeader, {
							icon: UiServerProcess,
							title: "Processes",
							status,
							actions: runActions
						}), /* @__PURE__ */ jsx("div", {
							className: "min-h-0 flex-1",
							children: /* @__PURE__ */ jsx(RunProcesses, {
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
			/* @__PURE__ */ jsx(RerunDialog, {
				open: rerunDialogOpen,
				onClose: () => setRerunDialogOpen(false),
				streamUrl: apiUrl("/api/rerun/stream", scope)
			})
		]
	});
}
export { App, sourceUrl };

//# sourceMappingURL=App.js.map