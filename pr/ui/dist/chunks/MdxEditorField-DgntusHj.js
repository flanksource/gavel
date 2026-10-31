import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { B as r, h as i, n as a, r as o, t as s } from "./Callout-ArstGfBl.js";
import { o as c, t as l } from "./button-G86_a8z9.js";
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-jsx-attributes.ts
var u = /* @__PURE__ */ e(n(), 1);
function d(e, t) {
	let n = {};
	for (let r of e ?? []) {
		if (r.type !== "mdxJsxAttribute") continue;
		let e = r.name;
		e && t.includes(e) && (typeof r.value == "string" ? n[e] = r.value : (r.value === null || r.value === void 0) && (n[e] = "true"));
	}
	return n;
}
function f(e, t, n) {
	let r = (e ?? []).filter((e) => e.type !== "mdxJsxAttribute" || e.name !== t);
	return n === "" ? r : [...r, {
		type: "mdxJsxAttribute",
		name: t,
		value: n
	}];
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-callout-model.ts
var p = [
	"variant",
	"icon",
	"title",
	"label",
	"badge",
	"source",
	"emphasis"
], m = p.map((e) => ({
	name: e,
	type: "string"
})), h = {
	variant: o,
	icon: a
};
function g(e) {
	return d(e, p);
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorCallout.tsx
var _ = /* @__PURE__ */ e(t(), 1), v = "h-6 rounded border border-border bg-background px-1 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring", y = c(v, "w-24 placeholder:text-placeholder"), b = c(v, "w-56 placeholder:text-placeholder");
function x({ attribute: e, value: t, onChange: n }) {
	if (e === "emphasis") return /* @__PURE__ */ (0, _.jsxs)("label", {
		className: "flex items-center gap-1 text-xs text-muted-foreground",
		children: [/* @__PURE__ */ (0, _.jsx)("input", {
			type: "checkbox",
			checked: t !== "" && t !== "false",
			onChange: (e) => n(e.target.checked ? "true" : "")
		}), "emphasis"]
	});
	let r = h[e];
	return /* @__PURE__ */ (0, _.jsxs)("label", {
		className: "flex items-center gap-1 text-xs text-muted-foreground",
		children: [e, r ? /* @__PURE__ */ (0, _.jsxs)("select", {
			className: v,
			value: t,
			onChange: (e) => n(e.target.value),
			children: [/* @__PURE__ */ (0, _.jsx)("option", {
				value: "",
				children: "—"
			}), r.map((e) => /* @__PURE__ */ (0, _.jsx)("option", {
				value: e,
				children: e
			}, e))]
		}) : /* @__PURE__ */ (0, _.jsx)("input", {
			className: e === "title" ? b : y,
			value: t,
			placeholder: "—",
			onChange: (e) => n(e.target.value)
		})]
	});
}
function S({ mdx: e, mdastNode: t }) {
	let n = e.NestedLexicalEditor, r = e.useMdastNodeUpdater(), i = g(t.attributes), a = (e) => (n) => {
		r({ attributes: f(t.attributes, e, n) });
	};
	return /* @__PURE__ */ (0, _.jsxs)("div", {
		className: "my-2",
		"data-callout-editor": i.variant ?? "default",
		children: [/* @__PURE__ */ (0, _.jsx)("div", {
			className: "mb-1 flex flex-wrap items-center gap-3 rounded-t border border-border bg-muted/40 px-2 py-1",
			children: p.map((e) => /* @__PURE__ */ (0, _.jsx)(x, {
				attribute: e,
				value: i[e] ?? "",
				onChange: a(e)
			}, e))
		}), /* @__PURE__ */ (0, _.jsx)(s, {
			badge: i.badge,
			emphasis: i.emphasis,
			icon: i.icon,
			label: i.label,
			source: i.source,
			title: i.title,
			variant: i.variant ?? "default",
			children: /* @__PURE__ */ (0, _.jsx)(n, {
				block: !0,
				getContent: (e) => e.children ?? [],
				getUpdatedMdastNode: (e, t) => ({
					...e,
					children: t
				})
			})
		})]
	});
}
function C({ mdx: e, name: t }) {
	let n = e.ButtonWithTooltip, i = e.usePublisher(e.insertJsx$);
	return /* @__PURE__ */ (0, _.jsx)(n, {
		title: "Insert callout",
		onClick: () => i({
			name: t,
			kind: "flow",
			props: { variant: "note" },
			children: [{
				type: "paragraph",
				children: []
			}]
		}),
		children: /* @__PURE__ */ (0, _.jsx)(r, { size: "1em" })
	});
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-callout.ts
var w = "CalloutBox", T = /* @__PURE__ */ new WeakMap(), E = /* @__PURE__ */ new WeakMap();
function D(e) {
	let t = T.get(e);
	if (t) return t;
	let n = ({ mdastNode: t }) => (0, u.createElement)(S, {
		mdastNode: t,
		mdx: e
	});
	return T.set(e, n), n;
}
function O(e, t = {}) {
	let n = t.name ?? "CalloutBox", r = E.get(e) ?? /* @__PURE__ */ new Map();
	E.set(e, r);
	let i = r.get(n);
	if (i) return i;
	let a = () => (0, u.createElement)(C, {
		mdx: e,
		name: n
	});
	return r.set(n, a), a;
}
function k(e, t = {}) {
	return {
		name: t.name ?? "CalloutBox",
		kind: "flow",
		...t.source ? { source: t.source } : {},
		props: m,
		hasChildren: !0,
		Editor: D(e)
	};
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorToolbar.tsx
function A(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function j(e) {
	return typeof e == "object" && e ? e : void 0;
}
function M(e) {
	if (e) return typeof e == "object" && e.viewModes?.length ? [...e.viewModes] : [
		"rich-text",
		"source",
		"diff"
	];
}
function N({ mdx: e, options: t }) {
	let { ConditionalContents: n, ChangeCodeMirrorLanguage: r } = e, i = [...A(t.codeBlocks) && A(t.codeMirror) ? [{
		when: (e) => e?.editorType === "codeblock",
		contents: () => /* @__PURE__ */ (0, _.jsx)(r, {})
	}] : [], { fallback: () => /* @__PURE__ */ (0, _.jsx)(P, {
		mdx: e,
		options: t
	}) }];
	return /* @__PURE__ */ (0, _.jsx)(n, { options: i });
}
function P({ mdx: e, options: t }) {
	let { BlockTypeSelect: n, BoldItalicUnderlineToggles: r, CodeToggle: i, CreateLink: a, DiffSourceToggleWrapper: o, InsertAdmonition: s, InsertCodeBlock: c, InsertFrontmatter: l, InsertImage: u, InsertTable: d, InsertThematicBreak: f, ListsToggle: p, Separator: m, UndoRedo: h } = e, g = O(e, j(t.callouts) ?? {}), v = typeof t.toolbar == "object" ? t.toolbar.extra : void 0, y = /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [
		/* @__PURE__ */ (0, _.jsx)(h, {}),
		/* @__PURE__ */ (0, _.jsx)(m, {}),
		/* @__PURE__ */ (0, _.jsx)(r, {}),
		/* @__PURE__ */ (0, _.jsx)(i, {}),
		A(t.headings) && /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [/* @__PURE__ */ (0, _.jsx)(m, {}), /* @__PURE__ */ (0, _.jsx)(n, {})] }),
		A(t.lists) && /* @__PURE__ */ (0, _.jsx)(p, {}),
		A(t.links) && /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [/* @__PURE__ */ (0, _.jsx)(m, {}), /* @__PURE__ */ (0, _.jsx)(a, {})] }),
		A(t.images) && /* @__PURE__ */ (0, _.jsx)(u, {}),
		(A(t.tables) || A(t.thematicBreak) || A(t.codeBlocks)) && /* @__PURE__ */ (0, _.jsx)(m, {}),
		A(t.tables) && /* @__PURE__ */ (0, _.jsx)(d, {}),
		A(t.thematicBreak) && /* @__PURE__ */ (0, _.jsx)(f, {}),
		A(t.codeBlocks) && /* @__PURE__ */ (0, _.jsx)(c, {}),
		(A(t.frontmatter) || A(t.admonitions) || A(t.callouts, !1)) && /* @__PURE__ */ (0, _.jsx)(m, {}),
		A(t.frontmatter) && /* @__PURE__ */ (0, _.jsx)(l, {}),
		A(t.admonitions) && /* @__PURE__ */ (0, _.jsx)(s, {}),
		A(t.callouts, !1) && /* @__PURE__ */ (0, _.jsx)(g, {}),
		v
	] }), b = M(t.diffMode);
	return b ? /* @__PURE__ */ (0, _.jsx)(o, {
		options: b,
		children: y
	}) : y;
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-field-plugins.tsx
var F = {
	bash: "Bash",
	css: "CSS",
	go: "Go",
	html: "HTML",
	javascript: "JavaScript",
	js: "JavaScript",
	json: "JSON",
	jsx: "JSX",
	markdown: "Markdown",
	python: "Python",
	shell: "Shell",
	sql: "SQL",
	ts: "TypeScript",
	tsx: "TSX",
	typescript: "TypeScript",
	yaml: "YAML"
};
function I(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function L(e) {
	return typeof e == "object" && e ? e : void 0;
}
function R(e) {
	if (e) return e === !0 ? {} : typeof e == "string" ? { viewMode: e } : e;
}
function z(e, t) {
	return () => (0, u.createElement)(N, {
		mdx: e,
		options: t
	});
}
function B(e, t) {
	if (t?.length) return t.map((t) => ({
		priority: t.priority,
		match: t.match,
		Editor: (n) => {
			let r = e.useCodeBlockEditorContext();
			return (0, u.createElement)(t.Editor, {
				...n,
				context: {
					setCode: r.setCode,
					setLanguage: r.setLanguage,
					setMeta: r.setMeta
				}
			});
		}
	}));
}
function V(e, t) {
	let n = [];
	return I(t.callouts, !1) && n.push(k(e, L(t.callouts) ?? {})), t.jsxComponents?.length && n.push(...t.jsxComponents), n;
}
function H(e, t = {}) {
	let n = [];
	if (I(t.toolbar)) {
		let r = L(t.toolbar);
		n.push(e.toolbarPlugin({
			toolbarContents: z(e, t),
			...r?.className ? { toolbarClassName: r.className } : {},
			...r?.position ? { toolbarPosition: r.position } : {}
		}));
	}
	if (I(t.headings)) {
		let r = L(t.headings);
		n.push(e.headingsPlugin(r));
	}
	if (I(t.lists) && n.push(e.listsPlugin()), I(t.quote) && n.push(e.quotePlugin()), I(t.links)) {
		let r = L(t.links);
		n.push(e.linkPlugin(r?.disableAutoLink === void 0 ? void 0 : { disableAutoLink: r.disableAutoLink }), e.linkDialogPlugin({
			...r?.linkAutocompleteSuggestions ? { linkAutocompleteSuggestions: r.linkAutocompleteSuggestions } : {},
			...r?.showLinkTitleField === void 0 ? {} : { showLinkTitleField: r.showLinkTitleField }
		}));
	}
	if (I(t.tables) && n.push(e.tablePlugin()), I(t.thematicBreak) && n.push(e.thematicBreakPlugin()), I(t.codeBlocks)) {
		let r = L(t.codeBlocks), i = { defaultCodeBlockLanguage: r?.defaultLanguage ?? "txt" }, a = B(e, r?.editorDescriptors);
		if (a && (i.codeBlockEditorDescriptors = a), n.push(e.codeBlockPlugin(i)), I(t.codeMirror)) {
			let r = L(t.codeMirror);
			n.push(e.codeMirrorPlugin({
				codeBlockLanguages: r?.languages ?? F,
				...r?.autoLoadLanguageSupport === void 0 ? {} : { autoLoadLanguageSupport: r.autoLoadLanguageSupport }
			}));
		}
	}
	let r = V(e, t);
	if (r.length && n.push(e.jsxPlugin({ jsxComponentDescriptors: r })), I(t.frontmatter) && n.push(e.frontmatterPlugin()), I(t.admonitions)) {
		let r = L(t.admonitions);
		n.push(e.directivesPlugin({
			directiveDescriptors: [e.AdmonitionDirectiveDescriptor],
			escapeUnknownTextDirectives: r?.escapeUnknownTextDirectives ?? !0
		}));
	}
	if (I(t.images)) {
		let r = L(t.images);
		n.push(e.imagePlugin({
			...r?.imageUploadHandler ? { imageUploadHandler: r.imageUploadHandler } : {},
			...r?.imageAutocompleteSuggestions ? { imageAutocompleteSuggestions: r.imageAutocompleteSuggestions } : {},
			...r?.imagePreviewHandler ? { imagePreviewHandler: r.imagePreviewHandler } : {},
			...r?.disableImageResize === void 0 ? {} : { disableImageResize: r.disableImageResize },
			...r?.disableImageSettingsButton === void 0 ? {} : { disableImageSettingsButton: r.disableImageSettingsButton },
			...r?.allowSetImageDimensions === void 0 ? {} : { allowSetImageDimensions: r.allowSetImageDimensions }
		}));
	}
	if (t.diffMode) {
		let r = R(t.diffMode);
		n.push(e.diffSourcePlugin({
			...r?.viewMode ? { viewMode: r.viewMode } : {},
			...r?.diffMarkdown === void 0 ? {} : { diffMarkdown: r.diffMarkdown },
			...r?.readOnlyDiff === void 0 ? {} : { readOnlyDiff: r.readOnlyDiff }
		}));
	}
	return I(t.markdownShortcuts) && n.push(e.markdownShortcutPlugin()), t.plugins?.length && n.push(...t.plugins), n;
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorField.tsx
var U = {
	xs: "min-h-32",
	sm: "min-h-36",
	md: "min-h-40",
	lg: "min-h-48",
	xl: "min-h-56"
}, W = null;
function G() {
	return W ??= import("./dist-Cxa-MbPV.js"), W;
}
function K({ id: e, value: t = "", onChange: n, readOnly: r = !1, disabled: a = !1, size: o = "md", placeholder: s, className: d, contentClassName: f, textareaClassName: p, "aria-label": m, commitInitialMarkdownNormalize: h = !1, ...g }) {
	let [v, y] = (0, u.useState)(null), [b, x] = (0, u.useState)(null), S = (0, u.useRef)(null), C = (0, u.useRef)(t), w = r || a;
	(0, u.useEffect)(() => {
		let e = !1;
		return G().then((t) => {
			e || y(t);
		}).catch(() => {
			e || y(null);
		}), () => {
			e = !0;
		};
	}, []), (0, u.useEffect)(() => {
		v && t !== C.current && (C.current = t, x(null), S.current?.setMarkdown(t));
	}, [v, t]);
	let T = (0, u.useCallback)((e, t = !1) => {
		(!t || h) && (C.current = e, n?.(e));
	}, [h, n]), E = (0, u.useMemo)(() => v ? H(v, g) : [], [v, g]), D = (0, u.useCallback)((e) => {
		queueMicrotask(() => x(e));
	}, []), O = b;
	if (!v || O) {
		let n = m ?? (O ? "Markdown source" : e ? void 0 : "Markdown");
		return /* @__PURE__ */ (0, _.jsxs)("div", {
			className: "min-w-0",
			children: [
				O && /* @__PURE__ */ (0, _.jsxs)("div", {
					role: "alert",
					className: "mb-2 text-sm text-destructive",
					children: ["Markdown could not be loaded: ", O.error]
				}),
				/* @__PURE__ */ (0, _.jsx)("textarea", {
					id: e,
					"data-jsf-input": !0,
					rows: 6,
					"aria-label": n,
					className: c(i(o), "h-auto resize-y font-mono", U[o], p),
					value: t,
					disabled: w,
					placeholder: typeof s == "string" ? s : void 0,
					onChange: (e) => T(e.target.value)
				}),
				O && !w && /* @__PURE__ */ (0, _.jsx)(l, {
					type: "button",
					variant: "outline",
					size: "sm",
					onClick: () => x(null),
					children: "Retry rich text"
				})
			]
		});
	}
	let { MDXEditor: k } = v;
	return /* @__PURE__ */ (0, _.jsx)("div", {
		id: e,
		"data-jsf-input": !0,
		className: "min-w-0",
		children: /* @__PURE__ */ (0, _.jsx)(k, {
			ref: S,
			markdown: t,
			readOnly: w,
			placeholder: s,
			plugins: E,
			className: c("clicky-mdx-editor-field", w && "clicky-mdx-editor-field-readonly", d),
			contentEditableClassName: c("clicky-mdx-editor-content", U[o], f),
			onChange: T,
			onError: D
		})
	});
}
export { K as MdxEditorField, h as a, d as c, p as i, f as l, k as n, m as o, O as r, g as s, w as t };
