import { n as e, r as t } from "./rolldown-runtime-QGRDbDUS.js";
import { n, t as r } from "./jsx-runtime-lVm_Myqd.js";
import { t as i } from "./react-dom-BBUmKE87.js";
import { C as a, S as o, T as s, _ as c } from "./Callout-DmCAGTzW.js";
import { D as l, H as u, K as d, U as f, V as p, Y as m, Z as h, _ as g, _t as _, bt as v, d as y, dt as b, et as x, ft as S, gt as C, it as w, k as T, lt as E, m as ee, pt as te, q as D, rt as ne, st as re, t as ie, tt as O, u as ae, xt as k } from "./clipboard-C0PFseuO.js";
import { D as oe, E as se, H as ce, L as A, R as le, k as ue, o as de, q as fe } from "./paths-RfM6xXIw.js";
import { n as j, t as pe } from "./Markdown-D7xfvz93.js";
import { a as me, i as he, t as ge, u as _e } from "./adapters-BVPn88xy.js";
import { t as ve } from "./Tabs-Cbv7DUex.js";
// ../../../clicky-ui/packages/ui/src/lib/palette.ts
function ye(e) {
	let t = 2166136261;
	for (let n = 0; n < e.length; n++) t ^= e.charCodeAt(n), t = Math.imul(t, 16777619);
	return t >>> 0;
}
// ../../../clicky-ui/packages/ui/src/data/Avatar.tsx
var M = /* @__PURE__ */ t(n(), 1), N = r();
function be(e) {
	return ye(e) % 360;
}
function xe(e, t, n) {
	if (t?.trim()) return t.trim().slice(0, n).toUpperCase();
	let r = e.replace(/^@/, "").trim(), i = (r.split(/[\\/]/).filter(Boolean).at(-1) ?? r).replace(/[^a-zA-Z0-9]+/g, " ").trim().split(/\s+/).filter(Boolean).map((e) => e.match(/[a-zA-Z0-9]/)?.[0] ?? "").filter(Boolean).slice(0, n).join("");
	return i ? i.toUpperCase() : (r.replace(/[^a-zA-Z0-9]+/g, "").slice(0, n) || "?").toUpperCase();
}
function P(e, t, n) {
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
function Se(e, t) {
	let n = t <= 16 ? 7 : t <= 20 ? 8 : 9;
	switch (e) {
		case "stamp": return Math.max(n, Math.round(t * .34));
		case "mono": return Math.max(n, Math.round(t * .36));
		default: return Math.max(n, Math.round(t * .38));
	}
}
function F({ src: e, alt: t, initials: n, size: r = "sm", rounded: i, kind: a = "user", variant: o = "duotone", title: c, href: l, colorKey: u, onError: d, className: f }) {
	let [p, m] = (0, M.useState)(), h = v(), g = k(r, h), _ = (i ?? (a === "group" ? "md" : "full")) === "full" ? "rounded-full" : "rounded-md", y = P(o, a, be(u ?? t)), b = xe(t, n, r === "xs" || r === "sm" ? 1 : 2), x = s("inline-flex shrink-0 select-none items-center justify-center overflow-hidden border", _), S = {
		width: g,
		height: g,
		borderColor: y.borderColor,
		borderStyle: y.borderStyle ?? "solid"
	};
	o === "stamp" && (S.transform = "rotate(-4deg)", S.boxShadow = y.shadow);
	let C = (e) => e === "stamp" ? 700 : r === "sm" ? 500 : 600, w = e && e !== p ? /* @__PURE__ */ (0, N.jsx)("img", {
		src: e,
		alt: t,
		title: c ?? t,
		width: g,
		height: g,
		"data-avatar-kind": a,
		"data-avatar-variant": o,
		"data-avatar-size": r,
		className: s(x, "bg-muted object-cover", f),
		style: S,
		loading: "lazy",
		onError: (t) => {
			m(e), d?.(t);
		}
	}) : /* @__PURE__ */ (0, N.jsx)("span", {
		role: "img",
		"aria-label": t,
		"data-avatar-kind": a,
		"data-avatar-variant": o,
		"data-avatar-size": r,
		className: s(x, o === "mono" || o === "stamp" ? "font-mono" : void 0, o === "mono" ? "tracking-[-0.05em]" : "tracking-[-0.02em]", f),
		style: {
			...S,
			background: y.background,
			color: y.color,
			fontSize: Se(o, g),
			fontWeight: C(o)
		},
		title: c ?? t,
		children: b
	});
	return l ? /* @__PURE__ */ (0, N.jsx)("a", {
		href: l,
		target: "_blank",
		rel: "noopener noreferrer",
		className: "inline-flex shrink-0",
		onClick: (e) => e.stopPropagation(),
		children: w
	}) : w;
}
// ../../../clicky-ui/packages/ui/src/icons/components/UiThumbsUp.tsx
var I = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Ce = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var we = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Te = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsx)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: /* @__PURE__ */ (0, N.jsx)("path", {
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
var Ee = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
var L = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#dc2626",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#dc2626",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
var De = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M229.19 213c-15.81-27.32-40.63-46.49-69.47-54.62a70 70 0 1 0-63.44 0C67.44 166.5 42.62 185.67 26.81 213a6 6 0 1 0 10.38 6c19.21-33.19 53.15-53 90.81-53s71.6 19.81 90.81 53a6 6 0 1 0 10.38-6M70 96a58 58 0 1 1 58 58a58.07 58.07 0 0 1-58-58"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, N.jsxs)("svg", {
	xmlns: "http://www.w3.org/2000/svg",
	width: e,
	height: e,
	viewBox: "0 0 256 256",
	role: n ? "img" : "presentation",
	"aria-label": n,
	"aria-hidden": !n || void 0,
	className: t,
	...r,
	children: [/* @__PURE__ */ (0, N.jsx)("path", {
		fill: "currentColor",
		d: "M230.93 220a8 8 0 0 1-6.93 4H32a8 8 0 0 1-6.92-12c15.23-26.33 38.7-45.21 66.09-54.16a72 72 0 1 1 73.66 0c27.39 8.95 50.86 27.83 66.09 54.16a8 8 0 0 1 .01 8"
	}), /* @__PURE__ */ (0, N.jsxs)("g", {
		transform: "translate(130.560 130.560) scale(4.8000)",
		color: "#16a34a",
		children: [/* @__PURE__ */ (0, N.jsx)("circle", {
			cx: "12",
			cy: "12",
			r: "11",
			fill: "white"
		}), /* @__PURE__ */ (0, N.jsx)("path", {
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
// ../../../clicky-ui/packages/ui/src/comments/comment-types.ts
var Oe = "__document__", ke = [
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
], Ae = [
	"open",
	"in_progress",
	"done"
];
// ../../../clicky-ui/packages/ui/src/comments/comment-utils.ts
function R(e) {
	return e?.name?.trim() || "Unknown";
}
function z(e) {
	switch (e) {
		case "success": return "success";
		case "danger": return "danger";
		case "warning": return "warning";
		case "info": return "info";
		default: return "neutral";
	}
}
function je(e, t) {
	if (t != null) return e.statuses.find((e) => e.value === t);
}
function Me(e, t) {
	if (t != null) return e.options.find((e) => e.value === t);
}
function B(e) {
	return e.filter((e) => !e.parentId);
}
function Ne(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (!n.parentId) continue;
		let e = t.get(n.parentId) ?? [];
		e.push(n), t.set(n.parentId, e);
	}
	return t;
}
function Pe(e, t) {
	let n = e.filter((e) => !e.parentId && (e.anchor ?? "__document__") === t), r = new Set(n.map((e) => String(e.id))), i = [];
	for (let t = !0; t;) {
		t = !1;
		for (let n of e) n.parentId && !r.has(String(n.id)) && r.has(String(n.parentId)) && (r.add(String(n.id)), i.push(n), t = !0);
	}
	return [...n, ...i];
}
function Fe(e) {
	return [...e].sort((e, t) => e.createdAt === t.createdAt ? String(e.id).localeCompare(String(t.id)) : e.createdAt.localeCompare(t.createdAt));
}
function Ie(e) {
	let t = {};
	for (let n of e) {
		let e = n.anchor ?? "__document__";
		t[e] = (t[e] ?? 0) + 1;
	}
	return t;
}
function V(e) {
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
function H(e, t = Ae) {
	let n = t.indexOf(e);
	return n === -1 ? t[0] ?? e : t[(n + 1) % t.length] ?? e;
}
function Le(e, t = 80) {
	let n = e.replace(/[#*_`~>[\]()!]/g, "").trim();
	return n.length > t ? `${n.slice(0, t)}…` : n;
}
function Re(e, t) {
	let n = e.toLowerCase();
	return t.filter((e) => n.includes(`@${e.name.toLowerCase()}`)).map((e) => ({
		id: e.id,
		name: e.name,
		kind: e.kind
	}));
}
function ze(e, t, n) {
	let r = {};
	if (n.onUpdateStatus && (r.onUpdateStatus = n.onUpdateStatus), n.onClose && (r.onClose = n.onClose), n.onAcknowledge && (r.onAcknowledge = n.onAcknowledge), n.onAssign && (r.onAssign = n.onAssign), n.onDelete && (r.onDelete = n.onDelete), n.onChecklistToggle) {
		let i = n.onChecklistToggle;
		r.onChecklistToggle = (n, r) => {
			let a = e.find((e) => e.id === n)?.checklist?.[r];
			a && i(n, r, H(a.status, t.checklistStatusCycle));
		};
	}
	if (n.onReply) {
		let e = n.onReply, i = n.onMention;
		r.onReply = async (n, r) => {
			let a = Re(r, t.mentionables ?? []);
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
function U(e, t) {
	if (t == null) return "active";
	let n = je(e, t);
	if (n) return n.stage ? n.stage : n.unresolved ? "active" : n.value.toLowerCase() === "closed" || n.label.toLowerCase() === "closed" ? "closed" : "resolved";
}
function W(e, t) {
	return e.statuses.find((n) => U(e, n.value) === t);
}
function Be(e, t, n) {
	let r = new Set(B(e).filter((e) => U(t, e.status) === n).map((e) => String(e.id)));
	return e.filter((e) => r.has(String(e.parentId ?? e.id)));
}
// ../../../clicky-ui/packages/ui/src/comments/CommentAuthor.tsx
function Ve({ author: e, size: t = "sm", bare: n = !1 }) {
	let r = v(), i = R(e), a = e?.kind === "agent", o = a ? /* @__PURE__ */ (0, N.jsx)(He, {
		author: e,
		size: k(t, r)
	}) : /* @__PURE__ */ (0, N.jsx)(F, {
		alt: i,
		size: t,
		...e?.avatar ? { src: e.avatar } : {},
		title: i
	});
	return n ? o : /* @__PURE__ */ (0, N.jsxs)(y, {
		trigger: o,
		placement: "top",
		children: [/* @__PURE__ */ (0, N.jsx)("span", {
			className: "font-medium",
			children: i
		}), a && /* @__PURE__ */ (0, N.jsx)("span", {
			className: "ml-1 text-muted-foreground",
			children: "· agent"
		})]
	});
}
function He({ author: e, size: t }) {
	let n = e.icon;
	return /* @__PURE__ */ (0, N.jsx)("span", {
		role: "img",
		"aria-label": e.name,
		className: "inline-flex shrink-0 items-center justify-center rounded-full border border-border bg-muted text-foreground",
		style: {
			width: t,
			height: t
		},
		children: /* @__PURE__ */ (0, N.jsx)(_, {
			...typeof n == "string" ? { name: n } : { icon: n ?? D },
			width: Math.max(10, Math.round(t * .6)),
			height: Math.max(10, Math.round(t * .6))
		})
	});
}
// ../../../clicky-ui/packages/ui/src/comments/MentionTextarea.tsx
var Ue = /* @__PURE__ */ t(i(), 1), G = /(^|\s)@([\w.-]*)$/, We = 8;
function K(e, t) {
	let n = e.slice(0, t).match(G);
	if (!n) return null;
	let r = n[2] ?? "";
	return {
		atIndex: t - r.length - 1,
		query: r
	};
}
function Ge(e, t) {
	let n = t.toLowerCase();
	return (n ? e.filter((e) => e.name.toLowerCase().includes(n) || e.id.toLowerCase().includes(n)) : e).slice(0, We);
}
var Ke = /* @__PURE__ */ "boxSizing.width.borderTopWidth.borderRightWidth.borderBottomWidth.borderLeftWidth.paddingTop.paddingRight.paddingBottom.paddingLeft.fontStyle.fontVariant.fontWeight.fontStretch.fontSize.lineHeight.fontFamily.textAlign.textTransform.textIndent.letterSpacing.wordSpacing.tabSize.whiteSpace.wordWrap.overflowWrap".split(".");
function qe(e, t) {
	if (typeof document > "u" || typeof getComputedStyle != "function") {
		let t = e.getBoundingClientRect();
		return {
			top: t.top,
			left: t.left,
			height: 16
		};
	}
	let n = getComputedStyle(e), r = document.createElement("div");
	for (let e of Ke) r.style[e] = n[e];
	r.style.position = "absolute", r.style.visibility = "hidden", r.style.whiteSpace = "pre-wrap", r.style.overflowWrap = "break-word", r.style.top = "0", r.style.left = "-9999px", r.textContent = e.value.slice(0, t);
	let i = document.createElement("span");
	i.textContent = e.value.slice(t) || ".", r.appendChild(i), document.body.appendChild(r);
	let a = parseInt(n.lineHeight, 10) || parseInt(n.fontSize, 10) || 16, o = i.offsetTop, s = i.offsetLeft;
	document.body.removeChild(r);
	let c = e.getBoundingClientRect();
	return {
		top: c.top + o - e.scrollTop,
		left: c.left + s - e.scrollLeft,
		height: a
	};
}
var Je = (0, M.forwardRef)(function({ value: e, onChange: t, mentionables: n, onMentionSelect: r, onSubmit: i, onCancel: a, placeholder: o, autoFocus: c, rows: l = 2, className: u, "data-testid": d }, f) {
	let p = (0, M.useRef)(null);
	(0, M.useImperativeHandle)(f, () => p.current);
	let [m, h] = (0, M.useState)(null), [g, _] = (0, M.useState)(null), [v, y] = (0, M.useState)(0), b = T(), x = m ? Ge(n, m.query) : [], S = m != null && x.length > 0;
	function C(e) {
		let t = K(e.value, e.selectionStart ?? e.value.length);
		h(t), y(0), _(t ? qe(e, e.selectionStart ?? 0) : null);
	}
	function w(n) {
		let i = p.current;
		if (!i || !m) return;
		let a = i.selectionStart ?? e.length, o = `${e.slice(0, m.atIndex)}@${n.name} ${e.slice(a)}`, s = m.atIndex + n.name.length + 2;
		t(o), r?.({
			id: n.id,
			name: n.name,
			kind: n.kind
		}), h(null), _(null), requestAnimationFrame(() => {
			i.focus(), i.setSelectionRange(s, s);
		});
	}
	function E(e) {
		if (S) {
			if (e.key === "ArrowDown") {
				e.preventDefault(), y((e) => (e + 1) % x.length);
				return;
			}
			if (e.key === "ArrowUp") {
				e.preventDefault(), y((e) => (e - 1 + x.length) % x.length);
				return;
			}
			if (e.key === "Enter" || e.key === "Tab") {
				e.preventDefault();
				let t = x[v];
				t && w(t);
				return;
			}
			if (e.key === "Escape") {
				e.preventDefault(), h(null), _(null);
				return;
			}
		}
		e.key === "Enter" && !e.shiftKey ? (e.preventDefault(), i?.()) : e.key === "Escape" && (e.preventDefault(), a?.());
	}
	return (0, M.useEffect)(() => {
		if (!c) return;
		let e = p.current, t = requestAnimationFrame(() => {
			e?.focus();
			let t = e?.value.length ?? 0;
			e?.setSelectionRange(t, t);
		});
		return () => cancelAnimationFrame(t);
	}, [c]), /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)("textarea", {
		ref: p,
		rows: l,
		value: e,
		placeholder: o,
		"data-testid": d,
		className: s("w-full resize-none bg-transparent text-sm leading-snug outline-none placeholder:text-placeholder [@media(pointer:coarse)]:text-base", u),
		onChange: (e) => {
			t(e.target.value), C(e.target);
		},
		onKeyUp: (e) => C(e.currentTarget),
		onClick: (e) => C(e.currentTarget),
		onKeyDown: E,
		onBlur: () => {
			h(null), _(null);
		}
	}), S && typeof document < "u" && (0, Ue.createPortal)(/* @__PURE__ */ (0, N.jsx)("ul", {
		role: "listbox",
		"data-testid": "mention-popover",
		style: {
			position: "fixed",
			top: (g?.top ?? 0) + (g?.height ?? 16) + 2,
			left: g?.left ?? 0,
			zIndex: b
		},
		className: "max-h-56 min-w-[12rem] overflow-auto rounded-md border border-border bg-popover py-1 shadow-md",
		children: x.map((e, t) => /* @__PURE__ */ (0, N.jsx)("li", { children: /* @__PURE__ */ (0, N.jsxs)("button", {
			type: "button",
			role: "option",
			"aria-selected": t === v,
			onMouseDown: (t) => {
				t.preventDefault(), w(e);
			},
			onMouseEnter: () => y(t),
			className: s("flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground", t === v && "bg-accent text-accent-foreground"),
			children: [
				/* @__PURE__ */ (0, N.jsx)(Ve, {
					author: {
						name: e.name,
						kind: e.kind,
						...e.avatar ? { avatar: e.avatar } : {},
						...e.icon ? { icon: e.icon } : {}
					},
					size: "xs",
					bare: !0
				}),
				/* @__PURE__ */ (0, N.jsx)("span", {
					className: "truncate",
					children: e.name
				}),
				e.kind === "agent" && /* @__PURE__ */ (0, N.jsx)("span", {
					className: "ml-auto text-[10px] text-muted-foreground",
					children: "agent"
				})
			]
		}) }, e.id))
	}), document.body)] });
}), q = {
	neutral: "bg-muted-foreground",
	success: "bg-green-500",
	danger: "bg-red-500",
	warning: "bg-yellow-500",
	info: "bg-blue-500"
};
function Ye({ facet: e, value: t, onChange: n }) {
	let r = e.options.find((e) => e.value === t);
	return /* @__PURE__ */ (0, N.jsx)(ae, {
		align: "left",
		menuClassName: "min-w-[150px]",
		trigger: /* @__PURE__ */ (0, N.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: s("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium transition-colors", r ? "bg-muted text-foreground" : "text-muted-foreground hover:bg-muted/60"),
			children: [r ? r.label : e.label, r && /* @__PURE__ */ (0, N.jsx)("span", {
				role: "presentation",
				className: "ml-0.5 cursor-pointer hover:text-foreground",
				onClick: (e) => {
					e.stopPropagation(), n(void 0);
				},
				children: "×"
			})]
		}),
		children: (r) => /* @__PURE__ */ (0, N.jsx)("div", {
			className: "py-1",
			children: e.options.map((e) => /* @__PURE__ */ (0, N.jsxs)("button", {
				type: "button",
				onClick: () => {
					n(e.value === t ? void 0 : e.value), r();
				},
				className: s("flex w-full items-center gap-2 px-3 py-1.5 text-left text-[11px] text-popover-foreground hover:bg-accent hover:text-accent-foreground", e.value === t && "bg-muted font-semibold"),
				children: [/* @__PURE__ */ (0, N.jsx)("span", { className: s("inline-block size-2 rounded-full", q[z(e.tone)]) }), e.label]
			}, e.value))
		})
	});
}
function Xe({ config: e, anchor: t, collapsible: n = !0, autoFocus: r = !1, placeholder: i = "Add a comment…", onCreate: a, createActions: o, onMention: c }) {
	let [l, u] = (0, M.useState)(!n || r), [d, f] = (0, M.useState)(""), [p, m] = (0, M.useState)({}), [h, g] = (0, M.useState)(), [v, y] = (0, M.useState)(!1), b = (0, M.useRef)(!1), x = (0, M.useRef)(/* @__PURE__ */ new Map()), S = (0, M.useRef)(null);
	function C() {
		f(""), m({}), g(void 0), x.current.clear();
	}
	function w(e, t) {
		m((n) => {
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
	async function T(e) {
		let r = d.trim();
		if (!r && !h || b.current) return;
		b.current = !0, y(!0);
		let i = [...x.current.values()].filter((e) => d.toLowerCase().includes(`@${e.name.toLowerCase()}`));
		try {
			let o = {
				body: r,
				anchor: t ?? null,
				...h ? { rating: h } : {},
				...Object.keys(p).length > 0 ? { facets: p } : {},
				...i.length > 0 ? { mentions: i } : {}
			};
			e === void 0 ? await a?.(o) : await a?.(o, e);
			for (let e of i) c?.(e, { body: r });
			C(), n && u(!1);
		} finally {
			b.current = !1, y(!1);
		}
	}
	return !l && d.trim().length === 0 ? /* @__PURE__ */ (0, N.jsx)("button", {
		type: "button",
		onClick: () => u(!0),
		className: "w-full rounded-full border border-border bg-muted/25 px-4 py-2 text-left text-sm text-muted-foreground/70 shadow-sm transition-colors hover:border-primary/30 hover:bg-muted/40",
		children: i
	}) : /* @__PURE__ */ (0, N.jsxs)("div", {
		"data-testid": "comment-compose-box",
		"aria-busy": v,
		className: "relative rounded-2xl border border-border bg-muted/25 px-3 pt-3 pb-10 shadow-sm transition-[border-color,box-shadow] focus-within:border-primary/40 focus-within:shadow-md [@media(pointer:coarse)]:pb-2",
		children: [
			/* @__PURE__ */ (0, N.jsx)(Je, {
				ref: S,
				value: d,
				onChange: f,
				mentionables: e.mentionables ?? [],
				onMentionSelect: (e) => x.current.set(e.id, e),
				onSubmit: () => void T(),
				onCancel: () => {
					C(), n && u(!1);
				},
				placeholder: i,
				autoFocus: r || n && l,
				"data-testid": "comment-compose-input",
				className: "min-h-[48px] pr-10 [@media(pointer:coarse)]:pr-12"
			}),
			n && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				onClick: () => {
					C(), u(!1);
				},
				"aria-label": "Close composer",
				className: "absolute right-3 top-3 inline-flex items-center justify-center rounded-full p-1 text-muted-foreground hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:right-1 [@media(pointer:coarse)]:top-1 [@media(pointer:coarse)]:size-11",
				children: /* @__PURE__ */ (0, N.jsx)(_, {
					icon: E,
					className: "text-xs"
				})
			}),
			/* @__PURE__ */ (0, N.jsxs)("div", {
				className: "absolute inset-x-3 bottom-2.5 flex items-center gap-1.5 [@media(pointer:coarse)]:static [@media(pointer:coarse)]:mt-2 [@media(pointer:coarse)]:flex-wrap [@media(pointer:coarse)]:gap-2",
				children: [
					[[
						"positive",
						"Positive rating",
						I
					], [
						"negative",
						"Negative rating",
						Ce
					]].map(([e, t, n]) => /* @__PURE__ */ (0, N.jsx)("button", {
						type: "button",
						"aria-label": t,
						"aria-pressed": h === e,
						onClick: () => g((t) => t === e ? void 0 : e),
						className: s("inline-flex size-7 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11", h === e && (e === "positive" ? "bg-green-100 text-green-700 [[data-theme=dark]_&]:bg-green-500/20 [[data-theme=dark]_&]:text-green-300" : "bg-red-100 text-red-700 [[data-theme=dark]_&]:bg-red-500/20 [[data-theme=dark]_&]:text-red-300")),
						children: /* @__PURE__ */ (0, N.jsx)(_, {
							icon: n,
							className: "text-sm"
						})
					}, e)),
					(e.facets ?? []).map((e) => /* @__PURE__ */ (0, N.jsx)(Ye, {
						facet: e,
						value: p[e.key],
						onChange: (t) => w(e.key, t)
					}, e.key)),
					o?.length ? /* @__PURE__ */ (0, N.jsx)("div", {
						className: "ml-auto flex gap-1 [@media(pointer:coarse)]:flex-wrap [@media(pointer:coarse)]:justify-end [@media(pointer:coarse)]:gap-2",
						children: o.map((e) => /* @__PURE__ */ (0, N.jsx)("button", {
							type: "button",
							onClick: () => void T(e.id),
							disabled: v || !d.trim() && !h,
							className: "whitespace-nowrap rounded-md bg-primary px-1.5 py-1 text-[10px] font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
							children: e.label
						}, e.id))
					}) : /* @__PURE__ */ (0, N.jsx)("button", {
						type: "button",
						"data-testid": "comment-compose-send",
						"aria-label": "Post comment",
						onClick: () => void T(),
						disabled: v || !d.trim() && !h,
						className: "ml-auto inline-flex size-7 items-center justify-center rounded-full bg-primary text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground [@media(pointer:coarse)]:size-11",
						children: /* @__PURE__ */ (0, N.jsx)(_, {
							icon: j,
							className: "text-xs"
						})
					})
				]
			})
		]
	});
}
// ../../../clicky-ui/packages/ui/src/comments/CommentCardParts.tsx
function Ze({ decision: e }) {
	return e ? /* @__PURE__ */ (0, N.jsx)(C, {
		variant: "soft",
		tone: e === "approved" ? "success" : e === "rejected" ? "danger" : "warning",
		size: "xs",
		children: e === "approved" ? "Approved" : e === "rejected" ? "Rejected" : "Revision requested"
	}) : null;
}
function Qe({ rating: e }) {
	if (!e) return null;
	let t = e === "positive";
	return /* @__PURE__ */ (0, N.jsx)(C, {
		variant: "soft",
		tone: t ? "success" : "danger",
		size: "xs",
		icon: t ? I : Ce,
		children: t ? "Positive rating" : "Negative rating"
	});
}
function $e({ comment: e, config: t }) {
	let n = je(t, e.status);
	if (!n) return null;
	let r = U(t, e.status), i = r === "closed" ? e.closedBy : void 0, a = r === "active" ? e.assignee : void 0, o = a?.kind === "agent" ? D : n.icon, s = a ? e.acknowledgedAt ? `Acknowledged ${new Date(e.acknowledgedAt).toLocaleString()}` : "Assigned, not yet acknowledged" : e.closedAt ? new Date(e.closedAt).toLocaleString() : void 0;
	return /* @__PURE__ */ (0, N.jsx)(C, {
		variant: "soft",
		tone: z(n.tone),
		size: "xs",
		...o ? { icon: o } : {},
		...s ? { title: s } : {},
		children: a ? `Assigned to ${a.name}` : i ? `Closed by ${i.name}` : n.label
	});
}
function et({ comment: e, config: t, compact: n }) {
	let r = (t.facets ?? []).flatMap((t) => {
		let r = Me(t, e.facets?.[t.key]);
		return r ? [/* @__PURE__ */ (0, N.jsx)(C, {
			variant: "soft",
			tone: z(r.tone),
			size: "xs",
			children: n ? r.short ?? r.label : r.label
		}, t.key)] : [];
	});
	return r.length > 0 ? /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: r }) : null;
}
function tt({ comment: e, onChecklistToggle: t }) {
	return e.checklist?.length ? /* @__PURE__ */ (0, N.jsx)("div", {
		className: "flex flex-wrap gap-1",
		children: e.checklist.map((e, n) => {
			let r = /done|complete|resolved/i.test(e.status), i = /progress|doing/i.test(e.status);
			return /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				onClick: () => t?.(n),
				disabled: !t,
				className: s("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] transition-opacity", t && "cursor-pointer hover:opacity-80", r ? "bg-green-100 text-green-800 line-through dark:bg-green-500/20 dark:text-green-300" : i ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-muted text-muted-foreground"),
				children: e.label
			}, n);
		})
	}) : null;
}
function nt({ comment: e }) {
	return e.refs?.length ? /* @__PURE__ */ (0, N.jsx)("div", {
		className: "flex flex-wrap gap-1.5",
		children: e.refs.map((e, t) => /* @__PURE__ */ (0, N.jsxs)("span", {
			className: "inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground",
			children: [/* @__PURE__ */ (0, N.jsxs)("span", {
				className: "font-medium",
				children: [e.label, ":"]
			}), /* @__PURE__ */ (0, N.jsx)("span", {
				className: e.mono ? "font-mono" : void 0,
				children: e.items.join(", ")
			})]
		}, t))
	}) : null;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentMarkdown.tsx
function rt(e) {
	let t = e.trim();
	return t && (/^(https?:|mailto:)/i.test(t) || t.startsWith("/") && !t.startsWith("//")) ? t : null;
}
var J = /(\*\*[^*\n]+\*\*|`[^`\n]+`|\[[^\]\n]+\]\(([^)\s]+)\)|https?:\/\/[^\s<)]+|\*[^*\n]+\*)/g;
function Y(e) {
	let t = [], n = 0;
	for (let r of e.matchAll(J)) {
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
			let e = a.indexOf("]("), n = rt(a.slice(e + 2, -1));
			t.push(n ? {
				type: "link",
				text: a.slice(1, e),
				href: n
			} : {
				type: "text",
				text: a
			});
		} else if (/^https?:\/\//i.test(a)) {
			let e = rt(a);
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
function it({ text: e }) {
	return /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: Y(e).map((e, t) => {
		let n = `${e.type}-${t}`;
		switch (e.type) {
			case "strong": return /* @__PURE__ */ (0, N.jsx)("strong", {
				className: "font-semibold",
				children: e.text
			}, n);
			case "em": return /* @__PURE__ */ (0, N.jsx)("em", { children: e.text }, n);
			case "code": return /* @__PURE__ */ (0, N.jsx)("code", {
				className: "rounded bg-muted px-1 py-0.5 font-mono text-[0.9em]",
				children: e.text
			}, n);
			case "link": return /* @__PURE__ */ (0, N.jsx)("a", {
				href: e.href,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary underline underline-offset-2",
				children: e.text
			}, n);
			default: return /* @__PURE__ */ (0, N.jsx)(M.Fragment, { children: e.text }, n);
		}
	}) });
}
function at(e, t) {
	return /* @__PURE__ */ (0, N.jsx)("p", { children: e.map((e, t) => /* @__PURE__ */ (0, N.jsxs)(M.Fragment, { children: [t > 0 && /* @__PURE__ */ (0, N.jsx)("br", {}), /* @__PURE__ */ (0, N.jsx)(it, { text: e })] }, t)) }, t);
}
function X(e, t) {
	return /* @__PURE__ */ (0, N.jsx)("pre", {
		className: "overflow-auto rounded bg-muted p-2 text-xs",
		children: /* @__PURE__ */ (0, N.jsx)("code", { children: e.join("\n") })
	}, t);
}
function ot({ text: e, className: t }) {
	let n = String(e ?? "").replace(/\r\n?/g, "\n").split("\n"), r = [], i = [], a = [], o = null;
	function c() {
		i.length > 0 && (r.push(at(i, `p-${r.length}`)), i = []);
	}
	function l() {
		a.length > 0 && (r.push(/* @__PURE__ */ (0, N.jsx)("ul", {
			className: "list-disc space-y-0.5 pl-4",
			children: a.map((e, t) => /* @__PURE__ */ (0, N.jsx)("li", { children: /* @__PURE__ */ (0, N.jsx)(it, { text: e }) }, t))
		}, `ul-${r.length}`)), a = []);
	}
	for (let e of n) {
		if (e.trim().startsWith("```")) {
			c(), l(), o ? (r.push(X(o, `code-${r.length}`)), o = null) : o = [];
			continue;
		}
		if (o) {
			o.push(e);
			continue;
		}
		let t = e.trim();
		if (!t) {
			c(), l();
			continue;
		}
		let n = t.match(/^[-*]\s+(.+)$/);
		n?.[1] ? (c(), a.push(n[1])) : (l(), i.push(t.replace(/^#{1,6}\s+/, "")));
	}
	return c(), l(), o && r.push(X(o, `code-${r.length}`)), /* @__PURE__ */ (0, N.jsx)("div", {
		className: s("space-y-2 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", t),
		children: r.length > 0 ? r : null
	});
}
// ../../../clicky-ui/packages/ui/src/comments/comment-assign.ts
function st(e, t) {
	return e.name === t.name && (e.kind ?? "user") === (t.kind ?? "user");
}
function ct(e) {
	let { comment: t, config: n, onAssign: r, extraAssignItems: i, run: a } = e, o = [];
	if (r) {
		for (let { id: e, name: i, kind: s } of n.assignees ?? []) {
			let n = {
				id: e,
				name: i,
				kind: s
			}, c = t.assignee !== void 0 && st(t.assignee, n);
			o.push({
				label: i,
				icon: s === "agent" ? D : Te,
				disabled: c,
				title: c ? `Already assigned to ${i}` : `Assign to ${i}`,
				onSelect: () => a("assign", () => r(n))
			});
		}
		t.assignee && o.push({
			label: "Unassign",
			icon: L,
			onSelect: () => a("assign", () => r(null))
		});
	}
	return o.push(...i?.(t) ?? []), o;
}
function lt(e) {
	let { comment: t, config: n, onAcknowledge: r, run: i } = e;
	if (t.parentId || U(n, t.status) !== "active") return [];
	let a = [];
	r && !t.assignee && a.push({
		label: "Acknowledge",
		icon: De,
		onSelect: () => i("acknowledge", r)
	});
	let o = ct(e);
	return o.length > 0 && a.push({
		label: "Assign…",
		icon: Ee,
		onSelect: () => {},
		children: o
	}), a;
}
// ../../../clicky-ui/packages/ui/src/comments/CommentCard.tsx
function ut(e, t) {
	let n = [];
	return t.onDelete && n.push({
		label: "Delete",
		icon: d,
		onSelect: t.onDelete
	}), e || n.push(...t.statusActions ?? []), t.onCopy && n.push({
		label: "Copy",
		icon: m,
		onSelect: t.onCopy
	}), t.onMaximize && n.push({
		label: "Maximise",
		icon: re,
		onSelect: t.onMaximize
	}), n;
}
function dt(e) {
	return (t, n) => {
		e(""), (async () => {
			try {
				await n();
			} catch (n) {
				e(`Couldn't ${t} comment: ${n instanceof Error ? n.message : "Unexpected error"}`);
			}
		})();
	};
}
function ft(e, t) {
	let { comment: n, config: r, onAcknowledge: i, onAssign: a, extraAssignItems: o } = e;
	return lt({
		comment: n,
		config: r,
		run: t,
		...i ? { onAcknowledge: i } : {},
		...a ? { onAssign: a } : {},
		...o ? { extraAssignItems: o } : {}
	});
}
function pt(e) {
	let { comment: t, config: n, renderBody: r, onUpdateStatus: i, onClose: a, onDelete: o, onCopy: s, onReply: c, onChecklistToggle: l, onMaximize: u } = e, [d, f] = (0, M.useState)(""), p = !!t.parentId, m = g(t.createdAt), h = W(n, "resolved"), v = W(n, "active"), y = W(n, "closed"), b = U(n, t.status), S = async (e) => {
		f("");
		try {
			await i?.(e);
		} catch (e) {
			f(`Couldn't update comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, C = async () => {
		f("");
		try {
			await a?.();
		} catch (e) {
			f(`Couldn't close comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, T = async () => {
		f("");
		try {
			await s?.();
		} catch (e) {
			f(`Couldn't copy comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, E = ut(p, {
		...o ? { onDelete: o } : {},
		statusActions: [
			...ft(e, dt(f)),
			...b === "active" && i && h ? [{
				label: "Resolve",
				icon: O,
				onSelect: () => void S(h.value)
			}] : [],
			...(b === "active" || b === "resolved") && a && y ? [{
				label: "Close",
				icon: x,
				onSelect: () => void C()
			}] : [],
			...(b === "resolved" || b === "closed") && i && v ? [{
				label: "Reopen",
				icon: w,
				onSelect: () => void S(v.value)
			}] : []
		],
		...s ? { onCopy: () => void T() } : {},
		...u ? { onMaximize: u } : {}
	});
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "space-y-2.5",
		children: [
			/* @__PURE__ */ (0, N.jsxs)("div", {
				className: "flex items-start gap-2.5",
				children: [
					/* @__PURE__ */ (0, N.jsx)(Ve, { author: t.author }),
					/* @__PURE__ */ (0, N.jsxs)("div", {
						className: "min-w-0 flex-1",
						children: [/* @__PURE__ */ (0, N.jsxs)("div", {
							className: "flex min-w-0 items-baseline gap-1.5",
							children: [/* @__PURE__ */ (0, N.jsx)("span", {
								className: "truncate text-xs font-semibold text-foreground",
								children: R(t.author)
							}), m && /* @__PURE__ */ (0, N.jsx)("span", {
								title: m.toLocaleString(),
								className: "shrink-0 text-[10px] text-muted-foreground",
								children: ee(m)
							})]
						}), !p && /* @__PURE__ */ (0, N.jsxs)("div", {
							className: "mt-1 flex flex-wrap gap-1",
							children: [
								/* @__PURE__ */ (0, N.jsx)($e, {
									comment: t,
									config: n
								}),
								/* @__PURE__ */ (0, N.jsx)(Qe, { rating: t.rating }),
								/* @__PURE__ */ (0, N.jsx)(Ze, { decision: t.reviewDecision }),
								/* @__PURE__ */ (0, N.jsx)(et, {
									comment: t,
									config: n
								})
							]
						})]
					}),
					/* @__PURE__ */ (0, N.jsxs)("div", {
						className: "flex shrink-0 items-center gap-0.5",
						children: [
							!p && (b === "active" || b === "resolved") && a && y && /* @__PURE__ */ (0, N.jsx)("button", {
								type: "button",
								"aria-label": "Close comment",
								title: "Close comment",
								onClick: () => void C(),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, N.jsx)(_, {
									icon: O,
									className: "text-sm"
								})
							}),
							!p && b === "closed" && i && v && /* @__PURE__ */ (0, N.jsx)("button", {
								type: "button",
								"aria-label": "Reopen comment",
								title: "Reopen comment",
								onClick: () => void S(v.value),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, N.jsx)(_, {
									icon: w,
									className: "text-sm"
								})
							}),
							E.length > 0 && /* @__PURE__ */ (0, N.jsx)(ae, {
								align: "right",
								menuClassName: "min-w-[150px]",
								trigger: /* @__PURE__ */ (0, N.jsx)("span", {
									role: "button",
									tabIndex: 0,
									"aria-label": "Comment actions",
									className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11",
									children: /* @__PURE__ */ (0, N.jsx)(_, {
										icon: ne,
										className: "text-sm"
									})
								}),
								items: E
							})
						]
					})
				]
			}),
			t.body.trim() && /* @__PURE__ */ (0, N.jsx)("div", {
				className: "text-[13px] leading-5 text-foreground",
				children: r ? r(t.body) : /* @__PURE__ */ (0, N.jsx)(ot, { text: t.body })
			}),
			d && /* @__PURE__ */ (0, N.jsx)("div", {
				role: "alert",
				className: "rounded-md border border-destructive/30 bg-destructive/5 px-2.5 py-2 text-xs text-destructive",
				children: d
			}),
			!p && /* @__PURE__ */ (0, N.jsx)(tt, {
				comment: t,
				...l ? { onChecklistToggle: l } : {}
			}),
			!p && /* @__PURE__ */ (0, N.jsx)(nt, { comment: t }),
			c && /* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				onClick: c,
				className: "-ml-2 rounded px-2 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary/10 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
				children: "Reply"
			})
		]
	});
}
function Z(e) {
	let { comment: t, config: n, compact: r, defaultExpanded: i } = e, [a, o] = (0, M.useState)(i ?? !1), [c, l] = (0, M.useState)(""), u = !!t.parentId, d = /* @__PURE__ */ (0, N.jsx)(pt, { ...e });
	async function f(t) {
		l("");
		try {
			if (t === "close") {
				await e.onClose?.();
				return;
			}
			let r = W(n, t === "resolve" ? "resolved" : "active");
			r && await e.onUpdateStatus?.(r.value);
		} catch (e) {
			l(`${t === "close" ? "Couldn't close comment" : "Couldn't update comment"}: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}
	async function p() {
		l("");
		try {
			await e.onCopy?.();
		} catch (e) {
			l(`Couldn't copy comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}
	return a ? /* @__PURE__ */ (0, N.jsx)(N.Fragment, { children: /* @__PURE__ */ (0, N.jsx)("div", {
		"data-testid": "comment-card",
		"data-comment-kind": u ? "reply" : "root",
		"data-comment-id": t.id,
		className: s("group/comment rounded-xl bg-background px-3.5 py-3 text-xs", !u && "border border-border/80 shadow-sm transition-shadow hover:shadow-md", r && "px-2.5 py-2.5", u && "ml-4 border-l-2 border-l-border bg-muted/20 shadow-none"),
		children: d
	}) }) : /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsxs)("div", {
		role: "button",
		tabIndex: 0,
		onClick: () => o(!0),
		onKeyDown: (e) => {
			(e.key === "Enter" || e.key === " ") && (e.preventDefault(), o(!0));
		},
		"data-testid": "comment-card",
		"data-comment-kind": u ? "reply" : "root",
		"data-comment-id": t.id,
		className: s("group/comment flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs transition-colors hover:bg-muted/50", !u && "border border-border bg-background shadow-sm", u && "ml-4 border-l-2 border-l-border bg-muted/30"),
		children: [
			/* @__PURE__ */ (0, N.jsx)(Ve, {
				author: t.author,
				bare: !0
			}),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "min-w-0 truncate text-[11px] font-semibold text-foreground",
				children: R(t.author)
			}),
			!u && /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [
				/* @__PURE__ */ (0, N.jsx)($e, {
					comment: t,
					config: n
				}),
				/* @__PURE__ */ (0, N.jsx)(Qe, { rating: t.rating }),
				/* @__PURE__ */ (0, N.jsx)(Ze, { decision: t.reviewDecision }),
				/* @__PURE__ */ (0, N.jsx)(et, {
					comment: t,
					config: n,
					compact: !0
				})
			] }),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "min-w-0 flex-1 truncate text-[12px] text-foreground/80",
				children: t.body.trim() ? Le(t.body) : t.rating === "positive" ? "Positive rating" : t.rating === "negative" ? "Negative rating" : t.reviewDecision === "approved" ? "Approved" : "Empty comment"
			}),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "shrink-0",
				role: "presentation",
				onClick: (e) => e.stopPropagation(),
				onKeyDown: (e) => e.stopPropagation(),
				children: /* @__PURE__ */ (0, N.jsx)(ae, {
					align: "right",
					trigger: /* @__PURE__ */ (0, N.jsx)("span", {
						role: "button",
						tabIndex: 0,
						"aria-label": "Comment actions",
						className: "rounded p-0.5 text-muted-foreground opacity-0 transition-opacity hover:bg-muted hover:opacity-100 group-hover/comment:opacity-70 [@media(hover:none)]:opacity-70 [@media(pointer:coarse)]:p-2.5",
						children: /* @__PURE__ */ (0, N.jsx)(_, {
							icon: ne,
							className: "text-sm"
						})
					}),
					items: ut(u, {
						...e.onDelete ? { onDelete: e.onDelete } : {},
						statusActions: [
							...ft(e, dt(l)),
							...U(n, t.status) === "active" && e.onUpdateStatus && W(n, "resolved") ? [{
								label: "Resolve",
								icon: O,
								onSelect: () => void f("resolve")
							}] : [],
							...(U(n, t.status) === "active" || U(n, t.status) === "resolved") && e.onClose && W(n, "closed") ? [{
								label: "Close",
								icon: x,
								onSelect: () => void f("close")
							}] : [],
							...(U(n, t.status) === "resolved" || U(n, t.status) === "closed") && e.onUpdateStatus && W(n, "active") ? [{
								label: "Reopen",
								icon: w,
								onSelect: () => void f("reopen")
							}] : []
						],
						...e.onCopy ? { onCopy: () => void p() } : {},
						...e.onMaximize ? { onMaximize: e.onMaximize } : {}
					})
				})
			})
		]
	}), c && /* @__PURE__ */ (0, N.jsx)("div", {
		role: "alert",
		className: "rounded-md border border-destructive/30 bg-destructive/5 px-2.5 py-2 text-xs text-destructive",
		children: c
	})] });
}
// ../../../clicky-ui/packages/ui/src/comments/CommentThreadList.tsx
function mt({ onSubmit: e, onCancel: t }) {
	let [n, r] = (0, M.useState)(""), [i, a] = (0, M.useState)(!1), o = (0, M.useRef)(!1);
	async function s() {
		if (n.trim() && !o.current) {
			o.current = !0, a(!0);
			try {
				await e(n.trim()), r("");
			} finally {
				o.current = !1, a(!1);
			}
		}
	}
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "ml-6 flex gap-1.5 py-1",
		"aria-busy": i,
		children: [/* @__PURE__ */ (0, N.jsx)("input", {
			autoFocus: !0,
			value: n,
			disabled: i,
			onChange: (e) => r(e.target.value),
			onKeyDown: (e) => {
				e.key === "Enter" && (e.preventDefault(), s()), e.key === "Escape" && t();
			},
			placeholder: "Write a reply…",
			"data-testid": "comment-reply-input",
			className: "h-7 min-w-0 flex-1 rounded-md border border-border bg-background px-2 py-1 text-xs outline-none focus:border-primary/40 [@media(pointer:coarse)]:h-11 [@media(pointer:coarse)]:text-base"
		}), /* @__PURE__ */ (0, N.jsx)(c, {
			size: "sm",
			className: "h-7 text-xs [@media(pointer:coarse)]:h-11 [@media(pointer:coarse)]:px-4 [@media(pointer:coarse)]:text-sm",
			disabled: i || !n.trim(),
			onClick: () => void s(),
			children: "Send"
		})]
	});
}
function ht({ comments: e, config: t, compact: n = !1, defaultExpanded: r = !1, renderRootMeta: i, renderRootDetails: a, getThreadProps: o, renderBody: c, threadToMarkdown: u, onUpdateStatus: d, onClose: f, onAcknowledge: p, onAssign: m, extraAssignItems: h, onChecklistToggle: g, onDelete: _, onReply: v }) {
	let [y, b] = (0, M.useState)(null), [x, S] = (0, M.useState)(null), [C, w] = (0, M.useState)("thread"), [T, E] = (0, M.useState)(null), ee = B(e), te = Ne(e);
	if (ee.length === 0) return null;
	let D = ee.find((e) => e.id === x), ne = D ? [D, ...Fe(te.get(D.id) ?? [])] : [];
	function re() {
		S(null), E(null), w("thread");
	}
	function O(e) {
		return {
			...d ? { onUpdateStatus: (t) => d(e, t) } : {},
			...f ? { onClose: () => f(e) } : {},
			...p ? { onAcknowledge: () => p(e) } : {},
			...m ? { onAssign: (t) => m(e, t) } : {},
			...h ? { extraAssignItems: h } : {},
			...g ? { onChecklistToggle: (t) => g(e, t) } : {},
			..._ ? { onDelete: () => _(e) } : {}
		};
	}
	return /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsx)("div", {
		className: "space-y-2",
		children: ee.map((e) => {
			let l = Fe(te.get(e.id) ?? []), d = [e, ...l], f = o?.(e) ?? {};
			return /* @__PURE__ */ (0, N.jsxs)("div", {
				"data-testid": "comment-thread-block",
				"data-comment-id": e.id,
				...f,
				className: s("space-y-1.5 p-1", f.className),
				children: [
					i ? /* @__PURE__ */ (0, N.jsx)("div", {
						className: "px-1",
						children: i(e)
					}) : null,
					/* @__PURE__ */ (0, N.jsx)(Z, {
						comment: e,
						config: t,
						compact: n,
						defaultExpanded: r,
						onMaximize: () => S(e.id),
						...u ? { onCopy: () => ie(u(d)) } : {},
						...c ? { renderBody: c } : {},
						...O(e.id),
						...v ? { onReply: () => b(e.id) } : {}
					}),
					a?.(e),
					l.length > 0 && /* @__PURE__ */ (0, N.jsx)("div", {
						className: "space-y-1.5",
						children: l.map((i) => /* @__PURE__ */ (0, N.jsx)(Z, {
							comment: i,
							config: t,
							compact: n,
							defaultExpanded: r,
							onMaximize: () => S(e.id),
							...u ? { onCopy: () => ie(u(d)) } : {},
							...c ? { renderBody: c } : {},
							...O(i.id)
						}, i.id))
					}),
					v && y === e.id && /* @__PURE__ */ (0, N.jsx)(mt, {
						onCancel: () => b(null),
						onSubmit: async (t) => {
							await v(e, t), b(null);
						}
					})
				]
			}, e.id);
		})
	}), D && /* @__PURE__ */ (0, N.jsx)(l, {
		open: !0,
		onClose: re,
		title: "Comment thread",
		size: "lg",
		children: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "space-y-3",
			"data-testid": "comment-thread-modal",
			children: [u && /* @__PURE__ */ (0, N.jsx)(ve, {
				tabs: [{
					id: "thread",
					label: "Thread"
				}, {
					id: "markdown",
					label: "Markdown"
				}],
				value: C,
				onChange: (e) => w(e === "markdown" ? "markdown" : "thread")
			}), C === "markdown" && u ? /* @__PURE__ */ (0, N.jsx)(pe, { text: u(ne) }) : /* @__PURE__ */ (0, N.jsx)("div", {
				className: "space-y-2",
				children: ne.map((e) => /* @__PURE__ */ (0, N.jsxs)("div", {
					className: "space-y-1.5",
					children: [
						/* @__PURE__ */ (0, N.jsx)(Z, {
							comment: e,
							config: t,
							defaultExpanded: !0,
							...u ? { onCopy: () => ie(u(ne)) } : {},
							...c ? { renderBody: c } : {},
							...O(e.id),
							...v ? { onReply: () => E(e.id) } : {}
						}),
						e.id === D.id && a?.(D),
						v && T === e.id && /* @__PURE__ */ (0, N.jsx)(mt, {
							onCancel: () => E(null),
							onSubmit: async (e) => {
								await v(D, e), E(null);
							}
						})
					]
				}, e.id))
			})]
		})
	})] });
}
// ../../../clicky-ui/packages/ui/src/comments/CommentThread.tsx
function gt({ comments: e, config: t, anchor: n, compact: r = !1, autoFocusComposer: i = !1, defaultExpanded: a = !1, composerPlaceholder: o, hideComposer: c = !1, createActions: l, renderRootMeta: u, renderRootDetails: d, renderBody: f, threadToMarkdown: p, extraAssignItems: m, onCreate: h, onReply: g, onUpdateStatus: _, onClose: v, onAcknowledge: y, onAssign: b, onDelete: x, onChecklistToggle: S, onMention: C }) {
	let w = ze(e, t, {
		..._ ? { onUpdateStatus: _ } : {},
		...v ? { onClose: v } : {},
		...y ? { onAcknowledge: y } : {},
		...b ? { onAssign: b } : {},
		...x ? { onDelete: x } : {},
		...S ? { onChecklistToggle: S } : {},
		...g ? { onReply: g } : {},
		...C ? { onMention: C } : {}
	});
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		className: s(r && "p-2"),
		children: [/* @__PURE__ */ (0, N.jsx)(ht, {
			comments: e,
			config: t,
			compact: r,
			defaultExpanded: a,
			...u ? { renderRootMeta: u } : {},
			...d ? { renderRootDetails: d } : {},
			...f ? { renderBody: f } : {},
			...p ? { threadToMarkdown: p } : {},
			...m ? { extraAssignItems: m } : {},
			...w
		}), !c && h && /* @__PURE__ */ (0, N.jsx)("div", {
			className: "pt-1",
			children: /* @__PURE__ */ (0, N.jsx)(Xe, {
				config: t,
				anchor: n ?? null,
				autoFocus: i,
				...o ? { placeholder: o } : {},
				onCreate: h,
				...l ? { createActions: l } : {},
				...C ? { onMention: C } : {}
			})
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/findings-comments.ts
function Q(e) {
	return `line:${e}`;
}
function _t(e, t) {
	return `L${e}-${t}`;
}
var vt = /^line:(\d+)$/, yt = /^L(\d+)-(\d+)$/;
function bt(e) {
	let t = vt.exec(e);
	if (!t) throw Error(`file-findings: expected a "line:<n>" anchor, got ${JSON.stringify(e)}`);
	return Number(t[1]);
}
var xt = (e, t) => {
	if (e === "__document__" || yt.test(e)) return t.includes(e) ? e : null;
	let n = bt(e), r = t.flatMap((e) => {
		let t = yt.exec(e);
		return t ? [{
			anchor: e,
			start: Number(t[1]),
			end: Number(t[2])
		}] : [];
	}), i = r.filter((e) => e.start <= n && n <= e.end).sort((e, t) => e.end - e.start - (t.end - t.start) || t.start - e.start)[0];
	if (i) return i.anchor;
	let a = r.filter((e) => e.start <= n).sort((e, t) => t.start - e.start)[0];
	return a ? a.anchor : t.includes("__document__") ? Oe : null;
}, St = {
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
}, Ct = "1970-01-01T00:00:00.000Z", wt = 6;
function Tt(e) {
	return e.map((e, t) => ({
		id: `finding-${String(t).padStart(wt, "0")}`,
		body: e.rule ? `${e.message}\n\n\`${e.rule}\`` : e.message,
		createdAt: Ct,
		author: {
			name: e.source,
			kind: "agent"
		},
		status: e.severity,
		anchor: Q(e.line),
		facets: {
			severity: e.severity,
			source: e.source,
			kind: e.kind
		}
	}));
}
// ../../../clicky-ui/packages/ui/src/file-findings/LazyAnnotatedCode.tsx
var Et = (0, M.lazy)(() => import("./AnnotatedCode-Dd5OqOyu.js").then((e) => ({ default: e.AnnotatedCode })));
function Dt(e) {
	return /* @__PURE__ */ (0, N.jsx)(M.Suspense, {
		fallback: /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
			children: [/* @__PURE__ */ (0, N.jsx)(_, {
				icon: ce,
				className: "mr-2 animate-spin"
			}), "Loading editor"]
		}),
		children: /* @__PURE__ */ (0, N.jsx)(Et, { ...e })
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/mdx-source.ts
var Ot = /^\s{0,3}(`{3,}|~{3,})/, kt = /^(import|export)\s/;
function At(e) {
	let t = 0;
	for (let n of e) "{[(".includes(n) ? t += 1 : "}])".includes(n) && --t;
	return t;
}
function jt(e) {
	let t = e.split("\n"), n = t.slice(), r = 0;
	if (t[0] === "---") {
		let e = t.indexOf("---", 1);
		if (e !== -1) for (; r <= e; r += 1) n[r] = "";
	}
	let i = null;
	for (; r < t.length; r += 1) {
		let e = t[r], a = Ot.exec(e);
		if (i) {
			a && a[1][0] === i[0] && a[1].length >= i.length && (i = null);
			continue;
		}
		if (a) {
			i = a[1];
			continue;
		}
		if (!kt.test(e)) continue;
		let o = At(e);
		for (n[r] = ""; o > 0 && r + 1 < t.length;) r += 1, o += At(t[r]), n[r] = "";
	}
	return n.join("\n");
}
var Mt = "[data-source-start]", Nt = /^h([1-6])$/, Pt = 2;
function Ft(e) {
	return `q:${JSON.stringify([
		e.path,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function It(e) {
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
function Lt(e) {
	let t = [];
	return [...e.querySelectorAll(Mt)].filter((e) => !e.parentElement?.closest(Mt)).map((e) => {
		let n = Nt.exec(e.tagName.toLowerCase())?.[1];
		if (n) {
			for (; t.length > 0 && (t.at(-1)?.level ?? 0) >= Number(n);) t.pop();
			let r = Bt(e).replace(/\s+/g, " ").trim().slice(0, 200);
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
var Rt = (e) => Lt(e);
function zt(e) {
	let t = document.createTreeWalker(e, NodeFilter.SHOW_TEXT), n = [];
	for (let e = t.nextNode(); e; e = t.nextNode()) e.parentElement?.closest("[data-comment-author]") || n.push(e);
	return n;
}
function Bt(e) {
	return zt(e).map((e) => e.data).join("");
}
function Vt(e, t, n) {
	let r = document.createRange(), i = 0, a = !1;
	for (let o of zt(e)) {
		if (!a && t < i + o.length && (r.setStart(o, t - i), a = !0), a && n <= i + o.length) return r.setEnd(o, n - i), r;
		i += o.length;
	}
	throw Error(`quote-anchor: offsets ${t}-${n} lie outside the block`);
}
function Ht(e) {
	return e.trim().split(/\s+/).filter(Boolean).map((e) => e.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("\\s*");
}
function Ut(e, t, n) {
	let r = n.prefix && t.prefix.trim() ? `(?<=${Ht(t.prefix)}\\s*)` : "", i = n.suffix && t.suffix.trim() ? `(?=\\s*${Ht(t.suffix)})` : "", a = RegExp(`${r}${Ht(t.quote)}${i}`, "g");
	return e.flatMap(({ block: e }) => [...Bt(e).matchAll(a)].map((t) => ({
		block: e,
		start: t.index,
		end: t.index + t[0].length
	})));
}
var Wt = (e, t) => e.length === t.length && e.every((e, n) => e === t[n]), Gt = [
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
function Kt(e, t) {
	let n = Rt(e), r = n.filter((e) => Wt(e.path, t.path));
	for (let e of Gt) {
		let n = Ut(r, t, e);
		if (n.length === 1) return Vt(n[0].block, n[0].start, n[0].end);
	}
	let i = Ut(n, t, Gt[0]);
	return i.length === 1 ? Vt(i[0].block, i[0].start, i[0].end) : null;
}
function qt(e, t) {
	let n = Rt(e), r = n.find(({ block: e }) => e.contains(t.startContainer) && e.contains(t.endContainer));
	if (!r) return null;
	let i = Bt(r.block), a = document.createRange();
	a.setStart(r.block, 0), a.setEnd(t.startContainer, t.startOffset);
	let o = en(t), s = en(a).length + (o.length - o.trimStart().length), c = en(a).length + o.trimEnd().length;
	if (c <= s) return null;
	let l = i.slice(0, s).split(/\s+/).filter(Boolean), u = i.slice(c).split(/\s+/).filter(Boolean), d = n.filter((e) => Wt(e.path, r.path));
	for (let e = Pt;; e++) {
		let t = {
			path: r.path,
			prefix: l.slice(-e).join(" "),
			quote: i.slice(s, c),
			suffix: u.slice(0, e).join(" ")
		}, n = Ut(d, t, Gt[0]);
		if (n.length === 1 && n[0].block === r.block && n[0].start === s) return t;
		if (e >= l.length && e >= u.length) return null;
	}
}
var Jt = /[*_`|>#~]|(?:^|\s)[-+](?=\s)/g;
function Yt(e) {
	return Ht(e.replace(Jt, " ")).replaceAll("\\s*", "[\\s*_`|>#~]*");
}
function Xt(e, t) {
	return Rt(e).map(({ block: e }) => {
		let n = Number(e.dataset.sourceStart), r = Number(e.dataset.sourceEnd);
		return {
			block: e,
			first: n,
			last: r,
			text: t.slice(n - 1, r).join("\n")
		};
	});
}
var Zt = (e, t) => [...e.matchAll(new RegExp(t, "g"))];
function Qt(e, t, n) {
	let r = t[n.line - 1]?.slice(n.start, n.end) ?? "";
	if (!r.replace(Jt, " ").trim()) return null;
	let i = Xt(e, t).find((e) => n.line >= e.first && n.line <= e.last);
	if (!i) return null;
	let a = t.slice(i.first - 1, n.line - 1).reduce((e, t) => e + t.length + 1, 0) + n.start, o = Zt(i.text, Yt(r)).findIndex((e) => e.index + e[0].length > a && e.index < a + r.length), s = o < 0 ? void 0 : Zt(Bt(i.block), Ht(r.replace(Jt, " ")))[o];
	return s ? qt(e, Vt(i.block, s.index, s.index + s[0].length)) : null;
}
function $t(e, t, n) {
	let r = Kt(e, n), i = r && Xt(e, t).find(({ block: e }) => e.contains(r.startContainer));
	if (!r || !i) return [];
	let a = document.createRange();
	a.setStart(i.block, 0), a.setEnd(r.startContainer, r.startOffset);
	let o = en(a).length, s = Zt(Bt(i.block), Ht(n.quote)).findIndex((e) => e.index === o), c = s < 0 ? void 0 : Zt(i.text, Yt(n.quote))[s];
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
function en(e) {
	let t = e.cloneContents();
	for (let e of t.querySelectorAll("[data-comment-author]")) e.remove();
	return t.textContent ?? "";
}
function tn(e) {
	let t = e.path.at(-1);
	return `${t ? `${t} › ` : ""}“${e.quote}”`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/geometry.ts
var nn = "[data-source-start]";
function rn(e) {
	return (e instanceof Element ? e : e?.parentElement ?? null)?.closest(nn) ?? null;
}
function an({ origin: e, markdown: t, anchors: n, mode: r }) {
	let i = e.getBoundingClientRect(), a = (e) => ({
		top: e.top - i.top,
		left: e.left - i.left,
		width: e.width,
		height: e.height
	}), o = /* @__PURE__ */ new Map();
	for (let e of t.querySelectorAll(nn)) {
		let t = _t(Number(e.dataset.sourceStart), Number(e.dataset.sourceEnd));
		o.has(t) || o.set(t, e);
	}
	let s = [];
	for (let e of n) {
		if (e.startsWith("q:")) {
			let n = Kt(t, It(e));
			if (!n) continue;
			let i = [...n.getClientRects()].filter((e) => e.width > 0).map(a), o = rn(n.startContainer), c = r === "text" ? i : o ? [a(o.getBoundingClientRect())] : [];
			c.length > 0 && s.push({
				anchor: e,
				boxes: c,
				at: i[0]?.top ?? c[0].top
			});
			continue;
		}
		let n = xt(e, [...o.keys()]), i = n ? o.get(n) : void 0;
		if (i) {
			let t = document.createRange();
			t.selectNodeContents(i);
			let n = on([...t.getClientRects()].filter((e) => e.width > 0).map(a)), r = n.length > 0 ? n : [a(i.getBoundingClientRect())];
			s.push({
				anchor: e,
				boxes: r,
				at: r[0].top
			});
		}
	}
	return new Map(s.sort((e, t) => e.at - t.at).map(({ anchor: e, boxes: t }) => [e, t]));
}
function on(e) {
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
function sn({ cards: e, focused: t, gap: n = 8 }) {
	let r = [...e].sort((e, t) => e.desired - t.desired || e.order - t.order), i = r.map((e) => e.desired), a = r.findIndex((e) => e.anchor === t);
	if (a >= 0) for (let e = a - 1; e >= 0; e--) i[e] = Math.min(i[e], i[e + 1] - n - r[e].height);
	for (let e = 0; e < r.length; e++) {
		let t = e === 0 ? 0 : i[e - 1] + r[e - 1].height + n;
		i[e] = Math.max(i[e], t);
	}
	return new Map(r.map((e, t) => [e.anchor, i[t]]));
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/hooks.ts
var cn = 640;
function ln() {
	let [e, t] = (0, M.useState)({}), n = (0, M.useRef)(null), r = (0, M.useRef)(/* @__PURE__ */ new Map());
	(0, M.useEffect)(() => () => n.current?.disconnect(), []);
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
function un({ origin: e, markdown: t, anchors: n, mode: r, visible: i }) {
	let [a, o] = (0, M.useState)(/* @__PURE__ */ new Map()), [s, c] = (0, M.useState)(0), [l, u] = (0, M.useState)(0);
	return (0, M.useLayoutEffect)(() => {
		let a = e.current, s = t.current;
		if (!a || !s || !i) return;
		let l = () => {
			o(an({
				origin: a,
				markdown: s,
				anchors: n,
				mode: r
			})), c(a.offsetHeight), u(a.offsetWidth);
		};
		l();
		let d = new MutationObserver(l);
		d.observe(s, {
			childList: !0,
			subtree: !0,
			characterData: !0
		});
		let f = new ResizeObserver(l);
		return f.observe(s), () => {
			d.disconnect(), f.disconnect();
		};
	}, [
		e,
		t,
		n,
		r,
		i
	]), {
		boxes: a,
		height: s,
		width: l
	};
}
function dn(e) {
	let [t, n] = (0, M.useState)(!1);
	return (0, M.useLayoutEffect)(() => {
		let t = e.current;
		if (!t) return;
		let r = new ResizeObserver(() => n(t.offsetWidth < cn));
		return r.observe(t), () => r.disconnect();
	}, [e]), t;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/marks.ts
var fn = {
	error: 3,
	warning: 2,
	info: 1
}, pn = {
	error: {
		icon: _e,
		tone: "text-red-500",
		stroke: "#ef4444",
		tint: "bg-red-500/10"
	},
	warning: {
		icon: a,
		tone: "text-amber-500",
		stroke: "#f59e0b",
		tint: "bg-amber-500/10"
	},
	info: {
		icon: o,
		tone: "text-sky-500",
		stroke: "#0ea5e9",
		tint: "bg-sky-500/10"
	}
};
function mn(e) {
	if (e === "error" || e === "warning" || e === "info") return e;
	throw Error(`lint finding has severity ${JSON.stringify(e)}, expected error, warning or info`);
}
var hn = (e) => `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='6' height='3' viewBox='0 0 6 3'%3E%3Cpath d='M0 2.5 1.5 0.5 3 2.5 4.5 0.5 6 2.5' fill='none' stroke='${encodeURIComponent(e)}' stroke-width='1'/%3E%3C/svg%3E")`, gn = (e) => ({
	backgroundImage: hn(e),
	backgroundRepeat: "repeat-x",
	backgroundPosition: "bottom",
	backgroundSize: "6px 3px",
	paddingBottom: 2
}), _n = (e, t) => fn[mn(t.severity)] - fn[mn(e.severity)] || e.line - t.line, vn = 3;
function yn({ finding: e, lines: t, onOpen: n, openLabel: r }) {
	let [i, a] = (0, M.useState)(!1), o = pn[mn(e.severity)], c = Math.max(1, e.line - vn), l = Math.min(t.length, e.line + vn), u = i ? t.slice(c - 1, l).map((e, t) => ({
		line: c + t,
		text: e
	})) : [];
	return /* @__PURE__ */ (0, N.jsxs)("li", {
		"data-problem-line": e.line,
		children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex items-start gap-1 px-1.5 py-1.5 text-xs hover:bg-muted/50",
			children: [/* @__PURE__ */ (0, N.jsx)("button", {
				type: "button",
				"aria-expanded": i,
				"aria-label": i ? "Hide context" : "Show context",
				onClick: () => a((e) => !e),
				className: "flex size-5 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground",
				children: /* @__PURE__ */ (0, N.jsx)(_, {
					icon: i ? S : b,
					className: "text-sm"
				})
			}), /* @__PURE__ */ (0, N.jsxs)("button", {
				type: "button",
				onClick: () => n(e),
				className: "grid min-w-0 flex-1 grid-cols-[auto_2rem_minmax(0,1fr)_auto] items-center gap-x-2 gap-y-0.5 text-left",
				title: r,
				children: [
					/* @__PURE__ */ (0, N.jsx)(_, {
						icon: o.icon,
						className: s("text-sm", o.tone)
					}),
					/* @__PURE__ */ (0, N.jsx)("span", {
						"data-problem-ln": !0,
						className: "text-right font-mono text-[11px] tabular-nums text-muted-foreground",
						children: e.line
					}),
					/* @__PURE__ */ (0, N.jsx)("span", {
						className: "truncate",
						children: e.message
					}),
					/* @__PURE__ */ (0, N.jsxs)("span", {
						className: "font-mono text-[11px] text-muted-foreground",
						children: [
							e.source,
							" ",
							e.rule
						]
					}),
					!i && /* @__PURE__ */ (0, N.jsx)("span", {
						"data-problem-context": "line",
						className: "col-span-2 col-start-3 truncate font-mono text-[11px] text-muted-foreground",
						children: t[e.line - 1]?.trim() || " "
					})
				]
			})]
		}), i && /* @__PURE__ */ (0, N.jsx)("ol", {
			"data-problem-context": "block",
			className: "mb-2 ml-8 mr-3 overflow-x-auto rounded border border-border bg-muted/30 py-1 font-mono text-[11px] leading-5",
			children: u.map(({ line: t, text: n }) => /* @__PURE__ */ (0, N.jsxs)("li", {
				className: "grid grid-cols-[2.5rem_1fr] pr-2",
				children: [/* @__PURE__ */ (0, N.jsx)("span", {
					className: "select-none pr-2 text-right tabular-nums text-muted-foreground",
					children: t
				}), /* @__PURE__ */ (0, N.jsx)("span", {
					className: "whitespace-pre",
					children: /* @__PURE__ */ (0, N.jsx)("span", {
						style: t === e.line ? gn(o.stroke) : void 0,
						children: n || " "
					})
				})]
			}, t))
		})]
	});
}
function bn({ source: e, findings: t, onOpen: n, openLabel: r }) {
	let i = (0, M.useMemo)(() => e.split("\n"), [e]);
	return t.length === 0 ? /* @__PURE__ */ (0, N.jsx)("p", {
		className: "p-6 text-center text-xs text-muted-foreground",
		children: "No problems in this file."
	}) : /* @__PURE__ */ (0, N.jsx)("ul", {
		className: "divide-y divide-border",
		children: [...t].sort(_n).map((e) => /* @__PURE__ */ (0, N.jsx)(yn, {
			finding: e,
			lines: i,
			onOpen: n,
			openLabel: r
		}, `${e.line}:${e.rule}:${e.message}`))
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/blame.ts
var xn = (e) => /^0+$/.test(e.sha);
function Sn(e, t) {
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
var Cn = 864e5;
function wn(e, t = /* @__PURE__ */ new Date()) {
	let n = Math.floor((t.getTime() - Date.parse(e)) / Cn);
	return n < 1 ? "today" : n < 30 ? `${n}d` : n < 365 ? `${Math.floor(n / 30)}mo` : `${Math.floor(n / 365)}y`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/source-marks.ts
var Tn = (e) => JSON.parse(e.dataset.sourceComment);
function En(e, t, n, r) {
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
function Dn(e) {
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
function On(e) {
	let t = Dn(e);
	return !t || t.start.line !== t.end.line ? null : {
		line: t.start.line,
		start: t.start.column,
		end: t.end.column
	};
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/SourceView.tsx
function kn({ blame: e }) {
	if (!e) return /* @__PURE__ */ (0, N.jsx)("span", {});
	let { commit: t, first: n, recency: r } = e, i = xn(t);
	return /* @__PURE__ */ (0, N.jsxs)("span", {
		"data-blame-sha": n ? t.sha : void 0,
		title: i ? "Not committed yet" : `${t.sha.slice(0, 12)} · ${t.author} · ${t.date.slice(0, 10)}\n${t.summary}`,
		className: s("relative flex min-w-0 items-start gap-1.5 pl-2 pr-2 font-sans text-[11px] text-muted-foreground", n && "border-t border-border/60"),
		children: [/* @__PURE__ */ (0, N.jsx)("span", {
			"aria-hidden": !0,
			className: s("absolute inset-y-0 left-0 w-[3px]", i ? "bg-amber-500" : "bg-sky-500"),
			style: { opacity: i ? 1 : .15 + .85 * r }
		}), n && (i ? /* @__PURE__ */ (0, N.jsx)("span", {
			className: "italic",
			children: "Uncommitted"
		}) : /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "shrink-0 font-mono text-foreground/70",
				children: t.sha.slice(0, 7)
			}),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "min-w-0 flex-1 truncate",
				children: t.author
			}),
			/* @__PURE__ */ (0, N.jsx)("span", {
				className: "shrink-0 tabular-nums",
				children: wn(t.date)
			})
		] }))]
	});
}
function An({ source: e, tokens: t, blame: n, findings: r, showFindings: i, marks: a, comments: o, selected: c, onSelect: l, onFocus: u, renderFindingActions: d }) {
	return /* @__PURE__ */ (0, N.jsx)("ol", {
		className: "py-2 font-mono text-xs leading-5",
		children: e.split("\n").map((e, f) => {
			let m = f + 1, h = r.filter((e) => e.line === m).sort(_n), g = h[0] ? pn[mn(h[0].severity)] : null, v = a.filter((e) => e.spans.at(-1)?.line === m), y = i || c === Q(m) ? h : [];
			return /* @__PURE__ */ (0, N.jsxs)("li", {
				"data-source-line": m,
				className: s("grid pr-3", n ? "grid-cols-[13rem_1.25rem_2.5rem_1fr]" : "grid-cols-[1.25rem_2.5rem_1fr]"),
				children: [
					n && /* @__PURE__ */ (0, N.jsx)(kn, { blame: n.get(m) }),
					g && !i ? /* @__PURE__ */ (0, N.jsx)("button", {
						type: "button",
						"data-source-mark": m,
						"aria-label": `${h.length} ${h.length === 1 ? "finding" : "findings"} on line ${m}`,
						"aria-pressed": c === Q(m),
						onClick: () => c === Q(m) ? u(null) : l(h[0]),
						className: s("flex h-5 items-center justify-center rounded-sm hover:bg-muted", g.tone),
						children: /* @__PURE__ */ (0, N.jsx)(_, {
							icon: g.icon,
							className: "text-[13px]"
						})
					}) : /* @__PURE__ */ (0, N.jsx)("span", {
						className: s("flex h-5 items-center justify-center", g?.tone),
						children: g && /* @__PURE__ */ (0, N.jsx)(_, {
							icon: g.icon,
							className: "text-[13px]"
						})
					}),
					/* @__PURE__ */ (0, N.jsx)("span", {
						className: "select-none pr-3 text-right text-muted-foreground tabular-nums",
						children: m
					}),
					/* @__PURE__ */ (0, N.jsxs)("span", {
						className: "min-w-0 whitespace-pre-wrap break-words",
						children: [
							/* @__PURE__ */ (0, N.jsx)("span", {
								"data-source-text": m,
								style: g ? gn(g.stroke) : void 0,
								children: e ? En(e, m, a, t?.[f]).map((e) => {
									let t = e.token ? p(e.token) : void 0;
									return e.anchors.length === 0 ? /* @__PURE__ */ (0, N.jsx)("span", {
										style: t,
										children: e.text
									}, e.start) : /* @__PURE__ */ (0, N.jsx)("mark", {
										"data-source-comment": JSON.stringify(e.anchors),
										onClick: () => window.getSelection()?.isCollapsed && u(e.anchors[0]),
										style: t,
										className: s("cursor-pointer rounded-sm text-inherit", e.anchors.includes(c ?? "") ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25"),
										children: e.text
									}, e.start);
								}) : " "
							}),
							v.map(({ anchor: e }) => {
								let t = Pe(o, e), n = B(t)[0];
								return n ? /* @__PURE__ */ (0, N.jsxs)("button", {
									type: "button",
									"data-source-thread": e,
									onClick: () => u(e),
									className: "flex w-full items-center gap-1.5 text-left font-sans text-[11px] text-amber-600 hover:underline [[data-theme=dark]_&]:text-amber-400",
									children: [
										/* @__PURE__ */ (0, N.jsx)(Ve, {
											author: n.author ?? null,
											size: "xs",
											bare: !0
										}),
										/* @__PURE__ */ (0, N.jsx)("span", {
											className: "shrink-0 font-medium text-foreground",
											children: n.author?.name
										}),
										/* @__PURE__ */ (0, N.jsx)("span", {
											className: "truncate text-muted-foreground",
											children: n.body
										}),
										t.length > 1 && /* @__PURE__ */ (0, N.jsxs)("span", {
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
							y.map((e) => /* @__PURE__ */ (0, N.jsxs)("span", {
								"data-source-finding": e.line,
								className: "flex items-center gap-2",
								children: [/* @__PURE__ */ (0, N.jsxs)("button", {
									type: "button",
									"aria-pressed": c === Q(e.line),
									onClick: () => l(e),
									className: s("flex min-w-0 flex-1 items-center gap-1.5 text-left font-sans text-[11px] hover:underline aria-pressed:font-semibold aria-pressed:underline", pn[mn(e.severity)].tone),
									children: [
										/* @__PURE__ */ (0, N.jsx)(_, { icon: pn[mn(e.severity)].icon }),
										/* @__PURE__ */ (0, N.jsx)("span", {
											className: "text-foreground",
											children: e.message
										}),
										/* @__PURE__ */ (0, N.jsxs)("span", {
											className: "text-muted-foreground",
											children: [
												e.source,
												" ",
												e.rule
											]
										})
									]
								}), d?.(e)]
							}, `${e.rule}:${e.message}`))
						]
					})
				]
			}, m);
		})
	});
}
var jn = 24, Mn = 60;
function Nn(e) {
	return `s:${JSON.stringify([
		e.line,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function Pn(e) {
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
function Fn(e) {
	let t = [0];
	for (let n = e.indexOf("\n"); n !== -1; n = e.indexOf("\n", n + 1)) t.push(n + 1);
	return t;
}
function In(e, t, n) {
	let r = t[n.line - 1];
	if (r === void 0) throw Error(`line ${n.line} is outside the ${t.length}-line source`);
	let i = t[n.line] === void 0 ? e.length : t[n.line] - 1;
	return Math.min(r + n.column, i);
}
function Ln(e, t) {
	let n = 1;
	for (; n < e.length && e[n] <= t;) n++;
	return n;
}
function Rn(e, t, n) {
	let r = Fn(e), i = In(e, r, t), a = In(e, r, n), o = e.slice(i, a);
	return o.trim() ? {
		line: Ln(r, i),
		prefix: e.slice(Math.max(0, i - jn), i),
		quote: o,
		suffix: e.slice(a, a + jn)
	} : null;
}
function zn(e, t) {
	let n = t.prefix + t.quote + t.suffix, r = Fn(e), i = null, a = Infinity;
	for (let o = e.indexOf(n); o !== -1; o = e.indexOf(n, o + 1)) {
		let e = o + t.prefix.length, n = Math.abs(Ln(r, e) - t.line);
		n < a && (i = e, a = n);
	}
	if (i === null) return [];
	let o = i + t.quote.length, s = [];
	for (let t = Ln(r, i); t <= r.length && r[t - 1] < o; t++) {
		let n = r[t - 1], a = r[t] === void 0 ? e.length : r[t] - 1, c = {
			line: t,
			start: Math.max(i, n) - n,
			end: Math.min(o, a) - n
		};
		c.end > c.start && s.push(c);
	}
	return s;
}
function Bn(e) {
	let t = e.quote.replace(/\s+/g, " ").trim();
	return `Line ${e.line} · “${t.length > Mn ? `${t.slice(0, Mn)}…` : t}”`;
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/InlineThread.tsx
var Vn = 340, Hn = 6;
function Un({ anchor: e, label: t, boxes: n, containerWidth: r, thread: i, config: a, callbacks: o, pinned: c, onPin: l, onClose: u, onPointerEnter: d, onPointerLeave: f, seamless: p = !1, renderRootDetails: m }) {
	(0, M.useEffect)(() => {
		let e = (e) => e.key === "Escape" && u();
		return window.addEventListener("keydown", e), () => window.removeEventListener("keydown", e);
	}, [u]);
	let h = n[0], g = n.at(-1);
	if (!h || !g) return null;
	let v = Math.min(Vn, r), y = Math.max(0, Math.min(h.left, r - v)), b = !!o.onCreate;
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		"data-inline-thread": e,
		role: "dialog",
		"aria-label": `Comments on ${t}`,
		onMouseEnter: d,
		onMouseLeave: f,
		onMouseDown: l,
		className: s("absolute z-30 bg-card", p ? "rounded-md font-sans shadow-[0_4px_16px_rgba(0,0,0,0.12)] ring-1 ring-border/60" : "rounded-lg border shadow-lg", !p && (c ? "border-primary/50" : "border-border")),
		style: {
			top: g.top + g.height + (p ? 2 : Hn),
			left: y,
			width: v
		},
		children: [!p && /* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, N.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: t
				}),
				!c && /* @__PURE__ */ (0, N.jsx)("span", { children: "click to pin" }),
				/* @__PURE__ */ (0, N.jsx)("button", {
					type: "button",
					"aria-label": "Close comments",
					onClick: u,
					className: "rounded p-0.5 hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
					children: /* @__PURE__ */ (0, N.jsx)(_, {
						icon: E,
						className: "text-xs"
					})
				})
			]
		}), /* @__PURE__ */ (0, N.jsx)("div", {
			className: "max-h-[360px] overflow-y-auto",
			children: /* @__PURE__ */ (0, N.jsx)(gt, {
				comments: i,
				config: a,
				anchor: e,
				compact: !0,
				defaultExpanded: !0,
				hideComposer: !b || i.length > 0 && !c,
				autoFocusComposer: i.length === 0,
				composerPlaceholder: i.length > 0 ? "Add another comment…" : "Add a comment…",
				...o,
				...m ? { renderRootDetails: m } : {}
			})
		})]
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/InlineAuthors.tsx
function Wn({ markdown: e, comments: t, focused: n, onFocus: r }) {
	let [i, a] = (0, M.useState)(/* @__PURE__ */ new Map());
	(0, M.useLayoutEffect)(() => {
		let n = e.current;
		if (!n) return;
		let r = /* @__PURE__ */ new Map(), i = () => {
			let e = !1;
			for (let i of B(t)) {
				let t = i.anchor;
				if (!t?.startsWith("q:") || r.get(i.id)?.isConnected) continue;
				let a = Kt(n, It(t));
				if (!a) continue;
				let o = document.createElement("span");
				o.dataset.commentAuthor = i.id, o.className = "ml-1 inline-flex max-w-full align-middle leading-none", o.contentEditable = "false", a.collapse(!1), a.insertNode(o), r.set(i.id, o), e = !0;
			}
			e && a(new Map(r));
		};
		i();
		let o = new MutationObserver(i);
		return o.observe(n, {
			childList: !0,
			subtree: !0
		}), () => {
			o.disconnect();
			for (let e of r.values()) e.remove();
		};
	}, [e, t]);
	let o = Ne(t);
	return B(t).map((e) => {
		let t = i.get(e.id), a = e.anchor;
		if (!t || !a) return null;
		let c = o.get(e.id)?.length ?? 0;
		return (0, Ue.createPortal)(/* @__PURE__ */ (0, N.jsxs)("button", {
			type: "button",
			"aria-label": `Open comment by ${R(e.author ?? null)}`,
			"aria-pressed": n === a,
			title: e.body,
			className: s("inline-flex max-w-[min(40ch,100%)] items-center gap-1 rounded-full border border-border bg-muted py-px pl-px pr-2 text-xs font-normal text-muted-foreground ring-offset-background hover:text-foreground hover:ring-2 hover:ring-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary", n === a && "text-foreground ring-2 ring-primary"),
			onClick: () => r(a),
			children: [
				/* @__PURE__ */ (0, N.jsx)(Ve, {
					author: e.author ?? null,
					size: "xs",
					bare: !0
				}),
				/* @__PURE__ */ (0, N.jsx)("span", {
					className: "truncate",
					children: Le(e.body)
				}),
				c > 0 && /* @__PURE__ */ (0, N.jsxs)("span", {
					className: "shrink-0 tabular-nums",
					children: ["· ", c]
				})
			]
		}), t, e.id);
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/CommentDraft.tsx
var Gn = {
	width: 104,
	height: 28
}, Kn = 320, qn = 190, Jn = 6;
function Yn({ anchor: e, label: t, position: n, containerWidth: r, expanded: i, config: a, onCreate: o, onExpand: c, onClose: l }) {
	(0, M.useEffect)(() => {
		if (!i) return;
		let e = (e) => e.key === "Escape" && l();
		return window.addEventListener("keydown", e), () => window.removeEventListener("keydown", e);
	}, [i, l]);
	let u = n.top - Gn.height - 8, d = n.top >= qn, f = Math.min(Kn, r), p = Math.max(f / 2, Math.min(n.left, r - f / 2)), m = i ? {
		top: d ? u + Gn.height : n.bottom + Jn,
		left: p,
		width: f,
		transform: `translate(-50%, ${d ? "-100%" : "0"})`
	} : {
		top: u,
		left: n.left,
		width: Gn.width,
		transform: "translate(-50%, 0)"
	};
	return /* @__PURE__ */ (0, N.jsx)("div", {
		"data-comment-draft": e,
		role: i ? "dialog" : void 0,
		"aria-label": i ? `New comment on ${t}` : void 0,
		onMouseDown: (e) => !i && e.preventDefault(),
		className: s("absolute z-30 overflow-hidden border bg-card shadow-md transition-[top,left,width,transform,border-radius,box-shadow] duration-200 ease-out", i ? "rounded-lg border-primary/50 shadow-lg" : "rounded-full border-border"),
		style: m,
		children: i ? /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, N.jsx)(_, {
					icon: h,
					className: "text-xs"
				}),
				/* @__PURE__ */ (0, N.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: t
				}),
				/* @__PURE__ */ (0, N.jsx)("button", {
					type: "button",
					"aria-label": "Discard comment",
					onClick: l,
					className: "rounded p-0.5 hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
					children: /* @__PURE__ */ (0, N.jsx)(_, {
						icon: E,
						className: "text-xs"
					})
				})
			]
		}), /* @__PURE__ */ (0, N.jsx)("div", {
			className: "p-2",
			children: /* @__PURE__ */ (0, N.jsx)(Xe, {
				config: a,
				anchor: e,
				collapsible: !1,
				autoFocus: !0,
				placeholder: "Add a comment…",
				onCreate: o
			})
		})] }) : /* @__PURE__ */ (0, N.jsxs)("button", {
			type: "button",
			onClick: c,
			className: "inline-flex w-full items-center justify-center gap-1 px-2.5 text-xs font-medium hover:bg-muted",
			style: { height: Gn.height },
			children: [/* @__PURE__ */ (0, N.jsx)(_, {
				icon: h,
				className: "text-sm"
			}), "Comment"]
		})
	});
}
// ../../../clicky-ui/packages/ui/src/file-findings/file-review/FileReview.tsx
var Xn = { statuses: ke }, Zn = 96, Qn = 450, $n = 300, er = 16, tr = [], nr = [], rr = {}, ir = Q(1).replace(/\d+$/, ""), $ = (e) => e.startsWith(ir), ar = (e) => /\.mdx?$/i.test(e), or = (e) => e.startsWith("q:") || e.startsWith("s:"), sr = (e) => ({
	anchor: Ft(e),
	label: tn(e)
});
function cr({ mode: e, active: t }) {
	return e === "block" ? s("border-l-[3px] border-amber-400", t ? "bg-amber-200/50 [[data-theme=dark]_&]:bg-amber-400/20" : "bg-amber-100/30 [[data-theme=dark]_&]:bg-amber-400/5") : s("rounded-sm", t ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25");
}
function lr(e, t) {
	let n = e.startsWith("q:") ? tn(It(e)) : e.startsWith("s:") ? Bn(Pn(e)) : e.replace("line:", "line ");
	return t ? n : `Text not found · ${n}`;
}
var ur = /[\p{L}\p{N}'’-]/u;
function dr(e, t) {
	let n = document.caretRangeFromPoint?.(e, t), r = n?.startContainer;
	if (!n || !r || r.nodeType !== Node.TEXT_NODE) return null;
	let i = r.textContent ?? "", a = n.startOffset, o = n.startOffset;
	for (; a > 0 && ur.test(i[a - 1]);) a--;
	for (; o < i.length && ur.test(i[o]);) o++;
	if (a === o) return null;
	let s = document.createRange();
	return s.setStart(r, a), s.setEnd(r, o), s;
}
var fr = [{
	id: "inline",
	label: "Inline comments"
}, {
	id: "margin",
	label: "Comment sidebar"
}], pr = [{
	id: "line",
	label: "Align to line"
}, {
	id: "stack",
	label: "Stacked list"
}], mr = 48;
function hr({ path: e, source: t, view: n, onViewChange: r, findings: i, comments: a, mode: o = "text", selectToComment: l = !1, authorFilter: d = !1, layout: p = "margin", onLayoutChange: m, onSourceChange: h, blame: g, line: v, onLineChange: y, renderFindingActions: b, toolbar: x, className: S }) {
	let C = i !== void 0, w = i ?? tr, T = ar(e);
	if (n === "problems" && !C) throw Error("FileReview: view \"problems\" needs `findings`");
	if (n === "preview" && !T) throw Error(`FileReview: view "preview" needs a markdown file, got ${e}`);
	let [E, ee] = (0, M.useState)(!1), [D, re] = (0, M.useState)(!0), ie = E && g !== void 0, k = h !== void 0 && !ie, oe = a?.items ?? nr, se = a?.callbacks ?? rr, ce = a?.config ?? Xn, A = (0, M.useMemo)(() => ({
		...ce,
		statuses: [...ce.statuses, ...St.statuses],
		facets: [...ce.facets ?? [], ...(St.facets ?? []).filter((e) => e.key !== "severity")]
	}), [ce]), de = l && !!se.onCreate, j = (0, M.useMemo)(() => e.endsWith(".mdx") ? jt(t) : t, [e, t]), me = (0, M.useMemo)(() => j.split("\n"), [j]), he = (0, M.useMemo)(() => ie ? Sn(g, me.length) : null, [
		ie,
		g,
		me.length
	]), ge = te(), _e = n === "source" && !k, [ye, be] = (0, M.useState)(null);
	(0, M.useEffect)(() => {
		if (!_e) return;
		let t = !0, n = ge === "dark" ? "github-dark" : "github-light";
		return u(j, {
			lang: f(e),
			theme: n
		}).then((e) => t && be({
			source: j,
			lines: e
		})), () => {
			t = !1;
		};
	}, [
		_e,
		j,
		e,
		ge
	]);
	let xe = ye?.source === j ? ye.lines : null, [P, Se] = (0, M.useState)("line"), [F, I] = (0, M.useState)(null), [Ce, Te] = (0, M.useState)(null), [Ee, L] = (0, M.useState)(null), [De, Oe] = (0, M.useState)("all"), [ke, Ae] = (0, M.useState)(void 0), [R, z] = (0, M.useState)(null), je = (0, M.useRef)(null), Me = (0, M.useRef)(null), Ne = (0, M.useRef)(null), Fe = (0, M.useRef)(null), Ie = (0, M.useRef)(null), V = (0, M.useRef)(null), H = dn(Fe), { heights: Le, refFor: Re } = ln(), ze = (0, M.useMemo)(() => Tt(w), [w]), W = (0, M.useMemo)(() => [...ze, ...oe], [ze, oe]), Ve = (0, M.useMemo)(() => new Map(ze.map((e, t) => [e.id, w[t]])), [ze, w]), He = (0, M.useMemo)(() => {
		if (b) return (e) => {
			let t = Ve.get(e.id);
			return t ? b(t) : null;
		};
	}, [b, Ve]), Ue = (0, M.useMemo)(() => W.filter((e) => e.parentId || De === "all" || (e.author?.kind ?? "user") === De), [W, De]), G = (0, M.useMemo)(() => Be(Ue, A, "active"), [Ue, A]), We = B(Ue).filter((e) => U(A, e.status) !== "active").length, K = (0, M.useMemo)(() => [...new Set(B(G).map((e) => e.anchor ?? ""))].filter(Boolean), [G]), Ge = Ee?.expanded ? Ee.anchor : null, Ke = Ge && !K.includes(Ge) ? Ge : null, qe = (0, M.useMemo)(() => Ke ? [...K, Ke] : K, [K, Ke]), Je = (0, M.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of B(G)) {
			if (!t.anchor || !$(t.anchor)) continue;
			let n = mn(t.status), r = e.get(t.anchor);
			(!r || fn[n] > fn[r]) && e.set(t.anchor, n);
		}
		return e;
	}, [G]), { boxes: q, height: Ye, width: Xe } = un({
		origin: Ie,
		markdown: V,
		anchors: qe,
		mode: o,
		visible: n === "preview"
	}), Ze = (0, M.useRef)(null), [Qe, $e] = (0, M.useState)(0), [et, tt] = (0, M.useState)([]);
	(0, M.useLayoutEffect)(() => {
		if (n !== "source" || !Ze.current) return;
		let e = V.current, t = () => tt(qe.flatMap((t) => t.startsWith("s:") ? [{
			anchor: t,
			spans: zn(j, Pn(t))
		}] : e && t.startsWith("q:") ? [{
			anchor: t,
			spans: $t(e, me, It(t))
		}] : []).filter((e) => e.spans.length > 0));
		t();
		let r = new MutationObserver(t);
		e && r.observe(e, {
			childList: !0,
			subtree: !0,
			characterData: !0
		});
		let i = Ze.current, a = new ResizeObserver(() => $e(i.offsetWidth));
		return a.observe(i), () => {
			r.disconnect(), a.disconnect();
		};
	}, [
		n,
		qe,
		j,
		me,
		k
	]);
	let nt = (0, M.useRef)(null), rt = (0, M.useRef)(null), J = F ?? Ce;
	(0, M.useEffect)(() => () => clearTimeout(nt.current ?? void 0), []);
	let [Y, it] = (0, M.useState)(null);
	(0, M.useLayoutEffect)(() => {
		let e = Ze.current;
		if (n !== "source" || !e || !J || !or(J)) return it(null);
		let t = e.getBoundingClientRect(), r = [...e.querySelectorAll("[data-source-comment]")].filter((e) => Tn(e).includes(J)).flatMap((e) => [...e.getClientRects()]).map((e) => ({
			top: e.top - t.top,
			left: e.left - t.left,
			width: e.width,
			height: e.height
		}));
		it(r.length > 0 ? {
			anchor: J,
			boxes: r
		} : null);
	}, [
		n,
		J,
		et,
		Qe
	]);
	let at = R ? q.get(R)?.[0] : void 0;
	(0, M.useEffect)(() => {
		let e = (H ? je : Me).current;
		if (n !== "preview" || !at || !e || !Ie.current) return;
		let t = Ie.current.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop + at.top;
		t >= e.scrollTop && t + at.height <= e.scrollTop + e.clientHeight || (e.scrollTop = t - mr), z(null);
	}, [
		n,
		at,
		H
	]), (0, M.useEffect)(() => {
		let e = je.current;
		if (n !== "source" || !R || !e) return;
		let t = $(R) ? R.slice(ir.length) : null, r = t ? e.querySelector(`[data-source-line="${t}"]`) : null;
		if (r) {
			let t = r.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
			t >= e.scrollTop && t + r.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = t - mr);
		}
		z(null);
	}, [n, R]);
	let X = p === "margin" && !H, ot = () => {
		X && P === "line" && Ne.current && Me.current && (Ne.current.scrollTop = Me.current.scrollTop);
	};
	(0, M.useEffect)(ot, [
		F,
		P,
		X
	]), (0, M.useEffect)(() => {
		let e = Ne.current;
		if (!X || P !== "stack" || !F || !e) return;
		let t = [...e.querySelectorAll("[data-card-anchor]")].find((e) => e.dataset.cardAnchor === F);
		if (!t) return;
		let n = t.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
		n >= e.scrollTop && n + t.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = n - mr);
	}, [
		F,
		P,
		X
	]);
	function st(e) {
		I(e), z(e), $(e) && y?.(Number(e.slice(ir.length)));
	}
	function ct(e) {
		let t = Q(e.line);
		r(T ? "preview" : "source"), I(t), z(t), T || Ae(e.line), y?.(e.line);
	}
	(0, M.useEffect)(() => {
		if (v === void 0) return;
		let e = Q(v);
		I(e), z(e), Ae(v);
	}, [v]);
	let lt = [...q.keys()], ut = sn({
		cards: K.map((e) => ({
			anchor: e,
			desired: q.get(e)?.[0]?.top ?? Ye,
			order: q.has(e) ? lt.indexOf(e) : lt.length,
			height: Le[e] ?? Zn
		})),
		focused: F
	}), dt = Math.max(Ye, ...K.map((e) => (ut.get(e) ?? 0) + (Le[e] ?? Zn))), ft = [...K].sort((e, t) => (ut.get(e) ?? 0) - (ut.get(t) ?? 0));
	function pt(e, t) {
		let n = e.currentTarget.getBoundingClientRect(), r = e.clientX - n.left, i = e.clientY - n.top, a = [...q].filter(([e, n]) => {
			if (e === Ke) return !1;
			if ($(e)) {
				let e = n[n.length - 1];
				return r <= er && i >= n[0].top && i <= e.top + e.height || t === "click" && n.some((e) => i >= e.top && i <= e.top + e.height && r >= e.left && r <= e.left + e.width);
			}
			return n.some((e) => {
				if (i < e.top || i > e.top + e.height) return !1;
				let n = o === "text";
				return t === "click" || n ? r >= e.left - (n ? 0 : 12) && r <= e.left + e.width : r >= e.left - er && r <= e.left;
			});
		}).map(([e]) => e);
		return a.find((e) => e.startsWith("q:")) ?? a[0] ?? null;
	}
	let Z = (e) => e.target.closest("[data-inline-thread], [data-comment-draft], [data-source-thread]") !== null;
	function mt(e, t) {
		if (!T) {
			let e = Dn(t), n = e && Rn(j, e.start, e.end);
			return n ? {
				anchor: Nn(n),
				label: Bn(n)
			} : null;
		}
		if (!V.current) return null;
		if (e === "preview") {
			let e = qt(V.current, t);
			return e && sr(e);
		}
		let n = On(t), r = n && Qt(V.current, me, n);
		return r && sr(r);
	}
	let ht = (e) => (e === "preview" ? Ie : Ze).current;
	function _t(e) {
		!window.getSelection()?.isCollapsed || Z(e) || e.target.closest("button, a") || I(pt(e, "click"));
	}
	function vt(e) {
		e !== rt.current && (rt.current = e, clearTimeout(nt.current ?? void 0), nt.current = setTimeout(() => Te(e), e ? Qn : $n));
	}
	function yt(e) {
		clearTimeout(nt.current ?? void 0), rt.current = e;
	}
	function bt() {
		clearTimeout(nt.current ?? void 0), rt.current = null, I(null), Te(null);
	}
	function xt(e) {
		if (Z(e)) return;
		let t = e.buttons !== 0 || !window.getSelection()?.isCollapsed, n = e.target.closest("[data-source-comment]");
		vt(t || !n ? null : Tn(n)[0]);
	}
	function Ct(e) {
		Z(e) || e.target.closest("[data-source-comment], button, a") || F && or(F) && window.getSelection()?.isCollapsed && bt();
	}
	function wt(e) {
		p !== "inline" || Z(e) || vt(e.buttons !== 0 || !window.getSelection()?.isCollapsed ? null : pt(e, "hover"));
	}
	function Et(e, t) {
		if (!de || Z(e) || e.target.closest("[data-comment-author]")) return;
		let n = window.getSelection(), r = n && !n.isCollapsed && n.rangeCount > 0 ? n.getRangeAt(0) : null, i = r ? mt(t, r) : null;
		if (!r || !i) {
			L((e) => e?.expanded ? e : null);
			return;
		}
		let a = ht(t).getBoundingClientRect(), o = r.getBoundingClientRect();
		L({
			...i,
			position: {
				top: o.top - a.top,
				bottom: o.bottom - a.top,
				left: o.left - a.left + o.width / 2
			},
			expanded: !1,
			surface: t
		});
	}
	function Ot(e, t) {
		if (!de || Z(e) || e.target.closest("[data-comment-author], button, a")) return;
		let n = window.getSelection(), r = n && !n.isCollapsed && n.rangeCount > 0 ? n.getRangeAt(0) : null, i = r && [...r.getClientRects()].some((t) => e.clientX >= t.left && e.clientX <= t.right && e.clientY >= t.top && e.clientY <= t.bottom), a = i ? r : dr(e.clientX, e.clientY), o = a ? mt(t, a) : null;
		if (!a || !o) return;
		e.preventDefault(), i || (n?.removeAllRanges(), n?.addRange(a));
		let s = ht(t).getBoundingClientRect(), c = a.getBoundingClientRect();
		L({
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
	function kt() {
		L((e) => e && {
			...e,
			expanded: !0
		}), window.getSelection()?.removeAllRanges();
	}
	let At = (0, M.useCallback)(() => L(null), []);
	function Mt(e) {
		let t = e.expanded && e.surface === "preview" ? q.get(e.anchor) : void 0, n = t?.[0], r = t?.at(-1);
		return !n || !r ? e.position : {
			top: n.top,
			bottom: r.top + r.height,
			left: n.left + n.width / 2
		};
	}
	function Nt(e, t) {
		let n = se.onCreate;
		if (!n) return null;
		let { anchor: r, label: i } = e;
		return /* @__PURE__ */ (0, N.jsx)(Yn, {
			anchor: r,
			label: i,
			position: Mt(e),
			containerWidth: t,
			expanded: e.expanded,
			config: A,
			onExpand: kt,
			onClose: At,
			onCreate: async (e) => {
				await n(e), L(null), I(r);
			}
		}, r);
	}
	return /* @__PURE__ */ (0, N.jsxs)("div", {
		ref: Fe,
		"data-file-review": e,
		className: s("flex min-h-0 flex-col gap-2", S ?? "h-[600px]"),
		children: [
			/* @__PURE__ */ (0, N.jsxs)("div", {
				className: "flex shrink-0 flex-wrap items-center gap-2 text-xs text-muted-foreground",
				children: [
					x,
					(a || C) && /* @__PURE__ */ (0, N.jsxs)("span", { children: [
						B(G).length,
						" open · ",
						We,
						" resolved (hidden)"
					] }),
					a && p === "inline" && /* @__PURE__ */ (0, N.jsx)("span", { children: "· hover a highlight to view or reply" }),
					de && /* @__PURE__ */ (0, N.jsx)("span", { children: "· select or right-click text to comment" }),
					d && /* @__PURE__ */ (0, N.jsx)("div", {
						role: "radiogroup",
						"aria-label": "Authors",
						className: "ml-auto inline-flex rounded-md border border-border p-0.5",
						children: [
							"all",
							"user",
							"agent"
						].map((e) => /* @__PURE__ */ (0, N.jsx)("button", {
							type: "button",
							role: "radio",
							"aria-checked": De === e,
							onClick: () => Oe(e),
							className: s("rounded px-2 py-0.5 font-medium", De === e ? "bg-primary text-primary-foreground" : "hover:text-foreground"),
							children: e === "all" ? "All" : e === "user" ? "People" : "Tools"
						}, e))
					})
				]
			}),
			/* @__PURE__ */ (0, N.jsxs)("div", {
				className: "flex shrink-0 items-center border-b border-border",
				children: [
					/* @__PURE__ */ (0, N.jsx)(ve, {
						className: "min-w-0 flex-1 border-b-0",
						value: n,
						onChange: (e) => {
							r(e), L(null);
						},
						tabs: [
							...T ? [{
								id: "preview",
								label: "Preview",
								icon: fe
							}] : [],
							{
								id: "source",
								label: "Source",
								icon: le
							},
							...C ? [{
								id: "problems",
								label: "Problems",
								icon: ue,
								count: w.length
							}] : []
						]
					}),
					C && w.length > 0 && n === "source" && !k && /* @__PURE__ */ (0, N.jsxs)(c, {
						size: "sm",
						variant: D ? "secondary" : "ghost",
						"aria-pressed": D,
						title: D ? "Show only the gutter marks; pick one to read its findings" : "Show each line's findings underneath it",
						onClick: () => re((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, N.jsx)(_, { icon: ue }), "Findings"]
					}),
					g && n === "source" && /* @__PURE__ */ (0, N.jsxs)(c, {
						size: "sm",
						variant: E ? "secondary" : "ghost",
						"aria-pressed": E,
						title: E ? "Hide who last changed each line" : "Show who last changed each line",
						onClick: () => ee((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, N.jsx)(_, { icon: we }), "Blame"]
					}),
					a && m && T && /* @__PURE__ */ (0, N.jsx)(ae, {
						align: "right",
						icon: ne,
						hideChevron: !0,
						variant: "ghost",
						size: "sm",
						title: "View options",
						menuLabel: "View options",
						className: "shrink-0",
						items: [...fr.map((e) => ({
							label: e.label,
							group: "Comments",
							icon: O,
							iconClassName: e.id === p ? void 0 : "invisible",
							onSelect: () => m(e.id)
						})), ...p === "margin" ? pr.map((e) => ({
							label: e.label,
							group: "Sidebar cards",
							icon: O,
							iconClassName: e.id === P ? void 0 : "invisible",
							onSelect: () => Se(e.id)
						})) : []]
					})
				]
			}),
			/* @__PURE__ */ (0, N.jsxs)("div", {
				ref: je,
				className: s("relative min-h-0 flex-1 rounded-md border border-border bg-background", k && n === "source" ? "overflow-hidden" : "overflow-y-auto"),
				children: [
					k && n === "source" && /* @__PURE__ */ (0, N.jsx)(Dt, {
						path: e,
						source: t,
						findings: w,
						line: ke,
						onLineChange: Ae,
						onChange: h,
						inlineChips: !0
					}),
					!k && n === "source" && /* @__PURE__ */ (0, N.jsxs)("div", {
						ref: Ze,
						className: "relative",
						onMouseUp: (e) => e.button === 0 && Et(e, "source"),
						onContextMenu: (e) => Ot(e, "source"),
						onClick: Ct,
						onMouseMove: xt,
						onMouseLeave: () => vt(null),
						children: [
							/* @__PURE__ */ (0, N.jsx)(An, {
								source: j,
								tokens: xe,
								blame: he,
								findings: w,
								showFindings: D,
								marks: et,
								comments: G,
								selected: J ?? Ge,
								onSelect: (e) => {
									I(Q(e.line)), y?.(e.line);
								},
								renderFindingActions: b,
								onFocus: I
							}),
							Ee?.surface === "source" && Nt(Ee, Qe),
							Y && /* @__PURE__ */ (0, N.jsx)(Un, {
								seamless: !0,
								anchor: Y.anchor,
								label: lr(Y.anchor, !0),
								boxes: Y.boxes,
								containerWidth: Qe,
								thread: Pe(G, Y.anchor),
								config: A,
								callbacks: se,
								pinned: F === Y.anchor,
								onPin: () => I(Y.anchor),
								onClose: bt,
								onPointerEnter: () => yt(Y.anchor),
								onPointerLeave: () => vt(null)
							}, Y.anchor)
						]
					}),
					n === "problems" && /* @__PURE__ */ (0, N.jsx)(bn, {
						source: j,
						findings: w,
						onOpen: ct,
						openLabel: T ? "Open in preview" : "Open in source"
					}),
					T && /* @__PURE__ */ (0, N.jsxs)("div", {
						className: s("flex", H ? "flex-col gap-4 p-4" : "h-full", n !== "preview" && "hidden"),
						children: [/* @__PURE__ */ (0, N.jsx)("div", {
							ref: Me,
							onScroll: ot,
							className: s("min-w-0 flex-1", !H && "overflow-y-auto p-4"),
							children: /* @__PURE__ */ (0, N.jsxs)("div", {
								ref: Ie,
								className: "relative",
								onMouseUp: (e) => e.button === 0 && Et(e, "preview"),
								onContextMenu: (e) => Ot(e, "preview"),
								onClick: _t,
								onMouseMove: wt,
								onMouseLeave: () => p === "inline" && vt(null),
								children: [
									[...q].flatMap(([e, t]) => {
										let n = e === F || e === Ce || e === Ge;
										if ($(e)) {
											let r = Je.get(e);
											if (!r) throw Error(`lint anchor ${e} has no open finding`);
											let i = pn[r], a = t[0];
											return [/* @__PURE__ */ (0, N.jsx)("span", {
												"data-lint-rail": r,
												"aria-hidden": !0,
												className: s("pointer-events-none absolute z-20 flex size-4 items-center justify-center rounded-sm", i.tone, n && i.tint),
												style: {
													top: a.top + a.height / 2 - 8,
													left: 0
												},
												children: /* @__PURE__ */ (0, N.jsx)(_, {
													icon: i.icon,
													className: "text-[13px]"
												})
											}, `${e}:rail`), ...t.map((t, a) => /* @__PURE__ */ (0, N.jsx)("div", {
												"data-anchor-box": e,
												"data-lint-squiggle": r,
												"aria-hidden": !0,
												className: s("pointer-events-none absolute z-0 transition-colors", n && i.tint),
												style: {
													top: t.top,
													left: t.left,
													width: t.width,
													height: t.height + 3
												},
												children: /* @__PURE__ */ (0, N.jsx)("span", {
													className: "absolute inset-x-0 bottom-0 h-[3px]",
													style: {
														backgroundImage: hn(i.stroke),
														backgroundRepeat: "repeat-x",
														backgroundSize: "6px 3px"
													}
												})
											}, `${e}:${a}`))];
										}
										return t.map((t, r) => /* @__PURE__ */ (0, N.jsx)("div", {
											"data-anchor-box": e,
											"aria-hidden": !0,
											className: s("pointer-events-none absolute z-0 transition-colors", cr({
												mode: o,
												active: n
											})),
											style: o === "text" ? {
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
									/* @__PURE__ */ (0, N.jsx)("div", {
										ref: V,
										"data-annotated-markdown": "",
										className: "relative z-10 pl-4",
										children: /* @__PURE__ */ (0, N.jsx)(pe, {
											sourceBlocks: !0,
											text: j
										})
									}),
									a && p === "inline" && /* @__PURE__ */ (0, N.jsx)(Wn, {
										markdown: V,
										comments: G,
										focused: F,
										onFocus: I
									}),
									Ee?.surface === "preview" && Nt(Ee, Xe),
									p === "inline" && J && /* @__PURE__ */ (0, N.jsx)(Un, {
										anchor: J,
										label: lr(J, !0),
										boxes: q.get(J) ?? [],
										containerWidth: Xe,
										thread: Pe(G, J),
										config: A,
										callbacks: $(J) ? rr : se,
										renderRootDetails: He,
										pinned: F === J,
										onPin: () => I(J),
										onClose: bt,
										onPointerEnter: () => yt(J),
										onPointerLeave: () => vt(null)
									}, J)
								]
							})
						}), p === "margin" && /* @__PURE__ */ (0, N.jsx)("div", {
							ref: Ne,
							"data-comment-sidebar": H ? "below" : P,
							className: s(H ? "w-full" : "w-[316px] shrink-0 overflow-y-auto overscroll-contain border-l border-border py-4 pl-2 pr-4"),
							children: /* @__PURE__ */ (0, N.jsx)("div", {
								className: s("relative", (H || P === "stack") && "space-y-2"),
								style: X && P === "line" ? { height: dt } : void 0,
								children: ft.map((e) => {
									let t = Pe(G, e), n = e === F;
									return /* @__PURE__ */ (0, N.jsxs)("div", {
										ref: Re(e),
										"data-card-anchor": e,
										onMouseEnter: () => Te(e),
										onMouseLeave: () => Te(null),
										onClick: () => st(e),
										className: s("rounded-lg border bg-card transition-[top,transform] duration-150", X && P === "line" && "absolute inset-x-0", n ? "z-10 -translate-x-2 border-primary/50 shadow-lg" : "border-border shadow-sm"),
										style: X && P === "line" ? { top: ut.get(e) ?? 0 } : void 0,
										children: [/* @__PURE__ */ (0, N.jsx)("div", {
											className: "truncate px-2 pt-2 text-[11px] text-muted-foreground",
											children: lr(e, q.has(e))
										}), /* @__PURE__ */ (0, N.jsx)(gt, {
											comments: t,
											config: A,
											anchor: e,
											compact: !0,
											defaultExpanded: n || $(e),
											hideComposer: t.length > 0,
											autoFocusComposer: t.length === 0,
											composerPlaceholder: "Add a comment…",
											...$(e) ? rr : se,
											...He ? { renderRootDetails: He } : {}
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
// ../../../clicky-ui/packages/ui/src/file-findings/severity-counts.ts
var gr = [
	"error",
	"warning",
	"info"
];
function _r(e) {
	let t = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let n of e) t[n.severity] += 1;
	return t;
}
// ../../../clicky-ui/packages/ui/src/file-findings/FileFindingsView.tsx
function vr(e) {
	return /\.mdx?$/i.test(e);
}
function yr({ findings: e }) {
	let t = _r(e);
	return /* @__PURE__ */ (0, N.jsx)("span", {
		className: "flex shrink-0 items-center gap-2 text-xs text-muted-foreground",
		children: gr.filter((e) => t[e] > 0).map((e) => /* @__PURE__ */ (0, N.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			title: `${t[e]} ${e}`,
			children: [/* @__PURE__ */ (0, N.jsx)(de, { severity: e }), /* @__PURE__ */ (0, N.jsx)("span", {
				className: "tabular-nums",
				children: t[e]
			})]
		}, e))
	});
}
function br({ node: e, selected: t }) {
	return /* @__PURE__ */ (0, N.jsxs)(N.Fragment, { children: [
		/* @__PURE__ */ (0, N.jsx)(_, {
			icon: e.kind === "folder" ? A : le,
			className: "shrink-0 text-muted-foreground"
		}),
		/* @__PURE__ */ (0, N.jsx)("span", {
			className: s("min-w-0 truncate text-foreground", t ? "font-semibold" : "font-medium"),
			title: e.name,
			children: e.name
		}),
		/* @__PURE__ */ (0, N.jsx)("span", { className: "flex-1" }),
		/* @__PURE__ */ (0, N.jsx)(yr, { findings: e.findings })
	] });
}
function xr(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.kind === "file" && me(n.ref ?? null, t)) return n;
		let e = xr(n.children ?? [], t);
		if (e) return e;
	}
	return null;
}
function Sr(e, t) {
	let [n, r] = (0, M.useState)(null), i = e ? he(e) : null, a = (0, M.useRef)({
		ref: e,
		loadFile: t
	});
	return a.current = {
		ref: e,
		loadFile: t
	}, (0, M.useEffect)(() => {
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
	}, [i]), n?.key === i ? n : null;
}
function Cr({ children: e }) {
	return /* @__PURE__ */ (0, N.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
		children: e
	});
}
function wr({ file: e, findings: t, state: n, line: r, onLineChange: i, view: a, onViewChange: o, renderFindingActions: s }) {
	let c = !vr(e.file) && a === "preview" ? "source" : a, l;
	return l = !n || n.status === "loading" ? /* @__PURE__ */ (0, N.jsxs)(Cr, { children: [
		/* @__PURE__ */ (0, N.jsx)(_, {
			icon: ce,
			className: "mr-2 animate-spin"
		}),
		"Loading ",
		e.file
	] }) : n.status === "error" ? /* @__PURE__ */ (0, N.jsxs)("div", {
		role: "alert",
		className: "m-density-4 rounded-md border border-destructive/40 p-3 text-sm text-destructive",
		children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: "font-semibold",
			children: ["Could not load ", e.file]
		}), /* @__PURE__ */ (0, N.jsx)("pre", {
			className: "mt-1 whitespace-pre-wrap break-words font-mono text-xs",
			children: n.message
		})]
	}) : /* @__PURE__ */ (0, N.jsx)(hr, {
		path: e.file,
		source: n.text,
		view: c,
		onViewChange: o,
		findings: t,
		line: r,
		onLineChange: i,
		renderFindingActions: s,
		className: "h-full p-density-2"
	}, he(e)), /* @__PURE__ */ (0, N.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, N.jsxs)("div", {
			className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
			"data-file-header": "",
			children: [/* @__PURE__ */ (0, N.jsx)("span", {
				className: "min-w-0 truncate font-mono text-sm",
				title: e.workDir ? `${e.workDir}/${e.file}` : e.file,
				children: e.file
			}), /* @__PURE__ */ (0, N.jsx)(yr, { findings: t })]
		}), /* @__PURE__ */ (0, N.jsx)("div", {
			className: "min-h-0 flex-1",
			children: l
		})]
	});
}
function Tr({ findings: e, selectedFile: t, onSelectFile: n, line: r, onLineChange: i, view: a, onViewChange: o, loadFile: c, renderFindingActions: l, defaultSplit: u = 28, className: d }) {
	let f = (0, M.useMemo)(() => ge(e), [e]), p = (0, M.useMemo)(() => xr(f, t), [f, t]), m = Sr(t, c), h = /* @__PURE__ */ (0, N.jsx)(oe, {
		className: "min-h-0",
		roots: f,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		selected: p,
		onSelect: (e) => {
			e.kind === "file" && e.ref && n(e.ref);
		},
		defaultOpen: () => !0,
		showSearch: !0,
		searchLabel: "Filter files",
		getSearchText: (e) => e.name,
		empty: /* @__PURE__ */ (0, N.jsx)("div", {
			className: "p-8 text-center text-sm text-muted-foreground",
			children: "No lint violations or failed tests with a file location"
		}),
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, N.jsx)(br, {
			node: e,
			selected: t
		})
	}), g = t ? p ? /* @__PURE__ */ (0, N.jsx)(wr, {
		file: t,
		findings: p.findings,
		state: m,
		line: r,
		onLineChange: i,
		view: a,
		onViewChange: o,
		renderFindingActions: l
	}) : /* @__PURE__ */ (0, N.jsxs)(Cr, { children: [t.file, " has no findings in this run."] }) : /* @__PURE__ */ (0, N.jsx)(Cr, { children: "Select a file to see its findings in context." });
	return /* @__PURE__ */ (0, N.jsx)("div", {
		className: s("flex h-full min-h-0 flex-col bg-background", d),
		"data-file-findings": "",
		children: /* @__PURE__ */ (0, N.jsx)(se, {
			className: "min-h-0 flex-1",
			defaultSplit: u,
			minLeft: 15,
			minRight: 40,
			leftClass: "border-r-2 border-border bg-muted/30",
			left: h,
			right: g
		})
	});
}
// src/FilesTab.tsx
var Er = /* @__PURE__ */ e({ default: () => Dr });
function Dr(e) {
	return /* @__PURE__ */ (0, N.jsx)(Tr, { ...e });
}
export { Fe as _, bt as a, ht as c, Ne as d, Dr as default, ze as f, Pe as g, B as h, Q as i, U as l, V as m, St as n, xt as o, Ie as p, Tt as r, gt as s, Er as t, Be as u, Oe as v };
