import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-a1ZmknZu.js";
import { G as r, a as i, g as a, n as o, r as s, t as c } from "./Callout-BvAoqcdf.js";
import { o as l, t as u } from "./button-Dbm88adW.js";
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
	variant: s,
	icon: o
};
function _(e) {
	return f(e, m);
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorCallout.tsx
var v = /* @__PURE__ */ e(n(), 1), y = "h-6 rounded border border-border bg-background px-1 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring", b = l(y, "w-24 placeholder:text-placeholder"), x = l(y, "w-56 placeholder:text-placeholder");
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
		}), /* @__PURE__ */ (0, v.jsx)(c, {
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
	let n = e.ButtonWithTooltip, i = e.usePublisher(e.insertJsx$);
	return /* @__PURE__ */ (0, v.jsx)(n, {
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
		children: /* @__PURE__ */ (0, v.jsx)(r, { size: "1em" })
	});
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-callout.ts
var T = "CalloutBox", E = /* @__PURE__ */ new WeakMap(), D = /* @__PURE__ */ new WeakMap();
function O(e) {
	let t = E.get(e);
	if (t) return t;
	let n = ({ mdastNode: t }) => (0, d.createElement)(C, {
		mdastNode: t,
		mdx: e
	});
	return E.set(e, n), n;
}
function k(e, t = {}) {
	let n = t.name ?? "CalloutBox", r = D.get(e) ?? /* @__PURE__ */ new Map();
	D.set(e, r);
	let i = r.get(n);
	if (i) return i;
	let a = () => (0, d.createElement)(w, {
		mdx: e,
		name: n
	});
	return r.set(n, a), a;
}
function A(e, t = {}) {
	return {
		name: t.name ?? "CalloutBox",
		kind: "flow",
		...t.source ? { source: t.source } : {},
		props: h,
		hasChildren: !0,
		Editor: O(e)
	};
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorToolbar.tsx
function j(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function M(e) {
	return typeof e == "object" && e ? e : void 0;
}
function N(e) {
	if (e) return typeof e == "object" && e.viewModes?.length ? [...e.viewModes] : [
		"rich-text",
		"source",
		"diff"
	];
}
function P({ mdx: e, options: t }) {
	let { ConditionalContents: n, ChangeCodeMirrorLanguage: r } = e, i = [...j(t.codeBlocks) && j(t.codeMirror) ? [{
		when: (e) => e?.editorType === "codeblock",
		contents: () => /* @__PURE__ */ (0, v.jsx)(r, {})
	}] : [], { fallback: () => /* @__PURE__ */ (0, v.jsx)(F, {
		mdx: e,
		options: t
	}) }];
	return /* @__PURE__ */ (0, v.jsx)(n, { options: i });
}
function F({ mdx: e, options: t }) {
	let { BlockTypeSelect: n, BoldItalicUnderlineToggles: r, CodeToggle: i, CreateLink: a, DiffSourceToggleWrapper: o, InsertAdmonition: s, InsertCodeBlock: c, InsertFrontmatter: l, InsertImage: u, InsertTable: d, InsertThematicBreak: f, ListsToggle: p, Separator: m, UndoRedo: h } = e, g = k(e, M(t.callouts) ?? {}), _ = typeof t.toolbar == "object" ? t.toolbar.extra : void 0, y = /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [
		/* @__PURE__ */ (0, v.jsx)(h, {}),
		/* @__PURE__ */ (0, v.jsx)(m, {}),
		/* @__PURE__ */ (0, v.jsx)(r, {}),
		/* @__PURE__ */ (0, v.jsx)(i, {}),
		j(t.headings) && /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [/* @__PURE__ */ (0, v.jsx)(m, {}), /* @__PURE__ */ (0, v.jsx)(n, {})] }),
		j(t.lists) && /* @__PURE__ */ (0, v.jsx)(p, {}),
		j(t.links) && /* @__PURE__ */ (0, v.jsxs)(v.Fragment, { children: [/* @__PURE__ */ (0, v.jsx)(m, {}), /* @__PURE__ */ (0, v.jsx)(a, {})] }),
		j(t.images) && /* @__PURE__ */ (0, v.jsx)(u, {}),
		(j(t.tables) || j(t.thematicBreak) || j(t.codeBlocks)) && /* @__PURE__ */ (0, v.jsx)(m, {}),
		j(t.tables) && /* @__PURE__ */ (0, v.jsx)(d, {}),
		j(t.thematicBreak) && /* @__PURE__ */ (0, v.jsx)(f, {}),
		j(t.codeBlocks) && /* @__PURE__ */ (0, v.jsx)(c, {}),
		(j(t.frontmatter) || j(t.admonitions) || j(t.callouts, !1)) && /* @__PURE__ */ (0, v.jsx)(m, {}),
		j(t.frontmatter) && /* @__PURE__ */ (0, v.jsx)(l, {}),
		j(t.admonitions) && /* @__PURE__ */ (0, v.jsx)(s, {}),
		j(t.callouts, !1) && /* @__PURE__ */ (0, v.jsx)(g, {}),
		_
	] }), b = N(t.diffMode);
	return b ? /* @__PURE__ */ (0, v.jsx)(o, {
		options: b,
		children: y
	}) : y;
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-field-plugins.tsx
var I = {
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
function L(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function R(e) {
	return typeof e == "object" && e ? e : void 0;
}
function z(e) {
	if (e) return e === !0 ? {} : typeof e == "string" ? { viewMode: e } : e;
}
function B(e, t) {
	return () => (0, d.createElement)(P, {
		mdx: e,
		options: t
	});
}
function V(e, t) {
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
function H(e, t) {
	let n = [];
	return L(t.callouts, !1) && n.push(A(e, R(t.callouts) ?? {})), t.jsxComponents?.length && n.push(...t.jsxComponents), t.mdxSyntax && n.push({
		name: "*",
		kind: "flow",
		props: [],
		hasChildren: !0,
		Editor: e.GenericJsxEditor
	}), n;
}
function U(e, t = {}) {
	let n = [];
	if (L(t.toolbar)) {
		let r = R(t.toolbar);
		n.push(e.toolbarPlugin({
			toolbarContents: B(e, t),
			...r?.className ? { toolbarClassName: r.className } : {},
			...r?.position ? { toolbarPosition: r.position } : {}
		}));
	}
	if (L(t.headings)) {
		let r = R(t.headings);
		n.push(e.headingsPlugin(r));
	}
	if (L(t.lists) && n.push(e.listsPlugin()), L(t.quote) && n.push(e.quotePlugin()), L(t.links)) {
		let r = R(t.links);
		n.push(e.linkPlugin(r?.disableAutoLink === void 0 ? void 0 : { disableAutoLink: r.disableAutoLink }), e.linkDialogPlugin({
			...r?.linkAutocompleteSuggestions ? { linkAutocompleteSuggestions: r.linkAutocompleteSuggestions } : {},
			...r?.showLinkTitleField === void 0 ? {} : { showLinkTitleField: r.showLinkTitleField }
		}));
	}
	if (L(t.tables) && n.push(e.tablePlugin()), L(t.thematicBreak) && n.push(e.thematicBreakPlugin()), L(t.codeBlocks)) {
		let r = R(t.codeBlocks), i = { defaultCodeBlockLanguage: r?.defaultLanguage ?? "txt" }, a = V(e, r?.editorDescriptors);
		if (a && (i.codeBlockEditorDescriptors = a), n.push(e.codeBlockPlugin(i)), L(t.codeMirror)) {
			let r = R(t.codeMirror);
			n.push(e.codeMirrorPlugin({
				codeBlockLanguages: r?.languages ?? I,
				...r?.autoLoadLanguageSupport === void 0 ? {} : { autoLoadLanguageSupport: r.autoLoadLanguageSupport }
			}));
		}
	}
	let r = H(e, t);
	if (r.length && n.push(e.jsxPlugin({ jsxComponentDescriptors: r })), L(t.frontmatter) && n.push(e.frontmatterPlugin()), L(t.admonitions)) {
		let r = R(t.admonitions);
		n.push(e.directivesPlugin({
			directiveDescriptors: [e.AdmonitionDirectiveDescriptor],
			escapeUnknownTextDirectives: r?.escapeUnknownTextDirectives ?? !0
		}));
	}
	if (L(t.images)) {
		let r = R(t.images);
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
		let r = z(t.diffMode);
		n.push(e.diffSourcePlugin({
			...r?.viewMode ? { viewMode: r.viewMode } : {},
			...r?.diffMarkdown === void 0 ? {} : { diffMarkdown: r.diffMarkdown },
			...r?.readOnlyDiff === void 0 ? {} : { readOnlyDiff: r.readOnlyDiff }
		}));
	}
	return L(t.markdownShortcuts) && n.push(e.markdownShortcutPlugin()), t.plugins?.length && n.push(...t.plugins), n;
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorField.tsx
var W = {
	xs: "min-h-32",
	sm: "min-h-36",
	md: "min-h-40",
	lg: "min-h-48",
	xl: "min-h-56"
}, G = null;
function K() {
	return G ??= import("./dist-C9JOkTA6.js"), G;
}
function q({ id: e, value: t = "", onChange: n, readOnly: r = !1, disabled: o = !1, size: s = "md", placeholder: c, className: f, contentClassName: p, textareaClassName: m, "aria-label": h, commitInitialMarkdownNormalize: g = !1, clearable: _, onClear: y, clearLabel: b, ...x }) {
	let [S, C] = (0, d.useState)(null), [w, T] = (0, d.useState)(null), E = (0, d.useRef)(null), D = (0, d.useRef)(t), O = r || o;
	(0, d.useEffect)(() => {
		let e = !1;
		return K().then((t) => {
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
	}, [g, n]), A = (0, d.useMemo)(() => S ? U(S, x) : [], [S, x]), j = (0, d.useCallback)((e) => {
		queueMicrotask(() => T(e));
	}, []), M = w;
	if (!S || M) {
		let n = h ?? (M ? "Markdown source" : e ? void 0 : "Markdown");
		return /* @__PURE__ */ (0, v.jsxs)(i, {
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
					className: l(a(s), "h-auto resize-y font-mono", W[s], m),
					value: t,
					disabled: O,
					placeholder: typeof c == "string" ? c : void 0,
					onChange: (e) => k(e.target.value)
				}),
				M && !O && /* @__PURE__ */ (0, v.jsx)(u, {
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
	return /* @__PURE__ */ (0, v.jsx)(i, {
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
				className: l("clicky-mdx-editor-field", O && "clicky-mdx-editor-field-readonly", f),
				contentEditableClassName: l("clicky-mdx-editor-content", W[s], p),
				onChange: k,
				onError: j
			})
		})
	});
}
export { q as MdxEditorField, g as a, f as c, m as i, p as l, A as n, h as o, k as r, _ as s, T as t };
