import { a as e, r as t } from "./rolldown-runtime-DUcC5eAW.js";
import { n, t as r } from "./jsx-runtime-B4NxTvNR.js";
import { t as i } from "./react-dom-BBpWcFtU.js";
import { a, t as o } from "./button-BLoR2pgR.js";
import { c as s, i as c, l, n as u, r as d, t as f } from "./UiInfo-CCiCR0Fr.js";
import { $ as p, B as m, C as h, G as g, I as _, L as v, M as y, N as b, R as x, V as S, W as C, X as w, Y as T, at as E, et as ee, f as te, it as ne, j as D, r as O, rt as re, s as k, t as ie, u as ae } from "./SplitButton-w2X8sfgS.js";
import { t as oe } from "./use-theme-sEqHIo2i.js";
import { a as se, d as A, l as ce, m as le, n as ue, o as de, p as fe, r as pe, s as me, t as he } from "./ErrorDetails-CbwPqfQb.js";
import { L as ge, R as _e, W as ve, h as ye, t as be } from "./Markdown-s_A3Eic8.js";
import { L as xe, Q as Se, R as Ce, Z as we, ct as Te, m as Ee, z as De } from "./locations-CUxJY9f0.js";
import { c as Oe, g as ke, i as Ae, n as je, o as Me, p as Ne, r as Pe, s as Fe, t as Ie } from "./ContextMeter.model-Vr4E2Guh.js";
import { a as Le, l as Re, n as ze, o as Be } from "./floating-ui.react-dom-CJoIjj1G.js";
import { a as Ve, i as j, o as He, t as Ue, v as M, y as We } from "./adapters-BoPmjJjW.js";
import { n as Ge, t as Ke } from "./ProblemActions-MzUA9awO.js";
import { t as qe } from "./Tabs-rMrrxeic.js";
// ../../../clicky-ui/packages/ui/src/lib/palette.ts
function Je(e) {
	let t = 2166136261;
	for (let n = 0; n < e.length; n++) t ^= e.charCodeAt(n), t = Math.imul(t, 16777619);
	return t >>> 0;
}
// ../../../clicky-ui/packages/ui/src/data/Avatar.tsx
var N = /* @__PURE__ */ e(n(), 1), P = r();
function Ye(e) {
	return Je(e) % 360;
}
function Xe(e, t, n) {
	if (t?.trim()) return t.trim().slice(0, n).toUpperCase();
	let r = e.replace(/^@/, "").trim(), i = (r.split(/[\\/]/).filter(Boolean).at(-1) ?? r).replace(/[^a-zA-Z0-9]+/g, " ").trim().split(/\s+/).filter(Boolean).map((e) => e.match(/[a-zA-Z0-9]/)?.[0] ?? "").filter(Boolean).slice(0, n).join("");
	return i ? i.toUpperCase() : (r.replace(/[^a-zA-Z0-9]+/g, "").slice(0, n) || "?").toUpperCase();
}
function Ze(e, t, n) {
	if (t === "group") switch (e) {
		case "solid": return {
			background: "#d8d4cc",
			borderColor: "#d8d4cc",
			color: "#5a574e"
		};
		case "stamp": return {
			background: "#fafaf7",
			borderColor: "#8d8778",
			color: "#5a574e",
			shadow: "2px 2px 0 rgba(90, 87, 78, 0.08)"
		};
		case "mono": return {
			background: "#f5f2eb",
			borderColor: "#8d8778",
			color: "#5a574e"
		};
		default: return {
			background: "#f0ede6",
			borderColor: "#b8b3a7",
			borderStyle: "dashed",
			color: "#5a574e"
		};
	}
	switch (e) {
		case "solid": return {
			background: `oklch(0.55 0.12 ${n})`,
			borderColor: `oklch(0.55 0.12 ${n})`,
			color: "#fff"
		};
		case "stamp": return {
			background: "#fafaf7",
			borderColor: `oklch(0.42 0.15 ${n})`,
			color: `oklch(0.42 0.15 ${n})`,
			shadow: "2px 2px 0 rgba(26, 26, 26, 0.08)"
		};
		case "mono": return {
			background: "#fff",
			borderColor: "#1a1a1a",
			color: "#1a1a1a"
		};
		default: return {
			background: `oklch(0.93 0.04 ${n})`,
			borderColor: `oklch(0.55 0.14 ${n} / 0.25)`,
			color: `oklch(0.35 0.14 ${n})`
		};
	}
}
function Qe(e, t) {
	let n = t <= 16 ? 7 : t <= 20 ? 8 : 9;
	switch (e) {
		case "stamp": return Math.max(n, Math.round(t * .34));
		case "mono": return Math.max(n, Math.round(t * .36));
		default: return Math.max(n, Math.round(t * .38));
	}
}
function $e({ src: e, alt: t, initials: n, size: r = "sm", rounded: i, kind: o = "user", variant: c = "duotone", title: u, href: d, colorKey: f, onError: p, className: m }) {
	let [h, g] = (0, N.useState)(), _ = s(), v = l(r, _), y = (i ?? (o === "group" ? "md" : "full")) === "full" ? "rounded-full" : "rounded-md", b = Ze(c, o, Ye(f ?? t)), x = Xe(t, n, r === "xs" || r === "sm" ? 1 : 2), S = a("inline-flex shrink-0 select-none items-center justify-center overflow-hidden border", y), C = {
		width: v,
		height: v,
		borderColor: b.borderColor,
		borderStyle: b.borderStyle ?? "solid"
	};
	c === "stamp" && (C.transform = "rotate(-4deg)", C.boxShadow = b.shadow);
	let w = (e) => e === "stamp" ? 700 : r === "sm" ? 500 : 600, T = e && e !== h ? /* @__PURE__ */ (0, P.jsx)("img", {
		src: e,
		alt: t,
		title: u ?? t,
		width: v,
		height: v,
		"data-avatar-kind": o,
		"data-avatar-variant": c,
		"data-avatar-size": r,
		className: a(S, "bg-muted object-cover", m),
		style: C,
		loading: "lazy",
		onError: (t) => {
			g(e), p?.(t);
		}
	}) : /* @__PURE__ */ (0, P.jsx)("span", {
		role: "img",
		"aria-label": t,
		"data-avatar-kind": o,
		"data-avatar-variant": c,
		"data-avatar-size": r,
		className: a(S, c === "mono" || c === "stamp" ? "font-mono" : void 0, c === "mono" ? "tracking-[-0.05em]" : "tracking-[-0.02em]", m),
		style: {
			...C,
			background: b.background,
			color: b.color,
			fontSize: Qe(c, v),
			fontWeight: w(c)
		},
		title: u ?? t,
		children: x
	});
	return d ? /* @__PURE__ */ (0, P.jsx)("a", {
		href: d,
		target: "_blank",
		rel: "noopener noreferrer",
		className: "inline-flex shrink-0",
		onClick: (e) => e.stopPropagation(),
		children: T
	}) : T;
}
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "m230.88 115.69l-168-95.88a14 14 0 0 0-20 16.87L73.66 128l-30.85 91.33A14 14 0 0 0 56 238a14.15 14.15 0 0 0 6.93-1.83l167.91-96.07a14 14 0 0 0 0-24.41Zm-5.95 14L57 225.73a2 2 0 0 1-2.86-2.42a.4.4 0 0 0 0-.1L84.3 134H144a6 6 0 0 0 0-12H84.3L54.17 32.8a.3.3 0 0 0 0-.1a1.87 1.87 0 0 1 .6-2.2a1.85 1.85 0 0 1 2.23-.25l168 95.89a1.93 1.93 0 0 1 1 1.74a2 2 0 0 1-1.07 1.78Z"
	})
}), {
	__source: "ph:paper-plane-right-light",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "paper-plane-right",
	displayName: "UiPaperPlaneRight"
});
var et = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M240 127.89a16 16 0 0 1-8.18 14L63.9 237.9A16.15 16.15 0 0 1 56 240a16 16 0 0 1-15-21.33l27-79.95a4 4 0 0 1 3.72-2.72H144a8 8 0 0 0 8-8.53a8.19 8.19 0 0 0-8.26-7.47h-72a4 4 0 0 1-3.79-2.72l-27-79.94a16 16 0 0 1 22.89-19.27l168 95.89a16 16 0 0 1 8.16 13.93"
	})
}), {
	__source: "ph:paper-plane-right-fill",
	__viewBox: "0 0 256 256",
	__group: "ui-controls",
	__consumerName: "paper-plane-right",
	displayName: "UiPaperPlaneRightFilled"
}), tt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "m148.2 84.28l-89.6 88a6 6 0 0 1-8.4 0l-38.4-37.71a6 6 0 1 1 8.4-8.57l34.2 33.58l85.4-83.87a6 6 0 1 1 8.4 8.56Zm96.08-8.48a6 6 0 0 0-8.48-.08l-85.4 83.87l-20.23-19.87a6 6 0 1 0-8.41 8.56l24.44 24a6 6 0 0 0 8.4 0l89.6-88a6 6 0 0 0 .08-8.48"
	})
}), {
	__source: "ph:checks-light",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "checks",
	displayName: "UiChecks"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "m152.41 88.56l-89.6 88a12 12 0 0 1-16.82 0l-38.4-37.71a12 12 0 0 1 16.82-17.13l30 29.46l81.19-79.74a12 12 0 0 1 16.82 17.12Zm96.15-17a12 12 0 0 0-17-.15l-81.16 79.77l-7.88-7.74a12 12 0 0 0-16.82 17.12l16.29 16a12 12 0 0 0 16.82 0l89.6-88a12 12 0 0 0 .15-16.97Z"
	})
}), {
	__source: "ph:checks-bold",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "checks",
	displayName: "UiChecksFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiThumbsUp.tsx
var nt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M232.49 81.44A22 22 0 0 0 216 74h-58V56a38 38 0 0 0-38-38a6 6 0 0 0-5.37 3.32L76.29 98H32a14 14 0 0 0-14 14v88a14 14 0 0 0 14 14h172a22 22 0 0 0 21.83-19.27l12-96a22 22 0 0 0-5.34-17.29M30 200v-88a2 2 0 0 1 2-2h42v92H32a2 2 0 0 1-2-2M225.92 97.24l-12 96A10 10 0 0 1 204 202H86v-96.58l37.58-75.17A26 26 0 0 1 146 56v24a6 6 0 0 0 6 6h64a10 10 0 0 1 9.92 11.24"
	})
}), {
	__source: "ph:thumbs-up-light",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "thumbs-up",
	displayName: "UiThumbsUp"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M234 80.12A24 24 0 0 0 216 72h-56V56a40 40 0 0 0-40-40a8 8 0 0 0-7.16 4.42L75.06 96H32a16 16 0 0 0-16 16v88a16 16 0 0 0 16 16h172a24 24 0 0 0 23.82-21l12-96A24 24 0 0 0 234 80.12M32 112h40v88H32Z"
	})
}), {
	__source: "ph:thumbs-up-fill",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "thumbs-up",
	displayName: "UiThumbsUpFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiThumbsDown.tsx
var rt = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "m237.83 157.27l-12-96A22 22 0 0 0 204 42H32a14 14 0 0 0-14 14v88a14 14 0 0 0 14 14h44.29l38.34 76.68A6 6 0 0 0 120 238a38 38 0 0 0 38-38v-18h58a22 22 0 0 0 21.83-24.73M74 146H32a2 2 0 0 1-2-2V56a2 2 0 0 1 2-2h42Zm149.5 20.62A9.9 9.9 0 0 1 216 170h-64a6 6 0 0 0-6 6v24a26 26 0 0 1-22.42 25.75L86 150.58V54h118a10 10 0 0 1 9.92 8.76l12 96a9.9 9.9 0 0 1-2.42 7.86"
	})
}), {
	__source: "ph:thumbs-down-light",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "thumbs-down",
	displayName: "UiThumbsDown"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "m239.82 157l-12-96A24 24 0 0 0 204 40H32a16 16 0 0 0-16 16v88a16 16 0 0 0 16 16h43.06l37.78 75.58A8 8 0 0 0 120 240a40 40 0 0 0 40-40v-16h56a24 24 0 0 0 23.82-27M72 144H32V56h40Z"
	})
}), {
	__source: "ph:thumbs-down-fill",
	__viewBox: "0 0 256 256",
	__group: "approval-review",
	__consumerName: "thumbs-down",
	displayName: "UiThumbsDownFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiGitCommit.tsx
var it = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M248 122h-66.34a54 54 0 0 0-107.32 0H8a6 6 0 0 0 0 12h66.34a54 54 0 0 0 107.32 0H248a6 6 0 0 0 0-12m-120 48a42 42 0 1 1 42-42a42 42 0 0 1-42 42"
	})
}), {
	__source: "ph:git-commit-light",
	__viewBox: "0 0 256 256",
	__group: "git-source-control",
	__consumerName: "git-commit",
	displayName: "UiGitCommit"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M256 128a8 8 0 0 1-8 8h-64.58a56 56 0 0 1-110.84 0H8a8 8 0 0 1 0-16h64.58a56 56 0 0 1 110.84 0H248a8 8 0 0 1 8 8"
	})
}), {
	__source: "ph:git-commit-fill",
	__viewBox: "0 0 256 256",
	__group: "git-source-control",
	__consumerName: "git-commit",
	displayName: "UiGitCommitFilled"
});
// ../../../clicky-ui/packages/ui/src/icons/components/UiUser.tsx
var at = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	})
}), {
	__source: "ph:user-light",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUser"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	})
}), {
	__source: "ph:user-fill",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserFilled"
});
var ot = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M12 4v16M4 12h16",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-light + plus marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserPlus"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M12 4v16M4 12h16",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-fill + plus marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserPlusFilled"
});
var st = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#dc2626",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M4 12h16",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-light + minus marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserMinus"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#dc2626",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M4 12h16",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-fill + minus marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserMinusFilled"
});
var ct = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M5 13l4 4 10-10",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			strokeLinejoin: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-light + tick marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserCheck"
});
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, P.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, P.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, P.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, P.jsx)("path", {
			d: "M5 13l4 4 10-10",
			stroke: "currentColor",
			strokeWidth: "3",
			strokeLinecap: "round",
			strokeLinejoin: "round",
			fill: "none"
		})]
	})]
}), {
	__source: "ph:user-fill + tick marker",
	__viewBox: "0 0 256 256",
	__group: "people-orgs",
	__consumerName: "user",
	displayName: "UiUserCheckFilled"
});
// ../../../clicky-ui/packages/ui/src/data/ImagePreview.tsx
var lt = {
	backgroundImage: "conic-gradient(color-mix(in srgb, var(--muted-foreground) 14%, transparent) 25%, transparent 0 50%, color-mix(in srgb, var(--muted-foreground) 14%, transparent) 0 75%, transparent 0)",
	backgroundSize: "16px 16px"
};
function ut({ src: e, alt: t, className: n }) {
	let [r, i] = (0, N.useState)(!1), [o, s] = (0, N.useState)(!1), c = t || e;
	return (0, N.useEffect)(() => {
		i(!1), s(!1);
	}, [e]), r ? /* @__PURE__ */ (0, P.jsxs)("div", {
		role: "alert",
		className: a("m-density-4 rounded-md border border-destructive/40 p-3 text-sm text-destructive", n),
		children: ["Could not load ", c]
	}) : /* @__PURE__ */ (0, P.jsx)("div", {
		"data-image-preview": "",
		className: a("flex size-full min-h-0 items-center justify-center overflow-auto bg-muted/30 p-density-4", n),
		style: lt,
		children: /* @__PURE__ */ (0, P.jsx)("img", {
			src: e,
			alt: c,
			"data-size": o ? "actual" : "fit",
			onClick: () => s((e) => !e),
			onError: () => i(!0),
			className: a("rounded-sm bg-background shadow-sm", o ? "max-w-none cursor-zoom-out" : "max-h-full max-w-full cursor-zoom-in object-contain")
		})
	});
}
// ../../../clicky-ui/packages/ui/src/data/git/DiffStat.tsx
var dt = 5;
function ft(e, t) {
	return `${e} ${t}${e === 1 ? "" : "s"}`;
}
function pt({ additions: e, deletions: t, files: n, bars: r = !0, className: i }) {
	let o = e + t, s = o > 0 ? Math.round(e / o * dt) : 0, c = [
		n === void 0 ? void 0 : ft(n, "changed file"),
		ft(e, "addition"),
		ft(t, "deletion")
	].filter(Boolean).join(", ");
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		role: "img",
		"aria-label": c,
		title: c,
		className: a("inline-flex items-center gap-1 whitespace-nowrap tabular-nums", i),
		children: [
			n !== void 0 && /* @__PURE__ */ (0, P.jsx)("span", {
				"aria-hidden": !0,
				className: "text-muted-foreground",
				children: n
			}),
			/* @__PURE__ */ (0, P.jsxs)("span", {
				"aria-hidden": !0,
				className: "text-emerald-600 dark:text-emerald-400",
				children: ["+", e]
			}),
			/* @__PURE__ */ (0, P.jsxs)("span", {
				"aria-hidden": !0,
				className: "text-rose-600 dark:text-rose-400",
				children: ["-", t]
			}),
			r && /* @__PURE__ */ (0, P.jsx)("span", {
				className: "inline-flex gap-px",
				"aria-hidden": !0,
				children: Array.from({ length: o === 0 ? 1 : dt }, (e, t) => {
					let n = o === 0 ? "empty" : t < s ? "add" : "del";
					return /* @__PURE__ */ (0, P.jsx)("span", {
						"data-diff-bar": n,
						className: a("inline-block h-2 w-1.5 rounded-sm", n === "add" && "bg-emerald-500", n === "del" && "bg-rose-500", n === "empty" && "bg-muted-foreground/30")
					}, t);
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/data/git/GitChangeList.tsx
function mt({ loading: e, payload: t, error: n, emptyLabel: r = "No text diff available", className: i, maxHeightClassName: o = "max-h-[520px]", onLineAction: s, lineWidgets: c, lineSelection: l }) {
	if (e) return /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("border-t border-border bg-muted/40 px-3 py-3 text-xs text-muted-foreground", i),
		children: "Loading diff..."
	});
	if (n) return /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("border-t border-border bg-rose-50 px-3 py-3 text-xs text-rose-700 dark:bg-rose-950/30 dark:text-rose-300", i),
		children: n
	});
	if (t?.binary) return /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("border-t border-border bg-muted/40 px-3 py-3 text-xs text-muted-foreground", i),
		children: "Binary file not shown."
	});
	if (!t?.diff) return /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("border-t border-border bg-muted/40 px-3 py-3 text-xs text-muted-foreground", i),
		children: r
	});
	let u = S(t.path);
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: a("border-t border-border", i),
		children: [t.truncated && /* @__PURE__ */ (0, P.jsx)("div", {
			className: "border-b border-border bg-amber-100/70 px-3 py-1.5 text-xs text-amber-800 [[data-theme=dark]_&]:bg-amber-950/40 [[data-theme=dark]_&]:text-amber-200",
			children: "Diff truncated."
		}), /* @__PURE__ */ (0, P.jsx)(_, {
			bare: !0,
			unified: t.diff,
			className: a("overflow-auto px-1 py-2", o),
			...u ? { language: u } : {},
			...s ? { onLineAction: s } : {},
			...c ? { lineWidgets: c } : {},
			...l ? { lineSelection: l } : {}
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/data/diagnostics/ActionError.tsx
function F({ error: e, title: t = "Error", context: n = [], actions: r, className: i }) {
	let a = ue(e) ?? {
		message: String(e),
		context: []
	};
	return /* @__PURE__ */ (0, P.jsx)("div", {
		role: "alert",
		className: i,
		children: /* @__PURE__ */ (0, P.jsx)(he, {
			title: t,
			diagnostics: {
				...a,
				context: [...n, ...a.context]
			},
			...r ? { actions: r } : {}
		})
	});
}
// ../../../clicky-ui/packages/ui/src/layout/ResizeHandle.tsx
function ht({ orientation: e, value: t, min: n, max: r, pointerFactor: i, label: o, testId: s, className: c, onChange: l }) {
	let u = (0, N.useRef)(null);
	(0, N.useEffect)(() => () => u.current?.(), []);
	let d = (e) => Math.max(n, Math.min(r, e));
	return /* @__PURE__ */ (0, P.jsx)("div", {
		role: "separator",
		tabIndex: 0,
		"aria-label": o,
		"aria-orientation": e,
		"aria-valuemin": n,
		"aria-valuemax": r,
		"aria-valuenow": Math.round(t),
		"data-testid": s,
		className: a("absolute z-20 bg-border transition-colors hover:bg-primary focus-visible:bg-primary focus-visible:outline-none touch-none", e === "vertical" ? "top-0 h-full w-1 cursor-col-resize" : "left-0 h-1 w-full cursor-row-resize", c),
		onPointerDown: (n) => {
			if (n.button !== 0) return;
			n.preventDefault(), u.current?.();
			let r = e === "vertical" ? n.clientX : n.clientY, a = document.body.style.cursor, o = document.body.style.userSelect, s = (n) => l(d(t + ((e === "vertical" ? n.clientX : n.clientY) - r) * i)), c = () => {
				document.removeEventListener("pointermove", s), document.removeEventListener("pointerup", c), document.removeEventListener("pointercancel", c), document.body.style.cursor = a, document.body.style.userSelect = o, u.current = null;
			};
			u.current = c, n.currentTarget.setPointerCapture?.(n.pointerId), document.addEventListener("pointermove", s), document.addEventListener("pointerup", c), document.addEventListener("pointercancel", c), document.body.style.cursor = e === "vertical" ? "col-resize" : "row-resize", document.body.style.userSelect = "none";
		},
		onKeyDown: (a) => {
			let o = a.shiftKey ? 48 : 16;
			if (a.key === "Home") l(n);
			else if (a.key === "End") l(r);
			else if (a.key === (e === "vertical" ? "ArrowLeft" : "ArrowUp")) l(d(t - o * i));
			else if (a.key === (e === "vertical" ? "ArrowRight" : "ArrowDown")) l(d(t + o * i));
			else return;
			a.preventDefault();
		}
	});
}
// ../../../clicky-ui/packages/ui/src/comments/comment-persistence.ts
var gt = (0, N.createContext)(null);
function _t(e) {
	let t = `clicky.comment.${encodeURIComponent(e)}.`, n = /* @__PURE__ */ new Set(), r = /* @__PURE__ */ new Map(), i = null, a = 0, o = () => {
		a++, n.forEach((e) => e()), window.dispatchEvent(new Event("playground-comment-state"));
	};
	return {
		discardDraft(e) {
			let t = e.replace(/:body$/, "");
			this.write(`${t}:body`, ""), this.write(`${t}:facets`, {}), this.write(`${t}:target`, void 0), this.write(`${t}:mentions`, []), this.write(`${t}:open`, !1);
		},
		get version() {
			return a;
		},
		drafts() {
			let e = new Set(r.keys());
			try {
				for (let n = 0; n < sessionStorage.length; n++) {
					let r = sessionStorage.key(n);
					r?.startsWith(t) && e.add(r.slice(t.length));
				}
			} catch (e) {
				i = `Comment recovery unavailable: ${String(e)}`;
			}
			return [...e].filter((e) => e.startsWith("draft:") && e.endsWith(":body")).flatMap((e) => {
				let t = this.read(e, "");
				return typeof t == "string" && t.trim() ? [{
					key: e,
					body: t
				}] : [];
			});
		},
		get error() {
			return i;
		},
		subscribe(e) {
			return n.add(e), () => {
				n.delete(e);
			};
		},
		read(e, n) {
			if (!r.has(e)) try {
				let i = sessionStorage.getItem(t + e);
				r.set(e, i === null ? n : JSON.parse(i));
			} catch (t) {
				i = `Comment recovery unavailable: ${String(t)}`, r.set(e, n);
			}
			return r.get(e);
		},
		write(e, n) {
			r.set(e, n);
			try {
				for (let [e, n] of r) n === void 0 ? sessionStorage.removeItem(t + e) : sessionStorage.setItem(t + e, JSON.stringify(n));
				i = null;
			} catch (e) {
				i = `Comment recovery unavailable: ${String(e)}`;
			}
			o();
		}
	};
}
function vt() {
	let e = (0, N.useContext)(gt);
	return (t) => e?.discardDraft(`draft:create:${t}:body`);
}
function yt(e, t) {
	let n = (0, N.useContext)(gt), [r, i] = (0, N.useState)(t);
	return [(0, N.useSyncExternalStore)(n?.subscribe ?? (() => () => {}), () => n ? n.read(e, r) : r), (0, N.useCallback)((r) => {
		if (!n) {
			i(r);
			return;
		}
		let a = typeof r == "function" ? r(n.read(e, t)) : r;
		n.write(e, a);
	}, [
		n,
		e,
		t
	])];
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@medv/finder/4.0.2/ffbf122184a50ce99735645cc3be9745a4a3ab74f5aedeb1f96d3135154e2178/node_modules/@medv/finder/finder.js
var bt = /* @__PURE__ */ new Set([
	"role",
	"name",
	"aria-label",
	"rel",
	"href"
]);
function xt(e, t) {
	let n = bt.has(e);
	n ||= e.startsWith("data-") && Dt(e);
	let r = Dt(t) && t.length < 100;
	return r ||= t.startsWith("#") && Dt(t.slice(1)), n && r;
}
function St(e) {
	return Dt(e);
}
function Ct(e) {
	return Dt(e);
}
function wt(e) {
	return !0;
}
function Tt(e, t) {
	if (e.nodeType !== Node.ELEMENT_NODE) throw Error("Can't generate CSS selector for non-element node type.");
	if (e.tagName.toLowerCase() === "html") return "html";
	let n = {
		root: document.body,
		idName: St,
		className: Ct,
		tagName: wt,
		attr: xt,
		timeoutMs: 1e3,
		seedMinLength: 3,
		optimizedMinLength: 2,
		maxNumberOfPathChecks: Infinity
	}, r = /* @__PURE__ */ new Date(), i = {
		...n,
		...t
	}, a = It(i.root, n), o, s = 0;
	for (let t of Et(e, i, a)) {
		if ((/* @__PURE__ */ new Date()).getTime() - r.getTime() > i.timeoutMs || s >= i.maxNumberOfPathChecks) {
			let t = Nt(e, a);
			if (!t) throw Error(`Timeout: Can't find a unique selector after ${i.timeoutMs}ms`);
			return kt(t);
		}
		if (s++, Lt(t, a)) {
			o = t;
			break;
		}
	}
	if (!o) throw Error("Selector was not found.");
	let c = [...Rt(o, e, i, a, r)];
	return c.sort(jt), c.length > 0 ? kt(c[0]) : kt(o);
}
function* Et(e, t, n) {
	let r = [], i = [], a = e, o = 0;
	for (; a && a !== n;) {
		let e = Ot(a, t);
		for (let t of e) t.level = o;
		if (r.push(e), a = a.parentElement, o++, i.push(...Ft(r)), o >= t.seedMinLength) {
			i.sort(jt);
			for (let e of i) yield e;
			i = [];
		}
	}
	i.sort(jt);
	for (let e of i) yield e;
}
function Dt(e) {
	if (/^[a-z\-]{3,}$/i.test(e)) {
		let t = e.split(/-|[A-Z]/);
		for (let e of t) if (e.length <= 2 || /[^aeiou]{4,}/i.test(e)) return !1;
		return !0;
	}
	return !1;
}
function Ot(e, t) {
	let n = [], r = e.getAttribute("id");
	r && t.idName(r) && n.push({
		name: "#" + CSS.escape(r),
		penalty: 0
	});
	for (let r = 0; r < e.classList.length; r++) {
		let i = e.classList[r];
		t.className(i) && n.push({
			name: "." + CSS.escape(i),
			penalty: 1
		});
	}
	for (let r = 0; r < e.attributes.length; r++) {
		let i = e.attributes[r];
		t.attr(i.name, i.value) && n.push({
			name: `[${CSS.escape(i.name)}="${CSS.escape(i.value)}"]`,
			penalty: 2
		});
	}
	let i = e.tagName.toLowerCase();
	if (t.tagName(i)) {
		n.push({
			name: i,
			penalty: 5
		});
		let t = Mt(e, i);
		t !== void 0 && n.push({
			name: Pt(i, t),
			penalty: 10
		});
	}
	let a = Mt(e);
	return a !== void 0 && n.push({
		name: I(i, a),
		penalty: 50
	}), n;
}
function kt(e) {
	let t = e[0], n = t.name;
	for (let r = 1; r < e.length; r++) {
		let i = e[r].level || 0;
		n = t.level === i - 1 ? `${e[r].name} > ${n}` : `${e[r].name} ${n}`, t = e[r];
	}
	return n;
}
function At(e) {
	return e.map((e) => e.penalty).reduce((e, t) => e + t, 0);
}
function jt(e, t) {
	return At(e) - At(t);
}
function Mt(e, t) {
	let n = e.parentNode;
	if (!n) return;
	let r = n.firstChild;
	if (!r) return;
	let i = 0;
	for (; r && (r.nodeType === Node.ELEMENT_NODE && (t === void 0 || r.tagName.toLowerCase() === t) && i++, r !== e);) r = r.nextSibling;
	return i;
}
function Nt(e, t) {
	let n = 0, r = e, i = [];
	for (; r && r !== t;) {
		let e = r.tagName.toLowerCase(), t = Mt(r, e);
		if (t === void 0) return;
		i.push({
			name: Pt(e, t),
			penalty: NaN,
			level: n
		}), r = r.parentElement, n++;
	}
	if (Lt(i, t)) return i;
}
function I(e, t) {
	return e === "html" ? "html" : `${e}:nth-child(${t})`;
}
function Pt(e, t) {
	return e === "html" ? "html" : `${e}:nth-of-type(${t})`;
}
function* Ft(e, t = []) {
	if (e.length > 0) for (let n of e[0]) yield* Ft(e.slice(1, e.length), t.concat(n));
	else yield t;
}
function It(e, t) {
	return e.nodeType === Node.DOCUMENT_NODE ? e : e === t.root ? e.ownerDocument : e;
}
function Lt(e, t) {
	let n = kt(e);
	switch (t.querySelectorAll(n).length) {
		case 0: throw Error(`Can't select any node with this selector: ${n}`);
		case 1: return !0;
		default: return !1;
	}
}
function* Rt(e, t, n, r, i) {
	if (e.length > 2 && e.length > n.optimizedMinLength) for (let a = 1; a < e.length - 1; a++) {
		if ((/* @__PURE__ */ new Date()).getTime() - i.getTime() > n.timeoutMs) return;
		let o = [...e];
		o.splice(a, 1), Lt(o, r) && r.querySelector(kt(o)) === t && (yield o, yield* Rt(o, t, n, r, i));
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/polyfills/array.from.mjs
var zt = Object.prototype.toString;
function L(e) {
	return typeof e == "function" || zt.call(e) === "[object Function]";
}
function Bt(e) {
	var t = Number(e);
	return isNaN(t) ? 0 : t === 0 || !isFinite(t) ? t : (t > 0 ? 1 : -1) * Math.floor(Math.abs(t));
}
var Vt = 2 ** 53 - 1;
function Ht(e) {
	var t = Bt(e);
	return Math.min(Math.max(t, 0), Vt);
}
function Ut(e, t) {
	var n = Array, r = Object(e);
	if (e == null) throw TypeError("Array.from requires an array-like object - not null or undefined");
	if (t !== void 0 && !L(t)) throw TypeError("Array.from: when provided, the second argument must be a function");
	for (var i = Ht(r.length), a = L(n) ? Object(new n(i)) : Array(i), o = 0, s; o < i;) s = r[o], t ? a[o] = t(s, o) : a[o] = s, o += 1;
	return a.length = i, a;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/polyfills/SetLike.mjs
function Wt(e) {
	"@babel/helpers - typeof";
	return Wt = typeof Symbol == "function" && typeof Symbol.iterator == "symbol" ? function(e) {
		return typeof e;
	} : function(e) {
		return e && typeof Symbol == "function" && e.constructor === Symbol && e !== Symbol.prototype ? "symbol" : typeof e;
	}, Wt(e);
}
function Gt(e, t) {
	if (!(e instanceof t)) throw TypeError("Cannot call a class as a function");
}
function Kt(e, t) {
	for (var n = 0; n < t.length; n++) {
		var r = t[n];
		r.enumerable = r.enumerable || !1, r.configurable = !0, "value" in r && (r.writable = !0), Object.defineProperty(e, Yt(r.key), r);
	}
}
function qt(e, t, n) {
	return t && Kt(e.prototype, t), n && Kt(e, n), Object.defineProperty(e, "prototype", { writable: !1 }), e;
}
function Jt(e, t, n) {
	return t = Yt(t), t in e ? Object.defineProperty(e, t, {
		value: n,
		enumerable: !0,
		configurable: !0,
		writable: !0
	}) : e[t] = n, e;
}
function Yt(e) {
	var t = Xt(e, "string");
	return Wt(t) === "symbol" ? t : String(t);
}
function Xt(e, t) {
	if (Wt(e) !== "object" || e === null) return e;
	var n = e[Symbol.toPrimitive];
	if (n !== void 0) {
		var r = n.call(e, t || "default");
		if (Wt(r) !== "object") return r;
		throw TypeError("@@toPrimitive must return a primitive value.");
	}
	return (t === "string" ? String : Number)(e);
}
var Zt = typeof Set > "u" ? Set : /* @__PURE__ */ function() {
	function e() {
		var t = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : [];
		Gt(this, e), Jt(this, "items", void 0), this.items = t;
	}
	return qt(e, [
		{
			key: "add",
			value: function(e) {
				return this.has(e) === !1 && this.items.push(e), this;
			}
		},
		{
			key: "clear",
			value: function() {
				this.items = [];
			}
		},
		{
			key: "delete",
			value: function(e) {
				var t = this.items.length;
				return this.items = this.items.filter(function(t) {
					return t !== e;
				}), t !== this.items.length;
			}
		},
		{
			key: "forEach",
			value: function(e) {
				var t = this;
				this.items.forEach(function(n) {
					e(n, n, t);
				});
			}
		},
		{
			key: "has",
			value: function(e) {
				return this.items.indexOf(e) !== -1;
			}
		},
		{
			key: "size",
			get: function() {
				return this.items.length;
			}
		}
	]), e;
}();
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/getRole.mjs
function R(e) {
	return e.localName ?? e.tagName.toLowerCase();
}
var Qt = {
	article: "article",
	aside: "complementary",
	button: "button",
	datalist: "listbox",
	dd: "definition",
	details: "group",
	dialog: "dialog",
	dt: "term",
	fieldset: "group",
	figure: "figure",
	form: "form",
	footer: "contentinfo",
	h1: "heading",
	h2: "heading",
	h3: "heading",
	h4: "heading",
	h5: "heading",
	h6: "heading",
	header: "banner",
	hr: "separator",
	html: "document",
	legend: "legend",
	li: "listitem",
	math: "math",
	main: "main",
	menu: "list",
	nav: "navigation",
	ol: "list",
	optgroup: "group",
	option: "option",
	output: "status",
	progress: "progressbar",
	section: "region",
	summary: "button",
	table: "table",
	tbody: "rowgroup",
	textarea: "textbox",
	tfoot: "rowgroup",
	td: "cell",
	th: "columnheader",
	thead: "rowgroup",
	tr: "row",
	ul: "list"
}, $t = {
	caption: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	code: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	deletion: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	emphasis: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	generic: /* @__PURE__ */ new Set([
		"aria-label",
		"aria-labelledby",
		"aria-roledescription"
	]),
	insertion: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	none: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	paragraph: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	presentation: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	strong: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	subscript: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"]),
	superscript: /* @__PURE__ */ new Set(["aria-label", "aria-labelledby"])
};
function en(e, t) {
	return [
		"aria-atomic",
		"aria-busy",
		"aria-controls",
		"aria-current",
		"aria-description",
		"aria-describedby",
		"aria-details",
		"aria-dropeffect",
		"aria-flowto",
		"aria-grabbed",
		"aria-hidden",
		"aria-keyshortcuts",
		"aria-label",
		"aria-labelledby",
		"aria-live",
		"aria-owns",
		"aria-relevant",
		"aria-roledescription"
	].some(function(n) {
		var r;
		return e.hasAttribute(n) && !((r = $t[t]) != null && r.has(n));
	});
}
function tn(e, t) {
	return en(e, t);
}
function nn(e) {
	var t = an(e);
	if (t === null || on.indexOf(t) !== -1) {
		var n = rn(e);
		if (on.indexOf(t || "") === -1 || tn(e, n || "")) return n;
	}
	return t;
}
function rn(e) {
	var t = Qt[R(e)];
	if (t !== void 0) return t;
	switch (R(e)) {
		case "a":
		case "area":
		case "link":
			if (e.hasAttribute("href")) return "link";
			break;
		case "img": return e.getAttribute("alt") === "" && !tn(e, "img") ? "presentation" : "img";
		case "input":
			var n = e.type;
			switch (n) {
				case "button":
				case "image":
				case "reset":
				case "submit": return "button";
				case "checkbox":
				case "radio": return n;
				case "range": return "slider";
				case "email":
				case "tel":
				case "text":
				case "url": return e.hasAttribute("list") ? "combobox" : "textbox";
				case "search": return e.hasAttribute("list") ? "combobox" : "searchbox";
				case "number": return "spinbutton";
				default: return null;
			}
		case "select": return e.hasAttribute("multiple") || e.size > 1 ? "listbox" : "combobox";
	}
	return null;
}
function an(e) {
	var t = e.getAttribute("role");
	if (t !== null) {
		var n = t.trim().split(" ")[0];
		if (n.length > 0) return n;
	}
	return null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/util.mjs
var on = ["presentation", "none"];
function z(e) {
	return e !== null && e.nodeType === e.ELEMENT_NODE;
}
function B(e) {
	return z(e) && R(e) === "caption";
}
function sn(e) {
	return z(e) && R(e) === "input";
}
function cn(e) {
	return z(e) && R(e) === "optgroup";
}
function ln(e) {
	return z(e) && R(e) === "select";
}
function un(e) {
	return z(e) && R(e) === "table";
}
function dn(e) {
	return z(e) && R(e) === "textarea";
}
function fn(e) {
	var t = (e.ownerDocument === null ? e : e.ownerDocument).defaultView;
	if (t === null) throw TypeError("no window available");
	return t;
}
function pn(e) {
	return z(e) && R(e) === "fieldset";
}
function mn(e) {
	return z(e) && R(e) === "legend";
}
function hn(e) {
	return z(e) && R(e) === "slot";
}
function gn(e) {
	return z(e) && e.ownerSVGElement !== void 0;
}
function _n(e) {
	return z(e) && R(e) === "svg";
}
function vn(e) {
	return gn(e) && R(e) === "title";
}
function yn(e, t) {
	if (z(e) && e.hasAttribute(t)) {
		var n = e.getAttribute(t).split(" "), r = e.getRootNode ? e.getRootNode() : e.ownerDocument;
		return n.map(function(e) {
			return r.getElementById(e);
		}).filter(function(e) {
			return e !== null;
		});
	}
	return [];
}
function bn(e, t) {
	return z(e) ? t.indexOf(nn(e)) !== -1 : !1;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/accessible-name-and-description.mjs
function xn(e) {
	return e.trim().replace(/\s\s+/g, " ");
}
function Sn(e, t) {
	if (!z(e)) return !1;
	if (e.hasAttribute("hidden") || e.getAttribute("aria-hidden") === "true") return !0;
	var n = t(e);
	return n.getPropertyValue("display") === "none" || n.getPropertyValue("visibility") === "hidden";
}
function Cn(e) {
	return bn(e, [
		"button",
		"combobox",
		"listbox",
		"textbox"
	]) || wn(e, "range");
}
function wn(e, t) {
	if (!z(e)) return !1;
	switch (t) {
		case "range": return bn(e, [
			"meter",
			"progressbar",
			"scrollbar",
			"slider",
			"spinbutton"
		]);
		default: throw TypeError(`No knowledge about abstract role '${t}'. This is likely a bug :(`);
	}
}
function Tn(e, t) {
	var n = Ut(e.querySelectorAll(t));
	return yn(e, "aria-owns").forEach(function(e) {
		n.push.apply(n, Ut(e.querySelectorAll(t)));
	}), n;
}
function En(e) {
	return ln(e) ? e.selectedOptions || Tn(e, "[selected]") : Tn(e, "[aria-selected=\"true\"]");
}
function Dn(e) {
	return bn(e, on);
}
function On(e) {
	return B(e);
}
function kn(e) {
	return bn(e, [
		"button",
		"cell",
		"checkbox",
		"columnheader",
		"gridcell",
		"heading",
		"label",
		"legend",
		"link",
		"menuitem",
		"menuitemcheckbox",
		"menuitemradio",
		"option",
		"radio",
		"row",
		"rowheader",
		"switch",
		"tab",
		"tooltip",
		"treeitem"
	]);
}
function An(e) {
	return !1;
}
function jn(e) {
	return sn(e) || dn(e) ? e.value : e.textContent || "";
}
function Mn(e) {
	var t = e.getPropertyValue("content");
	return /^["'].*["']$/.test(t) ? t.slice(1, -1) : "";
}
function Nn(e) {
	var t = R(e);
	return t === "button" || t === "input" && e.getAttribute("type") !== "hidden" || t === "meter" || t === "output" || t === "progress" || t === "select" || t === "textarea";
}
function Pn(e) {
	if (Nn(e)) return e;
	var t = null;
	return e.childNodes.forEach(function(e) {
		if (t === null && z(e)) {
			var n = Pn(e);
			n !== null && (t = n);
		}
	}), t;
}
function Fn(e) {
	if (e.control !== void 0) return e.control;
	var t = e.getAttribute("for");
	return t === null ? Pn(e) : e.ownerDocument.getElementById(t);
}
function In(e) {
	var t = e.labels;
	if (t === null) return t;
	if (t !== void 0) return Ut(t);
	if (!Nn(e)) return null;
	var n = e.ownerDocument;
	return Ut(n.querySelectorAll("label")).filter(function(t) {
		return Fn(t) === e;
	});
}
function Ln(e) {
	var t = e.assignedNodes();
	return t.length === 0 ? Ut(e.childNodes) : t;
}
function Rn(e) {
	var t = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {}, n = new Zt(), r = fn(e), i = t.compute, a = i === void 0 ? "name" : i, o = t.computedStyleSupportsPseudoElements, s = o === void 0 ? t.getComputedStyle !== void 0 : o, c = t.getComputedStyle, l = c === void 0 ? r.getComputedStyle.bind(r) : c, u = t.hidden, d = u !== void 0 && u;
	function f(e, t) {
		var n = "";
		if (z(e) && s && (n = `${Mn(l(e, "::before"))} ${n}`), (hn(e) ? Ln(e) : Ut(e.childNodes).concat(yn(e, "aria-owns"))).forEach(function(e) {
			var r = g(e, {
				isEmbeddedInLabel: t.isEmbeddedInLabel,
				isReferenced: !1,
				recursion: !0
			}), i = (z(e) ? l(e).getPropertyValue("display") : "inline") === "inline" ? "" : " ";
			n += `${i}${r}${i}`;
		}), z(e) && s) {
			var r = Mn(l(e, "::after"));
			n = `${n} ${r}`;
		}
		return n.trim();
	}
	function p(e, t) {
		var r = e.getAttributeNode(t);
		return r !== null && !n.has(r) && r.value.trim() !== "" ? (n.add(r), r.value) : null;
	}
	function m(e) {
		return z(e) ? p(e, "title") : null;
	}
	function h(e) {
		if (!z(e)) return null;
		if (pn(e)) {
			n.add(e);
			for (var t = Ut(e.childNodes), r = 0; r < t.length; r += 1) {
				var i = t[r];
				if (mn(i)) return g(i, {
					isEmbeddedInLabel: !1,
					isReferenced: !1,
					recursion: !1
				});
			}
		} else if (un(e)) {
			n.add(e);
			for (var a = Ut(e.childNodes), o = 0; o < a.length; o += 1) {
				var s = a[o];
				if (B(s)) return g(s, {
					isEmbeddedInLabel: !1,
					isReferenced: !1,
					recursion: !1
				});
			}
		} else if (_n(e)) {
			n.add(e);
			for (var c = Ut(e.childNodes), l = 0; l < c.length; l += 1) {
				var u = c[l];
				if (vn(u)) return u.textContent;
			}
			return null;
		} else if (R(e) === "img" || R(e) === "area") {
			var d = p(e, "alt");
			if (d !== null) return d;
		} else if (cn(e)) {
			var m = p(e, "label");
			if (m !== null) return m;
		}
		if (sn(e) && (e.type === "button" || e.type === "submit" || e.type === "reset")) {
			var h = p(e, "value");
			if (h !== null) return h;
			if (e.type === "submit") return "Submit";
			if (e.type === "reset") return "Reset";
		}
		var _ = In(e);
		if (_ !== null && _.length !== 0) return n.add(e), Ut(_).map(function(e) {
			return g(e, {
				isEmbeddedInLabel: !0,
				isReferenced: !1,
				recursion: !0
			});
		}).filter(function(e) {
			return e.length > 0;
		}).join(" ");
		if (sn(e) && e.type === "image") {
			var v = p(e, "alt");
			if (v !== null) return v;
			var y = p(e, "title");
			return y === null ? "Submit Query" : y;
		}
		if (bn(e, ["button"])) {
			var b = f(e, {
				isEmbeddedInLabel: !1,
				isReferenced: !1
			});
			if (b !== "") return b;
		}
		return null;
	}
	function g(e, t) {
		if (n.has(e)) return "";
		if (!d && Sn(e, l) && !t.isReferenced) return n.add(e), "";
		var r = z(e) ? e.getAttributeNode("aria-labelledby") : null, i = r !== null && !n.has(r) ? yn(e, "aria-labelledby") : [];
		if (a === "name" && !t.isReferenced && i.length > 0) return n.add(r), i.map(function(e) {
			return g(e, {
				isEmbeddedInLabel: t.isEmbeddedInLabel,
				isReferenced: !0,
				recursion: !1
			});
		}).join(" ");
		var o = t.recursion && Cn(e) && a === "name";
		if (!o) {
			var s = (z(e) && e.getAttribute("aria-label") || "").trim();
			if (s !== "" && a === "name") return n.add(e), s;
			if (!Dn(e)) {
				var c = h(e);
				if (c !== null) return n.add(e), c;
			}
		}
		if (bn(e, ["menu"])) return n.add(e), "";
		if (o || t.isEmbeddedInLabel || t.isReferenced) {
			if (bn(e, ["combobox", "listbox"])) {
				n.add(e);
				var u = En(e);
				return u.length === 0 ? sn(e) ? e.value : "" : Ut(u).map(function(e) {
					return g(e, {
						isEmbeddedInLabel: t.isEmbeddedInLabel,
						isReferenced: !1,
						recursion: !0
					});
				}).join(" ");
			}
			if (wn(e, "range")) return n.add(e), e.hasAttribute("aria-valuetext") ? e.getAttribute("aria-valuetext") : e.hasAttribute("aria-valuenow") ? e.getAttribute("aria-valuenow") : e.getAttribute("value") || "";
			if (bn(e, ["textbox"])) return n.add(e), jn(e);
		}
		if (kn(e) || z(e) && t.isReferenced || On(e) || An(e)) {
			var p = f(e, {
				isEmbeddedInLabel: t.isEmbeddedInLabel,
				isReferenced: !1
			});
			if (p !== "") return n.add(e), p;
		}
		if (e.nodeType === e.TEXT_NODE) return n.add(e), e.textContent || "";
		if (t.recursion) return n.add(e), f(e, {
			isEmbeddedInLabel: t.isEmbeddedInLabel,
			isReferenced: !1
		});
		var _ = m(e);
		return _ === null ? (n.add(e), "") : (n.add(e), _);
	}
	return xn(g(e, {
		isEmbeddedInLabel: !1,
		isReferenced: a === "description",
		recursion: !1
	}));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dom-accessibility-api/0.6.3/939b8cb1c72add2493e9017fdfa095253404db9c2f38dbbcfd4032c21ba2fdbe/node_modules/dom-accessibility-api/dist/accessible-name.mjs
function zn(e) {
	return bn(e, [
		"caption",
		"code",
		"deletion",
		"emphasis",
		"generic",
		"insertion",
		"none",
		"paragraph",
		"presentation",
		"strong",
		"subscript",
		"superscript"
	]);
}
function Bn(e) {
	var t = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
	return zn(e) ? "" : Rn(e, t);
}
// ../../../clicky-ui/packages/ui/src/comments/anchoring/markdown-anchor.ts
var Vn = "[data-annotated-markdown]", Hn = "[data-source-start]", Un = "li,tr", Wn = /^h([1-6])$/, Gn = .6;
function Kn(e) {
	return (e ?? "").replace(/\s+/g, " ").trim();
}
function qn(e) {
	return Kn(e.textContent).slice(0, 200);
}
function Jn(e) {
	let t = [];
	return [...e.querySelectorAll(Hn)].filter((e) => !e.parentElement?.closest(Hn)).map((e) => {
		let n = Wn.exec(e.tagName.toLowerCase())?.[1];
		if (n) {
			for (; t.length > 0 && (t.at(-1)?.level ?? 0) >= Number(n);) t.pop();
			t.push({
				level: Number(n),
				text: qn(e)
			});
		}
		return {
			block: e,
			path: t.map((e) => e.text),
			line: Number(e.dataset.sourceStart)
		};
	});
}
function Yn(e) {
	return Jn(e).flatMap(({ block: e, path: t, line: n }) => [{
		element: e,
		path: t,
		line: n
	}, ...[...e.querySelectorAll(Un)].map((e) => ({
		element: e,
		path: t,
		line: n
	}))]);
}
function Xn(e, t) {
	return e.length === t.length && e.every((e, n) => e === t[n]);
}
function Zn(e) {
	let t = e.closest(Vn), n = e.closest(Hn);
	if (!t || !n || !t.contains(n)) return null;
	let r = e.closest(Un), i = r && n.contains(r) ? r : n;
	if (qn(i) !== "") return {
		root: t,
		block: i
	};
	let a = Jn(t), o = a.findIndex((e) => e.block === n || e.block.contains(n)), s = a.slice(0, o + 1).reverse().find((e) => Wn.test(e.block.tagName.toLowerCase()));
	return s ? {
		root: t,
		block: s.block
	} : null;
}
function Qn(e, t) {
	let n = Yn(t).find((t) => t.element === e);
	if (!n) throw Error("markdownLocator: the block is not part of the rendered markdown");
	return {
		kind: "markdown",
		path: n.path,
		tag: e.tagName.toLowerCase(),
		text: qn(e),
		...Number.isInteger(n.line) && n.line > 0 ? { line: n.line } : {}
	};
}
function $n(e) {
	return new Set(e.toLowerCase().match(/[\p{L}\p{N}]+/gu) ?? []);
}
function er(e, t) {
	let n = $n(e), r = $n(t), i = [...n].filter((e) => r.has(e)).length, a = n.size + r.size - i;
	return a === 0 ? 0 : i / a;
}
function tr(e, t) {
	if (e.length <= 1) return e[0]?.element ?? null;
	if (t === void 0) return null;
	let n = (e) => Math.abs(e.line - t), [r, i] = [...e].sort((e, t) => n(e) - n(t));
	return r && i && n(r) === n(i) ? null : r?.element ?? null;
}
function nr(e, t) {
	let n = Yn(e).filter((e) => e.element.tagName.toLowerCase() === t.tag), r = n.filter((e) => Xn(e.path, t.path)), i = (e) => qn(e.element) === t.text, a = tr(r.filter(i), t.line);
	if (a) return {
		element: a,
		how: "exact"
	};
	let [o, s] = r.map((e) => ({
		candidate: e,
		score: er(qn(e.element), t.text)
	})).filter(({ score: e }) => e >= Gn).sort((e, t) => t.score - e.score);
	if (o && o.score !== s?.score) return {
		element: o.candidate.element,
		how: "edited"
	};
	let c = tr(n.filter(i), t.line);
	return c ? {
		element: c,
		how: "relocated"
	} : null;
}
// ../../../clicky-ui/packages/ui/src/comments/anchoring/frame-anchor.ts
var rr = "[data-page-agent],[data-react-grab],[data-react-grab-ignore]";
function ir(e) {
	return e.localName !== "iframe" || e.closest(rr) ? null : e.contentDocument;
}
function ar(e) {
	let t = new URL(e.URL);
	return t.username = "", t.password = "", t.hash = "", t.href;
}
function or(e, t) {
	let n = [];
	for (let r = e; r.ownerDocument !== t.ownerDocument;) {
		let e = r.ownerDocument, i = e.defaultView?.frameElement, a = e.body;
		if (!a?.contains(r) || r.closest(rr) || !i || ir(i) !== e) return null;
		let o = i.ownerDocument === t.ownerDocument ? t : i.ownerDocument.body;
		if (!o?.contains(i)) return null;
		n.unshift({
			frame: i,
			parent: o,
			body: a,
			url: ar(e)
		}), r = i;
	}
	return n.length ? n : null;
}
// ../../../clicky-ui/packages/ui/src/comments/anchoring/dom-anchor.ts
var sr = ":scope", cr = "data-comment-anchor", lr = /* @__PURE__ */ new Set([cr, "data-testid"]), ur = /* @__PURE__ */ new Set([
	"role",
	"aria-label",
	"name"
]), dr = /^data-(?:react-grab|review|comment|show-rejected)/;
function fr(e, t) {
	return e.startsWith("data-") && t === "" && !dr.test(e);
}
var pr = /* @__PURE__ */ new Set([
	"generic",
	"presentation",
	"none"
]), mr = [
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"[role=\"heading\"]",
	"figcaption",
	"legend"
], hr = mr.join(","), gr = "figcaption,legend", _r = "section,article,figure,li", vr = [
	/^(?::r[0-9a-z]+:|«r[0-9a-z]+»|_r_[0-9a-z]+_)/i,
	/^r[0-9a-v]{1,5}(?:-|$)/,
	/^(?:radix|headlessui|react-aria)[-:]/,
	/^mui-\d/
];
function yr(e) {
	return vr.some((t) => t.test(e));
}
function br(e, t) {
	if (e === t) return sr;
	let n = e.getAttribute(cr);
	if (n) {
		let e = `[${cr}="${CSS.escape(n)}"]`;
		if (Rr(t, e).length === 1) return e;
	}
	return Tt(e, {
		root: t,
		idName: (e) => !yr(e),
		className: () => !1,
		attr: (e, t) => lr.has(e) || fr(e, t) || ur.has(e) && xt(e, t)
	});
}
function xr(e) {
	return (e ?? "").replace(/\s+/g, " ").trim();
}
function Sr(e) {
	return e !== "" && e.length <= 200;
}
function Cr(e) {
	return e.length === 1 ? e[0] ?? null : null;
}
function wr(e) {
	let t = mr.find((t) => e.matches(t)), n = xr(e.textContent);
	return t && Sr(n) ? {
		selector: t,
		text: n
	} : null;
}
function Tr(e, t) {
	for (let n = e; n && n !== t; n = n.parentElement) {
		for (let e = n.previousElementSibling; e; e = e.previousElementSibling) {
			let t = e.firstElementChild, n = e.matches(hr) ? e : !e.matches(_r) && t?.matches(hr) ? t : null, r = n && wr(n);
			if (r) return r;
		}
		let e = n.parentElement ? [...n.parentElement.children].find((e) => e !== n && e.matches(gr)) : void 0, t = e && wr(e);
		if (t) return t;
	}
	return null;
}
function Er(e, t) {
	for (let n = e.parentElement; n && n !== t; n = n.parentElement) {
		let e = n.getAttribute("id");
		if (!e || yr(e)) continue;
		let r = `#${CSS.escape(e)}`;
		if (Rr(t, r).length === 1) return r;
	}
	return null;
}
function Dr(e, t, n) {
	return [e, ...Rr(e, t)].filter((e) => e.tagName.toLowerCase() === t && xr(e.textContent) === n);
}
function Or(e, t) {
	switch (t.kind) {
		case "css": return zr(e, t.selector);
		case "role": return Cr(Rr(e, t.tag).filter((e) => nn(e) === t.role && Bn(e) === t.name));
		case "text": {
			let n = t.scope ? zr(e, t.scope) : e;
			if (!n) return null;
			let { heading: r } = t;
			if (!r) return Cr(Dr(n, t.tag, t.text));
			for (let e = Cr(Rr(n, r.selector).filter((e) => xr(e.textContent) === r.text)); e && n.contains(e); e = e.parentElement) {
				let n = Dr(e, t.tag, t.text);
				if (n.length > 0) return Cr(n);
			}
			return null;
		}
		case "markdown": {
			let n = kr(e);
			return n ? nr(n, t)?.element ?? null : null;
		}
	}
}
function kr(e) {
	return e.matches("[data-annotated-markdown]") ? e : e.querySelector(Vn);
}
function Ar(e, t, n) {
	if (n?.frames && e.ownerDocument !== t.ownerDocument) {
		let n = or(e, t)?.at(-1);
		if (!n) return [];
		t = n.body;
	}
	if (e === t || !t.contains(e)) return [];
	let r = Zn(e);
	if (r) return [Qn(r.block, r.root)];
	let i = e.tagName.toLowerCase(), a = [], o = br(e, t);
	o.length <= 1024 && a.push({
		kind: "css",
		selector: o
	});
	let s = nn(e), c = Bn(e);
	s && !pr.has(s) && Sr(c) && a.push({
		kind: "role",
		tag: i,
		role: s,
		name: c
	});
	let l = xr(e.textContent);
	if (Sr(l)) {
		let n = Tr(e, t);
		n && a.push({
			kind: "text",
			tag: i,
			text: l,
			heading: n
		});
		let r = Er(e, t);
		r && a.push({
			kind: "text",
			tag: i,
			text: l,
			scope: r
		}), a.push({
			kind: "text",
			tag: i,
			text: l
		});
	}
	return a.filter((n) => Or(t, n) === e);
}
function jr(e, t) {
	return t.flatMap((t) => {
		let n = Or(e, t);
		return n ? [{
			by: t,
			element: n
		}] : [];
	});
}
function Mr(e) {
	if (e.length === 0) throw Error("an element anchor needs at least one locator");
	return `e:${JSON.stringify(e)}`;
}
function Nr(e) {
	return e.startsWith("e:");
}
function Pr(e) {
	if (!Nr(e)) throw Error(`expected an "e:" element anchor, got ${JSON.stringify(e.slice(0, 40))}`);
	let t;
	try {
		t = JSON.parse(e.slice(2));
	} catch (e) {
		throw Error(`element anchor is not valid JSON: ${e instanceof Error ? e.message : String(e)}`);
	}
	if (!Array.isArray(t) || t.length === 0 || t.some((e) => typeof e != "object" || !e || typeof e.kind != "string")) throw Error("element anchor is not a non-empty list of locators");
	return t;
}
function Fr(e, t) {
	return Ir(e, Pr(t));
}
function Ir(e, t) {
	return Lr(jr(e, t));
}
function Lr(e) {
	let t = e[0];
	if (!t) return null;
	let n = /* @__PURE__ */ new Map();
	for (let { element: t } of e) n.set(t, (n.get(t) ?? 0) + 1);
	let r = e.reduce((e, t) => (n.get(t.element) ?? 0) > (n.get(e.element) ?? 0) ? t : e, t);
	return {
		element: r.element,
		matchedBy: r.by,
		conflict: n.size > 1
	};
}
function Rr(e, t) {
	try {
		return [...e.querySelectorAll(t)];
	} catch {
		return [];
	}
}
function zr(e, t) {
	return Cr(Rr(e, t));
}
// ../../../clicky-ui/packages/ui/src/comments/anchoring/drift.ts
var Br = {
	moved: "Moved",
	ambiguous: "Ambiguous",
	outdated: "Outdated",
	orphaned: "Orphaned"
};
function Vr(e) {
	return e.state === "matched" ? null : e.state;
}
function Hr(e) {
	return e ? e.conflict ? "moved" : null : "orphaned";
}
var Ur = "Its element was found somewhere other than where it was written; check it before relying on it.", Wr = {
	lines: {
		outdated: "The lines this was written on have changed or been removed.",
		ambiguous: "The lines this was written on appear more than once in the file."
	},
	element: {
		orphaned: "The element this was written on is no longer in the document.",
		moved: Ur
	},
	text: {
		orphaned: "The words this was written on are no longer in the document.",
		moved: Ur,
		ambiguous: "The words this was written on appear more than once where it was written."
	}
};
function Gr(e, t) {
	let n = Wr[t][e];
	if (!n) throw Error(`an anchor on ${t} cannot be ${e}`);
	return n;
}
var Kr = 3, qr = (e) => e.replaceAll("\r\n", "\n").split("\n");
function Jr(e, t, n) {
	let r = qr(e);
	if (t < 1 || n < t || n > r.length) throw Error("Source selection is outside the file");
	return {
		version: 1,
		startLine: t,
		endLine: n,
		before: r.slice(Math.max(0, t - 1 - Kr), t - 1),
		lines: r.slice(t - 1, n),
		after: r.slice(n, n + Kr)
	};
}
var Yr = (e) => Array.isArray(e) && e.every((e) => typeof e == "string");
function Xr(e) {
	let t = e;
	if (!(typeof t == "object" && t && typeof t.version == "number" && typeof t.startLine == "number" && typeof t.endLine == "number" && Yr(t.before) && Yr(t.lines) && Yr(t.after))) throw Error("patch anchor is not a patch: expected {version, startLine, endLine, before, lines, after}");
}
function Zr(e) {
	return e.startsWith("r:");
}
function Qr(e) {
	return `r:${JSON.stringify(e)}`;
}
function $r(e) {
	if (!Zr(e)) throw Error(`expected an "r:" patch anchor, got ${JSON.stringify(e.slice(0, 40))}`);
	let t;
	try {
		t = JSON.parse(e.slice(2));
	} catch (e) {
		throw Error(`patch anchor is not valid JSON: ${e instanceof Error ? e.message : String(e)}`);
	}
	return Xr(t), t;
}
function ei(e) {
	return e.startLine === e.endLine ? `Line ${e.startLine}` : `Lines ${e.startLine}–${e.endLine}`;
}
function ti(e, t) {
	if (!(t.version === 1 && t.lines.length > 0 && t.startLine >= 1 && t.endLine - t.startLine + 1 === t.lines.length && t.before.length <= Kr && t.after.length <= Kr)) return { state: "outdated" };
	let n = qr(e), r = [
		...t.before,
		...t.lines,
		...t.after
	], i = (e) => e >= 0 && e + r.length <= n.length && r.every((t, r) => n[e + r] === t), a = (e) => {
		let n = e + t.before.length + 1;
		return {
			state: "matched",
			startLine: n,
			endLine: n + t.lines.length - 1
		};
	}, o = t.startLine - 1 - t.before.length;
	if (i(o)) return a(o);
	let s = -1;
	for (let e = 0; e + r.length <= n.length; e++) if (i(e)) {
		if (s >= 0) return { state: "ambiguous" };
		s = e;
	}
	return s >= 0 ? a(s) : { state: "outdated" };
}
// ../../../clicky-ui/packages/ui/src/comments/ElementOutline.tsx
function ni({ element: e, root: t }) {
	let n = e.getBoundingClientRect(), r = t.getBoundingClientRect();
	return /* @__PURE__ */ (0, P.jsx)("div", {
		"data-element-outline": "",
		className: "pointer-events-none absolute z-30 border-2 border-primary",
		style: {
			left: n.left - r.left,
			top: n.top - r.top,
			width: n.width,
			height: n.height
		}
	});
}
// ../../../clicky-ui/packages/ui/src/comments/element-picker.ts
var ri = 80;
function ii(e) {
	return (e.getAttribute("aria-label") || e.textContent?.trim() || e.localName).slice(0, ri);
}
function ai({ root: e, onPick: t, accepts: n }) {
	let [r, i] = (0, N.useState)(!1), [a, o] = (0, N.useState)(null), s = () => {
		i(!1), o(null);
	};
	y(r, s);
	let c = (e) => e instanceof Element && (n?.(e) ?? !0);
	return {
		picking: r,
		setPicking: (e) => e ? i(!0) : s(),
		rootProps: {
			onPointerOver: (e) => {
				r && c(e.target) && o(e.target);
			},
			onClickCapture: (n) => {
				r && e.current && c(n.target) && (n.preventDefault(), n.stopPropagation(), t(n.target), s());
			}
		},
		outline: r && a && e.current ? (0, N.createElement)(ni, {
			element: a,
			root: e.current
		}) : null
	};
}
// ../../../clicky-ui/packages/ui/src/comments/comment-types.ts
var oi = "__document__", si = [
	{
		value: "open",
		label: "Open",
		tone: "info",
		unresolved: !0,
		stage: "active"
	},
	{
		value: "in_progress",
		label: "In progress",
		tone: "warning",
		unresolved: !0,
		stage: "active"
	},
	{
		value: "resolved",
		label: "Resolved",
		tone: "success",
		stage: "resolved"
	},
	{
		value: "closed",
		label: "Closed",
		tone: "neutral",
		stage: "closed"
	}
], ci = [
	"open",
	"in_progress",
	"done"
];
// ../../../clicky-ui/packages/ui/src/comments/comment-utils.ts
function li(e) {
	let t = e?.name?.trim() || "Unknown";
	return e?.kind === "agent" && Ie(t) ? je(t) : t;
}
function ui(e) {
	switch (e) {
		case "success": return "success";
		case "danger": return "danger";
		case "warning": return "warning";
		case "info": return "info";
		default: return "neutral";
	}
}
function di(e, t) {
	if (t != null) return e.statuses.find((e) => e.value === t);
}
function fi(e, t) {
	if (t != null) return e.options.find((e) => e.value === t);
}
function pi(e) {
	return e.filter((e) => !e.parentId);
}
function mi(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (!n.parentId) continue;
		let e = t.get(n.parentId) ?? [];
		e.push(n), t.set(n.parentId, e);
	}
	return t;
}
function hi(e, t) {
	let n = e.filter((e) => !e.parentId && (e.anchor ?? "__document__") === t), r = new Set(n.map((e) => String(e.id))), i = [];
	for (let t = !0; t;) {
		t = !1;
		for (let n of e) n.parentId && !r.has(String(n.id)) && r.has(String(n.parentId)) && (r.add(String(n.id)), i.push(n), t = !0);
	}
	return [...n, ...i];
}
function gi(e) {
	return [...e].sort((e, t) => e.createdAt === t.createdAt ? String(e.id).localeCompare(String(t.id)) : e.createdAt.localeCompare(t.createdAt));
}
function _i(e) {
	let t = {};
	for (let n of e) {
		let e = n.anchor ?? "__document__";
		t[e] = (t[e] ?? 0) + 1;
	}
	return t;
}
function vi(e) {
	let t = {};
	for (let n of e) {
		let e = n.anchor ?? "__document__";
		(t[e] ??= []).push(n);
	}
	let n = {};
	for (let [e, r] of Object.entries(t)) {
		let t = [...new Set(r.map((e) => e.author?.name).filter((e) => !!e))], i = r.reduce((e, t) => (e.updatedAt ?? e.createdAt) > (t.updatedAt ?? t.createdAt) ? e : t);
		n[e] = {
			count: r.length,
			authors: t,
			...i.status == null ? {} : { latestStatus: i.status }
		};
	}
	return n;
}
function yi(e, t = ci) {
	let n = t.indexOf(e);
	return n === -1 ? t[0] ?? e : t[(n + 1) % t.length] ?? e;
}
function bi(e, t = 80) {
	let n = e.replace(/[#*_`~>[\]()!]/g, "").trim();
	return n.length > t ? `${n.slice(0, t)}…` : n;
}
function xi(e, t) {
	let n = e.toLowerCase();
	return t.filter((e) => n.includes(`@${e.name.toLowerCase()}`)).map((e) => ({
		id: e.id,
		name: e.name,
		kind: e.kind
	}));
}
function Si(e, t, n) {
	let r = {};
	if (n.onUpdateStatus && (r.onUpdateStatus = n.onUpdateStatus), n.onClose && (r.onClose = n.onClose), n.onAcknowledge && (r.onAcknowledge = n.onAcknowledge), n.onAssign && (r.onAssign = n.onAssign), n.onDelete && (r.onDelete = n.onDelete), n.onEdit && (r.onEdit = n.onEdit), n.onEditReply && (r.onEditReply = n.onEditReply), n.onDeleteReply && (r.onDeleteReply = n.onDeleteReply), n.onCreateTodo && (r.onCreateTodo = n.onCreateTodo), n.onChecklistToggle) {
		let i = n.onChecklistToggle;
		r.onChecklistToggle = (n, r) => {
			let a = e.find((e) => e.id === n)?.checklist?.[r];
			a && i(n, r, yi(a.status, t.checklistStatusCycle));
		};
	}
	if (n.onReply) {
		let e = n.onReply, i = n.onMention;
		r.onReply = async (n, r) => {
			let a = xi(r, t.mentionables ?? []);
			await e({
				parentId: n.id,
				body: r,
				anchor: n.anchor ?? null,
				...a.length > 0 ? { mentions: a } : {}
			});
			for (let e of a) i?.(e, { body: r });
		};
	}
	return r;
}
// ../../../clicky-ui/packages/ui/src/lib/comment-stage.ts
function Ci(e, t) {
	if (t == null) return "active";
	let n = di(e, t);
	if (n) return n.stage ? n.stage : n.unresolved ? "active" : n.value.toLowerCase() === "closed" || n.label.toLowerCase() === "closed" ? "closed" : "resolved";
}
function wi(e, t) {
	return e.statuses.find((n) => Ci(e, n.value) === t);
}
function Ti(e, t, n) {
	let r = new Set(pi(e).filter((e) => Ci(t, e.status) === n).map((e) => String(e.id)));
	return e.filter((e) => r.has(String(e.parentId ?? e.id)));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalComposerContext.prod.mjs
var Ei = (0, N.createContext)(null);
function Di(e, t) {
	let n = null;
	return e != null && (n = e[1]), { getTheme: function() {
		return t ?? (n == null ? null : n.getTheme());
	} };
}
function Oi() {
	let e = (0, N.useContext)(Ei);
	return e ?? function(e, ...t) {
		let n = new URL("https://lexical.dev/docs/error"), r = new URLSearchParams();
		r.append("code", e);
		for (let e of t) r.append("v", e);
		throw n.search = r.toString(), Error(`Minified Lexical error #${e}; visit ${n.toString()} for the full message or use the non-minified dev environment for full errors and additional helpful warnings.`);
	}(8), e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/lexical/0.35.0/9adf9f5aeee923fa8ecaeca56d2a445da8f6a34df8a28444e6f89770e3615572/node_modules/lexical/Lexical.prod.mjs
function V(e, ...t) {
	let n = new URL("https://lexical.dev/docs/error"), r = new URLSearchParams();
	r.append("code", e);
	for (let e of t) r.append("v", e);
	throw n.search = r.toString(), Error(`Minified Lexical error #${e}; visit ${n.toString()} for the full message or use the non-minified dev environment for full errors and additional helpful warnings.`);
}
var ki = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, Ai = ki && "documentMode" in document ? document.documentMode : null, ji = ki && /Mac|iPod|iPhone|iPad/.test(navigator.platform), Mi = ki && /^(?!.*Seamonkey)(?=.*Firefox).*/i.test(navigator.userAgent), Ni = !(!ki || !("InputEvent" in window) || Ai) && "getTargetRanges" in new window.InputEvent("input"), Pi = ki && /Version\/[\d.]+.*Safari/.test(navigator.userAgent), Fi = ki && /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream, Ii = ki && /Android/.test(navigator.userAgent), Li = ki && /^(?=.*Chrome).*/i.test(navigator.userAgent), Ri = ki && Ii && Li, zi = ki && /AppleWebKit\/[\d.]+/.test(navigator.userAgent) && !Li;
function Bi(...e) {
	let t = [];
	for (let n of e) if (n && typeof n == "string") for (let [e] of n.matchAll(/\S+/g)) t.push(e);
	return t;
}
var Vi = 1, Hi = 3, Ui = 9, Wi = 11, Gi = 0, Ki = 1, qi = 2, Ji = 0, Yi = 1, Xi = 2, Zi = 1, Qi = 2, $i = 3, ea = 4, ta = 5, na = 6, ra = Pi || Fi || zi ? "\xA0" : "​", ia = "\n\n", aa = Mi ? "\xA0" : ra, oa = {
	bold: 1,
	capitalize: 1024,
	code: 16,
	highlight: 128,
	italic: 2,
	lowercase: 256,
	strikethrough: 4,
	subscript: 32,
	superscript: 64,
	underline: 8,
	uppercase: 512
}, sa = {
	directionless: 1,
	unmergeable: 2
}, ca = {
	center: Qi,
	end: na,
	justify: ea,
	left: Zi,
	right: $i,
	start: ta
}, la = {
	[Qi]: "center",
	[na]: "end",
	[ea]: "justify",
	[Zi]: "left",
	[$i]: "right",
	[ta]: "start"
}, ua = {
	normal: 0,
	segmented: 2,
	token: 1
}, da = {
	[Ji]: "normal",
	[Xi]: "segmented",
	[Yi]: "token"
}, fa = "$config";
function pa(e, t, n, r, i, a) {
	let o = e.getFirstChild();
	for (; o !== null;) {
		let e = o.__key;
		o.__parent === t && (Y(o) && pa(o, e, n, r, i, a), n.has(e) || a.delete(e), i.push(e)), o = o.getNextSibling();
	}
}
var ma = 100, ha = !1, ga = 0;
function _a(e) {
	ga = e.timeStamp;
}
function va(e, t, n) {
	let r = e.nodeName === "BR", i = t.__lexicalLineBreak;
	return i && (e === i || r && e.previousSibling === i) || r && Yl(e, n) !== void 0;
}
function ya(e, t, n) {
	let r = Lu(ku(n)), i = null, a = null;
	r !== null && r.anchorNode === e && (i = r.anchorOffset, a = r.focusOffset);
	let o = e.nodeValue;
	o !== null && su(t, o, i, a, !1);
}
function ba(e, t, n) {
	if (K(e)) {
		let t = e.anchor.getNode();
		if (t.is(n) && e.format !== t.getFormat()) return !1;
	}
	return Rl(t) && n.isAttached();
}
function xa(e, t, n, r) {
	for (let i = e; i && !Qu(i); i = Cu(i)) {
		let e = Yl(i, t);
		if (e !== void 0) {
			let t = ql(e, n);
			if (t) return X(t) || !zu(i) ? void 0 : [i, t];
		} else if (i === r) return [r, $l(n)];
	}
}
function Sa(e, t, n) {
	ha = !0;
	let r = performance.now() - ga > ma;
	try {
		nl(e, (() => {
			let i = q() || function(e) {
				return e.getEditorState().read((() => {
					let e = q();
					return e === null ? null : e.clone();
				}));
			}(e), a = /* @__PURE__ */ new Map(), o = e.getRootElement(), s = e._editorState, c = e._blockCursorElement, l = !1, u = "";
			for (let n = 0; n < t.length; n++) {
				let d = t[n], f = d.type, p = d.target, m = xa(p, e, s, o);
				if (!m) continue;
				let [h, g] = m;
				if (f === "characterData") r && G(g) && Rl(p) && ba(i, p, g) && ya(p, g, e);
				else if (f === "childList") {
					l = !0;
					let t = d.addedNodes;
					for (let n = 0; n < t.length; n++) {
						let r = t[n], i = Jl(r), a = r.parentNode;
						if (a != null && r !== c && i === null && !va(r, a, e)) {
							if (Mi) {
								let e = (zu(r) ? r.innerText : null) || r.nodeValue;
								e && (u += e);
							}
							a.removeChild(r);
						}
					}
					let n = d.removedNodes, r = n.length;
					if (r > 0) {
						let t = 0;
						for (let i = 0; i < r; i++) {
							let r = n[i];
							(va(r, p, e) || c === r) && (p.appendChild(r), t++);
						}
						r !== t && a.set(h, g);
					}
				}
			}
			if (a.size > 0) for (let [t, n] of a) n.reconcileObservedMutation(t, e);
			let d = n.takeRecords();
			if (d.length > 0) {
				for (let t = 0; t < d.length; t++) {
					let n = d[t], r = n.addedNodes, i = n.target;
					for (let t = 0; t < r.length; t++) {
						let n = r[t], a = n.parentNode;
						a == null || n.nodeName !== "BR" || va(n, i, e) || a.removeChild(n);
					}
				}
				n.takeRecords();
			}
			i !== null && (l && eu(i), Mi && bu(e) && i.insertRawText(u));
		}));
	} finally {
		ha = !1;
	}
}
function Ca(e) {
	let t = e._observer;
	t !== null && Sa(e, t.takeRecords(), t);
}
function wa(e) {
	(function(e) {
		ga === 0 && ku(e).addEventListener("textInput", _a, !0);
	})(e), e._observer = new MutationObserver(((t, n) => {
		Sa(e, t, n);
	}));
}
function Ta(e) {
	let t = /* @__PURE__ */ new Map(), n = /* @__PURE__ */ new Set();
	for (let r = typeof e == "function" ? e : e.replace; r.prototype && r.prototype.getType !== void 0; r = Object.getPrototypeOf(r)) {
		let { ownNodeConfig: e } = ed(r);
		if (e && e.stateConfigs) for (let r of e.stateConfigs) {
			let e;
			"stateConfig" in r ? (e = r.stateConfig, r.flat && n.add(e.key)) : e = r, t.set(e.key, e);
		}
	}
	return {
		flatKeys: n,
		sharedConfigMap: t
	};
}
var Ea = class e {
	constructor(e, t, n = void 0, r = /* @__PURE__ */ new Map(), i = void 0) {
		this.node = e, this.sharedNodeState = t, this.unknownState = n, this.knownState = r;
		let { sharedConfigMap: a } = this.sharedNodeState, o = i === void 0 ? function(e, t, n) {
			let r = n.size;
			if (t) for (let i in t) {
				let t = e.get(i);
				t && n.has(t) || r++;
			}
			return r;
		}(a, n, r) : i;
		this.size = o;
	}
	getValue(e) {
		let t = this.knownState.get(e);
		if (t !== void 0) return t;
		this.sharedNodeState.sharedConfigMap.set(e.key, e);
		let n = e.defaultValue;
		if (this.unknownState && e.key in this.unknownState) {
			let t = this.unknownState[e.key];
			t !== void 0 && (n = e.parse(t)), this.updateFromKnown(e, n);
		}
		return n;
	}
	getInternalState() {
		return [this.unknownState, this.knownState];
	}
	toJSON() {
		let e = { ...this.unknownState }, t = {};
		for (let [t, n] of this.knownState) t.isEqual(n, t.defaultValue) ? delete e[t.key] : e[t.key] = t.unparse(n);
		for (let n of this.sharedNodeState.flatKeys) n in e && (t[n] = e[n], delete e[n]);
		return ka(e) && (t.$ = e), t;
	}
	getWritable(t) {
		if (this.node === t) return this;
		let { sharedNodeState: n, unknownState: r } = this, i = new Map(this.knownState);
		return new e(t, n, function(e, t, n) {
			let r;
			if (n) for (let [i, a] of Object.entries(n)) {
				let n = e.get(i);
				n ? t.has(n) || t.set(n, n.parse(a)) : (r ||= {}, r[i] = a);
			}
			return r;
		}(n.sharedConfigMap, i, r), i, this.size);
	}
	updateFromKnown(e, t) {
		let n = e.key;
		this.sharedNodeState.sharedConfigMap.set(n, e);
		let { knownState: r, unknownState: i } = this;
		r.has(e) || i && n in i || (i && (delete i[n], this.unknownState = ka(i)), this.size++), r.set(e, t);
	}
	updateFromUnknown(e, t) {
		let n = this.sharedNodeState.sharedConfigMap.get(e);
		n ? this.updateFromKnown(n, n.parse(t)) : (this.unknownState = this.unknownState || {}, e in this.unknownState || this.size++, this.unknownState[e] = t);
	}
	updateFromJSON(e) {
		let { knownState: t } = this;
		for (let e of t.keys()) t.set(e, e.defaultValue);
		if (this.size = t.size, this.unknownState = void 0, e) for (let [t, n] of Object.entries(e)) this.updateFromUnknown(t, n);
	}
};
function Da(e) {
	let t = e.getWritable(), n = t.__state ? t.__state.getWritable(t) : new Ea(t, Oa(t));
	return t.__state = n, n;
}
function Oa(e) {
	return e.__state ? e.__state.sharedNodeState : Dl(Ku(), e.getType()).sharedNodeState;
}
function ka(e) {
	if (e) for (let t in e) return e;
}
function Aa(e, t, n) {
	for (let [r, i] of t.knownState) {
		if (e.has(r.key)) continue;
		e.add(r.key);
		let t = n ? n.getValue(r) : r.defaultValue;
		if (t !== i && !r.isEqual(t, i)) return !0;
	}
	return !1;
}
function ja(e, t, n) {
	let { unknownState: r } = t, i = n ? n.unknownState : void 0;
	if (r) {
		for (let [t, n] of Object.entries(r)) if (!e.has(t) && (e.add(t), n !== (i ? i[t] : void 0))) return !0;
	}
	return !1;
}
function Ma(e, t) {
	let n = e.__state;
	return n && n.node === e ? n.getWritable(t) : n;
}
function Na(e, t) {
	let n = e.__mode, r = e.__format, i = e.__style, a = t.__mode, o = t.__format, s = t.__style, c = e.__state, l = t.__state;
	return (n === null || n === a) && (r === null || r === o) && (i === null || i === s) && (e.__state === null || c === l || function(e, t) {
		if (e === t) return !0;
		if (e && t && e.size !== t.size) return !1;
		let n = /* @__PURE__ */ new Set();
		return !(e && Aa(n, e, t) || t && Aa(n, t, e) || e && ja(n, e, t) || t && ja(n, t, e));
	}(c, l));
}
function Pa(e, t) {
	let n = e.mergeWithSibling(t), r = J()._normalizedNodes;
	return r.add(e.__key), r.add(t.__key), n;
}
function Fa(e) {
	let t, n, r = e;
	if (r.__text !== "" || !r.isSimpleText() || r.isUnmergeable()) {
		for (; (t = r.getPreviousSibling()) !== null && G(t) && t.isSimpleText() && !t.isUnmergeable();) {
			if (t.__text !== "") {
				if (Na(t, r)) {
					r = Pa(t, r);
					break;
				}
				break;
			}
			t.remove();
		}
		for (; (n = r.getNextSibling()) !== null && G(n) && n.isSimpleText() && !n.isUnmergeable();) {
			if (n.__text !== "") {
				if (Na(r, n)) {
					r = Pa(r, n);
					break;
				}
				break;
			}
			n.remove();
		}
	} else r.remove();
}
function Ia(e) {
	return La(e.anchor), La(e.focus), e;
}
function La(e) {
	for (; e.type === "element";) {
		let t = e.getNode(), n = e.offset, r, i;
		if (n === t.getChildrenSize() ? (r = t.getChildAtIndex(n - 1), i = !0) : (r = t.getChildAtIndex(n), i = !1), G(r)) {
			e.set(r.__key, i ? r.getTextContentSize() : 0, "text", !0);
			break;
		}
		if (!Y(r)) break;
		e.set(r.__key, i ? r.getChildrenSize() : 0, "element", !0);
	}
}
var Ra, za, Ba, Va, Ha, Ua, Wa, Ga, Ka, qa, H = "", Ja = null, Ya = "", Xa = "", Za = !1, Qa = !1;
function $a(e, t) {
	let n = Wa.get(e);
	if (t !== null) {
		let n = _o(e);
		n.parentNode === t && t.removeChild(n);
	}
	if (Ga.has(e) || za._keyToDOMMap.delete(e), Y(n)) {
		let e = fo(n, Wa);
		eo(e, 0, e.length - 1, null);
	}
	n !== void 0 && _u(qa, Ba, Va, n, "destroyed");
}
function eo(e, t, n, r) {
	let i = t;
	for (; i <= n; ++i) {
		let t = e[i];
		t !== void 0 && $a(t, r);
	}
}
function to(e, t) {
	e.setProperty("text-align", t);
}
var no = "40px";
function ro(e, t) {
	let n = Ra.theme.indent;
	if (typeof n == "string") {
		let r = e.classList.contains(n);
		t > 0 && !r ? e.classList.add(n) : t < 1 && r && e.classList.remove(n);
	}
	let r = getComputedStyle(e).getPropertyValue("--lexical-indent-base-value") || no;
	e.style.setProperty("padding-inline-start", t === 0 ? "" : `calc(${t} * ${r})`);
}
function io(e, t) {
	let n = e.style;
	t === 0 ? to(n, "") : t === Zi ? to(n, "left") : t === Qi ? to(n, "center") : t === $i ? to(n, "right") : t === ea ? to(n, "justify") : t === ta ? to(n, "start") : t === na && to(n, "end");
}
function ao(e, t) {
	let n = function(e) {
		let t = e.__dir;
		if (t !== null) return t;
		if (ll(e)) return null;
		let n = e.getParentOrThrow();
		return ll(n) && n.__dir === null ? "auto" : null;
	}(t);
	n === null ? e.removeAttribute("dir") : e.dir = n;
}
function oo(e, t) {
	let n = Ga.get(e);
	n === void 0 && V(60);
	let r = n.createDOM(Ra, za);
	if (function(e, t, n) {
		let r = n._keyToDOMMap;
		(function(e, t, n) {
			let r = `__lexicalKey_${t._key}`;
			e[r] = n;
		})(t, n, e), r.set(e, t);
	}(e, r, za), G(n) ? r.setAttribute("data-lexical-text", "true") : X(n) && r.setAttribute("data-lexical-decorator", "true"), Y(n)) {
		let e = n.__indent, t = n.__size;
		if (ao(r, n), e !== 0 && ro(r, e), t !== 0) {
			let e = t - 1;
			so(fo(n, Ga), n, 0, e, n.getDOMSlot(r));
		}
		let i = n.__format;
		i !== 0 && io(r, i), n.isInline() || lo(null, n, r), xu(n) && (H += ia, Xa += ia);
	} else {
		let t = n.getTextContent();
		if (X(n)) {
			let t = n.decorate(za, Ra);
			t !== null && mo(e, t), r.contentEditable = "false";
		}
		H += t, Xa += t;
	}
	return t !== null && t.insertChild(r), _u(qa, Ba, Va, n, "created"), r;
}
function so(e, t, n, r, i) {
	let a = H;
	H = "";
	let o = n;
	for (; o <= r; ++o) {
		oo(e[o], i);
		let t = Ga.get(e[o]);
		t !== null && G(t) && (Ja === null && (Ja = t.getFormat()), Ya === "" && (Ya = t.getStyle()));
	}
	xu(t) && (H += ia), i.element.__lexicalTextContent = H, H = a + H;
}
function co(e, t) {
	if (e) {
		let n = e.__last;
		if (n) {
			let e = t.get(n);
			if (e) return Rs(e) ? "line-break" : X(e) && e.isInline() ? "decorator" : null;
		}
		return "empty";
	}
	return null;
}
function lo(e, t, n) {
	let r = co(e, Wa), i = co(t, Ga);
	r !== i && t.getDOMSlot(n).setManagedLineBreak(i);
}
function uo(e, t, n) {
	var r;
	Ja = null, Ya = "", function(e, t, n) {
		let r = H, i = e.__size, a = t.__size;
		H = "";
		let o = n.element;
		if (i === 1 && a === 1) {
			let n = e.__first, r = t.__first;
			if (n === r) po(n, o);
			else {
				let e = _o(n), t = oo(r, null);
				try {
					o.replaceChild(t, e);
				} catch (i) {
					if (typeof i == "object" && i) {
						let a = `${i.toString()} Parent: ${o.tagName}, new child: {tag: ${t.tagName} key: ${r}}, old child: {tag: ${e.tagName}, key: ${n}}.`;
						throw Error(a);
					}
					throw i;
				}
				$a(n, null);
			}
			let i = Ga.get(r);
			G(i) && (Ja === null && (Ja = i.getFormat()), Ya === "" && (Ya = i.getStyle()));
		} else {
			let r = fo(e, Wa), s = fo(t, Ga);
			if (r.length !== i && V(227), s.length !== a && V(228), i === 0) a !== 0 && so(s, t, 0, a - 1, n);
			else if (a === 0) {
				if (i !== 0) {
					let e = n.after == null && n.before == null && n.element.__lexicalLineBreak == null;
					eo(r, 0, i - 1, e ? null : o), e && (o.textContent = "");
				}
			} else (function(e, t, n, r, i, a) {
				let o = r - 1, s = i - 1, c, l, u = a.getFirstChild(), d = 0, f = 0;
				for (; d <= o && f <= s;) {
					let e = t[d], r = n[f];
					if (e === r) u = ho(po(r, a.element)), d++, f++;
					else {
						c === void 0 && (c = new Set(t)), l === void 0 && (l = new Set(n));
						let i = l.has(e), o = c.has(r);
						if (i) {
							if (o) {
								let e = Su(za, r);
								e === u ? u = ho(po(r, a.element)) : (a.withBefore(u).insertChild(e), po(r, a.element)), d++, f++;
							} else oo(r, a.withBefore(u)), f++;
						} else u = ho(_o(e)), $a(e, a.element), d++;
					}
					let i = Ga.get(r);
					i !== null && G(i) && (Ja === null && (Ja = i.getFormat()), Ya === "" && (Ya = i.getStyle()));
				}
				let p = d > o, m = f > s;
				if (p && !m) {
					let t = n[s + 1], r = t === void 0 ? null : za.getElementByKey(t);
					so(n, e, f, s, a.withBefore(r));
				} else m && !p && eo(t, d, o, a.element);
			})(t, r, s, i, a, n);
		}
		xu(t) && (H += ia), o.__lexicalTextContent = H, H = r + H;
	}(e, t, t.getDOMSlot(n)), r = t, Ja == null || Ja === r.__textFormat || Qa || r.setTextFormat(Ja), function(e) {
		Ya === "" || Ya === e.__textStyle || Qa || e.setTextStyle(Ya);
	}(t);
}
function fo(e, t) {
	let n = [], r = e.__first;
	for (; r !== null;) {
		let e = t.get(r);
		e === void 0 && V(101), n.push(r), r = e.__next;
	}
	return n;
}
function po(e, t) {
	let n = Wa.get(e), r = Ga.get(e);
	n !== void 0 && r !== void 0 || V(61);
	let i = Za || Ua.has(e) || Ha.has(e), a = Su(za, e);
	if (n === r && !i) {
		if (Y(n)) {
			let e = a.__lexicalTextContent;
			e !== void 0 && (H += e, Xa += e);
		} else {
			let e = n.getTextContent();
			Xa += e, H += e;
		}
		return a;
	}
	if (n !== r && i && _u(qa, Ba, Va, r, "updated"), r.updateDOM(n, a, Ra)) {
		let n = oo(e, null);
		return t === null && V(62), t.replaceChild(n, a), $a(e, null), n;
	}
	if (Y(n) && Y(r)) {
		let e = r.__indent;
		(Za || e !== n.__indent) && ro(a, e);
		let t = r.__format;
		if ((Za || t !== n.__format) && io(a, t), i && (uo(n, r, a), ll(r) || r.isInline() || lo(n, r, a)), xu(r) && (H += ia, Xa += ia), (Za || r.__dir !== n.__dir) && (ao(a, r), ll(r) && !Za)) for (let e of r.getChildren()) Y(e) && ao(Su(za, e.getKey()), e);
	} else {
		let t = r.getTextContent();
		if (X(r)) {
			let t = r.decorate(za, Ra);
			t !== null && mo(e, t);
		}
		H += t, Xa += t;
	}
	if (!Qa && ll(r) && r.__cachedText !== Xa) {
		let e = r.getWritable();
		e.__cachedText = Xa, r = e;
	}
	return a;
}
function mo(e, t) {
	let n = za._pendingDecorators, r = za._decorators;
	if (n === null) {
		if (r[e] === t) return;
		n = Zl(za);
	}
	n[e] = t;
}
function ho(e) {
	let t = e.nextSibling;
	return t !== null && t === za._blockCursorElement && (t = t.nextSibling), t;
}
function go(e, t, n, r, i, a) {
	H = "", Xa = "", Za = r === qi, za = n, Ra = n._config, Ba = n._nodes, Va = za._listeners.mutation, Ha = i, Ua = a, Wa = e._nodeMap, Ga = t._nodeMap, Qa = t._readOnly, Ka = new Map(n._keyToDOMMap);
	let o = /* @__PURE__ */ new Map();
	return qa = o, po("root", null), za = void 0, Ba = void 0, Ha = void 0, Ua = void 0, Wa = void 0, Ga = void 0, Ra = void 0, Ka = void 0, qa = void 0, o;
}
function _o(e) {
	let t = Ka.get(e);
	return t === void 0 && V(75, e), t;
}
function U(e) {
	return { type: e };
}
var vo = U("SELECTION_CHANGE_COMMAND"), yo = U("CLICK_COMMAND"), bo = U("DELETE_CHARACTER_COMMAND"), xo = U("INSERT_LINE_BREAK_COMMAND"), So = U("INSERT_PARAGRAPH_COMMAND"), Co = U("CONTROLLED_TEXT_INSERTION_COMMAND"), wo = U("PASTE_COMMAND"), To = U("REMOVE_TEXT_COMMAND"), Eo = U("DELETE_WORD_COMMAND"), Do = U("DELETE_LINE_COMMAND"), Oo = U("FORMAT_TEXT_COMMAND"), ko = U("UNDO_COMMAND"), Ao = U("REDO_COMMAND"), jo = U("KEYDOWN_COMMAND"), Mo = U("KEY_ARROW_RIGHT_COMMAND"), No = U("MOVE_TO_END"), Po = U("KEY_ARROW_LEFT_COMMAND"), Fo = U("MOVE_TO_START"), Io = U("KEY_ARROW_UP_COMMAND"), Lo = U("KEY_ARROW_DOWN_COMMAND"), Ro = U("KEY_ENTER_COMMAND"), zo = U("KEY_SPACE_COMMAND"), Bo = U("KEY_BACKSPACE_COMMAND"), Vo = U("KEY_ESCAPE_COMMAND"), Ho = U("KEY_DELETE_COMMAND"), Uo = U("KEY_TAB_COMMAND"), Wo = U("DROP_COMMAND"), Go = U("DRAGSTART_COMMAND"), Ko = U("DRAGOVER_COMMAND"), qo = U("DRAGEND_COMMAND"), Jo = U("COPY_COMMAND"), Yo = U("CUT_COMMAND"), Xo = U("SELECT_ALL_COMMAND"), Zo = U("CLEAR_EDITOR_COMMAND"), Qo = U("CLEAR_HISTORY_COMMAND"), $o = U("CAN_REDO_COMMAND"), es = U("CAN_UNDO_COMMAND"), ts = U("FOCUS_COMMAND"), ns = U("BLUR_COMMAND"), rs = U("KEY_MODIFIER_COMMAND"), is = Object.freeze({}), as = 30, os = [
	["keydown", function(e, t) {
		if (ss = e.timeStamp, cs = e.key, !t.isComposing() && !$(t, jo, e) && e.key != null) {
			if (gs && pu(e)) return nl(t, (() => {
				Es(t, _s);
			})), gs = !1, void (_s = "");
			if (function(e) {
				return Q(e, "ArrowRight", { shiftKey: "any" });
			}(e)) $(t, Mo, e);
			else if (function(e) {
				return Q(e, "ArrowRight", du);
			}(e)) $(t, No, e);
			else if (function(e) {
				return Q(e, "ArrowLeft", { shiftKey: "any" });
			}(e)) $(t, Po, e);
			else if (function(e) {
				return Q(e, "ArrowLeft", du);
			}(e)) $(t, Fo, e);
			else if (function(e) {
				return Q(e, "ArrowUp", {
					altKey: "any",
					shiftKey: "any"
				});
			}(e)) $(t, Io, e);
			else if (function(e) {
				return Q(e, "ArrowDown", {
					altKey: "any",
					shiftKey: "any"
				});
			}(e)) $(t, Lo, e);
			else if (function(e) {
				return Q(e, "Enter", {
					altKey: "any",
					ctrlKey: "any",
					metaKey: "any",
					shiftKey: !0
				});
			}(e)) ms = !0, $(t, Ro, e);
			else if (function(e) {
				return e.key === " ";
			}(e)) $(t, zo, e);
			else if (function(e) {
				return ji && Q(e, "o", { ctrlKey: !0 });
			}(e)) e.preventDefault(), ms = !0, $(t, xo, !0);
			else if (function(e) {
				return Q(e, "Enter", {
					altKey: "any",
					ctrlKey: "any",
					metaKey: "any"
				});
			}(e)) ms = !1, $(t, Ro, e);
			else if (function(e) {
				return Q(e, "Backspace", { shiftKey: "any" }) || ji && Q(e, "h", { ctrlKey: !0 });
			}(e)) pu(e) ? $(t, Bo, e) : (e.preventDefault(), $(t, bo, !0));
			else if (function(e) {
				return e.key === "Escape";
			}(e)) $(t, Vo, e);
			else if (function(e) {
				return Q(e, "Delete", {}) || ji && Q(e, "d", { ctrlKey: !0 });
			}(e)) (function(e) {
				return e.key === "Delete";
			})(e) ? $(t, Ho, e) : (e.preventDefault(), $(t, bo, !1));
			else if (function(e) {
				return Q(e, "Backspace", fu);
			}(e)) e.preventDefault(), $(t, Eo, !0);
			else if (function(e) {
				return Q(e, "Delete", fu);
			}(e)) e.preventDefault(), $(t, Eo, !1);
			else if (function(e) {
				return ji && Q(e, "Backspace", { metaKey: !0 });
			}(e)) e.preventDefault(), $(t, Do, !0);
			else if (function(e) {
				return ji && (Q(e, "Delete", { metaKey: !0 }) || Q(e, "k", { ctrlKey: !0 }));
			}(e)) e.preventDefault(), $(t, Do, !1);
			else if (function(e) {
				return Q(e, "b", du);
			}(e)) e.preventDefault(), $(t, Oo, "bold");
			else if (function(e) {
				return Q(e, "u", du);
			}(e)) e.preventDefault(), $(t, Oo, "underline");
			else if (function(e) {
				return Q(e, "i", du);
			}(e)) e.preventDefault(), $(t, Oo, "italic");
			else if (function(e) {
				return Q(e, "Tab", { shiftKey: "any" });
			}(e)) $(t, Uo, e);
			else if (function(e) {
				return Q(e, "z", du);
			}(e)) e.preventDefault(), $(t, ko, void 0);
			else if (function(e) {
				return ji ? Q(e, "z", {
					metaKey: !0,
					shiftKey: !0
				}) : Q(e, "y", { ctrlKey: !0 }) || Q(e, "z", {
					ctrlKey: !0,
					shiftKey: !0
				});
			}(e)) e.preventDefault(), $(t, Ao, void 0);
			else {
				let n = t._editorState._selection;
				n === null || K(n) ? mu(e) && (e.preventDefault(), $(t, Xo, e)) : function(e) {
					return Q(e, "c", du);
				}(e) ? (e.preventDefault(), $(t, Jo, e)) : function(e) {
					return Q(e, "x", du);
				}(e) ? (e.preventDefault(), $(t, Yo, e)) : mu(e) && (e.preventDefault(), $(t, Xo, e));
			}
			(function(e) {
				return e.ctrlKey || e.shiftKey || e.altKey || e.metaKey;
			})(e) && $(t, rs, e);
		}
	}],
	["pointerdown", function(e, t) {
		let n = e.target, r = e.pointerType;
		Bu(n) && r !== "touch" && r !== "pen" && e.button === 0 && nl(t, (() => {
			Al(n) || (ps = !0);
		}));
	}],
	["compositionstart", function(e, t) {
		nl(t, (() => {
			let n = q();
			if (K(n) && !t.isComposing()) {
				let r = n.anchor, i = n.anchor.getNode();
				Gl(r.key), (e.timeStamp < ss + as || r.type === "element" || !n.isCollapsed() || i.getFormat() !== n.format || G(i) && i.getStyle() !== n.style) && $(t, Co, aa);
			}
		}));
	}],
	["compositionend", function(e, t) {
		Mi ? hs = !0 : Fi || !Pi && !zi ? nl(t, (() => {
			Es(t, e.data);
		})) : (gs = !0, _s = e.data);
	}],
	["input", function(e, t) {
		e.stopPropagation(), nl(t, (() => {
			if (zu(e.target) && Al(e.target)) return;
			let n = q(), r = e.data, i = Ts(e);
			if (r != null && K(n) && bs(n, i, r, e.timeStamp, !1)) {
				hs &&= (Es(t, r), !1);
				let i = n.anchor.getNode(), a = Lu(ku(t));
				if (a === null) return;
				let o = n.isBackward(), s = o ? n.anchor.offset : n.focus.offset, c = o ? n.focus.offset : n.anchor.offset;
				Ni && !n.isCollapsed() && G(i) && a.anchorNode !== null && i.getTextContent().slice(0, s) + r + i.getTextContent().slice(s + c) === au(a.anchorNode) || $(t, Co, r);
				let l = r.length;
				Mi && l > 1 && e.inputType === "insertCompositionText" && !t.isComposing() && (n.anchor.offset -= l), Pi || Fi || zi || !t.isComposing() || (ss = 0, Gl(null));
			} else ou(!1, t, r === null ? void 0 : r), hs &&= (Es(t, r || void 0), !1);
			(function() {
				Hc(), Ca(J());
			})();
		}), { event: e }), us = null;
	}],
	["click", function(e, t) {
		nl(t, (() => {
			let n = q(), r = Lu(ku(t)), i = Tc();
			if (r) {
				if (K(n)) {
					let t = n.anchor, a = t.getNode();
					if (t.type === "element" && t.offset === 0 && n.isCollapsed() && !ll(a) && Z().getChildrenSize() === 1 && a.getTopLevelElementOrThrow().isEmpty() && i !== null && n.is(i)) r.removeAllRanges(), n.dirty = !0;
					else if (e.detail === 3 && !n.isCollapsed() && a !== n.focus.getNode()) {
						let e = function(e, t) {
							let n = e;
							for (; n !== Z() && n != null;) {
								if (t(n)) return n;
								n = n.getParent();
							}
							return null;
						}(a, ((e) => Y(e) && !e.isInline()));
						Y(e) && e.select(0);
					}
				} else if (e.pointerType === "touch" || e.pointerType === "pen") {
					let n = r.anchorNode;
					(zu(n) || Rl(n)) && eu(wc(i, r, t, e));
				}
			}
			$(t, yo, e);
		}));
	}],
	["cut", is],
	["copy", is],
	["dragstart", is],
	["dragover", is],
	["dragend", is],
	["paste", is],
	["focus", is],
	["blur", is],
	["drop", is]
];
Ni && os.push(["beforeinput", (e, t) => function(e, t) {
	let n = e.inputType, r = Ts(e);
	n === "deleteCompositionText" || Mi && bu(t) || n !== "insertCompositionText" && nl(t, (() => {
		let i = q();
		if (n === "deleteContentBackward") {
			if (i === null) {
				let e = Tc();
				if (!K(e)) return;
				eu(e.clone());
			}
			if (K(i)) {
				let n = i.anchor.key === i.focus.key;
				if (a = e.timeStamp, cs === "MediaLast" && a < ss + as && t.isComposing() && n) {
					if (Gl(null), ss = 0, setTimeout((() => {
						nl(t, (() => {
							Gl(null);
						}));
					}), as), K(i)) {
						let e = i.anchor.getNode();
						e.markDirty(), G(e) || V(142), ws(i, e);
					}
				} else {
					Gl(null), e.preventDefault();
					let r = i.anchor.getNode(), a = r.getTextContent(), o = r.canInsertTextAfter(), s = i.anchor.offset === 0 && i.focus.offset === a.length, c = Ri && n && !s && o;
					if (c && i.isCollapsed() && (c = !X(yu(i.anchor, !0))), !c) {
						$(t, bo, !0);
						let e = q();
						Ri && K(e) && e.isCollapsed() && (vs = e, setTimeout((() => vs = null)));
					}
				}
				return;
			}
		}
		var a;
		if (!K(i)) return;
		let o = e.data;
		us !== null && ou(!1, t, us), i.dirty && us === null || !i.isCollapsed() || ll(i.anchor.getNode()) || r === null || i.applyDOMRange(r), us = null;
		let s = i.anchor, c = i.focus, l = s.getNode(), u = c.getNode();
		if (n !== "insertText" && n !== "insertTranspose") switch (e.preventDefault(), n) {
			case "insertFromYank":
			case "insertFromDrop":
			case "insertReplacementText":
				$(t, Co, e);
				break;
			case "insertFromComposition":
				Gl(null), $(t, Co, e);
				break;
			case "insertLineBreak":
				Gl(null), $(t, xo, !1);
				break;
			case "insertParagraph":
				Gl(null), ms && !Fi ? (ms = !1, $(t, xo, !1)) : $(t, So, void 0);
				break;
			case "insertFromPaste":
			case "insertFromPasteAsQuotation":
				$(t, wo, e);
				break;
			case "deleteByComposition":
				(function(e, t) {
					return e !== t || Y(e) || Y(t) || !Il(e) || !Il(t);
				})(l, u) && $(t, To, e);
				break;
			case "deleteByDrag":
			case "deleteByCut":
				$(t, To, e);
				break;
			case "deleteContent":
				$(t, bo, !1);
				break;
			case "deleteWordBackward":
				$(t, Eo, !0);
				break;
			case "deleteWordForward":
				$(t, Eo, !1);
				break;
			case "deleteHardLineBackward":
			case "deleteSoftLineBackward":
				$(t, Do, !0);
				break;
			case "deleteContentForward":
			case "deleteHardLineForward":
			case "deleteSoftLineForward":
				$(t, Do, !1);
				break;
			case "formatStrikeThrough":
				$(t, Oo, "strikethrough");
				break;
			case "formatBold":
				$(t, Oo, "bold");
				break;
			case "formatItalic":
				$(t, Oo, "italic");
				break;
			case "formatUnderline":
				$(t, Oo, "underline");
				break;
			case "historyUndo":
				$(t, ko, void 0);
				break;
			case "historyRedo": $(t, Ao, void 0);
		}
		else {
			if (o === "\n") e.preventDefault(), $(t, xo, !1);
			else if (o === ia) e.preventDefault(), $(t, So, void 0);
			else if (o == null && e.dataTransfer) {
				let t = e.dataTransfer.getData("text/plain");
				e.preventDefault(), i.insertRawText(t);
			} else o != null && bs(i, r, o, e.timeStamp, !0) ? (e.preventDefault(), $(t, Co, o)) : us = o;
			ls = e.timeStamp;
		}
	}));
}(e, t)]);
var ss = 0, cs = null, ls = 0, us = null, ds = /* @__PURE__ */ new WeakMap(), fs = !1, ps = !1, ms = !1, hs = !1, gs = !1, _s = "", vs = null, ys = [
	0,
	"",
	0,
	"root",
	0
];
function bs(e, t, n, r, i) {
	let a = e.anchor, o = e.focus, s = a.getNode(), c = J(), l = Lu(ku(c)), u = l === null ? null : l.anchorNode, d = a.key, f = c.getElementByKey(d), p = n.length;
	return d !== o.key || !G(s) || (!i && (!Ni || ls < r + 50) || s.isDirty() && p < 2 || nu(n)) && a.offset !== o.offset && !s.isComposing() || Ll(s) || s.isDirty() && p > 1 || (i || !Ni) && f !== null && !s.isComposing() && u !== Bl(f) || l !== null && t !== null && (!t.collapsed || t.startContainer !== l.anchorNode || t.startOffset !== l.anchorOffset) || s.getFormat() !== e.format || s.getStyle() !== e.style || function(e, t) {
		if (t.isSegmented()) return !0;
		if (!e.isCollapsed()) return !1;
		let n = e.anchor.offset, r = t.getParentOrThrow(), i = Il(t);
		return n === 0 ? !t.canInsertTextBefore() || !r.canInsertTextBefore() && !t.isComposing() || i || function(e) {
			let t = e.getPreviousSibling();
			return (G(t) || Y(t) && t.isInline()) && !t.canInsertTextAfter();
		}(t) : n === t.getTextContentSize() && (!t.canInsertTextAfter() || !r.canInsertTextAfter() && !t.isComposing() || i);
	}(e, s);
}
function xs(e, t) {
	return Rl(e) && e.nodeValue !== null && t !== 0 && t !== e.nodeValue.length;
}
function Ss(e, t, n) {
	let { anchorNode: r, anchorOffset: i, focusNode: a, focusOffset: o } = e;
	fs && (fs = !1, xs(r, i) && xs(a, o) && !vs) || nl(t, (() => {
		if (!n) return void eu(null);
		if (!Ml(t, r, a)) return;
		let s = q();
		if (vs && K(s) && s.isCollapsed()) {
			let e = s.anchor, t = vs.anchor;
			(e.key === t.key && e.offset === t.offset + 1 || e.offset === 1 && t.getNode().is(e.getNode().getPreviousSibling())) && (s = vs.clone(), eu(s));
		}
		if (vs = null, K(s)) {
			let n = s.anchor, r = n.getNode();
			if (s.isCollapsed()) {
				e.type === "Range" && e.anchorNode === e.focusNode && (s.dirty = !0);
				let i = ku(t).event, a = i ? i.timeStamp : performance.now(), [o, c, l, u, d] = ys, f = Z(), p = !1 === t.isComposing() && f.getTextContent() === "";
				if (a < d + 200 && n.offset === l && n.key === u) Cs(s, o, c);
				else if (n.type === "text") G(r) || V(141), ws(s, r);
				else if (n.type === "element" && !p) {
					Y(r) || V(259);
					let e = n.getNode();
					e.isEmpty() ? function(e, t) {
						Cs(e, t.getTextFormat(), t.getTextStyle());
					}(s, e) : Cs(s, 0, "");
				}
			} else {
				let e = n.key, t = s.focus.key, r = s.getNodes(), a = r.length, c = s.isBackward(), l = c ? o : i, u = c ? i : o, d = c ? t : e, f = c ? e : t, p = 2047, m = !1;
				for (let e = 0; e < a; e++) {
					let t = r[e], n = t.getTextContentSize();
					if (G(t) && n !== 0 && !(e === 0 && t.__key === d && l === n || e === a - 1 && t.__key === f && u === 0) && (m = !0, p &= t.getFormat(), p === 0)) break;
				}
				s.format = m ? p : 0;
			}
		}
		$(t, vo, void 0);
	}));
}
function Cs(e, t, n) {
	e.format === t && e.style === n || (e.format = t, e.style = n, e.dirty = !0);
}
function ws(e, t) {
	Cs(e, t.getFormat(), t.getStyle());
}
function Ts(e) {
	if (!e.getTargetRanges) return null;
	let t = e.getTargetRanges();
	return t.length === 0 ? null : t[0];
}
function Es(e, t) {
	let n = e._compositionKey;
	if (Gl(null), n !== null && t != null) {
		if (t === "") {
			let t = ql(n), r = Bl(e.getElementByKey(n));
			r !== null && r.nodeValue !== null && G(t) && su(t, r.nodeValue, null, null, !0);
			return;
		}
		if (t[t.length - 1] === "\n") {
			let t = q();
			if (K(t)) {
				let n = t.focus;
				t.anchor.set(n.key, n.offset, n.type), $(e, Ro, null);
				return;
			}
		}
	}
	ou(!0, e, t);
}
function Ds(e) {
	let t = e.__lexicalEventHandles;
	return t === void 0 && (t = [], e.__lexicalEventHandles = t), t;
}
var Os = /* @__PURE__ */ new Map();
function ks(e) {
	let t = Ru(e.target);
	if (t === null) return;
	let n = Pl(t.anchorNode);
	if (n === null) return;
	ps && (ps = !1, nl(n, (() => {
		let r = Tc(), i = t.anchorNode;
		(zu(i) || Rl(i)) && eu(wc(r, t, n, e));
	})));
	let r = ru(n), i = r[r.length - 1], a = i._key, o = Os.get(a), s = o || i;
	s !== n && Ss(t, s, !1), Ss(t, n, !0), n === i ? o && Os.delete(a) : Os.set(a, n);
}
function As(e) {
	e._lexicalHandled = !0;
}
function js(e) {
	return !0 === e._lexicalHandled;
}
function Ms(e) {
	let t = e.ownerDocument, n = ds.get(t);
	if (n === void 0) return;
	let r = n - 1;
	r >= 0 || V(164), ds.set(t, r), r === 0 && t.removeEventListener("selectionchange", ks);
	let i = Fl(e);
	Nl(i) ? (function(e) {
		if (e._parentEditor !== null) {
			let t = ru(e), n = t[t.length - 1]._key;
			Os.get(n) === e && Os.delete(n);
		} else Os.delete(e._key);
	}(i), e.__lexicalEditor = null) : i && V(198);
	let a = Ds(e);
	for (let e = 0; e < a.length; e++) a[e]();
	e.__lexicalEventHandles = [];
}
function Ns(e, t, n) {
	Hc();
	let r = e.__key, i = e.getParent();
	if (i === null) return;
	let a = function(e) {
		let t = q();
		if (!K(t) || !Y(e)) return t;
		let { anchor: n, focus: r } = t, i = n.getNode(), a = r.getNode();
		return Du(i, e) && n.set(e.__key, 0, "element"), Du(a, e) && r.set(e.__key, 0, "element"), t;
	}(e), o = !1;
	if (K(a) && t) {
		let t = a.anchor, n = a.focus;
		t.key === r && (Oc(t, e, i, e.getPreviousSibling(), e.getNextSibling()), o = !0), n.key === r && (Oc(n, e, i, e.getPreviousSibling(), e.getNextSibling()), o = !0);
	} else dc(a) && t && e.isSelected() && e.selectPrevious();
	if (K(a) && t && !o) {
		let t = e.getIndexWithinParent();
		Ul(e), Ec(a, i, t, -1);
	} else Ul(e);
	n || ju(i) || i.canBeEmpty() || !i.isEmpty() || Ns(i, t), t && a && ll(i) && i.isEmpty() && i.selectEnd();
}
var Ps = class {
	static getType() {
		let { ownNodeType: e } = ed(this);
		return e === void 0 && V(64, this.name), e;
	}
	static clone(e) {
		V(65, this.name);
	}
	$config() {
		return {};
	}
	config(e, t) {
		let n = t.extends || Object.getPrototypeOf(this.constructor);
		return Object.assign(t, {
			extends: n,
			type: e
		}), { [e]: t };
	}
	afterCloneFrom(e) {
		this.__key === e.__key ? (this.__parent = e.__parent, this.__next = e.__next, this.__prev = e.__prev, this.__state = e.__state) : e.__state && (this.__state = e.__state.getWritable(this));
	}
	constructor(e) {
		this.__type = this.constructor.getType(), this.__parent = null, this.__prev = null, this.__next = null, Object.defineProperty(this, "__state", {
			configurable: !0,
			enumerable: !1,
			value: void 0,
			writable: !0
		}), Hl(this, e);
	}
	getType() {
		return this.__type;
	}
	isInline() {
		V(137, this.constructor.name);
	}
	isAttached() {
		let e = this.__key;
		for (; e !== null;) {
			if (e === "root") return !0;
			let t = ql(e);
			if (t === null) break;
			e = t.__parent;
		}
		return !1;
	}
	isSelected(e) {
		let t = e || q();
		if (t == null) return !1;
		let n = t.getNodes().some(((e) => e.__key === this.__key));
		if (G(this)) return n;
		if (K(t) && t.anchor.type === "element" && t.focus.type === "element") {
			if (t.isCollapsed()) return !1;
			let e = this.getParent();
			if (X(this) && this.isInline() && e) {
				let n = t.isBackward() ? t.focus : t.anchor;
				if (e.is(n.getNode()) && n.offset === e.getChildrenSize() && this.is(e.getLastChild())) return !1;
			}
		}
		return n;
	}
	getKey() {
		return this.__key;
	}
	getIndexWithinParent() {
		let e = this.getParent();
		if (e === null) return -1;
		let t = e.getFirstChild(), n = 0;
		for (; t !== null;) {
			if (this.is(t)) return n;
			n++, t = t.getNextSibling();
		}
		return -1;
	}
	getParent() {
		let e = this.getLatest().__parent;
		return e === null ? null : ql(e);
	}
	getParentOrThrow() {
		let e = this.getParent();
		return e === null && V(66, this.__key), e;
	}
	getTopLevelElement() {
		let e = this;
		for (; e !== null;) {
			let t = e.getParent();
			if (ju(t)) return Y(e) || e === this && X(e) || V(194), e;
			e = t;
		}
		return null;
	}
	getTopLevelElementOrThrow() {
		let e = this.getTopLevelElement();
		return e === null && V(67, this.__key), e;
	}
	getParents() {
		let e = [], t = this.getParent();
		for (; t !== null;) e.push(t), t = t.getParent();
		return e;
	}
	getParentKeys() {
		let e = [], t = this.getParent();
		for (; t !== null;) e.push(t.__key), t = t.getParent();
		return e;
	}
	getPreviousSibling() {
		let e = this.getLatest().__prev;
		return e === null ? null : ql(e);
	}
	getPreviousSiblings() {
		let e = [], t = this.getParent();
		if (t === null) return e;
		let n = t.getFirstChild();
		for (; n !== null && !n.is(this);) e.push(n), n = n.getNextSibling();
		return e;
	}
	getNextSibling() {
		let e = this.getLatest().__next;
		return e === null ? null : ql(e);
	}
	getNextSiblings() {
		let e = [], t = this.getNextSibling();
		for (; t !== null;) e.push(t), t = t.getNextSibling();
		return e;
	}
	getCommonAncestor(e) {
		let t = Y(this) ? this : this.getParent(), n = Y(e) ? e : e.getParent(), r = t && n ? Md(t, n) : null;
		return r ? r.commonAncestor : null;
	}
	is(e) {
		return e != null && this.__key === e.__key;
	}
	isBefore(e) {
		let t = Md(this, e);
		return t !== null && (t.type === "descendant" || (t.type === "branch" ? kd(t) === -1 : (t.type !== "same" && t.type !== "ancestor" && V(279), !1)));
	}
	isParentOf(e) {
		let t = Md(this, e);
		return t !== null && t.type === "ancestor";
	}
	getNodesBetween(e) {
		let t = this.isBefore(e), n = [], r = /* @__PURE__ */ new Set(), i = this;
		for (; i !== null;) {
			let a = i.__key;
			if (r.has(a) || (r.add(a), n.push(i)), i === e) break;
			let o = Y(i) ? t ? i.getFirstChild() : i.getLastChild() : null;
			if (o !== null) {
				i = o;
				continue;
			}
			let s = t ? i.getNextSibling() : i.getPreviousSibling();
			if (s !== null) {
				i = s;
				continue;
			}
			let c = i.getParentOrThrow();
			if (r.has(c.__key) || n.push(c), c === e) break;
			let l = null, u = c;
			do {
				if (u === null && V(68), l = t ? u.getNextSibling() : u.getPreviousSibling(), u = u.getParent(), u === null) break;
				l !== null || r.has(u.__key) || n.push(u);
			} while (l === null);
			i = l;
		}
		return t || n.reverse(), n;
	}
	isDirty() {
		let e = J()._dirtyLeaves;
		return e !== null && e.has(this.__key);
	}
	getLatest() {
		let e = ql(this.__key);
		return e === null && V(113), e;
	}
	getWritable() {
		Hc();
		let e = Wc(), t = J(), n = e._nodeMap, r = this.__key, i = this.getLatest(), a = t._cloneNotNeeded, o = q();
		if (o !== null && o.setCachedNodes(null), a.has(r)) return Wl(i), i;
		let s = Xu(i);
		return a.add(r), Wl(s), n.set(r, s), s;
	}
	getTextContent() {
		return "";
	}
	getTextContentSize() {
		return this.getTextContent().length;
	}
	createDOM(e, t) {
		V(70);
	}
	updateDOM(e, t, n) {
		V(71);
	}
	exportDOM(e) {
		return { element: this.createDOM(e._config, e) };
	}
	exportJSON() {
		let e = this.__state ? this.__state.toJSON() : void 0;
		return {
			type: this.__type,
			version: 1,
			...e
		};
	}
	static importJSON(e) {
		V(18, this.name);
	}
	updateFromJSON(e) {
		return function(e, t) {
			let n = e.getWritable(), r = t.$, i = r;
			for (let e of Oa(n).flatKeys) e in t && (i !== void 0 && i !== r || (i = { ...r }), i[e] = t[e]);
			return (n.__state || i) && Da(e).updateFromJSON(i), n;
		}(this, e);
	}
	static transform() {
		return null;
	}
	remove(e) {
		Ns(this, !0, e);
	}
	replace(e, t) {
		Hc();
		let n = q();
		n !== null && (n = n.clone()), Nu(this, e);
		let r = this.getLatest(), i = this.__key, a = e.__key, o = e.getWritable(), s = this.getParentOrThrow().getWritable(), c = s.__size;
		Ul(o);
		let l = r.getPreviousSibling(), u = r.getNextSibling(), d = r.__prev, f = r.__next, p = r.__parent;
		if (Ns(r, !1, !0), l === null ? s.__first = a : l.getWritable().__next = a, o.__prev = d, u === null ? s.__last = a : u.getWritable().__prev = a, o.__next = f, o.__parent = p, s.__size = c, t && (Y(this) && Y(o) || V(139), this.getChildren().forEach(((e) => {
			o.append(e);
		}))), K(n)) {
			eu(n);
			let e = n.anchor, t = n.focus;
			e.key === i && cc(e, o), t.key === i && cc(t, o);
		}
		return Kl() === i && Gl(a), o;
	}
	insertAfter(e, t = !0) {
		Hc(), Nu(this, e);
		let n = this.getWritable(), r = e.getWritable(), i = r.getParent(), a = q(), o = !1, s = !1;
		if (i !== null) {
			let t = e.getIndexWithinParent();
			if (Ul(r), K(a)) {
				let e = i.__key, n = a.anchor, r = a.focus;
				o = n.type === "element" && n.key === e && n.offset === t + 1, s = r.type === "element" && r.key === e && r.offset === t + 1;
			}
		}
		let c = this.getNextSibling(), l = this.getParentOrThrow().getWritable(), u = r.__key, d = n.__next;
		if (c === null ? l.__last = u : c.getWritable().__prev = u, l.__size++, n.__next = u, r.__next = d, r.__prev = n.__key, r.__parent = n.__parent, t && K(a)) {
			let e = this.getIndexWithinParent();
			Ec(a, l, e + 1);
			let t = l.__key;
			o && a.anchor.set(t, e + 2, "element"), s && a.focus.set(t, e + 2, "element");
		}
		return e;
	}
	insertBefore(e, t = !0) {
		Hc(), Nu(this, e);
		let n = this.getWritable(), r = e.getWritable(), i = r.__key;
		Ul(r);
		let a = this.getPreviousSibling(), o = this.getParentOrThrow().getWritable(), s = n.__prev, c = this.getIndexWithinParent();
		a === null ? o.__first = i : a.getWritable().__next = i, o.__size++, n.__prev = i, r.__prev = s, r.__next = n.__key, r.__parent = n.__parent;
		let l = q();
		return t && K(l) && Ec(l, this.getParentOrThrow(), c), e;
	}
	isParentRequired() {
		return !1;
	}
	createParentElementNode() {
		return yl();
	}
	selectStart() {
		return this.selectPrevious();
	}
	selectEnd() {
		return this.selectNext(0, 0);
	}
	selectPrevious(e, t) {
		Hc();
		let n = this.getPreviousSibling(), r = this.getParentOrThrow();
		if (n === null) return r.select(0, 0);
		if (Y(n)) return n.select();
		if (!G(n)) {
			let e = n.getIndexWithinParent() + 1;
			return r.select(e, e);
		}
		return n.select(e, t);
	}
	selectNext(e, t) {
		Hc();
		let n = this.getNextSibling(), r = this.getParentOrThrow();
		if (n === null) return r.select();
		if (Y(n)) return n.select(0, 0);
		if (!G(n)) {
			let e = n.getIndexWithinParent();
			return r.select(e, e);
		}
		return n.select(e, t);
	}
	markDirty() {
		this.getWritable();
	}
	reconcileObservedMutation(e, t) {
		this.markDirty();
	}
}, Fs = class e extends Ps {
	static getType() {
		return "linebreak";
	}
	static clone(t) {
		return new e(t.__key);
	}
	constructor(e) {
		super(e);
	}
	getTextContent() {
		return "\n";
	}
	createDOM() {
		return document.createElement("br");
	}
	updateDOM() {
		return !1;
	}
	isInline() {
		return !0;
	}
	static importDOM() {
		return { br: (e) => function(e) {
			let t = e.parentElement;
			if (t !== null && Uu(t)) {
				let n = t.firstChild;
				if (n === e || n.nextSibling === e && zs(n)) {
					let n = t.lastChild;
					if (n === e || n.previousSibling === e && zs(n)) return !0;
				}
			}
			return !1;
		}(e) || function(e) {
			let t = e.parentElement;
			if (t !== null && Uu(t)) {
				let n = t.firstChild;
				if (n === e || n.nextSibling === e && zs(n)) return !1;
				let r = t.lastChild;
				if (r === e || r.previousSibling === e && zs(r)) return !0;
			}
			return !1;
		}(e) ? null : {
			conversion: Is,
			priority: 0
		} };
	}
	static importJSON(e) {
		return Ls().updateFromJSON(e);
	}
};
function Is(e) {
	return { node: Ls() };
}
function Ls() {
	return Mu(new Fs());
}
function Rs(e) {
	return e instanceof Fs;
}
function zs(e) {
	return Rl(e) && /^( |\t|\r?\n)+$/.test(e.textContent || "");
}
function Bs(e, t) {
	return 16 & t ? "code" : t & 128 ? "mark" : 32 & t ? "sub" : 64 & t ? "sup" : null;
}
function Vs(e, t) {
	return 1 & t ? "strong" : 2 & t ? "em" : "span";
}
function Hs(e, t, n, r, i) {
	let a = r.classList, o = gu(i, "base");
	o !== void 0 && a.add(...o), o = gu(i, "underlineStrikethrough");
	let s = !1, c = t & 8 && t & 4;
	o !== void 0 && (n & 8 && n & 4 ? (s = !0, c || a.add(...o)) : c && a.remove(...o));
	for (let e in oa) {
		let r = oa[e];
		if (o = gu(i, e), o !== void 0) {
			if (n & r) {
				if (s && (e === "underline" || e === "strikethrough")) {
					t & r && a.remove(...o);
					continue;
				}
				t & r && (!c || e !== "underline") && e !== "strikethrough" || a.add(...o);
			} else t & r && a.remove(...o);
		}
	}
}
function Us(e, t, n) {
	let r = t.firstChild, i = n.isComposing(), a = e + (i ? ra : "");
	if (r == null) t.textContent = a;
	else {
		let e = r.nodeValue;
		if (e !== a) {
			if (i || Mi) {
				let [t, n, i] = function(e, t) {
					let n = e.length, r = t.length, i = 0, a = 0;
					for (; i < n && i < r && e[i] === t[i];) i++;
					for (; a + i < n && a + i < r && e[n - a - 1] === t[r - a - 1];) a++;
					return [
						i,
						n - i - a,
						t.slice(i, r - a)
					];
				}(e, a);
				n !== 0 && r.deleteData(t, n), r.insertData(t, i);
			} else r.nodeValue = a;
		}
	}
}
function Ws(e, t, n, r, i, a) {
	Us(i, e, t);
	let o = a.theme.text;
	o !== void 0 && Hs(0, 0, r, e, o);
}
function Gs(e, t) {
	let n = document.createElement(t);
	return n.appendChild(e), n;
}
var Ks = class e extends Ps {
	static getType() {
		return "text";
	}
	static clone(t) {
		return new e(t.__text, t.__key);
	}
	afterCloneFrom(e) {
		super.afterCloneFrom(e), this.__text = e.__text, this.__format = e.__format, this.__style = e.__style, this.__mode = e.__mode, this.__detail = e.__detail;
	}
	constructor(e = "", t) {
		super(t), this.__text = e, this.__format = 0, this.__style = "", this.__mode = 0, this.__detail = 0;
	}
	getFormat() {
		return this.getLatest().__format;
	}
	getDetail() {
		return this.getLatest().__detail;
	}
	getMode() {
		return da[this.getLatest().__mode];
	}
	getStyle() {
		return this.getLatest().__style;
	}
	isToken() {
		return this.getLatest().__mode === 1;
	}
	isComposing() {
		return this.__key === Kl();
	}
	isSegmented() {
		return this.getLatest().__mode === 2;
	}
	isDirectionless() {
		return !!(1 & this.getLatest().__detail);
	}
	isUnmergeable() {
		return !!(2 & this.getLatest().__detail);
	}
	hasFormat(e) {
		let t = oa[e];
		return !!(this.getFormat() & t);
	}
	isSimpleText() {
		return this.__type === "text" && this.__mode === 0;
	}
	getTextContent() {
		return this.getLatest().__text;
	}
	getFormatFlags(e, t) {
		return Vl(this.getLatest().__format, e, t);
	}
	canHaveFormat() {
		return !0;
	}
	isInline() {
		return !0;
	}
	createDOM(e, t) {
		let n = this.__format, r = Bs(0, n), i = Vs(0, n), a = r === null ? i : r, o = document.createElement(a), s = o;
		this.hasFormat("code") && o.setAttribute("spellcheck", "false"), r !== null && (s = document.createElement(i), o.appendChild(s)), Ws(s, this, 0, n, this.__text, e);
		let c = this.__style;
		return c !== "" && (o.style.cssText = c), o;
	}
	updateDOM(e, t, n) {
		let r = this.__text, i = e.__format, a = this.__format, o = Bs(0, i), s = Bs(0, a), c = Vs(0, i), l = Vs(0, a);
		if ((o === null ? c : o) !== (s === null ? l : s)) return !0;
		if (o === s && c !== l) {
			let e = t.firstChild;
			e ?? V(48);
			let i = document.createElement(l);
			return Ws(i, this, 0, a, r, n), t.replaceChild(i, e), !1;
		}
		let u = t;
		s !== null && o !== null && (u = t.firstChild, u ?? V(49)), Us(r, u, this);
		let d = n.theme.text;
		d !== void 0 && i !== a && Hs(0, i, a, u, d);
		let f = e.__style, p = this.__style;
		return f !== p && (t.style.cssText = p), !1;
	}
	static importDOM() {
		return {
			"#text": () => ({
				conversion: Zs,
				priority: 0
			}),
			b: () => ({
				conversion: Js,
				priority: 0
			}),
			code: () => ({
				conversion: ec,
				priority: 0
			}),
			em: () => ({
				conversion: ec,
				priority: 0
			}),
			i: () => ({
				conversion: ec,
				priority: 0
			}),
			mark: () => ({
				conversion: ec,
				priority: 0
			}),
			s: () => ({
				conversion: ec,
				priority: 0
			}),
			span: () => ({
				conversion: qs,
				priority: 0
			}),
			strong: () => ({
				conversion: ec,
				priority: 0
			}),
			sub: () => ({
				conversion: ec,
				priority: 0
			}),
			sup: () => ({
				conversion: ec,
				priority: 0
			}),
			u: () => ({
				conversion: ec,
				priority: 0
			})
		};
	}
	static importJSON(e) {
		return W().updateFromJSON(e);
	}
	updateFromJSON(e) {
		return super.updateFromJSON(e).setTextContent(e.text).setFormat(e.format).setDetail(e.detail).setMode(e.mode).setStyle(e.style);
	}
	exportDOM(e) {
		let { element: t } = super.exportDOM(e);
		return zu(t) || V(132), t.style.whiteSpace = "pre-wrap", this.hasFormat("lowercase") ? t.style.textTransform = "lowercase" : this.hasFormat("uppercase") ? t.style.textTransform = "uppercase" : this.hasFormat("capitalize") && (t.style.textTransform = "capitalize"), this.hasFormat("bold") && (t = Gs(t, "b")), this.hasFormat("italic") && (t = Gs(t, "i")), this.hasFormat("strikethrough") && (t = Gs(t, "s")), this.hasFormat("underline") && (t = Gs(t, "u")), { element: t };
	}
	exportJSON() {
		return {
			detail: this.getDetail(),
			format: this.getFormat(),
			mode: this.getMode(),
			style: this.getStyle(),
			text: this.getTextContent(),
			...super.exportJSON()
		};
	}
	selectionTransform(e, t) {}
	setFormat(e) {
		let t = this.getWritable();
		return t.__format = typeof e == "string" ? oa[e] : e, t;
	}
	setDetail(e) {
		let t = this.getWritable();
		return t.__detail = typeof e == "string" ? sa[e] : e, t;
	}
	setStyle(e) {
		let t = this.getWritable();
		return t.__style = e, t;
	}
	toggleFormat(e) {
		let t = Vl(this.getFormat(), e, null);
		return this.setFormat(t);
	}
	toggleDirectionless() {
		let e = this.getWritable();
		return e.__detail ^= 1, e;
	}
	toggleUnmergeable() {
		let e = this.getWritable();
		return e.__detail ^= 2, e;
	}
	setMode(e) {
		let t = ua[e];
		if (this.__mode === t) return this;
		let n = this.getWritable();
		return n.__mode = t, n;
	}
	setTextContent(e) {
		if (this.__text === e) return this;
		let t = this.getWritable();
		return t.__text = e, t;
	}
	select(e, t) {
		Hc();
		let n = e, r = t, i = q(), a = this.getTextContent(), o = this.__key;
		if (typeof a == "string") {
			let e = a.length;
			n === void 0 && (n = e), r === void 0 && (r = e);
		} else n = 0, r = 0;
		if (!K(i)) return xc(o, n, o, r, "text", "text");
		{
			let e = Kl();
			e !== i.anchor.key && e !== i.focus.key || Gl(o), i.setTextNodeRange(this, n, this, r);
		}
		return i;
	}
	selectStart() {
		return this.select(0, 0);
	}
	selectEnd() {
		let e = this.getTextContentSize();
		return this.select(e, e);
	}
	spliceText(e, t, n, r) {
		let i = this.getWritable(), a = i.__text, o = n.length, s = e;
		s < 0 && (s = o + s, s < 0 && (s = 0));
		let c = q();
		if (r && K(c)) {
			let t = e + o;
			c.setTextNodeRange(i, t, i, t);
		}
		return i.__text = a.slice(0, s) + n + a.slice(s + t), i;
	}
	canInsertTextBefore() {
		return !0;
	}
	canInsertTextAfter() {
		return !0;
	}
	splitText(...e) {
		Hc();
		let t = this.getLatest(), n = t.getTextContent();
		if (n === "") return [];
		let r = t.__key, i = Kl(), a = n.length;
		e.sort(((e, t) => e - t)), e.push(a);
		let o = [], s = e.length;
		for (let t = 0, r = 0; t < a && r <= s; r++) {
			let i = e[r];
			i > t && (o.push(n.slice(t, i)), t = i);
		}
		let c = o.length;
		if (c === 1) return [t];
		let l = o[0], u = t.getParent(), d, f = t.getFormat(), p = t.getStyle(), m = t.__detail, h = !1, g = null, _ = null, v = q();
		if (K(v)) {
			let [e, t] = v.isBackward() ? [v.focus, v.anchor] : [v.anchor, v.focus];
			e.type === "text" && e.key === r && (g = e), t.type === "text" && t.key === r && (_ = t);
		}
		t.isSegmented() ? (d = W(l), d.__format = f, d.__style = p, d.__detail = m, d.__state = Ma(t, d), h = !0) : d = t.setTextContent(l);
		let y = [d];
		for (let e = 1; e < c; e++) {
			let n = W(o[e]);
			n.__format = f, n.__style = p, n.__detail = m, n.__state = Ma(t, n);
			let a = n.__key;
			i === r && Gl(a), y.push(n);
		}
		let b = g ? g.offset : null, x = _ ? _.offset : null, S = 0;
		for (let e of y) {
			if (!g && !_) break;
			let t = S + e.getTextContentSize();
			if (g !== null && b !== null && b <= t && b >= S && (g.set(e.getKey(), b - S, "text"), b < t && (g = null)), _ !== null && x !== null && x <= t && x >= S) {
				_.set(e.getKey(), x - S, "text");
				break;
			}
			S = t;
		}
		if (u !== null) {
			(function(e) {
				let t = e.getPreviousSibling(), n = e.getNextSibling();
				t !== null && Wl(t), n !== null && Wl(n);
			})(this);
			let e = u.getWritable(), t = this.getIndexWithinParent();
			h ? (e.splice(t, 0, y), this.remove()) : e.splice(t, 1, y), K(v) && Ec(v, u, t, c - 1);
		}
		return y;
	}
	mergeWithSibling(e) {
		let t = e === this.getPreviousSibling();
		t || e === this.getNextSibling() || V(50);
		let n = this.__key, r = e.__key, i = this.__text, a = i.length;
		Kl() === r && Gl(n);
		let o = q();
		if (K(o)) {
			let i = o.anchor, s = o.focus;
			i !== null && i.key === r && kc(i, t, n, e, a), s !== null && s.key === r && kc(s, t, n, e, a);
		}
		let s = e.__text, c = t ? s + i : i + s;
		this.setTextContent(c);
		let l = this.getWritable();
		return e.remove(), l;
	}
	isTextEntity() {
		return !1;
	}
};
function qs(e) {
	return {
		forChild: tc(e.style),
		node: null
	};
}
function Js(e) {
	let t = e, n = t.style.fontWeight === "normal";
	return {
		forChild: tc(t.style, n ? void 0 : "bold"),
		node: null
	};
}
var Ys = /* @__PURE__ */ new WeakMap();
function Xs(e) {
	if (!zu(e)) return !1;
	if (e.nodeName === "PRE") return !0;
	let t = e.style.whiteSpace;
	return typeof t == "string" && t.startsWith("pre");
}
function Zs(e) {
	let t = e;
	e.parentElement === null && V(129);
	let n = t.textContent || "";
	if (function(e) {
		let t, n = e.parentNode, r = [e];
		for (; n !== null && (t = Ys.get(n)) === void 0 && !Xs(n);) r.push(n), n = n.parentNode;
		let i = t === void 0 ? n : t;
		for (let e = 0; e < r.length; e++) Ys.set(r[e], i);
		return i;
	}(t) !== null) {
		let e = n.split(/(\r?\n|\t)/), t = [], r = e.length;
		for (let n = 0; n < r; n++) {
			let r = e[n];
			r === "\n" || r === "\r\n" ? t.push(Ls()) : r === "	" ? t.push(rc()) : r !== "" && t.push(W(r));
		}
		return { node: t };
	}
	if (n = n.replace(/\r/g, "").replace(/[ \t\n]+/g, " "), n === "") return { node: null };
	if (n[0] === " ") {
		let e = t, r = !0;
		for (; e !== null && (e = Qs(e, !1)) !== null;) {
			let t = e.textContent || "";
			if (t.length > 0) {
				/[ \t\n]$/.test(t) && (n = n.slice(1)), r = !1;
				break;
			}
		}
		r && (n = n.slice(1));
	}
	if (n[n.length - 1] === " ") {
		let e = t, r = !0;
		for (; e !== null && (e = Qs(e, !0)) !== null;) if ((e.textContent || "").replace(/^( |\t|\r?\n)+/, "").length > 0) {
			r = !1;
			break;
		}
		r && (n = n.slice(0, n.length - 1));
	}
	return n === "" ? { node: null } : { node: W(n) };
}
function Qs(e, t) {
	let n = e;
	for (;;) {
		let e;
		for (; (e = t ? n.nextSibling : n.previousSibling) === null;) {
			let e = n.parentElement;
			if (e === null) return null;
			n = e;
		}
		if (n = e, zu(n)) {
			let e = n.style.display;
			if (e === "" && !Hu(n) || e !== "" && !e.startsWith("inline")) return null;
		}
		let r = n;
		for (; (r = t ? n.firstChild : n.lastChild) !== null;) n = r;
		if (Rl(n)) return n;
		if (n.nodeName === "BR") return null;
	}
}
var $s = {
	code: "code",
	em: "italic",
	i: "italic",
	mark: "highlight",
	s: "strikethrough",
	strong: "bold",
	sub: "subscript",
	sup: "superscript",
	u: "underline"
};
function ec(e) {
	let t = $s[e.nodeName.toLowerCase()];
	return t === void 0 ? { node: null } : {
		forChild: tc(e.style, t),
		node: null
	};
}
function W(e = "") {
	return Mu(new Ks(e));
}
function G(e) {
	return e instanceof Ks;
}
function tc(e, t) {
	let n = e.fontWeight, r = e.textDecoration.split(" "), i = n === "700" || n === "bold", a = r.includes("line-through"), o = e.fontStyle === "italic", s = r.includes("underline"), c = e.verticalAlign;
	return (e) => (G(e) && (i && !e.hasFormat("bold") && e.toggleFormat("bold"), a && !e.hasFormat("strikethrough") && e.toggleFormat("strikethrough"), o && !e.hasFormat("italic") && e.toggleFormat("italic"), s && !e.hasFormat("underline") && e.toggleFormat("underline"), c !== "sub" || e.hasFormat("subscript") || e.toggleFormat("subscript"), c !== "super" || e.hasFormat("superscript") || e.toggleFormat("superscript"), t && !e.hasFormat(t) && e.toggleFormat(t)), e);
}
var nc = class e extends Ks {
	static getType() {
		return "tab";
	}
	static clone(t) {
		return new e(t.__key);
	}
	constructor(e) {
		super("	", e), this.__detail = 2;
	}
	static importDOM() {
		return null;
	}
	createDOM(e) {
		let t = super.createDOM(e), n = gu(e.theme, "tab");
		return n !== void 0 && t.classList.add(...n), t;
	}
	static importJSON(e) {
		return rc().updateFromJSON(e);
	}
	setTextContent(e) {
		return e !== "	" && e !== "" && V(126), super.setTextContent("	");
	}
	spliceText(e, t, n, r) {
		return n === "" && t === 0 || n === "	" && t === 1 || V(286), this;
	}
	setDetail(e) {
		return e !== 2 && V(127), this;
	}
	setMode(e) {
		return e !== "normal" && V(128), this;
	}
	canInsertTextBefore() {
		return !1;
	}
	canInsertTextAfter() {
		return !1;
	}
};
function rc() {
	return Mu(new nc());
}
function ic(e) {
	return e instanceof nc;
}
var ac = class {
	constructor(e, t, n) {
		this._selection = null, this.key = e, this.offset = t, this.type = n;
	}
	is(e) {
		return this.key === e.key && this.offset === e.offset && this.type === e.type;
	}
	isBefore(e) {
		return this.key === e.key ? this.offset < e.offset : Od(Hd(Nd(this, "next")), Hd(Nd(e, "next"))) < 0;
	}
	getNode() {
		let e = ql(this.key);
		return e === null && V(20), e;
	}
	set(e, t, n, r) {
		let i = this._selection, a = this.key;
		r && this.key === e && this.offset === t && this.type === n || (this.key = e, this.offset = t, this.type = n, Vc() || (Kl() === a && Gl(e), i !== null && (i.setCachedNodes(null), i.dirty = !0)));
	}
};
function oc(e, t, n) {
	return new ac(e, t, n);
}
function sc(e, t) {
	let n = t.__key, r = e.offset, i = "element";
	if (G(t)) {
		i = "text";
		let e = t.getTextContentSize();
		r > e && (r = e);
	} else if (!Y(t)) {
		let e = t.getNextSibling();
		if (G(e)) n = e.__key, r = 0, i = "text";
		else {
			let e = t.getParent();
			e && (n = e.__key, r = t.getIndexWithinParent() + 1);
		}
	}
	e.set(n, r, i);
}
function cc(e, t) {
	if (Y(t)) {
		let n = t.getLastDescendant();
		Y(n) || G(n) ? sc(e, n) : sc(e, t);
	} else sc(e, t);
}
var lc = class e {
	constructor(e) {
		this._cachedNodes = null, this._nodes = e, this.dirty = !1;
	}
	getCachedNodes() {
		return this._cachedNodes;
	}
	setCachedNodes(e) {
		this._cachedNodes = e;
	}
	is(e) {
		if (!dc(e)) return !1;
		let t = this._nodes, n = e._nodes;
		return t.size === n.size && Array.from(t).every(((e) => n.has(e)));
	}
	isCollapsed() {
		return !1;
	}
	isBackward() {
		return !1;
	}
	getStartEndPoints() {
		return null;
	}
	add(e) {
		this.dirty = !0, this._nodes.add(e), this._cachedNodes = null;
	}
	delete(e) {
		this.dirty = !0, this._nodes.delete(e), this._cachedNodes = null;
	}
	clear() {
		this.dirty = !0, this._nodes.clear(), this._cachedNodes = null;
	}
	has(e) {
		return this._nodes.has(e);
	}
	clone() {
		return new e(new Set(this._nodes));
	}
	extract() {
		return this.getNodes();
	}
	insertRawText(e) {}
	insertText() {}
	insertNodes(e) {
		let t = this.getNodes(), n = t.length, r = t[n - 1], i;
		if (G(r)) i = r.select();
		else {
			let e = r.getIndexWithinParent() + 1;
			i = r.getParentOrThrow().select(e, e);
		}
		i.insertNodes(e);
		for (let e = 0; e < n; e++) t[e].remove();
	}
	getNodes() {
		let e = this._cachedNodes;
		if (e !== null) return e;
		let t = this._nodes, n = [];
		for (let e of t) {
			let t = ql(e);
			t !== null && n.push(t);
		}
		return Vc() || (this._cachedNodes = n), n;
	}
	getTextContent() {
		let e = this.getNodes(), t = "";
		for (let n = 0; n < e.length; n++) t += e[n].getTextContent();
		return t;
	}
	deleteNodes() {
		let e = this.getNodes();
		if ((q() || Tc()) === this && e[0]) {
			let t = hd(e[0], "next");
			Fd(Ed(t, t));
		}
		for (let t of e) t.remove();
	}
};
function K(e) {
	return e instanceof uc;
}
var uc = class e {
	constructor(e, t, n, r) {
		this.anchor = e, this.focus = t, e._selection = this, t._selection = this, this._cachedNodes = null, this.format = n, this.style = r, this.dirty = !1;
	}
	getCachedNodes() {
		return this._cachedNodes;
	}
	setCachedNodes(e) {
		this._cachedNodes = e;
	}
	is(e) {
		return !!K(e) && this.anchor.is(e.anchor) && this.focus.is(e.focus) && this.format === e.format && this.style === e.style;
	}
	isCollapsed() {
		return this.anchor.is(this.focus);
	}
	getNodes() {
		let e = this._cachedNodes;
		if (e !== null) return e;
		let t = function(e) {
			let t = [], [n, r] = e.getTextSlices();
			n && t.push(n.caret.origin);
			let i = /* @__PURE__ */ new Set(), a = /* @__PURE__ */ new Set();
			for (let n of e) if (dd(n)) {
				let { origin: e } = n;
				t.length === 0 ? i.add(e) : (a.add(e), t.push(e));
			} else {
				let { origin: e } = n;
				Y(e) && a.has(e) || t.push(e);
			}
			if (r && t.push(r.caret.origin), ud(e.focus) && Y(e.focus.origin) && e.focus.getNodeAtCaret() === null) for (let n = yd(e.focus.origin, "previous"); dd(n) && i.has(n.origin) && !n.origin.isEmpty() && n.origin.is(t[t.length - 1]); n = xd(n)) i.delete(n.origin), t.pop();
			for (; t.length > 1;) {
				let e = t[t.length - 1];
				if (!Y(e) || a.has(e) || e.isEmpty() || i.has(e)) break;
				t.pop();
			}
			if (t.length === 0 && e.isCollapsed()) {
				let n = Hd(e.anchor), r = Hd(e.anchor.getFlipped()), i = (e) => ld(e) ? e.origin : e.getNodeAtCaret(), a = i(n) || i(r) || (e.anchor.getNodeAtCaret() ? n.origin : r.origin);
				t.push(a);
			}
			return t;
		}(Gd(Ld(this), "next"));
		return Vc() || (this._cachedNodes = t), t;
	}
	setTextNodeRange(e, t, n, r) {
		this.anchor.set(e.__key, t, "text"), this.focus.set(n.__key, r, "text");
	}
	getTextContent() {
		let e = this.getNodes();
		if (e.length === 0) return "";
		let t = e[0], n = e[e.length - 1], r = this.anchor, i = this.focus, a = r.isBefore(i), [o, s] = pc(this), c = "", l = !0;
		for (let u = 0; u < e.length; u++) {
			let d = e[u];
			if (Y(d) && !d.isInline()) l || (c += "\n"), l = !d.isEmpty();
			else if (l = !1, G(d)) {
				let e = d.getTextContent();
				d === t ? d === n ? r.type === "element" && i.type === "element" && i.offset !== r.offset || (e = o < s ? e.slice(o, s) : e.slice(s, o)) : e = a ? e.slice(o) : e.slice(s) : d === n && (e = a ? e.slice(0, s) : e.slice(0, o)), c += e;
			} else !X(d) && !Rs(d) || d === n && this.isCollapsed() || (c += d.getTextContent());
		}
		return c;
	}
	applyDOMRange(e) {
		let t = J(), n = t.getEditorState()._selection, r = bc(e.startContainer, e.startOffset, e.endContainer, e.endOffset, t, n);
		if (r === null) return;
		let [i, a] = r;
		this.anchor.set(i.key, i.offset, i.type, !0), this.focus.set(a.key, a.offset, a.type, !0), Ia(this);
	}
	clone() {
		let t = this.anchor, n = this.focus;
		return new e(oc(t.key, t.offset, t.type), oc(n.key, n.offset, n.type), this.format, this.style);
	}
	toggleFormat(e) {
		this.format = Vl(this.format, e, null), this.dirty = !0;
	}
	setFormat(e) {
		this.format = e, this.dirty = !0;
	}
	setStyle(e) {
		this.style = e, this.dirty = !0;
	}
	hasFormat(e) {
		let t = oa[e];
		return !!(this.format & t);
	}
	insertRawText(e) {
		let t = e.split(/(\r?\n|\t)/), n = [], r = t.length;
		for (let e = 0; e < r; e++) {
			let r = t[e];
			r === "\n" || r === "\r\n" ? n.push(Ls()) : r === "	" ? n.push(rc()) : n.push(W(r));
		}
		this.insertNodes(n);
	}
	insertText(e) {
		let t = this.anchor, n = this.focus, r = this.format, i = this.style, a = t, o = n;
		!this.isCollapsed() && n.isBefore(t) && (a = n, o = t), a.type === "element" && function(e, t, n, r) {
			let i = e.getNode(), a = i.getChildAtIndex(e.offset), o = W();
			if (o.setFormat(n), o.setStyle(r), bl(a)) a.splice(0, 0, [o]);
			else {
				let e = ll(i) ? yl().append(o) : o;
				a === null ? i.append(e) : a.insertBefore(e);
			}
			e.is(t) && t.set(o.__key, 0, "text"), e.set(o.__key, 0, "text");
		}(a, o, r, i), o.type === "element" && Pd(o, Hd(Nd(o, "next")));
		let s = a.offset, c = o.offset, l = this.getNodes(), u = l.length, d = l[0];
		G(d) || V(26);
		let f = d.getTextContent().length, p = d.getParentOrThrow(), m = l[u - 1];
		if (u === 1 && o.type === "element" && (c = f, o.set(a.key, c, "text")), this.isCollapsed() && s === f && (Ll(d) || !d.canInsertTextAfter() || !p.canInsertTextAfter() && d.getNextSibling() === null)) {
			let t = d.getNextSibling();
			if (G(t) && t.canInsertTextBefore() && !Ll(t) || (t = W(), t.setFormat(r), t.setStyle(i), p.canInsertTextAfter() ? d.insertAfter(t) : p.insertAfter(t)), t.select(0, 0), d = t, e !== "") return void this.insertText(e);
		} else if (this.isCollapsed() && s === 0 && (Ll(d) || !d.canInsertTextBefore() || !p.canInsertTextBefore() && d.getPreviousSibling() === null)) {
			let t = d.getPreviousSibling();
			if (G(t) && !Ll(t) || (t = W(), t.setFormat(r), p.canInsertTextBefore() ? d.insertBefore(t) : p.insertBefore(t)), t.select(), d = t, e !== "") return void this.insertText(e);
		} else if (d.isSegmented() && s !== f) {
			let e = W(d.getTextContent());
			e.setFormat(r), d.replace(e), d = e;
		} else if (!this.isCollapsed() && e !== "") {
			let t = m.getParent();
			if (!p.canInsertTextBefore() || !p.canInsertTextAfter() || Y(t) && (!t.canInsertTextBefore() || !t.canInsertTextAfter())) return this.insertText(""), yc(this.anchor, this.focus, null), void this.insertText(e);
		}
		if (u === 1) {
			if (Il(d)) {
				let t = W(e);
				t.select(), d.replace(t);
				return;
			}
			let t = d.getFormat(), n = d.getStyle();
			if (s !== c || t === r && n === i) {
				if (ic(d)) {
					let t = W(e);
					t.setFormat(r), t.setStyle(i), t.select(), d.replace(t);
					return;
				}
			} else {
				if (d.getTextContent() !== "") {
					let t = W(e);
					if (t.setFormat(r), t.setStyle(i), t.select(), s === 0) d.insertBefore(t, !1);
					else {
						let [e] = d.splitText(s);
						e.insertAfter(t, !1);
					}
					t.isComposing() && this.anchor.type === "text" && (this.anchor.offset -= e.length);
					return;
				}
				d.setFormat(r), d.setStyle(i);
			}
			let a = c - s;
			d = d.spliceText(s, a, e, !0), d.getTextContent() === "" ? d.remove() : this.anchor.type === "text" && (d.isComposing() ? this.anchor.offset -= e.length : (this.format = t, this.style = n));
		} else {
			let t = /* @__PURE__ */ new Set([...d.getParentKeys(), ...m.getParentKeys()]), n = Y(d) ? d : d.getParentOrThrow(), r = Y(m) ? m : m.getParentOrThrow(), i = m;
			if (!n.is(r) && r.isInline()) do
				i = r, r = r.getParentOrThrow();
			while (r.isInline());
			if (o.type === "text" && (c !== 0 || m.getTextContent() === "") || o.type === "element" && m.getIndexWithinParent() < c) {
				if (G(m) && !Il(m) && c !== m.getTextContentSize()) {
					if (m.isSegmented()) {
						let e = W(m.getTextContent());
						m.replace(e), m = e;
					}
					ll(o.getNode()) || o.type !== "text" || (m = m.spliceText(0, c, "")), t.add(m.__key);
				} else {
					let e = m.getParentOrThrow();
					e.canBeEmpty() || e.getChildrenSize() !== 1 ? m.remove() : e.remove();
				}
			} else t.add(m.__key);
			let a = r.getChildren(), p = new Set(l), h = n.is(r), g = n.isInline() && d.getNextSibling() === null ? n : d;
			for (let e = a.length - 1; e >= 0; e--) {
				let t = a[e];
				if (t.is(d) || Y(t) && t.isParentOf(d)) break;
				t.isAttached() && (!p.has(t) || t.is(i) ? h || g.insertAfter(t, !1) : t.remove());
			}
			if (!h) {
				let e = r, n = null;
				for (; e !== null;) {
					let r = e.getChildren(), i = r.length;
					(i === 0 || r[i - 1].is(n)) && (t.delete(e.__key), n = e), e = e.getParent();
				}
			}
			if (Il(d)) {
				if (s === f) d.select();
				else {
					let t = W(e);
					t.select(), d.replace(t);
				}
			} else d = d.spliceText(s, f - s, e, !0), d.getTextContent() === "" ? d.remove() : d.isComposing() && this.anchor.type === "text" && (this.anchor.offset -= e.length);
			for (let e = 1; e < u; e++) {
				let n = l[e], r = n.__key;
				t.has(r) || n.remove();
			}
		}
	}
	removeText() {
		let e = q() === this;
		Id(this, Vd(Ld(this))), e && q() !== this && eu(this);
	}
	formatText(e, t = null) {
		if (this.isCollapsed()) return this.toggleFormat(e), void Gl(null);
		let n = this.getNodes(), r = [];
		for (let e of n) G(e) && r.push(e);
		let i = (t) => {
			n.forEach(((n) => {
				if (Y(n)) {
					let r = n.getFormatFlags(e, t);
					n.setTextFormat(r);
				}
			}));
		}, a = r.length;
		if (a === 0) return this.toggleFormat(e), Gl(null), void i(t);
		let o = this.anchor, s = this.focus, c = this.isBackward(), l = c ? s : o, u = c ? o : s, d = 0, f = r[0], p = l.type === "element" ? 0 : l.offset;
		if (l.type === "text" && p === f.getTextContentSize() && (d = 1, f = r[1], p = 0), f == null) return;
		let m = f.getFormatFlags(e, t);
		i(m);
		let h = a - 1, g = r[h], _ = u.type === "text" ? u.offset : g.getTextContentSize();
		if (f.is(g)) {
			if (p === _) return;
			if (Ll(f) || p === 0 && _ === f.getTextContentSize()) f.setFormat(m);
			else {
				let e = f.splitText(p, _), t = p === 0 ? e[0] : e[1];
				t.setFormat(m), l.type === "text" && l.set(t.__key, 0, "text"), u.type === "text" && u.set(t.__key, _ - p, "text");
			}
			this.format = m;
			return;
		}
		p === 0 || Ll(f) || ([, f] = f.splitText(p), p = 0), f.setFormat(m);
		let v = g.getFormatFlags(e, m);
		_ > 0 && (_ === g.getTextContentSize() || Ll(g) || ([g] = g.splitText(_)), g.setFormat(v));
		for (let t = d + 1; t < h; t++) {
			let n = r[t], i = n.getFormatFlags(e, v);
			n.setFormat(i);
		}
		l.type === "text" && l.set(f.__key, p, "text"), u.type === "text" && u.set(g.__key, _, "text"), this.format = m | v;
	}
	insertNodes(e) {
		if (e.length === 0) return;
		if (this.isCollapsed() || this.removeText(), this.anchor.key === "root") {
			this.insertParagraph();
			let t = q();
			return K(t) || V(134), t.insertNodes(e);
		}
		let t = (this.isBackward() ? this.focus : this.anchor).getNode(), n = Gu(t, Wu), r = e[e.length - 1];
		if (Y(n) && "__language" in n) {
			if ("__language" in e[0]) this.insertText(e[0].getTextContent());
			else {
				let t = Mc(this);
				n.splice(t, 0, e), r.selectEnd();
			}
			return;
		}
		if (!e.some(((e) => (Y(e) || X(e)) && !e.isInline()))) {
			Y(n) || V(211, t.constructor.name, t.getType());
			let i = Mc(this);
			n.splice(i, 0, e), r.selectEnd();
			return;
		}
		let i = function(e) {
			let t = yl(), n = null;
			for (let r = 0; r < e.length; r++) {
				let i = e[r], a = Rs(i);
				if (a || X(i) && i.isInline() || Y(i) && i.isInline() || G(i) || i.isParentRequired()) {
					if (n === null && (n = i.createParentElementNode(), t.append(n), a)) continue;
					n !== null && n.append(i);
				} else t.append(i), n = null;
			}
			return t;
		}(e), a = i.getLastDescendant(), o = i.getChildren(), s = !Y(n) || !n.isEmpty() ? this.insertParagraph() : null, c = o[o.length - 1], l = o[0];
		var u;
		Y(u = l) && Wu(u) && !u.isEmpty() && Y(n) && (!n.isEmpty() || n.canMergeWhenEmpty()) && (Y(n) || V(211, t.constructor.name, t.getType()), n.append(...l.getChildren()), l = o[1]), l && (n === null && V(212, t.constructor.name, t.getType()), function(e, t, n) {
			let r = t.getParentOrThrow().getLastChild(), i = t, a = [t];
			for (; i !== r;) i.getNextSibling() || V(140), i = i.getNextSibling(), a.push(i);
			let o = e;
			for (let e of a) o = o.insertAfter(e);
		}(n, l));
		let d = Gu(a, Wu);
		s && Y(d) && (s.canMergeWhenEmpty() || Wu(c)) && (d.append(...s.getChildren()), s.remove()), Y(n) && n.isEmpty() && n.remove(), a.selectEnd();
		let f = Y(n) ? n.getLastChild() : null;
		Rs(f) && d !== n && f.remove();
	}
	insertParagraph() {
		if (this.anchor.key === "root") {
			let e = yl();
			return Z().splice(this.anchor.offset, 0, [e]), e.select(), e;
		}
		let e = Mc(this), t = Gu(this.anchor.getNode(), Wu);
		Y(t) || V(213);
		let n = t.getChildAtIndex(e), r = n ? [n, ...n.getNextSiblings()] : [], i = t.insertNewAfter(this, !1);
		return i ? (i.append(...r), i.selectStart(), i) : null;
	}
	insertLineBreak(e) {
		let t = Ls();
		if (this.insertNodes([t]), e) {
			let e = t.getParentOrThrow(), n = t.getIndexWithinParent();
			e.select(n, n);
		}
	}
	extract() {
		let e = this.getNodes(), t = e.length, n = t - 1, r = this.anchor, i = this.focus, a = e[0], o = e[n], [s, c] = pc(this);
		if (t === 0) return [];
		if (t === 1) {
			if (G(a) && !this.isCollapsed()) {
				let e = s > c ? c : s, t = s > c ? s : c, n = a.splitText(e, t), r = e === 0 ? n[0] : n[1];
				return r == null ? [] : [r];
			}
			return [a];
		}
		let l = r.isBefore(i);
		if (G(a)) {
			let t = l ? s : c;
			t === a.getTextContentSize() ? e.shift() : t !== 0 && ([, a] = a.splitText(t), e[0] = a);
		}
		if (G(o)) {
			let t = o.getTextContent().length, r = l ? c : s;
			r === 0 ? e.pop() : r !== t && ([o] = o.splitText(r), e[n] = o);
		}
		return e;
	}
	modify(e, t, n) {
		if (Pc(this, e, t, n)) return;
		let r = e === "move", i = J(), a = Lu(ku(i));
		if (!a) return;
		let o = i._blockCursorElement, s = i._rootElement, c = this.focus.getNode();
		if (s === null || o === null || !Y(c) || c.isInline() || c.canBeEmpty() || Iu(o, i, s), this.dirty) {
			let e = Su(i, this.anchor.key), t = Su(i, this.focus.key);
			this.anchor.type === "text" && (e = Bl(e)), this.focus.type === "text" && (t = Bl(t)), e && t && Ac(a, e, this.anchor.offset, t, this.focus.offset);
		}
		if (function(e, t, n, r) {
			e.modify(t, n, r);
		}(a, e, t ? "backward" : "forward", n), a.rangeCount > 0) {
			let e = a.getRangeAt(0), n = this.anchor.getNode(), i = ll(n) ? n : Au(n);
			if (this.applyDOMRange(e), this.dirty = !0, !r) {
				let n = this.getNodes(), r = [], o = !1;
				for (let e = 0; e < n.length; e++) {
					let t = n[e];
					Du(t, i) ? r.push(t) : o = !0;
				}
				if (o && r.length > 0) {
					if (t) {
						let e = r[0];
						Y(e) ? e.selectStart() : e.getParentOrThrow().selectStart();
					} else {
						let e = r[r.length - 1];
						Y(e) ? e.selectEnd() : e.getParentOrThrow().selectEnd();
					}
				}
				a.anchorNode === e.startContainer && a.anchorOffset === e.startOffset || function(e) {
					let t = e.focus, n = e.anchor, r = n.key, i = n.offset, a = n.type;
					n.set(t.key, t.offset, t.type, !0), t.set(r, i, a, !0);
				}(this);
			}
		}
		n === "lineboundary" && Pc(this, e, t, n, "decorators");
	}
	forwardDeletion(e, t, n) {
		if (!n && (e.type === "element" && Y(t) && e.offset === t.getChildrenSize() || e.type === "text" && e.offset === t.getTextContentSize())) {
			let e = t.getParent(), n = t.getNextSibling() || (e === null ? null : e.getNextSibling());
			if (Y(n) && n.isShadowRoot()) return !0;
		}
		return !1;
	}
	deleteCharacter(e) {
		let t = this.isCollapsed();
		if (this.isCollapsed()) {
			let t = this.anchor, n = t.getNode();
			if (this.forwardDeletion(t, n, e)) return;
			let r = wd(Nd(t, e ? "previous" : "next"));
			if (r.getTextSlices().every(((e) => e === null || e.distance === 0))) {
				let e = { type: "initial" };
				for (let t of r.iterNodeCarets("shadowRoot")) if (dd(t)) {
					if (!t.origin.isInline()) {
						if (t.origin.isShadowRoot()) {
							if (e.type === "merge-block") break;
							if (Y(r.anchor.origin) && r.anchor.origin.isEmpty()) {
								let e = Hd(t);
								Id(this, Ed(e, e)), r.anchor.origin.remove();
							}
							return;
						}
						e.type !== "merge-next-block" && e.type !== "merge-block" || (e = {
							block: e.block,
							caret: t,
							type: "merge-block"
						});
					}
				} else {
					if (e.type === "merge-block") break;
					if (ud(t)) {
						if (Y(t.origin)) {
							if (t.origin.isInline()) {
								if (!t.origin.isParentOf(r.anchor.origin)) break;
							} else e = {
								block: t.origin,
								type: "merge-next-block"
							};
							continue;
						}
						if (X(t.origin)) {
							if (!t.origin.isIsolated()) {
								if (e.type === "merge-next-block" && (t.origin.isKeyboardSelectable() || !t.origin.isInline()) && Y(r.anchor.origin) && r.anchor.origin.isEmpty()) {
									r.anchor.origin.remove();
									let e = Cc();
									e.add(t.origin.getKey()), eu(e);
								} else t.origin.remove();
							}
							return;
						}
						break;
					}
				}
				if (e.type === "merge-block") {
					let { caret: t, block: n } = e;
					return Id(this, Ed(!t.origin.isEmpty() && n.isEmpty() ? Rd(hd(n, t.direction)) : r.anchor, t)), this.removeText();
				}
			}
			let i = this.focus;
			if (this.modify("extend", e, "character"), this.isCollapsed()) {
				if (e && t.offset === 0 && mc(this, t.getNode())) return;
			} else {
				let r = i.type === "text" ? i.getNode() : null;
				if (n = t.type === "text" ? t.getNode() : null, r !== null && r.isSegmented()) {
					let t = i.offset, a = r.getTextContentSize();
					if (r.is(n) || e && t !== a || !e && t !== 0) return void gc(r, e, t);
				} else if (n !== null && n.isSegmented()) {
					let i = t.offset, a = n.getTextContentSize();
					if (n.is(r) || e && i !== 0 || !e && i !== a) return void gc(n, e, i);
				}
				(function(e, t) {
					let n = e.anchor, r = e.focus, i = n.getNode();
					if (i === r.getNode() && n.type === "text" && r.type === "text") {
						let e = n.offset, a = r.offset, o = e < a, s = o ? e : a, c = o ? a : e, l = c - 1;
						s !== l && (function(e) {
							return !(nu(e) || hc(e));
						})(i.getTextContent().slice(s, c)) && (t ? r.set(r.key, l, r.type) : n.set(n.key, l, n.type));
					}
				})(this, e);
			}
		}
		if (this.removeText(), e && !t && this.isCollapsed() && this.anchor.type === "element" && this.anchor.offset === 0) {
			let e = this.anchor.getNode();
			e.isEmpty() && ll(e.getParent()) && e.getPreviousSibling() === null && mc(this, e);
		}
	}
	deleteLine(e) {
		this.isCollapsed() && this.modify("extend", e, "lineboundary"), this.isCollapsed() ? this.deleteCharacter(e) : this.removeText();
	}
	deleteWord(e) {
		if (this.isCollapsed()) {
			let t = this.anchor, n = t.getNode();
			if (this.forwardDeletion(t, n, e)) return;
			this.modify("extend", e, "word");
		}
		this.removeText();
	}
	isBackward() {
		return this.focus.isBefore(this.anchor);
	}
	getStartEndPoints() {
		return [this.anchor, this.focus];
	}
};
function dc(e) {
	return e instanceof lc;
}
function fc(e) {
	let t = e.offset;
	if (e.type === "text") return t;
	let n = e.getNode();
	return t === n.getChildrenSize() ? n.getTextContent().length : 0;
}
function pc(e) {
	let t = e.getStartEndPoints();
	if (t === null) return [0, 0];
	let [n, r] = t;
	return n.type === "element" && r.type === "element" && n.key === r.key && n.offset === r.offset ? [0, 0] : [fc(n), fc(r)];
}
function mc(e, t) {
	for (let n = t; n; n = n.getParent()) {
		if (Y(n)) {
			if (n.collapseAtStart(e)) return !0;
			if (ju(n)) break;
		}
		if (n.getPreviousSibling()) break;
	}
	return !1;
}
var hc = (() => {
	try {
		let e = /* @__PURE__ */ RegExp("\\p{Emoji}", "u"), t = e.test.bind(e);
		if (t("❤️") && t("#️⃣") && t("👍")) return t;
	} catch {}
	return () => !1;
})();
function gc(e, t, n) {
	let r = e, i = r.getTextContent().split(/(?=\s)/g), a = i.length, o = 0, s = 0;
	for (let e = 0; e < a; e++) {
		let r = e === a - 1;
		if (s = o, o += i[e].length, t && o === n || o > n || r) {
			i.splice(e, 1), r && (s = void 0);
			break;
		}
	}
	let c = i.join("").trim();
	c === "" ? r.remove() : (r.setTextContent(c), r.select(s, s));
}
function _c(e, t, n, r) {
	let i, a = t;
	if (zu(e)) {
		let o = !1, s = e.childNodes, c = s.length, l = r._blockCursorElement;
		a === c && (o = !0, a = c - 1);
		let u = s[a], d = !1;
		if (u === l) u = s[a + 1], d = !0;
		else if (l !== null) {
			let n = l.parentNode;
			e === n && t > Array.prototype.indexOf.call(n.children, l) && a--;
		}
		if (i = tu(u), G(i)) a = _d(i, o ? "next" : "previous");
		else {
			let s = tu(e);
			if (s === null) return null;
			if (Y(s)) {
				let c = r.getElementByKey(s.getKey());
				c === null && V(214);
				let l = s.getDOMSlot(c);
				[s, a] = l.resolveChildIndex(s, c, e, t), Y(s) || V(215), o && a >= s.getChildrenSize() && (a = Math.max(0, s.getChildrenSize() - 1));
				let u = s.getChildAtIndex(a);
				if (Y(u) && function(e, t, n) {
					let r = e.getParent();
					return n === null || r === null || !r.canBeEmpty() || r !== n.getNode();
				}(u, 0, n)) {
					let e = o ? u.getLastDescendant() : u.getFirstDescendant();
					e === null ? s = u : (u = e, s = Y(u) ? u : u.getParentOrThrow()), a = 0;
				}
				G(u) ? (i = u, s = null, a = _d(u, o ? "next" : "previous")) : u !== s && o && !d && (Y(s) || V(216), a = Math.min(s.getChildrenSize(), a + 1));
			} else {
				let n = s.getIndexWithinParent();
				a = t === 0 && X(s) && tu(e) === s ? n : n + 1, s = s.getParentOrThrow();
			}
			if (Y(s)) return oc(s.__key, a, "element");
		}
	} else i = tu(e);
	return G(i) ? oc(i.__key, _d(i, a, "clamp"), "text") : null;
}
function vc(e, t, n) {
	let r = e.offset, i = e.getNode();
	if (r === 0) {
		let r = i.getPreviousSibling(), a = i.getParent();
		if (t) {
			if ((n || !t) && r === null && Y(a) && a.isInline()) {
				let t = a.getPreviousSibling();
				G(t) && e.set(t.__key, t.getTextContent().length, "text");
			}
		} else Y(r) && !n && r.isInline() ? e.set(r.__key, r.getChildrenSize(), "element") : G(r) && e.set(r.__key, r.getTextContent().length, "text");
	} else if (r === i.getTextContent().length) {
		let r = i.getNextSibling(), a = i.getParent();
		if (t && Y(r) && r.isInline()) e.set(r.__key, 0, "element");
		else if ((n || t) && r === null && Y(a) && a.isInline() && !a.canInsertTextAfter()) {
			let t = a.getNextSibling();
			G(t) && e.set(t.__key, 0, "text");
		}
	}
}
function yc(e, t, n) {
	if (e.type === "text" && t.type === "text") {
		let r = e.isBefore(t), i = e.is(t);
		vc(e, r, i), vc(t, !r, i), i && t.set(e.key, e.offset, e.type);
		let a = J();
		if (a.isComposing() && a._compositionKey !== e.key && K(n)) {
			let r = n.anchor, i = n.focus;
			e.set(r.key, r.offset, r.type, !0), t.set(i.key, i.offset, i.type, !0);
		}
	}
}
function bc(e, t, n, r, i, a) {
	if (e === null || n === null || !Ml(i, e, n)) return null;
	let o = _c(e, t, K(a) ? a.anchor : null, i);
	if (o === null) return null;
	let s = _c(n, r, K(a) ? a.focus : null, i);
	if (s === null) return null;
	if (o.type === "element" && s.type === "element") {
		let t = tu(e), r = tu(n);
		if (X(t) && X(r)) return null;
	}
	return yc(o, s, a), [o, s];
}
function xc(e, t, n, r, i, a) {
	let o = Wc(), s = new uc(oc(e, t, i), oc(n, r, a), 0, "");
	return s.dirty = !0, o._selection = s, s;
}
function Sc() {
	return new uc(oc("root", 0, "element"), oc("root", 0, "element"), 0, "");
}
function Cc() {
	return new lc(/* @__PURE__ */ new Set());
}
function wc(e, t, n, r) {
	let i = n._window;
	if (i === null) return null;
	let a = r || i.event, o = a ? a.type : void 0, s = o === "selectionchange", c = !ha && (s || o === "beforeinput" || o === "compositionstart" || o === "compositionend" || o === "click" && a && a.detail === 3 || o === "drop" || o === void 0), l, u, d, f;
	if (K(e) && !c) return e.clone();
	if (t === null) return null;
	if (l = t.anchorNode, u = t.focusNode, d = t.anchorOffset, f = t.focusOffset, s && K(e) && !Ml(n, l, u)) return e.clone();
	let p = bc(l, d, u, f, n, e);
	if (p === null) return null;
	let [m, h] = p;
	return new uc(m, h, K(e) ? e.format : 0, K(e) ? e.style : "");
}
function q() {
	return Wc()._selection;
}
function Tc() {
	return J()._editorState._selection;
}
function Ec(e, t, n, r = 1) {
	let i = e.anchor, a = e.focus, o = i.getNode(), s = a.getNode();
	if (!t.is(o) && !t.is(s)) return;
	let c = t.__key;
	if (e.isCollapsed()) {
		let t = i.offset;
		if (n <= t && r > 0 || n < t && r < 0) {
			let n = Math.max(0, t + r);
			i.set(c, n, "element"), a.set(c, n, "element"), Dc(e);
		}
	} else {
		let o = e.isBackward(), s = o ? a : i, l = s.getNode(), u = o ? i : a, d = u.getNode();
		if (t.is(l)) {
			let e = s.offset;
			(n <= e && r > 0 || n < e && r < 0) && s.set(c, Math.max(0, e + r), "element");
		}
		if (t.is(d)) {
			let e = u.offset;
			(n <= e && r > 0 || n < e && r < 0) && u.set(c, Math.max(0, e + r), "element");
		}
	}
	Dc(e);
}
function Dc(e) {
	let t = e.anchor, n = t.offset, r = e.focus, i = r.offset, a = t.getNode(), o = r.getNode();
	if (e.isCollapsed()) {
		if (!Y(a)) return;
		let e = a.getChildrenSize(), i = n >= e, o = i ? a.getChildAtIndex(e - 1) : a.getChildAtIndex(n);
		if (G(o)) {
			let e = 0;
			i && (e = o.getTextContentSize()), t.set(o.__key, e, "text"), r.set(o.__key, e, "text");
		}
	} else {
		if (Y(a)) {
			let e = a.getChildrenSize(), r = n >= e, i = r ? a.getChildAtIndex(e - 1) : a.getChildAtIndex(n);
			if (G(i)) {
				let e = 0;
				r && (e = i.getTextContentSize()), t.set(i.__key, e, "text");
			}
		}
		if (Y(o)) {
			let e = o.getChildrenSize(), t = i >= e, n = t ? o.getChildAtIndex(e - 1) : o.getChildAtIndex(i);
			if (G(n)) {
				let e = 0;
				t && (e = n.getTextContentSize()), r.set(n.__key, e, "text");
			}
		}
	}
}
function Oc(e, t, n, r, i) {
	let a = null, o = 0, s = null;
	r === null ? i !== null && (a = i.__key, G(i) ? s = "text" : Y(i) && (s = "element")) : (a = r.__key, G(r) ? (o = r.getTextContentSize(), s = "text") : Y(r) && (o = r.getChildrenSize(), s = "element")), a !== null && s !== null ? e.set(a, o, s) : (o = t.getIndexWithinParent(), o === -1 && (o = n.getChildrenSize()), e.set(n.__key, o, "element"));
}
function kc(e, t, n, r, i) {
	e.type === "text" ? e.set(n, e.offset + (t ? 0 : i), "text") : e.offset > r.getIndexWithinParent() && e.set(e.key, e.offset - 1, "element");
}
function Ac(e, t, n, r, i) {
	try {
		e.setBaseAndExtent(t, n, r, i);
	} catch {}
}
function jc(e, t, n, r, i, a, o) {
	let s = r.anchorNode, c = r.focusNode, l = r.anchorOffset, u = r.focusOffset, d = document.activeElement;
	if (i.has("collaboration") && d !== a || d !== null && jl(d)) return;
	if (!K(t)) return void (e !== null && Ml(n, s, c) && r.removeAllRanges());
	let f = t.anchor, p = t.focus, m = f.key, h = p.key, g = Su(n, m), _ = Su(n, h), v = f.offset, y = p.offset, b = t.format, x = t.style, S = t.isCollapsed(), C = g, w = _, T = !1;
	if (f.type === "text") {
		C = Bl(g);
		let e = f.getNode();
		T = e.getFormat() !== b || e.getStyle() !== x;
	} else K(e) && e.anchor.type === "text" && (T = !0);
	var E, ee, te, ne, D;
	if (p.type === "text" && (w = Bl(_)), C !== null && w !== null && (S && (e === null || T || K(e) && (e.format !== b || e.style !== x)) && (E = b, ee = x, te = v, ne = m, D = performance.now(), ys = [
		E,
		ee,
		te,
		ne,
		D
	]), l !== v || u !== y || s !== C || c !== w || r.type === "Range" && S || (d !== null && a.contains(d) || a.focus({ preventScroll: !0 }), f.type === "element"))) {
		if (Ac(r, C, v, w, y), !i.has("skip-scroll-into-view") && t.isCollapsed() && a !== null && a === document.activeElement) {
			let e = K(t) && t.anchor.type === "element" ? C.childNodes[v] || null : r.rangeCount > 0 ? r.getRangeAt(0) : null;
			if (e !== null) {
				let t;
				if (e instanceof Text) {
					let n = document.createRange();
					n.selectNode(e), t = n.getBoundingClientRect();
				} else t = e.getBoundingClientRect();
				(function(e, t, n) {
					let r = wu(n), i = Ou(r);
					if (r === null || i === null) return;
					let { top: a, bottom: o } = t, s = 0, c = 0, l = n;
					for (; l !== null;) {
						let t = l === r.body;
						if (t) s = 0, c = ku(e).innerHeight;
						else {
							let e = l.getBoundingClientRect();
							s = e.top, c = e.bottom;
						}
						let n = 0;
						if (a < s ? n = -(s - a) : o > c && (n = o - c), n !== 0) {
							if (t) i.scrollBy(0, n);
							else {
								let e = l.scrollTop;
								l.scrollTop += n;
								let t = l.scrollTop - e;
								a -= t, o -= t;
							}
						}
						if (t) break;
						l = Cu(l);
					}
				})(n, t, a);
			}
		}
		fs = !0;
	}
}
function Mc(e) {
	let t = e;
	e.isCollapsed() || t.removeText();
	let n = q();
	K(n) && (t = n), K(t) || V(161);
	let r = t.anchor, i = r.getNode(), a = r.offset;
	for (; !Wu(i);) {
		let e = i;
		if ([i, a] = Nc(i, a), e.is(i)) break;
	}
	return a;
}
function Nc(e, t) {
	let n = e.getParent();
	if (!n) {
		let e = yl();
		return Z().append(e), e.select(), [Z(), 0];
	}
	if (G(e)) {
		let r = e.splitText(t);
		if (r.length === 0) return [n, e.getIndexWithinParent()];
		let i = t === 0 ? 0 : 1;
		return [n, r[0].getIndexWithinParent() + i];
	}
	if (!Y(e) || t === 0) return [n, e.getIndexWithinParent()];
	let r = e.getChildAtIndex(t);
	if (r) {
		let n = new uc(oc(e.__key, t, "element"), oc(e.__key, t, "element"), 0, ""), i = e.insertNewAfter(n);
		i && i.append(r, ...r.getNextSiblings());
	}
	return [n, e.getIndexWithinParent() + 1];
}
function Pc(e, t, n, r, i = "decorators-and-blocks") {
	if (t === "move" && r === "character" && !e.isCollapsed()) {
		let [t, r] = n === e.isBackward() ? [e.focus, e.anchor] : [e.anchor, e.focus];
		return r.set(t.key, t.offset, t.type), !0;
	}
	let a = Nd(e.focus, n ? "previous" : "next"), o = r === "lineboundary", s = t === "move", c = a, l = i === "decorators-and-blocks";
	if (!Ud(c)) {
		for (let e of c) {
			l = !1;
			let { origin: t } = e;
			if (!X(t) || t.isIsolated() || (c = e, !o || !t.isInline())) break;
		}
		if (l) for (let e of wd(a).iterNodeCarets(t === "extend" ? "shadowRoot" : "root")) {
			if (dd(e)) e.origin.isInline() || (c = e);
			else {
				if (Y(e.origin)) continue;
				X(e.origin) && !e.origin.isInline() && (c = e);
			}
			break;
		}
	}
	if (c === a) return !1;
	if (s && !o && X(c.origin) && c.origin.isKeyboardSelectable()) {
		let e = Cc();
		return e.add(c.origin.getKey()), eu(e), !0;
	}
	return c = Hd(c), s && Pd(e.anchor, c), Pd(e.focus, c), l || !o;
}
var Fc = null, Ic = null, Lc = !1, Rc = !1, zc = 0, Bc = {
	characterData: !0,
	childList: !0,
	subtree: !0
};
function Vc() {
	return Lc || Fc !== null && Fc._readOnly;
}
function Hc() {
	Lc && V(13);
}
function Uc() {
	zc > 99 && V(14);
}
function Wc() {
	return Fc === null && V(195, Gc()), Fc;
}
function J() {
	return Ic === null && V(196, Gc()), Ic;
}
function Gc() {
	let e = 0, t = /* @__PURE__ */ new Set(), n = Cl.version;
	if (typeof window < "u") for (let r of document.querySelectorAll("[contenteditable]")) {
		let i = Fl(r);
		if (Nl(i)) e++;
		else if (i) {
			let e = String(i.constructor.version || "<0.17.1");
			e === n && (e += " (separately built, likely a bundler configuration issue)"), t.add(e);
		}
	}
	let r = ` Detected on the page: ${e} compatible editor(s) with version ${n}`;
	return t.size && (r += ` and incompatible editors with versions ${Array.from(t).join(", ")}`), r;
}
function Kc() {
	return Ic;
}
function qc(e, t, n) {
	let r = t.__type, i = Dl(e, r), a = n.get(r);
	a === void 0 && (a = Array.from(i.transforms), n.set(r, a));
	let o = a.length;
	for (let e = 0; e < o && (a[e](t), t.isAttached()); e++);
}
function Jc(e, t) {
	return e !== void 0 && e.__key !== t && e.isAttached();
}
function Yc(e, t) {
	if (!t) return;
	let n = e._updateTags, r = t;
	Array.isArray(t) || (r = [t]);
	for (let e of r) n.add(e);
}
function Xc(e, t) {
	let n = e.type, r = t.get(n);
	r === void 0 && V(17, n);
	let i = r.klass;
	e.type !== i.getType() && V(18, i.name);
	let a = i.importJSON(e), o = e.children;
	if (Y(a) && Array.isArray(o)) for (let e = 0; e < o.length; e++) {
		let n = Xc(o[e], t);
		a.append(n);
	}
	return a;
}
function Zc(e, t, n) {
	let r = Fc, i = Lc, a = Ic;
	Fc = t, Lc = !0, Ic = e;
	try {
		return n();
	} finally {
		Fc = r, Lc = i, Ic = a;
	}
}
function Qc(e, t) {
	let n = e._pendingEditorState, r = e._rootElement, i = e._headless || r === null;
	if (n === null) return;
	let a = e._editorState, o = a._selection, s = n._selection, c = e._dirtyType !== Gi, l = Fc, u = Lc, d = Ic, f = e._updating, p = e._observer, m = null;
	if (e._pendingEditorState = null, e._editorState = n, !i && c && p !== null) {
		Ic = e, Fc = n, Lc = !1, e._updating = !0;
		try {
			let t = e._dirtyType, r = e._dirtyElements, i = e._dirtyLeaves;
			p.disconnect(), m = go(a, n, e, t, r, i);
		} catch (t) {
			if (t instanceof Error && e._onError(t), Rc) throw t;
			xl(e, null, r, n), wa(e), e._dirtyType = qi, Rc = !0, Qc(e, a), Rc = !1;
			return;
		} finally {
			p.observe(r, Bc), e._updating = f, Fc = l, Lc = u, Ic = d;
		}
	}
	n._readOnly ||= !0;
	let h = e._dirtyLeaves, g = e._dirtyElements, _ = e._normalizedNodes, v = e._updateTags, y = e._deferred;
	c && (e._dirtyType = Gi, e._cloneNotNeeded.clear(), e._dirtyLeaves = /* @__PURE__ */ new Set(), e._dirtyElements = /* @__PURE__ */ new Map(), e._normalizedNodes = /* @__PURE__ */ new Set(), e._updateTags = /* @__PURE__ */ new Set()), function(e, t) {
		let n = e._decorators, r = e._pendingDecorators || n, i = t._nodeMap, a;
		for (a in r) i.has(a) || (r === n && (r = Zl(e)), delete r[a]);
	}(e, n);
	let b = i ? null : Lu(ku(e));
	if (e._editable && b !== null && (c || s === null || s.dirty) && r !== null && !v.has("skip-dom-selection")) {
		Ic = e, Fc = n;
		try {
			if (p !== null && p.disconnect(), c || s === null || s.dirty) {
				let t = e._blockCursorElement;
				t !== null && Iu(t, e, r), jc(o, s, e, b, v, r);
			}
			(function(e, t, n) {
				let r = e._blockCursorElement;
				if (K(n) && n.isCollapsed() && n.anchor.type === "element" && t.contains(document.activeElement)) {
					let i = n.anchor, a = i.getNode(), o = i.offset, s = !1, c = null;
					if (o === a.getChildrenSize()) Fu(a.getChildAtIndex(o - 1)) && (s = !0);
					else {
						let t = a.getChildAtIndex(o);
						if (t !== null && Fu(t)) {
							let n = t.getPreviousSibling();
							(n === null || Fu(n)) && (s = !0, c = e.getElementByKey(t.__key));
						}
					}
					if (s) {
						let n = e.getElementByKey(a.__key);
						r === null && (e._blockCursorElement = r = function(e) {
							let t = e.theme, n = document.createElement("div");
							n.contentEditable = "false", n.setAttribute("data-lexical-cursor", "true");
							let r = t.blockCursor;
							return r !== void 0 && (typeof r == "string" && (r = t.blockCursor = Bi(r)), r !== void 0 && n.classList.add(...r)), n;
						}(e._config)), t.style.caretColor = "transparent", c === null ? n.appendChild(r) : n.insertBefore(r, c);
						return;
					}
				}
				r !== null && Iu(r, e, t);
			})(e, r, s);
		} finally {
			p !== null && p.observe(r, Bc), Ic = d, Fc = l;
		}
	}
	m !== null && function(e, t, n, r, i) {
		let a = Array.from(e._listeners.mutation), o = a.length;
		for (let e = 0; e < o; e++) {
			let [o, s] = a[e];
			for (let e of s) {
				let a = t.get(e);
				a !== void 0 && o(a, {
					dirtyLeaves: r,
					prevEditorState: i,
					updateTags: n
				});
			}
		}
	}(e, m, v, h, a), K(s) || s === null || o !== null && o.is(s) || e.dispatchCommand(vo, void 0);
	let x = e._pendingDecorators;
	x !== null && (e._decorators = x, e._pendingDecorators = null, $c("decorator", e, !0, x)), function(e, t, n) {
		let r = Ql(t), i = Ql(n);
		r !== i && $c("textcontent", e, !0, i);
	}(e, t || a, n), $c("update", e, !0, {
		dirtyElements: g,
		dirtyLeaves: h,
		editorState: n,
		mutatedNodes: m,
		normalizedNodes: _,
		prevEditorState: t || a,
		tags: v
	}), function(e, t) {
		if (e._deferred = [], t.length !== 0) {
			let n = e._updating;
			e._updating = !0;
			try {
				for (let e = 0; e < t.length; e++) t[e]();
			} finally {
				e._updating = n;
			}
		}
	}(e, y), function(e) {
		let t = e._updates;
		if (t.length !== 0) {
			let n = t.shift();
			if (n) {
				let [t, r] = n;
				tl(e, t, r);
			}
		}
	}(e);
}
function $c(e, t, n, ...r) {
	let i = t._updating;
	t._updating = n;
	try {
		let n = Array.from(t._listeners[e]);
		for (let e = 0; e < n.length; e++) n[e].apply(null, r);
	} finally {
		t._updating = i;
	}
}
function el(e, t) {
	let n = e._updates, r = t || !1;
	for (; n.length !== 0;) {
		let t = n.shift();
		if (t) {
			let [n, i] = t, a = e._pendingEditorState, o;
			i !== void 0 && (o = i.onUpdate, i.skipTransforms && (r = !0), i.discrete && (a === null && V(191), a._flushSync = !0), o && e._deferred.push(o), Yc(e, i.tag)), a == null ? tl(e, n, i) : n();
		}
	}
	return r;
}
function tl(e, t, n) {
	let r = e._updateTags, i, a = !1, o = !1;
	n !== void 0 && (i = n.onUpdate, Yc(e, n.tag), a = n.skipTransforms || !1, o = n.discrete || !1), i && e._deferred.push(i);
	let s = e._editorState, c = e._pendingEditorState, l = !1;
	(c === null || c._readOnly) && (c = e._pendingEditorState = ul(c || s), l = !0), c._flushSync = o;
	let u = Fc, d = Lc, f = Ic, p = e._updating;
	Fc = c, Lc = !1, e._updating = !0, Ic = e;
	let m = e._headless || e.getRootElement() === null;
	Tl(null);
	try {
		l && (m ? s._selection !== null && (c._selection = s._selection.clone()) : c._selection = function(e, t) {
			let n = e.getEditorState()._selection, r = Lu(ku(e));
			return K(n) || n == null ? wc(n, r, e, t) : n.clone();
		}(e, n && n.event || null));
		let r = e._compositionKey;
		t(), a = el(e, a), function(e, t) {
			let n = t.getEditorState()._selection, r = e._selection;
			if (K(r)) {
				let e = r.anchor, t = r.focus, i;
				if (e.type === "text" && (i = e.getNode(), i.selectionTransform(n, r)), t.type === "text") {
					let e = t.getNode();
					i !== e && e.selectionTransform(n, r);
				}
			}
		}(c, e), e._dirtyType !== Gi && (a ? function(e, t) {
			let n = t._dirtyLeaves, r = e._nodeMap;
			for (let e of n) {
				let t = r.get(e);
				G(t) && t.isAttached() && t.isSimpleText() && !t.isUnmergeable() && Fa(t);
			}
		}(c, e) : function(e, t) {
			let n = t._dirtyLeaves, r = t._dirtyElements, i = e._nodeMap, a = Kl(), o = /* @__PURE__ */ new Map(), s = n, c = s.size, l = r, u = l.size;
			for (; c > 0 || u > 0;) {
				if (c > 0) {
					t._dirtyLeaves = /* @__PURE__ */ new Set();
					for (let e of s) {
						let r = i.get(e);
						G(r) && r.isAttached() && r.isSimpleText() && !r.isUnmergeable() && Fa(r), r !== void 0 && Jc(r, a) && qc(t, r, o), n.add(e);
					}
					if (s = t._dirtyLeaves, c = s.size, c > 0) {
						zc++;
						continue;
					}
				}
				t._dirtyLeaves = /* @__PURE__ */ new Set(), t._dirtyElements = /* @__PURE__ */ new Map(), l.delete("root") && l.set("root", !0);
				for (let e of l) {
					let n = e[0], s = e[1];
					if (r.set(n, s), !s) continue;
					let c = i.get(n);
					c !== void 0 && Jc(c, a) && qc(t, c, o);
				}
				s = t._dirtyLeaves, c = s.size, l = t._dirtyElements, u = l.size, zc++;
			}
			t._dirtyLeaves = n, t._dirtyElements = r;
		}(c, e), el(e), function(e, t, n, r) {
			let i = e._nodeMap, a = t._nodeMap, o = [];
			for (let [e] of r) {
				let t = a.get(e);
				t !== void 0 && (t.isAttached() || (Y(t) && pa(t, e, i, a, o, r), i.has(e) || r.delete(e), o.push(e)));
			}
			for (let e of o) a.delete(e);
			for (let e of n) {
				let t = a.get(e);
				t === void 0 || t.isAttached() || (i.has(e) || n.delete(e), a.delete(e));
			}
		}(s, c, e._dirtyLeaves, e._dirtyElements)), r !== e._compositionKey && (c._flushSync = !0);
		let i = c._selection;
		if (K(i)) {
			let e = c._nodeMap, t = i.anchor.key, n = i.focus.key;
			e.get(t) !== void 0 && e.get(n) !== void 0 || V(19);
		} else dc(i) && i._nodes.size === 0 && (c._selection = null);
	} catch (t) {
		t instanceof Error && e._onError(t), e._pendingEditorState = s, e._dirtyType = qi, e._cloneNotNeeded.clear(), e._dirtyLeaves = /* @__PURE__ */ new Set(), e._dirtyElements.clear(), Qc(e);
		return;
	} finally {
		Fc = u, Lc = d, Ic = f, e._updating = p, zc = 0;
	}
	e._dirtyType !== Gi || e._deferred.length > 0 || function(e, t) {
		let n = t.getEditorState()._selection, r = e._selection;
		if (r !== null) {
			if (r.dirty || !r.is(n)) return !0;
		} else if (n !== null) return !0;
		return !1;
	}(c, e) ? c._flushSync ? (c._flushSync = !1, Qc(e)) : l && kl((() => {
		Qc(e);
	})) : (c._flushSync = !1, l && (r.clear(), e._deferred = [], e._pendingEditorState = null));
}
function nl(e, t, n) {
	Ic === e && n === void 0 ? t() : tl(e, t, n);
}
var rl = class e {
	constructor(e, t, n) {
		this.element = e, this.before = t || null, this.after = n || null;
	}
	withBefore(t) {
		return new e(this.element, t, this.after);
	}
	withAfter(t) {
		return new e(this.element, this.before, t);
	}
	withElement(t) {
		return this.element === t ? this : new e(t, this.before, this.after);
	}
	insertChild(e) {
		let t = this.before || this.getManagedLineBreak();
		return t !== null && t.parentElement !== this.element && V(222), this.element.insertBefore(e, t), this;
	}
	removeChild(e) {
		return e.parentElement !== this.element && V(223), this.element.removeChild(e), this;
	}
	replaceChild(e, t) {
		return t.parentElement !== this.element && V(224), this.element.replaceChild(e, t), this;
	}
	getFirstChild() {
		let e = this.after ? this.after.nextSibling : this.element.firstChild;
		return e === this.before || e === this.getManagedLineBreak() ? null : e;
	}
	getManagedLineBreak() {
		return this.element.__lexicalLineBreak || null;
	}
	setManagedLineBreak(e) {
		if (e === null) this.removeManagedLineBreak();
		else {
			let t = e === "decorator" && (Fi || Pi);
			this.insertManagedLineBreak(t);
		}
	}
	removeManagedLineBreak() {
		let e = this.getManagedLineBreak();
		if (e) {
			let t = this.element, n = e.nodeName === "IMG" ? e.nextSibling : null;
			n && t.removeChild(n), t.removeChild(e), t.__lexicalLineBreak = void 0;
		}
	}
	insertManagedLineBreak(e) {
		let t = this.getManagedLineBreak();
		if (t) {
			if (e === (t.nodeName === "IMG")) return;
			this.removeManagedLineBreak();
		}
		let n = this.element, r = this.before, i = document.createElement("br");
		if (n.insertBefore(i, r), e) {
			let e = document.createElement("img");
			e.setAttribute("data-lexical-linebreak", "true"), e.style.cssText = "display: inline !important; border: 0px !important; margin: 0px !important;", e.alt = "", n.insertBefore(e, i), n.__lexicalLineBreak = e;
		} else n.__lexicalLineBreak = i;
	}
	getFirstChildOffset() {
		let e = 0;
		for (let t = this.after; t !== null; t = t.previousSibling) e++;
		return e;
	}
	resolveChildIndex(e, t, n, r) {
		if (n === this.element) {
			let t = this.getFirstChildOffset();
			return [e, Math.min(t + e.getChildrenSize(), Math.max(t, r))];
		}
		let i = il(t, n);
		i.push(r);
		let a = il(t, this.element), o = e.getIndexWithinParent();
		for (let e = 0; e < a.length; e++) {
			let t = i[e], n = a[e];
			if (t === void 0 || t < n) break;
			if (t > n) {
				o += 1;
				break;
			}
		}
		return [e.getParentOrThrow(), o];
	}
};
function il(e, t) {
	let n = [], r = t;
	for (; r !== e && r !== null; r = r.parentNode) {
		let e = 0;
		for (let t = r.previousSibling; t !== null; t = t.previousSibling) e++;
		n.push(e);
	}
	return r !== e && V(225), n.reverse();
}
var al = class extends Ps {
	constructor(e) {
		super(e), this.__first = null, this.__last = null, this.__size = 0, this.__format = 0, this.__style = "", this.__indent = 0, this.__dir = null, this.__textFormat = 0, this.__textStyle = "";
	}
	afterCloneFrom(e) {
		super.afterCloneFrom(e), this.__key === e.__key && (this.__first = e.__first, this.__last = e.__last, this.__size = e.__size), this.__indent = e.__indent, this.__format = e.__format, this.__style = e.__style, this.__dir = e.__dir, this.__textFormat = e.__textFormat, this.__textStyle = e.__textStyle;
	}
	getFormat() {
		return this.getLatest().__format;
	}
	getFormatType() {
		return la[this.getFormat()] || "";
	}
	getStyle() {
		return this.getLatest().__style;
	}
	getIndent() {
		return this.getLatest().__indent;
	}
	getChildren() {
		let e = [], t = this.getFirstChild();
		for (; t !== null;) e.push(t), t = t.getNextSibling();
		return e;
	}
	getChildrenKeys() {
		let e = [], t = this.getFirstChild();
		for (; t !== null;) e.push(t.__key), t = t.getNextSibling();
		return e;
	}
	getChildrenSize() {
		return this.getLatest().__size;
	}
	isEmpty() {
		return this.getChildrenSize() === 0;
	}
	isDirty() {
		let e = J()._dirtyElements;
		return e !== null && e.has(this.__key);
	}
	isLastChild() {
		let e = this.getLatest(), t = this.getParentOrThrow().getLastChild();
		return t !== null && t.is(e);
	}
	getAllTextNodes() {
		let e = [], t = this.getFirstChild();
		for (; t !== null;) {
			if (G(t) && e.push(t), Y(t)) {
				let n = t.getAllTextNodes();
				e.push(...n);
			}
			t = t.getNextSibling();
		}
		return e;
	}
	getFirstDescendant() {
		let e = this.getFirstChild();
		for (; Y(e);) {
			let t = e.getFirstChild();
			if (t === null) break;
			e = t;
		}
		return e;
	}
	getLastDescendant() {
		let e = this.getLastChild();
		for (; Y(e);) {
			let t = e.getLastChild();
			if (t === null) break;
			e = t;
		}
		return e;
	}
	getDescendantByIndex(e) {
		let t = this.getChildren(), n = t.length;
		if (e >= n) {
			let e = t[n - 1];
			return Y(e) && e.getLastDescendant() || e || null;
		}
		let r = t[e];
		return Y(r) && r.getFirstDescendant() || r || null;
	}
	getFirstChild() {
		let e = this.getLatest().__first;
		return e === null ? null : ql(e);
	}
	getFirstChildOrThrow() {
		let e = this.getFirstChild();
		return e === null && V(45, this.__key), e;
	}
	getLastChild() {
		let e = this.getLatest().__last;
		return e === null ? null : ql(e);
	}
	getLastChildOrThrow() {
		let e = this.getLastChild();
		return e === null && V(96, this.__key), e;
	}
	getChildAtIndex(e) {
		let t = this.getChildrenSize(), n, r;
		if (e < t / 2) {
			for (n = this.getFirstChild(), r = 0; n !== null && r <= e;) {
				if (r === e) return n;
				n = n.getNextSibling(), r++;
			}
			return null;
		}
		for (n = this.getLastChild(), r = t - 1; n !== null && r >= e;) {
			if (r === e) return n;
			n = n.getPreviousSibling(), r--;
		}
		return null;
	}
	getTextContent() {
		let e = "", t = this.getChildren(), n = t.length;
		for (let r = 0; r < n; r++) {
			let i = t[r];
			e += i.getTextContent(), Y(i) && r !== n - 1 && !i.isInline() && (e += ia);
		}
		return e;
	}
	getTextContentSize() {
		let e = 0, t = this.getChildren(), n = t.length;
		for (let r = 0; r < n; r++) {
			let i = t[r];
			e += i.getTextContentSize(), Y(i) && r !== n - 1 && !i.isInline() && (e += 2);
		}
		return e;
	}
	getDirection() {
		return this.getLatest().__dir;
	}
	getTextFormat() {
		return this.getLatest().__textFormat;
	}
	hasFormat(e) {
		if (e !== "") {
			let t = ca[e];
			return !!(this.getFormat() & t);
		}
		return !1;
	}
	hasTextFormat(e) {
		let t = oa[e];
		return !!(this.getTextFormat() & t);
	}
	getFormatFlags(e, t) {
		return Vl(this.getLatest().__textFormat, e, t);
	}
	getTextStyle() {
		return this.getLatest().__textStyle;
	}
	select(e, t) {
		Hc();
		let n = q(), r = e, i = t, a = this.getChildrenSize();
		if (!this.canBeEmpty()) {
			if (e === 0 && t === 0) {
				let e = this.getFirstChild();
				if (G(e) || Y(e)) return e.select(0, 0);
			} else if (!(e !== void 0 && e !== a || t !== void 0 && t !== a)) {
				let e = this.getLastChild();
				if (G(e) || Y(e)) return e.select();
			}
		}
		r === void 0 && (r = a), i === void 0 && (i = a);
		let o = this.__key;
		return K(n) ? (n.anchor.set(o, r, "element"), n.focus.set(o, i, "element"), n.dirty = !0, n) : xc(o, r, o, i, "element", "element");
	}
	selectStart() {
		let e = this.getFirstDescendant();
		return e ? e.selectStart() : this.select();
	}
	selectEnd() {
		let e = this.getLastDescendant();
		return e ? e.selectEnd() : this.select();
	}
	clear() {
		let e = this.getWritable();
		return this.getChildren().forEach(((e) => e.remove())), e;
	}
	append(...e) {
		return this.splice(this.getChildrenSize(), 0, e);
	}
	setDirection(e) {
		let t = this.getWritable();
		return t.__dir = e, t;
	}
	setFormat(e) {
		return this.getWritable().__format = e === "" ? 0 : ca[e], this;
	}
	setStyle(e) {
		return this.getWritable().__style = e || "", this;
	}
	setTextFormat(e) {
		let t = this.getWritable();
		return t.__textFormat = e, t;
	}
	setTextStyle(e) {
		let t = this.getWritable();
		return t.__textStyle = e, t;
	}
	setIndent(e) {
		return this.getWritable().__indent = e, this;
	}
	splice(e, t, n) {
		let r = n.length, i = this.getChildrenSize(), a = this.getWritable();
		e + t <= i || V(226, String(e), String(t), String(i));
		let o = a.__key, s = [], c = [], l = this.getChildAtIndex(e + t), u = null, d = i - t + r;
		if (e !== 0) {
			if (e === i) u = this.getLastChild();
			else {
				let t = this.getChildAtIndex(e);
				t !== null && (u = t.getPreviousSibling());
			}
		}
		if (t > 0) {
			let e = u === null ? this.getFirstChild() : u.getNextSibling();
			for (let n = 0; n < t; n++) {
				e === null && V(100);
				let t = e.getNextSibling(), n = e.__key;
				Ul(e.getWritable()), c.push(n), e = t;
			}
		}
		let f = u;
		for (let e = 0; e < r; e++) {
			let t = n[e];
			f !== null && t.is(f) && (u = f = f.getPreviousSibling());
			let r = t.getWritable();
			r.__parent === o && d--, Ul(r);
			let i = t.__key;
			if (f === null) a.__first = i, r.__prev = null;
			else {
				let e = f.getWritable();
				e.__next = i, r.__prev = e.__key;
			}
			t.__key === o && V(76), r.__parent = o, s.push(i), f = t;
		}
		if (e + t === i) f !== null && (f.getWritable().__next = null, a.__last = f.__key);
		else if (l !== null) {
			let e = l.getWritable();
			if (f !== null) {
				let t = f.getWritable();
				e.__prev = f.__key, t.__next = l.__key;
			} else e.__prev = null;
		}
		if (a.__size = d, c.length) {
			let e = q();
			if (K(e)) {
				let t = new Set(c), n = new Set(s), { anchor: r, focus: i } = e;
				ol(r, t, n) && Oc(r, r.getNode(), this, u, l), ol(i, t, n) && Oc(i, i.getNode(), this, u, l), d !== 0 || this.canBeEmpty() || ju(this) || this.remove();
			}
		}
		return a;
	}
	getDOMSlot(e) {
		return new rl(e);
	}
	exportDOM(e) {
		let { element: t } = super.exportDOM(e);
		if (zu(t)) {
			let e = this.getIndent();
			e > 0 && (t.style.paddingInlineStart = 40 * e + "px");
			let n = this.getDirection();
			n && (t.dir = n);
		}
		return { element: t };
	}
	exportJSON() {
		let e = {
			children: [],
			direction: this.getDirection(),
			format: this.getFormatType(),
			indent: this.getIndent(),
			...super.exportJSON()
		}, t = this.getTextFormat(), n = this.getTextStyle();
		return t !== 0 && (e.textFormat = t), n !== "" && (e.textStyle = n), e;
	}
	updateFromJSON(e) {
		return super.updateFromJSON(e).setFormat(e.format).setIndent(e.indent).setDirection(e.direction).setTextFormat(e.textFormat || 0).setTextStyle(e.textStyle || "");
	}
	insertNewAfter(e, t) {
		return null;
	}
	canIndent() {
		return !0;
	}
	collapseAtStart(e) {
		return !1;
	}
	excludeFromCopy(e) {
		return !1;
	}
	canReplaceWith(e) {
		return !0;
	}
	canInsertAfter(e) {
		return !0;
	}
	canBeEmpty() {
		return !0;
	}
	canInsertTextBefore() {
		return !0;
	}
	canInsertTextAfter() {
		return !0;
	}
	isInline() {
		return !1;
	}
	isShadowRoot() {
		return !1;
	}
	canMergeWith(e) {
		return !1;
	}
	extractWithChild(e, t, n) {
		return !1;
	}
	canMergeWhenEmpty() {
		return !1;
	}
	reconcileObservedMutation(e, t) {
		let n = this.getDOMSlot(e), r = n.getFirstChild();
		for (let e = this.getFirstChild(); e; e = e.getNextSibling()) {
			let i = t.getElementByKey(e.getKey());
			i !== null && (r == null ? (n.insertChild(i), r = i) : r !== i && n.replaceChild(i, r), r = r.nextSibling);
		}
	}
};
function Y(e) {
	return e instanceof al;
}
function ol(e, t, n) {
	let r = e.getNode();
	for (; r;) {
		let e = r.__key;
		if (t.has(e) && !n.has(e)) return !0;
		r = r.getParent();
	}
	return !1;
}
var sl = class extends Ps {
	decorate(e, t) {
		V(47);
	}
	isIsolated() {
		return !1;
	}
	isInline() {
		return !0;
	}
	isKeyboardSelectable() {
		return !0;
	}
};
function X(e) {
	return e instanceof sl;
}
var cl = class e extends al {
	static getType() {
		return "root";
	}
	static clone() {
		return new e();
	}
	constructor() {
		super("root"), this.__cachedText = null;
	}
	getTopLevelElementOrThrow() {
		V(51);
	}
	getTextContent() {
		let e = this.__cachedText;
		return !Vc() && J()._dirtyType !== Gi || e === null ? super.getTextContent() : e;
	}
	remove() {
		V(52);
	}
	replace(e) {
		V(53);
	}
	insertBefore(e) {
		V(54);
	}
	insertAfter(e) {
		V(55);
	}
	updateDOM(e, t) {
		return !1;
	}
	splice(e, t, n) {
		for (let e of n) Y(e) || X(e) || V(282);
		return super.splice(e, t, n);
	}
	static importJSON(e) {
		return Z().updateFromJSON(e);
	}
	collapseAtStart() {
		return !0;
	}
};
function ll(e) {
	return e instanceof cl;
}
function ul(e) {
	return new pl(new Map(e._nodeMap));
}
function dl() {
	return new pl(/* @__PURE__ */ new Map([["root", new cl()]]));
}
function fl(e) {
	let t = e.exportJSON(), n = e.constructor;
	if (t.type !== n.getType() && V(130, n.name), Y(e)) {
		let r = t.children;
		Array.isArray(r) || V(59, n.name);
		let i = e.getChildren();
		for (let e = 0; e < i.length; e++) {
			let t = fl(i[e]);
			r.push(t);
		}
	}
	return t;
}
var pl = class e {
	constructor(e, t) {
		this._nodeMap = e, this._selection = t || null, this._flushSync = !1, this._readOnly = !1;
	}
	isEmpty() {
		return this._nodeMap.size === 1 && this._selection === null;
	}
	read(e, t) {
		return Zc(t && t.editor || null, this, e);
	}
	clone(t) {
		let n = new e(this._nodeMap, t === void 0 ? this._selection : t);
		return n._readOnly = !0, n;
	}
	toJSON() {
		return Zc(null, this, (() => ({ root: fl(Z()) })));
	}
}, ml = "history-push", hl = "history-merge", gl = class extends al {
	static getType() {
		return "artificial";
	}
	createDOM(e) {
		return document.createElement("div");
	}
}, _l = class e extends al {
	static getType() {
		return "paragraph";
	}
	static clone(t) {
		return new e(t.__key);
	}
	createDOM(e) {
		let t = document.createElement("p"), n = gu(e.theme, "paragraph");
		return n !== void 0 && t.classList.add(...n), t;
	}
	updateDOM(e, t, n) {
		return !1;
	}
	static importDOM() {
		return { p: (e) => ({
			conversion: vl,
			priority: 0
		}) };
	}
	exportDOM(e) {
		let { element: t } = super.exportDOM(e);
		if (zu(t)) {
			this.isEmpty() && t.append(document.createElement("br"));
			let e = this.getFormatType();
			e && (t.style.textAlign = e);
		}
		return { element: t };
	}
	static importJSON(e) {
		return yl().updateFromJSON(e);
	}
	exportJSON() {
		return {
			...super.exportJSON(),
			textFormat: this.getTextFormat(),
			textStyle: this.getTextStyle()
		};
	}
	insertNewAfter(e, t) {
		let n = yl();
		n.setTextFormat(e.format), n.setTextStyle(e.style);
		let r = this.getDirection();
		return n.setDirection(r), n.setFormat(this.getFormatType()), n.setStyle(this.getStyle()), this.insertAfter(n, t), n;
	}
	collapseAtStart() {
		let e = this.getChildren();
		if (e.length === 0 || G(e[0]) && e[0].getTextContent().trim() === "") {
			if (this.getNextSibling() !== null) return this.selectNext(), this.remove(), !0;
			if (this.getPreviousSibling() !== null) return this.selectPrevious(), this.remove(), !0;
		}
		return !1;
	}
};
function vl(e) {
	let t = yl();
	return e.style && (t.setFormat(e.style.textAlign), Zu(e, t)), { node: t };
}
function yl() {
	return Mu(new _l());
}
function bl(e) {
	return e instanceof _l;
}
function xl(e, t, n, r) {
	let i = e._keyToDOMMap;
	i.clear(), e._editorState = dl(), e._pendingEditorState = r, e._compositionKey = null, e._dirtyType = Gi, e._cloneNotNeeded.clear(), e._dirtyLeaves = /* @__PURE__ */ new Set(), e._dirtyElements.clear(), e._normalizedNodes = /* @__PURE__ */ new Set(), e._updateTags = /* @__PURE__ */ new Set(), e._updates = [], e._blockCursorElement = null;
	let a = e._observer;
	a !== null && (a.disconnect(), e._observer = null), t !== null && (t.textContent = ""), n !== null && (n.textContent = "", i.set("root", n));
}
function Sl(e) {
	let t = e || {}, n = Kc(), r = t.theme || {}, i = e === void 0 ? n : t.parentEditor || null, a = t.disableEvents || !1, o = dl(), s = t.namespace || (i === null ? iu() : i._config.namespace), c = t.editorState, l = [
		cl,
		Ks,
		Fs,
		nc,
		_l,
		gl,
		...t.nodes || []
	], { onError: u, html: d } = t, f = t.editable === void 0 || t.editable, p;
	if (e === void 0 && n !== null) p = n._nodes;
	else {
		p = /* @__PURE__ */ new Map();
		for (let e = 0; e < l.length; e++) {
			let t = l[e], n = null, r = null;
			if (typeof t != "function") {
				let e = t;
				t = e.replace, n = e.with, r = e.withKlass || null;
			}
			let { ownNodeConfig: i } = ed(t), a = t.getType(), o = t.transform(), s = /* @__PURE__ */ new Set();
			i && i.$transform && s.add(i.$transform), o !== null && s.add(o), p.set(a, {
				exportDOM: d && d.export ? d.export.get(t) : void 0,
				klass: t,
				replace: n,
				replaceWithKlass: r,
				sharedNodeState: Ta(l[e]),
				transforms: s
			});
		}
	}
	let m = new Cl(o, i, p, {
		disableEvents: a,
		namespace: s,
		theme: r
	}, u || console.error, function(e, t) {
		let n = /* @__PURE__ */ new Map(), r = /* @__PURE__ */ new Set(), i = (e) => {
			Object.keys(e).forEach(((t) => {
				let r = n.get(t);
				r === void 0 && (r = [], n.set(t, r)), r.push(e[t]);
			}));
		};
		return e.forEach(((e) => {
			let t = e.klass.importDOM;
			if (t == null || r.has(t)) return;
			r.add(t);
			let n = t.call(e.klass);
			n !== null && i(n);
		})), t && i(t), n;
	}(p, d ? d.import : void 0), f, e);
	return c !== void 0 && (m._pendingEditorState = c, m._dirtyType = qi), m;
}
var Cl = class {
	constructor(e, t, n, r, i, a, o, s) {
		this._createEditorArgs = s, this._parentEditor = t, this._rootElement = null, this._editorState = e, this._pendingEditorState = null, this._compositionKey = null, this._deferred = [], this._keyToDOMMap = /* @__PURE__ */ new Map(), this._updates = [], this._updating = !1, this._listeners = {
			decorator: /* @__PURE__ */ new Set(),
			editable: /* @__PURE__ */ new Set(),
			mutation: /* @__PURE__ */ new Map(),
			root: /* @__PURE__ */ new Set(),
			textcontent: /* @__PURE__ */ new Set(),
			update: /* @__PURE__ */ new Set()
		}, this._commands = /* @__PURE__ */ new Map(), this._config = r, this._nodes = n, this._decorators = {}, this._pendingDecorators = null, this._dirtyType = Gi, this._cloneNotNeeded = /* @__PURE__ */ new Set(), this._dirtyLeaves = /* @__PURE__ */ new Set(), this._dirtyElements = /* @__PURE__ */ new Map(), this._normalizedNodes = /* @__PURE__ */ new Set(), this._updateTags = /* @__PURE__ */ new Set(), this._observer = null, this._key = iu(), this._onError = i, this._htmlConversions = a, this._editable = o, this._headless = t !== null && t._headless, this._window = null, this._blockCursorElement = null;
	}
	isComposing() {
		return this._compositionKey != null;
	}
	registerUpdateListener(e) {
		let t = this._listeners.update;
		return t.add(e), () => {
			t.delete(e);
		};
	}
	registerEditableListener(e) {
		let t = this._listeners.editable;
		return t.add(e), () => {
			t.delete(e);
		};
	}
	registerDecoratorListener(e) {
		let t = this._listeners.decorator;
		return t.add(e), () => {
			t.delete(e);
		};
	}
	registerTextContentListener(e) {
		let t = this._listeners.textcontent;
		return t.add(e), () => {
			t.delete(e);
		};
	}
	registerRootListener(e) {
		let t = this._listeners.root;
		return e(this._rootElement, null), t.add(e), () => {
			e(null, this._rootElement), t.delete(e);
		};
	}
	registerCommand(e, t, n) {
		n === void 0 && V(35);
		let r = this._commands;
		r.has(e) || r.set(e, [
			/* @__PURE__ */ new Set(),
			/* @__PURE__ */ new Set(),
			/* @__PURE__ */ new Set(),
			/* @__PURE__ */ new Set(),
			/* @__PURE__ */ new Set()
		]);
		let i = r.get(e);
		i === void 0 && V(36, String(e));
		let a = i[n];
		return a.add(t), () => {
			a.delete(t), i.every(((e) => e.size === 0)) && r.delete(e);
		};
	}
	registerMutationListener(e, t, n) {
		let r = this.resolveRegisteredNodeAfterReplacements(this.getRegisteredNode(e)).klass, i = this._listeners.mutation, a = i.get(t);
		a === void 0 && (a = /* @__PURE__ */ new Set(), i.set(t, a)), a.add(r);
		let o = n && n.skipInitialization;
		return o !== void 0 && o || this.initializeMutationListener(t, r), () => {
			a.delete(r), a.size === 0 && i.delete(t);
		};
	}
	getRegisteredNode(e) {
		let t = this._nodes.get(e.getType());
		return t === void 0 && V(37, e.name), t;
	}
	resolveRegisteredNodeAfterReplacements(e) {
		for (; e.replaceWithKlass;) e = this.getRegisteredNode(e.replaceWithKlass);
		return e;
	}
	initializeMutationListener(e, t) {
		let n = this._editorState, r = Yu(n).get(t.getType());
		if (!r) return;
		let i = /* @__PURE__ */ new Map();
		for (let e of r.keys()) i.set(e, "created");
		i.size > 0 && e(i, {
			dirtyLeaves: /* @__PURE__ */ new Set(),
			prevEditorState: n,
			updateTags: /* @__PURE__ */ new Set(["registerMutationListener"])
		});
	}
	registerNodeTransformToKlass(e, t) {
		let n = this.getRegisteredNode(e);
		return n.transforms.add(t), n;
	}
	registerNodeTransform(e, t) {
		let n = this.registerNodeTransformToKlass(e, t), r = [n], i = n.replaceWithKlass;
		if (i != null) {
			let e = this.registerNodeTransformToKlass(i, t);
			r.push(e);
		}
		return function(e, t) {
			let n = Yu(e.getEditorState()), r = [];
			for (let e of t) {
				let t = n.get(e);
				t && r.push(t);
			}
			r.length !== 0 && e.update((() => {
				for (let e of r) for (let t of e.keys()) {
					let e = ql(t);
					e && e.markDirty();
				}
			}), e._pendingEditorState === null ? { tag: hl } : void 0);
		}(this, r.map(((e) => e.klass.getType()))), () => {
			r.forEach(((e) => e.transforms.delete(t)));
		};
	}
	hasNode(e) {
		return this._nodes.has(e.getType());
	}
	hasNodes(e) {
		return e.every(this.hasNode.bind(this));
	}
	dispatchCommand(e, t) {
		return $(this, e, t);
	}
	getDecorators() {
		return this._decorators;
	}
	getRootElement() {
		return this._rootElement;
	}
	getKey() {
		return this._key;
	}
	setRootElement(e) {
		let t = this._rootElement;
		if (e !== t) {
			let n = gu(this._config.theme, "root"), r = this._pendingEditorState || this._editorState;
			if (this._rootElement = e, xl(this, t, e, r), t !== null && (this._config.disableEvents || Ms(t), n != null && t.classList.remove(...n)), e !== null) {
				let t = Ou(e), r = e.style;
				r.userSelect = "text", r.whiteSpace = "pre-wrap", r.wordBreak = "break-word", e.setAttribute("data-lexical-editor", "true"), this._window = t, this._dirtyType = qi, wa(this), this._updateTags.add(hl), Qc(this), this._config.disableEvents || function(e, t) {
					let n = e.ownerDocument, r = ds.get(n);
					(r === void 0 || r < 1) && n.addEventListener("selectionchange", ks), ds.set(n, (r || 0) + 1), e.__lexicalEditor = t;
					let i = Ds(e);
					for (let n = 0; n < os.length; n++) {
						let [r, a] = os[n], o = typeof a == "function" ? (e) => {
							js(e) || (As(e), (t.isEditable() || r === "click") && a(e, t));
						} : (e) => {
							if (js(e)) return;
							As(e);
							let n = t.isEditable();
							switch (r) {
								case "cut": return n && $(t, Yo, e);
								case "copy": return $(t, Jo, e);
								case "paste": return n && $(t, wo, e);
								case "dragstart": return n && $(t, Go, e);
								case "dragover": return n && $(t, Ko, e);
								case "dragend": return n && $(t, qo, e);
								case "focus": return n && $(t, ts, e);
								case "blur": return n && $(t, ns, e);
								case "drop": return n && $(t, Wo, e);
							}
						};
						e.addEventListener(r, o), i.push((() => {
							e.removeEventListener(r, o);
						}));
					}
				}(e, this), n != null && e.classList.add(...n);
			} else this._window = null, this._updateTags.add(hl), Qc(this);
			$c("root", this, !1, e, t);
		}
	}
	getElementByKey(e) {
		return this._keyToDOMMap.get(e) || null;
	}
	getEditorState() {
		return this._editorState;
	}
	setEditorState(e, t) {
		e.isEmpty() && V(38);
		let n = e;
		n._readOnly && (n = ul(e), n._selection = e._selection ? e._selection.clone() : null), Ca(this);
		let r = this._pendingEditorState, i = this._updateTags, a = t === void 0 ? null : t.tag;
		r === null || r.isEmpty() || (a != null && i.add(a), Qc(this)), this._pendingEditorState = n, this._dirtyType = qi, this._dirtyElements.set("root", !1), this._compositionKey = null, a != null && i.add(a), this._updating || Qc(this);
	}
	parseEditorState(e, t) {
		return function(e, t, n) {
			let r = dl(), i = Fc, a = Lc, o = Ic, s = t._dirtyElements, c = t._dirtyLeaves, l = t._cloneNotNeeded, u = t._dirtyType;
			t._dirtyElements = /* @__PURE__ */ new Map(), t._dirtyLeaves = /* @__PURE__ */ new Set(), t._cloneNotNeeded = /* @__PURE__ */ new Set(), t._dirtyType = 0, Fc = r, Lc = !1, Ic = t, Tl(null);
			try {
				let i = t._nodes;
				Xc(e.root, i), n && n(), r._readOnly = !0;
			} catch (e) {
				e instanceof Error && t._onError(e);
			} finally {
				t._dirtyElements = s, t._dirtyLeaves = c, t._cloneNotNeeded = l, t._dirtyType = u, Fc = i, Lc = a, Ic = o;
			}
			return r;
		}(typeof e == "string" ? JSON.parse(e) : e, this, t);
	}
	read(e) {
		return Qc(this), this.getEditorState().read(e, { editor: this });
	}
	update(e, t) {
		(function(e, t, n) {
			e._updating ? e._updates.push([t, n]) : tl(e, t, n);
		})(this, e, t);
	}
	focus(e, t = {}) {
		let n = this._rootElement;
		n !== null && (n.setAttribute("autocapitalize", "off"), nl(this, (() => {
			let r = q(), i = Z();
			r === null ? i.getChildrenSize() !== 0 && (t.defaultSelection === "rootStart" ? i.selectStart() : i.selectEnd()) : r.dirty || eu(r.clone()), Tu("focus"), Eu((() => {
				n.removeAttribute("autocapitalize"), e && e();
			}));
		})), this._pendingEditorState === null && n.removeAttribute("autocapitalize"));
	}
	blur() {
		let e = this._rootElement;
		e !== null && e.blur();
		let t = Lu(this._window);
		t !== null && t.removeAllRanges();
	}
	isEditable() {
		return this._editable;
	}
	setEditable(e) {
		this._editable !== e && (this._editable = e, $c("editable", this, !0, e));
	}
	toJSON() {
		return { editorState: this._editorState.toJSON() };
	}
};
Cl.version = "0.35.0+prod.esm";
var wl = null;
function Tl(e) {
	wl = e;
}
var El = 1;
function Dl(e, t) {
	let n = Ol(e, t);
	return n === void 0 && V(30, t), n;
}
function Ol(e, t) {
	return e._nodes.get(t);
}
var kl = typeof queueMicrotask == "function" ? queueMicrotask : (e) => {
	Promise.resolve().then(e);
};
function Al(e) {
	return X(Xl(e));
}
function jl(e) {
	let t = document.activeElement;
	if (!zu(t)) return !1;
	let n = t.nodeName;
	return X(Xl(e)) && (n === "INPUT" || n === "TEXTAREA" || t.contentEditable === "true" && Fl(t) == null);
}
function Ml(e, t, n) {
	let r = e.getRootElement();
	try {
		return r !== null && r.contains(t) && r.contains(n) && t !== null && !jl(t) && Pl(t) === e;
	} catch {
		return !1;
	}
}
function Nl(e) {
	return e instanceof Cl;
}
function Pl(e) {
	let t = e;
	for (; t != null;) {
		let e = Fl(t);
		if (Nl(e)) return e;
		t = Cu(t);
	}
	return null;
}
function Fl(e) {
	return e ? e.__lexicalEditor : null;
}
function Il(e) {
	return ic(e) || e.isToken();
}
function Ll(e) {
	return Il(e) || e.isSegmented();
}
function Rl(e) {
	return Bu(e) && e.nodeType === Hi;
}
function zl(e) {
	return Bu(e) && e.nodeType === Ui;
}
function Bl(e) {
	let t = e;
	for (; t != null;) {
		if (Rl(t)) return t;
		t = t.firstChild;
	}
	return null;
}
function Vl(e, t, n) {
	let r = oa[t];
	if (n !== null && (e & r) == (n & r)) return e;
	let i = e ^ r;
	return t === "subscript" ? i &= ~oa.superscript : t === "superscript" ? i &= ~oa.subscript : t === "lowercase" ? (i &= ~oa.uppercase, i &= ~oa.capitalize) : t === "uppercase" ? (i &= ~oa.lowercase, i &= ~oa.capitalize) : t === "capitalize" && (i &= ~oa.lowercase, i &= ~oa.uppercase), i;
}
function Hl(e, t) {
	let n = function() {
		let e = wl;
		return wl = null, e;
	}();
	if ((t ||= n && n.__key) != null) return void (e.__key = t);
	Hc(), Uc();
	let r = J(), i = Wc(), a = "" + El++;
	i._nodeMap.set(a, e), Y(e) ? r._dirtyElements.set(a, !0) : r._dirtyLeaves.add(a), r._cloneNotNeeded.add(a), r._dirtyType = Ki, e.__key = a;
}
function Ul(e) {
	let t = e.getParent();
	if (t !== null) {
		let n = e.getWritable(), r = t.getWritable(), i = e.getPreviousSibling(), a = e.getNextSibling(), o = a === null ? null : a.__key, s = i === null ? null : i.__key, c = i === null ? null : i.getWritable(), l = a === null ? null : a.getWritable();
		i === null && (r.__first = o), a === null && (r.__last = s), c !== null && (c.__next = o), l !== null && (l.__prev = s), n.__prev = null, n.__next = null, n.__parent = null, r.__size--;
	}
}
function Wl(e) {
	Uc();
	let t = e.getLatest(), n = t.__parent, r = Wc(), i = J(), a = r._nodeMap, o = i._dirtyElements;
	n !== null && function(e, t, n) {
		let r = e;
		for (; r !== null;) {
			if (n.has(r)) return;
			let e = t.get(r);
			if (e === void 0) break;
			n.set(r, !1), r = e.__parent;
		}
	}(n, a, o);
	let s = t.__key;
	i._dirtyType = Ki, Y(e) ? o.set(s, !0) : i._dirtyLeaves.add(s);
}
function Gl(e) {
	Hc();
	let t = J(), n = t._compositionKey;
	if (e !== n) {
		if (t._compositionKey = e, n !== null) {
			let e = ql(n);
			e !== null && e.getWritable();
		}
		if (e !== null) {
			let t = ql(e);
			t !== null && t.getWritable();
		}
	}
}
function Kl() {
	return Vc() ? null : J()._compositionKey;
}
function ql(e, t) {
	let n = (t || Wc())._nodeMap.get(e);
	return n === void 0 ? null : n;
}
function Jl(e, t) {
	let n = Yl(e, J());
	return n === void 0 ? null : ql(n, t);
}
function Yl(e, t) {
	return e[`__lexicalKey_${t._key}`];
}
function Xl(e, t) {
	let n = e;
	for (; n != null;) {
		let e = Jl(n, t);
		if (e !== null) return e;
		n = Cu(n);
	}
	return null;
}
function Zl(e) {
	let t = e._decorators, n = Object.assign({}, t);
	return e._pendingDecorators = n, n;
}
function Ql(e) {
	return e.read((() => Z().getTextContent()));
}
function Z() {
	return $l(Wc());
}
function $l(e) {
	return e._nodeMap.get("root");
}
function eu(e) {
	Hc();
	let t = Wc();
	e !== null && (e.dirty = !0, e.setCachedNodes(null)), t._selection = e;
}
function tu(e) {
	let t = J(), n = function(e, t) {
		let n = e;
		for (; n != null;) {
			let e = Yl(n, t);
			if (e !== void 0) return e;
			n = Cu(n);
		}
		return null;
	}(e, t);
	return n === null ? e === t.getRootElement() ? ql("root") : null : ql(n);
}
function nu(e) {
	return /[\uD800-\uDBFF][\uDC00-\uDFFF]/g.test(e);
}
function ru(e) {
	let t = [], n = e;
	for (; n !== null;) t.push(n), n = n._parentEditor;
	return t;
}
function iu() {
	return Math.random().toString(36).replace(/[^a-z]+/g, "").substring(0, 5);
}
function au(e) {
	return Rl(e) ? e.nodeValue : null;
}
function ou(e, t, n) {
	let r = Lu(ku(t));
	if (r === null) return;
	let i = r.anchorNode, { anchorOffset: a, focusOffset: o } = r;
	if (i !== null) {
		let t = au(i), r = Xl(i);
		if (t !== null && G(r)) {
			if (t === ra && n) {
				let e = n.length;
				t = n, a = e, o = e;
			}
			t !== null && su(r, t, a, o, e);
		}
	}
}
function su(e, t, n, r, i) {
	let a = e;
	if (a.isAttached() && (i || !a.isDirty())) {
		let o = a.isComposing(), s = t;
		(o || i) && t[t.length - 1] === ra && (s = t.slice(0, -1));
		let c = a.getTextContent();
		if (i || s !== c) {
			if (s === "") {
				if (Gl(null), Pi || Fi || zi) a.remove();
				else {
					let e = J();
					setTimeout((() => {
						e.update((() => {
							a.isAttached() && a.remove();
						}));
					}), 20);
				}
				return;
			}
			let t = a.getParent(), i = Tc(), c = a.getTextContentSize(), l = Kl(), u = a.getKey();
			if (a.isToken() || l !== null && u === l && !o || K(i) && (t !== null && !t.canInsertTextBefore() && i.anchor.offset === 0 || i.anchor.key === e.__key && i.anchor.offset === 0 && !a.canInsertTextBefore() && !o || i.focus.key === e.__key && i.focus.offset === c && !a.canInsertTextAfter() && !o)) return void a.markDirty();
			let d = q();
			if (!K(d) || n === null || r === null) return void cu(a, s, d);
			if (d.setTextNodeRange(a, n, a, r), a.isSegmented()) {
				let e = W(a.getTextContent());
				a.replace(e), a = e;
			}
			cu(a, s, d);
		}
	}
}
function cu(e, t, n) {
	if (e.setTextContent(t), K(n)) {
		let t = e.getKey();
		for (let r of ["anchor", "focus"]) {
			let i = n[r];
			i.type === "text" && i.key === t && (i.offset = _d(e, i.offset, "clamp"));
		}
	}
}
function lu(e, t, n) {
	let r = t[n] || !1;
	return r === "any" || r === e[n];
}
function uu(e, t) {
	return lu(e, t, "altKey") && lu(e, t, "ctrlKey") && lu(e, t, "shiftKey") && lu(e, t, "metaKey");
}
function Q(e, t, n) {
	return uu(e, n) && e.key.toLowerCase() === t.toLowerCase();
}
var du = {
	ctrlKey: !ji,
	metaKey: ji
}, fu = {
	altKey: ji,
	ctrlKey: !ji
};
function pu(e) {
	return e.key === "Backspace";
}
function mu(e) {
	return Q(e, "a", du);
}
function hu(e) {
	let t = Z();
	if (K(e)) {
		let t = e.anchor, n = e.focus, r = t.getNode().getTopLevelElementOrThrow().getParentOrThrow();
		return t.set(r.getKey(), 0, "element"), n.set(r.getKey(), r.getChildrenSize(), "element"), Ia(e), e;
	}
	{
		let e = t.select(0, t.getChildrenSize());
		return eu(Ia(e)), e;
	}
}
function gu(e, t) {
	e.__lexicalClassNameCache === void 0 && (e.__lexicalClassNameCache = {});
	let n = e.__lexicalClassNameCache, r = n[t];
	if (r !== void 0) return r;
	let i = e[t];
	if (typeof i == "string") {
		let e = Bi(i);
		return n[t] = e, e;
	}
	return i;
}
function _u(e, t, n, r, i) {
	if (n.size === 0) return;
	let a = r.__type, o = r.__key, s = t.get(a);
	s === void 0 && V(33, a);
	let c = s.klass, l = e.get(c);
	l === void 0 && (l = /* @__PURE__ */ new Map(), e.set(c, l));
	let u = l.get(o), d = u === "destroyed" && i === "created";
	(u === void 0 || d) && l.set(o, d ? "updated" : i);
}
function vu(e, t, n) {
	let r = e.getParent(), i = n, a = e;
	return r !== null && (t && n === 0 ? (i = a.getIndexWithinParent(), a = r) : t || n !== a.getChildrenSize() || (i = a.getIndexWithinParent() + 1, a = r)), a.getChildAtIndex(t ? i - 1 : i);
}
function yu(e, t) {
	let n = e.offset;
	if (e.type === "element") return vu(e.getNode(), t, n);
	{
		let r = e.getNode();
		if (t && n === 0 || !t && n === r.getTextContentSize()) {
			let e = t ? r.getPreviousSibling() : r.getNextSibling();
			return e === null ? vu(r.getParentOrThrow(), t, r.getIndexWithinParent() + +!t) : e;
		}
	}
	return null;
}
function bu(e) {
	let t = ku(e).event, n = t && t.inputType;
	return n === "insertFromPaste" || n === "insertFromPasteAsQuotation";
}
function $(e, t, n) {
	return function(e, t, n) {
		let r = ru(e);
		for (let i = 4; i >= 0; i--) for (let a = 0; a < r.length; a++) {
			let o = r[a], s = o._commands.get(t);
			if (s !== void 0) {
				let t = s[i];
				if (t !== void 0) {
					let r = Array.from(t), i = r.length, a = !1;
					if (nl(o, (() => {
						for (let t = 0; t < i; t++) if (r[t](n, e)) return void (a = !0);
					})), a) return a;
				}
			}
		}
		return !1;
	}(e, t, n);
}
function xu(e) {
	return !ll(e) && !e.isLastChild() && !e.isInline();
}
function Su(e, t) {
	let n = e._keyToDOMMap.get(t);
	return n === void 0 && V(75, t), n;
}
function Cu(e) {
	let t = e.assignedSlot || e.parentElement;
	return Vu(t) ? t.host : t;
}
function wu(e) {
	return zl(e) ? e : zu(e) ? e.ownerDocument : null;
}
function Tu(e) {
	Hc(), J()._updateTags.add(e);
}
function Eu(e) {
	Hc(), J()._deferred.push(e);
}
function Du(e, t) {
	let n = e.getParent();
	for (; n !== null;) {
		if (n.is(t)) return !0;
		n = n.getParent();
	}
	return !1;
}
function Ou(e) {
	let t = wu(e);
	return t ? t.defaultView : null;
}
function ku(e) {
	let t = e._window;
	return t === null && V(78), t;
}
function Au(e) {
	let t = e.getParentOrThrow();
	for (; t !== null;) {
		if (ju(t)) return t;
		t = t.getParentOrThrow();
	}
	return t;
}
function ju(e) {
	return ll(e) || Y(e) && e.isShadowRoot();
}
function Mu(e) {
	let t = J(), n = e.getType(), r = Ol(t, n);
	r === void 0 && V(200, e.constructor.name, n);
	let { replace: i, replaceWithKlass: a } = r;
	if (i !== null) {
		let t = i(e), r = t.constructor;
		return a === null ? t instanceof e.constructor && r !== e.constructor || V(202, r.name, r.getType(), e.constructor.name, n) : t instanceof a || V(201, a.name, a.getType(), r.name, r.getType(), e.constructor.name, n), t.__key === e.__key && V(203, e.constructor.name, n, r.name, r.getType()), t;
	}
	return e;
}
function Nu(e, t) {
	!ll(e.getParent()) || Y(t) || X(t) || V(99);
}
function Pu(e) {
	let t = ql(e);
	return t === null && V(63, e), t;
}
function Fu(e) {
	return (X(e) || Y(e) && !e.canBeEmpty()) && !e.isInline();
}
function Iu(e, t, n) {
	n.style.removeProperty("caret-color"), t._blockCursorElement = null;
	let r = e.parentElement;
	r !== null && r.removeChild(e);
}
function Lu(e) {
	return ki ? (e || window).getSelection() : null;
}
function Ru(e) {
	let t = Ou(e);
	return t ? t.getSelection() : null;
}
function zu(e) {
	return Bu(e) && e.nodeType === Vi;
}
function Bu(e) {
	return typeof e == "object" && !!e && "nodeType" in e && typeof e.nodeType == "number";
}
function Vu(e) {
	return Bu(e) && e.nodeType === Wi;
}
function Hu(e) {
	let t = /* @__PURE__ */ new RegExp(/^(a|abbr|acronym|b|cite|code|del|em|i|ins|kbd|label|mark|output|q|ruby|s|samp|span|strong|sub|sup|time|u|tt|var|#text)$/, "i");
	return e.nodeName.match(t) !== null;
}
function Uu(e) {
	let t = /* @__PURE__ */ new RegExp(/^(address|article|aside|blockquote|canvas|dd|div|dl|dt|fieldset|figcaption|figure|footer|form|h1|h2|h3|h4|h5|h6|header|hr|li|main|nav|noscript|ol|p|pre|section|table|td|tfoot|ul|video)$/, "i");
	return e.nodeName.match(t) !== null;
}
function Wu(e) {
	if (X(e) && !e.isInline()) return !0;
	if (!Y(e) || ju(e)) return !1;
	let t = e.getFirstChild(), n = t === null || Rs(t) || G(t) || t.isInline();
	return !e.isInline() && !1 !== e.canBeEmpty() && n;
}
function Gu(e, t) {
	let n = e;
	for (; n !== null && n.getParent() !== null && !t(n);) n = n.getParentOrThrow();
	return t(n) ? n : null;
}
function Ku() {
	return J();
}
var qu = /* @__PURE__ */ new WeakMap(), Ju = /* @__PURE__ */ new Map();
function Yu(e) {
	if (!e._readOnly && e.isEmpty()) return Ju;
	e._readOnly || V(192);
	let t = qu.get(e);
	return t || (t = function(e) {
		let t = /* @__PURE__ */ new Map();
		for (let [n, r] of e._nodeMap) {
			let e = r.__type, i = t.get(e);
			i || (i = /* @__PURE__ */ new Map(), t.set(e, i)), i.set(n, r);
		}
		return t;
	}(e), qu.set(e, t)), t;
}
function Xu(e) {
	let t = e.constructor.clone(e);
	return t.afterCloneFrom(e), t;
}
function Zu(e, t) {
	let n = parseInt(e.style.paddingInlineStart, 10) || 0, r = Math.round(n / 40);
	t.setIndent(r);
}
function Qu(e) {
	return !0 === e.__lexicalUnmanaged;
}
function $u(e, t) {
	return function(e, t) {
		return Object.prototype.hasOwnProperty.call(e, t);
	}(e, t) && e[t] !== Ps[t];
}
function ed(e) {
	let t = fa in e.prototype ? e.prototype[fa]() : void 0, n = function(e) {
		return e === sl || e === al || e === Ps;
	}(e), r = !n && $u(e, "getType") ? e.getType() : void 0, i, a = r;
	if (t) {
		if (r) i = t[r];
		else for (let [e, n] of Object.entries(t)) a = e, i = n;
	}
	if (!n && a && ($u(e, "getType") || (e.getType = () => a), $u(e, "clone") || (e.clone = (t) => (Tl(t), new e())), $u(e, "importJSON") || (e.importJSON = i && i.$importJSON || ((t) => new e().updateFromJSON(t))), !$u(e, "importDOM") && i)) {
		let { importDOM: t } = i;
		t && (e.importDOM = () => t);
	}
	return {
		ownNodeConfig: i,
		ownNodeType: a
	};
}
var td = {
	next: "previous",
	previous: "next"
}, nd = class {
	constructor(e) {
		this.origin = e;
	}
	[Symbol.iterator]() {
		return Dd({
			hasNext: ud,
			initial: this.getAdjacentCaret(),
			map: (e) => e,
			step: (e) => e.getAdjacentCaret()
		});
	}
	getAdjacentCaret() {
		return hd(this.getNodeAtCaret(), this.direction);
	}
	getSiblingCaret() {
		return hd(this.origin, this.direction);
	}
	remove() {
		let e = this.getNodeAtCaret();
		return e && e.remove(), this;
	}
	replaceOrInsert(e, t) {
		let n = this.getNodeAtCaret();
		return e.is(this.origin) || e.is(n) || (n === null ? this.insert(e) : n.replace(e, t)), this;
	}
	splice(e, t, n = "next") {
		let r = n === this.direction ? t : Array.from(t).reverse(), i = this, a = this.getParentAtCaret(), o = /* @__PURE__ */ new Map();
		for (let t = i.getAdjacentCaret(); t !== null && o.size < e; t = t.getAdjacentCaret()) {
			let e = t.origin.getWritable();
			o.set(e.getKey(), e);
		}
		for (let e of r) {
			if (o.size > 0) {
				let t = i.getNodeAtCaret();
				if (t) {
					if (o.delete(t.getKey()), o.delete(e.getKey()), !(t.is(e) || i.origin.is(e))) {
						let n = e.getParent();
						n && n.is(a) && e.remove(), t.replace(e);
					}
				} else t === null && V(263, Array.from(o).join(" "));
			} else i.insert(e);
			i = hd(e, this.direction);
		}
		for (let e of o.values()) e.remove();
		return this;
	}
}, rd = class e extends nd {
	type = "child";
	getLatest() {
		let e = this.origin.getLatest();
		return e === this.origin ? this : yd(e, this.direction);
	}
	getParentCaret(e = "root") {
		return hd(od(this.getParentAtCaret(), e), this.direction);
	}
	getFlipped() {
		let e = ad(this.direction);
		return hd(this.getNodeAtCaret(), e) || yd(this.origin, e);
	}
	getParentAtCaret() {
		return this.origin;
	}
	getChildCaret() {
		return this;
	}
	isSameNodeCaret(t) {
		return t instanceof e && this.direction === t.direction && this.origin.is(t.origin);
	}
	isSamePointCaret(e) {
		return this.isSameNodeCaret(e);
	}
}, id = {
	root: ll,
	shadowRoot: ju
};
function ad(e) {
	return td[e];
}
function od(e, t = "root") {
	return id[t](e) ? null : e;
}
var sd = class e extends nd {
	type = "sibling";
	getLatest() {
		let e = this.origin.getLatest();
		return e === this.origin ? this : hd(e, this.direction);
	}
	getSiblingCaret() {
		return this;
	}
	getParentAtCaret() {
		return this.origin.getParent();
	}
	getChildCaret() {
		return Y(this.origin) ? yd(this.origin, this.direction) : null;
	}
	getParentCaret(e = "root") {
		return hd(od(this.getParentAtCaret(), e), this.direction);
	}
	getFlipped() {
		let e = ad(this.direction);
		return hd(this.getNodeAtCaret(), e) || yd(this.origin.getParentOrThrow(), e);
	}
	isSamePointCaret(t) {
		return t instanceof e && this.direction === t.direction && this.origin.is(t.origin);
	}
	isSameNodeCaret(t) {
		return (t instanceof e || t instanceof cd) && this.direction === t.direction && this.origin.is(t.origin);
	}
}, cd = class e extends nd {
	type = "text";
	constructor(e, t) {
		super(e), this.offset = t;
	}
	getLatest() {
		let e = this.origin.getLatest();
		return e === this.origin ? this : gd(e, this.direction, this.offset);
	}
	getParentAtCaret() {
		return this.origin.getParent();
	}
	getChildCaret() {
		return null;
	}
	getParentCaret(e = "root") {
		return hd(od(this.getParentAtCaret(), e), this.direction);
	}
	getFlipped() {
		return gd(this.origin, ad(this.direction), this.offset);
	}
	isSamePointCaret(t) {
		return t instanceof e && this.direction === t.direction && this.origin.is(t.origin) && this.offset === t.offset;
	}
	isSameNodeCaret(t) {
		return (t instanceof sd || t instanceof e) && this.direction === t.direction && this.origin.is(t.origin);
	}
	getSiblingCaret() {
		return hd(this.origin, this.direction);
	}
};
function ld(e) {
	return e instanceof cd;
}
function ud(e) {
	return e instanceof sd;
}
function dd(e) {
	return e instanceof rd;
}
var fd = {
	next: class extends cd {
		direction = "next";
		getNodeAtCaret() {
			return this.origin.getNextSibling();
		}
		insert(e) {
			return this.origin.insertAfter(e), this;
		}
	},
	previous: class extends cd {
		direction = "previous";
		getNodeAtCaret() {
			return this.origin.getPreviousSibling();
		}
		insert(e) {
			return this.origin.insertBefore(e), this;
		}
	}
}, pd = {
	next: class extends sd {
		direction = "next";
		getNodeAtCaret() {
			return this.origin.getNextSibling();
		}
		insert(e) {
			return this.origin.insertAfter(e), this;
		}
	},
	previous: class extends sd {
		direction = "previous";
		getNodeAtCaret() {
			return this.origin.getPreviousSibling();
		}
		insert(e) {
			return this.origin.insertBefore(e), this;
		}
	}
}, md = {
	next: class extends rd {
		direction = "next";
		getNodeAtCaret() {
			return this.origin.getFirstChild();
		}
		insert(e) {
			return this.origin.splice(0, 0, [e]), this;
		}
	},
	previous: class extends rd {
		direction = "previous";
		getNodeAtCaret() {
			return this.origin.getLastChild();
		}
		insert(e) {
			return this.origin.splice(this.origin.getChildrenSize(), 0, [e]), this;
		}
	}
};
function hd(e, t) {
	return e ? new pd[t](e) : null;
}
function gd(e, t, n) {
	return e ? new fd[t](e, _d(e, n)) : null;
}
function _d(e, t, n = "error") {
	let r = e.getTextContentSize(), i = t === "next" ? r : t === "previous" ? 0 : t;
	return (i < 0 || i > r) && (n !== "clamp" && function(e, ...t) {
		let n = new URL("https://lexical.dev/docs/error"), r = new URLSearchParams();
		r.append("code", e);
		for (let e of t) r.append("v", e);
		n.search = r.toString(), console.warn(`Minified Lexical warning #${e}; visit ${n.toString()} for the full message or use the non-minified dev environment for full errors and additional helpful warnings.`);
	}(284, String(t), String(r), e.getKey()), i = i < 0 ? 0 : r), i;
}
function vd(e, t) {
	return new Cd(e, t);
}
function yd(e, t) {
	return Y(e) ? new md[t](e) : null;
}
function bd(e) {
	return e && e.getChildCaret() || e;
}
function xd(e) {
	return e && bd(e.getAdjacentCaret());
}
var Sd = class e {
	type = "node-caret-range";
	constructor(e, t, n) {
		this.anchor = e, this.focus = t, this.direction = n;
	}
	getLatest() {
		let t = this.anchor.getLatest(), n = this.focus.getLatest();
		return t === this.anchor && n === this.focus ? this : new e(t, n, this.direction);
	}
	isCollapsed() {
		return this.anchor.isSamePointCaret(this.focus);
	}
	getTextSlices() {
		let e = (e) => {
			let t = this[e].getLatest();
			return ld(t) ? function(e, t) {
				let { direction: n, origin: r } = e;
				return vd(e, _d(r, t === "focus" ? ad(n) : n) - e.offset);
			}(t, e) : null;
		}, t = e("anchor"), n = e("focus");
		if (t && n) {
			let { caret: e } = t, { caret: r } = n;
			if (e.isSameNodeCaret(r)) return [vd(e, r.offset - e.offset), null];
		}
		return [t, n];
	}
	iterNodeCarets(e = "root") {
		let t = ld(this.anchor) ? this.anchor.getSiblingCaret() : this.anchor.getLatest(), n = this.focus.getLatest(), r = ld(n), i = (t) => t.isSameNodeCaret(n) ? null : xd(t) || t.getParentCaret(e);
		return Dd({
			hasNext: (e) => e !== null && !(r && n.isSameNodeCaret(e)),
			initial: t.isSameNodeCaret(n) ? null : i(t),
			map: (e) => e,
			step: i
		});
	}
	[Symbol.iterator]() {
		return this.iterNodeCarets("root");
	}
}, Cd = class {
	type = "slice";
	constructor(e, t) {
		this.caret = e, this.distance = t;
	}
	getSliceIndices() {
		let { distance: e, caret: { offset: t } } = this, n = t + e;
		return n < t ? [n, t] : [t, n];
	}
	getTextContent() {
		let [e, t] = this.getSliceIndices();
		return this.caret.origin.getTextContent().slice(e, t);
	}
	getTextContentSize() {
		return Math.abs(this.distance);
	}
	removeTextSlice() {
		let { caret: { origin: e, direction: t } } = this, [n, r] = this.getSliceIndices(), i = e.getTextContent();
		return gd(e.setTextContent(i.slice(0, n) + i.slice(r)), t, n);
	}
};
function wd(e) {
	return Ed(e, hd(Z(), e.direction));
}
function Td(e) {
	return Ed(e, e);
}
function Ed(e, t) {
	return e.direction !== t.direction && V(265), new Sd(e, t, e.direction);
}
function Dd(e) {
	let { initial: t, hasNext: n, step: r, map: i } = e, a = t;
	return {
		[Symbol.iterator]() {
			return this;
		},
		next() {
			if (!n(a)) return {
				done: !0,
				value: void 0
			};
			let e = {
				done: !1,
				value: i(a)
			};
			return a = r(a), e;
		}
	};
}
function Od(e, t) {
	let n = Md(e.origin, t.origin);
	switch (n === null && V(275, e.origin.getKey(), t.origin.getKey()), n.type) {
		case "same": {
			let n = e.type === "text", r = t.type === "text";
			return n && r ? function(e, t) {
				return Math.sign(e - t);
			}(e.offset, t.offset) : e.type === t.type ? 0 : n ? -1 : r ? 1 : e.type === "child" ? -1 : 1;
		}
		case "ancestor": return e.type === "child" ? -1 : 1;
		case "descendant": return t.type === "child" ? 1 : -1;
		case "branch": return kd(n);
	}
}
function kd(e) {
	let { a: t, b: n } = e, r = t.__key, i = n.__key, a = t, o = n;
	for (; a && o; a = a.getNextSibling(), o = o.getNextSibling()) {
		if (a.__key === i) return -1;
		if (o.__key === r) return 1;
	}
	return a === null ? 1 : -1;
}
function Ad(e, t) {
	return t.is(e);
}
function jd(e) {
	return Y(e) ? [e.getLatest(), null] : [e.getParent(), e.getLatest()];
}
function Md(e, t) {
	if (e.is(t)) return {
		commonAncestor: e,
		type: "same"
	};
	let n = /* @__PURE__ */ new Map();
	for (let [t, r] = jd(e); t; r = t, t = t.getParent()) n.set(t, r);
	for (let [r, i] = jd(t); r; i = r, r = r.getParent()) {
		let a = n.get(r);
		if (a !== void 0) return a === null ? (Ad(e, r) || V(276), {
			commonAncestor: r,
			type: "ancestor"
		}) : i === null ? (Ad(t, r) || V(277), {
			commonAncestor: r,
			type: "descendant"
		}) : ((Y(a) || Ad(e, a)) && (Y(i) || Ad(t, i)) && r.is(a.getParent()) && r.is(i.getParent()) || V(278), {
			a,
			b: i,
			commonAncestor: r,
			type: "branch"
		});
	}
	return null;
}
function Nd(e, t) {
	let { type: n, key: r, offset: i } = e, a = Pu(e.key);
	return n === "text" ? (G(a) || V(266, a.getType(), r), gd(a, t, i)) : (Y(a) || V(267, a.getType(), r), Kd(a, e.offset, t));
}
function Pd(e, t) {
	let { origin: n, direction: r } = t, i = r === "next";
	ld(t) ? e.set(n.getKey(), t.offset, "text") : ud(t) ? G(n) ? e.set(n.getKey(), _d(n, r), "text") : e.set(n.getParentOrThrow().getKey(), n.getIndexWithinParent() + +!!i, "element") : (dd(t) && Y(n) || V(268), e.set(n.getKey(), i ? 0 : n.getChildrenSize(), "element"));
}
function Fd(e) {
	let t = q(), n = K(t) ? t : Sc();
	return Id(n, e), eu(n), n;
}
function Id(e, t) {
	Pd(e.anchor, t.anchor), Pd(e.focus, t.focus);
}
function Ld(e) {
	let { anchor: t, focus: n } = e, r = Nd(t, "next"), i = Nd(n, "next"), a = Od(r, i) <= 0 ? "next" : "previous";
	return Ed(Wd(r, a), Wd(i, a));
}
function Rd(e) {
	let { direction: t, origin: n } = e, r = hd(n, ad(t)).getNodeAtCaret();
	return r ? hd(r, t) : yd(n.getParentOrThrow(), t);
}
function zd(e, t = "root") {
	let n = [e];
	for (let r = dd(e) ? e.getParentCaret(t) : e.getSiblingCaret(); r !== null; r = r.getParentCaret(t)) n.push(Rd(r));
	return n;
}
function Bd(e) {
	return !!e && e.origin.isAttached();
}
function Vd(e, t = "removeEmptySlices") {
	if (e.isCollapsed()) return e;
	let n = "root", r = "next", i = t, a = Gd(e, r), o = zd(a.anchor, n), s = zd(a.focus.getFlipped(), n), c = /* @__PURE__ */ new Set(), l = [];
	for (let e of a.iterNodeCarets(n)) if (dd(e)) c.add(e.origin.getKey());
	else if (ud(e)) {
		let { origin: t } = e;
		Y(t) && !c.has(t.getKey()) || l.push(t);
	}
	for (let e of l) e.remove();
	for (let e of a.getTextSlices()) {
		if (!e) continue;
		let { origin: t } = e.caret, n = t.getTextContentSize(), a = Rd(hd(t, r)), c = t.getMode();
		if (Math.abs(e.distance) === n && i === "removeEmptySlices" || c === "token" && e.distance !== 0) a.remove();
		else if (e.distance !== 0) {
			i = "removeEmptySlices";
			let t = e.removeTextSlice(), n = e.caret.origin;
			if (c === "segmented") {
				let e = t.origin, n = W(e.getTextContent()).setStyle(e.getStyle()).setFormat(e.getFormat());
				a.replaceOrInsert(n), t = gd(n, r, t.offset);
			}
			n.is(o[0].origin) && (o[0] = t), n.is(s[0].origin) && (s[0] = t.getFlipped());
		}
	}
	let u, d;
	for (let e of o) if (Bd(e)) {
		u = Hd(e);
		break;
	}
	for (let e of s) if (Bd(e)) {
		d = Hd(e);
		break;
	}
	let f = function(e, t, n) {
		if (!e || !t) return null;
		let r = e.getParentAtCaret(), i = t.getParentAtCaret();
		if (!r || !i) return null;
		let a = r.getParents().reverse();
		a.push(r);
		let o = i.getParents().reverse();
		o.push(i);
		let s = Math.min(a.length, o.length), c = 0;
		for (; c < s && a[c] === o[c]; c++);
		let l = (e, t) => {
			let n;
			for (let r = c; r < e.length; r++) {
				let i = e[r];
				if (ju(i)) return;
				!n && t(i) && (n = i);
			}
			return n;
		}, u = l(a, Wu), d = u && l(o, ((e) => n.has(e.getKey()) && Wu(e)));
		return u && d ? [u, d] : null;
	}(u, d, c);
	if (f) {
		let [e, t] = f;
		yd(e, "previous").splice(0, t.getChildren()), t.remove();
	}
	let p = [
		u,
		d,
		...o,
		...s
	].find(Bd);
	if (p) return Td(Wd(Hd(p), e.direction));
	V(269, JSON.stringify(o.map(((e) => e.origin.__key))));
}
function Hd(e) {
	let t = function(e) {
		let t = e;
		for (; dd(t);) {
			let e = xd(t);
			if (!dd(e)) break;
			t = e;
		}
		return t;
	}(e.getLatest()), { direction: n } = t;
	if (G(t.origin)) return ld(t) ? t : gd(t.origin, n, n);
	let r = t.getAdjacentCaret();
	return ud(r) && G(r.origin) ? gd(r.origin, n, ad(n)) : t;
}
function Ud(e) {
	return ld(e) && e.offset !== _d(e.origin, e.direction);
}
function Wd(e, t) {
	return e.direction === t ? e : e.getFlipped();
}
function Gd(e, t) {
	return e.direction === t ? e : Ed(Wd(e.focus, t), Wd(e.anchor, t));
}
function Kd(e, t, n) {
	let r = yd(e, "next");
	for (let e = 0; e < t; e++) {
		let e = r.getAdjacentCaret();
		if (e === null) break;
		r = e;
	}
	return Wd(r, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalComposer.prod.mjs
var qd = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, Jd = qd ? N.useLayoutEffect : N.useEffect, Yd = { tag: hl };
function Xd({ initialConfig: e, children: t }) {
	let n = (0, N.useMemo)((() => {
		let { theme: t, namespace: n, nodes: r, onError: i, editorState: a, html: o } = e, s = Di(null, t), c = Sl({
			editable: e.editable,
			html: o,
			namespace: n,
			nodes: r,
			onError: (e) => i(e, c),
			theme: t
		});
		return function(e, t) {
			if (t !== null) {
				if (t === void 0) e.update((() => {
					let t = Z();
					if (t.isEmpty()) {
						let n = yl();
						t.append(n);
						let r = qd ? document.activeElement : null;
						(q() !== null || r !== null && r === e.getRootElement()) && n.select();
					}
				}), Yd);
				else if (t !== null) switch (typeof t) {
					case "string": {
						let n = e.parseEditorState(t);
						e.setEditorState(n, Yd);
						break;
					}
					case "object":
						e.setEditorState(t, Yd);
						break;
					case "function": e.update((() => {
						Z().isEmpty() && t(e);
					}), Yd);
				}
			}
		}(c, a), [c, s];
	}), []);
	return Jd((() => {
		let t = e.editable, [r] = n;
		r.setEditable(t === void 0 || t);
	}), []), (0, P.jsx)(Ei.Provider, {
		value: n,
		children: t
	});
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/useLexicalEditable.prod.mjs
var Zd = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0 ? N.useLayoutEffect : N.useEffect;
function Qd(e) {
	return {
		initialValueFn: () => e.isEditable(),
		subscribe: (t) => e.registerEditableListener(t)
	};
}
function $d() {
	return function(e) {
		let [t] = Oi(), n = (0, N.useMemo)((() => e(t)), [t, e]), [r, i] = (0, N.useState)((() => n.initialValueFn())), a = (0, N.useRef)(r);
		return Zd((() => {
			let { initialValueFn: e, subscribe: t } = n, r = e();
			return a.current !== r && (a.current = r, i(r)), t(((e) => {
				a.current = e, i(e);
			}));
		}), [n, e]), r;
	}(Qd);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/text/0.35.0/4794399b98f2048b2d2d4998efc69b3d4fe404b5a2b66271a324f512817cd3c0/node_modules/@lexical/text/LexicalText.prod.mjs
function ef() {
	return Z().getTextContent();
}
function tf(e, t = !0) {
	if (e) return !1;
	let n = ef();
	return t && (n = n.trim()), n === "";
}
function nf(e) {
	if (!tf(e, !1)) return !1;
	let t = Z().getChildren(), n = t.length;
	if (n > 1) return !1;
	for (let e = 0; e < n; e++) {
		let n = t[e];
		if (X(n)) return !1;
		if (Y(n)) {
			if (!bl(n) || n.__indent !== 0) return !1;
			let t = n.getChildren(), r = t.length;
			for (let n = 0; n < r; n++) {
				let n = t[e];
				if (!G(n)) return !1;
			}
		}
	}
	return !0;
}
function rf(e) {
	return () => nf(e);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/selection/0.35.0/ff558ed86062196423283cb31e0f3c76dad90aa8a73c8162a63a0351f3911343/node_modules/@lexical/selection/LexicalSelection.prod.mjs
function af(e) {
	let t = Ku().getElementByKey(e.getKey());
	if (t === null) return null;
	let n = t.ownerDocument.defaultView;
	return n === null ? null : n.getComputedStyle(t);
}
function of(e) {
	return af(ll(e) ? e : e.getParentOrThrow());
}
function sf(e, t) {
	let n = e.getStartEndPoints();
	if (t.isSelected(e) && !Ll(t) && n !== null) {
		let [r, i] = n, a = e.isBackward(), o = r.getNode(), s = i.getNode(), c = t.is(o), l = t.is(s);
		if (c || l) {
			let [n, r] = pc(e), i = o.is(s), c = t.is(a ? s : o), l = t.is(a ? o : s), u, d = 0;
			i ? (d = n > r ? r : n, u = n > r ? n : r) : c ? (d = a ? r : n, u = void 0) : l && (d = 0, u = a ? n : r), t.__text = t.__text.slice(d, u);
		}
	}
	return t;
}
function cf(e) {
	let t = lf(e);
	return t !== null && t.writingMode === "vertical-rl";
}
function lf(e) {
	let t = e.anchor.getNode();
	return Y(t) ? af(t) : of(t);
}
function uf(e, t) {
	let n = cf(e) ? !t : t;
	ff(e) && (n = !n);
	let r = Nd(e.focus, n ? "previous" : "next");
	if (Ud(r)) return !1;
	for (let e of wd(r)) {
		if (dd(e)) return !e.origin.isInline();
		if (!Y(e.origin)) {
			if (X(e.origin)) return !0;
			break;
		}
	}
	return !1;
}
function df(e, t, n, r) {
	e.modify(t ? "extend" : "move", n, r);
}
function ff(e) {
	let t = lf(e);
	return t !== null && t.direction === "rtl";
}
function pf(e, t, n) {
	let r = ff(e), i;
	i = cf(e) || r ? !n : n, df(e, t, i, "character");
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/utils/0.35.0/cc980af960a164acefb1918eec77142cf245d659ddb4911210cebff0515a63bd/node_modules/@lexical/utils/LexicalUtils.prod.mjs
var mf = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, hf = mf && "documentMode" in document ? document.documentMode : null;
mf && /Mac|iPod|iPhone|iPad/.test(navigator.platform), mf && /^(?!.*Seamonkey)(?=.*Firefox).*/i.test(navigator.userAgent), mf && "InputEvent" in window && !hf && "getTargetRanges" in new window.InputEvent("input"), mf && /Version\/[\d.]+.*Safari/.test(navigator.userAgent), mf && /iPad|iPhone|iPod/.test(navigator.userAgent) && window.MSStream, mf && /Android/.test(navigator.userAgent), mf && /^(?=.*Chrome).*/i.test(navigator.userAgent), mf && /AppleWebKit\/[\d.]+/.test(navigator.userAgent);
function gf(...e) {
	return () => {
		for (let t = e.length - 1; t >= 0; t--) e[t]();
		e.length = 0;
	};
}
function _f(e, t) {
	return e !== null && Object.getPrototypeOf(e).constructor.name === t.name;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/dragon/0.35.0/0e7233b4751a9f1d499e386ef0d1c40bc41f5ec1b215a445a25648b6d26ca49c/node_modules/@lexical/dragon/LexicalDragon.prod.mjs
function vf(e) {
	let t = window.location.origin, n = (n) => {
		if (n.origin !== t) return;
		let r = e.getRootElement();
		if (document.activeElement !== r) return;
		let i = n.data;
		if (typeof i == "string") {
			let t;
			try {
				t = JSON.parse(i);
			} catch {
				return;
			}
			if (t && t.protocol === "nuanria_messaging" && t.type === "request") {
				let r = t.payload;
				if (r && r.functionId === "makeChanges") {
					let t = r.args;
					if (t) {
						let [r, i, a, o, s, c] = t;
						e.update((() => {
							let e = q();
							if (K(e)) {
								let t = e.anchor, c = t.getNode(), l = 0, u = 0;
								if (G(c) && r >= 0 && i >= 0 && (l = r, u = r + i, e.setTextNodeRange(c, l, c, u)), l === u && a === "" || (e.insertRawText(a), c = t.getNode()), G(c)) {
									l = o, u = o + s;
									let t = c.getTextContentSize();
									l = l > t ? t : l, u = u > t ? t : u, e.setTextNodeRange(c, l, c, u);
								}
								n.stopImmediatePropagation();
							}
						}));
					}
				}
			}
		}
	};
	return window.addEventListener("message", n, !0), () => {
		window.removeEventListener("message", n, !0);
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/html/0.35.0/6efbce2fa15e3081cbdc22752efaecdb49c738d801b476c3f91ce7d4065a3ceb/node_modules/@lexical/html/LexicalHtml.prod.mjs
function yf(e, t) {
	if (typeof document > "u" || typeof window > "u" && global.window === void 0) throw Error("To use $generateHtmlFromNodes in headless mode please initialize a headless browser implementation such as JSDom before calling this function.");
	let n = document.createElement("div"), r = Z().getChildren();
	for (let i = 0; i < r.length; i++) bf(e, r[i], n, t);
	return n.innerHTML;
}
function bf(e, t, n, r = null) {
	let i = r === null || t.isSelected(r), a = Y(t) && t.excludeFromCopy("html"), o = t;
	if (r !== null) {
		let e = Xu(t);
		e = G(e) && r !== null ? sf(r, e) : e, o = e;
	}
	let s = Y(o) ? o.getChildren() : [], c = Ol(e, o.getType()), l;
	l = c && c.exportDOM !== void 0 ? c.exportDOM(e, o) : o.exportDOM(e);
	let { element: u, after: d } = l;
	if (!u) return !1;
	let f = document.createDocumentFragment();
	for (let n = 0; n < s.length; n++) {
		let a = s[n], o = bf(e, a, f, r);
		!i && Y(t) && o && t.extractWithChild(a, r, "html") && (i = !0);
	}
	if (i && !a) {
		if ((zu(u) || Vu(u)) && u.append(f), n.append(u), d) {
			let e = d.call(o, u);
			e && (Vu(u) ? u.replaceChildren(e) : u.replaceWith(e));
		}
	} else n.append(f);
	return i;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/clipboard/0.35.0/a3747864708938cda9f551d20afe1b552cbdb03b1e9604476908248cb499e2ed/node_modules/@lexical/clipboard/LexicalClipboard.prod.mjs
function xf(e, ...t) {
	let n = new URL("https://lexical.dev/docs/error"), r = new URLSearchParams();
	r.append("code", e);
	for (let e of t) r.append("v", e);
	throw n.search = r.toString(), Error(`Minified Lexical error #${e}; visit ${n.toString()} for the full message or use the non-minified dev environment for full errors and additional helpful warnings.`);
}
function Sf(e, t = q()) {
	return t ?? xf(166), K(t) && t.isCollapsed() || t.getNodes().length === 0 ? "" : yf(e, t);
}
function Cf(e, t) {
	let n = e.getData("text/plain") || e.getData("text/uri-list");
	n != null && t.insertRawText(n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/plain-text/0.35.0/e9e3f6e43289721091b22aacc6eeef1fdfb1924dd14fd5dbaffb000ed4b92147/node_modules/@lexical/plain-text/LexicalPlainText.prod.mjs
var wf = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0, Tf = wf && "documentMode" in document ? document.documentMode : null, Ef = !(!wf || !("InputEvent" in window) || Tf) && "getTargetRanges" in new window.InputEvent("input"), Df = wf && /Version\/[\d.]+.*Safari/.test(navigator.userAgent), Of = wf && /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream, kf = wf && /^(?=.*Chrome).*/i.test(navigator.userAgent), Af = wf && /AppleWebKit\/[\d.]+/.test(navigator.userAgent) && !kf;
function jf(e, t) {
	t.update((() => {
		if (e !== null) {
			let n = _f(e, KeyboardEvent) ? null : e.clipboardData, r = q();
			if (r !== null && n != null) {
				e.preventDefault();
				let i = Sf(t);
				i !== null && n.setData("text/html", i), n.setData("text/plain", r.getTextContent());
			}
		}
	}));
}
function Mf(e) {
	return gf(e.registerCommand(bo, ((e) => {
		let t = q();
		return !!K(t) && (t.deleteCharacter(e), !0);
	}), 0), e.registerCommand(Eo, ((e) => {
		let t = q();
		return !!K(t) && (t.deleteWord(e), !0);
	}), 0), e.registerCommand(Do, ((e) => {
		let t = q();
		return !!K(t) && (t.deleteLine(e), !0);
	}), 0), e.registerCommand(Co, ((e) => {
		let t = q();
		if (!K(t)) return !1;
		if (typeof e == "string") t.insertText(e);
		else {
			let n = e.dataTransfer;
			if (n != null) Cf(n, t);
			else {
				let n = e.data;
				n && t.insertText(n);
			}
		}
		return !0;
	}), 0), e.registerCommand(To, (() => {
		let e = q();
		return !!K(e) && (e.removeText(), !0);
	}), 0), e.registerCommand(xo, ((e) => {
		let t = q();
		return !!K(t) && (t.insertLineBreak(e), !0);
	}), 0), e.registerCommand(So, (() => {
		let e = q();
		return !!K(e) && (e.insertLineBreak(), !0);
	}), 0), e.registerCommand(Po, ((e) => {
		let t = q();
		if (!K(t)) return !1;
		let n = e, r = n.shiftKey;
		return !!uf(t, !0) && (n.preventDefault(), pf(t, r, !0), !0);
	}), 0), e.registerCommand(Mo, ((e) => {
		let t = q();
		if (!K(t)) return !1;
		let n = e, r = n.shiftKey;
		return !!uf(t, !1) && (n.preventDefault(), pf(t, r, !1), !0);
	}), 0), e.registerCommand(Bo, ((t) => !!K(q()) && (!Of || navigator.language !== "ko-KR") && (t.preventDefault(), e.dispatchCommand(bo, !0))), 0), e.registerCommand(Ho, ((t) => !!K(q()) && (t.preventDefault(), e.dispatchCommand(bo, !1))), 0), e.registerCommand(Ro, ((t) => {
		if (!K(q())) return !1;
		if (t !== null) {
			if ((Of || Df || Af) && Ef) return !1;
			t.preventDefault();
		}
		return e.dispatchCommand(xo, !1);
	}), 0), e.registerCommand(Xo, (() => (hu(), !0)), 0), e.registerCommand(Jo, ((t) => !!K(q()) && (jf(t, e), !0)), 0), e.registerCommand(Yo, ((t) => !!K(q()) && (function(e, t) {
		jf(e, t), t.update((() => {
			let e = q();
			K(e) && e.removeText();
		}));
	}(t, e), !0)), 0), e.registerCommand(wo, ((t) => !!K(q()) && (function(e, t) {
		e.preventDefault(), t.update((() => {
			let t = q(), n = _f(e, ClipboardEvent) ? e.clipboardData : null;
			n != null && K(t) && Cf(n, t);
		}), { tag: "paste" });
	}(t, e), !0)), 0), e.registerCommand(Wo, ((e) => !!K(q()) && (e.preventDefault(), !0)), 0), e.registerCommand(Go, ((e) => !!K(q()) && (e.preventDefault(), !0)), 0));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalPlainTextPlugin.prod.mjs
var Nf = /* @__PURE__ */ e(i(), 1), Pf = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0 ? N.useLayoutEffect : N.useEffect;
function Ff(e) {
	return e.getEditorState().read(rf(e.isComposing()));
}
function If({ contentEditable: e, placeholder: t = null, ErrorBoundary: n }) {
	let [r] = Oi(), i = function(e, t) {
		let [n, r] = (0, N.useState)((() => e.getDecorators()));
		return Pf((() => e.registerDecoratorListener(((e) => {
			(0, Nf.flushSync)((() => {
				r(e);
			}));
		}))), [e]), (0, N.useEffect)((() => {
			r(e.getDecorators());
		}), [e]), (0, N.useMemo)((() => {
			let r = [], i = Object.keys(n);
			for (let a = 0; a < i.length; a++) {
				let o = i[a], s = (0, P.jsx)(t, {
					onError: (t) => e._onError(t),
					children: (0, P.jsx)(N.Suspense, {
						fallback: null,
						children: n[o]
					})
				}), c = e.getElementByKey(o);
				c !== null && r.push((0, Nf.createPortal)(s, c, o));
			}
			return r;
		}), [
			t,
			n,
			e
		]);
	}(r, n);
	return function(e) {
		Pf((() => gf(Mf(e), vf(e))), [e]);
	}(r), (0, P.jsxs)(P.Fragment, { children: [
		e,
		(0, P.jsx)(Lf, { content: t }),
		i
	] });
}
function Lf({ content: e }) {
	let [t] = Oi(), n = function(e) {
		let [t, n] = (0, N.useState)((() => Ff(e)));
		return Pf((() => {
			function t() {
				let t = Ff(e);
				n(t);
			}
			return t(), gf(e.registerUpdateListener((() => {
				t();
			})), e.registerEditableListener((() => {
				t();
			})));
		}), [e]), t;
	}(t), r = $d();
	return n ? typeof e == "function" ? e(r) : e : null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalContentEditable.prod.mjs
var Rf = typeof window < "u" && window.document !== void 0 && window.document.createElement !== void 0 ? N.useLayoutEffect : N.useEffect;
function zf({ editor: e, ariaActiveDescendant: t, ariaAutoComplete: n, ariaControls: r, ariaDescribedBy: i, ariaErrorMessage: a, ariaExpanded: o, ariaInvalid: s, ariaLabel: c, ariaLabelledBy: l, ariaMultiline: u, ariaOwns: d, ariaRequired: f, autoCapitalize: p, className: m, id: h, role: g = "textbox", spellCheck: _ = !0, style: v, tabIndex: y, "data-testid": b, ...x }, S) {
	let [C, w] = (0, N.useState)(e.isEditable()), T = (0, N.useCallback)(((t) => {
		t && t.ownerDocument && t.ownerDocument.defaultView ? e.setRootElement(t) : e.setRootElement(null);
	}), [e]), E = (0, N.useMemo)((() => function(...e) {
		return (t) => {
			e.forEach(((e) => {
				typeof e == "function" ? e(t) : e != null && (e.current = t);
			}));
		};
	}(S, T)), [T, S]);
	return Rf((() => (w(e.isEditable()), e.registerEditableListener(((e) => {
		w(e);
	})))), [e]), (0, P.jsx)("div", {
		"aria-activedescendant": C ? t : void 0,
		"aria-autocomplete": C ? n : "none",
		"aria-controls": C ? r : void 0,
		"aria-describedby": i,
		...a == null ? {} : { "aria-errormessage": a },
		"aria-expanded": C && g === "combobox" ? !!o : void 0,
		...s == null ? {} : { "aria-invalid": s },
		"aria-label": c,
		"aria-labelledby": l,
		"aria-multiline": u,
		"aria-owns": C ? d : void 0,
		"aria-readonly": !C || void 0,
		"aria-required": f,
		autoCapitalize: p,
		className: m,
		contentEditable: C,
		"data-testid": b,
		id: h,
		ref: E,
		role: g,
		spellCheck: _,
		style: v,
		tabIndex: y,
		...x
	});
}
var Bf = (0, N.forwardRef)(zf);
function Vf(e) {
	return e.getEditorState().read(rf(e.isComposing()));
}
var Hf = (0, N.forwardRef)(Uf);
function Uf(e, t) {
	let { placeholder: n, ...r } = e, [i] = Oi();
	return (0, P.jsxs)(P.Fragment, { children: [(0, P.jsx)(Bf, {
		editor: i,
		...r,
		ref: t
	}), n != null && (0, P.jsx)(Wf, {
		editor: i,
		content: n
	})] });
}
function Wf({ content: e, editor: t }) {
	let n = function(e) {
		let [t, n] = (0, N.useState)((() => Vf(e)));
		return Rf((() => {
			function t() {
				let t = Vf(e);
				n(t);
			}
			return t(), gf(e.registerUpdateListener((() => {
				t();
			})), e.registerEditableListener((() => {
				t();
			})));
		}), [e]), t;
	}(t), [r, i] = (0, N.useState)(t.isEditable());
	if ((0, N.useLayoutEffect)((() => (i(t.isEditable()), t.registerEditableListener(((e) => {
		i(e);
	})))), [t]), !n) return null;
	let a = null;
	return typeof e == "function" ? a = e(r) : e !== null && (a = e), a === null ? null : (0, P.jsx)("div", {
		"aria-hidden": !0,
		children: a
	});
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/history/0.35.0/6fa90706fdfceced48747f7b9b2cba9b5d5417fc9db4fa24a79dd7b0c891c3c2/node_modules/@lexical/history/LexicalHistory.prod.mjs
var Gf = 0, Kf = 1, qf = 2, Jf = 0, Yf = 1, Xf = 2, Zf = 3, Qf = 4;
function $f(e, t, n, r, i) {
	if (e === null || n.size === 0 && r.size === 0 && !i) return Jf;
	let a = t._selection, o = e._selection;
	if (i) return Yf;
	if (!(K(a) && K(o) && o.isCollapsed() && a.isCollapsed())) return Jf;
	let s = function(e, t, n) {
		let r = e._nodeMap, i = [];
		for (let e of t) {
			let t = r.get(e);
			t !== void 0 && i.push(t);
		}
		for (let [e, t] of n) {
			if (!t) continue;
			let n = r.get(e);
			n === void 0 || ll(n) || i.push(n);
		}
		return i;
	}(t, n, r);
	if (s.length === 0) return Jf;
	if (s.length > 1) {
		let n = t._nodeMap, r = n.get(a.anchor.key), i = n.get(o.anchor.key);
		return r && i && !e._nodeMap.has(r.__key) && G(r) && r.__text.length === 1 && a.anchor.offset === 1 ? Xf : Jf;
	}
	let c = s[0], l = e._nodeMap.get(c.__key);
	if (!G(l) || !G(c) || l.__mode !== c.__mode) return Jf;
	let u = l.__text, d = c.__text;
	if (u === d) return Jf;
	let f = a.anchor, p = o.anchor;
	if (f.key !== p.key || f.type !== "text") return Jf;
	let m = f.offset, h = p.offset, g = d.length - u.length;
	return g === 1 && h === m - 1 ? Xf : g === -1 && h === m + 1 ? Zf : g === -1 && h === m ? Qf : Jf;
}
function ep(e, t) {
	let n = Date.now(), r = Jf;
	return (i, a, o, s, c, l) => {
		let u = Date.now();
		if (l.has("historic")) return r = Jf, n = u, qf;
		let d = $f(i, a, s, c, e.isComposing()), f = (() => {
			let f = o === null || o.editor === e, p = l.has(ml);
			if (!p && f && l.has("history-merge")) return Gf;
			if (i === null) return Kf;
			let m = a._selection;
			return s.size > 0 || c.size > 0 ? !1 === p && d !== Jf && d === r && u < n + t && f || s.size === 1 && function(e, t, n) {
				let r = t._nodeMap.get(e), i = n._nodeMap.get(e), a = t._selection, o = n._selection;
				return !(K(a) && K(o) && a.anchor.type === "element" && a.focus.type === "element" && o.anchor.type === "text" && o.focus.type === "text" || !G(r) || !G(i) || r.__parent !== i.__parent) && JSON.stringify(t.read((() => r.exportJSON()))) === JSON.stringify(n.read((() => i.exportJSON())));
			}(Array.from(s)[0], i, a) ? Gf : Kf : m === null ? qf : Gf;
		})();
		return n = u, r = d, f;
	};
}
function tp(e) {
	e.undoStack = [], e.redoStack = [], e.current = null;
}
function np(e, t, n) {
	let r = ep(e, n);
	return gf(e.registerCommand(ko, (() => (function(e, t) {
		let n = t.redoStack, r = t.undoStack;
		if (r.length !== 0) {
			let i = t.current, a = r.pop();
			i !== null && (n.push(i), e.dispatchCommand($o, !0)), r.length === 0 && e.dispatchCommand(es, !1), t.current = a || null, a && a.editor.setEditorState(a.editorState, { tag: "historic" });
		}
	}(e, t), !0)), 0), e.registerCommand(Ao, (() => (function(e, t) {
		let n = t.redoStack, r = t.undoStack;
		if (n.length !== 0) {
			let i = t.current;
			i !== null && (r.push(i), e.dispatchCommand(es, !0));
			let a = n.pop();
			n.length === 0 && e.dispatchCommand($o, !1), t.current = a || null, a && a.editor.setEditorState(a.editorState, { tag: "historic" });
		}
	}(e, t), !0)), 0), e.registerCommand(Zo, (() => (tp(t), !1)), 0), e.registerCommand(Qo, (() => (tp(t), e.dispatchCommand($o, !1), e.dispatchCommand(es, !1), !0)), 0), e.registerUpdateListener((({ editorState: n, prevEditorState: i, dirtyLeaves: a, dirtyElements: o, tags: s }) => {
		let c = t.current, l = t.redoStack, u = t.undoStack, d = c === null ? null : c.editorState;
		if (c !== null && n === d) return;
		let f = r(i, n, c, a, o, s);
		if (f === Kf) l.length !== 0 && (t.redoStack = [], e.dispatchCommand($o, !1)), c !== null && (u.push({ ...c }), e.dispatchCommand(es, !0));
		else if (f === qf) return;
		t.current = {
			editor: e,
			editorState: n
		};
	})));
}
function rp() {
	return {
		current: null,
		redoStack: [],
		undoStack: []
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalHistoryPlugin.prod.mjs
function ip({ delay: e, externalHistoryState: t }) {
	let [n] = Oi();
	return function(e, t, n = 1e3) {
		let r = (0, N.useMemo)((() => t || rp()), [t]);
		(0, N.useEffect)((() => np(e, r, n)), [
			n,
			e,
			r
		]);
	}(n, t, e), null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@lexical/react/0.35.0/9fa3e5df6b50a8023b7692cb94539c161ec97bc43fcbe1a0f671d4b3338b27f1/node_modules/@lexical/react/LexicalErrorBoundary.prod.mjs
function ap(e, t) {
	return ap = Object.setPrototypeOf ? Object.setPrototypeOf.bind() : function(e, t) {
		return e.__proto__ = t, e;
	}, ap(e, t);
}
var op = { error: null }, sp = function(e) {
	var t, n;
	function r() {
		for (var t, n = arguments.length, r = Array(n), i = 0; i < n; i++) r[i] = arguments[i];
		return (t = e.call.apply(e, [this].concat(r)) || this).state = op, t.resetErrorBoundary = function() {
			for (var e, n = arguments.length, r = Array(n), i = 0; i < n; i++) r[i] = arguments[i];
			t.props.onReset == null || (e = t.props).onReset.apply(e, r), t.reset();
		}, t;
	}
	n = e, (t = r).prototype = Object.create(n.prototype), t.prototype.constructor = t, ap(t, n), r.getDerivedStateFromError = function(e) {
		return { error: e };
	};
	var i = r.prototype;
	return i.reset = function() {
		this.setState(op);
	}, i.componentDidCatch = function(e, t) {
		var n, r;
		(n = (r = this.props).onError) == null || n.call(r, e, t);
	}, i.componentDidUpdate = function(e, t) {
		var n, r, i, a, o = this.state.error, s = this.props.resetKeys;
		o !== null && t.error !== null && ((i = e.resetKeys) === void 0 && (i = []), (a = s) === void 0 && (a = []), i.length !== a.length || i.some((function(e, t) {
			return !Object.is(e, a[t]);
		}))) && ((n = (r = this.props).onResetKeysChange) == null || n.call(r, e.resetKeys, s), this.reset());
	}, i.render = function() {
		var e = this.state.error, t = this.props, n = t.fallbackRender, r = t.FallbackComponent, i = t.fallback;
		if (e !== null) {
			var a = {
				error: e,
				resetErrorBoundary: this.resetErrorBoundary
			};
			if (N.isValidElement(i)) return i;
			if (typeof n == "function") return n(a);
			if (r) return N.createElement(r, a);
			throw Error("react-error-boundary requires either a fallback, fallbackRender, or FallbackComponent prop");
		}
		return this.props.children;
	}, r;
}(N.Component);
function cp({ children: e, onError: t }) {
	return (0, P.jsx)(sp, {
		fallback: (0, P.jsx)("div", {
			style: {
				border: "1px solid #f00",
				color: "#f00",
				padding: "8px"
			},
			children: "An error was thrown."
		}),
		onError: t,
		children: e
	});
}
// ../../../clicky-ui/packages/ui/src/comments/comment-vocabulary.ts
function lp(e) {
	return `/?vocabulary=${encodeURIComponent(`${e.catalog}:${e.id}`)}`;
}
function up(e, t) {
	let n = t.toLowerCase().replace(/^lines(?=\/|$)/, "line");
	return n.includes("/") ? e.filter((e) => `${e.id} ${e.label}`.toLowerCase().includes(n)).map((e) => ({
		label: e.label,
		item: e
	})) : [...new Set(e.map((e) => e.type))].filter((e) => e.includes(n)).map((e) => ({
		label: e,
		query: `${e}/`
	}));
}
function dp(e, t) {
	let n = /* @__PURE__ */ new Map(), r = e.replace(/```[\s\S]*?(?:```|$)|`[^`\n]*`/g, "");
	for (let e of r.matchAll(/\[[^\]\n]*\]\((\/\?vocabulary=[^)\s]+)\)/g)) {
		let r = e[1], i = t.find((e) => lp(e) === r);
		if (!i) throw Error(`Unknown vocabulary reference: ${r}`);
		n.set(`${i.catalog}:${i.id}`, i);
	}
	return [...n.values()];
}
// ../../../clicky-ui/packages/ui/src/comments/VocabularyChip.tsx
function fp({ item: e, vocabulary: t }) {
	return /* @__PURE__ */ (0, P.jsx)(O, {
		placement: "top",
		trigger: /* @__PURE__ */ (0, P.jsxs)("span", {
			tabIndex: 0,
			className: "inline-flex max-w-full items-center gap-1 rounded border border-border bg-muted px-1 align-middle text-xs",
			"aria-label": e.id,
			children: [t && /* @__PURE__ */ (0, P.jsx)("span", {
				"aria-hidden": !0,
				className: "pointer-events-none relative inline-flex h-5 w-8 overflow-hidden",
				children: /* @__PURE__ */ (0, P.jsx)("span", {
					className: "absolute left-0 top-0 origin-top-left scale-[0.2]",
					children: t.renderPreview(e)
				})
			}), /* @__PURE__ */ (0, P.jsx)("span", { children: e.label })]
		}),
		children: /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "max-w-80 space-y-2 p-2",
			children: [
				t?.renderPreview(e),
				/* @__PURE__ */ (0, P.jsx)("strong", {
					className: "text-xs",
					children: e.id
				}),
				/* @__PURE__ */ (0, P.jsx)("p", {
					className: "text-xs",
					children: e.description
				}),
				/* @__PURE__ */ (0, P.jsx)("a", {
					href: lp(e),
					target: "_blank",
					rel: "noreferrer",
					className: "text-xs text-primary underline",
					children: "Open reference"
				}),
				e.references.map((e) => /* @__PURE__ */ (0, P.jsx)("div", {
					className: "break-all font-mono text-[10px] text-muted-foreground",
					children: e
				}, e))
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentEditor.tsx
var pp = class e extends sl {
	__item;
	static getType() {
		return "comment-vocabulary";
	}
	static clone(t) {
		return new e(t.__item, t.__key);
	}
	constructor(e, t) {
		super(t), this.__item = e;
	}
	createDOM() {
		return document.createElement("span");
	}
	exportDOM() {
		let e = document.createElement("a");
		return e.href = lp(this.__item), e.textContent = this.__item.label, { element: e };
	}
	updateDOM() {
		return !1;
	}
	isInline() {
		return !0;
	}
	getTextContent() {
		return `[${this.__item.label}](${lp(this.__item)})`;
	}
	exportJSON() {
		return {
			...super.exportJSON(),
			item: this.__item
		};
	}
	static importJSON(t) {
		return new e(t.item);
	}
	decorate() {
		return /* @__PURE__ */ (0, P.jsx)(hp, { item: this.__item });
	}
}, mp = (0, N.createContext)(void 0);
function hp({ item: e }) {
	let t = (0, N.useContext)(mp);
	return /* @__PURE__ */ (0, P.jsx)(fp, {
		item: e,
		...t ? { vocabulary: t } : {}
	});
}
function gp(e, t) {
	if (!t) return [W(e)];
	let n = [], r = dp(e, t?.items ?? []), i = 0, a = e.replace(/```[\s\S]*?(?:```|$)|`[^`\n]*`/g, (e) => " ".repeat(e.length));
	for (let t of a.matchAll(/\[[^\]\n]+\]\((\/\?vocabulary=[^)\s]+)\)/g)) {
		let a = r.find((e) => lp(e) === t[1]);
		a && (n.push(W(e.slice(i, t.index)), new pp(a)), i = t.index + t[0].length);
	}
	return n.push(W(e.slice(i))), n;
}
function _p(e, t) {
	Z().clear().append(yl().append(...gp(e, t)));
}
function vp(e) {
	let [t] = Oi(), n = (0, N.useRef)(e.value), [r, i] = (0, N.useState)({
		query: null,
		activeIndex: 0
	}), { query: a, activeIndex: o } = r, s = (e) => i((t) => t.query === e ? t : {
		query: e,
		activeIndex: 0
	}), c = (e) => i((t) => ({
		...t,
		activeIndex: e
	})), l = (0, N.useId)(), u = b(), { refs: d, floatingStyles: f } = h({
		strategy: "fixed",
		middleware: [
			Le(6),
			ze(),
			Be({ padding: 8 })
		],
		whileElementsMounted: Re
	}), p = e.vocabulary?.items, { choices: m, results: g } = (0, N.useMemo)(() => {
		let t = a === null ? [] : up(p ?? [], a), n = a === null || a.includes("/") ? [] : e.mentionables.filter((e) => e.name.toLowerCase().includes(a.toLowerCase()));
		return {
			choices: t,
			results: [...t, ...n.map((e) => ({
				label: `@${e.name}`,
				person: e
			}))]
		};
	}, [
		a,
		p,
		e.mentionables
	]), _ = a !== null;
	(0, N.useEffect)(() => {
		t.setEditable(!e.disabled);
	}, [t, e.disabled]), (0, N.useEffect)(() => {
		e.value !== n.current && (n.current = e.value, t.update(() => _p(e.value, e.vocabulary), { discrete: !0 }));
	}, [
		t,
		e.value,
		e.vocabulary
	]), (0, N.useEffect)(() => {
		e.autoFocus && (t.update(() => Z().selectEnd(), { discrete: !0 }), t.getRootElement()?.focus());
	}, [t, e.autoFocus]), (0, N.useEffect)(() => t.registerCommand(wo, (t) => {
		if (!(t instanceof ClipboardEvent)) return !1;
		let n = t.clipboardData?.getData("text/plain");
		if (!n?.includes("/?vocabulary=")) return !1;
		let r = q();
		return K(r) ? (r.insertNodes(gp(n, e.vocabulary)), t.preventDefault(), !0) : !1;
	}, 3), [t, e.vocabulary]), (0, N.useEffect)(() => {
		let e = () => s(null), n = t.registerRootListener((t, n) => {
			n?.removeEventListener("blur", e), t?.addEventListener("blur", e);
		});
		return () => {
			t.getRootElement()?.removeEventListener("blur", e), n();
		};
	}, [t]), (0, N.useEffect)(() => t.registerUpdateListener(({ editorState: t }) => {
		t.read(() => {
			let t = Z().getTextContent();
			t !== n.current && (n.current = t, e.onChange(t));
			let r = q(), i = K(r) && r.isCollapsed() ? r.anchor.getNode() : null, a = (G(i) && K(r) ? i.getTextContent().slice(0, r.anchor.offset).match(/(?:^|\s)@([\w./-]*)$/) : null)?.[1] ?? null;
			s(a);
			let o = window.getSelection();
			if (a !== null && o?.rangeCount && o.getRangeAt(0).getBoundingClientRect) {
				let e = o.getRangeAt(0).getBoundingClientRect();
				d.setPositionReference({ getBoundingClientRect: () => e });
			}
		});
	}), [
		t,
		e.onChange,
		d
	]);
	function v(n) {
		let r = g[n];
		r && a !== null && t.update(() => {
			let t = q();
			if (!K(t)) return;
			let n = t.anchor.getNode();
			if (!G(n)) return;
			let i = t.anchor.offset, o = n.select(i - a.length - 1, i);
			"person" in r ? (o.insertText(`@${r.person.name} `), e.onMentionSelect?.(r.person)) : r.item ? o.insertNodes([new pp(r.item), W(" ")]) : o.insertText(`@${r.query}`);
		});
	}
	(0, N.useEffect)(() => t.registerCommand(jo, (t) => {
		if (t.isComposing) return !1;
		if (_) {
			if (t.key === "Escape") return t.preventDefault(), t.stopPropagation(), s(null), !0;
			if ([
				"ArrowUp",
				"ArrowDown",
				"Enter",
				"Tab"
			].includes(t.key)) {
				if (t.preventDefault(), t.stopPropagation(), t.key === "ArrowUp" || t.key === "ArrowDown") {
					let e = t.key === "ArrowDown" ? 1 : -1;
					i((t) => ({
						...t,
						activeIndex: (t.activeIndex + e + g.length) % Math.max(1, g.length)
					}));
				} else v(o);
				return !0;
			}
		}
		return t.key === "Enter" && !t.shiftKey ? (t.preventDefault(), e.onSubmit?.(), !0) : t.key === "Escape" && (t.preventDefault(), e.onCancel?.(), !0);
	}, 3), [
		t,
		_,
		o,
		a,
		g,
		e.onSubmit,
		e.onCancel
	]), (0, N.useEffect)(() => {
		let e = t.getRootElement();
		e?.setAttribute("aria-expanded", String(_)), _ ? e?.setAttribute("aria-controls", l) : e?.removeAttribute("aria-controls"), _ && g.length ? e?.setAttribute("aria-activedescendant", `${l}-${o}`) : e?.removeAttribute("aria-activedescendant"), document.getElementById(`${l}-${o}`)?.scrollIntoView?.({ block: "nearest" });
	}, [
		t,
		_,
		o,
		l,
		g.length
	]);
	let y = m[o]?.item;
	return _ && (0, Nf.createPortal)(/* @__PURE__ */ (0, P.jsxs)("div", {
		ref: d.setFloating,
		style: {
			...f,
			zIndex: u
		},
		className: "flex max-h-80 w-[min(36rem,calc(100vw-1rem))] overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-lg",
		onMouseDown: (e) => e.preventDefault(),
		children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			role: "listbox",
			"aria-label": "Vocabulary and mentions",
			id: l,
			className: "min-w-0 flex-1 overflow-auto p-1",
			children: [g.length === 0 && /* @__PURE__ */ (0, P.jsx)("p", {
				role: "status",
				className: "p-2 text-xs",
				children: "No matching items"
			}), g.map((e, t) => /* @__PURE__ */ (0, P.jsx)("button", {
				id: `${l}-${t}`,
				role: "option",
				"aria-selected": t === o,
				type: "button",
				onMouseEnter: () => c(t),
				onClick: () => v(t),
				className: `block w-full rounded px-2 py-1.5 text-left text-xs ${t === o ? "bg-accent text-accent-foreground" : ""}`,
				children: e.label
			}, `${e.label}-${t}`))]
		}), y && /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "w-1/2 overflow-auto border-l border-border p-3 text-xs",
			children: [/* @__PURE__ */ (0, P.jsx)("div", {
				className: "overflow-hidden",
				children: e.vocabulary?.renderPreview(y)
			}), /* @__PURE__ */ (0, P.jsx)("p", {
				className: "mt-2",
				children: y.description
			})]
		})]
	}), document.body);
}
function yp(e) {
	return /* @__PURE__ */ (0, P.jsx)(mp.Provider, {
		value: e.vocabulary,
		children: /* @__PURE__ */ (0, P.jsxs)(Xd, {
			initialConfig: {
				namespace: "comment",
				nodes: [pp],
				onError: (e) => {
					throw e;
				},
				editorState: () => _p(e.value, e.vocabulary)
			},
			children: [
				/* @__PURE__ */ (0, P.jsx)("div", {
					className: "relative",
					children: /* @__PURE__ */ (0, P.jsx)(If, {
						contentEditable: /* @__PURE__ */ (0, P.jsx)(Hf, {
							role: "textbox",
							"aria-label": e.placeholder ?? "Comment",
							"aria-describedby": e["aria-describedby"],
							"aria-multiline": !0,
							"data-testid": e["data-testid"],
							className: `w-full whitespace-pre-wrap bg-transparent text-sm outline-none [@media(pointer:coarse)]:text-base ${e.className ?? ""}`
						}),
						placeholder: /* @__PURE__ */ (0, P.jsx)("span", {
							className: "pointer-events-none absolute left-0 top-0 text-sm text-placeholder",
							children: e.placeholder
						}),
						ErrorBoundary: cp
					})
				}),
				/* @__PURE__ */ (0, P.jsx)(ip, {}),
				/* @__PURE__ */ (0, P.jsx)(vp, { ...e })
			]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/comments/DefaultDeliveryLine.tsx
function bp({ delivery: e, className: t }) {
	return e ? /* @__PURE__ */ (0, P.jsx)("p", {
		"data-testid": "default-delivery",
		...e.description ? { title: e.description } : {},
		className: t ?? "text-[11px] leading-4 text-muted-foreground",
		children: `Comments without an @target go to ${e.label}${e.layer ? ` (rule from ${e.layer})` : ""}`
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentComposer.tsx
var xp = {
	neutral: "bg-muted-foreground",
	success: "bg-green-500",
	danger: "bg-red-500",
	warning: "bg-yellow-500",
	info: "bg-blue-500"
};
function Sp({ facet: e, value: t, onChange: n }) {
	let r = e.options.find((e) => e.value === t);
	return /* @__PURE__ */ (0, P.jsx)(te, {
		align: "left",
		menuClassName: "min-w-[150px]",
		trigger: /* @__PURE__ */ (0, P.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: a("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium transition-colors", r ? "bg-muted text-foreground" : "text-muted-foreground hover:bg-muted/60"),
			children: [r ? r.label : e.label, r && /* @__PURE__ */ (0, P.jsx)("span", {
				role: "presentation",
				className: "ml-0.5 cursor-pointer hover:text-foreground",
				onClick: (e) => {
					e.stopPropagation(), n(void 0);
				},
				children: "×"
			})]
		}),
		children: (r) => /* @__PURE__ */ (0, P.jsx)("div", {
			className: "py-1",
			children: e.options.map((e) => /* @__PURE__ */ (0, P.jsxs)("button", {
				type: "button",
				onClick: () => {
					n(e.value === t ? void 0 : e.value), r();
				},
				className: a("flex w-full items-center gap-2 px-3 py-1.5 text-left text-[11px] text-popover-foreground hover:bg-accent hover:text-accent-foreground", e.value === t && "bg-muted font-semibold"),
				children: [/* @__PURE__ */ (0, P.jsx)("span", { className: a("inline-block size-2 rounded-full", xp[ui(e.tone)]) }), e.label]
			}, e.value))
		})
	});
}
function Cp(e) {
	return e === void 0 ? {
		text: "Send to…",
		label: "Send to…",
		title: "Send this comment to an agent, or type @ and its name"
	} : e === null ? {
		text: "Not delivered",
		label: "Send to… (not delivered by default)",
		title: "No agent rule covers this page, so the comment is only posted. Pick a target, or type @ and its name, to deliver it"
	} : {
		text: `Auto: ${e.label}`,
		label: `Send to… (default: ${e.label})`,
		title: e.description ?? `New comments here are delivered to ${e.label}. Pick a target to send this one elsewhere`
	};
}
function wp({ targets: e, target: t, defaultDelivery: n, onPick: r, onClear: i }) {
	let o = Cp(n);
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		className: "inline-flex items-center",
		children: [/* @__PURE__ */ (0, P.jsx)(te, {
			align: "left",
			menuClassName: "min-w-[150px]",
			menuLabel: "Comment targets",
			items: e.map((e) => ({
				label: `@${e.name}`,
				onSelect: () => r(e)
			})),
			trigger: /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				"aria-label": t ? `Send to @${t.name}` : o.label,
				title: t ? `This comment is sent to @${t.name}` : o.title,
				className: a("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium transition-colors [@media(pointer:coarse)]:min-h-11", t ? "bg-primary/10 text-primary" : "text-muted-foreground hover:bg-muted/60", !t && n && "bg-muted/60 text-foreground"),
				children: t ? `@${t.name}` : o.text
			})
		}), t && /* @__PURE__ */ (0, P.jsx)("button", {
			type: "button",
			"aria-label": "Clear target",
			onClick: i,
			className: "inline-flex items-center justify-center rounded-full p-1 text-muted-foreground hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11",
			children: /* @__PURE__ */ (0, P.jsx)(c, {
				icon: d,
				className: "text-[10px]"
			})
		})]
	});
}
function Tp({ config: e, anchor: t, collapsible: n = !0, autoFocus: r = !1, placeholder: i = "Add a comment…", onCreate: a, createActions: o, onMention: s, onCancel: l }) {
	let u = `draft:create:${t ?? "general"}`, [f, p] = yt(`${u}:open`, !n || r), [m, h] = yt(`${u}:body`, ""), [g, _] = yt(`${u}:facets`, {}), [v, y] = yt(`${u}:target`, void 0), b = e.targets ?? [], x = b.find((e) => e.id === v), [S, C] = (0, N.useState)(!1), [w, T] = (0, N.useState)(null), E = (0, N.useRef)(!1), ee = (0, N.useRef)(!0);
	(0, N.useEffect)(() => (ee.current = !0, () => {
		ee.current = !1;
	}), []);
	let [te, ne] = yt(`${u}:mentions`, []), D = (0, N.useRef)(new Map(te.map((e) => [e.id, e])));
	function O() {
		h(""), _({}), y(void 0), D.current.clear(), ne([]);
	}
	function re(e, t) {
		_((n) => {
			if (t === void 0) {
				let { [e]: t, ...r } = n;
				return r;
			}
			return {
				...n,
				[e]: t
			};
		});
	}
	async function k(e) {
		let r = m.trim();
		if (!r || E.current) return;
		E.current = !0, C(!0), T(null);
		let i = [...D.current.values()].filter((e) => e.id !== x?.id && m.toLowerCase().includes(`@${e.name.toLowerCase()}`)), o = {
			body: m,
			facets: g,
			pickedTargetId: v,
			mentions: [...D.current.values()]
		};
		O();
		try {
			let o = {
				body: r,
				anchor: t ?? null,
				...Object.keys(g).length > 0 ? { facets: g } : {},
				...i.length > 0 ? { mentions: i } : {},
				...x ? { target: {
					id: x.id,
					name: x.name,
					kind: x.kind
				} } : {}
			};
			e === void 0 ? await a?.(o) : await a?.(o, e);
			for (let e of i) s?.(e, { body: r });
			n && p(!1);
		} catch (e) {
			if (!ee.current) return;
			h(o.body), _(o.facets), y(o.pickedTargetId), D.current = new Map(o.mentions.map((e) => [e.id, e])), ne(o.mentions), p(!0), T(e);
		} finally {
			E.current = !1, C(!1);
		}
	}
	return n && !f && m.trim().length === 0 ? /* @__PURE__ */ (0, P.jsx)("button", {
		type: "button",
		onClick: () => p(!0),
		className: "w-full rounded-full border border-border bg-muted/25 px-4 py-2 text-left text-sm text-muted-foreground/70 shadow-sm transition-colors hover:border-primary/30 hover:bg-muted/40",
		children: i
	}) : /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-testid": "comment-compose-box",
		"aria-busy": S,
		className: "relative rounded-2xl border border-border bg-muted/25 px-3 pt-3 pb-10 shadow-sm transition-[border-color,box-shadow] focus-within:border-primary/40 focus-within:shadow-md [@media(pointer:coarse)]:pb-2",
		children: [
			/* @__PURE__ */ (0, P.jsx)(yp, {
				disabled: S,
				...e.vocabulary ? { vocabulary: e.vocabulary } : {},
				value: m,
				onChange: h,
				mentionables: [...e.mentionables ?? [], ...b.filter((t) => !e.mentionables?.some((e) => e.name.toLowerCase() === t.name.toLowerCase()))],
				onMentionSelect: (e) => {
					D.current.set(e.id, e), ne([...D.current.values()]);
				},
				onSubmit: () => void k(o?.[0]?.id),
				onCancel: () => {
					E.current || (O(), n && p(!1), l?.());
				},
				placeholder: i,
				autoFocus: r || n && f,
				"data-testid": "comment-compose-input",
				className: "min-h-[48px] pr-10 [@media(pointer:coarse)]:pr-12"
			}),
			w !== null && /* @__PURE__ */ (0, P.jsx)(F, {
				error: w,
				context: [["Operation", "Post comment"], ["Anchor", t ?? "Whole page"]]
			}),
			n && /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: () => {
					O(), p(!1);
				},
				"aria-label": "Close composer",
				className: "absolute right-3 top-3 inline-flex items-center justify-center rounded-full p-1 text-muted-foreground hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:right-1 [@media(pointer:coarse)]:top-1 [@media(pointer:coarse)]:size-11",
				children: /* @__PURE__ */ (0, P.jsx)(c, {
					icon: d,
					className: "text-xs"
				})
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				"data-comment-actions": !0,
				className: "absolute inset-x-3 bottom-2.5 flex items-center gap-1.5 [@media(pointer:coarse)]:static [@media(pointer:coarse)]:mt-2 [@media(pointer:coarse)]:flex-wrap [@media(pointer:coarse)]:gap-2",
				children: [
					b.length > 0 && /* @__PURE__ */ (0, P.jsx)(wp, {
						targets: b,
						target: x,
						defaultDelivery: e.defaultDelivery,
						onPick: (e) => y(e.id),
						onClear: () => y(void 0)
					}),
					(e.facets ?? []).map((e) => /* @__PURE__ */ (0, P.jsx)(Sp, {
						facet: e,
						value: g[e.key],
						onChange: (t) => re(e.key, t)
					}, e.key)),
					/* @__PURE__ */ (0, P.jsx)("div", {
						className: "flex min-w-0 flex-1 items-center overflow-hidden",
						children: e.renderDeliveryPreview ? e.renderDeliveryPreview({
							body: m,
							...x ? { target: x } : {}
						}) : /* @__PURE__ */ (0, P.jsx)(bp, {
							delivery: e.defaultDelivery,
							className: "truncate text-[11px] leading-4 text-muted-foreground"
						})
					}),
					/* @__PURE__ */ (0, P.jsx)(ie, {
						className: "ml-auto shrink-0 [@media(pointer:coarse)]:min-h-11",
						size: "sm",
						label: /* @__PURE__ */ (0, P.jsx)("span", {
							className: "sr-only",
							children: o?.[0]?.label ?? "Post comment"
						}),
						icon: et,
						loading: S,
						title: "Send options",
						disabled: S || !m.trim(),
						onClick: () => void k(o?.[0]?.id),
						items: (o?.slice(1) ?? []).map((e) => ({
							label: e.label,
							onSelect: () => {
								k(e.id);
							}
						}))
					})
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentReplyInput.tsx
function Ep({ parentId: e, config: t, onSubmit: n, onCancel: r, pinned: i = !1 }) {
	let a = (0, N.useId)(), [s, c] = yt(`draft:reply:${e}:body`, ""), [l, u] = (0, N.useState)(!1), [d, f] = (0, N.useState)(null), p = (0, N.useRef)(!1);
	async function m() {
		if (s.trim() && !p.current) {
			p.current = !0, u(!0), f(null);
			try {
				await n(s.trim()), c("");
			} catch (e) {
				f(e);
			} finally {
				p.current = !1, u(!1);
			}
		}
	}
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-1.5",
		"aria-busy": l,
		"data-comment-reply-bar": i || void 0,
		children: [
			i && /* @__PURE__ */ (0, P.jsx)("label", {
				id: a,
				className: "block text-xs font-medium",
				children: "Reply to thread"
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex items-end gap-2 py-1",
				children: [/* @__PURE__ */ (0, P.jsx)("div", {
					className: "min-w-0 flex-1",
					children: /* @__PURE__ */ (0, P.jsx)(yp, {
						...t.vocabulary ? { vocabulary: t.vocabulary } : {},
						mentionables: t.mentionables ?? [],
						autoFocus: !i,
						value: s,
						disabled: l,
						onChange: c,
						onSubmit: () => void m(),
						...i ? {} : { onCancel: r },
						placeholder: "Write a reply…",
						...i ? { "aria-describedby": a } : {},
						"data-testid": "comment-reply-input",
						className: "min-h-9 max-h-32 min-w-0 flex-1 overflow-y-auto rounded-md border border-border bg-background px-2 py-1 text-xs outline-none focus:border-primary/40 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:text-base"
					})
				}), /* @__PURE__ */ (0, P.jsx)(o, {
					size: "sm",
					disabled: l || !s.trim(),
					onClick: () => void m(),
					children: l ? "Sending…" : "Send"
				})]
			}),
			(t.mentionables?.length ?? 0) > 0 && /* @__PURE__ */ (0, P.jsx)("p", {
				"data-testid": "reply-delivery",
				className: "text-[11px] leading-4 text-muted-foreground",
				children: "Replies are delivered only when they name an @target"
			}),
			d !== null && /* @__PURE__ */ (0, P.jsx)(F, {
				error: d,
				context: [["Operation", "Send reply"], ["Thread", e]]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/agent-delivery.ts
function Dp(e) {
	return e.state === "error" ? "failed" : e.repliedAt || e.state === "done" ? "replied" : e.state === "sent" ? "sent" : "queued";
}
function Op(e, t = Ie(e)) {
	let n = Pe(t), r = Ae(t);
	return n ? {
		icon: n,
		...r ? { className: r } : {}
	} : { icon: me };
}
// ../../../clicky-ui/packages/ui/src/data/chat/agent-identity-context.ts
var kp = (0, N.createContext)({});
// ../../../clicky-ui/packages/ui/src/data/chat/AgentIdentity.tsx
function Ap({ model: e, provider: t, effort: n, label: r, identityId: i, showSession: o = !1, iconOnly: s = !1, hideIcon: l = !1, className: u }) {
	let d = (0, N.useContext)(kp), f = i ? d[i] : void 0, p = f?.model ?? e, m = f?.provider ?? t, h = f?.effort ?? n, g = m ?? (p ? Ie(p) : void 0), _ = h ? Fe(h) : void 0, v = (o ? f?.sessionName : void 0) ?? r;
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		className: a("inline-flex min-w-0 items-center gap-1.5", u),
		title: [
			v,
			o && i,
			p,
			m,
			h && `${h} effort`
		].filter(Boolean).join(" · "),
		children: [
			!l && /* @__PURE__ */ (0, P.jsx)(c, {
				icon: Pe(g) ?? me,
				className: a("size-3.5 shrink-0", Ae(g))
			}),
			!s && /* @__PURE__ */ (0, P.jsx)("span", {
				className: "truncate",
				children: v ?? (p ? Ie(p) ? je(p) : p : m ?? "Agent")
			}),
			!s && o && i && v !== i && /* @__PURE__ */ (0, P.jsx)("code", {
				className: "shrink-0 text-muted-foreground",
				children: Ne(i)
			}),
			h && /* @__PURE__ */ (0, P.jsx)("span", {
				title: `${Oe(h)} effort`,
				"aria-label": `${Oe(h)} effort`,
				children: _ ? /* @__PURE__ */ (0, P.jsx)(_, { className: a("size-3.5", Me(h)) }) : /* @__PURE__ */ (0, P.jsx)("span", {
					className: "text-xs",
					children: h
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentAuthor.tsx
function jp({ author: e }) {
	return e?.kind === "agent" ? /* @__PURE__ */ (0, P.jsx)(Ap, {
		hideIcon: !0,
		showSession: !0,
		model: e.model ?? e.name,
		...e.model ? { label: e.name } : {},
		...e.effort ? { effort: e.effort } : {},
		...e.sessionId ? { identityId: e.sessionId } : {}
	}) : /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: li(e) });
}
function Mp({ author: e, size: t = "sm", bare: n = !1 }) {
	let r = s(), i = e?.name ?? li(e), a = e?.kind === "agent", o = a ? /* @__PURE__ */ (0, P.jsx)(Np, {
		author: e,
		size: l(t, r)
	}) : /* @__PURE__ */ (0, P.jsx)($e, {
		alt: i,
		size: t,
		...e?.avatar ? { src: e.avatar } : {},
		title: i
	});
	return n ? o : /* @__PURE__ */ (0, P.jsxs)(O, {
		trigger: o,
		placement: "top",
		children: [/* @__PURE__ */ (0, P.jsx)("span", {
			className: "font-medium",
			children: i
		}), a && /* @__PURE__ */ (0, P.jsx)("span", {
			className: "ml-1 text-muted-foreground",
			children: "· agent"
		})]
	});
}
function Np({ author: e, size: t }) {
	let n = e.icon, r = (0, N.useContext)(kp), i = e.sessionId ? r[e.sessionId] : void 0, a = Op(i?.model ?? e.model ?? e.name, i?.provider);
	return /* @__PURE__ */ (0, P.jsx)("span", {
		role: "img",
		"aria-label": e.name,
		className: "inline-flex shrink-0 items-center justify-center rounded-full border border-border bg-muted text-foreground",
		style: {
			width: t,
			height: t
		},
		children: /* @__PURE__ */ (0, P.jsx)(c, {
			...typeof n == "string" ? { name: n } : { icon: n ?? a.icon },
			...a.className ? { className: a.className } : {},
			width: Math.max(10, Math.round(t * .6)),
			height: Math.max(10, Math.round(t * .6))
		})
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentBodyEditor.tsx
function Pp({ comment: e, config: t, onSave: n, onDone: r }) {
	let [i, a] = (0, N.useState)(e.body), [s, c] = (0, N.useState)(!1), [l, u] = (0, N.useState)(null), d = (0, N.useRef)(!1);
	async function f() {
		let t = i.trim();
		if (t && !d.current) {
			if (t === e.body.trim()) return r();
			d.current = !0, c(!0), u(null);
			try {
				await n(t), r();
			} catch (e) {
				u(e);
			} finally {
				d.current = !1, c(!1);
			}
		}
	}
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-1.5",
		"aria-busy": s,
		children: [
			/* @__PURE__ */ (0, P.jsx)(yp, {
				...t.vocabulary ? { vocabulary: t.vocabulary } : {},
				mentionables: t.mentionables ?? [],
				autoFocus: !0,
				value: i,
				disabled: s,
				onChange: a,
				onSubmit: () => void f(),
				onCancel: () => {
					d.current || r();
				},
				"data-testid": "comment-edit-input",
				className: "min-h-9 max-h-48 w-full overflow-y-auto rounded-md border border-border bg-background px-2 py-1 text-[13px] outline-none focus:border-primary/40 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:text-base"
			}),
			l !== null && /* @__PURE__ */ (0, P.jsx)(F, {
				error: l,
				context: [
					["Operation", "Edit comment"],
					["Comment", e.id],
					["Anchor", e.anchor ?? "Whole page"]
				]
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex justify-end gap-1.5",
				children: [/* @__PURE__ */ (0, P.jsx)(o, {
					size: "sm",
					variant: "ghost",
					disabled: s,
					onClick: r,
					children: "Cancel"
				}), /* @__PURE__ */ (0, P.jsx)(o, {
					size: "sm",
					disabled: s || !i.trim(),
					onClick: () => void f(),
					children: s ? "Saving…" : "Save"
				})]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentCardParts.tsx
function Fp({ decision: e }) {
	return e ? /* @__PURE__ */ (0, P.jsx)(E, {
		variant: "soft",
		tone: e === "approved" ? "success" : e === "rejected" ? "danger" : "warning",
		size: "xs",
		children: e === "approved" ? "Approved" : e === "rejected" ? "Rejected" : "Revision requested"
	}) : null;
}
function Ip({ rating: e }) {
	if (!e) return null;
	let t = e === "positive";
	return /* @__PURE__ */ (0, P.jsx)(E, {
		variant: "soft",
		tone: t ? "success" : "danger",
		size: "xs",
		icon: t ? nt : rt,
		children: t ? "Positive rating" : "Negative rating"
	});
}
var Lp = {
	queued: {
		icon: de,
		className: "text-muted-foreground",
		label: "Queued"
	},
	sent: {
		icon: w,
		className: "text-muted-foreground",
		label: "Sent"
	},
	replied: {
		icon: tt,
		className: "text-primary",
		label: "Replied"
	},
	failed: {
		icon: u,
		className: "text-destructive",
		label: "Hand-off failed"
	}
};
function Rp(e, t, n) {
	let r = (e) => new Date(e).toLocaleString();
	return e === "failed" ? `Hand-off to ${n} failed: ${t.error ?? "no reason recorded"}` : e === "replied" ? t.repliedAt ? `${n} replied ${r(t.repliedAt)}` : `${n} handled this thread` : e === "sent" ? `Sent to ${n} ${r(t.requestedAt)}` : `Queued for ${n} since ${r(t.requestedAt)}`;
}
function zp({ assignee: e, run: t }) {
	let n = (0, N.useContext)(kp), r = t.sessionId ? n[t.sessionId] : void 0, i = Dp(t), o = Op(r?.model ?? e.model ?? e.name), s = Lp[i], l = t.sessionId ? `@${(r?.sessionName ?? t.sessionId).replace(/^@/, "")}` : t.target?.label ?? e.name;
	return /* @__PURE__ */ (0, P.jsx)(E, {
		variant: "soft",
		tone: i === "failed" ? "danger" : "neutral",
		size: "xs",
		children: /* @__PURE__ */ (0, P.jsxs)("span", {
			"data-testid": "agent-delivery",
			"data-delivery": i,
			title: Rp(i, t, l),
			className: "inline-flex min-w-0 items-center gap-1",
			children: [
				/* @__PURE__ */ (0, P.jsx)(c, {
					icon: o.icon,
					className: a("size-3 shrink-0", o.className)
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "max-w-40 truncate",
					children: l
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					role: "img",
					"aria-label": s.label,
					className: "inline-flex shrink-0",
					children: /* @__PURE__ */ (0, P.jsx)(c, {
						icon: s.icon,
						className: a("size-3.5", s.className)
					})
				})
			]
		})
	});
}
function Bp({ comment: e, config: t }) {
	let n = di(t, e.status);
	if (!n) return null;
	let r = Ci(t, e.status), i = r === "closed" ? e.closedBy : void 0, a = r === "active" ? e.assignee : void 0;
	if (a?.kind === "agent" && e.agentRun) return /* @__PURE__ */ (0, P.jsx)(zp, {
		assignee: a,
		run: e.agentRun
	});
	let o = a?.kind === "agent" ? me : n.icon, s = a ? e.acknowledgedAt ? `Acknowledged ${new Date(e.acknowledgedAt).toLocaleString()}` : "Assigned, not yet acknowledged" : e.closedAt ? new Date(e.closedAt).toLocaleString() : void 0;
	return /* @__PURE__ */ (0, P.jsx)(E, {
		variant: "soft",
		tone: ui(n.tone),
		size: "xs",
		...o ? { icon: o } : {},
		...s ? { title: s } : {},
		children: a ? `Assigned to ${a.name}` : i ? `Closed by ${i.name}` : n.label
	});
}
function Vp({ comment: e, config: t, compact: n }) {
	let r = (t.facets ?? []).flatMap((t) => {
		let r = fi(t, e.facets?.[t.key]);
		return r ? [/* @__PURE__ */ (0, P.jsx)(E, {
			variant: "soft",
			tone: ui(r.tone),
			size: "xs",
			children: n ? r.short ?? r.label : r.label
		}, t.key)] : [];
	});
	return r.length > 0 ? /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: r }) : null;
}
function Hp({ comment: e, onChecklistToggle: t }) {
	return e.checklist?.length ? /* @__PURE__ */ (0, P.jsx)("div", {
		className: "flex flex-wrap gap-1",
		children: e.checklist.map((e, n) => {
			let r = /done|complete|resolved/i.test(e.status), i = /progress|doing/i.test(e.status);
			return /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: () => t?.(n),
				disabled: !t,
				className: a("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] transition-opacity", t && "cursor-pointer hover:opacity-80", r ? "bg-green-100 text-green-800 line-through dark:bg-green-500/20 dark:text-green-300" : i ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-muted text-muted-foreground"),
				children: e.label
			}, n);
		})
	}) : null;
}
function Up({ comment: e }) {
	return e.refs?.length ? /* @__PURE__ */ (0, P.jsx)("div", {
		className: "flex flex-wrap gap-1.5",
		children: e.refs.map((e, t) => /* @__PURE__ */ (0, P.jsxs)("span", {
			className: "inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
			children: [/* @__PURE__ */ (0, P.jsxs)("span", {
				className: "font-medium",
				children: [e.label, ":"]
			}), /* @__PURE__ */ (0, P.jsx)("span", {
				className: e.mono ? "font-mono" : void 0,
				children: e.items.join(", ")
			})]
		}, t))
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentMarkdown.tsx
function Wp(e) {
	let t = e.trim();
	return t && (/^(https?:|mailto:)/i.test(t) || t.startsWith("/") && !t.startsWith("//")) ? t : null;
}
var Gp = /(\*\*[^*\n]+\*\*|`[^`\n]+`|\[[^\]\n]+\]\(([^)\s]+)\)|https?:\/\/[^\s<)]+|\*[^*\n]+\*)/g;
function Kp(e) {
	let t = [], n = 0;
	for (let r of e.matchAll(Gp)) {
		let i = r.index ?? 0;
		i > n && t.push({
			type: "text",
			text: e.slice(n, i)
		});
		let a = r[0];
		if (a.startsWith("**") && a.endsWith("**")) t.push({
			type: "strong",
			text: a.slice(2, -2)
		});
		else if (a.startsWith("`") && a.endsWith("`")) t.push({
			type: "code",
			text: a.slice(1, -1)
		});
		else if (a.startsWith("[")) {
			let e = a.indexOf("]("), n = Wp(a.slice(e + 2, -1));
			t.push(n ? {
				type: "link",
				text: a.slice(1, e),
				href: n
			} : {
				type: "text",
				text: a
			});
		} else if (/^https?:\/\//i.test(a)) {
			let e = Wp(a);
			t.push(e ? {
				type: "link",
				text: a,
				href: e
			} : {
				type: "text",
				text: a
			});
		} else a.startsWith("*") && a.endsWith("*") ? t.push({
			type: "em",
			text: a.slice(1, -1)
		}) : t.push({
			type: "text",
			text: a
		});
		n = i + a.length;
	}
	return n < e.length && t.push({
		type: "text",
		text: e.slice(n)
	}), t;
}
function qp({ text: e, vocabulary: t, vocabularyRefs: n }) {
	return /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: Kp(e).map((e, r) => {
		let i = `${e.type}-${r}`;
		switch (e.type) {
			case "strong": return /* @__PURE__ */ (0, P.jsx)("strong", {
				className: "font-semibold",
				children: e.text
			}, i);
			case "em": return /* @__PURE__ */ (0, P.jsx)("em", { children: e.text }, i);
			case "code": return /* @__PURE__ */ (0, P.jsx)("code", {
				className: "rounded bg-muted px-1 py-0.5 font-mono text-[0.9em]",
				children: e.text
			}, i);
			case "link": {
				let r = n?.find((t) => lp(t) === e.href) ?? t?.items.find((t) => lp(t) === e.href);
				return r ? /* @__PURE__ */ (0, P.jsx)(fp, {
					item: r,
					...t ? { vocabulary: t } : {}
				}, i) : /* @__PURE__ */ (0, P.jsx)("a", {
					href: e.href,
					target: "_blank",
					rel: "noreferrer",
					className: "text-primary underline underline-offset-2",
					children: e.text
				}, i);
			}
			default: return /* @__PURE__ */ (0, P.jsx)(N.Fragment, { children: e.text }, i);
		}
	}) });
}
function Jp(e, t, n) {
	return /* @__PURE__ */ (0, P.jsx)("p", { children: e.map((e, t) => /* @__PURE__ */ (0, P.jsxs)(N.Fragment, { children: [t > 0 && /* @__PURE__ */ (0, P.jsx)("br", {}), /* @__PURE__ */ (0, P.jsx)(qp, {
		text: e,
		...n
	})] }, t)) }, t);
}
function Yp(e, t) {
	return /* @__PURE__ */ (0, P.jsx)("pre", {
		className: "overflow-auto rounded bg-muted p-2 text-xs",
		children: /* @__PURE__ */ (0, P.jsx)("code", { children: e.join("\n") })
	}, t);
}
function Xp({ text: e, className: t, ...n }) {
	let r = String(e ?? "").replace(/\r\n?/g, "\n").split("\n"), i = [], o = [], s = [], c = null;
	function l() {
		o.length > 0 && (i.push(Jp(o, `p-${i.length}`, n)), o = []);
	}
	function u() {
		s.length > 0 && (i.push(/* @__PURE__ */ (0, P.jsx)("ul", {
			className: "list-disc space-y-0.5 pl-4",
			children: s.map((e, t) => /* @__PURE__ */ (0, P.jsx)("li", { children: /* @__PURE__ */ (0, P.jsx)(qp, {
				text: e,
				...n
			}) }, t))
		}, `ul-${i.length}`)), s = []);
	}
	for (let e of r) {
		if (e.trim().startsWith("```")) {
			l(), u(), c ? (i.push(Yp(c, `code-${i.length}`)), c = null) : c = [];
			continue;
		}
		if (c) {
			c.push(e);
			continue;
		}
		let t = e.trim();
		if (!t) {
			l(), u();
			continue;
		}
		let n = t.match(/^[-*]\s+(.+)$/);
		n?.[1] ? (l(), s.push(n[1])) : (u(), o.push(t.replace(/^#{1,6}\s+/, "")));
	}
	return l(), u(), c && i.push(Yp(c, `code-${i.length}`)), /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("space-y-2 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", t),
		children: i.length > 0 ? i : null
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentPreview.tsx
function Zp({ id: e, lines: t, children: n }) {
	let r = (0, N.useRef)(null), [i, a] = yt(`preview:${e}`, !1), [o, s] = (0, N.useState)(!1);
	return (0, N.useLayoutEffect)(() => {
		let e = r.current;
		if (!e) return;
		let n = () => {
			let n = Number.parseFloat(getComputedStyle(e).lineHeight);
			s(e.scrollHeight > n * t + 1);
		};
		n();
		let i = typeof ResizeObserver > "u" ? null : new ResizeObserver(n);
		return i?.observe(e), () => i?.disconnect();
	}, [n, t]), /* @__PURE__ */ (0, P.jsxs)("div", { children: [/* @__PURE__ */ (0, P.jsx)("div", {
		ref: r,
		"data-comment-preview": "",
		className: "overflow-hidden",
		style: i ? void 0 : {
			display: "-webkit-box",
			WebkitBoxOrient: "vertical",
			WebkitLineClamp: t,
			maxHeight: `${t * 1.25}rem`
		},
		children: n
	}), o && /* @__PURE__ */ (0, P.jsx)("button", {
		type: "button",
		className: "mt-1 text-xs font-medium text-primary hover:underline",
		"aria-expanded": i,
		onClick: (e) => {
			e.stopPropagation(), a(!i);
		},
		children: i ? "Show less" : "Show more"
	})] });
}
// ../../../clicky-ui/packages/ui/src/comments/ProposalCard.tsx
var Qp = {
	proposed: {
		label: "Proposed",
		tone: "info"
	},
	applied: {
		label: "Applied",
		tone: "success"
	},
	staged: {
		label: "Staged",
		tone: "success"
	},
	rejected: {
		label: "Rejected",
		tone: "danger"
	},
	revision_requested: {
		label: "Revision requested",
		tone: "warning"
	},
	superseded: {
		label: "Superseded",
		tone: "neutral"
	}
}, $p = "min-h-16 w-full resize-y rounded-md border border-border bg-background px-2 py-1.5 text-xs outline-none focus:border-primary/40 disabled:opacity-50 [@media(pointer:coarse)]:text-base";
function em({ proposal: e }) {
	let { state: t, decidedBy: n, decidedAt: r, feedback: i } = e;
	if (t === "proposed") return null;
	let a = ae(r), o = t === "revision_requested" ? "Revisions requested" : Qp[t].label;
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-1 text-xs text-muted-foreground",
		children: [t === "superseded" ? /* @__PURE__ */ (0, P.jsx)("p", { children: "Replaced by a newer proposal." }) : /* @__PURE__ */ (0, P.jsxs)("p", { children: [
			o,
			n && /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [" by ", /* @__PURE__ */ (0, P.jsx)("span", {
				className: "font-medium text-foreground",
				children: /* @__PURE__ */ (0, P.jsx)(jp, { author: n })
			})] }),
			a && /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [" ", /* @__PURE__ */ (0, P.jsx)("span", {
				title: a.toLocaleString(),
				children: k(a)
			})] })
		] }), i && /* @__PURE__ */ (0, P.jsx)("blockquote", {
			className: "whitespace-pre-wrap border-l-2 border-border pl-2 text-foreground",
			children: i
		})]
	});
}
function tm(e, t) {
	let [n, r] = (0, N.useState)(!1), [i, a] = (0, N.useState)({ status: "idle" });
	function o() {
		let o = !n;
		r(o), o && t && i.status !== "loading" && i.status !== "ready" && (a({ status: "loading" }), t(e).then((e) => a({
			status: "ready",
			payload: e
		}), (e) => a({
			status: "error",
			message: e instanceof Error ? e.message : String(e)
		})));
	}
	return {
		expanded: n,
		diff: i,
		toggle: o
	};
}
function nm({ commentId: e, proposal: t, callbacks: n = {}, className: r }) {
	let { onApplyProposal: i, onRejectProposal: s, onRequestRevision: l, loadProposalDiff: u } = n, { expanded: d, diff: f, toggle: p } = tm(e, u), [m, h] = (0, N.useState)(null), [g, _] = (0, N.useState)(""), [v, y] = (0, N.useState)(null), [b, x] = (0, N.useState)(null), S = (0, N.useRef)(!1), C = (0, N.useId)(), w = Qp[t.state], T = t.files.reduce((e, t) => e + t.additions, 0), ee = t.files.reduce((e, t) => e + t.deletions, 0), te = t.state === "proposed" && !!(i || s || l);
	async function D(e, t, n) {
		if (!S.current) {
			S.current = !0, y(e), x(null);
			try {
				await n(), h(null), _("");
			} catch (e) {
				x({
					error: e,
					title: t
				});
			} finally {
				S.current = !1, y(null);
			}
		}
	}
	function O(e) {
		x(null), _(""), h(m === e ? null : e);
	}
	let k = g.trim();
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-testid": "proposal-card",
		"data-proposal-state": t.state,
		"aria-busy": v !== null,
		className: a("space-y-2 rounded-md border border-border bg-card p-2.5 text-xs", r),
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex flex-wrap items-center gap-x-2 gap-y-1",
				children: [
					/* @__PURE__ */ (0, P.jsx)(c, {
						icon: it,
						className: "size-3.5 text-muted-foreground"
					}),
					/* @__PURE__ */ (0, P.jsx)("span", {
						className: "font-semibold text-foreground",
						children: "Proposed change"
					}),
					/* @__PURE__ */ (0, P.jsx)("code", {
						className: "font-mono text-[10px] text-muted-foreground",
						title: t.commit,
						children: t.commit.slice(0, 7)
					}),
					/* @__PURE__ */ (0, P.jsx)(E, {
						variant: "soft",
						tone: w.tone,
						size: "xs",
						children: w.label
					}),
					/* @__PURE__ */ (0, P.jsx)(pt, {
						files: t.files.length,
						additions: T,
						deletions: ee,
						className: "ml-auto text-[11px]"
					})
				]
			}),
			/* @__PURE__ */ (0, P.jsx)("ul", {
				className: "space-y-0.5",
				children: t.files.map((e) => /* @__PURE__ */ (0, P.jsxs)("li", {
					className: "flex items-center justify-between gap-2",
					children: [/* @__PURE__ */ (0, P.jsx)("span", {
						className: "min-w-0 truncate font-mono text-[11px] text-foreground",
						title: e.path,
						children: e.path
					}), /* @__PURE__ */ (0, P.jsx)(pt, {
						additions: e.additions,
						deletions: e.deletions,
						className: "text-[11px]"
					})]
				}, e.path))
			}),
			u && /* @__PURE__ */ (0, P.jsxs)("div", {
				className: "overflow-hidden rounded-md border border-border",
				children: [/* @__PURE__ */ (0, P.jsxs)("button", {
					type: "button",
					"aria-expanded": d,
					onClick: p,
					className: "flex w-full items-center gap-1 px-2 py-1 text-left font-medium text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
					children: [/* @__PURE__ */ (0, P.jsx)(c, {
						icon: d ? ne : re,
						className: "size-3.5"
					}), d ? "Hide diff" : "Show diff"]
				}), d && /* @__PURE__ */ (0, P.jsx)(mt, {
					loading: f.status === "loading" || f.status === "idle",
					payload: f.status === "ready" ? f.payload : null,
					error: f.status === "error" ? f.message : "",
					maxHeightClassName: "max-h-80"
				})]
			}),
			/* @__PURE__ */ (0, P.jsx)(em, { proposal: t }),
			te && /* @__PURE__ */ (0, P.jsxs)("div", {
				className: "space-y-2",
				children: [
					/* @__PURE__ */ (0, P.jsxs)("div", {
						className: "flex flex-wrap items-center gap-2",
						children: [
							i && /* @__PURE__ */ (0, P.jsx)(ie, {
								size: "sm",
								label: "Apply",
								title: "More ways to apply",
								loading: v === "apply" || v === "stage",
								disabled: v !== null,
								onClick: () => void D("apply", "Couldn't apply proposal", () => i(e, "apply")),
								items: [{
									label: "Stage",
									onSelect: () => void D("stage", "Couldn't stage proposal", () => i(e, "stage"))
								}]
							}),
							s && /* @__PURE__ */ (0, P.jsx)(o, {
								size: "sm",
								variant: "outline",
								disabled: v !== null,
								"aria-expanded": m === "reject",
								onClick: () => O("reject"),
								children: "Reject"
							}),
							l && /* @__PURE__ */ (0, P.jsx)(o, {
								size: "sm",
								variant: "outline",
								disabled: v !== null,
								"aria-expanded": m === "revise",
								onClick: () => O("revise"),
								children: "Ask for revisions"
							})
						]
					}),
					m === "reject" && s && /* @__PURE__ */ (0, P.jsxs)("div", {
						className: "space-y-1.5",
						children: [
							/* @__PURE__ */ (0, P.jsx)("label", {
								htmlFor: C,
								className: "block font-medium text-foreground",
								children: "Reason (optional)"
							}),
							/* @__PURE__ */ (0, P.jsx)("textarea", {
								id: C,
								className: $p,
								value: g,
								disabled: v !== null,
								onChange: (e) => _(e.target.value)
							}),
							/* @__PURE__ */ (0, P.jsxs)("div", {
								className: "flex gap-2",
								children: [/* @__PURE__ */ (0, P.jsx)(o, {
									size: "sm",
									variant: "destructive",
									loading: v === "reject",
									disabled: v !== null,
									onClick: () => void D("reject", "Couldn't reject proposal", () => s(e, k || void 0)),
									children: "Reject proposal"
								}), /* @__PURE__ */ (0, P.jsx)(o, {
									size: "sm",
									variant: "ghost",
									disabled: v !== null,
									onClick: () => h(null),
									children: "Cancel"
								})]
							})
						]
					}),
					m === "revise" && l && /* @__PURE__ */ (0, P.jsxs)("div", {
						className: "space-y-1.5",
						children: [
							/* @__PURE__ */ (0, P.jsx)("label", {
								htmlFor: C,
								className: "block font-medium text-foreground",
								children: "Revision feedback"
							}),
							/* @__PURE__ */ (0, P.jsx)("textarea", {
								id: C,
								className: $p,
								value: g,
								disabled: v !== null,
								placeholder: "What should change?",
								onChange: (e) => _(e.target.value)
							}),
							/* @__PURE__ */ (0, P.jsxs)("div", {
								className: "flex gap-2",
								children: [/* @__PURE__ */ (0, P.jsx)(o, {
									size: "sm",
									loading: v === "revise",
									disabled: v !== null || !k,
									onClick: () => void D("revise", "Couldn't request revisions", () => l(e, k)),
									children: "Send feedback"
								}), /* @__PURE__ */ (0, P.jsx)(o, {
									size: "sm",
									variant: "ghost",
									disabled: v !== null,
									onClick: () => h(null),
									children: "Cancel"
								})]
							})
						]
					})
				]
			}),
			b && /* @__PURE__ */ (0, P.jsx)(F, {
				error: b.error,
				title: b.title,
				context: [["Comment", e], ["Proposal", t.commit]]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/proposal-context.ts
var rm = (0, N.createContext)({});
function im(e) {
	let { onApplyProposal: t, onRejectProposal: n, onRequestRevision: r, loadProposalDiff: i } = e;
	return (0, N.useMemo)(() => !t && !n && !r && !i ? null : {
		...t ? { onApplyProposal: t } : {},
		...n ? { onRejectProposal: n } : {},
		...r ? { onRequestRevision: r } : {},
		...i ? { loadProposalDiff: i } : {}
	}, [
		t,
		n,
		r,
		i
	]);
}
// ../../../clicky-ui/packages/ui/src/comments/comment-assign.ts
function am(e, t) {
	return e.name === t.name && (e.kind ?? "user") === (t.kind ?? "user");
}
function om(e) {
	let { comment: t, config: n, onAssign: r, extraAssignItems: i, run: a } = e, o = [];
	if (r) {
		for (let { id: e, name: i, kind: s } of n.assignees ?? []) {
			let n = {
				id: e,
				name: i,
				kind: s
			}, c = t.assignee !== void 0 && am(t.assignee, n);
			o.push({
				label: i,
				icon: s === "agent" ? me : at,
				disabled: c,
				title: c ? `Already assigned to ${i}` : `Assign to ${i}`,
				onSelect: () => a("assign", () => r(n))
			});
		}
		t.assignee && o.push({
			label: "Unassign",
			icon: st,
			onSelect: () => a("assign", () => r(null))
		});
	}
	return o.push(...i?.(t) ?? []), o;
}
function sm(e) {
	let { comment: t, config: n, onAcknowledge: r, run: i } = e;
	if (t.parentId || Ci(n, t.status) !== "active") return [];
	let a = [];
	r && !t.assignee && a.push({
		label: "Acknowledge",
		icon: ct,
		onSelect: () => i("acknowledge", r)
	});
	let o = om(e);
	return o.length > 0 && a.push({
		label: "Assign…",
		icon: ot,
		onSelect: () => {},
		children: o
	}), a;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentCard.tsx
function cm(e, t) {
	let n = [];
	return t.onEdit && n.push({
		label: "Edit",
		icon: ke,
		onSelect: t.onEdit
	}), t.onDelete && n.push({
		label: "Delete",
		icon: C,
		onSelect: t.onDelete
	}), e || n.push(...t.statusActions ?? []), !e && t.onCreateTodo && n.push({
		label: "Create todo",
		icon: se,
		onSelect: t.onCreateTodo
	}), t.onCopy && n.push({
		label: "Copy",
		icon: g,
		onSelect: t.onCopy
	}), t.onMaximize && n.push({
		label: "Maximise",
		icon: ee,
		onSelect: t.onMaximize
	}), n;
}
function lm(e, t) {
	let n = ue(e) ?? {
		message: String(e),
		context: []
	};
	return {
		...n,
		message: `${t}: ${n.message}`
	};
}
function um(e) {
	return (t, n) => {
		e(null), (async () => {
			try {
				await n();
			} catch (n) {
				e(lm(n, `Couldn't ${t} comment`));
			}
		})();
	};
}
function dm(e, t) {
	let { comment: n, config: r, onAcknowledge: i, onAssign: a, extraAssignItems: o } = e;
	return sm({
		comment: n,
		config: r,
		run: t,
		...i ? { onAcknowledge: i } : {},
		...a ? { onAssign: a } : {},
		...o ? { extraAssignItems: o } : {}
	});
}
function fm(e, t, n) {
	let { onEdit: r, onDelete: i, onCreateTodo: a } = e;
	return {
		...r ? { onEdit: n } : {},
		...i ? { onDelete: () => t("delete", i) } : {},
		...a ? { onCreateTodo: () => t("create todo from", a) } : {}
	};
}
function pm(e) {
	let { comment: t, config: n, renderBody: r, onUpdateStatus: i, onClose: a, onEdit: o, editing: s, onEditingChange: l, onCopy: u, onReply: d, onChecklistToggle: f, onMaximize: m } = e, [h, g] = (0, N.useState)(null), _ = !!t.parentId, v = ae(t.createdAt), y = wi(n, "resolved"), b = wi(n, "active"), x = wi(n, "closed"), S = Ci(n, t.status), C = async (e) => {
		g(null);
		try {
			await i?.(e);
		} catch (e) {
			g(lm(e, "Couldn't update comment"));
		}
	}, E = async () => {
		g(null);
		try {
			await a?.();
		} catch (e) {
			g(lm(e, "Couldn't close comment"));
		}
	}, ee = async () => {
		g(null);
		try {
			await u?.();
		} catch (e) {
			g(lm(e, "Couldn't copy comment"));
		}
	}, ne = cm(_, {
		...fm(e, um(g), () => l(!0)),
		statusActions: [
			...dm(e, um(g)),
			...S === "active" && i && y ? [{
				label: "Resolve",
				icon: w,
				onSelect: () => void C(y.value)
			}] : [],
			...(S === "active" || S === "resolved") && a && x ? [{
				label: "Close",
				icon: T,
				onSelect: () => void E()
			}] : [],
			...(S === "resolved" || S === "closed") && i && b ? [{
				label: "Reopen",
				icon: fe,
				onSelect: () => void C(b.value)
			}] : []
		],
		...u ? { onCopy: () => void ee() } : {},
		...m ? { onMaximize: m } : {}
	});
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-2.5",
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex items-start gap-2.5",
				children: [
					/* @__PURE__ */ (0, P.jsx)(Mp, { author: t.author }),
					/* @__PURE__ */ (0, P.jsx)("div", {
						className: "min-w-0 flex-1",
						children: /* @__PURE__ */ (0, P.jsxs)("div", {
							className: "flex min-w-0 flex-wrap items-center gap-x-1.5 gap-y-1",
							children: [
								/* @__PURE__ */ (0, P.jsx)("span", {
									className: "min-w-0 truncate text-xs font-semibold text-foreground",
									children: /* @__PURE__ */ (0, P.jsx)(jp, { author: t.author })
								}),
								v && /* @__PURE__ */ (0, P.jsx)("span", {
									title: v.toLocaleString(),
									className: "shrink-0 text-[10px] text-muted-foreground",
									children: k(v)
								}),
								!_ && /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
									/* @__PURE__ */ (0, P.jsx)(Bp, {
										comment: t,
										config: n
									}),
									/* @__PURE__ */ (0, P.jsx)(Ip, { rating: t.rating }),
									/* @__PURE__ */ (0, P.jsx)(Fp, { decision: t.reviewDecision }),
									/* @__PURE__ */ (0, P.jsx)(Vp, {
										comment: t,
										config: n
									})
								] })
							]
						})
					}),
					/* @__PURE__ */ (0, P.jsxs)("div", {
						className: "flex shrink-0 items-center gap-0.5",
						children: [
							!_ && (S === "active" || S === "resolved") && a && x && /* @__PURE__ */ (0, P.jsx)("button", {
								type: "button",
								"aria-label": "Close comment",
								title: "Close comment",
								onClick: () => void E(),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, P.jsx)(c, {
									icon: w,
									className: "text-sm"
								})
							}),
							!_ && S === "closed" && i && b && /* @__PURE__ */ (0, P.jsx)("button", {
								type: "button",
								"aria-label": "Reopen comment",
								title: "Reopen comment",
								onClick: () => void C(b.value),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, P.jsx)(c, {
									icon: fe,
									className: "text-sm"
								})
							}),
							ne.length > 0 && /* @__PURE__ */ (0, P.jsx)(te, {
								align: "right",
								menuClassName: "min-w-[150px]",
								trigger: /* @__PURE__ */ (0, P.jsx)("span", {
									role: "button",
									tabIndex: 0,
									"aria-label": "Comment actions",
									className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11",
									children: /* @__PURE__ */ (0, P.jsx)(c, {
										icon: p,
										className: "text-sm"
									})
								}),
								items: ne
							})
						]
					})
				]
			}),
			s && o && /* @__PURE__ */ (0, P.jsx)(Pp, {
				comment: t,
				config: n,
				onSave: o,
				onDone: () => l(!1)
			}),
			!s && t.body.trim() && /* @__PURE__ */ (0, P.jsx)("div", {
				className: "text-[13px] leading-5 text-foreground",
				children: e.previewLines === void 0 ? /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: r ? r(t.body) : /* @__PURE__ */ (0, P.jsx)(Xp, {
					...t.vocabularyRefs ? { vocabularyRefs: t.vocabularyRefs } : {},
					text: t.body,
					...n.vocabulary ? { vocabulary: n.vocabulary } : {}
				}) }) : /* @__PURE__ */ (0, P.jsx)(Zp, {
					id: t.id,
					lines: e.previewLines,
					children: r ? r(t.body) : /* @__PURE__ */ (0, P.jsx)(Xp, {
						text: t.body,
						...t.vocabularyRefs ? { vocabularyRefs: t.vocabularyRefs } : {},
						...n.vocabulary ? { vocabulary: n.vocabulary } : {}
					})
				})
			}),
			h !== null && /* @__PURE__ */ (0, P.jsx)(F, {
				error: h,
				context: [["Comment", t.id], ["Anchor", t.anchor ?? "Whole page"]]
			}),
			!_ && /* @__PURE__ */ (0, P.jsx)(Hp, {
				comment: t,
				...f ? { onChecklistToggle: f } : {}
			}),
			!_ && /* @__PURE__ */ (0, P.jsx)(Up, { comment: t }),
			d && /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: d,
				className: "-ml-2 rounded px-2 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary/10 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
				children: "Reply"
			})
		]
	});
}
function mm(e) {
	return e.comment.submission ? /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-2 rounded-lg border border-border bg-card p-3",
		children: [
			e.comment.submission.state === "failed" && /* @__PURE__ */ (0, P.jsx)("p", {
				role: "status",
				children: "Comment could not be sent"
			}),
			/* @__PURE__ */ (0, P.jsx)("p", {
				className: "whitespace-pre-wrap",
				children: e.comment.body
			}),
			e.comment.submission.error && /* @__PURE__ */ (0, P.jsx)(F, { error: Error(e.comment.submission.error) })
		]
	}) : /* @__PURE__ */ (0, P.jsx)(hm, { ...e });
}
function hm(e) {
	let { comment: t, config: n, compact: r, defaultExpanded: i } = e, [o, s] = yt(`expanded:${t.id}`, i ?? !1), [l, u] = (0, N.useState)(null), [d, f] = (0, N.useState)(!1), m = (0, N.useContext)(rm), h = !!t.parentId, g = /* @__PURE__ */ (0, P.jsx)(pm, {
		...e,
		editing: d,
		onEditingChange: f
	}), _ = !o && !e.fullBody && e.previewLines === void 0 && !d, v = t.proposal ? /* @__PURE__ */ (0, P.jsx)(nm, {
		commentId: t.id,
		proposal: t.proposal,
		callbacks: m,
		...h && _ ? { className: "ml-4" } : {}
	}) : null;
	async function y(t) {
		u(null);
		try {
			if (t === "close") {
				await e.onClose?.();
				return;
			}
			let r = wi(n, t === "resolve" ? "resolved" : "active");
			r && await e.onUpdateStatus?.(r.value);
		} catch (e) {
			u(lm(e, t === "close" ? "Couldn't close comment" : "Couldn't update comment"));
		}
	}
	async function b() {
		u(null);
		try {
			await e.onCopy?.();
		} catch (e) {
			u(lm(e, "Couldn't copy comment"));
		}
	}
	return _ ? /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
		/* @__PURE__ */ (0, P.jsxs)("div", {
			role: "button",
			tabIndex: 0,
			onClick: () => s(!0),
			onKeyDown: (e) => {
				(e.key === "Enter" || e.key === " ") && (e.preventDefault(), s(!0));
			},
			"data-testid": "comment-card",
			"data-comment-kind": h ? "reply" : "root",
			"data-comment-id": t.id,
			className: a("group/comment flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs transition-colors hover:bg-muted/50", !h && "border border-border bg-background shadow-sm", h && "ml-4 border-l-2 border-l-border bg-muted/30"),
			children: [
				/* @__PURE__ */ (0, P.jsx)(Mp, {
					author: t.author,
					bare: !0
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 truncate text-[11px] font-semibold text-foreground",
					children: /* @__PURE__ */ (0, P.jsx)(jp, { author: t.author })
				}),
				!h && /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
					/* @__PURE__ */ (0, P.jsx)(Bp, {
						comment: t,
						config: n
					}),
					/* @__PURE__ */ (0, P.jsx)(Ip, { rating: t.rating }),
					/* @__PURE__ */ (0, P.jsx)(Fp, { decision: t.reviewDecision }),
					/* @__PURE__ */ (0, P.jsx)(Vp, {
						comment: t,
						config: n,
						compact: !0
					})
				] }),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 flex-1 truncate text-[12px] text-foreground/80",
					children: t.body.trim() ? t.vocabularyRefs?.length ? /* @__PURE__ */ (0, P.jsx)(Xp, {
						text: t.body,
						vocabularyRefs: t.vocabularyRefs,
						...n.vocabulary ? { vocabulary: n.vocabulary } : {}
					}) : bi(t.body) : t.rating === "positive" ? "Positive rating" : t.rating === "negative" ? "Negative rating" : t.reviewDecision === "approved" ? "Approved" : "Empty comment"
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "shrink-0",
					role: "presentation",
					onClick: (e) => e.stopPropagation(),
					onKeyDown: (e) => e.stopPropagation(),
					children: /* @__PURE__ */ (0, P.jsx)(te, {
						align: "right",
						trigger: /* @__PURE__ */ (0, P.jsx)("span", {
							role: "button",
							tabIndex: 0,
							"aria-label": "Comment actions",
							className: "rounded p-0.5 text-muted-foreground opacity-0 transition-opacity hover:bg-muted hover:opacity-100 group-hover/comment:opacity-70 [@media(hover:none)]:opacity-70 [@media(pointer:coarse)]:p-2.5",
							children: /* @__PURE__ */ (0, P.jsx)(c, {
								icon: p,
								className: "text-sm"
							})
						}),
						items: cm(h, {
							...fm(e, um(u), () => f(!0)),
							statusActions: [
								...dm(e, um(u)),
								...Ci(n, t.status) === "active" && e.onUpdateStatus && wi(n, "resolved") ? [{
									label: "Resolve",
									icon: w,
									onSelect: () => void y("resolve")
								}] : [],
								...(Ci(n, t.status) === "active" || Ci(n, t.status) === "resolved") && e.onClose && wi(n, "closed") ? [{
									label: "Close",
									icon: T,
									onSelect: () => void y("close")
								}] : [],
								...(Ci(n, t.status) === "resolved" || Ci(n, t.status) === "closed") && e.onUpdateStatus && wi(n, "active") ? [{
									label: "Reopen",
									icon: fe,
									onSelect: () => void y("reopen")
								}] : []
							],
							...e.onCopy ? { onCopy: () => void b() } : {},
							...e.onMaximize ? { onMaximize: e.onMaximize } : {}
						})
					})
				})
			]
		}),
		l !== null && /* @__PURE__ */ (0, P.jsx)(F, {
			error: l,
			context: [["Comment", t.id], ["Anchor", t.anchor ?? "Whole page"]]
		}),
		v
	] }) : /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-testid": "comment-card",
		"data-comment-kind": h ? "reply" : "root",
		"data-comment-id": t.id,
		className: a("group/comment rounded-xl bg-background px-3.5 py-3 text-xs", !h && "border border-border/80 shadow-sm transition-shadow hover:shadow-md", r && "px-2.5 py-2.5", h && "ml-4 border-l-2 border-l-border bg-muted/20 shadow-none"),
		children: [g, v && /* @__PURE__ */ (0, P.jsx)("div", {
			className: "mt-2.5",
			children: v
		})]
	}) });
}
// ../../../clicky-ui/packages/ui/src/comments/CommentThreadList.tsx
function gm({ comments: e, config: t, compact: n = !1, defaultExpanded: r = !1, previewLines: i, detailTabs: o = [], renderMaximizedActions: s, onMaximizedChange: c, renderRootMeta: l, renderRootDetails: u, getThreadProps: d, renderBody: f, threadToMarkdown: p, onUpdateStatus: m, onClose: h, onAcknowledge: g, onAssign: _, extraAssignItems: v, onChecklistToggle: y, onDelete: b, onEdit: x, onEditReply: S, onDeleteReply: C, onCreateTodo: w, onReply: T }) {
	let [E, ee] = yt(`thread-list:${e[0]?.anchor ?? "general"}:replying`, null), [te, ne] = yt(`thread-list:${e[0]?.anchor ?? "general"}:maximized`, null), [O, re] = yt(`thread-list:${e[0]?.anchor ?? "general"}:tab`, "thread"), k = (0, N.useRef)(null), ie = pi(e), ae = mi(e), oe = ie.find((e) => e.id === te), se = !!oe;
	if ((0, N.useEffect)(() => {
		if (se) return c?.(!0), () => c?.(!1);
	}, [se, c]), ie.length === 0) return null;
	let A = oe ? [oe, ...gi(ae.get(oe.id) ?? [])] : [];
	function ce() {
		ne(null), re("thread");
	}
	function le(e, n) {
		let r = e.id, i = n[0], a = i.id, o = r !== a, s = o ? C : void 0, c = o ? S && ((e) => S(a, r, e)) : x && ((e) => x(r, e));
		return {
			...c && t.canEditThread?.(i) !== !1 ? { onEdit: c } : {},
			...!o && w ? { onCreateTodo: () => w(n) } : {},
			...m ? { onUpdateStatus: (e) => m(r, e) } : {},
			...h ? { onClose: () => h(r) } : {},
			...g ? { onAcknowledge: () => g(r) } : {},
			..._ ? { onAssign: (e) => _(r, e) } : {},
			...v ? { extraAssignItems: v } : {},
			...y ? { onChecklistToggle: (e) => y(r, e) } : {},
			...s ? { onDelete: () => s(a, r) } : b ? { onDelete: () => b(r) } : {}
		};
	}
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)("div", {
		className: "space-y-2",
		children: ie.map((e) => {
			let o = gi(ae.get(e.id) ?? []), s = [e, ...o], c = d?.(e) ?? {};
			return /* @__PURE__ */ (0, P.jsxs)("div", {
				"data-testid": "comment-thread-block",
				"data-comment-id": e.id,
				...c,
				className: a("space-y-1.5 p-1", c.className),
				children: [
					l ? /* @__PURE__ */ (0, P.jsx)("div", {
						className: "px-1",
						children: l(e)
					}) : null,
					/* @__PURE__ */ (0, P.jsx)(mm, {
						comment: e,
						config: t,
						compact: n,
						defaultExpanded: r,
						...i === void 0 ? {} : { previewLines: i },
						onMaximize: () => ne(e.id),
						...p ? { onCopy: () => pe(p(s)) } : {},
						...f ? { renderBody: f } : {},
						...le(e, s),
						...T ? { onReply: () => ee(e.id) } : {}
					}),
					u?.(e),
					o.length > 0 && /* @__PURE__ */ (0, P.jsx)("div", {
						className: "space-y-1.5",
						children: o.map((a) => /* @__PURE__ */ (0, P.jsx)(mm, {
							comment: a,
							config: t,
							compact: n,
							defaultExpanded: r,
							...i === void 0 ? {} : { previewLines: i },
							onMaximize: () => ne(e.id),
							...p ? { onCopy: () => pe(p(s)) } : {},
							...f ? { renderBody: f } : {},
							...le(a, s)
						}, a.id))
					}),
					T && E === e.id && /* @__PURE__ */ (0, P.jsx)(Ep, {
						parentId: e.id,
						config: t,
						onCancel: () => ee(null),
						onSubmit: async (t) => {
							await T(e, t), ee(null);
						}
					})
				]
			}, e.id);
		})
	}), oe && /* @__PURE__ */ (0, P.jsx)(D, {
		open: !0,
		onClose: ce,
		title: "Comment thread",
		size: "xl",
		className: "h-[min(800px,90dvh)]",
		headerSlot: s?.(A),
		subtitle: /* @__PURE__ */ (0, P.jsx)(qe, {
			tabs: [{
				id: "thread",
				label: "Thread"
			}, ...o.map(({ id: e, label: t }) => ({
				id: e,
				label: t
			}))],
			value: O,
			onChange: re
		}),
		footer: T && /* @__PURE__ */ (0, P.jsx)("div", {
			ref: k,
			children: /* @__PURE__ */ (0, P.jsx)(Ep, {
				pinned: !0,
				parentId: oe.id,
				config: t,
				onCancel: () => {},
				onSubmit: (e) => T(oe, e)
			}, oe.id)
		}),
		children: /* @__PURE__ */ (0, P.jsx)("div", {
			className: "space-y-3",
			"data-testid": "comment-thread-modal",
			children: o.some((e) => e.id === O) ? o.find((e) => e.id === O)?.render(A) : /* @__PURE__ */ (0, P.jsx)("div", {
				className: "space-y-2",
				children: A.map((e) => /* @__PURE__ */ (0, P.jsxs)("div", {
					className: "space-y-1.5",
					children: [/* @__PURE__ */ (0, P.jsx)(mm, {
						comment: e,
						config: t,
						defaultExpanded: !0,
						fullBody: !0,
						...p ? { onCopy: () => pe(p(A)) } : {},
						...f ? { renderBody: f } : {},
						...le(e, A),
						...T ? { onReply: () => k.current?.querySelector("[role=\"textbox\"]")?.focus() } : {}
					}), e.id === oe.id && u?.(oe)]
				}, e.id))
			})
		})
	})] });
}
// ../../../clicky-ui/packages/ui/src/comments/CommentThread.tsx
function _m({ comments: e, config: t, anchor: n, compact: r = !1, autoFocusComposer: i = !1, defaultExpanded: o = !1, previewLines: s, detailTabs: c, renderMaximizedActions: l, onMaximizedChange: u, composerPlaceholder: d, hideComposer: f = !1, createActions: p, renderRootMeta: m, renderRootDetails: h, renderBody: g, threadToMarkdown: _, extraAssignItems: v, onCreate: y, onReply: b, onUpdateStatus: x, onClose: S, onAcknowledge: C, onAssign: w, onDelete: T, onEdit: E, onEditReply: ee, onDeleteReply: te, onCreateTodo: ne, onChecklistToggle: D, onMention: O, onApplyProposal: re, onRejectProposal: k, onRequestRevision: ie, loadProposalDiff: ae }) {
	let oe = im({
		...re ? { onApplyProposal: re } : {},
		...k ? { onRejectProposal: k } : {},
		...ie ? { onRequestRevision: ie } : {},
		...ae ? { loadProposalDiff: ae } : {}
	}), se = Si(e, t, {
		...x ? { onUpdateStatus: x } : {},
		...S ? { onClose: S } : {},
		...C ? { onAcknowledge: C } : {},
		...w ? { onAssign: w } : {},
		...T ? { onDelete: T } : {},
		...E ? { onEdit: E } : {},
		...ee ? { onEditReply: ee } : {},
		...te ? { onDeleteReply: te } : {},
		...ne ? { onCreateTodo: ne } : {},
		...D ? { onChecklistToggle: D } : {},
		...b ? { onReply: b } : {},
		...O ? { onMention: O } : {}
	}), A = /* @__PURE__ */ (0, P.jsxs)("div", {
		className: a(r && "p-2"),
		children: [/* @__PURE__ */ (0, P.jsx)(gm, {
			comments: e,
			config: t,
			compact: r,
			defaultExpanded: o,
			...s === void 0 ? {} : { previewLines: s },
			...c ? { detailTabs: c } : {},
			...l ? { renderMaximizedActions: l } : {},
			...u ? { onMaximizedChange: u } : {},
			...m ? { renderRootMeta: m } : {},
			...h ? { renderRootDetails: h } : {},
			...g ? { renderBody: g } : {},
			..._ ? { threadToMarkdown: _ } : {},
			...v ? { extraAssignItems: v } : {},
			...se
		}), !f && y && /* @__PURE__ */ (0, P.jsx)("div", {
			className: "pt-1",
			children: /* @__PURE__ */ (0, P.jsx)(Tp, {
				config: t,
				anchor: n ?? null,
				autoFocus: i,
				...d ? { placeholder: d } : {},
				onCreate: y,
				...p ? { createActions: p } : {},
				...O ? { onMention: O } : {}
			})
		})]
	});
	return oe ? /* @__PURE__ */ (0, P.jsx)(rm.Provider, {
		value: oe,
		children: A
	}) : A;
}
// ../../../clicky-ui/packages/ui/src/comments/use-comment-rail-width.ts
var vm = "clicky.comment-sidebar.width", ym = "clicky.comment-sidebar.resize", bm = (e) => (window.addEventListener(ym, e), window.addEventListener("storage", e), () => {
	window.removeEventListener(ym, e), window.removeEventListener("storage", e);
});
function xm(e = 320, t = vm) {
	return [(0, N.useSyncExternalStore)(bm, () => {
		let n = localStorage.getItem(t);
		if (n === null) return e;
		let r = Number(n);
		if (!Number.isFinite(r) || r < 280 || r > 640) throw Error(`Invalid comment sidebar width: ${n}`);
		return r;
	}, () => e), (e) => {
		if (!Number.isFinite(e) || e < 280 || e > 640) throw Error(`Invalid comment sidebar width: ${e}`);
		localStorage.setItem(t, String(e)), window.dispatchEvent(new Event(ym));
	}];
}
// ../../../clicky-ui/packages/ui/src/comments/comment-context.ts
var Sm = (0, N.createContext)(null), Cm = Object.freeze({
	meta: void 0,
	active: !1,
	highlighted: !1
});
function wm(e, t) {
	return e === t ? !0 : !e || !t ? !1 : e.count === t.count && e.latestStatus === t.latestStatus && e.authors.length === t.authors.length && e.authors.every((e, n) => e === t.authors[n]);
}
function Tm(e, t) {
	return wm(e.meta, t.meta) && e.active === t.active && e.highlighted === t.highlighted;
}
function Em() {
	let e = {}, t = "closed", n = null, r = null, i = /* @__PURE__ */ new Map(), a = /* @__PURE__ */ new Map();
	function o(i) {
		return {
			meta: e[i],
			active: t === "focused" && n === i,
			highlighted: r === i || t === "focused" && n === i
		};
	}
	return {
		update(s, c, l, u) {
			e = s, t = c, n = l, r = u;
			for (let [e, t] of a) {
				let n = i.get(e) ?? Cm, r = o(e);
				if (!Tm(n, r)) {
					i.set(e, r);
					for (let e of t) e();
				}
			}
		},
		getSnapshot(e) {
			let t = i.get(e) ?? Cm, n = o(e);
			return Tm(t, n) ? t : (i.set(e, n), n);
		},
		subscribe(e, t) {
			let n = a.get(e) ?? /* @__PURE__ */ new Set();
			return n.add(t), a.set(e, n), () => {
				n.delete(t), n.size === 0 && a.delete(e);
			};
		}
	};
}
var Dm = (0, N.createContext)(null), Om = (0, N.createContext)(null);
function km(e, t) {
	return t.includes(e) ? e : t.includes("__document__") ? oi : null;
}
function Am() {
	let e = (0, N.useContext)(Sm);
	if (!e) throw Error("useCommentContext must be used within a CommentProvider");
	return e;
}
function jm() {
	return (0, N.useContext)(Sm);
}
// ../../../clicky-ui/packages/ui/src/comments/CommentRail.tsx
function Mm({ children: e, className: t, style: n, width: r, onWidthChange: i, stacked: s, scrollRef: c, open: l = !0, onOpenChange: u, launcherLabel: d = "Open comments" }) {
	let f = jm()?.railHeaderActions, p = (0, N.useRef)(null), m = (0, N.useRef)(null), h = (0, N.useRef)(l);
	(0, N.useEffect)(() => {
		h.current && !l && m.current?.focus(), h.current = l;
	}, [l]);
	let [g, _] = xm(), v = b(), [y, x] = (0, N.useState)(960);
	(0, N.useLayoutEffect)(() => {
		let e = p.current?.parentElement;
		if (!e) return;
		let t = () => x(e.clientWidth);
		t();
		let n = typeof ResizeObserver > "u" ? null : new ResizeObserver(t);
		return n?.observe(e), () => n?.disconnect();
	}, []);
	let S = s ?? y < 640, C = Math.max(280, Math.min(640, y - 320)), w = Math.max(280, Math.min(C, r ?? g));
	if (!l && !u) throw Error("A closed comment rail requires onOpenChange to restore it");
	return !l || S && u ? /* @__PURE__ */ (0, P.jsx)("div", {
		ref: p,
		className: "contents",
		children: l ? /* @__PURE__ */ (0, P.jsx)(D, {
			open: !0,
			title: "Comments",
			headerSlot: f,
			expandable: !1,
			onClose: () => u?.(!1),
			children: e
		}) : /* @__PURE__ */ (0, P.jsxs)(o, {
			ref: m,
			type: "button",
			"aria-label": d,
			title: d,
			onClick: () => u?.(!0),
			style: { zIndex: v },
			className: "fixed bottom-[calc(1rem+env(safe-area-inset-bottom))] right-4 h-12 gap-2 rounded-full px-4 shadow-lg",
			children: [/* @__PURE__ */ (0, P.jsx)(we, { className: "size-5" }), d]
		})
	}) : /* @__PURE__ */ (0, P.jsxs)("div", {
		ref: p,
		"data-comment-rail": "",
		className: a("relative flex min-h-0 shrink-0 flex-col border-border bg-background", S ? "w-full border-t" : "border-l", t),
		style: {
			...n,
			...S ? {} : { width: w }
		},
		children: [
			!S && /* @__PURE__ */ (0, P.jsx)(ht, {
				orientation: "vertical",
				value: w,
				min: 280,
				max: C,
				pointerFactor: -1,
				label: "Resize comment sidebar",
				className: "left-0",
				onChange: i ?? _
			}),
			f && /* @__PURE__ */ (0, P.jsxs)("header", {
				className: "flex shrink-0 items-center justify-between gap-2 border-b border-border px-3 py-1.5",
				children: [/* @__PURE__ */ (0, P.jsx)("span", {
					className: "text-xs font-semibold",
					children: "Comments"
				}), f]
			}),
			/* @__PURE__ */ (0, P.jsx)("div", {
				ref: c,
				className: "min-h-0 flex-1 overflow-y-auto overscroll-contain",
				"data-comment-rail-scroll": "",
				children: e
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/findings-comments.ts
function Nm(e) {
	return `line:${e}`;
}
function Pm(e, t) {
	return `L${e}-${t}`;
}
var Fm = /^line:(\d+)$/, Im = /^L(\d+)-(\d+)$/;
function Lm(e) {
	let t = Fm.exec(e);
	if (!t) throw Error(`file-findings: expected a "line:<n>" anchor, got ${JSON.stringify(e)}`);
	return Number(t[1]);
}
var Rm = (e, t) => {
	if (e === "__document__" || Im.test(e)) return t.includes(e) ? e : null;
	let n = Lm(e), r = t.flatMap((e) => {
		let t = Im.exec(e);
		return t ? [{
			anchor: e,
			start: Number(t[1]),
			end: Number(t[2])
		}] : [];
	}), i = r.filter((e) => e.start <= n && n <= e.end).sort((e, t) => e.end - e.start - (t.end - t.start) || t.start - e.start)[0];
	if (i) return i.anchor;
	let a = r.filter((e) => e.start <= n).sort((e, t) => t.start - e.start)[0];
	return a ? a.anchor : t.includes("__document__") ? oi : null;
}, zm = {
	statuses: [
		{
			value: "error",
			label: "Error",
			tone: "danger",
			stage: "active",
			unresolved: !0
		},
		{
			value: "warning",
			label: "Warning",
			tone: "warning",
			stage: "active",
			unresolved: !0
		},
		{
			value: "info",
			label: "Info",
			tone: "info",
			stage: "active",
			unresolved: !0
		}
	],
	facets: [{
		key: "severity",
		label: "Severity",
		options: [
			{
				value: "error",
				label: "Error",
				tone: "danger"
			},
			{
				value: "warning",
				label: "Warning",
				tone: "warning"
			},
			{
				value: "info",
				label: "Info",
				tone: "info"
			}
		]
	}, {
		key: "kind",
		label: "Kind",
		options: [{
			value: "lint",
			label: "Lint",
			tone: "neutral"
		}, {
			value: "test",
			label: "Test",
			tone: "neutral"
		}]
	}]
}, Bm = "1970-01-01T00:00:00.000Z", Vm = 6;
function Hm(e) {
	return e.map((e, t) => ({
		id: `finding-${String(t).padStart(Vm, "0")}`,
		body: e.rule ? `${e.message}\n\n\`${e.rule}\`` : e.message,
		createdAt: Bm,
		author: {
			name: e.source,
			kind: "agent"
		},
		status: e.severity,
		anchor: Nm(e.line),
		facets: {
			severity: e.severity,
			source: e.source,
			kind: e.kind
		}
	}));
}
// ../../../clicky-ui/packages/ui/src/file-findings/LazyAnnotatedCode.tsx
var Um = (0, N.lazy)(() => import("./AnnotatedCode-CxgXcnDV.js").then((e) => ({ default: e.AnnotatedCode })));
function Wm(e) {
	return /* @__PURE__ */ (0, P.jsx)(N.Suspense, {
		fallback: /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
			children: [/* @__PURE__ */ (0, P.jsx)(c, {
				icon: Te,
				className: "mr-2 animate-spin"
			}), "Loading editor"]
		}),
		children: /* @__PURE__ */ (0, P.jsx)(Um, { ...e })
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/mdx-source.ts
var Gm = /^\s{0,3}(`{3,}|~{3,})/, Km = /^(import|export)\s/;
function qm(e) {
	let t = 0;
	for (let n of e) "{[(".includes(n) ? t += 1 : "}])".includes(n) && --t;
	return t;
}
function Jm(e) {
	let t = e.split("\n"), n = t.slice(), r = 0;
	if (t[0] === "---") {
		let e = t.indexOf("---", 1);
		if (e !== -1) for (; r <= e; r += 1) n[r] = "";
	}
	let i = null;
	for (; r < t.length; r += 1) {
		let e = t[r], a = Gm.exec(e);
		if (i) {
			a && a[1][0] === i[0] && a[1].length >= i.length && (i = null);
			continue;
		}
		if (a) {
			i = a[1];
			continue;
		}
		if (!Km.test(e)) continue;
		let o = qm(e);
		for (n[r] = ""; o > 0 && r + 1 < t.length;) r += 1, o += qm(t[r]), n[r] = "";
	}
	return n.join("\n");
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/table-cell.ts
var Ym = "[data-source-start]";
function Xm(e) {
	let t = e.querySelectorAll("table");
	if (t.length !== 1) return [];
	let n = [...t[0].rows], { sourceStart: r, sourceEnd: i } = e.dataset, a = Number(r);
	return n.length === 0 || t[0].tHead?.rows.length !== 1 || n.length !== Number(i) - a ? [] : n.map((e, t) => ({
		row: e,
		index: t - 1,
		line: t === 0 ? a : a + 1 + t
	}));
}
var Zm = (e) => (e?.textContent ?? "").replace(/\s+/g, " ").trim();
function Qm(e, t) {
	let n = null;
	for (let r = e instanceof Element ? e : e.parentElement; r && r !== t; r = r.parentElement) r.matches(Ym) && (n = r);
	return n;
}
var $m = (e) => (e instanceof Element ? e : e.parentElement)?.closest("td, th") ?? null, eh = (e) => [...e[0].row.cells].map((e) => Zm(e)), th = (e, t) => e.map((e) => Zm(e.row.cells[t]));
function nh(e) {
	let t = e.slice(1);
	for (let n = 0; n < e[0].row.cells.length; n++) {
		let e = th(t, n);
		if (e.every(Boolean) && new Set(e).size === e.length) return n;
	}
}
function rh(e, t) {
	let n = $m(e.startContainer);
	if (!n || n !== $m(e.endContainer) || !t.contains(n)) return null;
	let r = Qm(n, t), i = r ? Xm(r) : [], a = i.find(({ row: e }) => e === n.parentElement);
	if (!a) return null;
	let o = a.index < 0 ? void 0 : nh(i), s = eh(i);
	return {
		cell: {
			keyColumn: o === void 0 ? "" : s[o],
			key: o === void 0 ? "" : Zm(a.row.cells[o]),
			column: s[n.cellIndex],
			row: a.index,
			col: n.cellIndex
		},
		line: a.line
	};
}
function ih(e, t) {
	if (!t) return;
	let n = e.flatMap((e, n) => e === t ? [n] : []);
	return n.length === 1 ? n[0] : void 0;
}
function ah(e, t) {
	let n = Xm(e);
	if (n.length === 0) return null;
	let [r, ...i] = n, a = eh(n), o = ih(a, t.keyColumn), s = o === void 0 ? void 0 : ih(th(i, o), t.key), c = t.row < 0 ? r : i[s ?? t.row], l = c?.row.cells[ih(a, t.column) ?? t.col];
	return c && l ? {
		cell: l,
		line: c.line
	} : null;
}
function oh(e) {
	return e.row < 0 && e.column ? `header[${e.column}]` : e.key && e.column ? `row[${e.key}][${e.column}]` : `cell[${e.row}][${e.col}]`;
}
var sh = "[data-source-start]", ch = /^h([1-6])$/, lh = 2;
function uh(e) {
	return `q:${JSON.stringify([
		e.path,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function dh(e) {
	if (!e.startsWith("q:")) throw Error(`not a quote anchor: ${JSON.stringify(e)}`);
	let t = JSON.parse(e.slice(2)), [n, r, i, a] = Array.isArray(t) ? t : [];
	if (!Array.isArray(n) || !n.every((e) => typeof e == "string") || typeof r != "string" || typeof i != "string" || typeof a != "string") throw Error(`malformed quote anchor ${JSON.stringify(e)}`);
	return {
		path: n,
		prefix: r,
		quote: i,
		suffix: a
	};
}
function fh(e) {
	let t = [];
	return [...e.querySelectorAll(sh)].filter((e) => !e.parentElement?.closest(sh)).map((e) => {
		let n = ch.exec(e.tagName.toLowerCase())?.[1];
		if (n) {
			for (; t.length > 0 && (t.at(-1)?.level ?? 0) >= Number(n);) t.pop();
			let r = hh(e).replace(/\s+/g, " ").trim().slice(0, 200);
			t.push({
				level: Number(n),
				text: r
			});
		}
		return {
			block: e,
			path: t.map((e) => e.text),
			line: Number(e.dataset.sourceStart)
		};
	});
}
var ph = (e) => fh(e);
function mh(e) {
	let t = document.createTreeWalker(e, NodeFilter.SHOW_TEXT), n = [];
	for (let e = t.nextNode(); e; e = t.nextNode()) e.parentElement?.closest("[data-comment-author]") || n.push(e);
	return n;
}
function hh(e) {
	return mh(e).map((e) => e.data).join("");
}
function gh(e, t, n) {
	let r = document.createRange(), i = 0, a = !1;
	for (let o of mh(e)) {
		if (!a && t < i + o.length && (r.setStart(o, t - i), a = !0), a && n <= i + o.length) return r.setEnd(o, n - i), r;
		i += o.length;
	}
	throw Error(`quote-anchor: offsets ${t}-${n} lie outside the block`);
}
function _h(e) {
	return e.trim().split(/\s+/).filter(Boolean).map((e) => e.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("\\s*");
}
function vh(e, t, n) {
	let r = n.prefix && t.prefix.trim() ? `(?<=${_h(t.prefix)}\\s*)` : "", i = n.suffix && t.suffix.trim() ? `(?=\\s*${_h(t.suffix)})` : "", a = RegExp(`${r}${_h(t.quote)}${i}`, "g");
	return e.flatMap(({ block: e }) => [...hh(e).matchAll(a)].map((t) => ({
		block: e,
		start: t.index,
		end: t.index + t[0].length
	})));
}
var yh = (e, t) => e.length === t.length && e.every((e, n) => e === t[n]), bh = [
	{
		prefix: !0,
		suffix: !0
	},
	{
		prefix: !0,
		suffix: !1
	},
	{
		prefix: !1,
		suffix: !0
	}
];
function xh(e, t) {
	let n = ph(e), r = n.filter((e) => yh(e.path, t.path));
	for (let e of bh) {
		let n = vh(r, t, e);
		if (n.length === 1) return gh(n[0].block, n[0].start, n[0].end);
	}
	let i = vh(n, t, bh[0]);
	return i.length === 1 ? gh(i[0].block, i[0].start, i[0].end) : null;
}
function Sh(e, t) {
	let n = ph(e), r = n.find(({ block: e }) => e.contains(t.startContainer) && e.contains(t.endContainer));
	if (!r) return null;
	let i = hh(r.block), a = document.createRange();
	a.setStart(r.block, 0), a.setEnd(t.startContainer, t.startOffset);
	let o = Ph(t), s = Ph(a).length + (o.length - o.trimStart().length), c = Ph(a).length + o.trimEnd().length;
	if (c <= s) return null;
	let l = i.slice(0, s).split(/\s+/).filter(Boolean), u = i.slice(c).split(/\s+/).filter(Boolean), d = n.filter((e) => yh(e.path, r.path));
	for (let e = lh;; e++) {
		let t = {
			path: r.path,
			prefix: l.slice(-e).join(" "),
			quote: i.slice(s, c),
			suffix: u.slice(0, e).join(" ")
		}, n = vh(d, t, bh[0]);
		if (n.length === 1 && n[0].block === r.block && n[0].start === s) return t;
		if (e >= l.length && e >= u.length) return null;
	}
}
function Ch(e, t) {
	let n = Sh(e, t);
	return n && {
		prefix: n.prefix,
		exact: n.quote,
		suffix: n.suffix
	};
}
var wh = [...bh, {
	prefix: !1,
	suffix: !1
}];
function Th(e, t) {
	return ph(e).flatMap(({ block: e }) => {
		let { sourceStart: n, sourceEnd: r } = e.dataset;
		if (Number(n) > t.end || Number(r) < t.start) return [];
		let i = Xm(e);
		return i.length > 0 ? i.filter(({ line: e }) => e >= t.start && e <= t.end).map(({ row: e }) => e) : [e];
	});
}
function Eh(e, t) {
	let n = e.map((e) => ({
		block: e,
		path: []
	})), r = {
		path: [],
		prefix: t.prefix,
		quote: t.exact,
		suffix: t.suffix
	};
	for (let e of wh) {
		let t = vh(n, r, e);
		if (t.length === 1) return gh(t[0].block, t[0].start, t[0].end);
		if (t.length > 1) return null;
	}
	return null;
}
function Dh(e, t) {
	return e.exact !== "" && t.join("\n").includes(e.exact);
}
var Oh = /[*_`|>#~]|(?:^|\s)[-+](?=\s)/g;
function kh(e) {
	return _h(e.replace(Oh, " ")).replaceAll("\\s*", "[\\s*_`|>#~]*");
}
function Ah(e, t) {
	return ph(e).map(({ block: e }) => {
		let n = Number(e.dataset.sourceStart), r = Number(e.dataset.sourceEnd);
		return {
			block: e,
			first: n,
			last: r,
			text: t.slice(n - 1, r).join("\n")
		};
	});
}
var jh = (e, t) => [...e.matchAll(new RegExp(t, "g"))];
function Mh(e, t, n) {
	let r = t[n.line - 1]?.slice(n.start, n.end) ?? "";
	if (!r.replace(Oh, " ").trim()) return null;
	let i = Ah(e, t).find((e) => n.line >= e.first && n.line <= e.last);
	if (!i) return null;
	let a = t.slice(i.first - 1, n.line - 1).reduce((e, t) => e + t.length + 1, 0) + n.start, o = jh(i.text, kh(r)).findIndex((e) => e.index + e[0].length > a && e.index < a + r.length), s = o < 0 ? void 0 : jh(hh(i.block), _h(r.replace(Oh, " ")))[o];
	return s ? Sh(e, gh(i.block, s.index, s.index + s[0].length)) : null;
}
function Nh(e, t, n) {
	let r = xh(e, n), i = r && Ah(e, t).find(({ block: e }) => e.contains(r.startContainer));
	if (!r || !i) return [];
	let a = document.createRange();
	a.setStart(i.block, 0), a.setEnd(r.startContainer, r.startOffset);
	let o = Ph(a).length, s = jh(hh(i.block), _h(n.quote)).findIndex((e) => e.index === o), c = s < 0 ? void 0 : jh(i.text, kh(n.quote))[s];
	if (!c) return [];
	let l = [], u = 0;
	for (let e = i.first; e <= i.last; e++) {
		let n = t[e - 1] ?? "", r = Math.max(c.index, u) - u, i = Math.min(c.index + c[0].length, u + n.length) - u;
		i > r && l.push({
			line: e,
			start: r,
			end: i
		}), u += n.length + 1;
	}
	return l;
}
function Ph(e) {
	let t = e.cloneContents();
	for (let e of t.querySelectorAll("[data-comment-author]")) e.remove();
	return t.textContent ?? "";
}
function Fh(e) {
	let t = e.path.at(-1);
	return `${t ? `${t} › ` : ""}“${e.quote}”`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/geometry.ts
var Ih = "[data-source-start]";
function Lh(e) {
	return (e instanceof Element ? e : e?.parentElement ?? null)?.closest(Ih) ?? null;
}
function Rh(e, t, n) {
	for (let r of e.querySelectorAll(Ih)) {
		if (Number(r.dataset.sourceStart) > t.end || Number(r.dataset.sourceEnd) < t.start) continue;
		let e = ah(r, n);
		if (e && e.line >= t.start && e.line <= t.end) return e.cell;
	}
	return null;
}
function zh({ origin: e, markdown: t, anchors: n, mode: r, patchRanges: i }) {
	let a = e.getBoundingClientRect(), o = (e) => ({
		top: e.top - a.top,
		left: e.left - a.left,
		width: e.width,
		height: e.height
	}), s = /* @__PURE__ */ new Map();
	for (let e of t.querySelectorAll(Ih)) {
		let t = Pm(Number(e.dataset.sourceStart), Number(e.dataset.sourceEnd));
		s.has(t) || s.set(t, e);
	}
	let c = (e) => {
		let t = document.createRange();
		t.selectNodeContents(e);
		let n = Bh([...t.getClientRects()].filter((e) => e.width > 0).map(o));
		return n.length > 0 ? n : [o(e.getBoundingClientRect())];
	}, l = [];
	for (let e of n) {
		if (e.startsWith("q:")) {
			let n = xh(t, dh(e));
			if (!n) continue;
			let i = [...n.getClientRects()].filter((e) => e.width > 0).map(o), a = Lh(n.startContainer), s = r === "text" ? i : a ? [o(a.getBoundingClientRect())] : [];
			s.length > 0 && l.push({
				anchor: e,
				boxes: s,
				at: i[0]?.top ?? s[0].top
			});
			continue;
		}
		if (e.startsWith("e:")) {
			let n = Fr(t, e)?.element, r = n && o(n.getBoundingClientRect());
			r && l.push({
				anchor: e,
				boxes: [r],
				at: r.top
			});
			continue;
		}
		if (Zr(e)) {
			let n = i?.get(e);
			if (!n) continue;
			let a = r === "text" && n.cell ? Rh(t, n, n.cell) : null, s = a ? [a] : Th(t, n), u = r === "text" && n.quote ? Eh(s, n.quote) : null, d = u ? [...u.getClientRects()].filter((e) => e.width > 0).map(o) : [], f = d.length > 0 ? d : s.flatMap(c);
			f.length > 0 && l.push({
				anchor: e,
				boxes: f,
				at: f[0].top
			});
			continue;
		}
		let n = Rm(e, [...s.keys()]), a = n ? s.get(n) : void 0;
		if (a) {
			let t = c(a);
			l.push({
				anchor: e,
				boxes: t,
				at: t[0].top
			});
		}
	}
	return new Map(l.sort((e, t) => e.at - t.at).map(({ anchor: e, boxes: t }) => [e, t]));
}
function Bh(e) {
	let t = [];
	for (let n of [...e].sort((e, t) => e.top - t.top || e.left - t.left)) {
		let e = t.find((e) => n.top < e.top + e.height && n.top + n.height > e.top);
		if (!e) {
			t.push({ ...n });
			continue;
		}
		let r = Math.max(e.left + e.width, n.left + n.width), i = Math.max(e.top + e.height, n.top + n.height);
		e.left = Math.min(e.left, n.left), e.top = Math.min(e.top, n.top), e.width = r - e.left, e.height = i - e.top;
	}
	return t;
}
function Vh({ cards: e, focused: t, gap: n = 8 }) {
	let r = [...e].sort((e, t) => e.desired - t.desired || e.order - t.order), i = r.map((e) => e.desired), a = r.findIndex((e) => e.anchor === t);
	if (a >= 0) for (let e = a - 1; e >= 0; e--) i[e] = Math.min(i[e], i[e + 1] - n - r[e].height);
	for (let e = 0; e < r.length; e++) {
		let t = e === 0 ? 0 : i[e - 1] + r[e - 1].height + n;
		i[e] = Math.max(i[e], t);
	}
	return new Map(r.map((e, t) => [e.anchor, i[t]]));
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/hooks.ts
var Hh = 640;
function Uh() {
	let [e, t] = (0, N.useState)({}), n = (0, N.useRef)(null), r = (0, N.useRef)(/* @__PURE__ */ new Map());
	(0, N.useEffect)(() => () => n.current?.disconnect(), []);
	let i = () => n.current ??= new ResizeObserver((e) => t((t) => {
		let n = t;
		for (let r of e) {
			let e = r.target, i = e.dataset.cardAnchor ?? "";
			t[i] !== e.offsetHeight && (n = {
				...n,
				[i]: e.offsetHeight
			});
		}
		return n;
	}));
	return {
		heights: e,
		refFor: (e) => {
			let t = r.current.get(e);
			if (!t) {
				let n = null;
				t = (e) => {
					n && i().unobserve(n), n = e, e && i().observe(e);
				}, r.current.set(e, t);
			}
			return t;
		}
	};
}
function Wh({ origin: e, markdown: t, anchors: n, mode: r, patchRanges: i, visible: a }) {
	let [o, s] = (0, N.useState)(/* @__PURE__ */ new Map()), [c, l] = (0, N.useState)(0), [u, d] = (0, N.useState)(0);
	return (0, N.useLayoutEffect)(() => {
		let o = e.current, c = t.current;
		if (!o || !c || !a) return;
		let u = () => {
			s(zh({
				origin: o,
				markdown: c,
				anchors: n,
				mode: r,
				...i ? { patchRanges: i } : {}
			})), l(o.offsetHeight), d(o.offsetWidth);
		};
		u();
		let f = new MutationObserver(u);
		f.observe(c, {
			childList: !0,
			subtree: !0,
			characterData: !0
		});
		let p = new ResizeObserver(u);
		return p.observe(c), () => {
			f.disconnect(), p.disconnect();
		};
	}, [
		e,
		t,
		n,
		r,
		i,
		a
	]), {
		boxes: (0, N.useMemo)(() => {
			let e = new Set(n);
			return new Map([...o].filter(([t]) => e.has(t)));
		}, [o, n]),
		height: c,
		width: u
	};
}
function Gh(e) {
	let [t, n] = (0, N.useState)(!1);
	return (0, N.useLayoutEffect)(() => {
		let t = e.current;
		if (!t) return;
		let r = new ResizeObserver(() => n(t.offsetWidth < Hh));
		return r.observe(t), () => r.disconnect();
	}, [e]), t;
}
function Kh({ markdown: e, anchors: t }) {
	let [n, r] = (0, N.useState)(/* @__PURE__ */ new Map());
	return (0, N.useLayoutEffect)(() => {
		let n = e.current, i = t.filter((e) => e.startsWith("e:"));
		if (!n || i.length === 0) return r((e) => e.size === 0 ? e : /* @__PURE__ */ new Map());
		let a = () => r((e) => {
			let t = new Map(i.map((e) => [e, Hr(Fr(n, e))]));
			return t.size === e.size && [...t].every(([t, n]) => e.get(t) === n) ? e : t;
		});
		a();
		let o = new MutationObserver(a);
		return o.observe(n, {
			childList: !0,
			subtree: !0,
			characterData: !0
		}), () => o.disconnect();
	}, [e, t]), n;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/marks.ts
var qh = {
	error: 3,
	warning: 2,
	info: 1
}, Jh = {
	error: {
		icon: A,
		tone: "text-red-500",
		stroke: "#ef4444",
		tint: "bg-red-500/10"
	},
	warning: {
		icon: u,
		tone: "text-amber-500",
		stroke: "#f59e0b",
		tint: "bg-amber-500/10"
	},
	info: {
		icon: f,
		tone: "text-sky-500",
		stroke: "#0ea5e9",
		tint: "bg-sky-500/10"
	}
};
function Yh(e) {
	if (e === "error" || e === "warning" || e === "info") return e;
	throw Error(`lint finding has severity ${JSON.stringify(e)}, expected error, warning or info`);
}
var Xh = (e) => `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='6' height='3' viewBox='0 0 6 3'%3E%3Cpath d='M0 2.5 1.5 0.5 3 2.5 4.5 0.5 6 2.5' fill='none' stroke='${encodeURIComponent(e)}' stroke-width='1'/%3E%3C/svg%3E")`, Zh = (e) => ({
	backgroundImage: Xh(e),
	backgroundRepeat: "repeat-x",
	backgroundPosition: "bottom",
	backgroundSize: "6px 3px",
	paddingBottom: 2
}), Qh = (e, t) => qh[Yh(t.severity)] - qh[Yh(e.severity)] || e.line - t.line, $h = [
	"error",
	"warning",
	"info"
];
function eg(e) {
	let t = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let n of e) t[n.severity] += 1;
	return t;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/problem-filters.ts
function tg() {
	return {
		severity: /* @__PURE__ */ new Map(),
		type: /* @__PURE__ */ new Map()
	};
}
function ng(e) {
	return e.rule || e.source;
}
function rg(e, t) {
	return xe(e.severity, t.severity) && xe(ng(e), t.type);
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/ProblemsList.tsx
var ig = {
	error: "Errors",
	warning: "Warnings",
	info: "Info"
};
function ag(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) n.set(t(r), (n.get(t(r)) ?? 0) + 1);
	return n;
}
function og({ findings: e, shown: t, filters: n, onFiltersChange: r }) {
	let i = ag(e.filter((e) => xe(ng(e), n.type)), (e) => e.severity), a = ag(e.filter((e) => xe(e.severity, n.severity)), ng), o = [.../* @__PURE__ */ new Set([...e.map(ng), ...n.type.keys()])].sort(), s = n.severity.size + n.type.size;
	return /* @__PURE__ */ (0, P.jsx)("div", {
		className: "border-b border-border p-1 text-xs",
		"data-problem-filters": "",
		children: /* @__PURE__ */ (0, P.jsx)(De, {
			autoSubmit: !1,
			overflowMode: "wrap",
			leading: /* @__PURE__ */ (0, P.jsxs)("span", {
				className: "px-1 text-muted-foreground",
				children: [
					t,
					" of ",
					e.length,
					" problems"
				]
			}),
			filters: [{
				key: "severity",
				kind: "multi",
				label: "Severity",
				options: $h.map((e) => ({
					value: e,
					label: ig[e],
					count: i.get(e) ?? 0
				})),
				value: Object.fromEntries(n.severity),
				onChange: (e) => r({
					...n,
					severity: new Map(Object.entries(e))
				})
			}, {
				key: "type",
				kind: "multi",
				label: "Issue Type",
				options: o.map((e) => ({
					value: e,
					label: e,
					count: a.get(e) ?? 0
				})),
				value: Object.fromEntries(n.type),
				onChange: (e) => r({
					...n,
					type: new Map(Object.entries(e))
				})
			}],
			trailing: s > 0 && /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				className: "text-xs text-muted-foreground hover:text-foreground",
				onClick: () => r(tg()),
				children: "Clear filters"
			})
		})
	});
}
function sg({ source: e, tokens: t, findings: n, filters: r, onFiltersChange: i, onOpen: a, openLabel: o, getFindingActions: s }) {
	let c = (0, N.useMemo)(() => e.split("\n"), [e]), l = (0, N.useMemo)(() => {
		let e = He(n.map((e) => [
			e.workDir,
			e.file,
			e.source,
			e.rule,
			e.line,
			e.column,
			e.endLine,
			e.endColumn,
			e.message,
			e.severity,
			e.kind
		]));
		return n.map((t, n) => ({
			finding: t,
			key: e[n]
		}));
	}, [n]), u = (0, N.useMemo)(() => l.map((e) => e.key), [l]), [d, f] = We(u), p = l.filter(({ finding: e }) => rg(e, r)).sort((e, t) => Qh(e.finding, t.finding));
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)(og, {
		findings: n,
		shown: p.length,
		filters: r,
		onFiltersChange: i
	}), p.length === 0 ? /* @__PURE__ */ (0, P.jsx)("p", {
		className: "p-6 text-center text-xs text-muted-foreground",
		children: n.length ? "No problems match the current filters." : "No problems in this file."
	}) : /* @__PURE__ */ (0, P.jsx)("ul", { children: p.map(({ finding: e, key: n }) => {
		let r = Math.max(1, e.line - 3), i = Math.min(c.length, (e.endLine ?? e.line) + 3);
		return /* @__PURE__ */ (0, P.jsx)("li", {
			"data-problem-line": e.line,
			children: /* @__PURE__ */ (0, P.jsx)(M, {
				message: e.message,
				severity: e.severity,
				metadata: `${e.source} ${e.rule ?? ""}`,
				location: `${e.line}${e.column ? `:${e.column}` : ""}`,
				expanded: d.includes(n),
				onExpandedChange: (e) => f((t) => {
					let r = new Set([...t].filter((e) => l.some((t) => t.key === e)));
					return e ? r.add(n) : r.delete(n), [...r];
				}),
				actions: [
					{
						label: o,
						onSelect: () => a(e)
					},
					{
						label: "Copy description",
						onSelect: () => {
							navigator.clipboard.writeText(e.message);
						}
					},
					...s?.(e) ?? []
				],
				children: e.line > c.length ? /* @__PURE__ */ (0, P.jsx)("p", {
					role: "alert",
					children: "Source location is outside this file."
				}) : /* @__PURE__ */ (0, P.jsx)("ol", {
					"data-problem-context": "block",
					className: "overflow-x-auto border-l border-border bg-muted/20 py-1 font-mono text-xs leading-5",
					children: c.slice(r - 1, i).map((n, i) => {
						let a = r + i;
						return /* @__PURE__ */ (0, P.jsxs)("li", {
							className: "grid min-w-max grid-cols-[2.5rem_1fr] gap-2 pr-2",
							children: [/* @__PURE__ */ (0, P.jsx)("span", {
								className: "select-none text-right text-muted-foreground",
								children: a
							}), /* @__PURE__ */ (0, P.jsx)("code", {
								className: "whitespace-pre",
								style: a >= e.line && a <= (e.endLine ?? e.line) ? Zh(Jh[e.severity].stroke) : void 0,
								children: /* @__PURE__ */ (0, P.jsx)(v, {
									tokens: t?.[a - 1],
									content: n || " "
								})
							})]
						}, a);
					})
				})
			})
		}, n);
	}) })] });
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/blame.ts
var cg = (e) => /^0+$/.test(e.sha);
function lg(e, t) {
	let n = e.map((e) => Date.parse(e.commit.date)), r = Math.min(...n), i = Math.max(...n) - r, a = /* @__PURE__ */ new Map();
	return e.forEach((e, o) => {
		let s = e.startLine + e.lineCount - 1;
		if (e.startLine < 1 || e.lineCount < 1 || s > t) throw Error(`blame covers lines ${e.startLine}-${s} but the source has ${t} lines`);
		let c = i > 0 ? (n[o] - r) / i : 1;
		for (let t = e.startLine; t <= s; t++) a.set(t, {
			commit: e.commit,
			first: t === e.startLine,
			recency: c
		});
	}), a;
}
var ug = 864e5;
function dg(e, t = /* @__PURE__ */ new Date()) {
	let n = Math.floor((t.getTime() - Date.parse(e)) / ug);
	return n < 1 ? "today" : n < 30 ? `${n}d` : n < 365 ? `${Math.floor(n / 30)}mo` : `${Math.floor(n / 365)}y`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/finding-message.ts
function fg({ message: e, path: t }) {
	let n = t ? `${t}: ` : "";
	return n && e.startsWith(n) ? e.slice(n.length) : e;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/FindingRow.tsx
function pg({ finding: e, selected: t, onSelect: n, getFindingActions: r, truncate: i = !1 }) {
	let o = Jh[Yh(e.severity)];
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		"data-source-finding": e.line,
		className: a("flex items-center gap-2", i && "h-5"),
		children: [/* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			"aria-pressed": t,
			onClick: () => n(e),
			className: a("flex min-w-0 flex-1 items-center gap-1.5 text-left font-sans text-[11px] hover:underline aria-pressed:font-semibold aria-pressed:underline", o.tone),
			children: [
				/* @__PURE__ */ (0, P.jsx)(c, {
					icon: o.icon,
					className: "shrink-0"
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: a("text-foreground", i && "min-w-0 truncate"),
					children: fg(e)
				}),
				/* @__PURE__ */ (0, P.jsxs)("span", {
					className: "shrink-0 text-muted-foreground",
					children: [
						e.source,
						" ",
						e.rule
					]
				})
			]
		}), r && /* @__PURE__ */ (0, P.jsx)(Ke, { items: r(e) })]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/source-marks.ts
var mg = (e) => JSON.parse(e.dataset.sourceComment);
function hg(e, t, n, r) {
	let i = n.flatMap((e) => e.spans.filter((e) => e.line === t).map((t) => ({
		...t,
		anchor: e.anchor
	}))), a = 0, o = (r ?? []).map((e) => {
		let t = a;
		return a += e.content.length, {
			start: t,
			token: e
		};
	}), s = [.../* @__PURE__ */ new Set([
		0,
		e.length,
		...i.flatMap((e) => [e.start, e.end]),
		...o.map((e) => e.start)
	])].filter((t) => t <= e.length).sort((e, t) => e - t);
	return s.slice(0, -1).map((t, n) => {
		let r = o.filter((e) => e.start <= t).at(-1)?.token;
		return {
			start: t,
			text: e.slice(t, s[n + 1]),
			anchors: i.filter((e) => e.start <= t && e.end > t).map((e) => e.anchor),
			...r ? { token: r } : {}
		};
	});
}
function gg(e) {
	let t = (e) => (e instanceof Element ? e : e.parentElement)?.closest("[data-source-text]"), n = (e, n) => {
		let r = t(e);
		if (!r) return null;
		let i = document.createRange();
		return i.setStart(r, 0), i.setEnd(e, n), {
			line: Number(r.dataset.sourceText),
			column: i.toString().length
		};
	}, r = n(e.startContainer, e.startOffset), i = n(e.endContainer, e.endOffset);
	return r && i ? {
		start: r,
		end: i
	} : null;
}
function _g(e) {
	let t = gg(e);
	return !t || t.start.line !== t.end.line ? null : {
		line: t.start.line,
		start: t.start.column,
		end: t.end.column
	};
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/SourceView.tsx
function vg({ blame: e }) {
	if (!e) return /* @__PURE__ */ (0, P.jsx)("span", {});
	let { commit: t, first: n, recency: r } = e, i = cg(t);
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		"data-blame-sha": n ? t.sha : void 0,
		title: i ? "Not committed yet" : `${t.sha.slice(0, 12)} · ${t.author} · ${t.date.slice(0, 10)}\n${t.summary}`,
		className: a("relative flex min-w-0 items-start gap-1.5 pl-2 pr-2 font-sans text-[11px] text-muted-foreground", n && "border-t border-border/60"),
		children: [/* @__PURE__ */ (0, P.jsx)("span", {
			"aria-hidden": !0,
			className: a("absolute inset-y-0 left-0 w-[3px]", i ? "bg-amber-500" : "bg-sky-500"),
			style: { opacity: i ? 1 : .15 + .85 * r }
		}), n && (i ? /* @__PURE__ */ (0, P.jsx)("span", {
			className: "italic",
			children: "Uncommitted"
		}) : /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "shrink-0 font-mono text-foreground/70",
				children: t.sha.slice(0, 7)
			}),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "min-w-0 flex-1 truncate",
				children: t.author
			}),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "shrink-0 tabular-nums",
				children: dg(t.date)
			})
		] }))]
	});
}
function yg({ source: e, tokens: t, blame: n, findings: r, showFindings: i, marks: o, comments: s, selected: l, onSelect: u, onFocus: d, getFindingActions: f }) {
	return /* @__PURE__ */ (0, P.jsx)("ol", {
		className: "py-2 font-mono text-xs leading-5",
		children: e.split("\n").map((e, p) => {
			let m = p + 1, h = r.filter((e) => e.line === m).sort(Qh), g = h[0] ? Jh[Yh(h[0].severity)] : null, _ = o.filter((e) => e.spans.at(-1)?.line === m), v = i || l === Nm(m) ? h : [];
			return /* @__PURE__ */ (0, P.jsxs)("li", {
				"data-source-line": m,
				className: a("grid pr-3", n ? "grid-cols-[13rem_1.25rem_2.5rem_1fr]" : "grid-cols-[1.25rem_2.5rem_1fr]"),
				children: [
					n && /* @__PURE__ */ (0, P.jsx)(vg, { blame: n.get(m) }),
					g && !i ? /* @__PURE__ */ (0, P.jsx)("button", {
						type: "button",
						"data-source-mark": m,
						"aria-label": `${h.length} ${h.length === 1 ? "finding" : "findings"} on line ${m}`,
						"aria-pressed": l === Nm(m),
						onClick: () => l === Nm(m) ? d(null) : u(h[0]),
						className: a("flex h-5 items-center justify-center rounded-sm hover:bg-muted", g.tone),
						children: /* @__PURE__ */ (0, P.jsx)(c, {
							icon: g.icon,
							className: "text-[13px]"
						})
					}) : /* @__PURE__ */ (0, P.jsx)("span", {
						className: a("flex h-5 items-center justify-center", g?.tone),
						children: g && /* @__PURE__ */ (0, P.jsx)(c, {
							icon: g.icon,
							className: "text-[13px]"
						})
					}),
					/* @__PURE__ */ (0, P.jsx)("span", {
						className: "select-none pr-3 text-right text-muted-foreground tabular-nums",
						children: m
					}),
					/* @__PURE__ */ (0, P.jsxs)("span", {
						className: "min-w-0 whitespace-pre-wrap break-words",
						children: [
							/* @__PURE__ */ (0, P.jsx)("span", {
								"data-source-text": m,
								style: g ? Zh(g.stroke) : void 0,
								children: e ? hg(e, m, o, t?.[p]).map((e) => {
									let t = e.token ? x(e.token) : void 0;
									return e.anchors.length === 0 ? /* @__PURE__ */ (0, P.jsx)("span", {
										style: t,
										children: e.text
									}, e.start) : /* @__PURE__ */ (0, P.jsx)("mark", {
										"data-source-comment": JSON.stringify(e.anchors),
										onClick: () => window.getSelection()?.isCollapsed && d(e.anchors[0]),
										style: t,
										className: a("cursor-pointer rounded-sm text-inherit", e.anchors.includes(l ?? "") ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25"),
										children: e.text
									}, e.start);
								}) : " "
							}),
							_.map(({ anchor: e }) => {
								let t = hi(s, e), n = pi(t)[0];
								return n ? /* @__PURE__ */ (0, P.jsxs)("button", {
									type: "button",
									"data-source-thread": e,
									onClick: () => d(e),
									className: "flex w-full items-center gap-1.5 text-left font-sans text-[11px] text-amber-600 hover:underline [[data-theme=dark]_&]:text-amber-400",
									children: [
										/* @__PURE__ */ (0, P.jsx)(Mp, {
											author: n.author ?? null,
											size: "xs",
											bare: !0
										}),
										/* @__PURE__ */ (0, P.jsx)("span", {
											className: "shrink-0 font-medium text-foreground",
											title: n.author?.name,
											children: li(n.author)
										}),
										/* @__PURE__ */ (0, P.jsx)("span", {
											className: "truncate text-muted-foreground",
											children: n.body
										}),
										t.length > 1 && /* @__PURE__ */ (0, P.jsxs)("span", {
											className: "shrink-0 text-muted-foreground",
											children: [
												"· ",
												t.length - 1,
												" ",
												t.length === 2 ? "reply" : "replies"
											]
										})
									]
								}, e) : null;
							}),
							v.map((e) => /* @__PURE__ */ (0, P.jsx)(pg, {
								finding: e,
								selected: l === Nm(e.line),
								onSelect: u,
								getFindingActions: f
							}, `${e.rule}:${e.message}`))
						]
					})
				]
			}, m);
		})
	});
}
var bg = 24, xg = 60;
function Sg(e) {
	return `s:${JSON.stringify([
		e.line,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function Cg(e) {
	if (!e.startsWith("s:")) throw Error(`not a source anchor: ${JSON.stringify(e)}`);
	let t = JSON.parse(e.slice(2)), [n, r, i, a] = Array.isArray(t) ? t : [];
	if (!Number.isInteger(n) || n < 1 || typeof r != "string" || typeof i != "string" || typeof a != "string") throw Error(`malformed source anchor ${JSON.stringify(e)}`);
	return {
		line: n,
		prefix: r,
		quote: i,
		suffix: a
	};
}
function wg(e) {
	let t = [0];
	for (let n = e.indexOf("\n"); n !== -1; n = e.indexOf("\n", n + 1)) t.push(n + 1);
	return t;
}
function Tg(e, t, n) {
	let r = t[n.line - 1];
	if (r === void 0) throw Error(`line ${n.line} is outside the ${t.length}-line source`);
	let i = t[n.line] === void 0 ? e.length : t[n.line] - 1;
	return Math.min(r + n.column, i);
}
function Eg(e, t) {
	let n = 1;
	for (; n < e.length && e[n] <= t;) n++;
	return n;
}
function Dg(e, t, n) {
	let r = wg(e), i = Tg(e, r, t), a = Tg(e, r, n), o = e.slice(i, a);
	return o.trim() ? {
		line: Eg(r, i),
		prefix: e.slice(Math.max(0, i - bg), i),
		quote: o,
		suffix: e.slice(a, a + bg)
	} : null;
}
function Og(e, t) {
	let n = t.prefix + t.quote + t.suffix, r = wg(e), i = null, a = Infinity;
	for (let o = e.indexOf(n); o !== -1; o = e.indexOf(n, o + 1)) {
		let e = o + t.prefix.length, n = Math.abs(Eg(r, e) - t.line);
		n < a && (i = e, a = n);
	}
	if (i === null) return [];
	let o = i + t.quote.length, s = [];
	for (let t = Eg(r, i); t <= r.length && r[t - 1] < o; t++) {
		let n = r[t - 1], a = r[t] === void 0 ? e.length : r[t] - 1, c = {
			line: t,
			start: Math.max(i, n) - n,
			end: Math.min(o, a) - n
		};
		c.end > c.start && s.push(c);
	}
	return s;
}
function kg(e) {
	let t = e.quote.replace(/\s+/g, " ").trim();
	return `Line ${e.line} · “${t.length > xg ? `${t.slice(0, xg)}…` : t}”`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/InlineThread.tsx
var Ag = 340, jg = 6;
function Mg({ anchor: e, label: t, boxes: n, containerWidth: r, thread: i, config: o, callbacks: s, pinned: l, onPin: u, onClose: f, onPointerEnter: p, onPointerLeave: m, seamless: h = !1, renderRootDetails: g, detailOptions: _ }) {
	(0, N.useEffect)(() => {
		let e = (e) => e.key === "Escape" && f();
		return window.addEventListener("keydown", e), () => window.removeEventListener("keydown", e);
	}, [f]);
	let v = n[0], y = n.at(-1);
	if (!v || !y) return null;
	let b = Math.min(Ag, r), x = Math.max(0, Math.min(v.left, r - b)), S = !!s.onCreate;
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-inline-thread": e,
		role: "dialog",
		"aria-label": `Comments on ${t}`,
		onMouseEnter: p,
		onMouseLeave: m,
		onMouseDown: u,
		className: a("absolute z-30 bg-card", h ? "rounded-md font-sans shadow-[0_4px_16px_rgba(0,0,0,0.12)] ring-1 ring-border/60" : "rounded-lg border shadow-lg", !h && (l ? "border-primary/50" : "border-border")),
		style: {
			top: y.top + y.height + (h ? 2 : jg),
			left: x,
			width: b
		},
		children: [!h && /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: t
				}),
				!l && /* @__PURE__ */ (0, P.jsx)("span", { children: "click to pin" }),
				/* @__PURE__ */ (0, P.jsx)("button", {
					type: "button",
					"aria-label": "Close comments",
					onClick: f,
					className: "rounded p-0.5 hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
					children: /* @__PURE__ */ (0, P.jsx)(c, {
						icon: d,
						className: "text-xs"
					})
				})
			]
		}), /* @__PURE__ */ (0, P.jsx)("div", {
			className: "max-h-[360px] overflow-y-auto",
			children: /* @__PURE__ */ (0, P.jsx)(_m, {
				..._,
				comments: i,
				config: o,
				anchor: e,
				compact: !0,
				defaultExpanded: !0,
				hideComposer: !S || i.length > 0 && !l,
				autoFocusComposer: i.length === 0,
				composerPlaceholder: i.length > 0 ? "Add another comment…" : "Add a comment…",
				...s,
				...g ? { renderRootDetails: g } : {}
			})
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/InlineAuthors.tsx
function Ng({ markdown: e, comments: t, focused: n, onFocus: r }) {
	let [i, o] = (0, N.useState)(/* @__PURE__ */ new Map());
	(0, N.useLayoutEffect)(() => {
		let n = e.current;
		if (!n) return;
		let r = /* @__PURE__ */ new Map(), i = () => {
			let e = !1;
			for (let i of pi(t)) {
				let t = i.anchor;
				if (!t?.startsWith("q:") || r.get(i.id)?.isConnected) continue;
				let a = xh(n, dh(t));
				if (!a) continue;
				let o = document.createElement("span");
				o.dataset.commentAuthor = i.id, o.className = "ml-1 inline-flex max-w-full align-middle leading-none", o.contentEditable = "false", a.collapse(!1), a.insertNode(o), r.set(i.id, o), e = !0;
			}
			e && o(new Map(r));
		};
		i();
		let a = new MutationObserver(i);
		return a.observe(n, {
			childList: !0,
			subtree: !0
		}), () => {
			a.disconnect();
			for (let e of r.values()) e.remove();
		};
	}, [e, t]);
	let s = mi(t);
	return pi(t).map((e) => {
		let t = i.get(e.id), o = e.anchor;
		if (!t || !o) return null;
		let c = s.get(e.id)?.length ?? 0;
		return (0, Nf.createPortal)(/* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			"aria-label": `Open comment by ${li(e.author ?? null)}`,
			"aria-pressed": n === o,
			title: e.body,
			className: a("inline-flex max-w-[min(40ch,100%)] items-center gap-1 rounded-full border border-border bg-muted py-px pl-px pr-2 text-xs font-normal text-muted-foreground ring-offset-background hover:text-foreground hover:ring-2 hover:ring-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary", n === o && "text-foreground ring-2 ring-primary"),
			onClick: () => r(o),
			children: [
				/* @__PURE__ */ (0, P.jsx)(Mp, {
					author: e.author ?? null,
					size: "xs",
					bare: !0
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "truncate",
					children: bi(e.body)
				}),
				c > 0 && /* @__PURE__ */ (0, P.jsxs)("span", {
					className: "shrink-0 tabular-nums",
					children: ["· ", c]
				})
			]
		}), t, e.id);
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentDraftCard.tsx
function Pg({ label: e, onDiscard: t, onCreate: n, ...r }) {
	let [i, a] = (0, N.useState)(!1), o = (0, N.useRef)(!1), s = vt(), l = () => {
		o.current || (r.anchor && s(r.anchor), t());
	};
	return y(!0, l), /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
		children: [
			/* @__PURE__ */ (0, P.jsx)(c, {
				icon: ce,
				className: "text-xs"
			}),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "min-w-0 flex-1 truncate",
				children: e
			}),
			/* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				"aria-label": "Discard comment",
				disabled: i,
				onClick: l,
				className: "rounded p-0.5 hover:bg-muted hover:text-foreground disabled:opacity-40 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
				children: /* @__PURE__ */ (0, P.jsx)(c, {
					icon: d,
					className: "text-xs"
				})
			})
		]
	}), /* @__PURE__ */ (0, P.jsx)("div", {
		className: "p-2",
		children: /* @__PURE__ */ (0, P.jsx)(Tp, {
			...r,
			collapsible: !1,
			autoFocus: !0,
			placeholder: "Add a comment…",
			onCancel: l,
			onCreate: async (e, t) => {
				if (!n) throw Error("Comment creation is not configured");
				o.current = !0, a(!0);
				try {
					t === void 0 ? await n(e) : await n(e, t);
				} finally {
					o.current = !1, a(!1);
				}
			}
		})
	})] });
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/CommentDraft.tsx
var Fg = {
	width: 104,
	height: 28
}, Ig = 320, Lg = 190, Rg = 6;
function zg({ anchor: e, label: t, position: n, containerWidth: r, expanded: i, config: o, onCreate: s, onExpand: l, onClose: u, createActions: d }) {
	let f = n.top - Fg.height - 8, p = n.top >= Lg, m = Math.min(Ig, r), h = Math.max(m / 2, Math.min(n.left, r - m / 2)), g = i ? {
		top: p ? f + Fg.height : n.bottom + Rg,
		left: h,
		width: m,
		transform: `translate(-50%, ${p ? "-100%" : "0"})`
	} : {
		top: f,
		left: n.left,
		width: Fg.width,
		transform: "translate(-50%, 0)"
	};
	return /* @__PURE__ */ (0, P.jsx)("div", {
		"data-comment-draft": e,
		role: i ? "dialog" : void 0,
		"aria-label": i ? `New comment on ${t}` : void 0,
		onMouseDown: (e) => !i && e.preventDefault(),
		className: a("absolute z-30 overflow-hidden border bg-card shadow-md transition-[top,left,width,transform,border-radius,box-shadow] duration-200 ease-out", i ? "rounded-lg border-primary/50 shadow-lg" : "rounded-full border-border"),
		style: g,
		children: i ? /* @__PURE__ */ (0, P.jsx)(Pg, {
			label: t,
			config: o,
			anchor: e,
			onCreate: s,
			onDiscard: u,
			...d ? { createActions: d } : {}
		}) : /* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			onClick: l,
			className: "inline-flex w-full items-center justify-center gap-1 px-2.5 text-xs font-medium hover:bg-muted",
			style: { height: Fg.height },
			children: [/* @__PURE__ */ (0, P.jsx)(c, {
				icon: ce,
				className: "text-sm"
			}), "Comment"]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/views.ts
var Bg = [
	"preview",
	"source",
	"problems"
];
function Vg(e) {
	return Bg.includes(e);
}
var Hg = "problems", Ug = {
	moved: "warning",
	ambiguous: "warning",
	outdated: "danger",
	orphaned: "danger"
};
function Wg({ state: e }) {
	return /* @__PURE__ */ (0, P.jsx)("span", {
		"data-drift-badge": e,
		className: "inline-flex shrink-0",
		children: /* @__PURE__ */ (0, P.jsx)(E, {
			tone: Ug[e],
			variant: "soft",
			size: "xs",
			clickToCopy: !1,
			children: Br[e]
		})
	});
}
function Gg({ item: e, retarget: t }) {
	let [n, r] = (0, N.useState)({
		busy: !1,
		error: null
	});
	async function i() {
		r({
			busy: !0,
			error: null
		});
		try {
			let n = await t.pickAnchor(e.anchor);
			if (n) for (let r of e.ids) await (n.quote ? t.onRetarget(r, n.anchor, n.quote) : t.onRetarget(r, n.anchor));
			r({
				busy: !1,
				error: null
			});
		} catch (e) {
			r({
				busy: !1,
				error: `Could not re-anchor: ${e instanceof Error ? e.message : String(e)}`
			});
		}
	}
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)(o, {
		type: "button",
		size: "sm",
		variant: "outline",
		loading: n.busy,
		onClick: () => void i(),
		className: "shrink-0",
		children: "Re-anchor"
	}), n.error && /* @__PURE__ */ (0, P.jsx)("span", {
		role: "alert",
		className: "basis-full text-[11px] text-destructive",
		children: n.error
	})] });
}
function Kg({ items: e, retarget: t, onSelect: n, className: r }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, P.jsxs)("section", {
		"data-drift-list": "",
		"aria-label": "Comments needing re-anchoring",
		className: a("rounded-md border border-border bg-card", r),
		children: [/* @__PURE__ */ (0, P.jsxs)("h3", {
			className: "px-2 pt-2 text-xs font-medium text-muted-foreground",
			children: ["Needs re-anchoring · ", e.length]
		}), /* @__PURE__ */ (0, P.jsx)("ul", {
			className: "space-y-1 p-2",
			children: e.map((e) => /* @__PURE__ */ (0, P.jsxs)("li", {
				"data-drift-item": e.anchor,
				"data-drift-state": e.drift.state,
				className: "flex flex-wrap items-center gap-2 rounded-md px-1 py-1 text-xs",
				children: [
					/* @__PURE__ */ (0, P.jsx)(Wg, { state: e.drift.state }),
					n ? /* @__PURE__ */ (0, P.jsx)("button", {
						type: "button",
						onClick: () => n(e),
						className: "min-w-0 flex-1 truncate text-left font-medium hover:underline",
						children: e.label
					}) : /* @__PURE__ */ (0, P.jsx)("span", {
						className: "min-w-0 flex-1 truncate font-medium",
						children: e.label
					}),
					t && e.drift.reanchor !== !1 && /* @__PURE__ */ (0, P.jsx)(Gg, {
						item: e,
						retarget: t
					}),
					e.drift.note && /* @__PURE__ */ (0, P.jsx)("p", {
						"data-drift-note": "",
						className: "basis-full text-[11px] text-muted-foreground",
						children: e.drift.note
					})
				]
			}, e.anchor))
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/FileThreads.tsx
function qg({ entries: e, drafting: t, config: n, callbacks: r, options: i, renderRootDetails: a, composerActions: o, retarget: s, onCreate: l, onDiscardDraft: u }) {
	let [d, f] = (0, N.useState)(!1);
	return e.length === 0 && !t ? null : /* @__PURE__ */ (0, P.jsxs)("section", {
		"data-file-threads": "",
		"aria-label": "File comments",
		className: "shrink-0 rounded-md border border-border bg-card",
		children: [/* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			"aria-expanded": !d,
			onClick: () => f((e) => !e),
			className: "flex w-full items-center gap-1 px-2 py-1 text-left text-xs font-medium text-muted-foreground hover:text-foreground",
			children: [
				/* @__PURE__ */ (0, P.jsx)(c, { icon: d ? re : ne }),
				"File comments · ",
				e.length
			]
		}), !d && /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "max-h-72 space-y-2 overflow-y-auto border-t border-border p-2",
			children: [t && /* @__PURE__ */ (0, P.jsx)("div", {
				className: "rounded-lg border border-primary/50 bg-card",
				children: /* @__PURE__ */ (0, P.jsx)(Jg, {
					config: n,
					composerActions: o,
					onCreate: l,
					onDiscard: u
				})
			}), e.map((e) => /* @__PURE__ */ (0, P.jsx)("div", {
				"data-file-thread": e.anchor,
				className: "rounded-lg border border-border bg-card",
				children: /* @__PURE__ */ (0, P.jsx)(Yg, {
					entry: e,
					config: n,
					callbacks: r,
					options: i,
					renderRootDetails: a,
					retarget: s
				})
			}, e.anchor))]
		})]
	});
}
function Jg({ config: e, composerActions: t, onCreate: n, onDiscard: r }) {
	return /* @__PURE__ */ (0, P.jsx)(Pg, {
		label: "Comment on file",
		anchor: oi,
		config: e,
		onCreate: n,
		onDiscard: r,
		...t ? { createActions: t } : {}
	});
}
function Yg({ entry: { anchor: e, label: t, drift: n, note: r, thread: i }, config: a, callbacks: o, options: s, renderRootDetails: c, retarget: l, defaultExpanded: u = !0 }) {
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex flex-wrap items-center gap-2 px-2 pt-2 text-[11px] text-muted-foreground",
		children: [
			n && /* @__PURE__ */ (0, P.jsx)(Wg, { state: n }),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "font-medium",
				children: t
			}),
			r && /* @__PURE__ */ (0, P.jsxs)("span", {
				className: "min-w-0 flex-1",
				children: [" — ", r]
			}),
			n && l && /* @__PURE__ */ (0, P.jsx)(Gg, {
				item: {
					ids: pi(i).map((e) => e.id),
					anchor: e,
					label: t,
					drift: { state: n }
				},
				retarget: l
			})
		]
	}), /* @__PURE__ */ (0, P.jsx)(_m, {
		...s,
		comments: i,
		config: a,
		anchor: e,
		compact: !0,
		defaultExpanded: u,
		hideComposer: i.length > 0,
		composerPlaceholder: "Add a comment…",
		...o,
		...c ? { renderRootDetails: c } : {}
	})] });
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/EditThreads.tsx
function Xg({ entries: e, draft: t, focused: n, config: r, callbacks: i, options: o, renderRootDetails: s, composerActions: c, onFocus: l, onCreate: u, onDiscardDraft: d, files: f }) {
	let p = (0, N.useRef)(null);
	return (0, N.useEffect)(() => {
		[...p.current?.querySelectorAll("[data-card-anchor]") ?? []].find((e) => e.dataset.cardAnchor === n)?.scrollIntoView?.({ block: "nearest" });
	}, [n]), e.length === 0 && !t && !f?.drafting && !f?.entries.length ? null : /* @__PURE__ */ (0, P.jsxs)("aside", {
		ref: p,
		"data-edit-threads": "",
		"aria-label": "Comments on the source",
		className: "w-80 shrink-0 space-y-2 overflow-y-auto border-l border-border p-2",
		children: [
			t && /* @__PURE__ */ (0, P.jsx)("div", {
				"data-comment-draft": t.anchor,
				className: "rounded-lg border border-primary/50 bg-card shadow-lg",
				children: /* @__PURE__ */ (0, P.jsx)(Pg, {
					label: t.label,
					anchor: t.anchor,
					config: r,
					onCreate: u,
					onDiscard: d,
					...c ? { createActions: c } : {}
				})
			}),
			f?.drafting && /* @__PURE__ */ (0, P.jsx)("div", {
				"data-comment-draft": "__document__",
				className: "rounded-lg border border-primary/50 bg-card shadow-lg",
				children: /* @__PURE__ */ (0, P.jsx)(Jg, {
					config: r,
					composerActions: c,
					onCreate: f.onCreate,
					onDiscard: f.onDiscardDraft
				})
			}),
			f?.entries.map((e) => /* @__PURE__ */ (0, P.jsx)("div", {
				"data-card-anchor": e.anchor,
				"data-file-thread": e.anchor,
				className: a("rounded-lg border bg-card", e.anchor === n ? "border-primary/50 shadow-lg" : "border-border shadow-sm"),
				children: /* @__PURE__ */ (0, P.jsx)(Yg, {
					entry: e,
					config: r,
					callbacks: i,
					options: o,
					renderRootDetails: s
				})
			}, e.anchor)),
			e.map((e) => /* @__PURE__ */ (0, P.jsxs)("div", {
				"data-card-anchor": e.anchor,
				onClick: () => l(e),
				className: a("rounded-lg border bg-card", e.anchor === n ? "border-primary/50 shadow-lg" : "border-border shadow-sm"),
				children: [/* @__PURE__ */ (0, P.jsx)("div", {
					className: "truncate px-2 pt-2 text-[11px] text-muted-foreground",
					children: e.label
				}), /* @__PURE__ */ (0, P.jsx)(_m, {
					...o,
					comments: e.thread,
					config: r,
					anchor: e.anchor,
					compact: !0,
					defaultExpanded: e.anchor === n,
					hideComposer: e.thread.length > 0,
					composerPlaceholder: "Add a comment…",
					...i,
					...s ? { renderRootDetails: s } : {}
				})]
			}, e.anchor))
		]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/patch-comments.ts
function Zg(e, t) {
	return new Map(t.filter(Zr).map((t) => {
		let n = $r(t);
		return [t, {
			anchor: t,
			patch: n,
			match: ti(e, n)
		}];
	}));
}
var Qg = ({ quote: e, cell: t }) => ({
	...e ? { quote: e } : {},
	...t ? { cell: t } : {}
});
function $g(e, t = /* @__PURE__ */ new Map()) {
	let n = /* @__PURE__ */ new Map();
	for (let { anchor: r, match: i } of e.values()) i.state === "matched" && n.set(r, {
		start: i.startLine,
		end: i.endLine,
		...Qg(t.get(r) ?? {})
	});
	return n;
}
function e_(e, t) {
	let n = e.split("\n");
	return [...$g(t)].map(([e, { start: t, end: r }]) => ({
		anchor: e,
		spans: Array.from({ length: r - t + 1 }, (e, r) => ({
			line: t + r,
			start: 0,
			end: n[t + r - 1]?.length ?? 0
		}))
	}));
}
function t_({ patch: e, match: t }) {
	return t.state === "matched" ? ei({
		...e,
		startLine: t.startLine,
		endLine: t.endLine
	}) : `${t.state === "outdated" ? "Outdated" : "Ambiguous"} · ${ei(e)}`;
}
function n_({ raw: e, lines: t, quote: n, cell: r }) {
	if (r && t.startLine !== t.endLine) throw Error(`a table cell lies on one line, not lines ${t.startLine}-${t.endLine}`);
	let i = Jr(e, t.startLine, t.endLine);
	return {
		anchor: Qr(i),
		label: r ? oh(r) : ei(i),
		...n && Dh(n, i.lines) ? { quote: n } : {},
		...r ? { cell: r } : {}
	};
}
// ../../../clicky-ui/packages/ui/src/comments/annotation-screenshot.ts
var r_, i_, a_, o_ = 0, s_ = 1e4;
function c_(e) {
	e.getTracks().forEach((e) => e.stop());
}
function l_() {
	let e = navigator.mediaDevices;
	if (a_ ??= crypto.randomUUID(), !e.setCaptureHandleConfig) throw Error("browser does not support capture handles");
	return e.setCaptureHandleConfig({
		handle: a_,
		exposeOrigin: !1,
		permittedOrigins: [window.location.origin]
	}), navigator.mediaDevices.getDisplayMedia({
		video: { displaySurface: "browser" },
		audio: !1,
		preferCurrentTab: !0,
		selfBrowserSurface: "include",
		monitorTypeSurfaces: "exclude"
	});
}
function u_(e) {
	let t = e.getVideoTracks()[0];
	return t?.readyState === "live" && t.getSettings().displaySurface === "browser" && t.getCaptureHandle?.()?.handle === a_;
}
function d_(e) {
	let t = e.getVideoTracks()[0];
	if (!t || t.readyState !== "live") throw Error("selected display stream has no live video track");
	if (!u_(e)) throw Error("select this tab to capture a playground element");
}
function f_(e) {
	try {
		d_(e);
	} catch (t) {
		throw c_(e), t;
	}
	return r_ = e, e.getTracks().forEach((t) => {
		t.addEventListener("ended", () => {
			r_ === e && (r_ = void 0);
		}, { once: !0 });
	}), e;
}
function p_() {
	if (r_ && u_(r_)) return Promise.resolve(r_);
	if (r_ && c_(r_), r_ = void 0, i_) return i_;
	let e = o_, t = l_().then((t) => {
		if (e !== o_) throw c_(t), Error("Screenshot capture session was closed");
		return f_(t);
	});
	return i_ = t, t.then(() => {
		i_ === t && (i_ = void 0);
	}, () => {
		i_ === t && (i_ = void 0);
	}), t;
}
function m_() {
	return new Promise((e) => {
		requestAnimationFrame(() => requestAnimationFrame(() => e()));
	});
}
function h_(e, t) {
	let n = "rect" in t ? t.rect : t.getBoundingClientRect();
	if (!("rect" in t ? t.isLive() : t.isConnected) || !Object.values({
		x: n.x,
		y: n.y,
		width: n.width,
		height: n.height
	}).every(Number.isFinite) || n.width <= 0 || n.height <= 0) throw Error("selected element is no longer visible");
	if (!e.videoWidth || !e.videoHeight) throw Error("captured tab has no video frame");
	let r = e.videoWidth / window.innerWidth, i = e.videoHeight / window.innerHeight;
	if (!Number.isFinite(r) || !Number.isFinite(i) || Math.abs(r - i) > Math.max(r, i) * .02) throw Error("captured tab has invalid viewport dimensions");
	let a = Math.max(0, Math.floor(n.x * r)), o = Math.max(0, Math.floor(n.y * i)), s = Math.min(e.videoWidth, Math.ceil((n.x + n.width) * r)), c = Math.min(e.videoHeight, Math.ceil((n.y + n.height) * i));
	if (s <= a || c <= o) throw Error("selected element is outside the captured viewport");
	return {
		x: a,
		y: o,
		width: s - a,
		height: c - o
	};
}
async function g_(e, t) {
	let n = document.createElement("video");
	n.muted = !0, n.playsInline = !0, n.srcObject = e;
	try {
		await n.play(), await m_();
		let e = h_(n, t), r = document.createElement("canvas");
		r.width = e.width, r.height = e.height;
		let i = r.getContext("2d");
		if (!i) throw Error("browser could not create a screenshot canvas");
		i.drawImage(n, e.x, e.y, e.width, e.height, 0, 0, r.width, r.height);
		let a = await new Promise((e, t) => {
			r.toBlob((n) => n ? e(n) : t(/* @__PURE__ */ Error("PNG encoding failed")), "image/png");
		});
		return await new Promise((e, t) => {
			let n = new FileReader();
			n.onerror = () => t(/* @__PURE__ */ Error("PNG serialization failed")), n.onload = () => e(String(n.result)), n.readAsDataURL(a);
		});
	} finally {
		n.pause(), n.srcObject = null;
	}
}
function __(e) {
	return e instanceof DOMException && ["NotAllowedError", "AbortError"].includes(e.name) ? "cancelled" : "failed";
}
function v_() {
	let e = navigator.mediaDevices;
	return !!(e?.getDisplayMedia && e.setCaptureHandleConfig && crypto.randomUUID);
}
async function y_(e) {
	if (!v_()) return {
		status: "unavailable",
		reason: "unsupported"
	};
	try {
		let t = await p_();
		try {
			d_(t);
		} catch (e) {
			throw r_ === t && (r_ = void 0), c_(t), e;
		}
		let n = t.getVideoTracks()[0];
		if (!n) throw Error("selected display stream has no video track");
		let r = n.clone(), i = new MediaStream([r]), a;
		try {
			let n = g_(i, e), r = new Promise((e, n) => {
				a = setTimeout(() => {
					r_ === t && (r_ = void 0), c_(t), n(/* @__PURE__ */ Error("screenshot capture timed out"));
				}, s_);
			});
			return {
				status: "captured",
				dataUrl: await Promise.race([n, r])
			};
		} finally {
			clearTimeout(a), r.stop();
		}
	} catch (e) {
		return {
			status: "unavailable",
			reason: __(e)
		};
	}
}
var b_ = "screenshot";
function x_({ host: e, screenshot: t }) {
	if (!t || !v_()) return e;
	if (e?.some((e) => e.id === "screenshot")) throw Error(`FileReview: the create action id "${b_}" is reserved for "Post with screenshot"`);
	return [...e ?? [{
		id: "post",
		label: "Post comment"
	}], {
		id: b_,
		label: "Post with screenshot"
	}];
}
async function S_({ onCreate: e, input: t, action: n, hostActions: r, region: i }) {
	return n === "screenshot" ? e({
		...t,
		screenshot: await y_(i())
	}) : n === void 0 || n === "post" && !r?.length ? e(t) : e(t, n);
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/selection-lines.ts
var C_ = "[data-source-start]";
function w_(e, t) {
	let n = document.createRange();
	return n.selectNodeContents(t), e.compareBoundaryPoints(Range.END_TO_START, n) < 0 && e.compareBoundaryPoints(Range.START_TO_END, n) > 0;
}
function T_(e, t) {
	let n = [];
	for (let r of t.querySelectorAll(C_)) {
		if (r.parentElement?.closest(C_) || !w_(e, r)) continue;
		let t = Xm(r).filter(({ row: t }) => w_(e, t));
		t.length > 0 ? n.push(...t.map(({ line: e }) => ({
			startLine: e,
			endLine: e
		}))) : n.push({
			startLine: Number(r.dataset.sourceStart),
			endLine: Number(r.dataset.sourceEnd)
		});
	}
	return n.length === 0 ? null : {
		startLine: Math.min(...n.map((e) => e.startLine)),
		endLine: Math.max(...n.map((e) => e.endLine))
	};
}
function E_(e) {
	let t = gg(e);
	if (!t) return null;
	let { start: n, end: r } = t, i = r.column === 0 && r.line > n.line ? r.line - 1 : r.line;
	return {
		startLine: n.line,
		endLine: i
	};
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/FileReview.tsx
var D_ = { statuses: si }, O_ = 96, k_ = "file-draft", A_ = 450, j_ = 300, M_ = 16, N_ = [], P_ = [], F_ = {}, I_ = [], L_ = Nm(1).replace(/\d+$/, ""), R_ = (e) => e.startsWith(L_), z_ = {
	top: 0,
	bottom: 0,
	left: 0
}, B_ = (e) => /\.(md|mdx|markdown)$/i.test(e), V_ = (e) => e.startsWith("q:") || e.startsWith("s:") || e.startsWith("r:"), H_ = (e) => ({
	anchor: uh(e),
	label: Fh(e)
});
function U_({ mode: e, active: t }) {
	return e === "block" ? a("border-l-[3px] border-amber-400", t ? "bg-amber-200/50 [[data-theme=dark]_&]:bg-amber-400/20" : "bg-amber-100/30 [[data-theme=dark]_&]:bg-amber-400/5") : a("rounded-sm", t ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25");
}
function W_({ anchor: e, found: t, placements: n, extras: r, raw: i }) {
	if (e === "__document__") return "File";
	if (e.startsWith("e:")) return t ? "Element" : "Element not found";
	if (e.startsWith("r:")) {
		let t = n.get(e) ?? Zg(i, [e]).get(e), a = r.get(e)?.cell;
		return a && t.match.state === "matched" ? oh(a) : t_(t);
	}
	let a = e.startsWith("q:") ? Fh(dh(e)) : e.startsWith("s:") ? kg(Cg(e)) : e.replace("line:", "line ");
	return t ? a : `Text not found · ${a}`;
}
function G_({ anchor: e, placements: t, elements: n }) {
	return Zr(e) ? Vr(t.get(e)?.match ?? { state: "outdated" }) : e.startsWith("e:") ? n.has(e) ? n.get(e) : "orphaned" : null;
}
var K_ = /[\p{L}\p{N}'’-]/u;
function q_(e, t) {
	let n = document.caretRangeFromPoint?.(e, t), r = n?.startContainer;
	if (!n || !r || r.nodeType !== Node.TEXT_NODE) return null;
	let i = r.textContent ?? "", a = n.startOffset, o = n.startOffset;
	for (; a > 0 && K_.test(i[a - 1]);) a--;
	for (; o < i.length && K_.test(i[o]);) o++;
	if (a === o) return null;
	let s = document.createRange();
	return s.setStart(r, a), s.setEnd(r, o), s;
}
function J_(e, t) {
	return e ? {
		...e.previewLines === void 0 ? {} : { previewLines: e.previewLines },
		...e.detailTabs ? { detailTabs: e.detailTabs } : {},
		...e.onMaximizedChange ? { onMaximizedChange: e.onMaximizedChange } : {},
		...t && e.createActions ? { createActions: e.createActions } : {},
		...t && e.extraAssignItems ? { extraAssignItems: e.extraAssignItems } : {}
	} : {};
}
var Y_ = [{
	id: "inline",
	label: "Inline comments"
}, {
	id: "margin",
	label: "Comment sidebar"
}], X_ = [{
	id: "line",
	label: "Align to line"
}, {
	id: "stack",
	label: "Stacked list"
}], Z_ = 48;
function Q_({ sidebarWidth: e, onSidebarWidthChange: t, resolveUrl: n, path: r, source: i, imageUrl: s, view: l, onViewChange: u, findings: d, problemFilters: f, onProblemFiltersChange: h, comments: g, mode: _ = "text", selectToComment: v = !1, authorFilter: b = !1, layout: x = "margin", onLayoutChange: C, onSourceChange: T, blame: E, line: ee, onLineChange: ne, getFindingActions: D, tabs: O = I_, toolbar: re, preview: k, className: ie }) {
	let ae = d !== void 0, A = d ?? N_, le = B_(r);
	if (i === void 0 && !s) throw Error(`FileReview: ${r} needs \`source\` unless \`imageUrl\` is set`);
	if (s && le) throw Error(`FileReview: \`imageUrl\` is for image files, got markdown ${r}`);
	let ue = i !== void 0, de = i ?? "";
	for (let e of O) if (Vg(e.id)) throw Error(`FileReview: host tab id ${JSON.stringify(e.id)} collides with a built-in tab`);
	let fe = O.find((e) => e.id === l);
	if (!fe && !Vg(l)) throw Error(`FileReview: view ${JSON.stringify(l)} is neither a built-in tab nor one of \`tabs\` (${O.map((e) => e.id).join(", ") || "none"})`);
	if ((l === "source" || l === "problems") && !ue) throw Error(`FileReview: view ${JSON.stringify(l)} needs \`source\`, ${r} is shown as an image only`);
	if (l === "problems" && !ae) throw Error("FileReview: view \"problems\" needs `findings`");
	if (l === "problems" && (!f || !h)) throw Error("FileReview: view \"problems\" needs `problemFilters` and `onProblemFiltersChange`");
	if (l === "preview" && !le && !s && !k) throw Error(`FileReview: view "preview" needs a markdown file, got ${r}`);
	let [pe, me] = (0, N.useState)(!1), [he, ge] = (0, N.useState)(!0), [ye, xe] = (0, N.useState)(!0), Se = pe && E !== void 0, Ce = T !== void 0 && ye && !Se, we = g?.anchoring === "patch", Te = g?.targets, Ee = g?.items ?? P_, De = g?.callbacks ?? F_, Oe = g?.config ?? D_, ke = (0, N.useMemo)(() => ({
		...Oe,
		statuses: [...Oe.statuses, ...zm.statuses],
		facets: [...Oe.facets ?? [], ...(zm.facets ?? []).filter((e) => e.key !== "severity")]
	}), [Oe]), Ae = v && !!De.onCreate, je = (0, N.useMemo)(() => r.endsWith(".mdx") ? Jm(de) : de, [r, de]), Me = (0, N.useMemo)(() => je.split("\n"), [je]), Ne = (0, N.useMemo)(() => Se ? lg(E, Me.length) : null, [
		Se,
		E,
		Me.length
	]), Pe = oe(), Fe = l === "source" && !Ce || l === "problems", [Ie, Le] = (0, N.useState)(null);
	(0, N.useEffect)(() => {
		if (!Fe) return;
		let e = !0, t = Pe === "dark" ? "github-dark" : "github-light";
		return m(je, {
			lang: S(r),
			theme: t
		}).then((t) => e && Le({
			source: je,
			lines: t
		})), () => {
			e = !1;
		};
	}, [
		Fe,
		je,
		r,
		Pe
	]);
	let Re = Ie?.source === je ? Ie.lines : null, [ze, Be] = yt(`file:${r}:align`, "line"), [Ve, j] = yt(`file:${r}:focused`, null), [He, Ue] = (0, N.useState)(null), [M, We] = (0, N.useState)(null), [Ge, Je] = (0, N.useState)(!1), Ye = (0, N.useRef)(null), Xe = (0, N.useRef)(null), [Ze, Qe] = (0, N.useState)(null), $e = vt(), [et, tt] = yt(`file:${r}:selection`, null);
	(0, N.useEffect)(() => {
		if (M?.expanded && M.surface !== "edit") {
			let { position: e, ...t } = M;
			tt(t);
		}
	}, [M, tt]);
	let [nt, rt] = yt(`file:${r}:filter`, "all"), [at, ot] = (0, N.useState)(void 0), [st, ct] = (0, N.useState)(null), lt = (0, N.useRef)(null), dt = (0, N.useRef)(null), ft = (0, N.useRef)(null), pt = (0, N.useRef)(null), mt = (0, N.useRef)(null), F = (0, N.useRef)(null), ht = Gh(pt), { heights: gt, refFor: _t } = Uh(), bt = (0, N.useMemo)(() => Hm(A), [A]), xt = (0, N.useMemo)(() => [...bt, ...Ee], [bt, Ee]), St = (0, N.useMemo)(() => new Map(bt.map((e, t) => [e.id, A[t]])), [bt, A]), Ct = (0, N.useMemo)(() => new Map(A.map((e, t) => [e, bt[t]])), [bt, A]), [wt, Tt] = (0, N.useState)(null), Et = !!De.onCreate && (ke.assignees?.length ?? 0) > 0;
	function Dt(e) {
		let t = Me[e.line - 1];
		if (t === void 0) throw Error(`line ${e.line} is outside ${r} (${Me.length} lines)`);
		let n = t.length - t.trimStart().length, i = t.trimEnd().length, a = e.column ? e.column - 1 : n, o = (e.endLine ?? e.line) === e.line && e.endColumn ? e.endColumn - 1 : i, [s, c] = o > a ? [a, o] : [n, i];
		if (le) {
			let t = F.current && Mh(F.current, Me, {
				line: e.line,
				start: s,
				end: c
			});
			if (!t) throw Error(`line ${e.line} of ${r} has no rendered text to anchor a thread on`);
			return uh(t);
		}
		let l = Dg(je, {
			line: e.line,
			column: s
		}, {
			line: e.line,
			column: c
		});
		if (!l) throw Error(`line ${e.line} of ${r} is blank; there is no text to anchor a thread on`);
		return Sg(l);
	}
	let Ot = (0, N.useMemo)(() => {
		let e = De.onCreate;
		if (!D && !(Et && e)) return;
		let t = (e, t) => {
			Tt(null), Promise.resolve().then(t).catch((t) => Tt(`Could not ${e}: ${t instanceof Error ? t.message : String(t)}`));
		};
		return (n) => [...Et && e ? sm({
			comment: Ct.get(n),
			config: ke,
			onAssign: async (t) => {
				if (!t) throw Error("a finding has no assignee to clear");
				await e({
					body: Ct.get(n).body,
					anchor: Dt(n),
					assignee: t
				});
			},
			run: t
		}) : [], ...D?.(n) ?? []];
	}, [
		D,
		Et,
		De.onCreate,
		Ct,
		ke,
		je,
		Me,
		le,
		r
	]), kt = g?.renderRootDetails, At = (0, N.useMemo)(() => {
		if (Ot || kt) return (e) => {
			let t = St.get(e.id);
			return t ? Ot ? /* @__PURE__ */ (0, P.jsx)(Ke, { items: Ot(t) }) : null : kt?.(e) ?? null;
		};
	}, [
		Ot,
		kt,
		St
	]), jt = (0, N.useMemo)(() => J_(g, !0), [g]), Mt = (0, N.useMemo)(() => J_(g, !1), [g]), Nt = (0, N.useMemo)(() => xt.filter((e) => e.parentId || nt === "all" || (e.author?.kind ?? "user") === nt), [xt, nt]), I = (0, N.useMemo)(() => Ti(Nt, ke, "active"), [Nt, ke]), Pt = pi(Nt).filter((e) => Ci(ke, e.status) !== "active").length, Ft = !!Te?.document, It = (0, N.useMemo)(() => [...new Set(pi(I).map((e) => e.anchor || (Ft ? "__document__" : "")))].filter(Boolean), [I, Ft]), Lt = M?.expanded ? M.anchor : et?.anchor ?? null, Rt = Lt && !It.includes(Lt) ? Lt : null, zt = (0, N.useMemo)(() => Rt ? [...It, Rt] : It, [It, Rt]), L = (0, N.useMemo)(() => Zg(de, zt), [de, zt]), Bt = M?.expanded ? M : et, Vt = Bt?.quote, Ht = Bt?.cell, Ut = (0, N.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of pi(I)) t.anchor && (t.quote || t.cell) && e.set(t.anchor, Qg(t));
		return Rt && (Vt || Ht) && e.set(Rt, Qg({
			quote: Vt,
			cell: Ht
		})), e;
	}, [
		I,
		Rt,
		Vt,
		Ht
	]), Wt = (0, N.useMemo)(() => $g(L, Ut), [L, Ut]), Gt = Kh({
		markdown: F,
		anchors: zt
	}), Kt = (0, N.useMemo)(() => new Map(It.map((e) => [e, G_({
		anchor: e,
		placements: L,
		elements: Gt
	})])), [
		It,
		L,
		Gt
	]), qt = (0, N.useMemo)(() => new Set(It.filter((e) => e === "__document__" || Kt.get(e))), [It, Kt]), Jt = (0, N.useMemo)(() => zt.filter((e) => !qt.has(e)), [zt, qt]), Yt = (0, N.useMemo)(() => It.filter((e) => !qt.has(e)), [It, qt]), Xt = (0, N.useMemo)(() => It.filter((e) => qt.has(e)).map((e) => {
		let t = Kt.get(e);
		return {
			anchor: e,
			label: W_({
				anchor: e,
				found: t !== "orphaned",
				placements: L,
				extras: Ut,
				raw: de
			}),
			...t ? { drift: t } : {},
			note: t ? Gr(t, Zr(e) ? "lines" : "element") : null,
			thread: hi(I, e)
		};
	}), [
		It,
		qt,
		Kt,
		L,
		Ut,
		de,
		I
	]), Zt = (0, N.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of pi(I)) {
			if (!t.anchor || !R_(t.anchor)) continue;
			let n = Yh(t.status), r = e.get(t.anchor);
			(!r || qh[n] > qh[r]) && e.set(t.anchor, n);
		}
		return e;
	}, [I]), { boxes: R, height: Qt, width: $t } = Wh({
		origin: mt,
		markdown: F,
		anchors: Jt,
		mode: _,
		patchRanges: Wt,
		visible: l === "preview"
	});
	(0, N.useEffect)(() => {
		if (M || !et || et.surface !== "preview") return;
		let e = R.get(et.anchor)?.[0];
		e && We({
			...et,
			position: {
				top: e.top,
				bottom: e.top + e.height,
				left: e.left + e.width / 2
			}
		});
	}, [
		R,
		M,
		et
	]);
	let en = (0, N.useRef)(null), [tn, nn] = (0, N.useState)(0), [rn, an] = (0, N.useState)([]);
	(0, N.useEffect)(() => {
		if (M || !et || et.surface !== "source" || !en.current) return;
		let e = [...en.current.querySelectorAll("[data-source-comment]")].find((e) => mg(e).includes(et.anchor));
		if (!e) return;
		let t = e.getBoundingClientRect(), n = en.current.getBoundingClientRect();
		We({
			...et,
			position: {
				top: t.top - n.top,
				bottom: t.bottom - n.top,
				left: t.left - n.left + t.width / 2
			}
		});
	}, [
		M,
		et,
		rn
	]), (0, N.useLayoutEffect)(() => {
		if (l !== "source" || !en.current) return;
		let e = F.current, t = () => an([...Jt.flatMap((t) => t.startsWith("s:") ? [{
			anchor: t,
			spans: Og(je, Cg(t))
		}] : e && t.startsWith("q:") ? [{
			anchor: t,
			spans: Nh(e, Me, dh(t))
		}] : []).filter((e) => e.spans.length > 0), ...e_(de, L)]);
		t();
		let n = new MutationObserver(t);
		e && n.observe(e, {
			childList: !0,
			subtree: !0,
			characterData: !0
		});
		let r = en.current, i = new ResizeObserver(() => nn(r.offsetWidth));
		return i.observe(r), () => {
			n.disconnect(), i.disconnect();
		};
	}, [
		l,
		Jt,
		je,
		Me,
		Ce,
		de,
		L
	]);
	let on = (0, N.useRef)(null), z = (0, N.useRef)(null), B = Ve ?? He;
	(0, N.useEffect)(() => () => clearTimeout(on.current ?? void 0), []);
	let [sn, cn] = (0, N.useState)(null);
	(0, N.useLayoutEffect)(() => {
		let e = en.current;
		if (l !== "source" || !e || !B || !V_(B)) return cn(null);
		let t = e.getBoundingClientRect(), n = [...e.querySelectorAll("[data-source-comment]")].filter((e) => mg(e).includes(B)).flatMap((e) => [...e.getClientRects()]).map((e) => ({
			top: e.top - t.top,
			left: e.left - t.left,
			width: e.width,
			height: e.height
		}));
		cn(n.length > 0 ? {
			anchor: B,
			boxes: n
		} : null);
	}, [
		l,
		B,
		rn,
		tn
	]);
	let ln = st ? R.get(st)?.[0] : void 0;
	(0, N.useEffect)(() => {
		let e = (ht ? lt : dt).current;
		if (l !== "preview" || !ln || !e || !mt.current) return;
		let t = mt.current.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop + ln.top;
		t >= e.scrollTop && t + ln.height <= e.scrollTop + e.clientHeight || (e.scrollTop = t - Z_), ct(null);
	}, [
		l,
		ln,
		ht
	]), (0, N.useEffect)(() => {
		let e = lt.current;
		if (l !== "source" || !st || !e) return;
		let t = L.get(st)?.match, n = R_(st) ? st.slice(L_.length) : t?.state === "matched" ? String(t.startLine) : null, r = n ? e.querySelector(`[data-source-line="${n}"]`) : null;
		if (r) {
			let t = r.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
			t >= e.scrollTop && t + r.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = t - Z_);
		}
		ct(null);
	}, [
		l,
		st,
		L
	]);
	let un = x === "margin" && !ht, dn = () => {
		un && ze === "line" && ft.current && dt.current && (ft.current.scrollTop = dt.current.scrollTop);
	};
	(0, N.useEffect)(dn, [
		Ve,
		ze,
		un
	]), (0, N.useEffect)(() => {
		let e = ft.current;
		if (!un || ze !== "stack" || !Ve || !e) return;
		let t = [...e.querySelectorAll("[data-card-anchor]")].find((e) => e.dataset.cardAnchor === Ve);
		if (!t) return;
		let n = t.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
		n >= e.scrollTop && n + t.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = n - Z_);
	}, [
		Ve,
		ze,
		un
	]);
	function fn(e) {
		j(e), ct(e), R_(e) && ne?.(Number(e.slice(L_.length)));
	}
	function pn(e) {
		let t = Nm(e.line);
		u(le ? "preview" : "source"), j(t), ct(t), le || ot(e.line), ne?.(e.line);
	}
	(0, N.useEffect)(() => {
		if (ee === void 0) return;
		let e = Nm(ee);
		j(e), ct(e), ot(ee);
	}, [ee]);
	let mn = Ge && !!De.onCreate, hn = !!g && le && x === "margin" && l === "preview", gn = Ce && l === "source" && we, _n = hn ? [...mn ? [k_] : [], ...Xt.map((e) => e.anchor)] : [], vn = [..._n, ...Yt], yn = [...R.keys()], bn = Vh({
		cards: [..._n.map((e, t) => ({
			anchor: e,
			desired: 0,
			order: t - _n.length,
			height: gt[e] ?? O_
		})), ...Yt.map((e) => ({
			anchor: e,
			desired: R.get(e)?.[0]?.top ?? Qt,
			order: R.has(e) ? yn.indexOf(e) : yn.length,
			height: gt[e] ?? O_
		}))],
		focused: Ve
	}), xn = Math.max(Qt, ...vn.map((e) => (bn.get(e) ?? 0) + (gt[e] ?? O_))), Sn = [...vn].sort((e, t) => (bn.get(e) ?? 0) - (bn.get(t) ?? 0));
	function Cn(e, t) {
		let n = e.currentTarget.getBoundingClientRect(), r = e.clientX - n.left, i = e.clientY - n.top, a = [...R].filter(([e, n]) => {
			if (e === Rt) return !1;
			if (R_(e)) {
				let e = n[n.length - 1];
				return r <= M_ && i >= n[0].top && i <= e.top + e.height || t === "click" && n.some((e) => i >= e.top && i <= e.top + e.height && r >= e.left && r <= e.left + e.width);
			}
			return n.some((e) => {
				if (i < e.top || i > e.top + e.height) return !1;
				let n = _ === "text";
				return t === "click" || n ? r >= e.left - (n ? 0 : 12) && r <= e.left + e.width : r >= e.left - M_ && r <= e.left;
			});
		}).map(([e]) => e);
		return a.find((e) => e.startsWith("q:")) ?? a[0] ?? null;
	}
	let wn = (e) => e.target.closest("[data-inline-thread], [data-comment-draft], [data-source-thread]") !== null;
	function Tn(e, t) {
		if (we) {
			let n = e === "preview" ? F.current : null, r = n && rh(t, n), i = r ? {
				startLine: r.line,
				endLine: r.line
			} : e === "preview" ? n && T_(t, n) : E_(t);
			return i ? n_({
				raw: de,
				lines: i,
				quote: n && Ch(n, t),
				cell: r?.cell
			}) : null;
		}
		if (!le) {
			let e = gg(t), n = e && Dg(je, e.start, e.end);
			return n ? {
				anchor: Sg(n),
				label: kg(n)
			} : null;
		}
		if (!F.current) return null;
		if (e === "preview") {
			let e = Sh(F.current, t);
			return e && H_(e);
		}
		let n = _g(t), r = n && Mh(F.current, Me, n);
		return r && H_(r);
	}
	let En = (e) => (e === "preview" ? mt : en).current;
	function Dn(e) {
		!window.getSelection()?.isCollapsed || wn(e) || e.target.closest("button, a") || j(Cn(e, "click"));
	}
	function On(e) {
		e !== z.current && (z.current = e, clearTimeout(on.current ?? void 0), on.current = setTimeout(() => Ue(e), e ? A_ : j_));
	}
	function kn(e) {
		clearTimeout(on.current ?? void 0), z.current = e;
	}
	function An() {
		clearTimeout(on.current ?? void 0), z.current = null, j(null), Ue(null);
	}
	function jn(e) {
		if (wn(e)) return;
		let t = e.buttons !== 0 || !window.getSelection()?.isCollapsed, n = e.target.closest("[data-source-comment]");
		On(t || !n ? null : mg(n)[0]);
	}
	function Mn(e) {
		wn(e) || e.target.closest("[data-source-comment], button, a") || Ve && V_(Ve) && window.getSelection()?.isCollapsed && An();
	}
	function Nn(e) {
		x !== "inline" || wn(e) || On(e.buttons !== 0 || !window.getSelection()?.isCollapsed ? null : Cn(e, "hover"));
	}
	function Pn(e, t) {
		let n = Xe.current?.kind === "text";
		if (!(Ae || n) || wn(e) || e.target.closest("[data-comment-author]")) return;
		let r = window.getSelection(), i = r && !r.isCollapsed && r.rangeCount > 0 ? r.getRangeAt(0) : null;
		if (n) {
			let e = i && Tn(t, i);
			if (!e) return;
			r?.removeAllRanges(), Gn({
				anchor: e.anchor,
				...e.quote ? { quote: e.quote } : {}
			});
			return;
		}
		let a = i ? Tn(t, i) : null;
		if (!i || !a) {
			We((e) => e?.expanded ? e : null);
			return;
		}
		let o = En(t).getBoundingClientRect(), s = i.getBoundingClientRect();
		We({
			...a,
			position: {
				top: s.top - o.top,
				bottom: s.bottom - o.top,
				left: s.left - o.left + s.width / 2
			},
			expanded: !1,
			surface: t
		});
	}
	function Fn(e, t) {
		if (!Ae || wn(e) || e.target.closest("[data-comment-author], button, a")) return;
		let n = window.getSelection(), r = n && !n.isCollapsed && n.rangeCount > 0 ? n.getRangeAt(0) : null, i = r && [...r.getClientRects()].some((t) => e.clientX >= t.left && e.clientX <= t.right && e.clientY >= t.top && e.clientY <= t.bottom), a = i ? r : q_(e.clientX, e.clientY), o = a ? Tn(t, a) : null;
		if (!a || !o) return;
		e.preventDefault(), i || (n?.removeAllRanges(), n?.addRange(a));
		let s = En(t).getBoundingClientRect(), c = a.getBoundingClientRect();
		We({
			...o,
			position: {
				top: c.top - s.top,
				bottom: c.bottom - s.top,
				left: e.clientX - s.left
			},
			expanded: !1,
			surface: t
		});
	}
	function In() {
		We((e) => e && {
			...e,
			expanded: !0
		}), window.getSelection()?.removeAllRanges();
	}
	let Ln = () => {
		M && $e(M.anchor), Ye.current = null, We(null), tt(null);
	};
	function Rn(e) {
		let t = e.expanded && e.surface === "preview" ? R.get(e.anchor) : void 0, n = t?.[0], r = t?.at(-1);
		return !n || !r ? e.position : {
			top: n.top,
			bottom: r.top + r.height,
			left: n.left + n.width / 2
		};
	}
	let zn = jt.createActions, Bn = (0, N.useMemo)(() => x_({
		host: zn,
		screenshot: !!Te?.screenshot
	}), [zn, Te?.screenshot]);
	function Vn(e) {
		let t = Ye.current;
		if (t?.isConnected && M?.anchor === e) return t;
		let n = mt.current, r = l === "preview" ? R.get(e) : void 0;
		if (n && r && r.length > 0) {
			let e = n.getBoundingClientRect(), t = Math.min(...r.map((e) => e.left)), i = Math.min(...r.map((e) => e.top)), a = Math.max(...r.map((e) => e.left + e.width)), o = Math.max(...r.map((e) => e.top + e.height));
			return {
				rect: {
					x: e.left + t,
					y: e.top + i,
					width: a - t,
					height: o - i
				},
				isLive: () => n.isConnected
			};
		}
		if (!lt.current) throw Error("FileReview: there is nothing mounted to capture");
		return lt.current;
	}
	async function Hn(e, t, n) {
		let r = De.onCreate;
		if (!r) throw Error("FileReview: comment creation is not configured");
		await S_({
			onCreate: r,
			input: t,
			action: n,
			hostActions: zn,
			region: () => Vn(e)
		});
	}
	async function Un(e, t) {
		await Hn(oi, e, t), Je(!1);
	}
	let Wn = ai({
		root: mt,
		accepts: (e) => F.current?.contains(e) ?? !1,
		onPick: (e) => {
			let t = F.current;
			if (!t || !mt.current) return;
			let n = Ar(e, t);
			if (n.length === 0) return Gn(null), Tt("Could not use that element: it cannot be told apart from the rest of the page");
			if (Xe.current) return Gn({ anchor: Mr(n) });
			let r = mt.current.getBoundingClientRect(), i = e.getBoundingClientRect();
			Ye.current = e, Je(!1), We({
				anchor: Mr(n),
				label: ii(e),
				position: {
					top: i.top - r.top,
					bottom: i.bottom - r.top,
					left: i.left - r.left + i.width / 2
				},
				expanded: !0,
				surface: "preview"
			});
		}
	});
	function Gn(e) {
		let t = Xe.current;
		t && (Xe.current = null, Qe(null), Wn.setPicking(!1), t.resolve(e));
	}
	function Kn(e) {
		Xe.current?.resolve(null);
		let t = e.startsWith("e:") ? "element" : "text";
		return new Promise((e) => {
			Xe.current = {
				kind: t,
				resolve: e
			}, Qe(t), We(null), Je(!1), t === "element" && (u("preview"), Wn.setPicking(!0));
		});
	}
	let qn = g?.onRetarget ? {
		pickAnchor: Kn,
		onRetarget: g.onRetarget
	} : void 0;
	y(Ze === "text", () => Gn(null)), (0, N.useEffect)(() => {
		Ze === "element" && !Wn.picking && Gn(null);
	}, [Ze, Wn.picking]), (0, N.useEffect)(() => () => Xe.current?.resolve(null), []);
	let Jn = [...Te?.element && le ? [{
		label: "Comment on element",
		onSelect: () => {
			u("preview"), Wn.setPicking(!0);
		}
	}] : [], ...Te?.document && ue ? [{
		label: "Comment on file",
		onSelect: () => {
			We(null), Je(!0);
		}
	}] : []], Yn = (0, N.useMemo)(() => Yt.flatMap((e) => {
		let t = L.get(e);
		return t?.match.state === "matched" ? [{
			anchor: e,
			label: t_(t),
			startLine: t.match.startLine,
			thread: hi(I, e)
		}] : [];
	}).sort((e, t) => e.startLine - t.startLine), [
		Yt,
		L,
		I
	]), Xn = (0, N.useMemo)(() => {
		if (we && Ce) return {
			ranges: Yn.map((e) => {
				let t = L.get(e.anchor).match;
				if (t.state !== "matched") throw Error(`FileReview: edit thread ${e.anchor} is not placed`);
				return {
					id: e.anchor,
					startLine: t.startLine,
					endLine: t.endLine
				};
			}),
			onSelect: j,
			...Ae ? { onCreate: (e) => We({
				...n_({
					raw: de,
					lines: e
				}),
				position: z_,
				expanded: !0,
				surface: "edit"
			}) } : {}
		};
	}, [
		we,
		Ce,
		Yn,
		L,
		Ae,
		de,
		j
	]);
	function Zn(e, t) {
		if (!De.onCreate) return null;
		let { anchor: n, label: r } = e;
		return /* @__PURE__ */ (0, P.jsx)(zg, {
			anchor: n,
			label: r,
			position: Rn(e),
			containerWidth: t,
			expanded: e.expanded,
			config: ke,
			onExpand: In,
			onClose: Ln,
			...Bn ? { createActions: Bn } : {},
			onCreate: async (t, r) => {
				await Hn(n, {
					...t,
					...Qg(e)
				}, r), We(null), tt(null), j(n);
			}
		}, n);
	}
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		ref: pt,
		"data-file-review": r,
		className: a("flex min-h-0 flex-col gap-2", ie ?? "h-[600px]"),
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex shrink-0 flex-wrap items-center gap-2 text-xs text-muted-foreground",
				children: [
					re,
					wt && /* @__PURE__ */ (0, P.jsx)("span", {
						role: "alert",
						className: "text-destructive",
						children: wt
					}),
					(g || ae) && /* @__PURE__ */ (0, P.jsxs)("span", { children: [
						pi(I).length,
						" open · ",
						Pt,
						" resolved (hidden)"
					] }),
					g && x === "inline" && /* @__PURE__ */ (0, P.jsx)("span", { children: "· hover a highlight to view or reply" }),
					Ae && /* @__PURE__ */ (0, P.jsx)("span", { children: "· select or right-click text to comment" }),
					Wn.picking && !Ze && /* @__PURE__ */ (0, P.jsx)("span", {
						role: "status",
						children: "· click an element to comment on it, Esc to cancel"
					}),
					Ze && /* @__PURE__ */ (0, P.jsxs)("span", {
						role: "status",
						children: [
							"· ",
							Ze === "element" ? "click an element" : "select text",
							" to re-anchor, Esc to cancel"
						]
					}),
					b && /* @__PURE__ */ (0, P.jsx)("div", {
						role: "radiogroup",
						"aria-label": "Authors",
						className: "ml-auto inline-flex rounded-md border border-border p-0.5",
						children: [
							"all",
							"user",
							"agent"
						].map((e) => /* @__PURE__ */ (0, P.jsx)("button", {
							type: "button",
							role: "radio",
							"aria-checked": nt === e,
							onClick: () => rt(e),
							className: a("rounded px-2 py-0.5 font-medium", nt === e ? "bg-primary text-primary-foreground" : "hover:text-foreground"),
							children: e === "all" ? "All" : e === "user" ? "People" : "Tools"
						}, e))
					})
				]
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex shrink-0 items-center border-b border-border",
				children: [
					/* @__PURE__ */ (0, P.jsx)(qe, {
						className: "min-w-0 flex-1 border-b-0",
						value: l,
						onChange: (e) => {
							u(e), We(null);
						},
						tabs: [
							...le || s || k ? [{
								id: "preview",
								label: "Preview",
								icon: ve
							}] : [],
							...ue ? [{
								id: "source",
								label: "Source",
								icon: _e
							}] : [],
							...ae && ue ? [{
								id: "problems",
								label: "Problems",
								icon: se,
								count: A.length
							}] : [],
							...O.map(({ render: e, ...t }) => t)
						]
					}),
					T && ue && l === "source" && /* @__PURE__ */ (0, P.jsx)("div", {
						role: "group",
						"aria-label": "Source mode",
						className: "mr-1 inline-flex shrink-0 rounded-md border border-border p-0.5 text-xs",
						children: ["edit", "review"].map((e) => /* @__PURE__ */ (0, P.jsx)("button", {
							type: "button",
							"aria-pressed": Ce === (e === "edit"),
							title: e === "edit" ? "Edit the source in the editor" : "Read the source and its comment threads",
							onClick: () => {
								xe(e === "edit"), e === "edit" && me(!1);
							},
							className: a("rounded px-2 py-0.5 font-medium", Ce === (e === "edit") ? "bg-primary text-primary-foreground" : "hover:text-foreground"),
							children: e === "edit" ? "Edit" : "Review"
						}, e))
					}),
					De.onCreate && Jn.length > 0 && /* @__PURE__ */ (0, P.jsx)(te, {
						align: "right",
						icon: ce,
						label: "Comment",
						hideChevron: !0,
						variant: "ghost",
						size: "sm",
						title: "Comment on an element or the whole file",
						menuLabel: "Comment",
						className: "shrink-0",
						items: Jn
					}),
					ae && A.length > 0 && l === "source" && !Ce && /* @__PURE__ */ (0, P.jsxs)(o, {
						size: "sm",
						variant: he ? "secondary" : "ghost",
						"aria-pressed": he,
						title: he ? "Show only the gutter marks; pick one to read its findings" : "Show each line's findings underneath it",
						onClick: () => ge((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, P.jsx)(c, { icon: se }), "Findings"]
					}),
					E && l === "source" && /* @__PURE__ */ (0, P.jsxs)(o, {
						size: "sm",
						variant: pe ? "secondary" : "ghost",
						"aria-pressed": pe,
						title: pe ? "Hide who last changed each line" : "Show who last changed each line",
						onClick: () => me((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, P.jsx)(c, { icon: it }), "Blame"]
					}),
					g && C && le && l !== "problems" && /* @__PURE__ */ (0, P.jsx)(te, {
						align: "right",
						icon: p,
						hideChevron: !0,
						variant: "ghost",
						size: "sm",
						title: "View options",
						menuLabel: "View options",
						className: "shrink-0",
						items: [...Y_.map((e) => ({
							label: e.label,
							group: "Comments",
							icon: w,
							iconClassName: e.id === x ? void 0 : "invisible",
							onSelect: () => C(e.id)
						})), ...x === "margin" ? X_.map((e) => ({
							label: e.label,
							group: "Sidebar cards",
							icon: w,
							iconClassName: e.id === ze ? void 0 : "invisible",
							onSelect: () => Be(e.id)
						})) : []]
					})
				]
			}),
			g && !hn && !gn && /* @__PURE__ */ (0, P.jsx)(qg, {
				entries: Xt,
				drafting: mn,
				config: ke,
				callbacks: De,
				options: jt,
				renderRootDetails: At,
				composerActions: Bn,
				retarget: qn,
				onCreate: Un,
				onDiscardDraft: () => Je(!1)
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				ref: lt,
				className: a("relative min-h-0 flex-1 rounded-md border border-border bg-background", Ce && l === "source" ? "overflow-hidden" : "overflow-y-auto"),
				children: [
					fe && /* @__PURE__ */ (0, P.jsx)("div", {
						"data-file-review-tab": fe.id,
						className: "h-full",
						children: fe.render({
							path: r,
							source: de,
							findings: A,
							...T ? { onSourceChange: T } : {}
						})
					}),
					Ce && l === "source" && /* @__PURE__ */ (0, P.jsxs)("div", {
						className: "flex h-full min-h-0",
						children: [/* @__PURE__ */ (0, P.jsx)("div", {
							className: "h-full min-w-0 flex-1",
							children: /* @__PURE__ */ (0, P.jsx)(Wm, {
								path: r,
								source: de,
								findings: A,
								line: at,
								onLineChange: ot,
								onChange: T,
								inlineFindings: !0,
								sourceComments: Xn
							})
						}), we && /* @__PURE__ */ (0, P.jsx)(Xg, {
							entries: Yn,
							draft: M?.surface === "edit" ? M : null,
							focused: Ve,
							config: ke,
							callbacks: De,
							options: jt,
							renderRootDetails: At,
							composerActions: Bn,
							onFocus: (e) => {
								j(e.anchor), ot(e.startLine);
							},
							onCreate: async (e, t) => {
								if (!M) throw Error("FileReview: no edit-mode comment is being written");
								await Hn(M.anchor, e, t), j(M.anchor), We(null);
							},
							onDiscardDraft: Ln,
							files: {
								entries: Xt,
								drafting: mn,
								onCreate: Un,
								onDiscardDraft: () => Je(!1)
							}
						})]
					}),
					!Ce && l === "source" && /* @__PURE__ */ (0, P.jsxs)("div", {
						ref: en,
						className: "relative",
						onMouseUp: (e) => e.button === 0 && Pn(e, "source"),
						onContextMenu: (e) => Fn(e, "source"),
						onClick: Mn,
						onMouseMove: jn,
						onMouseLeave: () => On(null),
						children: [
							/* @__PURE__ */ (0, P.jsx)(yg, {
								source: je,
								tokens: Re,
								blame: Ne,
								findings: A,
								showFindings: he,
								marks: rn,
								comments: I,
								selected: B ?? Lt,
								onSelect: (e) => {
									j(Nm(e.line)), ne?.(e.line);
								},
								getFindingActions: Ot,
								onFocus: j
							}),
							M?.surface === "source" && Zn(M, tn),
							sn && /* @__PURE__ */ (0, P.jsx)(Mg, {
								seamless: !0,
								anchor: sn.anchor,
								label: W_({
									anchor: sn.anchor,
									found: !0,
									placements: L,
									extras: Ut,
									raw: de
								}),
								boxes: sn.boxes,
								containerWidth: tn,
								thread: hi(I, sn.anchor),
								config: ke,
								callbacks: De,
								renderRootDetails: At,
								detailOptions: jt,
								pinned: Ve === sn.anchor,
								onPin: () => j(sn.anchor),
								onClose: An,
								onPointerEnter: () => kn(sn.anchor),
								onPointerLeave: () => On(null)
							}, sn.anchor)
						]
					}),
					l === "problems" && f && h && /* @__PURE__ */ (0, P.jsx)(sg, {
						source: je,
						tokens: Re,
						findings: A,
						filters: f,
						onFiltersChange: h,
						onOpen: pn,
						openLabel: le ? "Open in preview" : "Open in source",
						getFindingActions: Ot
					}),
					s && l === "preview" && /* @__PURE__ */ (0, P.jsx)(ut, {
						src: s,
						alt: r,
						className: "h-full"
					}),
					k && /* @__PURE__ */ (0, P.jsxs)("div", {
						ref: mt,
						className: a("relative h-full", l !== "preview" && "hidden"),
						"data-file-review-preview": "custom",
						children: [
							k({
								path: r,
								source: de,
								findings: A,
								comments: I,
								focusedAnchor: Ve,
								onSelectElement: (e) => {
									Pr(e.anchor), We({
										...e,
										expanded: !0,
										surface: "preview"
									});
								},
								onFocus: j,
								...T ? { onSourceChange: T } : {}
							}),
							M?.surface === "preview" && Zn(M, mt.current?.clientWidth ?? $t),
							Ve && /* @__PURE__ */ (0, P.jsxs)("div", {
								className: "absolute right-2 top-12 z-30 max-h-[80%] w-80 overflow-auto rounded border bg-card p-2 shadow",
								children: [/* @__PURE__ */ (0, P.jsx)("button", {
									type: "button",
									onClick: () => j(null),
									children: "Close comments"
								}), /* @__PURE__ */ (0, P.jsx)(_m, {
									comments: hi(I, Ve),
									config: ke,
									anchor: Ve,
									...De,
									...jt
								})]
							})
						]
					}),
					le && !k && /* @__PURE__ */ (0, P.jsxs)("div", {
						className: a("flex", ht ? "flex-col gap-4 p-4" : "h-full", l !== "preview" && "hidden"),
						children: [/* @__PURE__ */ (0, P.jsx)("div", {
							ref: dt,
							onScroll: dn,
							className: a("min-w-0 flex-1", !ht && "overflow-y-auto p-4"),
							children: /* @__PURE__ */ (0, P.jsxs)("div", {
								ref: mt,
								className: "relative",
								onMouseUp: (e) => e.button === 0 && Pn(e, "preview"),
								onContextMenu: (e) => Fn(e, "preview"),
								onClick: Dn,
								onMouseMove: Nn,
								onMouseLeave: () => x === "inline" && On(null),
								...Wn.rootProps,
								children: [
									Wn.outline,
									[...R].flatMap(([e, t]) => {
										let n = e === Ve || e === He || e === Lt;
										if (R_(e)) {
											let r = Zt.get(e);
											if (!r) throw Error(`lint anchor ${e} has no open finding`);
											let i = Jh[r], o = t[0];
											return [/* @__PURE__ */ (0, P.jsx)("span", {
												"data-lint-rail": r,
												"aria-hidden": !0,
												className: a("pointer-events-none absolute z-20 flex size-4 items-center justify-center rounded-sm", i.tone, n && i.tint),
												style: {
													top: o.top + o.height / 2 - 8,
													left: 0
												},
												children: /* @__PURE__ */ (0, P.jsx)(c, {
													icon: i.icon,
													className: "text-[13px]"
												})
											}, `${e}:rail`), ...t.map((t, o) => /* @__PURE__ */ (0, P.jsx)("div", {
												"data-anchor-box": e,
												"data-lint-squiggle": r,
												"aria-hidden": !0,
												className: a("pointer-events-none absolute z-0 transition-colors", n && i.tint),
												style: {
													top: t.top,
													left: t.left,
													width: t.width,
													height: t.height + 3
												},
												children: /* @__PURE__ */ (0, P.jsx)("span", {
													className: "absolute inset-x-0 bottom-0 h-[3px]",
													style: {
														backgroundImage: Xh(i.stroke),
														backgroundRepeat: "repeat-x",
														backgroundSize: "6px 3px"
													}
												})
											}, `${e}:${o}`))];
										}
										return t.map((t, r) => /* @__PURE__ */ (0, P.jsx)("div", {
											"data-anchor-box": e,
											"aria-hidden": !0,
											className: a("pointer-events-none absolute z-0 transition-colors", U_({
												mode: _,
												active: n
											})),
											style: _ === "text" ? {
												top: t.top,
												left: t.left,
												width: t.width,
												height: t.height
											} : {
												top: t.top - 2,
												left: t.left - 10,
												width: t.width + 10,
												height: t.height + 4
											}
										}, `${e}:${r}`));
									}),
									/* @__PURE__ */ (0, P.jsx)("div", {
										ref: F,
										"data-annotated-markdown": "",
										className: "relative z-10 pl-4",
										children: /* @__PURE__ */ (0, P.jsx)(be, {
											sourceBlocks: !0,
											text: je,
											resolveUrl: n
										})
									}),
									g && x === "inline" && /* @__PURE__ */ (0, P.jsx)(Ng, {
										markdown: F,
										comments: I,
										focused: Ve,
										onFocus: j
									}),
									M?.surface === "preview" && Zn(M, $t),
									x === "inline" && B && /* @__PURE__ */ (0, P.jsx)(Mg, {
										anchor: B,
										label: W_({
											anchor: B,
											found: !0,
											placements: L,
											extras: Ut,
											raw: de
										}),
										boxes: R.get(B) ?? [],
										containerWidth: $t,
										thread: hi(I, B),
										config: ke,
										callbacks: R_(B) ? F_ : De,
										renderRootDetails: At,
										detailOptions: R_(B) ? Mt : jt,
										pinned: Ve === B,
										onPin: () => j(B),
										onClose: An,
										onPointerEnter: () => kn(B),
										onPointerLeave: () => On(null)
									}, B)
								]
							})
						}), x === "margin" && /* @__PURE__ */ (0, P.jsx)(Mm, {
							scrollRef: ft,
							stacked: ht,
							...e === void 0 ? {} : { width: e },
							...t ? { onWidthChange: t } : {},
							children: /* @__PURE__ */ (0, P.jsx)("div", {
								"data-comment-sidebar": ht ? "below" : ze,
								className: a("relative mx-2 my-4", (ht || ze === "stack") && "space-y-2"),
								style: un && ze === "line" ? { height: xn } : void 0,
								children: Sn.map((e) => {
									let t = hi(I, e), n = e === Ve, r = un && ze === "line" ? { top: bn.get(e) ?? 0 } : void 0, i = un && ze === "line" && "absolute inset-x-0";
									if (e === k_) return /* @__PURE__ */ (0, P.jsx)("div", {
										ref: _t(e),
										"data-comment-draft": "__document__",
										style: r,
										className: a("rounded-lg border border-primary/50 bg-card shadow-lg", i),
										children: /* @__PURE__ */ (0, P.jsx)(Jg, {
											config: ke,
											composerActions: Bn,
											onCreate: Un,
											onDiscard: () => Je(!1)
										})
									}, e);
									let o = Xt.find((t) => t.anchor === e);
									return o ? /* @__PURE__ */ (0, P.jsx)("div", {
										ref: _t(e),
										"data-card-anchor": e,
										"data-file-thread": e,
										onClick: () => j(e),
										style: r,
										className: a("rounded-lg border bg-card", i, n ? "z-10 -translate-x-2 border-primary/50 shadow-lg" : "border-border shadow-sm"),
										children: /* @__PURE__ */ (0, P.jsx)(Yg, {
											entry: o,
											config: ke,
											callbacks: De,
											options: jt,
											renderRootDetails: At,
											retarget: qn
										})
									}, `${e}:${n}`) : /* @__PURE__ */ (0, P.jsxs)("div", {
										ref: _t(e),
										"data-card-anchor": e,
										onMouseEnter: () => Ue(e),
										onMouseLeave: () => Ue(null),
										onClick: () => fn(e),
										className: a("rounded-lg border bg-card transition-[top,transform] duration-150", un && ze === "line" && "absolute inset-x-0", n ? "z-10 -translate-x-2 border-primary/50 shadow-lg" : "border-border shadow-sm"),
										style: un && ze === "line" ? { top: bn.get(e) ?? 0 } : void 0,
										children: [/* @__PURE__ */ (0, P.jsx)("div", {
											className: "truncate px-2 pt-2 text-[11px] text-muted-foreground",
											children: W_({
												anchor: e,
												found: R.has(e),
												placements: L,
												extras: Ut,
												raw: de
											})
										}), /* @__PURE__ */ (0, P.jsx)(_m, {
											...R_(e) ? Mt : jt,
											comments: t,
											config: ke,
											anchor: e,
											compact: !0,
											defaultExpanded: n || R_(e),
											hideComposer: t.length > 0,
											autoFocusComposer: t.length === 0,
											composerPlaceholder: "Add a comment…",
											...R_(e) ? F_ : De,
											...At ? { renderRootDetails: At } : {}
										})]
									}, `${e}:${n}`);
								})
							})
						})]
					})
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-urls.ts
var $_ = /\.(png|jpe?g|gif|webp|svg|avif|bmp|ico)$/i, ev = /^[a-z][a-z0-9+.-]*:/i;
function tv(e) {
	return $_.test(e);
}
function nv(e) {
	return tv(e) && !/\.svg$/i.test(e);
}
function rv(e) {
	let t = e.trim();
	if (!t || t.startsWith("/") || t.startsWith("#") || ev.test(t)) return null;
	let n = t.replace(/[?#].*$/s, "");
	if (!n) return null;
	try {
		return decodeURIComponent(n);
	} catch {
		return null;
	}
}
function iv(e) {
	let t = [];
	for (let n of e.split("/")) n && n !== "." && (n === ".." && t.length > 0 && t[t.length - 1] !== ".." ? t.pop() : t.push(n));
	return t.join("/");
}
function av(e, t) {
	let n = rv(t);
	if (n === null) return null;
	let r = e.file.lastIndexOf("/"), i = r < 0 ? "" : e.file.slice(0, r), a = iv(i ? `${i}/${n}` : n);
	return e.workDir === void 0 ? { file: a } : {
		workDir: e.workDir,
		file: a
	};
}
// ../../../clicky-ui/packages/ui/src/file-findings/NodeActions.tsx
function ov({ items: e, label: t, className: n, variant: r = "header" }) {
	return e.length === 0 ? null : /* @__PURE__ */ (0, P.jsx)("span", {
		className: a("inline-flex shrink-0", n),
		"data-node-actions": "",
		onClick: (e) => e.stopPropagation(),
		onKeyDown: (e) => e.stopPropagation(),
		children: r === "row" ? /* @__PURE__ */ (0, P.jsx)(te, {
			align: "right",
			menuLabel: t,
			items: e,
			trigger: /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				title: t,
				"aria-label": t,
				className: "inline-flex h-5 w-5 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
				children: /* @__PURE__ */ (0, P.jsx)(c, {
					icon: le,
					className: "text-sm"
				})
			})
		}) : /* @__PURE__ */ (0, P.jsx)(te, {
			align: "right",
			icon: p,
			hideChevron: !0,
			variant: "ghost",
			size: "sm",
			title: t,
			menuLabel: t,
			items: e
		})
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/SeverityCounts.tsx
function sv({ findings: e }) {
	let t = eg(e);
	return /* @__PURE__ */ (0, P.jsx)("span", {
		className: "flex shrink-0 items-center gap-2 text-xs text-muted-foreground",
		children: $h.filter((e) => t[e] > 0).map((e) => /* @__PURE__ */ (0, P.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			title: `${t[e]} ${e}`,
			children: [/* @__PURE__ */ (0, P.jsx)(Ee, { severity: e }), /* @__PURE__ */ (0, P.jsx)("span", {
				className: "tabular-nums",
				children: t[e]
			})]
		}, e))
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/FolderFindings.tsx
function cv(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		let e = {
			workDir: n.workDir,
			file: n.file
		}, r = j(e), i = t.get(r) ?? {
			ref: e,
			findings: []
		};
		i.findings.push(n), t.set(r, i);
	}
	return [...t.values()].map((e) => ({
		...e,
		findings: e.findings.sort(Qh)
	}));
}
function lv({ findings: e, problemFilters: t, onProblemFiltersChange: n, onOpenFile: r, getFindingActions: i }) {
	let a = (0, N.useMemo)(() => e.filter((e) => rg(e, t)), [e, t]), o = (0, N.useMemo)(() => cv(a), [a]);
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)(og, {
		findings: e,
		shown: a.length,
		filters: t,
		onFiltersChange: n
	}), o.length === 0 ? /* @__PURE__ */ (0, P.jsx)("p", {
		className: "p-6 text-center text-xs text-muted-foreground",
		children: "No problems match the current filters."
	}) : o.map(({ ref: e, findings: t }) => /* @__PURE__ */ (0, P.jsxs)("section", {
		"data-folder-file": e.file,
		children: [/* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			onClick: () => r(e),
			title: `Open ${e.file}`,
			className: "flex w-full min-w-0 items-center gap-2 border-b border-border bg-muted/30 px-2 py-1 text-left text-xs font-medium hover:bg-muted",
			children: [
				/* @__PURE__ */ (0, P.jsx)(c, {
					icon: _e,
					className: "shrink-0 text-muted-foreground"
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 flex-1 truncate font-mono",
					children: e.file
				}),
				/* @__PURE__ */ (0, P.jsx)(sv, { findings: t })
			]
		}), /* @__PURE__ */ (0, P.jsx)("ul", { children: t.map((t, n) => /* @__PURE__ */ (0, P.jsx)("li", {
			"data-problem-line": t.line,
			children: /* @__PURE__ */ (0, P.jsx)(M, {
				message: t.message,
				severity: t.severity,
				metadata: `${t.source} ${t.rule ?? ""}`,
				location: `${t.line}${t.column ? `:${t.column}` : ""}`,
				actions: [
					{
						label: "Open in file",
						onSelect: () => r(e, t.line)
					},
					{
						label: "Copy description",
						onSelect: () => void navigator.clipboard.writeText(t.message)
					},
					...i?.(t) ?? []
				],
				children: null
			})
		}, `${t.line}:${t.column ?? 0}:${n}`)) })]
	}, j(e)))] });
}
function uv({ folder: e, name: t, showPath: n = !0, findings: r, view: i, onViewChange: a, actions: o, tabs: s = [], ...c }) {
	let l = s.find((e) => e.id === i);
	if (!l && i !== "problems") throw Error(`FolderFindings: view ${JSON.stringify(i)} is neither "${Hg}" nor one of \`tabs\` (${s.map((e) => e.id).join(", ") || "none"})`);
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		"data-folder-findings": e.file,
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
				"data-file-header": "",
				children: [
					n && /* @__PURE__ */ (0, P.jsx)("span", {
						className: "min-w-0 truncate font-mono text-sm",
						title: e.workDir ? `${e.workDir}/${e.file}` : e.file,
						children: t
					}),
					/* @__PURE__ */ (0, P.jsx)(sv, { findings: r }),
					/* @__PURE__ */ (0, P.jsx)("span", { className: "flex-1" }),
					/* @__PURE__ */ (0, P.jsx)(ov, {
						items: o ?? [],
						label: "Folder actions"
					})
				]
			}),
			/* @__PURE__ */ (0, P.jsx)(qe, {
				className: "shrink-0 px-density-2",
				value: i,
				onChange: a,
				tabs: [{
					id: Hg,
					label: "Problems",
					icon: se,
					count: r.length
				}, ...s.map(({ render: e, ...t }) => t)]
			}),
			/* @__PURE__ */ (0, P.jsx)("div", {
				className: "min-h-0 flex-1 overflow-y-auto",
				children: l ? l.render({
					folder: e,
					findings: r
				}) : /* @__PURE__ */ (0, P.jsx)(lv, {
					findings: r,
					...c
				})
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/FileFindingsView.tsx
function dv(e) {
	return /\.mdx?$/i.test(e);
}
function fv(e, t, n) {
	return n.some((t) => t.id === e) ? e : t.raster ? "preview" : e === "source" || e === "problems" ? e : t.markdown || t.image ? "preview" : "source";
}
function pv({ file: e, url: t, kind: n, fileUrl: r }) {
	let i = n === "image" ? av(e, t) : null;
	return i ? r(i) : t;
}
function mv({ node: e, selected: t, slots: n, actions: r }) {
	let i = (r, i) => {
		let a = n?.[r];
		return a ? a({
			node: e,
			selected: t,
			default: i
		}) : i;
	}, [o, s] = (0, N.useState)(null), l = e.kind === "folder" ? "Folder actions" : "File actions";
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
		r.length > 0 && /* @__PURE__ */ (0, P.jsx)("span", {
			hidden: !0,
			ref: (e) => s(e?.parentElement ?? null),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			children: /* @__PURE__ */ (0, P.jsx)(Ge, {
				contextTarget: o,
				menuLabel: l,
				menuItems: r
			})
		}),
		i("icon", /* @__PURE__ */ (0, P.jsx)(c, {
			icon: e.kind === "folder" ? ge : tv(e.ref.file) ? Se : _e,
			className: "shrink-0 text-muted-foreground"
		})),
		i("name", /* @__PURE__ */ (0, P.jsx)("span", {
			className: a("min-w-0 truncate text-foreground", t ? "font-semibold" : "font-medium"),
			title: e.name,
			children: e.name
		})),
		/* @__PURE__ */ (0, P.jsx)("span", { className: "flex-1" }),
		i("badges", /* @__PURE__ */ (0, P.jsx)(sv, { findings: e.findings })),
		/* @__PURE__ */ (0, P.jsx)(ov, {
			variant: "row",
			items: r,
			label: l
		})
	] });
}
function hv(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (Ve({
			kind: n.kind,
			ref: n.ref
		}, t)) return n;
		let e = hv(n.children ?? [], t);
		if (e) return e;
	}
	return null;
}
function gv(e, t) {
	let [n, r] = (0, N.useState)(null), i = e ? j(e) : null, a = (0, N.useRef)({
		ref: e,
		loadFile: t
	});
	return (0, N.useEffect)(() => {
		a.current = {
			ref: e,
			loadFile: t
		};
	}), (0, N.useEffect)(() => {
		let { ref: e, loadFile: t } = a.current;
		if (!e || !i) {
			r(null);
			return;
		}
		let n = !0;
		return r({
			key: i,
			status: "loading"
		}), t(e).then((e) => n && r({
			key: i,
			status: "ready",
			text: e
		}), (e) => n && r({
			key: i,
			status: "error",
			message: e instanceof Error ? e.message : String(e)
		})), () => {
			n = !1;
		};
	}, [i]), {
		state: n?.key === i ? n : null,
		replaceText: (e) => r((t) => t?.key === i && t.status === "ready" ? {
			...t,
			text: e
		} : t)
	};
}
function _v({ children: e }) {
	return /* @__PURE__ */ (0, P.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
		children: e
	});
}
function vv({ file: e, showPath: t, findings: n, state: r, view: i, tabs: a, actions: o, toolbar: s, preview: l, comments: u, imageUrl: d, raster: f, resolveUrl: p, onSourceChange: m, selectToComment: h, authorFilter: g, layout: _, onLayoutChange: v, ...y }) {
	let b = r?.status === "ready" ? r : null, x;
	return x = !f && !b ? r?.status === "error" ? /* @__PURE__ */ (0, P.jsxs)("div", {
		role: "alert",
		className: "m-density-4 rounded-md border border-destructive/40 p-3 text-sm text-destructive",
		children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			className: "font-semibold",
			children: ["Could not load ", e.file]
		}), /* @__PURE__ */ (0, P.jsx)("pre", {
			className: "mt-1 whitespace-pre-wrap break-words font-mono text-xs",
			children: r.message
		})]
	}) : /* @__PURE__ */ (0, P.jsxs)(_v, { children: [
		/* @__PURE__ */ (0, P.jsx)(c, {
			icon: Te,
			className: "mr-2 animate-spin"
		}),
		"Loading ",
		e.file
	] }) : /* @__PURE__ */ (0, P.jsx)(Q_, {
		path: e.file,
		view: i,
		findings: n,
		tabs: a,
		toolbar: s,
		preview: l,
		...b ? { source: b.text } : {},
		...d ? { imageUrl: d } : {},
		...p ? { resolveUrl: p } : {},
		...b && m ? { onSourceChange: m } : {},
		...u ? { comments: u } : {},
		...h === void 0 ? {} : { selectToComment: h },
		...g === void 0 ? {} : { authorFilter: g },
		..._ ? { layout: _ } : {},
		...v ? { onLayoutChange: v } : {},
		...y,
		className: "h-full p-density-2"
	}, j(e)), /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
			"data-file-header": "",
			children: [
				t && /* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 truncate font-mono text-sm",
					title: e.workDir ? `${e.workDir}/${e.file}` : e.file,
					children: e.file
				}),
				/* @__PURE__ */ (0, P.jsx)(sv, { findings: n }),
				/* @__PURE__ */ (0, P.jsx)("span", { className: "flex-1" }),
				/* @__PURE__ */ (0, P.jsx)(ov, {
					items: o,
					label: "File actions"
				})
			]
		}), /* @__PURE__ */ (0, P.jsx)("div", {
			className: "min-h-0 flex-1",
			children: x
		})]
	});
}
function yv({ findings: e, files: t, folders: n, selected: r, onSelect: i, view: o, onViewChange: s, loadFile: c, onSourceChange: l, fileUrl: u, getFileActions: d, getFolderActions: f, fileToolbar: p, fileTabs: m, filePreview: h, folderTabs: g, comments: _, treeNode: v, treeActions: y, showPath: b = !0, defaultSplit: x = 28, className: S, ...C }) {
	let w = (0, N.useMemo)(() => Ue(e, t, n), [
		e,
		t,
		n
	]), T = (0, N.useMemo)(() => hv(w, r), [w, r]), E = T?.kind === "file" ? T.ref : null, ee = E ? j(E) : null, te = (0, N.useMemo)(() => E && u && dv(E.file) ? (e, t) => pv({
		file: E,
		url: e,
		kind: t,
		fileUrl: u
	}) : void 0, [ee, u]), { state: ne, replaceText: D } = gv(r?.kind === "file" && !(u && nv(r.ref.file)) ? r.ref : null, c), O = (e) => (e.kind === "folder" ? f : d)?.(e.ref) ?? [], re = /* @__PURE__ */ (0, P.jsx)(ye, {
		className: "min-h-0",
		roots: w,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		selected: T,
		revealSelected: !0,
		onSelect: (e) => i({
			kind: e.kind,
			ref: e.ref
		}),
		defaultOpen: () => !0,
		showSearch: !0,
		searchLabel: "Filter files",
		getSearchText: (e) => e.name,
		...y ? { actions: y } : {},
		empty: /* @__PURE__ */ (0, P.jsx)("div", {
			className: "p-8 text-center text-sm text-muted-foreground",
			children: "No lint violations or failed tests with a file location"
		}),
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, P.jsx)(mv, {
			node: e,
			selected: t,
			slots: v,
			actions: O(e)
		})
	}), k;
	if (!r) k = /* @__PURE__ */ (0, P.jsx)(_v, { children: "Select a file to see its findings in context." });
	else if (!T) k = /* @__PURE__ */ (0, P.jsxs)(_v, { children: [r.ref.file || "This folder", " has no findings in this run."] });
	else if (T.kind === "folder") {
		let e = g?.(T.ref) ?? [];
		k = /* @__PURE__ */ (0, P.jsx)(uv, {
			folder: T.ref,
			name: T.name,
			showPath: b,
			findings: T.findings,
			view: e.some((e) => e.id === o) ? o : Hg,
			onViewChange: s,
			problemFilters: C.problemFilters,
			onProblemFiltersChange: C.onProblemFiltersChange,
			onOpenFile: (e, t) => {
				i({
					kind: "file",
					ref: e
				}), t !== void 0 && C.onLineChange(t);
			},
			getFindingActions: C.getFindingActions,
			actions: O(T),
			tabs: e
		});
	} else {
		let e = T.ref, t = m?.(e) ?? [], n = h?.(e), r = dv(e.file), i = !!u && tv(e.file), a = !!u && nv(e.file);
		k = /* @__PURE__ */ (0, P.jsx)(vv, {
			...C,
			file: e,
			showPath: b,
			findings: T.findings,
			state: ne,
			view: fv(o, {
				markdown: r || !!n,
				image: i,
				raster: a
			}, t),
			preview: n,
			onViewChange: s,
			tabs: t,
			actions: O(T),
			toolbar: p?.(e),
			comments: _?.(e),
			imageUrl: u && i ? u(e) : void 0,
			raster: a,
			resolveUrl: te,
			onSourceChange: l && !a ? (t) => {
				D(t), l(e, t);
			} : void 0
		});
	}
	return /* @__PURE__ */ (0, P.jsx)("div", {
		className: a("flex h-full min-h-0 flex-col bg-background", S),
		"data-file-findings": "",
		children: /* @__PURE__ */ (0, P.jsx)(Ce, {
			className: "min-h-0 flex-1",
			defaultSplit: x,
			minLeft: 15,
			minRight: 40,
			leftClass: "border-r-2 border-border bg-muted/30",
			left: re,
			right: k
		})
	});
}
// src/FilesTab.tsx
var bv = /* @__PURE__ */ t({ default: () => xv });
function xv(e) {
	return /* @__PURE__ */ (0, P.jsx)(yv, { ...e });
}
export { oi as A, mi as C, pi as D, vi as E, gt as M, _t as N, hi as O, yt as P, Ti as S, _i as T, _m as _, zm as a, im as b, Lm as c, Dm as d, xv as default, Sm as f, jm as g, Am as h, Qh as i, Br as j, gi as k, Rm as l, km as m, Kg as n, Hm as o, Em as p, pg as r, Nm as s, bv as t, Om as u, gm as v, Si as w, Ci as x, rm as y };
