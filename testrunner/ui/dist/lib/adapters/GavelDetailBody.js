import { Field } from "../components/Field.js";
import { AnsiHtml, JsonView, TestAttemptDetail, TestFailureDetail, TestOutput, sum, useTestRunner } from "@flanksource/clicky-ui/data";
import { Clicky } from "@flanksource/clicky-ui/clicky";
import { Fragment, jsx, jsxs } from "react/jsx-runtime";
// src/adapters/GavelDetailBody.tsx
function isLegacyClickyDocument(value) {
	return typeof value === "object" && value !== null && "version" in value && "node" in value;
}
function taskMeta(t) {
	if (t.framework !== "task" || !t.context || typeof t.context !== "object") return null;
	const ctx = t.context;
	return {
		status: typeof ctx.status === "string" ? ctx.status : void 0,
		type: typeof ctx.type === "string" ? ctx.type : void 0
	};
}
function GoTestContextFields({ ctx }) {
	return /* @__PURE__ */ jsxs(Fragment, { children: [ctx.parent_test && /* @__PURE__ */ jsx(Field, {
		title: "Parent Test",
		children: /* @__PURE__ */ jsx("span", {
			className: "font-mono text-sm",
			children: ctx.parent_test
		})
	}), ctx.import_path && /* @__PURE__ */ jsx(Field, {
		title: "Import Path",
		children: /* @__PURE__ */ jsx("span", {
			className: "font-mono text-sm",
			children: ctx.import_path
		})
	})] });
}
function GinkgoContextFields({ ctx, failureLocation, suiteAbove }) {
	const showLocation = ctx.failure_location && ctx.failure_location !== failureLocation;
	const showDescription = ctx.suite_description && ctx.suite_description !== suiteAbove;
	return /* @__PURE__ */ jsxs(Fragment, { children: [
		(showDescription || ctx.suite_path) && /* @__PURE__ */ jsxs(Field, {
			title: "Ginkgo suite",
			children: [showDescription && /* @__PURE__ */ jsx("div", {
				className: "text-sm font-medium",
				children: ctx.suite_description
			}), ctx.suite_path && /* @__PURE__ */ jsx("div", {
				className: "font-mono text-xs text-muted-foreground",
				children: ctx.suite_path
			})]
		}),
		showLocation && /* @__PURE__ */ jsx(Field, {
			title: "Failure Location",
			children: /* @__PURE__ */ jsx("span", {
				className: "font-mono text-sm text-red-600 dark:text-red-400",
				children: ctx.failure_location
			})
		}),
		!!ctx.labels?.length && /* @__PURE__ */ jsx(Field, {
			title: "Labels",
			children: /* @__PURE__ */ jsx("div", {
				className: "flex flex-wrap gap-1",
				children: ctx.labels.map((label) => /* @__PURE__ */ jsx("span", {
					className: "rounded bg-muted px-2 py-0.5 text-xs",
					children: label
				}, label))
			})
		}),
		!!ctx.num_attempts && ctx.num_attempts > 1 && /* @__PURE__ */ jsx(Field, {
			title: "Ginkgo attempts",
			children: /* @__PURE__ */ jsx("span", {
				className: "text-sm",
				children: ctx.num_attempts
			})
		}),
		!!ctx.report_entries?.length && /* @__PURE__ */ jsx(Field, {
			title: "Report entries",
			children: /* @__PURE__ */ jsx("div", {
				className: "space-y-2",
				children: ctx.report_entries.map((entry, index) => /* @__PURE__ */ jsxs("div", {
					className: "rounded border border-border bg-muted/50 p-2",
					children: [
						/* @__PURE__ */ jsx("div", {
							className: "text-sm font-medium",
							children: entry.name
						}),
						entry.file && /* @__PURE__ */ jsxs("div", {
							className: "font-mono text-xs text-muted-foreground",
							children: [entry.file, entry.line ? `:${entry.line}` : ""]
						}),
						/* @__PURE__ */ jsx("div", {
							className: "mt-1 break-words",
							children: /* @__PURE__ */ jsx(JsonView, { data: entry.value ?? entry.representation })
						})
					]
				}, `${entry.name}-${index}`))
			})
		})
	] });
}
/**
* The detail body of every node that is not a fixture step (fixture steps use
* clicky's verification adapter): suite, package and command, the error, the
* framework's own context (Go subtest parent, Ginkgo suite/labels/report
* entries), the latest attempt's process metrics and stack, output and the
* provider's structured detail.
*/
function GavelDetailBody({ node: clickyNode }) {
	const node = clickyNode;
	const runner = useTestRunner();
	const attempt = node.attempts?.at(-1);
	const task = taskMeta(node);
	const counts = (node.children?.length ?? 0) > 0 ? sum(clickyNode) : null;
	const suite = node.suite?.join(" > ");
	const failure = node.failure_detail && node.failure_detail.kind && node.failure_detail.kind !== "raw" ? node.failure_detail : void 0;
	const showFullMessage = !!node.message && !!failure && node.message !== failure.summary && node.message.includes("\n");
	return /* @__PURE__ */ jsxs("div", {
		className: "space-y-4 p-density-4",
		children: [
			node.cached && /* @__PURE__ */ jsx("span", {
				className: "inline-block rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: "cached"
			}),
			suite && /* @__PURE__ */ jsx(Field, {
				title: "Suite",
				children: /* @__PURE__ */ jsx("span", {
					className: "text-sm",
					children: suite
				})
			}),
			node.package_path && /* @__PURE__ */ jsx(Field, {
				title: "Package",
				children: /* @__PURE__ */ jsx("span", {
					className: "font-mono text-sm",
					children: node.package_path
				})
			}),
			failure ? /* @__PURE__ */ jsxs(Field, {
				title: "Error",
				children: [/* @__PURE__ */ jsx(TestFailureDetail, { failure }), showFullMessage && /* @__PURE__ */ jsxs("details", {
					className: "rounded bg-muted/50 p-2",
					children: [/* @__PURE__ */ jsx("summary", {
						className: "cursor-pointer text-xs font-semibold text-muted-foreground",
						children: "Full message"
					}), /* @__PURE__ */ jsx(AnsiHtml, {
						text: node.message,
						className: "mt-2 block max-h-64 overflow-y-auto whitespace-pre-wrap font-mono text-xs"
					})]
				})]
			}) : node.message ? /* @__PURE__ */ jsx(Field, {
				title: "Error",
				children: /* @__PURE__ */ jsx(AnsiHtml, {
					text: node.message,
					className: "block max-h-64 overflow-y-auto whitespace-pre-wrap rounded bg-red-50 p-3 font-mono text-sm text-red-700 dark:bg-red-500/10 dark:text-red-300"
				})
			}) : null,
			node.command && /* @__PURE__ */ jsx(Field, {
				title: "Command",
				children: /* @__PURE__ */ jsxs("pre", {
					className: "whitespace-pre-wrap rounded bg-muted p-2 font-mono text-sm",
					children: [/* @__PURE__ */ jsx("span", {
						className: "text-muted-foreground",
						children: "$ "
					}), node.command]
				})
			}),
			task && (task.status || task.type) && /* @__PURE__ */ jsx(Field, {
				title: "Task",
				children: /* @__PURE__ */ jsxs("div", {
					className: "flex gap-4 text-sm",
					children: [task.type && /* @__PURE__ */ jsxs("span", { children: ["type: ", /* @__PURE__ */ jsx("span", {
						className: "font-mono",
						children: task.type
					})] }), task.status && /* @__PURE__ */ jsxs("span", { children: ["status: ", /* @__PURE__ */ jsx("span", {
						className: "font-mono",
						children: task.status
					})] })]
				})
			}),
			node.framework === "go test" && node.context && /* @__PURE__ */ jsx(GoTestContextFields, { ctx: node.context }),
			node.framework === "ginkgo" && node.context && /* @__PURE__ */ jsx(GinkgoContextFields, {
				ctx: node.context,
				failureLocation: node.failure_detail?.location,
				suiteAbove: suite
			}),
			attempt && /* @__PURE__ */ jsx(Field, {
				title: "Attempt",
				children: /* @__PURE__ */ jsx(TestAttemptDetail, {
					attempt,
					...runner.fetchProcessMetrics ? { fetchProcessMetrics: runner.fetchProcessMetrics } : {}
				}, `${attempt.run_kind ?? ""}:${attempt.sequence}`)
			}),
			/* @__PURE__ */ jsx(TestOutput, { node: clickyNode }),
			counts && counts.total > 0 && /* @__PURE__ */ jsxs("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ jsxs("span", { children: [counts.total, " total"] }),
					counts.passed > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.passed, " passed"] }),
					counts.failed > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.failed, " failed"] }),
					counts.warned > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.warned, " warned"] }),
					counts.aborted > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.aborted, " aborted"] }),
					counts.skipped > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.skipped, " skipped"] }),
					counts.pending > 0 && /* @__PURE__ */ jsxs("span", { children: [counts.pending, " pending"] })
				]
			}),
			node.detail != null && /* @__PURE__ */ jsx(Field, {
				title: "Detail",
				children: isLegacyClickyDocument(node.detail) ? /* @__PURE__ */ jsx(Clicky, { data: node.detail }) : /* @__PURE__ */ jsx(JsonView, { data: node.detail })
			})
		]
	});
}
export { GavelDetailBody };

//# sourceMappingURL=GavelDetailBody.js.map