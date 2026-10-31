import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./jsx-runtime-lVm_Myqd.js";
import { T as r, _ as i, b as a, n as o, r as s, t as c } from "./Callout-DmCAGTzW.js";
import { n as l } from "./PromptRunView-CKU3_Eci.js";
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-jsx-attributes.ts
var u = /* @__PURE__ */ e(t(), 1);
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
	variant: s,
	icon: o
};
function g(e) {
	return d(e, p);
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorCallout.tsx
var _ = n(), v = "h-6 rounded border border-border bg-background px-1 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-ring", y = r(v, "w-24 placeholder:text-placeholder"), b = r(v, "w-56 placeholder:text-placeholder");
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
		}), /* @__PURE__ */ (0, _.jsx)(c, {
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
	let n = e.ButtonWithTooltip, r = e.usePublisher(e.insertJsx$);
	return /* @__PURE__ */ (0, _.jsx)(n, {
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
		children: /* @__PURE__ */ (0, _.jsx)(a, { size: "1em" })
	});
}
var w = /* @__PURE__ */ new WeakMap(), T = /* @__PURE__ */ new WeakMap();
function E(e) {
	let t = w.get(e);
	if (t) return t;
	let n = ({ mdastNode: t }) => (0, u.createElement)(S, {
		mdastNode: t,
		mdx: e
	});
	return w.set(e, n), n;
}
function D(e, t = {}) {
	let n = t.name ?? "CalloutBox", r = T.get(e) ?? /* @__PURE__ */ new Map();
	T.set(e, r);
	let i = r.get(n);
	if (i) return i;
	let a = () => (0, u.createElement)(C, {
		mdx: e,
		name: n
	});
	return r.set(n, a), a;
}
function O(e, t = {}) {
	return {
		name: t.name ?? "CalloutBox",
		kind: "flow",
		...t.source ? { source: t.source } : {},
		props: m,
		hasChildren: !0,
		Editor: E(e)
	};
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorToolbar.tsx
function k(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function A(e) {
	return typeof e == "object" && e ? e : void 0;
}
function j(e) {
	if (e) return typeof e == "object" && e.viewModes?.length ? [...e.viewModes] : [
		"rich-text",
		"source",
		"diff"
	];
}
function M({ mdx: e, options: t }) {
	let { ConditionalContents: n, ChangeCodeMirrorLanguage: r } = e, i = [...k(t.codeBlocks) && k(t.codeMirror) ? [{
		when: (e) => e?.editorType === "codeblock",
		contents: () => /* @__PURE__ */ (0, _.jsx)(r, {})
	}] : [], { fallback: () => /* @__PURE__ */ (0, _.jsx)(N, {
		mdx: e,
		options: t
	}) }];
	return /* @__PURE__ */ (0, _.jsx)(n, { options: i });
}
function N({ mdx: e, options: t }) {
	let { BlockTypeSelect: n, BoldItalicUnderlineToggles: r, CodeToggle: i, CreateLink: a, DiffSourceToggleWrapper: o, InsertAdmonition: s, InsertCodeBlock: c, InsertFrontmatter: l, InsertImage: u, InsertTable: d, InsertThematicBreak: f, ListsToggle: p, Separator: m, UndoRedo: h } = e, g = D(e, A(t.callouts) ?? {}), v = typeof t.toolbar == "object" ? t.toolbar.extra : void 0, y = /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [
		/* @__PURE__ */ (0, _.jsx)(h, {}),
		/* @__PURE__ */ (0, _.jsx)(m, {}),
		/* @__PURE__ */ (0, _.jsx)(r, {}),
		/* @__PURE__ */ (0, _.jsx)(i, {}),
		k(t.headings) && /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [/* @__PURE__ */ (0, _.jsx)(m, {}), /* @__PURE__ */ (0, _.jsx)(n, {})] }),
		k(t.lists) && /* @__PURE__ */ (0, _.jsx)(p, {}),
		k(t.links) && /* @__PURE__ */ (0, _.jsxs)(_.Fragment, { children: [/* @__PURE__ */ (0, _.jsx)(m, {}), /* @__PURE__ */ (0, _.jsx)(a, {})] }),
		k(t.images) && /* @__PURE__ */ (0, _.jsx)(u, {}),
		(k(t.tables) || k(t.thematicBreak) || k(t.codeBlocks)) && /* @__PURE__ */ (0, _.jsx)(m, {}),
		k(t.tables) && /* @__PURE__ */ (0, _.jsx)(d, {}),
		k(t.thematicBreak) && /* @__PURE__ */ (0, _.jsx)(f, {}),
		k(t.codeBlocks) && /* @__PURE__ */ (0, _.jsx)(c, {}),
		(k(t.frontmatter) || k(t.admonitions) || k(t.callouts, !1)) && /* @__PURE__ */ (0, _.jsx)(m, {}),
		k(t.frontmatter) && /* @__PURE__ */ (0, _.jsx)(l, {}),
		k(t.admonitions) && /* @__PURE__ */ (0, _.jsx)(s, {}),
		k(t.callouts, !1) && /* @__PURE__ */ (0, _.jsx)(g, {}),
		v
	] }), b = j(t.diffMode);
	return b ? /* @__PURE__ */ (0, _.jsx)(o, {
		options: b,
		children: y
	}) : y;
}
// ../../../clicky-ui/packages/ui/src/components/mdx-editor-field-plugins.tsx
var P = {
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
function F(e, t = !0) {
	return e === void 0 ? t : e !== !1;
}
function I(e) {
	return typeof e == "object" && e ? e : void 0;
}
function L(e) {
	if (e) return e === !0 ? {} : typeof e == "string" ? { viewMode: e } : e;
}
function R(e, t) {
	return () => (0, u.createElement)(M, {
		mdx: e,
		options: t
	});
}
function z(e, t) {
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
function B(e, t) {
	let n = [];
	return F(t.callouts, !1) && n.push(O(e, I(t.callouts) ?? {})), t.jsxComponents?.length && n.push(...t.jsxComponents), n;
}
function V(e, t = {}) {
	let n = [];
	if (F(t.toolbar)) {
		let r = I(t.toolbar);
		n.push(e.toolbarPlugin({
			toolbarContents: R(e, t),
			...r?.className ? { toolbarClassName: r.className } : {},
			...r?.position ? { toolbarPosition: r.position } : {}
		}));
	}
	if (F(t.headings)) {
		let r = I(t.headings);
		n.push(e.headingsPlugin(r));
	}
	if (F(t.lists) && n.push(e.listsPlugin()), F(t.quote) && n.push(e.quotePlugin()), F(t.links)) {
		let r = I(t.links);
		n.push(e.linkPlugin(r?.disableAutoLink === void 0 ? void 0 : { disableAutoLink: r.disableAutoLink }), e.linkDialogPlugin({
			...r?.linkAutocompleteSuggestions ? { linkAutocompleteSuggestions: r.linkAutocompleteSuggestions } : {},
			...r?.showLinkTitleField === void 0 ? {} : { showLinkTitleField: r.showLinkTitleField }
		}));
	}
	if (F(t.tables) && n.push(e.tablePlugin()), F(t.thematicBreak) && n.push(e.thematicBreakPlugin()), F(t.codeBlocks)) {
		let r = I(t.codeBlocks), i = { defaultCodeBlockLanguage: r?.defaultLanguage ?? "txt" }, a = z(e, r?.editorDescriptors);
		if (a && (i.codeBlockEditorDescriptors = a), n.push(e.codeBlockPlugin(i)), F(t.codeMirror)) {
			let r = I(t.codeMirror);
			n.push(e.codeMirrorPlugin({
				codeBlockLanguages: r?.languages ?? P,
				...r?.autoLoadLanguageSupport === void 0 ? {} : { autoLoadLanguageSupport: r.autoLoadLanguageSupport }
			}));
		}
	}
	let r = B(e, t);
	if (r.length && n.push(e.jsxPlugin({ jsxComponentDescriptors: r })), F(t.frontmatter) && n.push(e.frontmatterPlugin()), F(t.admonitions)) {
		let r = I(t.admonitions);
		n.push(e.directivesPlugin({
			directiveDescriptors: [e.AdmonitionDirectiveDescriptor],
			escapeUnknownTextDirectives: r?.escapeUnknownTextDirectives ?? !0
		}));
	}
	if (F(t.images)) {
		let r = I(t.images);
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
		let r = L(t.diffMode);
		n.push(e.diffSourcePlugin({
			...r?.viewMode ? { viewMode: r.viewMode } : {},
			...r?.diffMarkdown === void 0 ? {} : { diffMarkdown: r.diffMarkdown },
			...r?.readOnlyDiff === void 0 ? {} : { readOnlyDiff: r.readOnlyDiff }
		}));
	}
	return F(t.markdownShortcuts) && n.push(e.markdownShortcutPlugin()), t.plugins?.length && n.push(...t.plugins), n;
}
// ../../../clicky-ui/packages/ui/src/components/MdxEditorField.tsx
var H = {
	xs: "min-h-32",
	sm: "min-h-36",
	md: "min-h-40",
	lg: "min-h-48",
	xl: "min-h-56"
}, U = null;
function W() {
	return U ??= import("./dist-CT27Kp6A.js"), U;
}
function G({ id: e, value: t = "", onChange: n, readOnly: a = !1, disabled: o = !1, size: s = "md", placeholder: c, className: d, contentClassName: f, textareaClassName: p, "aria-label": m, commitInitialMarkdownNormalize: h = !1, ...g }) {
	let [v, y] = (0, u.useState)(null), [b, x] = (0, u.useState)(null), S = (0, u.useRef)(null), C = (0, u.useRef)(t), w = a || o;
	(0, u.useEffect)(() => {
		let e = !1;
		return W().then((t) => {
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
	}, [h, n]), E = (0, u.useMemo)(() => v ? V(v, g) : [], [v, g]), D = (0, u.useCallback)((e) => {
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
					className: r(l(s), "h-auto resize-y font-mono", H[s], p),
					value: t,
					disabled: w,
					placeholder: typeof c == "string" ? c : void 0,
					onChange: (e) => T(e.target.value)
				}),
				O && !w && /* @__PURE__ */ (0, _.jsx)(i, {
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
			placeholder: c,
			plugins: E,
			className: r("clicky-mdx-editor-field", w && "clicky-mdx-editor-field-readonly", d),
			contentEditableClassName: r("clicky-mdx-editor-content", H[s], f),
			onChange: T,
			onError: D
		})
	});
}
export { G as MdxEditorField };
