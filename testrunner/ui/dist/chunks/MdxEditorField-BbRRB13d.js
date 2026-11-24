import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { a as r, t as i } from "./button-BLoR2pgR.js";
import { n as a, r as o, t as s, v as c } from "./Callout-BqeseRSA.js";
import { f as l, t as u } from "./clearable-field-group-D6WgCteb.js";
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-jsx-attributes.ts
var d = /* @__PURE__ */ e(t(), 1);
function f(e, t) {
	let n = {};
	for (let r of e ?? []) {
		if (r.type !== "mdxJsxAttribute") continue;
		let e = r.name;
		e && t.includes(e) && (typeof r.value == "string" ? n[e] = r.value : (r.value === null || r.value === void 0) && (n[e] = "true"));
	}
	return n;
}
function p(e, t, n) {
	let r = (e ?? []).filter((e) => e.type !== "mdxJsxAttribute" || e.name !== t);
	return n === "" ? r : [...r, {
		type: "mdxJsxAttribute",
		name: t,
		value: n
	}];
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-callout-model.ts
var m = [
	"variant",
	"icon",
	"title",
	"label",
	"badge",
	"source",
	"emphasis"
], h = m.map((e) => ({
	name: e,
	type: "string"
})), g = {
	variant: o,
	icon: a
};
function _(e) {
	return f(e, m);
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorCallout.tsx
var v = n(), y = "h-6 rounded border border-border bg-background px-1 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring", b = r(y, "w-24 placeholder:text-placeholder"), x = r(y, "w-56 placeholder:text-placeholder");
function S({ attribute: e, value: t, onChange: n }) {
	if (e === "emphasis") return /* @__PURE__ */ (0, v.jsxs)("label", {
		className: "flex items-center gap-1 text-xs text-muted-foreground",
		children: [/* @__PURE__ */ (0, v.jsx)("input", {
			type: "checkbox",
			checked: t !== "" && t !== "false",
			onChange: (e) => n(e.target.checked ? "true" : "")
		}), "emphasis"]
	});
	let r = g[e];
	return /* @__PURE__ */ (0, v.jsxs)("label", {
		className: "flex items-center gap-1 text-xs text-muted-foreground",
		children: [e, r ? /* @__PURE__ */ (0, v.jsxs)("select", {
			className: y,
			value: t,
			onChange: (e) => n(e.target.value),
			children: [/* @__PURE__ */ (0, v.jsx)("option", {
				value: "",
				children: "—"
			}), r.map((e) => /* @__PURE__ */ (0, v.jsx)("option", {
				value: e,
				children: e
			}, e))]
		}) : /* @__PURE__ */ (0, v.jsx)("input", {
			className: e === "title" ? x : b,
			value: t,
			placeholder: "—",
			onChange: (e) => n(e.target.value)
		})]
	});
}
function C({ mdx: e, mdastNode: t }) {
	let n = e.NestedLexicalEditor, r = e.useMdastNodeUpdater(), i = _(t.attributes), a = (e) => (n) => {
		r({ attributes: p(t.attributes, e, n) });
	};
	return /* @__PURE__ */ (0, v.jsxs)("div", {
		className: "my-2",
		"data-callout-editor": i.variant ?? "default",
		children: [/* @__PURE__ */ (0, v.jsx)("div", {
			className: "mb-1 flex flex-wrap items-center gap-3 rounded-t border border-border bg-muted/40 px-2 py-1",
			children: m.map((e) => /* @__PURE__ */ (0, v.jsx)(S, {
				attribute: e,
				value: i[e] ?? "",
				onChange: a(e)
			}, e))
		}), /* @__PURE__ */ (0, v.jsx)(s, {
			badge: i.badge,
			emphasis: i.emphasis,
			icon: i.icon,
			label: i.label,
			source: i.source,
			title: i.title,
			variant: i.variant ?? "default",
			children: /* @__PURE__ */ (0, v.jsx)(n, {
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
function w({ mdx: e, name: t }) {
	let n = e.ButtonWithTooltip, r = e.usePublisher(e.insertJsx$);
	return /* @__PURE__ */ (0, v.jsx)(n, {
		title: "Insert callout",
		onClick: () => r({
			name: t,
			kind: "flow",
			props: { variant: "note" },
			children: [{
				type: "paragraph",
				children: []
			}]
		}),
		children: /* @__PURE__ */ (0, v.jsx)(c, { size: "1em" })
	});
}
var T = /* @__PURE__ */ new WeakMap(), E = /* @__PURE__ */ new WeakMap();
function D(e) {
	let t = T.get(e);
	if (t) return t;
	let n = ({ mdastNode: t }) => (0, d.createElement)(C, {
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
	let a = () => (0, d.createElement)(w, {
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
		props: h,
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
		contents: () => /* @__PURE__ */ (0, v.jsx)(r, {})
	}] : [], { fallback: () => /* @__PURE__ */ (0, v.jsx)(P, {
		mdx: e,
		options: t
	}) }];
	return /* @__PURE__ */ (0, v.jsx)(n, { options: i });
}
function P({ mdx: e, options: t }) {
	let { BlockTypeSelect: n, BoldItalicUnderlineToggles: r, CodeToggle: i, CreateLink: a, DiffSourceToggleWrapper: o, InsertAdmonition: s, InsertCodeBlock: c, InsertFrontmatter: l, InsertImage: u, InsertTable: d, InsertThematicBreak: f, ListsToggle: p, Separator: m, UndoRedo: h } = e, g = O(e, j(t.callouts) ?? {}), _ = typeof t.toolbar == "object" ? t.toolbar.extra : void 0, y = /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [
		/* @__PURE__ */ (0, v.jsx)(h, {}),
		/* @__PURE__ */ (0, v.jsx)(m, {}),
		/* @__PURE__ */ (0, v.jsx)(r, {}),
		/* @__PURE__ */ (0, v.jsx)(i, {}),
		A(t.headings) && /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [/* @__PURE__ */ (0, v.jsx)(m, {}), /* @__PURE__ */ (0, v.jsx)(n, {})] }),
		A(t.lists) && /* @__PURE__ */ (0, v.jsx)(p, {}),
		A(t.links) && /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [/* @__PURE__ */ (0, v.jsx)(m, {}), /* @__PURE__ */ (0, v.jsx)(a, {})] }),
		A(t.images) && /* @__PURE__ */ (0, v.jsx)(u, {}),
		(A(t.tables) || A(t.thematicBreak) || A(t.codeBlocks)) && /* @__PURE__ */ (0, v.jsx)(m, {}),
		A(t.tables) && /* @__PURE__ */ (0, v.jsx)(d, {}),
		A(t.thematicBreak) && /* @__PURE__ */ (0, v.jsx)(f, {}),
		A(t.codeBlocks) && /* @__PURE__ */ (0, v.jsx)(c, {}),
		(A(t.frontmatter) || A(t.admonitions) || A(t.callouts, !1)) && /* @__PURE__ */ (0, v.jsx)(m, {}),
		A(t.frontmatter) && /* @__PURE__ */ (0, v.jsx)(l, {}),
		A(t.admonitions) && /* @__PURE__ */ (0, v.jsx)(s, {}),
		A(t.callouts, !1) && /* @__PURE__ */ (0, v.jsx)(g, {}),
		_
	] }), b = M(t.diffMode);
	return b ? /* @__PURE__ */ (0, v.jsx)(o, {
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
	return () => (0, d.createElement)(N, {
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
			return (0, d.createElement)(t.Editor, {
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
	return I(t.callouts, !1) && n.push(k(e, L(t.callouts) ?? {})), t.jsxComponents?.length && n.push(...t.jsxComponents), t.mdxSyntax && n.push({
		name: "*",
		kind: "flow",
		props: [],
		hasChildren: !0,
		Editor: e.GenericJsxEditor
	}), n;
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
	return W ??= import("./dist-Q4DdRFqk.js"), W;
}
function K({ id: e, value: t = "", onChange: n, readOnly: a = !1, disabled: o = !1, size: s = "md", placeholder: c, className: f, contentClassName: p, textareaClassName: m, "aria-label": h, commitInitialMarkdownNormalize: g = !1, clearable: _, onClear: y, clearLabel: b, ...x }) {
	let [S, C] = (0, d.useState)(null), [w, T] = (0, d.useState)(null), E = (0, d.useRef)(null), D = (0, d.useRef)(t), O = a || o;
	(0, d.useEffect)(() => {
		let e = !1;
		return G().then((t) => {
			e || C(t);
		}).catch(() => {
			e || C(null);
		}), () => {
			e = !0;
		};
	}, []), (0, d.useEffect)(() => {
		S && t !== D.current && (D.current = t, T(null), E.current?.setMarkdown(t));
	}, [S, t]);
	let k = (0, d.useCallback)((e, t = !1) => {
		(!t || g) && (D.current = e, n?.(e));
	}, [g, n]), A = (0, d.useMemo)(() => S ? H(S, x) : [], [S, x]), j = (0, d.useCallback)((e) => {
		queueMicrotask(() => T(e));
	}, []), M = w;
	if (!S || M) {
		let n = h ?? (M ? "Markdown source" : e ? void 0 : "Markdown");
		return /* @__PURE__ */ (0, v.jsxs)(u, {
			clearable: _ === !0 && !O && t !== "",
			...y ? { onClear: y } : {},
			...b ? { clearLabel: b } : {},
			children: [
				M && /* @__PURE__ */ (0, v.jsxs)("div", {
					role: "alert",
					className: "mb-2 text-sm text-destructive",
					children: ["Markdown could not be loaded: ", M.error]
				}),
				/* @__PURE__ */ (0, v.jsx)("textarea", {
					id: e,
					"data-jsf-input": !0,
					rows: 6,
					"aria-label": n,
					className: r(l(s), "h-auto resize-y font-mono", U[s], m),
					value: t,
					disabled: O,
					placeholder: typeof c == "string" ? c : void 0,
					onChange: (e) => k(e.target.value)
				}),
				M && !O && /* @__PURE__ */ (0, v.jsx)(i, {
					type: "button",
					variant: "outline",
					size: "sm",
					onClick: () => T(null),
					children: "Retry rich text"
				})
			]
		});
	}
	let { MDXEditor: N } = S;
	return /* @__PURE__ */ (0, v.jsx)(u, {
		clearable: _ === !0 && !O && t !== "",
		...y ? { onClear: y } : {},
		...b ? { clearLabel: b } : {},
		children: /* @__PURE__ */ (0, v.jsx)("div", {
			id: e,
			"data-jsf-input": !0,
			className: "min-w-0",
			children: /* @__PURE__ */ (0, v.jsx)(N, {
				ref: E,
				markdown: t,
				readOnly: O,
				placeholder: c,
				plugins: A,
				className: r("clicky-mdx-editor-field", O && "clicky-mdx-editor-field-readonly", f),
				contentEditableClassName: r("clicky-mdx-editor-content", U[s], p),
				onChange: k,
				onError: j
			})
		})
	});
}
export { K as MdxEditorField };
