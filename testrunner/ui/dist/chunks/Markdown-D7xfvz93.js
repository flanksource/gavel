import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-lVm_Myqd.js";
import { T as r, t as i } from "./Callout-DmCAGTzW.js";
import { M as a, N as o, l as s } from "./clipboard-C0PFseuO.js";
// ../../../clicky-ui/packages/ui/src/icons/components/UiArrowUp.tsx
var c = /* @__PURE__ */ e(t(), 1), l = n(), u = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, l.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, l.jsx)("path", {
		fill: "currentColor",
		d: "M204.24 116.24a6 6 0 0 1-8.48 0L134 54.49V216a6 6 0 0 1-12 0V54.49l-61.76 61.75a6 6 0 0 1-8.48-8.48l72-72a6 6 0 0 1 8.48 0l72 72a6 6 0 0 1 0 8.48"
	})
}), {
	__source: "ph:arrow-up-light",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-up",
	displayName: "UiArrowUp"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, l.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, l.jsx)("path", {
		fill: "currentColor",
		d: "M207.39 115.06A8 8 0 0 1 200 120h-64v96a8 8 0 0 1-16 0v-96H56a8 8 0 0 1-5.66-13.66l72-72a8 8 0 0 1 11.32 0l72 72a8 8 0 0 1 1.73 8.72"
	})
}), {
	__source: "ph:arrow-up-fill",
	__viewBox: "0 0 256 256",
	__group: "navigation",
	__consumerName: "arrow-up",
	displayName: "UiArrowUpFilled"
});
// ../../../clicky-ui/packages/ui/src/data/markdown-source-blocks.ts
var d = "data-source-start", f = "data-source-end";
function p() {
	return (e) => {
		for (let t of e.children ?? []) t.type === "element" && t.position && (t.properties = {
			...t.properties,
			dataSourceStart: t.position.start.line,
			dataSourceEnd: t.position.end.line
		});
	};
}
function m(e) {
	let t = e[d], n = e[f];
	return {
		...typeof t == "string" || typeof t == "number" ? { [d]: t } : {},
		...typeof n == "string" || typeof n == "number" ? { [f]: n } : {}
	};
}
// ../../../clicky-ui/packages/ui/src/data/Markdown.tsx
var h = {
	h1: "mt-4 mb-2 text-lg font-semibold",
	h2: "mt-3 mb-1.5 text-base font-semibold",
	h3: "mt-3 mb-1 text-sm font-semibold",
	h4: "mt-2 mb-1 text-sm font-semibold",
	h5: "mt-2 mb-1 text-sm font-medium",
	h6: "mt-2 mb-1 text-sm font-medium"
};
function g({ as: e, className: t, children: n, node: i, ...a }) {
	return /* @__PURE__ */ (0, l.jsx)(e, {
		className: r(h[e], t),
		"data-streamdown": `heading-${e.slice(1)}`,
		...a,
		children: n
	});
}
var _ = null;
function v() {
	return _ ||= import("./dist-BgZVLhq1.js").then((e) => ({
		Streamdown: e.Streamdown,
		defaultRehypePlugins: e.defaultRehypePlugins
	})).catch((e) => (console.warn("clicky-ui: streamdown not available; rendering markdown as plain text", e), null)), _;
}
function y(e) {
	return typeof e == "string" ? e : typeof e == "number" ? String(e) : Array.isArray(e) ? e.map(y).join("") : (0, c.isValidElement)(e) ? y(e.props.children) : "";
}
function b({ className: e, children: t, ...n }) {
	if (!("data-block" in n)) return /* @__PURE__ */ (0, l.jsx)("code", {
		className: r("rounded bg-muted px-1.5 py-0.5 font-mono text-sm", e),
		children: t
	});
	let i = /language-([\w-]+)/.exec(e ?? "")?.[1], a = y(t).replace(/\n$/, ""), s = /* @__PURE__ */ (0, l.jsx)(o, {
		source: a,
		language: i,
		className: "not-prose my-4",
		copyable: !0,
		downloadable: !0,
		themeToggle: !0
	}), c = m(n);
	return Object.keys(c).length > 0 ? /* @__PURE__ */ (0, l.jsx)("div", {
		...c,
		children: s
	}) : s;
}
function x({ children: e, ...t }) {
	return (0, c.isValidElement)(e) ? (0, c.cloneElement)(e, {
		"data-block": "true",
		...m(t)
	}) : e;
}
function S({ className: e, children: t, ...n }) {
	let i = e?.includes("contains-task-list");
	return /* @__PURE__ */ (0, l.jsx)("ul", {
		className: r("whitespace-normal", i ? "list-none pl-0" : "list-disc pl-6", e),
		...m(n),
		children: t
	});
}
function C({ className: e, children: t, ...n }) {
	return /* @__PURE__ */ (0, l.jsx)("ol", {
		className: r("list-decimal whitespace-normal pl-6", e),
		...m(n),
		children: t
	});
}
function w({ className: e, children: t, ...n }) {
	return /* @__PURE__ */ (0, l.jsx)("div", {
		className: "not-prose my-4 max-w-full overflow-x-auto rounded-md border border-border",
		...m(n),
		children: /* @__PURE__ */ (0, l.jsx)("table", {
			className: r("w-max min-w-full table-auto text-left text-sm", e),
			children: t
		})
	});
}
function T({ className: e, children: t }) {
	return /* @__PURE__ */ (0, l.jsx)("thead", {
		className: r("border-b border-border bg-muted", e),
		children: t
	});
}
function E({ className: e, children: t }) {
	return /* @__PURE__ */ (0, l.jsx)("tbody", {
		className: r("divide-y divide-border", e),
		children: t
	});
}
function D({ className: e, style: t, children: n }) {
	return /* @__PURE__ */ (0, l.jsx)("th", {
		className: r("whitespace-nowrap px-density-3 py-density-2 font-semibold", e),
		style: t,
		children: n
	});
}
function O({ className: e, style: t, children: n }) {
	return /* @__PURE__ */ (0, l.jsx)("td", {
		className: r("px-density-3 py-density-2 align-top", e),
		style: t,
		children: n
	});
}
function k({ href: e, title: t, className: n, children: i }) {
	let { renderLink: o } = a(), [u, d] = (0, c.useState)(!1), f = r("wrap-anywhere font-medium text-primary underline", n);
	return e ? e.startsWith("#") ? /* @__PURE__ */ (0, l.jsx)("a", {
		href: e,
		title: t,
		className: f,
		children: i
	}) : e.startsWith("/") && !e.startsWith("//") ? o({
		to: e,
		className: f,
		children: i,
		...t ? { title: t } : {}
	}) : /* @__PURE__ */ (0, l.jsxs)(l.Fragment, { children: [/* @__PURE__ */ (0, l.jsx)("a", {
		href: e,
		title: t,
		target: "_blank",
		rel: "noopener noreferrer",
		className: f,
		onClick: (e) => {
			e.preventDefault(), d(!0);
		},
		children: i
	}), /* @__PURE__ */ (0, l.jsx)(s, {
		open: u,
		title: "Open external link?",
		message: /* @__PURE__ */ (0, l.jsx)("code", {
			className: "break-all font-mono text-sm",
			children: e
		}),
		confirmLabel: "Open link",
		onCancel: () => d(!1),
		onConfirm: () => {
			d(!1), window.open(e, "_blank", "noopener,noreferrer");
		}
	})] }) : /* @__PURE__ */ (0, l.jsx)("span", {
		className: n,
		children: i
	});
}
var A = {
	h1: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h1",
		...e
	}),
	h2: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h2",
		...e
	}),
	h3: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h3",
		...e
	}),
	h4: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h4",
		...e
	}),
	h5: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h5",
		...e
	}),
	h6: (e) => /* @__PURE__ */ (0, l.jsx)(g, {
		as: "h6",
		...e
	}),
	a: k,
	pre: x,
	code: b,
	ul: S,
	ol: C,
	table: w,
	thead: T,
	tbody: E,
	th: D,
	td: O
};
function j({ badge: e, children: t, className: n, emphasis: r, icon: a, label: o, source: s, title: c, variant: u }) {
	return /* @__PURE__ */ (0, l.jsx)(i, {
		badge: e,
		className: n,
		emphasis: r,
		icon: a,
		label: o,
		source: s,
		title: c,
		variant: u,
		children: t
	});
}
var M = {
	...A,
	calloutbox: j
}, N = { calloutbox: [
	"variant",
	"label",
	"badge",
	"source",
	"icon",
	"emphasis",
	"title"
] };
function P({ text: e, className: t, callouts: n = !1, sourceBlocks: i = !1 }) {
	if (i && n) throw Error("Markdown: sourceBlocks and callouts cannot be combined; callouts replaces the sanitiser schema that sourceBlocks keeps");
	let [a, o] = (0, c.useState)(null);
	(0, c.useEffect)(() => {
		let e = !1;
		return v().then((t) => {
			e || o(t);
		}), () => {
			e = !0;
		};
	}, []);
	let s = (0, c.useMemo)(() => a ? [...Object.values(a.defaultRehypePlugins), p] : void 0, [a]);
	if (!a) return /* @__PURE__ */ (0, l.jsx)("pre", {
		className: r("whitespace-pre-wrap text-sm", t),
		children: e
	});
	let { Streamdown: u } = a;
	return /* @__PURE__ */ (0, l.jsx)("div", {
		className: r("prose prose-sm max-w-none dark:prose-invert", t),
		children: i ? /* @__PURE__ */ (0, l.jsx)(u, {
			mode: "static",
			components: A,
			rehypePlugins: s,
			children: e
		}) : /* @__PURE__ */ (0, l.jsx)(u, {
			parseIncompleteMarkdown: !0,
			components: n ? M : A,
			allowedTags: n ? N : void 0,
			children: e
		})
	});
}
export { u as n, P as t };
