const require_Field = require("../components/Field.cjs");
let _flanksource_clicky_ui_data = require("@flanksource/clicky-ui/data");
let _flanksource_clicky_ui_clicky = require("@flanksource/clicky-ui/clicky");
let react_jsx_runtime = require("react/jsx-runtime");
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
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [ctx.parent_test && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
		title: "Parent Test",
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
			className: "font-mono text-sm",
			children: ctx.parent_test
		})
	}), ctx.import_path && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
		title: "Import Path",
		children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
			className: "font-mono text-sm",
			children: ctx.import_path
		})
	})] });
}
function GinkgoContextFields({ ctx, failureLocation, suiteAbove }) {
	const showLocation = ctx.failure_location && ctx.failure_location !== failureLocation;
	const showDescription = ctx.suite_description && ctx.suite_description !== suiteAbove;
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(react_jsx_runtime.Fragment, { children: [
		(showDescription || ctx.suite_path) && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(require_Field.Field, {
			title: "Ginkgo suite",
			children: [showDescription && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "text-sm font-medium",
				children: ctx.suite_description
			}), ctx.suite_path && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "font-mono text-xs text-muted-foreground",
				children: ctx.suite_path
			})]
		}),
		showLocation && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
			title: "Failure Location",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
				className: "font-mono text-sm text-red-600 dark:text-red-400",
				children: ctx.failure_location
			})
		}),
		!!ctx.labels?.length && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
			title: "Labels",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "flex flex-wrap gap-1",
				children: ctx.labels.map((label) => /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
					className: "rounded bg-muted px-2 py-0.5 text-xs",
					children: label
				}, label))
			})
		}),
		!!ctx.num_attempts && ctx.num_attempts > 1 && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
			title: "Ginkgo attempts",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
				className: "text-sm",
				children: ctx.num_attempts
			})
		}),
		!!ctx.report_entries?.length && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
			title: "Report entries",
			children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
				className: "space-y-2",
				children: ctx.report_entries.map((entry, index) => /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
					className: "rounded border border-border bg-muted/50 p-2",
					children: [
						/* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "text-sm font-medium",
							children: entry.name
						}),
						entry.file && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
							className: "font-mono text-xs text-muted-foreground",
							children: [entry.file, entry.line ? `:${entry.line}` : ""]
						}),
						/* @__PURE__ */ (0, react_jsx_runtime.jsx)("div", {
							className: "mt-1 break-words",
							children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.JsonView, { data: entry.value ?? entry.representation })
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
	const runner = (0, _flanksource_clicky_ui_data.useTestRunner)();
	const attempt = node.attempts?.at(-1);
	const task = taskMeta(node);
	const counts = (node.children?.length ?? 0) > 0 ? (0, _flanksource_clicky_ui_data.sum)(clickyNode) : null;
	const suite = node.suite?.join(" > ");
	const failure = node.failure_detail && node.failure_detail.kind && node.failure_detail.kind !== "raw" ? node.failure_detail : void 0;
	const showFullMessage = !!node.message && !!failure && node.message !== failure.summary && node.message.includes("\n");
	return /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
		className: "space-y-4 p-density-4",
		children: [
			node.cached && /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
				className: "inline-block rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground",
				children: "cached"
			}),
			suite && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Suite",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
					className: "text-sm",
					children: suite
				})
			}),
			node.package_path && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Package",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
					className: "font-mono text-sm",
					children: node.package_path
				})
			}),
			failure ? /* @__PURE__ */ (0, react_jsx_runtime.jsxs)(require_Field.Field, {
				title: "Error",
				children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.TestFailureDetail, { failure }), showFullMessage && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("details", {
					className: "rounded bg-muted/50 p-2",
					children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("summary", {
						className: "cursor-pointer text-xs font-semibold text-muted-foreground",
						children: "Full message"
					}), /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.AnsiHtml, {
						text: node.message,
						className: "mt-2 block max-h-64 overflow-y-auto whitespace-pre-wrap font-mono text-xs"
					})]
				})]
			}) : node.message ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Error",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.AnsiHtml, {
					text: node.message,
					className: "block max-h-64 overflow-y-auto whitespace-pre-wrap rounded bg-red-50 p-3 font-mono text-sm text-red-700 dark:bg-red-500/10 dark:text-red-300"
				})
			}) : null,
			node.command && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Command",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("pre", {
					className: "whitespace-pre-wrap rounded bg-muted p-2 font-mono text-sm",
					children: [/* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "text-muted-foreground",
						children: "$ "
					}), node.command]
				})
			}),
			task && (task.status || task.type) && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Task",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
					className: "flex gap-4 text-sm",
					children: [task.type && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: ["type: ", /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "font-mono",
						children: task.type
					})] }), task.status && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: ["status: ", /* @__PURE__ */ (0, react_jsx_runtime.jsx)("span", {
						className: "font-mono",
						children: task.status
					})] })]
				})
			}),
			node.framework === "go test" && node.context && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(GoTestContextFields, { ctx: node.context }),
			node.framework === "ginkgo" && node.context && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(GinkgoContextFields, {
				ctx: node.context,
				failureLocation: node.failure_detail?.location,
				suiteAbove: suite
			}),
			attempt && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Attempt",
				children: /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.TestAttemptDetail, {
					attempt,
					...runner.fetchProcessMetrics ? { fetchProcessMetrics: runner.fetchProcessMetrics } : {}
				}, `${attempt.run_kind ?? ""}:${attempt.sequence}`)
			}),
			/* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.TestOutput, { node: clickyNode }),
			counts && counts.total > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("div", {
				className: "flex flex-wrap gap-3 text-xs text-muted-foreground",
				children: [
					/* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.total, " total"] }),
					counts.passed > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.passed, " passed"] }),
					counts.failed > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.failed, " failed"] }),
					counts.warned > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.warned, " warned"] }),
					counts.aborted > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.aborted, " aborted"] }),
					counts.skipped > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.skipped, " skipped"] }),
					counts.pending > 0 && /* @__PURE__ */ (0, react_jsx_runtime.jsxs)("span", { children: [counts.pending, " pending"] })
				]
			}),
			node.detail != null && /* @__PURE__ */ (0, react_jsx_runtime.jsx)(require_Field.Field, {
				title: "Detail",
				children: isLegacyClickyDocument(node.detail) ? /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_clicky.Clicky, { data: node.detail }) : /* @__PURE__ */ (0, react_jsx_runtime.jsx)(_flanksource_clicky_ui_data.JsonView, { data: node.detail })
			})
		]
	});
}
exports.GavelDetailBody = GavelDetailBody;

//# sourceMappingURL=GavelDetailBody.cjs.map