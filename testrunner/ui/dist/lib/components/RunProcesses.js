import { Field } from "./Field.js";
import { useMemo } from "react";
import { GoProfileView, ProcessTraceDetail, ProcessTraceTree, armedFacets } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { Button, Section, SplitPane } from "@flanksource/clicky-ui/components";
import { UiBeaker, UiLoader, UiRefresh } from "@flanksource/clicky-ui/icons";
// src/components/RunProcesses.tsx
/**
* The Processes view: the run's process tree beside the selected process's
* detail (command, lifetime, usage, stack dumps with their collect buttons,
* environment and the facet tabs the samples carry) and, when the run ingested
* Go profiles for it, its profiles. Collect and refresh need the run's live
* capture (`trace.canCollect`): a stored run shows what was recorded and offers
* neither.
*/
function RunProcesses({ trace, selectedKey, onSelect, hideExited, onHideExitedChange, expandAll, onExpandAllChange, stepLinks, onOpenStep, className }) {
	const selected = useMemo(() => trace.nodes.find((node) => node.key === selectedKey) ?? null, [trace.nodes, selectedKey]);
	const stepLink = selected ? stepLinks?.get(selected.pid) : void 0;
	const canRefresh = trace.canCollect && armedFacets(trace.samples).length > 0;
	const hasProfiles = !!selected && trace.goprof.some((record) => record.key === selected.key);
	return /* @__PURE__ */ jsx(SplitPane, {
		className: className ?? "h-full min-h-0",
		defaultSplit: 45,
		minLeft: 25,
		minRight: 30,
		leftClass: "border-r-2 border-border bg-muted/30",
		left: /* @__PURE__ */ jsx(ProcessTraceTree, {
			className: "h-full",
			nodes: trace.nodes,
			samples: trace.samples,
			selectedKey,
			onSelect,
			expandAll,
			onExpandAllChange,
			hideExited,
			onHideExitedChange,
			loading: trace.loading,
			...trace.error ? { error: trace.error } : {}
		}),
		right: /* @__PURE__ */ jsxs("div", {
			className: "flex h-full min-h-0 flex-col",
			children: [
				selected && canRefresh && /* @__PURE__ */ jsxs("div", {
					className: "flex shrink-0 items-center justify-end gap-2 border-b border-border px-density-4 py-density-1",
					children: [trace.refreshError && /* @__PURE__ */ jsx("span", {
						className: "min-w-0 truncate text-xs text-red-700 dark:text-red-400",
						role: "alert",
						title: trace.refreshError,
						children: trace.refreshError
					}), /* @__PURE__ */ jsxs(Button, {
						type: "button",
						size: "sm",
						variant: "outline",
						disabled: trace.refreshing,
						title: "Read the open files, sockets, ports and connections of this process in full on the next sample",
						onClick: () => void trace.refreshFacets(selected.key),
						children: [trace.refreshing ? /* @__PURE__ */ jsx(UiLoader, { className: "mr-1 animate-spin" }) : /* @__PURE__ */ jsx(UiRefresh, { className: "mr-1" }), "Refresh facets"]
					})]
				}),
				/* @__PURE__ */ jsx(ProcessTraceDetail, {
					className: "min-h-0 flex-1",
					node: selected,
					samples: trace.samples,
					stacks: trace.stacks,
					...trace.canCollect ? { onCollectStack: trace.collectStack } : {},
					collecting: trace.collecting,
					...trace.collectError ? { collectError: trace.collectError } : {},
					...stepLink && onOpenStep ? { links: /* @__PURE__ */ jsx(Field, {
						title: "Step",
						children: /* @__PURE__ */ jsxs(Button, {
							type: "button",
							variant: "link",
							size: "sm",
							className: "h-auto p-0",
							onClick: () => onOpenStep(stepLink),
							children: [/* @__PURE__ */ jsx(UiBeaker, { className: "mr-1" }), stepLink.name]
						})
					}) } : {}
				}),
				selected && hasProfiles && /* @__PURE__ */ jsx("div", {
					className: "max-h-[45%] shrink-0 overflow-y-auto border-t border-border px-density-4 py-density-2",
					"data-go-profiles": "",
					children: /* @__PURE__ */ jsxs(Section, {
						title: "Go profiles",
						defaultOpen: true,
						children: [/* @__PURE__ */ jsx(GoProfileView, {
							records: trace.goprof,
							processKey: selected.key
						}), trace.goprofExportUrl && /* @__PURE__ */ jsx("a", {
							className: "mt-2 inline-block text-xs text-primary hover:underline",
							href: trace.goprofExportUrl,
							download: true,
							children: "Download pprof"
						})]
					})
				})
			]
		})
	});
}
export { RunProcesses };

//# sourceMappingURL=RunProcesses.js.map