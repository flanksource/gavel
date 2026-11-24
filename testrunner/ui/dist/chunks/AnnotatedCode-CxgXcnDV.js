import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { t as r } from "./react-dom-BBpWcFtU.js";
import { a as i } from "./button-BLoR2pgR.js";
import { i as a } from "./UiInfo-CCiCR0Fr.js";
import { V as o } from "./SplitButton-w2X8sfgS.js";
import { l as s, r as c } from "./ErrorDetails-CbwPqfQb.js";
import { t as l } from "./ProblemActions-MzUA9awO.js";
import { t as u } from "./MonacoEditor-0E72ufhs.js";
import { r as d, t as f } from "./css-BSHA9lyv.js";
import { r as p, t as m } from "./html-DIGm8MDP.js";
import { n as h, t as g } from "./typescript-9qocCrJq.js";
import { r as _, t as v } from "./javascript-BqUICE53.js";
import { n as y, t as b } from "./less-Dd2Povnz.js";
import { n as x, t as S } from "./scss-D0OUqUOV.js";
import { C, D as w, E as T, M as ee, N as E, O as D, P as O, S as k, T as te, _ as A, a as j, b as ne, c as M, d as re, f as ie, g as N, h as P, i as ae, j as oe, k as F, l as I, m as se, n as L, o as R, p as ce, r as z, s as B, u as le, v as V, w as H, x as U, y as ue } from "./FilesTab-B1iR4D-6.js";
// ../../../clicky-ui/packages/ui/src/file-findings/monaco-language.ts
var W = /* @__PURE__ */ e(t(), 1), G = /* @__PURE__ */ e(r(), 1), K = n(), q = {
	go: "go",
	ts: "typescript",
	tsx: "typescript",
	mts: "typescript",
	cts: "typescript",
	js: "javascript",
	jsx: "javascript",
	mjs: "javascript",
	cjs: "javascript",
	json: "json",
	jsonc: "json",
	yaml: "yaml",
	yml: "yaml",
	md: "markdown",
	mdx: "markdown",
	py: "python",
	rb: "ruby",
	rs: "rust",
	java: "java",
	kt: "kotlin",
	c: "c",
	h: "c",
	cc: "cpp",
	cpp: "cpp",
	hpp: "cpp",
	cs: "csharp",
	php: "php",
	sh: "shell",
	bash: "shell",
	zsh: "shell",
	sql: "sql",
	css: "css",
	scss: "scss",
	less: "less",
	html: "html",
	xml: "xml",
	toml: "ini",
	ini: "ini",
	dockerfile: "dockerfile",
	proto: "protobuf",
	tf: "hcl",
	hcl: "hcl"
};
function J(e) {
	let t = o(e);
	return t && q[t] || "plaintext";
}
var de = {
	typescript: "findings-typescript",
	javascript: "findings-javascript",
	json: "findings-json",
	css: "findings-css",
	scss: "findings-scss",
	less: "findings-less",
	html: "findings-html"
};
function Y(e) {
	let t = J(e);
	return de[t] ?? t;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentPersistenceBoundary.tsx
function X({ scope: e, children: t }) {
	let n = (0, W.useMemo)(() => e ? E(e) : null, [e]);
	return /* @__PURE__ */ (0, K.jsxs)(ee.Provider, {
		value: n,
		children: [t, n && /* @__PURE__ */ (0, K.jsx)(fe, { store: n })]
	});
}
function fe({ store: e }) {
	(0, W.useSyncExternalStore)(e.subscribe, () => e.version);
	let t = e.drafts();
	return !e.error && t.length === 0 ? null : /* @__PURE__ */ (0, K.jsxs)("div", {
		className: "fixed bottom-2 left-2 max-w-sm rounded border bg-background p-2 text-xs shadow",
		children: [e.error && /* @__PURE__ */ (0, K.jsx)("p", {
			role: "alert",
			"data-comment-storage-error": !0,
			children: e.error
		}), t.length > 0 && /* @__PURE__ */ (0, K.jsxs)("details", { children: [
			/* @__PURE__ */ (0, K.jsxs)("summary", { children: [
				"Saved drafts (",
				t.length,
				")"
			] }),
			/* @__PURE__ */ (0, K.jsx)("p", { children: "Drafts remain available here if their original target disappears." }),
			t.map(({ key: t, body: n }) => /* @__PURE__ */ (0, K.jsxs)("div", {
				className: "space-y-1 border-t py-2",
				children: [
					/* @__PURE__ */ (0, K.jsx)("p", {
						className: "whitespace-pre-wrap",
						children: n
					}),
					/* @__PURE__ */ (0, K.jsx)("button", {
						type: "button",
						onClick: () => void c(n),
						children: "Copy draft"
					}),
					" ",
					/* @__PURE__ */ (0, K.jsx)("button", {
						type: "button",
						onClick: () => e.discardDraft(t),
						children: "Discard draft"
					})
				]
			}, t))
		] })]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentProvider.tsx
var pe = {};
function Z(e) {
	return /* @__PURE__ */ (0, K.jsx)(X, {
		...e.persistenceKey ? { scope: e.persistenceKey } : {},
		children: /* @__PURE__ */ (0, K.jsx)(Q, { ...e })
	});
}
function Q({ children: e, comments: t, config: n, resolveAnchor: r = se, onRailModeChange: i, onScrollToAnchor: a, getAnchorTop: o, onCreate: s, onReply: c, onUpdateStatus: l, onClose: u, onAcknowledge: d, onAssign: f, onUpdateRating: p, onDelete: m, onEdit: h, onEditReply: g, onDeleteReply: _, onCreateTodo: v, onChecklistToggle: y, onMention: b, onApplyProposal: x, onRejectProposal: S, onRequestRevision: C, loadProposalDiff: w, initialRailMode: ee = "closed", railHeaderActions: E, persistenceKey: D }) {
	let [k, A] = O("rail", ee), [j, M] = O("focused-anchor", null), [N, P] = (0, W.useState)(null), [ae, oe] = (0, W.useState)(D);
	ae !== D && (oe(D), P(null));
	let F = (0, W.useRef)({
		setRailMode: A,
		setFocusedAnchor: M
	}), I = (0, W.useRef)(k), L = (0, W.useRef)(i), R = (0, W.useRef)(a), z = (0, W.useRef)(o), B = (0, W.useRef)(r), V = (0, W.useRef)({}), H = (0, W.useRef)(null), U = (0, W.useMemo)(() => te(t), [t]), G = (0, W.useMemo)(() => T(t), [t]), [q] = (0, W.useState)(ce);
	(0, W.useLayoutEffect)(() => {
		F.current = {
			setRailMode: A,
			setFocusedAnchor: M
		}, I.current = k, L.current = i, R.current = a, z.current = o, B.current = r;
	}, [
		A,
		M,
		k,
		i,
		a,
		o,
		r
	]);
	let J = (0, W.useCallback)((e) => {
		F.current.setFocusedAnchor(e), I.current = "focused", F.current.setRailMode("focused"), L.current?.("focused");
	}, []), de = (0, W.useCallback)(() => {
		let e = I.current === "all" ? "closed" : "all";
		I.current = e, F.current.setRailMode(e), L.current?.(e);
	}, []), Y = (0, W.useCallback)(() => {
		I.current = "closed", F.current.setRailMode("closed"), L.current?.("closed");
	}, []), X = (0, W.useCallback)((e, t) => {
		if (!t) {
			delete V.current[e];
			return;
		}
		let n = V.current[e];
		if (n && n !== t) throw Error(`Comment anchor "${e}" is registered by multiple elements`);
		V.current[e] = t;
	}, []), fe = (0, W.useCallback)((e) => {
		if (z.current) return z.current(e);
		let t = B.current(e, Object.keys(V.current)), n = t ? V.current[t] : null, r = H.current;
		return !n || !r ? null : n.getBoundingClientRect().top - r.getBoundingClientRect().top;
	}, []), Z = (0, W.useCallback)((e, t = {}) => {
		if (R.current) return R.current(e, t);
		let n = B.current(e, Object.keys(V.current)), r = n ? V.current[n] : null, i = H.current;
		if (!r || !i) return !1;
		let a = i.getBoundingClientRect(), o = r.getBoundingClientRect(), s = t.offset ?? 0, c = t.block === "center" ? i.scrollTop + o.top - a.top + o.height / 2 - i.clientHeight / 2 - s : i.scrollTop + o.top - a.top - s;
		return i.scrollTo({
			top: Math.max(0, c),
			behavior: t.behavior ?? "smooth"
		}), !0;
	}, []);
	(0, W.useEffect)(() => {
		k === "focused" && !j && Y();
	}, [
		Y,
		k,
		j
	]);
	let Q = (0, W.useMemo)(() => ({
		...s ? { onCreate: s } : {},
		...c ? { onReply: c } : {},
		...l ? { onUpdateStatus: l } : {},
		...u ? { onClose: u } : {},
		...d ? { onAcknowledge: d } : {},
		...f ? { onAssign: f } : {},
		...p ? { onUpdateRating: p } : {},
		...m ? { onDelete: m } : {},
		...h ? { onEdit: h } : {},
		...g ? { onEditReply: g } : {},
		..._ ? { onDeleteReply: _ } : {},
		...v ? { onCreateTodo: v } : {},
		...y ? { onChecklistToggle: y } : {},
		...b ? { onMention: b } : {},
		...x ? { onApplyProposal: x } : {},
		...S ? { onRejectProposal: S } : {},
		...C ? { onRequestRevision: C } : {},
		...w ? { loadProposalDiff: w } : {}
	}), [
		s,
		c,
		l,
		u,
		d,
		f,
		p,
		m,
		h,
		g,
		_,
		v,
		y,
		b,
		x,
		S,
		C,
		w
	]), me = ne(Q) ?? pe;
	(0, W.useLayoutEffect)(() => {
		q.update(G, k, j, N);
	}, [
		q,
		G,
		k,
		j,
		N
	]);
	let [he] = (0, W.useState)(() => ({
		focusAnchor: J,
		registerAnchor: X,
		scrollToAnchor: Z,
		contentRef: H
	})), ge = (0, W.useMemo)(() => ({
		comments: t,
		config: n,
		callbacks: Q,
		commentCounts: U,
		commentMeta: G,
		railMode: k,
		railHeaderActions: E,
		focusedAnchor: j,
		focusAnchor: J,
		openCommentList: de,
		closeRail: Y,
		registerAnchor: X,
		getAnchorTop: fe,
		scrollToAnchor: Z,
		contentRef: H,
		highlightAnchor: N,
		setHighlightAnchor: P
	}), [
		t,
		n,
		Q,
		U,
		G,
		k,
		E,
		j,
		J,
		de,
		Y,
		X,
		fe,
		Z,
		N
	]);
	return /* @__PURE__ */ (0, K.jsx)(re.Provider, {
		value: q,
		children: /* @__PURE__ */ (0, K.jsx)(le.Provider, {
			value: he,
			children: /* @__PURE__ */ (0, K.jsx)(ie.Provider, {
				value: ge,
				children: /* @__PURE__ */ (0, K.jsx)(ue.Provider, {
					value: me,
					children: e
				})
			})
		})
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentSidePanel.tsx
function me(e) {
	return e === "__document__" ? "General" : e.replace(/\[\d+\]/g, "").replaceAll(".", " › ").replaceAll("_", " ").trim();
}
function he(e) {
	return (t) => e.anchorLabels?.[t] ?? e.formatAnchorLabel?.(t) ?? me(t);
}
function ge({ roots: e, drift: t, label: n }) {
	let r = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = n.anchor ?? "__document__";
		t[e] && r.set(e, [...r.get(e) ?? [], n.id]);
	}
	return [...r].map(([e, r]) => ({
		ids: r,
		anchor: e,
		label: n(e),
		drift: t[e]
	}));
}
function _e({ label: e }) {
	return /* @__PURE__ */ (0, K.jsx)("span", {
		"data-testid": "comment-location-meta",
		title: e,
		className: "inline-block max-w-full truncate rounded-full bg-muted px-2 py-1 align-middle text-[10px] font-medium text-muted-foreground",
		children: e
	});
}
function ve({ active: e, onClick: t, children: n, testId: r }) {
	return /* @__PURE__ */ (0, K.jsxs)("button", {
		type: "button",
		onClick: t,
		"data-testid": r,
		className: i("inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium shadow-sm transition-colors [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:text-sm", e ? "border-primary bg-primary text-primary-foreground" : "border-border bg-background text-muted-foreground hover:text-foreground"),
		children: [/* @__PURE__ */ (0, K.jsx)(a, {
			icon: s,
			className: "text-xs"
		}), n]
	});
}
function ye(e) {
	return Object.keys(e.commentMeta).sort((t, n) => {
		if (t === "__document__") return -1;
		if (n === "__document__") return 1;
		let r = e.getAnchorTop(t), i = e.getAnchorTop(n);
		return r == null && i == null ? t.localeCompare(n) : r == null ? 1 : i == null ? -1 : r === i ? t.localeCompare(n) : r - i;
	});
}
function be(e, t) {
	let n = C(e);
	return w(e).sort((e, n) => {
		let r = t.get(e.anchor ?? "__document__") ?? 2 ** 53 - 1, i = t.get(n.anchor ?? "__document__") ?? 2 ** 53 - 1;
		return r === i ? e.createdAt === n.createdAt ? String(e.id).localeCompare(String(n.id)) : e.createdAt.localeCompare(n.createdAt) : r - i;
	}).flatMap((e) => [e, ...F(n.get(e.id) ?? [])]);
}
function xe({ ctx: e, comments: t, label: n, emptyLabel: r, threadToMarkdown: i, renderRootDetails: a, extraAssignItems: o, detailOptions: s, drift: c }) {
	let l = (0, W.useMemo)(() => be(t, new Map(ye(e).map((e, t) => [e, t]))), [e, t]), u = H(l, e.config, e.callbacks);
	function d(t, n) {
		let r = t.target;
		if (r.closest("button, a, details, input, textarea, select, [contenteditable=\"true\"]")) return;
		let i = r.closest("[role=\"button\"]");
		i && i !== r.closest("[data-comment-kind]") || (n === "__document__" || e.scrollToAnchor(n, {
			behavior: "smooth",
			block: "start",
			offset: 12
		})) && e.focusAnchor(n);
	}
	return l.length === 0 ? /* @__PURE__ */ (0, K.jsx)("p", {
		className: "rounded-lg bg-muted/40 px-3 py-2 text-xs text-muted-foreground",
		children: r
	}) : /* @__PURE__ */ (0, K.jsx)("div", {
		"data-testid": "comment-all-rail",
		className: "space-y-3 overflow-y-auto pr-1",
		children: /* @__PURE__ */ (0, K.jsx)(V, {
			...s,
			comments: l,
			config: e.config,
			compact: !0,
			renderRootMeta: (t) => {
				let r = t.anchor ?? "__document__", i = c?.[r]?.state, a = r === "__document__" || e.getAnchorTop(r) != null;
				return /* @__PURE__ */ (0, K.jsx)(_e, { label: `${n(r)}${i ? ` · ${oe[i]}` : a ? "" : " · Unavailable"}` });
			},
			getThreadProps: (t) => ({
				"data-testid": "comment-feed-item",
				onMouseEnter: () => e.setHighlightAnchor(t.anchor ?? "__document__"),
				onMouseLeave: () => e.setHighlightAnchor(null),
				onClick: (e) => d(e, t.anchor ?? "__document__")
			}),
			...u,
			...a ? { renderRootDetails: a } : {},
			...o ? { extraAssignItems: o } : {},
			...i ? { threadToMarkdown: i } : {}
		})
	});
}
function Se({ ctx: e, comments: t, anchor: n, label: r, compact: i, threadToMarkdown: a, renderRootDetails: o, createActions: s, extraAssignItems: c, hideComposer: l, detailOptions: u }) {
	let d = D(t, n), f = d.length > 0;
	return /* @__PURE__ */ (0, K.jsxs)("div", {
		className: "space-y-3",
		"data-comment-anchor": n,
		children: [/* @__PURE__ */ (0, K.jsx)(_e, { label: n === "__document__" ? "Whole-page comment" : `Attached to ${r}` }), /* @__PURE__ */ (0, K.jsx)(A, {
			...u,
			comments: d,
			config: e.config,
			anchor: n,
			compact: i ?? !1,
			autoFocusComposer: !f,
			hideComposer: l ?? !1,
			defaultExpanded: f,
			composerPlaceholder: f ? "Add another top-level comment…" : "Add a comment…",
			...e.callbacks,
			...o ? { renderRootDetails: o } : {},
			...s ? { createActions: s } : {},
			...c ? { extraAssignItems: c } : {},
			...a ? { threadToMarkdown: a } : {}
		})]
	});
}
function Ce(e) {
	let t = {
		...e.previewLines === void 0 ? {} : { previewLines: e.previewLines },
		...e.detailTabs ? { detailTabs: e.detailTabs } : {},
		...e.onMaximizedChange ? { onMaximizedChange: e.onMaximizedChange } : {}
	}, n = N(), r = he(e), a = (0, W.useRef)(null), o = (0, W.useRef)(null), [s, c] = (0, W.useState)(null), [l, u] = O("stage", "active"), d = n?.railMode === "focused" ? n.focusedAnchor : null;
	if ((0, W.useEffect)(() => {
		if (!n || e.focusedAlignment !== "anchor" || !d) {
			c(null);
			return;
		}
		let t = n.contentRef.current, r = a.current, i = o.current;
		if (!t || !r || !i) return;
		let s = () => {
			let e = n.getAnchorTop(d), a = t.getBoundingClientRect(), o = a.height > 0 && r.getBoundingClientRect().top >= a.bottom;
			if (e == null || o) {
				c(null);
				return;
			}
			let s = e + t.getBoundingClientRect().top, l = (Number.parseFloat(i.style.top) || 0) + s - i.getBoundingClientRect().top;
			c(Math.max(0, l));
		};
		s(), window.addEventListener("scroll", s, {
			capture: !0,
			passive: !0
		}), window.addEventListener("resize", s);
		let l = typeof ResizeObserver > "u" ? null : new ResizeObserver(s);
		return l?.observe(t), l?.observe(r), l?.observe(i), () => {
			window.removeEventListener("scroll", s, !0), window.removeEventListener("resize", s), l?.disconnect();
		};
	}, [
		n,
		d,
		e.focusedAlignment
	]), e.onRetarget && !e.pickAnchor) throw Error("CommentSidePanel: `onRetarget` needs `pickAnchor`, the pick that finds the new anchor");
	if (!n) return null;
	let f = e.onRetarget && e.pickAnchor ? {
		pickAnchor: e.pickAnchor,
		onRetarget: e.onRetarget
	} : void 0, p = w(n.comments).reduce((e, t) => {
		let r = U(n.config, t.status);
		return r && (e[r] += 1), e;
	}, {
		active: 0,
		resolved: 0,
		closed: 0
	}), m = k(n.comments, n.config, l);
	if (n.railMode === "closed" && p.active + p.resolved + p.closed === 0) return null;
	function h(e) {
		u(e), n?.railMode !== "all" && n?.openCommentList();
	}
	return /* @__PURE__ */ (0, K.jsxs)("aside", {
		ref: a,
		"data-testid": "comment-side-panel",
		className: i("relative w-[320px] max-w-full space-y-3", e.className),
		children: [
			/* @__PURE__ */ (0, K.jsxs)("div", {
				"data-testid": "comment-rail-header",
				className: "sticky top-0 z-20 flex flex-wrap items-center gap-2 bg-background py-2",
				children: [
					(p.active > 0 || n.railMode === "all") && /* @__PURE__ */ (0, K.jsx)(ve, {
						active: n.railMode === "all" && l === "active",
						onClick: () => h("active"),
						testId: "comment-open-all",
						children: n.railMode === "closed" ? `Open comments (${p.active})` : n.railMode === "focused" ? `All comments (${p.active})` : `Open (${p.active})`
					}),
					p.resolved > 0 && /* @__PURE__ */ (0, K.jsxs)("button", {
						type: "button",
						"aria-pressed": n.railMode === "all" && l === "resolved",
						onClick: () => h("resolved"),
						className: i("rounded-full border px-2 py-1 text-xs font-medium transition-colors [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm", n.railMode === "all" && l === "resolved" ? "border-primary bg-primary text-primary-foreground" : "border-border bg-background text-muted-foreground hover:text-foreground"),
						children: [
							"Resolved (",
							p.resolved,
							")"
						]
					}),
					p.closed > 0 && /* @__PURE__ */ (0, K.jsxs)("button", {
						type: "button",
						"aria-pressed": n.railMode === "all" && l === "closed",
						onClick: () => h("closed"),
						className: i("rounded-full border px-2 py-1 text-xs font-medium transition-colors [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm", n.railMode === "all" && l === "closed" ? "border-primary bg-primary text-primary-foreground" : "border-border bg-background text-muted-foreground hover:text-foreground"),
						children: [
							"Closed (",
							p.closed,
							")"
						]
					}),
					n.railMode !== "closed" && /* @__PURE__ */ (0, K.jsx)("button", {
						type: "button",
						onClick: n.closeRail,
						className: "ml-auto rounded-md px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
						children: n.railMode === "all" ? "Close" : "Hide"
					})
				]
			}),
			n.railMode !== "closed" && e.drift && /* @__PURE__ */ (0, K.jsx)(L, {
				items: ge({
					roots: w(m),
					drift: e.drift,
					label: r
				}),
				retarget: f,
				onSelect: (e) => n.focusAnchor(e.anchor)
			}),
			n.railMode === "all" ? /* @__PURE__ */ (0, K.jsx)(xe, {
				detailOptions: t,
				drift: e.drift,
				ctx: n,
				comments: m,
				label: r,
				emptyLabel: l === "active" ? "No open comments." : l === "resolved" ? "No resolved comments." : "No closed comments.",
				...e.threadToMarkdown ? { threadToMarkdown: e.threadToMarkdown } : {},
				...e.renderRootDetails ? { renderRootDetails: e.renderRootDetails } : {},
				...e.extraAssignItems ? { extraAssignItems: e.extraAssignItems } : {}
			}) : n.railMode === "focused" && n.focusedAnchor ? /* @__PURE__ */ (0, K.jsx)("div", {
				ref: o,
				"data-testid": "comment-focused-rail",
				className: "relative space-y-3",
				style: s == null ? void 0 : { top: s },
				children: /* @__PURE__ */ (0, K.jsx)(Se, {
					detailOptions: t,
					hideComposer: e.hideComposer ?? !1,
					ctx: n,
					comments: n.comments,
					anchor: n.focusedAnchor,
					...e.createActions ? { createActions: e.createActions } : {},
					label: r(n.focusedAnchor),
					...e.compact === void 0 ? {} : { compact: e.compact },
					...e.threadToMarkdown ? { threadToMarkdown: e.threadToMarkdown } : {},
					...e.renderRootDetails ? { renderRootDetails: e.renderRootDetails } : {},
					...e.extraAssignItems ? { extraAssignItems: e.extraAssignItems } : {}
				})
			}) : null
		]
	});
}
// ../../../clicky-ui/packages/ui/src/monaco/source-comments.ts
function we(e, t) {
	let n = t()?.onCreate ? e.addAction({
		id: "comment-selection",
		label: "Comment on selection",
		contextMenuGroupId: "navigation",
		contextMenuOrder: 2,
		run: () => {
			let n = e.getSelection();
			n && t()?.onCreate?.({
				startLine: n.startLineNumber,
				endLine: n.endColumn === 1 && n.endLineNumber > n.startLineNumber ? n.endLineNumber - 1 : n.endLineNumber
			});
		}
	}) : void 0;
	e.updateOptions({ glyphMargin: !0 });
	let r = e.onMouseDown((e) => {
		if (e.target.type !== 2 || !e.target.position) return;
		let n = e.target.position.lineNumber, r = t()?.ranges.find((e) => e.startLine <= n && e.endLine >= n);
		r ? t()?.onSelect(r.id) : t()?.canCreateOnLine?.(n) !== !1 && t()?.onCreate?.({
			startLine: n,
			endLine: n
		});
	});
	return { dispose() {
		n?.dispose(), r.dispose();
	} };
}
// ../../../clicky-ui/packages/ui/src/file-findings/findings-languages.ts
var Te = {
	conf: v,
	language: _
}, Ee = {
	"findings-typescript": {
		conf: g,
		language: h
	},
	"findings-javascript": Te,
	"findings-json": Te,
	"findings-css": {
		conf: f,
		language: d
	},
	"findings-scss": {
		conf: S,
		language: x
	},
	"findings-less": {
		conf: b,
		language: y
	},
	"findings-html": {
		conf: m,
		language: p
	}
};
function De(e) {
	let t = new Set(e.getLanguages().map((e) => e.id));
	for (let [n, { conf: r, language: i }] of Object.entries(Ee)) t.has(n) || (e.register({ id: n }), e.setMonarchTokensProvider(n, i), e.setLanguageConfiguration(n, r));
}
// ../../../clicky-ui/packages/ui/src/file-findings/AnnotatedCode.tsx
var Oe = "findings", ke = {
	error: "bg-red-500 opacity-10",
	warning: "bg-amber-500 opacity-10",
	info: "bg-sky-500 opacity-10"
}, $ = "ml-1 mt-1 size-3 rounded-full", Ae = {
	error: `${$} bg-red-500`,
	warning: `${$} bg-amber-500`,
	info: `${$} bg-sky-500`
}, je = `${$} bg-primary`, Me = 20;
function Ne(e, t) {
	let { MarkerSeverity: n } = e;
	return {
		error: n.Error,
		warning: n.Warning,
		info: n.Info
	}[t];
}
function Pe(e, t) {
	return `file:///${(t.startsWith("/") ? t : `${e ?? ""}/${t}`).replace(/^\/+/, "").split("/").map(encodeURIComponent).join("/")}`;
}
function Fe(e, t, n) {
	return e.map((e) => {
		let r = e.endLine ?? e.line;
		return {
			severity: Ne(n, e.severity),
			message: e.message,
			source: e.source,
			...e.rule ? { code: e.rule } : {},
			startLineNumber: e.line,
			startColumn: e.column ?? 1,
			endLineNumber: r,
			endColumn: e.endColumn ?? t.getLineMaxColumn(r)
		};
	});
}
function Ie(e, t, n) {
	let [r, i] = (0, W.useState)([]);
	return (0, W.useEffect)(() => {
		if (!e || !n) return i([]);
		let r = /* @__PURE__ */ new Map();
		for (let e of t) r.set(e.line, [...r.get(e.line) ?? [], e]);
		let a = [...r].map(([e, t]) => ({
			line: e,
			findings: t.sort(ae),
			node: document.createElement("div")
		})), o = [];
		return e.editor.changeViewZones((e) => {
			for (let t of a) o.push(e.addZone({
				afterLineNumber: t.line,
				heightInPx: t.findings.length * Me,
				domNode: t.node,
				suppressMouseDown: !0
			}));
		}), i(a), () => e.editor.changeViewZones((e) => o.forEach((t) => e.removeZone(t)));
	}, [
		e,
		t,
		n
	]), r;
}
function Le(e, t) {
	return e.find((e) => e.line <= t && t <= (e.endLine ?? e.line));
}
function Re({ mounted: e, onMount: t, contentElement: n, path: r, workDir: a, source: o, findings: s, line: c, onLineChange: d, getFindingActions: f, onChange: p, inlineFindings: m, sourceComments: h, className: g }) {
	let _ = P(), { focusAnchor: v, focusedAnchor: y, commentMeta: b } = _, x = (0, W.useMemo)(() => {
		let e = R(s);
		return new Map(e.map((e, t) => [e.id, s[t]]));
	}, [s]);
	(0, W.useEffect)(() => {
		if (!e) return;
		let { editor: t, monaco: n } = e, r = t.getModel();
		if (!r) return;
		t.updateOptions({ glyphMargin: !0 }), n.editor.setModelMarkers(r, Oe, Fe(s, r, n));
		let i = t.createDecorationsCollection(s.map((e) => ({
			range: {
				startLineNumber: e.line,
				startColumn: 1,
				endLineNumber: e.endLine ?? e.line,
				endColumn: 1
			},
			options: {
				isWholeLine: !0,
				className: ke[e.severity],
				glyphMarginClassName: Ae[e.severity],
				glyphMarginHoverMessage: { value: e.message }
			}
		})));
		return () => {
			i.clear(), n.editor.setModelMarkers(r, Oe, []);
		};
	}, [e, s]);
	let S = Ie(e, s, m === !0), C = (0, W.useCallback)((e) => {
		v(B(e.line)), d?.(e.line);
	}, [v, d]), w = (0, W.useRef)(h);
	(0, W.useEffect)(() => {
		w.current = h;
	});
	let T = h !== void 0, ee = h?.onCreate !== void 0;
	return (0, W.useEffect)(() => {
		if (!e || !T) return;
		let t = we(e.editor, () => {
			let e = w.current;
			return e && {
				...e,
				canCreateOnLine: (t) => !Le(s, t) && (e.canCreateOnLine?.(t) ?? !0)
			};
		});
		return () => t.dispose();
	}, [
		e,
		T,
		ee,
		s
	]), (0, W.useEffect)(() => {
		if (!e || !h) return;
		let t = e.editor.createDecorationsCollection(h.ranges.map((e) => ({
			range: {
				startLineNumber: e.startLine,
				startColumn: 1,
				endLineNumber: e.endLine,
				endColumn: 1
			},
			options: {
				isWholeLine: !0,
				glyphMarginClassName: je,
				glyphMarginHoverMessage: { value: "Open comment" }
			}
		})));
		return () => t.clear();
	}, [e, h]), (0, W.useEffect)(() => {
		if (!e) return;
		let { editor: t } = e, n = t.onMouseDown((e) => {
			let t = e.target.position?.lineNumber, n = t === void 0 ? void 0 : Le(s, t);
			n && C(n);
		});
		return () => n.dispose();
	}, [
		e,
		s,
		C
	]), (0, W.useEffect)(() => {
		let t = n.current;
		if (!e || !t) return;
		let r = () => t.dispatchEvent(new Event("scroll")), i = e.editor.onDidScrollChange(r);
		return r(), () => i.dispose();
	}, [e, n]), (0, W.useLayoutEffect)(() => {
		e && c !== void 0 && (e.editor.revealLineInCenter(c), b[B(c)] && v(B(c)));
	}, [
		e,
		c,
		v
	]), /* @__PURE__ */ (0, K.jsxs)("div", {
		className: i("flex h-full min-h-0", g),
		"data-annotated-code": "",
		children: [/* @__PURE__ */ (0, K.jsxs)("div", {
			ref: (e) => {
				_.contentRef.current = e, n.current = e;
			},
			className: "min-w-0 flex-1 [&_[data-slot=monaco-editor]]:h-full [&_[data-slot=monaco-editor]]:rounded-none [&_[data-slot=monaco-editor]]:border-0",
			children: [/* @__PURE__ */ (0, K.jsx)(u, {
				...p ? { onChange: p } : { readOnly: !0 },
				value: o,
				language: Y(r),
				path: Pe(a, r),
				height: "100%",
				beforeMount: (e) => De(e.languages),
				onMount: (e, n) => t({
					editor: e,
					monaco: n
				})
			}), S.map((e) => (0, G.createPortal)(/* @__PURE__ */ (0, K.jsx)("div", {
				className: "leading-5",
				children: e.findings.map((e) => /* @__PURE__ */ (0, K.jsx)(z, {
					finding: e,
					selected: y === B(e.line),
					onSelect: C,
					getFindingActions: f,
					truncate: !0
				}, `${e.rule}:${e.message}`))
			}), e.node, e.line))]
		}), s.length > 0 && /* @__PURE__ */ (0, K.jsx)("div", {
			className: "shrink-0 overflow-y-auto border-l border-border px-density-3",
			children: /* @__PURE__ */ (0, K.jsx)(Ce, {
				focusedAlignment: "anchor",
				formatAnchorLabel: (e) => `line ${M(e)}`,
				renderRootDetails: (e) => {
					let t = x.get(e.id);
					return t && f ? /* @__PURE__ */ (0, K.jsx)(l, { items: f(t) }) : null;
				}
			})
		})]
	});
}
function ze({ findings: e, ...t }) {
	let [n, r] = (0, W.useState)(null), i = (0, W.useRef)(null), a = (0, W.useMemo)(() => R(e), [e]), o = (0, W.useCallback)((e) => {
		let t = i.current, r = n?.editor.getModel(), a = n?.editor.getDomNode();
		if (!n || !t || !r || !a) return null;
		let o = M(e);
		return o > r.getLineCount() ? null : a.getBoundingClientRect().top - t.getBoundingClientRect().top + n.editor.getTopForLineNumber(o) - n.editor.getScrollTop();
	}, [n]), s = (0, W.useCallback)((e, t) => {
		if (!n) return !1;
		let { editor: r, monaco: i } = n, a = M(e);
		return a > (r.getModel()?.getLineCount() ?? 0) ? !1 : (r.revealLineInCenter(a, t.behavior === "smooth" ? i.editor.ScrollType.Smooth : i.editor.ScrollType.Immediate), !0);
	}, [n]);
	return /* @__PURE__ */ (0, K.jsx)(Z, {
		comments: a,
		config: j,
		resolveAnchor: I,
		getAnchorTop: o,
		onScrollToAnchor: s,
		initialRailMode: "all",
		children: /* @__PURE__ */ (0, K.jsx)(Re, {
			mounted: n,
			onMount: r,
			contentElement: i,
			findings: e,
			...t
		})
	});
}
export { ze as AnnotatedCode };
