import { n as e, r as t } from "./rolldown-runtime-QGRDbDUS.js";
import { n, t as r } from "./jsx-runtime-f3E3oPuc.js";
import { t as i } from "./react-dom-D8HDTZ2P.js";
import { C as a, S as o, T as s, _ as c } from "./Callout-DALwzckz.js";
import { At as l, Et as u, F as d, J as f, L as p, St as m, Tt as h, X as g, Y as _, Z as v, at as y, ct as b, dt as x, et as S, ft as C, g as w, ht as T, kt as E, l as ee, lt as te, o as D, p as O, rt as k, t as ne, tt as re, vt as ie, y as ae } from "./clipboard-CMVj3cUp.js";
import { E as oe, Q as se, T as ce, U as le, Z as ue, nt as de, o as fe, st as pe } from "./paths-D_tC_cYe.js";
import { n as A, t as me } from "./Markdown-4Pj2JdiG.js";
import { a as he, c as j, f as ge, g as _e, i as ve, o as ye, p as be, s as xe, t as Se } from "./adapters-B0rO-m88.js";
import { m as Ce, t as we } from "./Tabs-Dd-b3nbz.js";
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/lib/palette.ts
function M(e) {
	let t = 2166136261;
	for (let n = 0; n < e.length; n++) t ^= e.charCodeAt(n), t = Math.imul(t, 16777619);
	return t >>> 0;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/data/Avatar.tsx
var N = /* @__PURE__ */ t(n(), 1), P = r();
function Te(e) {
	return M(e) % 360;
}
function F(e, t, n) {
	if (t?.trim()) return t.trim().slice(0, n).toUpperCase();
	let r = e.replace(/^@/, "").trim(), i = (r.split(/[\\/]/).filter(Boolean).at(-1) ?? r).replace(/[^a-zA-Z0-9]+/g, " ").trim().split(/\s+/).filter(Boolean).map((e) => e.match(/[a-zA-Z0-9]/)?.[0] ?? "").filter(Boolean).slice(0, n).join("");
	return i ? i.toUpperCase() : (r.replace(/[^a-zA-Z0-9]+/g, "").slice(0, n) || "?").toUpperCase();
}
function I(e, t, n) {
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
function Ee(e, t) {
	let n = t <= 16 ? 7 : t <= 20 ? 8 : 9;
	switch (e) {
		case "stamp": return Math.max(n, Math.round(t * .34));
		case "mono": return Math.max(n, Math.round(t * .36));
		default: return Math.max(n, Math.round(t * .38));
	}
}
function De({ src: e, alt: t, initials: n, size: r = "sm", rounded: i, kind: a = "user", variant: o = "duotone", title: c, href: u, colorKey: d, onError: f, className: p }) {
	let [m, h] = (0, N.useState)(), g = E(), _ = l(r, g), v = (i ?? (a === "group" ? "md" : "full")) === "full" ? "rounded-full" : "rounded-md", y = I(o, a, Te(d ?? t)), b = F(t, n, r === "xs" || r === "sm" ? 1 : 2), x = s("inline-flex shrink-0 select-none items-center justify-center overflow-hidden border", v), S = {
		width: _,
		height: _,
		borderColor: y.borderColor,
		borderStyle: y.borderStyle ?? "solid"
	};
	o === "stamp" && (S.transform = "rotate(-4deg)", S.boxShadow = y.shadow);
	let C = (e) => e === "stamp" ? 700 : r === "sm" ? 500 : 600, w = e && e !== m ? /* @__PURE__ */ (0, P.jsx)("img", {
		src: e,
		alt: t,
		title: c ?? t,
		width: _,
		height: _,
		"data-avatar-kind": a,
		"data-avatar-variant": o,
		"data-avatar-size": r,
		className: s(x, "bg-muted object-cover", p),
		style: S,
		loading: "lazy",
		onError: (t) => {
			h(e), f?.(t);
		}
	}) : /* @__PURE__ */ (0, P.jsx)("span", {
		role: "img",
		"aria-label": t,
		"data-avatar-kind": a,
		"data-avatar-variant": o,
		"data-avatar-size": r,
		className: s(x, o === "mono" || o === "stamp" ? "font-mono" : void 0, o === "mono" ? "tracking-[-0.05em]" : "tracking-[-0.02em]", p),
		style: {
			...S,
			background: y.background,
			color: y.color,
			fontSize: Ee(o, _),
			fontWeight: C(o)
		},
		title: c ?? t,
		children: b
	});
	return u ? /* @__PURE__ */ (0, P.jsx)("a", {
		href: u,
		target: "_blank",
		rel: "noopener noreferrer",
		className: "inline-flex shrink-0",
		onClick: (e) => e.stopPropagation(),
		children: w
	}) : w;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiThumbsUp.tsx
var L = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiThumbsDown.tsx
var R = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiGitCommit.tsx
var Oe = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/icons/components/UiUser.tsx
var ke = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsx)("svg", {
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
var Ae = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
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
var je = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
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
var Me = Object.assign(({ size: e = "1em", className: t, title: n, ...r }) => /* @__PURE__ */ (0, P.jsxs)("svg", {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/comment-types.ts
var z = "__document__", Ne = [
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
], Pe = [
	"open",
	"in_progress",
	"done"
];
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/comment-utils.ts
function B(e) {
	return e?.name?.trim() || "Unknown";
}
function Fe(e) {
	switch (e) {
		case "success": return "success";
		case "danger": return "danger";
		case "warning": return "warning";
		case "info": return "info";
		default: return "neutral";
	}
}
function Ie(e, t) {
	if (t != null) return e.statuses.find((e) => e.value === t);
}
function Le(e, t) {
	if (t != null) return e.options.find((e) => e.value === t);
}
function V(e) {
	return e.filter((e) => !e.parentId);
}
function H(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) {
		if (!n.parentId) continue;
		let e = t.get(n.parentId) ?? [];
		e.push(n), t.set(n.parentId, e);
	}
	return t;
}
function Re(e, t) {
	let n = e.filter((e) => !e.parentId && (e.anchor ?? "__document__") === t), r = new Set(n.map((e) => String(e.id))), i = [];
	for (let t = !0; t;) {
		t = !1;
		for (let n of e) n.parentId && !r.has(String(n.id)) && r.has(String(n.parentId)) && (r.add(String(n.id)), i.push(n), t = !0);
	}
	return [...n, ...i];
}
function U(e) {
	return [...e].sort((e, t) => e.createdAt === t.createdAt ? String(e.id).localeCompare(String(t.id)) : e.createdAt.localeCompare(t.createdAt));
}
function ze(e) {
	let t = {};
	for (let n of e) {
		let e = n.anchor ?? "__document__";
		t[e] = (t[e] ?? 0) + 1;
	}
	return t;
}
function Be(e) {
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
function Ve(e, t = Pe) {
	let n = t.indexOf(e);
	return n === -1 ? t[0] ?? e : t[(n + 1) % t.length] ?? e;
}
function He(e, t = 80) {
	let n = e.replace(/[#*_`~>[\]()!]/g, "").trim();
	return n.length > t ? `${n.slice(0, t)}…` : n;
}
function Ue(e, t) {
	let n = e.toLowerCase();
	return t.filter((e) => n.includes(`@${e.name.toLowerCase()}`)).map((e) => ({
		id: e.id,
		name: e.name,
		kind: e.kind
	}));
}
function We(e, t, n) {
	let r = {};
	if (n.onUpdateStatus && (r.onUpdateStatus = n.onUpdateStatus), n.onClose && (r.onClose = n.onClose), n.onAcknowledge && (r.onAcknowledge = n.onAcknowledge), n.onAssign && (r.onAssign = n.onAssign), n.onDelete && (r.onDelete = n.onDelete), n.onChecklistToggle) {
		let i = n.onChecklistToggle;
		r.onChecklistToggle = (n, r) => {
			let a = e.find((e) => e.id === n)?.checklist?.[r];
			a && i(n, r, Ve(a.status, t.checklistStatusCycle));
		};
	}
	if (n.onReply) {
		let e = n.onReply, i = n.onMention;
		r.onReply = async (n, r) => {
			let a = Ue(r, t.mentionables ?? []);
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/lib/comment-stage.ts
function W(e, t) {
	if (t == null) return "active";
	let n = Ie(e, t);
	if (n) return n.stage ? n.stage : n.unresolved ? "active" : n.value.toLowerCase() === "closed" || n.label.toLowerCase() === "closed" ? "closed" : "resolved";
}
function G(e, t) {
	return e.statuses.find((n) => W(e, n.value) === t);
}
function Ge(e, t, n) {
	let r = new Set(V(e).filter((e) => W(t, e.status) === n).map((e) => String(e.id)));
	return e.filter((e) => r.has(String(e.parentId ?? e.id)));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentAuthor.tsx
function K({ author: e, size: t = "sm", bare: n = !1 }) {
	let r = E(), i = B(e), a = e?.kind === "agent", o = a ? /* @__PURE__ */ (0, P.jsx)(Ke, {
		author: e,
		size: l(t, r)
	}) : /* @__PURE__ */ (0, P.jsx)(De, {
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
function Ke({ author: e, size: t }) {
	let n = e.icon;
	return /* @__PURE__ */ (0, P.jsx)("span", {
		role: "img",
		"aria-label": e.name,
		className: "inline-flex shrink-0 items-center justify-center rounded-full border border-border bg-muted text-foreground",
		style: {
			width: t,
			height: t
		},
		children: /* @__PURE__ */ (0, P.jsx)(u, {
			...typeof n == "string" ? { name: n } : { icon: n ?? re },
			width: Math.max(10, Math.round(t * .6)),
			height: Math.max(10, Math.round(t * .6))
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/MentionTextarea.tsx
var q = /* @__PURE__ */ t(i(), 1), qe = /(^|\s)@([\w.-]*)$/, Je = 8;
function Ye(e, t) {
	let n = e.slice(0, t).match(qe);
	if (!n) return null;
	let r = n[2] ?? "";
	return {
		atIndex: t - r.length - 1,
		query: r
	};
}
function Xe(e, t) {
	let n = t.toLowerCase();
	return (n ? e.filter((e) => e.name.toLowerCase().includes(n) || e.id.toLowerCase().includes(n)) : e).slice(0, Je);
}
var J = /* @__PURE__ */ "boxSizing.width.borderTopWidth.borderRightWidth.borderBottomWidth.borderLeftWidth.paddingTop.paddingRight.paddingBottom.paddingLeft.fontStyle.fontVariant.fontWeight.fontStretch.fontSize.lineHeight.fontFamily.textAlign.textTransform.textIndent.letterSpacing.wordSpacing.tabSize.whiteSpace.wordWrap.overflowWrap".split(".");
function Ze(e, t) {
	if (typeof document > "u" || typeof getComputedStyle != "function") {
		let t = e.getBoundingClientRect();
		return {
			top: t.top,
			left: t.left,
			height: 16
		};
	}
	let n = getComputedStyle(e), r = document.createElement("div");
	for (let e of J) r.style[e] = n[e];
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
var Qe = (0, N.forwardRef)(function({ value: e, onChange: t, mentionables: n, onMentionSelect: r, onSubmit: i, onCancel: a, placeholder: o, autoFocus: c, rows: l = 2, className: u, "data-testid": d }, f) {
	let m = (0, N.useRef)(null);
	(0, N.useImperativeHandle)(f, () => m.current);
	let [h, g] = (0, N.useState)(null), [_, v] = (0, N.useState)(null), [y, b] = (0, N.useState)(0), x = p(), S = h ? Xe(n, h.query) : [], C = h != null && S.length > 0;
	function w(e) {
		let t = Ye(e.value, e.selectionStart ?? e.value.length);
		g(t), b(0), v(t ? Ze(e, e.selectionStart ?? 0) : null);
	}
	function T(n) {
		let i = m.current;
		if (!i || !h) return;
		let a = i.selectionStart ?? e.length, o = `${e.slice(0, h.atIndex)}@${n.name} ${e.slice(a)}`, s = h.atIndex + n.name.length + 2;
		t(o), r?.({
			id: n.id,
			name: n.name,
			kind: n.kind
		}), g(null), v(null), requestAnimationFrame(() => {
			i.focus(), i.setSelectionRange(s, s);
		});
	}
	function E(e) {
		if (C) {
			if (e.key === "ArrowDown") {
				e.preventDefault(), b((e) => (e + 1) % S.length);
				return;
			}
			if (e.key === "ArrowUp") {
				e.preventDefault(), b((e) => (e - 1 + S.length) % S.length);
				return;
			}
			if (e.key === "Enter" || e.key === "Tab") {
				e.preventDefault();
				let t = S[y];
				t && T(t);
				return;
			}
			if (e.key === "Escape") {
				e.preventDefault(), g(null), v(null);
				return;
			}
		}
		e.key === "Enter" && !e.shiftKey ? (e.preventDefault(), i?.()) : e.key === "Escape" && (e.preventDefault(), a?.());
	}
	return (0, N.useEffect)(() => {
		if (!c) return;
		let e = m.current, t = requestAnimationFrame(() => {
			e?.focus();
			let t = e?.value.length ?? 0;
			e?.setSelectionRange(t, t);
		});
		return () => cancelAnimationFrame(t);
	}, [c]), /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)("textarea", {
		ref: m,
		rows: l,
		value: e,
		placeholder: o,
		"data-testid": d,
		className: s("w-full resize-none bg-transparent text-sm leading-snug outline-none placeholder:text-placeholder [@media(pointer:coarse)]:text-base", u),
		onChange: (e) => {
			t(e.target.value), w(e.target);
		},
		onKeyUp: (e) => w(e.currentTarget),
		onClick: (e) => w(e.currentTarget),
		onKeyDown: E,
		onBlur: () => {
			g(null), v(null);
		}
	}), C && typeof document < "u" && (0, q.createPortal)(/* @__PURE__ */ (0, P.jsx)("ul", {
		role: "listbox",
		"data-testid": "mention-popover",
		style: {
			position: "fixed",
			top: (_?.top ?? 0) + (_?.height ?? 16) + 2,
			left: _?.left ?? 0,
			zIndex: x
		},
		className: "max-h-56 min-w-[12rem] overflow-auto rounded-md border border-border bg-popover py-1 shadow-md",
		children: S.map((e, t) => /* @__PURE__ */ (0, P.jsx)("li", { children: /* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			role: "option",
			"aria-selected": t === y,
			onMouseDown: (t) => {
				t.preventDefault(), T(e);
			},
			onMouseEnter: () => b(t),
			className: s("flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs text-popover-foreground", t === y && "bg-accent text-accent-foreground"),
			children: [
				/* @__PURE__ */ (0, P.jsx)(K, {
					author: {
						name: e.name,
						kind: e.kind,
						...e.avatar ? { avatar: e.avatar } : {},
						...e.icon ? { icon: e.icon } : {}
					},
					size: "xs",
					bare: !0
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "truncate",
					children: e.name
				}),
				e.kind === "agent" && /* @__PURE__ */ (0, P.jsx)("span", {
					className: "ml-auto text-[10px] text-muted-foreground",
					children: "agent"
				})
			]
		}) }, e.id))
	}), document.body)] });
}), $e = {
	neutral: "bg-muted-foreground",
	success: "bg-green-500",
	danger: "bg-red-500",
	warning: "bg-yellow-500",
	info: "bg-blue-500"
};
function et({ facet: e, value: t, onChange: n }) {
	let r = e.options.find((e) => e.value === t);
	return /* @__PURE__ */ (0, P.jsx)(ee, {
		align: "left",
		menuClassName: "min-w-[150px]",
		trigger: /* @__PURE__ */ (0, P.jsxs)("span", {
			role: "button",
			tabIndex: 0,
			className: s("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium transition-colors", r ? "bg-muted text-foreground" : "text-muted-foreground hover:bg-muted/60"),
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
				className: s("flex w-full items-center gap-2 px-3 py-1.5 text-left text-[11px] text-popover-foreground hover:bg-accent hover:text-accent-foreground", e.value === t && "bg-muted font-semibold"),
				children: [/* @__PURE__ */ (0, P.jsx)("span", { className: s("inline-block size-2 rounded-full", $e[Fe(e.tone)]) }), e.label]
			}, e.value))
		})
	});
}
function tt({ config: e, anchor: t, collapsible: n = !0, autoFocus: r = !1, placeholder: i = "Add a comment…", onCreate: a, createActions: o, onMention: c }) {
	let [l, d] = (0, N.useState)(!n || r), [f, p] = (0, N.useState)(""), [m, h] = (0, N.useState)({}), [g, _] = (0, N.useState)(), [v, y] = (0, N.useState)(!1), b = (0, N.useRef)(!1), x = (0, N.useRef)(/* @__PURE__ */ new Map()), S = (0, N.useRef)(null);
	function C() {
		p(""), h({}), _(void 0), x.current.clear();
	}
	function w(e, t) {
		h((n) => {
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
		let r = f.trim();
		if (!r && !g || b.current) return;
		b.current = !0, y(!0);
		let i = [...x.current.values()].filter((e) => f.toLowerCase().includes(`@${e.name.toLowerCase()}`));
		try {
			let o = {
				body: r,
				anchor: t ?? null,
				...g ? { rating: g } : {},
				...Object.keys(m).length > 0 ? { facets: m } : {},
				...i.length > 0 ? { mentions: i } : {}
			};
			e === void 0 ? await a?.(o) : await a?.(o, e);
			for (let e of i) c?.(e, { body: r });
			C(), n && d(!1);
		} finally {
			b.current = !1, y(!1);
		}
	}
	return !l && f.trim().length === 0 ? /* @__PURE__ */ (0, P.jsx)("button", {
		type: "button",
		onClick: () => d(!0),
		className: "w-full rounded-full border border-border bg-muted/25 px-4 py-2 text-left text-sm text-muted-foreground/70 shadow-sm transition-colors hover:border-primary/30 hover:bg-muted/40",
		children: i
	}) : /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-testid": "comment-compose-box",
		"aria-busy": v,
		className: "relative rounded-2xl border border-border bg-muted/25 px-3 pt-3 pb-10 shadow-sm transition-[border-color,box-shadow] focus-within:border-primary/40 focus-within:shadow-md [@media(pointer:coarse)]:pb-2",
		children: [
			/* @__PURE__ */ (0, P.jsx)(Qe, {
				ref: S,
				value: f,
				onChange: p,
				mentionables: e.mentionables ?? [],
				onMentionSelect: (e) => x.current.set(e.id, e),
				onSubmit: () => void T(),
				onCancel: () => {
					C(), n && d(!1);
				},
				placeholder: i,
				autoFocus: r || n && l,
				"data-testid": "comment-compose-input",
				className: "min-h-[48px] pr-10 [@media(pointer:coarse)]:pr-12"
			}),
			n && /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: () => {
					C(), d(!1);
				},
				"aria-label": "Close composer",
				className: "absolute right-3 top-3 inline-flex items-center justify-center rounded-full p-1 text-muted-foreground hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:right-1 [@media(pointer:coarse)]:top-1 [@media(pointer:coarse)]:size-11",
				children: /* @__PURE__ */ (0, P.jsx)(u, {
					icon: ie,
					className: "text-xs"
				})
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "absolute inset-x-3 bottom-2.5 flex items-center gap-1.5 [@media(pointer:coarse)]:static [@media(pointer:coarse)]:mt-2 [@media(pointer:coarse)]:flex-wrap [@media(pointer:coarse)]:gap-2",
				children: [
					[[
						"positive",
						"Positive rating",
						L
					], [
						"negative",
						"Negative rating",
						R
					]].map(([e, t, n]) => /* @__PURE__ */ (0, P.jsx)("button", {
						type: "button",
						"aria-label": t,
						"aria-pressed": g === e,
						onClick: () => _((t) => t === e ? void 0 : e),
						className: s("inline-flex size-7 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11", g === e && (e === "positive" ? "bg-green-100 text-green-700 [[data-theme=dark]_&]:bg-green-500/20 [[data-theme=dark]_&]:text-green-300" : "bg-red-100 text-red-700 [[data-theme=dark]_&]:bg-red-500/20 [[data-theme=dark]_&]:text-red-300")),
						children: /* @__PURE__ */ (0, P.jsx)(u, {
							icon: n,
							className: "text-sm"
						})
					}, e)),
					(e.facets ?? []).map((e) => /* @__PURE__ */ (0, P.jsx)(et, {
						facet: e,
						value: m[e.key],
						onChange: (t) => w(e.key, t)
					}, e.key)),
					o?.length ? /* @__PURE__ */ (0, P.jsx)("div", {
						className: "ml-auto flex gap-1 [@media(pointer:coarse)]:flex-wrap [@media(pointer:coarse)]:justify-end [@media(pointer:coarse)]:gap-2",
						children: o.map((e) => /* @__PURE__ */ (0, P.jsx)("button", {
							type: "button",
							onClick: () => void T(e.id),
							disabled: v || !f.trim() && !g,
							className: "whitespace-nowrap rounded-md bg-primary px-1.5 py-1 text-[10px] font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
							children: e.label
						}, e.id))
					}) : /* @__PURE__ */ (0, P.jsx)("button", {
						type: "button",
						"data-testid": "comment-compose-send",
						"aria-label": "Post comment",
						onClick: () => void T(),
						disabled: v || !f.trim() && !g,
						className: "ml-auto inline-flex size-7 items-center justify-center rounded-full bg-primary text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground [@media(pointer:coarse)]:size-11",
						children: /* @__PURE__ */ (0, P.jsx)(u, {
							icon: A,
							className: "text-xs"
						})
					})
				]
			})
		]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentCardParts.tsx
function nt({ decision: e }) {
	return e ? /* @__PURE__ */ (0, P.jsx)(h, {
		variant: "soft",
		tone: e === "approved" ? "success" : e === "rejected" ? "danger" : "warning",
		size: "xs",
		children: e === "approved" ? "Approved" : e === "rejected" ? "Rejected" : "Revision requested"
	}) : null;
}
function rt({ rating: e }) {
	if (!e) return null;
	let t = e === "positive";
	return /* @__PURE__ */ (0, P.jsx)(h, {
		variant: "soft",
		tone: t ? "success" : "danger",
		size: "xs",
		icon: t ? L : R,
		children: t ? "Positive rating" : "Negative rating"
	});
}
function it({ comment: e, config: t }) {
	let n = Ie(t, e.status);
	if (!n) return null;
	let r = W(t, e.status), i = r === "closed" ? e.closedBy : void 0, a = r === "active" ? e.assignee : void 0, o = a?.kind === "agent" ? re : n.icon, s = a ? e.acknowledgedAt ? `Acknowledged ${new Date(e.acknowledgedAt).toLocaleString()}` : "Assigned, not yet acknowledged" : e.closedAt ? new Date(e.closedAt).toLocaleString() : void 0;
	return /* @__PURE__ */ (0, P.jsx)(h, {
		variant: "soft",
		tone: Fe(n.tone),
		size: "xs",
		...o ? { icon: o } : {},
		...s ? { title: s } : {},
		children: a ? `Assigned to ${a.name}` : i ? `Closed by ${i.name}` : n.label
	});
}
function at({ comment: e, config: t, compact: n }) {
	let r = (t.facets ?? []).flatMap((t) => {
		let r = Le(t, e.facets?.[t.key]);
		return r ? [/* @__PURE__ */ (0, P.jsx)(h, {
			variant: "soft",
			tone: Fe(r.tone),
			size: "xs",
			children: n ? r.short ?? r.label : r.label
		}, t.key)] : [];
	});
	return r.length > 0 ? /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: r }) : null;
}
function Y({ comment: e, onChecklistToggle: t }) {
	return e.checklist?.length ? /* @__PURE__ */ (0, P.jsx)("div", {
		className: "flex flex-wrap gap-1",
		children: e.checklist.map((e, n) => {
			let r = /done|complete|resolved/i.test(e.status), i = /progress|doing/i.test(e.status);
			return /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: () => t?.(n),
				disabled: !t,
				className: s("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] transition-opacity", t && "cursor-pointer hover:opacity-80", r ? "bg-green-100 text-green-800 line-through dark:bg-green-500/20 dark:text-green-300" : i ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-500/20 dark:text-yellow-300" : "bg-muted text-muted-foreground"),
				children: e.label
			}, n);
		})
	}) : null;
}
function X({ comment: e }) {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentMarkdown.tsx
function ot(e) {
	let t = e.trim();
	return t && (/^(https?:|mailto:)/i.test(t) || t.startsWith("/") && !t.startsWith("//")) ? t : null;
}
var st = /(\*\*[^*\n]+\*\*|`[^`\n]+`|\[[^\]\n]+\]\(([^)\s]+)\)|https?:\/\/[^\s<)]+|\*[^*\n]+\*)/g;
function Z(e) {
	let t = [], n = 0;
	for (let r of e.matchAll(st)) {
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
			let e = a.indexOf("]("), n = ot(a.slice(e + 2, -1));
			t.push(n ? {
				type: "link",
				text: a.slice(1, e),
				href: n
			} : {
				type: "text",
				text: a
			});
		} else if (/^https?:\/\//i.test(a)) {
			let e = ot(a);
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
function ct({ text: e }) {
	return /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: Z(e).map((e, t) => {
		let n = `${e.type}-${t}`;
		switch (e.type) {
			case "strong": return /* @__PURE__ */ (0, P.jsx)("strong", {
				className: "font-semibold",
				children: e.text
			}, n);
			case "em": return /* @__PURE__ */ (0, P.jsx)("em", { children: e.text }, n);
			case "code": return /* @__PURE__ */ (0, P.jsx)("code", {
				className: "rounded bg-muted px-1 py-0.5 font-mono text-[0.9em]",
				children: e.text
			}, n);
			case "link": return /* @__PURE__ */ (0, P.jsx)("a", {
				href: e.href,
				target: "_blank",
				rel: "noreferrer",
				className: "text-primary underline underline-offset-2",
				children: e.text
			}, n);
			default: return /* @__PURE__ */ (0, P.jsx)(N.Fragment, { children: e.text }, n);
		}
	}) });
}
function lt(e, t) {
	return /* @__PURE__ */ (0, P.jsx)("p", { children: e.map((e, t) => /* @__PURE__ */ (0, P.jsxs)(N.Fragment, { children: [t > 0 && /* @__PURE__ */ (0, P.jsx)("br", {}), /* @__PURE__ */ (0, P.jsx)(ct, { text: e })] }, t)) }, t);
}
function ut(e, t) {
	return /* @__PURE__ */ (0, P.jsx)("pre", {
		className: "overflow-auto rounded bg-muted p-2 text-xs",
		children: /* @__PURE__ */ (0, P.jsx)("code", { children: e.join("\n") })
	}, t);
}
function dt({ text: e, className: t }) {
	let n = String(e ?? "").replace(/\r\n?/g, "\n").split("\n"), r = [], i = [], a = [], o = null;
	function c() {
		i.length > 0 && (r.push(lt(i, `p-${r.length}`)), i = []);
	}
	function l() {
		a.length > 0 && (r.push(/* @__PURE__ */ (0, P.jsx)("ul", {
			className: "list-disc space-y-0.5 pl-4",
			children: a.map((e, t) => /* @__PURE__ */ (0, P.jsx)("li", { children: /* @__PURE__ */ (0, P.jsx)(ct, { text: e }) }, t))
		}, `ul-${r.length}`)), a = []);
	}
	for (let e of n) {
		if (e.trim().startsWith("```")) {
			c(), l(), o ? (r.push(ut(o, `code-${r.length}`)), o = null) : o = [];
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
	return c(), l(), o && r.push(ut(o, `code-${r.length}`)), /* @__PURE__ */ (0, P.jsx)("div", {
		className: s("space-y-2 [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", t),
		children: r.length > 0 ? r : null
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/comment-assign.ts
function ft(e, t) {
	return e.name === t.name && (e.kind ?? "user") === (t.kind ?? "user");
}
function pt(e) {
	let { comment: t, config: n, onAssign: r, extraAssignItems: i, run: a } = e, o = [];
	if (r) {
		for (let { id: e, name: i, kind: s } of n.assignees ?? []) {
			let n = {
				id: e,
				name: i,
				kind: s
			}, c = t.assignee !== void 0 && ft(t.assignee, n);
			o.push({
				label: i,
				icon: s === "agent" ? re : ke,
				disabled: c,
				title: c ? `Already assigned to ${i}` : `Assign to ${i}`,
				onSelect: () => a("assign", () => r(n))
			});
		}
		t.assignee && o.push({
			label: "Unassign",
			icon: je,
			onSelect: () => a("assign", () => r(null))
		});
	}
	return o.push(...i?.(t) ?? []), o;
}
function mt(e) {
	let { comment: t, config: n, onAcknowledge: r, run: i } = e;
	if (t.parentId || W(n, t.status) !== "active") return [];
	let a = [];
	r && !t.assignee && a.push({
		label: "Acknowledge",
		icon: Me,
		onSelect: () => i("acknowledge", r)
	});
	let o = pt(e);
	return o.length > 0 && a.push({
		label: "Assign…",
		icon: Ae,
		onSelect: () => {},
		children: o
	}), a;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentCard.tsx
function ht(e, t) {
	let n = [];
	return t.onDelete && n.push({
		label: "Delete",
		icon: S,
		onSelect: t.onDelete
	}), e || n.push(...t.statusActions ?? []), t.onCopy && n.push({
		label: "Copy",
		icon: k,
		onSelect: t.onCopy
	}), t.onMaximize && n.push({
		label: "Maximise",
		icon: T,
		onSelect: t.onMaximize
	}), n;
}
function Q(e) {
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
function gt(e, t) {
	let { comment: n, config: r, onAcknowledge: i, onAssign: a, extraAssignItems: o } = e;
	return mt({
		comment: n,
		config: r,
		run: t,
		...i ? { onAcknowledge: i } : {},
		...a ? { onAssign: a } : {},
		...o ? { extraAssignItems: o } : {}
	});
}
function _t(e) {
	let { comment: t, config: n, renderBody: r, onUpdateStatus: i, onClose: a, onDelete: o, onCopy: s, onReply: c, onChecklistToggle: l, onMaximize: d } = e, [f, p] = (0, N.useState)(""), m = !!t.parentId, h = ae(t.createdAt), g = G(n, "resolved"), _ = G(n, "active"), v = G(n, "closed"), y = W(n, t.status), S = async (e) => {
		p("");
		try {
			await i?.(e);
		} catch (e) {
			p(`Couldn't update comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, T = async () => {
		p("");
		try {
			await a?.();
		} catch (e) {
			p(`Couldn't close comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, E = async () => {
		p("");
		try {
			await s?.();
		} catch (e) {
			p(`Couldn't copy comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}, D = ht(m, {
		...o ? { onDelete: o } : {},
		statusActions: [
			...gt(e, Q(p)),
			...y === "active" && i && g ? [{
				label: "Resolve",
				icon: te,
				onSelect: () => void S(g.value)
			}] : [],
			...(y === "active" || y === "resolved") && a && v ? [{
				label: "Close",
				icon: b,
				onSelect: () => void T()
			}] : [],
			...(y === "resolved" || y === "closed") && i && _ ? [{
				label: "Reopen",
				icon: C,
				onSelect: () => void S(_.value)
			}] : []
		],
		...s ? { onCopy: () => void E() } : {},
		...d ? { onMaximize: d } : {}
	});
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "space-y-2.5",
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex items-start gap-2.5",
				children: [
					/* @__PURE__ */ (0, P.jsx)(K, { author: t.author }),
					/* @__PURE__ */ (0, P.jsxs)("div", {
						className: "min-w-0 flex-1",
						children: [/* @__PURE__ */ (0, P.jsxs)("div", {
							className: "flex min-w-0 items-baseline gap-1.5",
							children: [/* @__PURE__ */ (0, P.jsx)("span", {
								className: "truncate text-xs font-semibold text-foreground",
								children: B(t.author)
							}), h && /* @__PURE__ */ (0, P.jsx)("span", {
								title: h.toLocaleString(),
								className: "shrink-0 text-[10px] text-muted-foreground",
								children: w(h)
							})]
						}), !m && /* @__PURE__ */ (0, P.jsxs)("div", {
							className: "mt-1 flex flex-wrap gap-1",
							children: [
								/* @__PURE__ */ (0, P.jsx)(it, {
									comment: t,
									config: n
								}),
								/* @__PURE__ */ (0, P.jsx)(rt, { rating: t.rating }),
								/* @__PURE__ */ (0, P.jsx)(nt, { decision: t.reviewDecision }),
								/* @__PURE__ */ (0, P.jsx)(at, {
									comment: t,
									config: n
								})
							]
						})]
					}),
					/* @__PURE__ */ (0, P.jsxs)("div", {
						className: "flex shrink-0 items-center gap-0.5",
						children: [
							!m && (y === "active" || y === "resolved") && a && v && /* @__PURE__ */ (0, P.jsx)("button", {
								type: "button",
								"aria-label": "Close comment",
								title: "Close comment",
								onClick: () => void T(),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, P.jsx)(u, {
									icon: te,
									className: "text-sm"
								})
							}),
							!m && y === "closed" && i && _ && /* @__PURE__ */ (0, P.jsx)("button", {
								type: "button",
								"aria-label": "Reopen comment",
								title: "Reopen comment",
								onClick: () => void S(_.value),
								className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [@media(pointer:coarse)]:size-11",
								children: /* @__PURE__ */ (0, P.jsx)(u, {
									icon: C,
									className: "text-sm"
								})
							}),
							D.length > 0 && /* @__PURE__ */ (0, P.jsx)(ee, {
								align: "right",
								menuClassName: "min-w-[150px]",
								trigger: /* @__PURE__ */ (0, P.jsx)("span", {
									role: "button",
									tabIndex: 0,
									"aria-label": "Comment actions",
									className: "inline-flex items-center justify-center rounded-full p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:size-11",
									children: /* @__PURE__ */ (0, P.jsx)(u, {
										icon: x,
										className: "text-sm"
									})
								}),
								items: D
							})
						]
					})
				]
			}),
			t.body.trim() && /* @__PURE__ */ (0, P.jsx)("div", {
				className: "text-[13px] leading-5 text-foreground",
				children: r ? r(t.body) : /* @__PURE__ */ (0, P.jsx)(dt, { text: t.body })
			}),
			f && /* @__PURE__ */ (0, P.jsx)("div", {
				role: "alert",
				className: "rounded-md border border-destructive/30 bg-destructive/5 px-2.5 py-2 text-xs text-destructive",
				children: f
			}),
			!m && /* @__PURE__ */ (0, P.jsx)(Y, {
				comment: t,
				...l ? { onChecklistToggle: l } : {}
			}),
			!m && /* @__PURE__ */ (0, P.jsx)(X, { comment: t }),
			c && /* @__PURE__ */ (0, P.jsx)("button", {
				type: "button",
				onClick: c,
				className: "-ml-2 rounded px-2 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary/10 [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:px-3 [@media(pointer:coarse)]:text-sm",
				children: "Reply"
			})
		]
	});
}
function vt(e) {
	let { comment: t, config: n, compact: r, defaultExpanded: i } = e, [a, o] = (0, N.useState)(i ?? !1), [c, l] = (0, N.useState)(""), d = !!t.parentId, f = /* @__PURE__ */ (0, P.jsx)(_t, { ...e });
	async function p(t) {
		l("");
		try {
			if (t === "close") {
				await e.onClose?.();
				return;
			}
			let r = G(n, t === "resolve" ? "resolved" : "active");
			r && await e.onUpdateStatus?.(r.value);
		} catch (e) {
			l(`${t === "close" ? "Couldn't close comment" : "Couldn't update comment"}: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}
	async function m() {
		l("");
		try {
			await e.onCopy?.();
		} catch (e) {
			l(`Couldn't copy comment: ${e instanceof Error ? e.message : "Unexpected error"}`);
		}
	}
	return a ? /* @__PURE__ */ (0, P.jsx)(P.Fragment, { children: /* @__PURE__ */ (0, P.jsx)("div", {
		"data-testid": "comment-card",
		"data-comment-kind": d ? "reply" : "root",
		"data-comment-id": t.id,
		className: s("group/comment rounded-xl bg-background px-3.5 py-3 text-xs", !d && "border border-border/80 shadow-sm transition-shadow hover:shadow-md", r && "px-2.5 py-2.5", d && "ml-4 border-l-2 border-l-border bg-muted/20 shadow-none"),
		children: f
	}) }) : /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsxs)("div", {
		role: "button",
		tabIndex: 0,
		onClick: () => o(!0),
		onKeyDown: (e) => {
			(e.key === "Enter" || e.key === " ") && (e.preventDefault(), o(!0));
		},
		"data-testid": "comment-card",
		"data-comment-kind": d ? "reply" : "root",
		"data-comment-id": t.id,
		className: s("group/comment flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs transition-colors hover:bg-muted/50", !d && "border border-border bg-background shadow-sm", d && "ml-4 border-l-2 border-l-border bg-muted/30"),
		children: [
			/* @__PURE__ */ (0, P.jsx)(K, {
				author: t.author,
				bare: !0
			}),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "min-w-0 truncate text-[11px] font-semibold text-foreground",
				children: B(t.author)
			}),
			!d && /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
				/* @__PURE__ */ (0, P.jsx)(it, {
					comment: t,
					config: n
				}),
				/* @__PURE__ */ (0, P.jsx)(rt, { rating: t.rating }),
				/* @__PURE__ */ (0, P.jsx)(nt, { decision: t.reviewDecision }),
				/* @__PURE__ */ (0, P.jsx)(at, {
					comment: t,
					config: n,
					compact: !0
				})
			] }),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "min-w-0 flex-1 truncate text-[12px] text-foreground/80",
				children: t.body.trim() ? He(t.body) : t.rating === "positive" ? "Positive rating" : t.rating === "negative" ? "Negative rating" : t.reviewDecision === "approved" ? "Approved" : "Empty comment"
			}),
			/* @__PURE__ */ (0, P.jsx)("span", {
				className: "shrink-0",
				role: "presentation",
				onClick: (e) => e.stopPropagation(),
				onKeyDown: (e) => e.stopPropagation(),
				children: /* @__PURE__ */ (0, P.jsx)(ee, {
					align: "right",
					trigger: /* @__PURE__ */ (0, P.jsx)("span", {
						role: "button",
						tabIndex: 0,
						"aria-label": "Comment actions",
						className: "rounded p-0.5 text-muted-foreground opacity-0 transition-opacity hover:bg-muted hover:opacity-100 group-hover/comment:opacity-70 [@media(hover:none)]:opacity-70 [@media(pointer:coarse)]:p-2.5",
						children: /* @__PURE__ */ (0, P.jsx)(u, {
							icon: x,
							className: "text-sm"
						})
					}),
					items: ht(d, {
						...e.onDelete ? { onDelete: e.onDelete } : {},
						statusActions: [
							...gt(e, Q(l)),
							...W(n, t.status) === "active" && e.onUpdateStatus && G(n, "resolved") ? [{
								label: "Resolve",
								icon: te,
								onSelect: () => void p("resolve")
							}] : [],
							...(W(n, t.status) === "active" || W(n, t.status) === "resolved") && e.onClose && G(n, "closed") ? [{
								label: "Close",
								icon: b,
								onSelect: () => void p("close")
							}] : [],
							...(W(n, t.status) === "resolved" || W(n, t.status) === "closed") && e.onUpdateStatus && G(n, "active") ? [{
								label: "Reopen",
								icon: C,
								onSelect: () => void p("reopen")
							}] : []
						],
						...e.onCopy ? { onCopy: () => void m() } : {},
						...e.onMaximize ? { onMaximize: e.onMaximize } : {}
					})
				})
			})
		]
	}), c && /* @__PURE__ */ (0, P.jsx)("div", {
		role: "alert",
		className: "rounded-md border border-destructive/30 bg-destructive/5 px-2.5 py-2 text-xs text-destructive",
		children: c
	})] });
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentThreadList.tsx
function yt({ onSubmit: e, onCancel: t }) {
	let [n, r] = (0, N.useState)(""), [i, a] = (0, N.useState)(!1), o = (0, N.useRef)(!1);
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
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "ml-6 flex gap-1.5 py-1",
		"aria-busy": i,
		children: [/* @__PURE__ */ (0, P.jsx)("input", {
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
		}), /* @__PURE__ */ (0, P.jsx)(c, {
			size: "sm",
			className: "h-7 text-xs [@media(pointer:coarse)]:h-11 [@media(pointer:coarse)]:px-4 [@media(pointer:coarse)]:text-sm",
			disabled: i || !n.trim(),
			onClick: () => void s(),
			children: "Send"
		})]
	});
}
function bt({ comments: e, config: t, compact: n = !1, defaultExpanded: r = !1, renderRootMeta: i, renderRootDetails: a, getThreadProps: o, renderBody: c, threadToMarkdown: l, onUpdateStatus: u, onClose: f, onAcknowledge: p, onAssign: m, extraAssignItems: h, onChecklistToggle: g, onDelete: _, onReply: v }) {
	let [y, b] = (0, N.useState)(null), [x, S] = (0, N.useState)(null), [C, w] = (0, N.useState)("thread"), [T, E] = (0, N.useState)(null), ee = V(e), te = H(e);
	if (ee.length === 0) return null;
	let D = ee.find((e) => e.id === x), O = D ? [D, ...U(te.get(D.id) ?? [])] : [];
	function k() {
		S(null), E(null), w("thread");
	}
	function re(e) {
		return {
			...u ? { onUpdateStatus: (t) => u(e, t) } : {},
			...f ? { onClose: () => f(e) } : {},
			...p ? { onAcknowledge: () => p(e) } : {},
			...m ? { onAssign: (t) => m(e, t) } : {},
			...h ? { extraAssignItems: h } : {},
			...g ? { onChecklistToggle: (t) => g(e, t) } : {},
			..._ ? { onDelete: () => _(e) } : {}
		};
	}
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsx)("div", {
		className: "space-y-2",
		children: ee.map((e) => {
			let u = U(te.get(e.id) ?? []), d = [e, ...u], f = o?.(e) ?? {};
			return /* @__PURE__ */ (0, P.jsxs)("div", {
				"data-testid": "comment-thread-block",
				"data-comment-id": e.id,
				...f,
				className: s("space-y-1.5 p-1", f.className),
				children: [
					i ? /* @__PURE__ */ (0, P.jsx)("div", {
						className: "px-1",
						children: i(e)
					}) : null,
					/* @__PURE__ */ (0, P.jsx)(vt, {
						comment: e,
						config: t,
						compact: n,
						defaultExpanded: r,
						onMaximize: () => S(e.id),
						...l ? { onCopy: () => ne(l(d)) } : {},
						...c ? { renderBody: c } : {},
						...re(e.id),
						...v ? { onReply: () => b(e.id) } : {}
					}),
					a?.(e),
					u.length > 0 && /* @__PURE__ */ (0, P.jsx)("div", {
						className: "space-y-1.5",
						children: u.map((i) => /* @__PURE__ */ (0, P.jsx)(vt, {
							comment: i,
							config: t,
							compact: n,
							defaultExpanded: r,
							onMaximize: () => S(e.id),
							...l ? { onCopy: () => ne(l(d)) } : {},
							...c ? { renderBody: c } : {},
							...re(i.id)
						}, i.id))
					}),
					v && y === e.id && /* @__PURE__ */ (0, P.jsx)(yt, {
						onCancel: () => b(null),
						onSubmit: async (t) => {
							await v(e, t), b(null);
						}
					})
				]
			}, e.id);
		})
	}), D && /* @__PURE__ */ (0, P.jsx)(d, {
		open: !0,
		onClose: k,
		title: "Comment thread",
		size: "lg",
		children: /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "space-y-3",
			"data-testid": "comment-thread-modal",
			children: [l && /* @__PURE__ */ (0, P.jsx)(we, {
				tabs: [{
					id: "thread",
					label: "Thread"
				}, {
					id: "markdown",
					label: "Markdown"
				}],
				value: C,
				onChange: (e) => w(e === "markdown" ? "markdown" : "thread")
			}), C === "markdown" && l ? /* @__PURE__ */ (0, P.jsx)(me, { text: l(O) }) : /* @__PURE__ */ (0, P.jsx)("div", {
				className: "space-y-2",
				children: O.map((e) => /* @__PURE__ */ (0, P.jsxs)("div", {
					className: "space-y-1.5",
					children: [
						/* @__PURE__ */ (0, P.jsx)(vt, {
							comment: e,
							config: t,
							defaultExpanded: !0,
							...l ? { onCopy: () => ne(l(O)) } : {},
							...c ? { renderBody: c } : {},
							...re(e.id),
							...v ? { onReply: () => E(e.id) } : {}
						}),
						e.id === D.id && a?.(D),
						v && T === e.id && /* @__PURE__ */ (0, P.jsx)(yt, {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/comments/CommentThread.tsx
function xt({ comments: e, config: t, anchor: n, compact: r = !1, autoFocusComposer: i = !1, defaultExpanded: a = !1, composerPlaceholder: o, hideComposer: c = !1, createActions: l, renderRootMeta: u, renderRootDetails: d, renderBody: f, threadToMarkdown: p, extraAssignItems: m, onCreate: h, onReply: g, onUpdateStatus: _, onClose: v, onAcknowledge: y, onAssign: b, onDelete: x, onChecklistToggle: S, onMention: C }) {
	let w = We(e, t, {
		..._ ? { onUpdateStatus: _ } : {},
		...v ? { onClose: v } : {},
		...y ? { onAcknowledge: y } : {},
		...b ? { onAssign: b } : {},
		...x ? { onDelete: x } : {},
		...S ? { onChecklistToggle: S } : {},
		...g ? { onReply: g } : {},
		...C ? { onMention: C } : {}
	});
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		className: s(r && "p-2"),
		children: [/* @__PURE__ */ (0, P.jsx)(bt, {
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
		}), !c && h && /* @__PURE__ */ (0, P.jsx)("div", {
			className: "pt-1",
			children: /* @__PURE__ */ (0, P.jsx)(tt, {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/findings-comments.ts
function $(e) {
	return `line:${e}`;
}
function St(e, t) {
	return `L${e}-${t}`;
}
var Ct = /^line:(\d+)$/, wt = /^L(\d+)-(\d+)$/;
function Tt(e) {
	let t = Ct.exec(e);
	if (!t) throw Error(`file-findings: expected a "line:<n>" anchor, got ${JSON.stringify(e)}`);
	return Number(t[1]);
}
var Et = (e, t) => {
	if (e === "__document__" || wt.test(e)) return t.includes(e) ? e : null;
	let n = Tt(e), r = t.flatMap((e) => {
		let t = wt.exec(e);
		return t ? [{
			anchor: e,
			start: Number(t[1]),
			end: Number(t[2])
		}] : [];
	}), i = r.filter((e) => e.start <= n && n <= e.end).sort((e, t) => e.end - e.start - (t.end - t.start) || t.start - e.start)[0];
	if (i) return i.anchor;
	let a = r.filter((e) => e.start <= n).sort((e, t) => t.start - e.start)[0];
	return a ? a.anchor : t.includes("__document__") ? z : null;
}, Dt = {
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
}, Ot = "1970-01-01T00:00:00.000Z", kt = 6;
function At(e) {
	return e.map((e, t) => ({
		id: `finding-${String(t).padStart(kt, "0")}`,
		body: e.rule ? `${e.message}\n\n\`${e.rule}\`` : e.message,
		createdAt: Ot,
		author: {
			name: e.source,
			kind: "agent"
		},
		status: e.severity,
		anchor: $(e.line),
		facets: {
			severity: e.severity,
			source: e.source,
			kind: e.kind
		}
	}));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/LazyAnnotatedCode.tsx
var jt = (0, N.lazy)(() => import("./AnnotatedCode-BEzvyG8W.js").then((e) => ({ default: e.AnnotatedCode })));
function Mt(e) {
	return /* @__PURE__ */ (0, P.jsx)(N.Suspense, {
		fallback: /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
			children: [/* @__PURE__ */ (0, P.jsx)(u, {
				icon: de,
				className: "mr-2 animate-spin"
			}), "Loading editor"]
		}),
		children: /* @__PURE__ */ (0, P.jsx)(jt, { ...e })
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/mdx-source.ts
var Nt = /^\s{0,3}(`{3,}|~{3,})/, Pt = /^(import|export)\s/;
function Ft(e) {
	let t = 0;
	for (let n of e) "{[(".includes(n) ? t += 1 : "}])".includes(n) && --t;
	return t;
}
function It(e) {
	let t = e.split("\n"), n = t.slice(), r = 0;
	if (t[0] === "---") {
		let e = t.indexOf("---", 1);
		if (e !== -1) for (; r <= e; r += 1) n[r] = "";
	}
	let i = null;
	for (; r < t.length; r += 1) {
		let e = t[r], a = Nt.exec(e);
		if (i) {
			a && a[1][0] === i[0] && a[1].length >= i.length && (i = null);
			continue;
		}
		if (a) {
			i = a[1];
			continue;
		}
		if (!Pt.test(e)) continue;
		let o = Ft(e);
		for (n[r] = ""; o > 0 && r + 1 < t.length;) r += 1, o += Ft(t[r]), n[r] = "";
	}
	return n.join("\n");
}
var Lt = "[data-source-start]", Rt = /^h([1-6])$/, zt = 2;
function Bt(e) {
	return `q:${JSON.stringify([
		e.path,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function Vt(e) {
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
function Ht(e) {
	let t = [];
	return [...e.querySelectorAll(Lt)].filter((e) => !e.parentElement?.closest(Lt)).map((e) => {
		let n = Rt.exec(e.tagName.toLowerCase())?.[1];
		if (n) {
			for (; t.length > 0 && (t.at(-1)?.level ?? 0) >= Number(n);) t.pop();
			let r = Gt(e).replace(/\s+/g, " ").trim().slice(0, 200);
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
var Ut = (e) => Ht(e);
function Wt(e) {
	let t = document.createTreeWalker(e, NodeFilter.SHOW_TEXT), n = [];
	for (let e = t.nextNode(); e; e = t.nextNode()) e.parentElement?.closest("[data-comment-author]") || n.push(e);
	return n;
}
function Gt(e) {
	return Wt(e).map((e) => e.data).join("");
}
function Kt(e, t, n) {
	let r = document.createRange(), i = 0, a = !1;
	for (let o of Wt(e)) {
		if (!a && t < i + o.length && (r.setStart(o, t - i), a = !0), a && n <= i + o.length) return r.setEnd(o, n - i), r;
		i += o.length;
	}
	throw Error(`quote-anchor: offsets ${t}-${n} lie outside the block`);
}
function qt(e) {
	return e.trim().split(/\s+/).filter(Boolean).map((e) => e.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("\\s*");
}
function Jt(e, t, n) {
	let r = n.prefix && t.prefix.trim() ? `(?<=${qt(t.prefix)}\\s*)` : "", i = n.suffix && t.suffix.trim() ? `(?=\\s*${qt(t.suffix)})` : "", a = RegExp(`${r}${qt(t.quote)}${i}`, "g");
	return e.flatMap(({ block: e }) => [...Gt(e).matchAll(a)].map((t) => ({
		block: e,
		start: t.index,
		end: t.index + t[0].length
	})));
}
var Yt = (e, t) => e.length === t.length && e.every((e, n) => e === t[n]), Xt = [
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
function Zt(e, t) {
	let n = Ut(e), r = n.filter((e) => Yt(e.path, t.path));
	for (let e of Xt) {
		let n = Jt(r, t, e);
		if (n.length === 1) return Kt(n[0].block, n[0].start, n[0].end);
	}
	let i = Jt(n, t, Xt[0]);
	return i.length === 1 ? Kt(i[0].block, i[0].start, i[0].end) : null;
}
function Qt(e, t) {
	let n = Ut(e), r = n.find(({ block: e }) => e.contains(t.startContainer) && e.contains(t.endContainer));
	if (!r) return null;
	let i = Gt(r.block), a = document.createRange();
	a.setStart(r.block, 0), a.setEnd(t.startContainer, t.startOffset);
	let o = on(t), s = on(a).length + (o.length - o.trimStart().length), c = on(a).length + o.trimEnd().length;
	if (c <= s) return null;
	let l = i.slice(0, s).split(/\s+/).filter(Boolean), u = i.slice(c).split(/\s+/).filter(Boolean), d = n.filter((e) => Yt(e.path, r.path));
	for (let e = zt;; e++) {
		let t = {
			path: r.path,
			prefix: l.slice(-e).join(" "),
			quote: i.slice(s, c),
			suffix: u.slice(0, e).join(" ")
		}, n = Jt(d, t, Xt[0]);
		if (n.length === 1 && n[0].block === r.block && n[0].start === s) return t;
		if (e >= l.length && e >= u.length) return null;
	}
}
var $t = /[*_`|>#~]|(?:^|\s)[-+](?=\s)/g;
function en(e) {
	return qt(e.replace($t, " ")).replaceAll("\\s*", "[\\s*_`|>#~]*");
}
function tn(e, t) {
	return Ut(e).map(({ block: e }) => {
		let n = Number(e.dataset.sourceStart), r = Number(e.dataset.sourceEnd);
		return {
			block: e,
			first: n,
			last: r,
			text: t.slice(n - 1, r).join("\n")
		};
	});
}
var nn = (e, t) => [...e.matchAll(new RegExp(t, "g"))];
function rn(e, t, n) {
	let r = t[n.line - 1]?.slice(n.start, n.end) ?? "";
	if (!r.replace($t, " ").trim()) return null;
	let i = tn(e, t).find((e) => n.line >= e.first && n.line <= e.last);
	if (!i) return null;
	let a = t.slice(i.first - 1, n.line - 1).reduce((e, t) => e + t.length + 1, 0) + n.start, o = nn(i.text, en(r)).findIndex((e) => e.index + e[0].length > a && e.index < a + r.length), s = o < 0 ? void 0 : nn(Gt(i.block), qt(r.replace($t, " ")))[o];
	return s ? Qt(e, Kt(i.block, s.index, s.index + s[0].length)) : null;
}
function an(e, t, n) {
	let r = Zt(e, n), i = r && tn(e, t).find(({ block: e }) => e.contains(r.startContainer));
	if (!r || !i) return [];
	let a = document.createRange();
	a.setStart(i.block, 0), a.setEnd(r.startContainer, r.startOffset);
	let o = on(a).length, s = nn(Gt(i.block), qt(n.quote)).findIndex((e) => e.index === o), c = s < 0 ? void 0 : nn(i.text, en(n.quote))[s];
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
function on(e) {
	let t = e.cloneContents();
	for (let e of t.querySelectorAll("[data-comment-author]")) e.remove();
	return t.textContent ?? "";
}
function sn(e) {
	let t = e.path.at(-1);
	return `${t ? `${t} › ` : ""}“${e.quote}”`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/geometry.ts
var cn = "[data-source-start]";
function ln(e) {
	return (e instanceof Element ? e : e?.parentElement ?? null)?.closest(cn) ?? null;
}
function un({ origin: e, markdown: t, anchors: n, mode: r }) {
	let i = e.getBoundingClientRect(), a = (e) => ({
		top: e.top - i.top,
		left: e.left - i.left,
		width: e.width,
		height: e.height
	}), o = /* @__PURE__ */ new Map();
	for (let e of t.querySelectorAll(cn)) {
		let t = St(Number(e.dataset.sourceStart), Number(e.dataset.sourceEnd));
		o.has(t) || o.set(t, e);
	}
	let s = [];
	for (let e of n) {
		if (e.startsWith("q:")) {
			let n = Zt(t, Vt(e));
			if (!n) continue;
			let i = [...n.getClientRects()].filter((e) => e.width > 0).map(a), o = ln(n.startContainer), c = r === "text" ? i : o ? [a(o.getBoundingClientRect())] : [];
			c.length > 0 && s.push({
				anchor: e,
				boxes: c,
				at: i[0]?.top ?? c[0].top
			});
			continue;
		}
		let n = Et(e, [...o.keys()]), i = n ? o.get(n) : void 0;
		if (i) {
			let t = document.createRange();
			t.selectNodeContents(i);
			let n = dn([...t.getClientRects()].filter((e) => e.width > 0).map(a)), r = n.length > 0 ? n : [a(i.getBoundingClientRect())];
			s.push({
				anchor: e,
				boxes: r,
				at: r[0].top
			});
		}
	}
	return new Map(s.sort((e, t) => e.at - t.at).map(({ anchor: e, boxes: t }) => [e, t]));
}
function dn(e) {
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
function fn({ cards: e, focused: t, gap: n = 8 }) {
	let r = [...e].sort((e, t) => e.desired - t.desired || e.order - t.order), i = r.map((e) => e.desired), a = r.findIndex((e) => e.anchor === t);
	if (a >= 0) for (let e = a - 1; e >= 0; e--) i[e] = Math.min(i[e], i[e + 1] - n - r[e].height);
	for (let e = 0; e < r.length; e++) {
		let t = e === 0 ? 0 : i[e - 1] + r[e - 1].height + n;
		i[e] = Math.max(i[e], t);
	}
	return new Map(r.map((e, t) => [e.anchor, i[t]]));
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/hooks.ts
var pn = 640;
function mn() {
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
function hn({ origin: e, markdown: t, anchors: n, mode: r, visible: i }) {
	let [a, o] = (0, N.useState)(/* @__PURE__ */ new Map()), [s, c] = (0, N.useState)(0), [l, u] = (0, N.useState)(0);
	return (0, N.useLayoutEffect)(() => {
		let a = e.current, s = t.current;
		if (!a || !s || !i) return;
		let l = () => {
			o(un({
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
function gn(e) {
	let [t, n] = (0, N.useState)(!1);
	return (0, N.useLayoutEffect)(() => {
		let t = e.current;
		if (!t) return;
		let r = new ResizeObserver(() => n(t.offsetWidth < pn));
		return r.observe(t), () => r.disconnect();
	}, [e]), t;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/marks.ts
var _n = {
	error: 3,
	warning: 2,
	info: 1
}, vn = {
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
function yn(e) {
	if (e === "error" || e === "warning" || e === "info") return e;
	throw Error(`lint finding has severity ${JSON.stringify(e)}, expected error, warning or info`);
}
var bn = (e) => `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='6' height='3' viewBox='0 0 6 3'%3E%3Cpath d='M0 2.5 1.5 0.5 3 2.5 4.5 0.5 6 2.5' fill='none' stroke='${encodeURIComponent(e)}' stroke-width='1'/%3E%3C/svg%3E")`, xn = (e) => ({
	backgroundImage: bn(e),
	backgroundRepeat: "repeat-x",
	backgroundPosition: "bottom",
	backgroundSize: "6px 3px",
	paddingBottom: 2
}), Sn = (e, t) => _n[yn(t.severity)] - _n[yn(e.severity)] || e.line - t.line, Cn = [
	"error",
	"warning",
	"info"
];
function wn(e) {
	let t = {
		error: 0,
		warning: 0,
		info: 0
	};
	for (let n of e) t[n.severity] += 1;
	return t;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/problem-filters.ts
function Tn() {
	return {
		severity: /* @__PURE__ */ new Map(),
		type: /* @__PURE__ */ new Map()
	};
}
function En(e) {
	return e.rule || e.source;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/ProblemsList.tsx
var Dn = {
	error: "Errors",
	warning: "Warnings",
	info: "Info"
};
function On(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) n.set(t(r), (n.get(t(r)) ?? 0) + 1);
	return n;
}
function kn({ source: e, tokens: t, findings: n, filters: r, onFiltersChange: i, onOpen: a, openLabel: o, getFindingActions: s }) {
	let c = (0, N.useMemo)(() => e.split("\n"), [e]), l = (0, N.useMemo)(() => {
		let e = j(n.map((e) => [
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
	}, [n]), u = (0, N.useMemo)(() => l.map((e) => e.key), [l]), [d, p] = be(u), m = l.filter(({ finding: e }) => D(e.severity, r.severity) && D(En(e), r.type)).sort((e, t) => Sn(e.finding, t.finding)), h = On(n.filter((e) => D(En(e), r.type)), (e) => e.severity), g = On(n.filter((e) => D(e.severity, r.severity)), En), _ = [.../* @__PURE__ */ new Set([...n.map(En), ...r.type.keys()])].sort(), v = r.severity.size + r.type.size;
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex items-center justify-between border-b border-border p-1 text-xs",
		"data-problem-filters": "",
		children: [/* @__PURE__ */ (0, P.jsxs)("span", {
			className: "px-2 text-muted-foreground",
			children: [
				m.length,
				" of ",
				n.length,
				" problems"
			]
		}), /* @__PURE__ */ (0, P.jsxs)(ye, {
			active: v,
			children: [
				/* @__PURE__ */ (0, P.jsx)(xe, {
					label: "Severity",
					options: Cn.map((e) => ({
						value: e,
						label: Dn[e],
						count: h.get(e) ?? 0
					})),
					state: r.severity,
					onChange: (e) => i({
						...r,
						severity: e
					})
				}),
				/* @__PURE__ */ (0, P.jsx)(xe, {
					label: "Issue Type",
					options: _.map((e) => ({
						value: e,
						label: e,
						count: g.get(e) ?? 0
					})),
					state: r.type,
					onChange: (e) => i({
						...r,
						type: e
					})
				}),
				/* @__PURE__ */ (0, P.jsx)("button", {
					type: "button",
					className: "text-xs text-muted-foreground",
					disabled: !v,
					onClick: () => i(Tn()),
					children: "Clear filters"
				})
			]
		})]
	}), m.length === 0 ? /* @__PURE__ */ (0, P.jsx)("p", {
		className: "p-6 text-center text-xs text-muted-foreground",
		children: n.length ? "No problems match the current filters." : "No problems in this file."
	}) : /* @__PURE__ */ (0, P.jsx)("ul", { children: m.map(({ finding: e, key: n }) => {
		let r = Math.max(1, e.line - 3), i = Math.min(c.length, (e.endLine ?? e.line) + 3);
		return /* @__PURE__ */ (0, P.jsx)("li", {
			"data-problem-line": e.line,
			children: /* @__PURE__ */ (0, P.jsx)(ge, {
				message: e.message,
				severity: e.severity,
				metadata: `${e.source} ${e.rule ?? ""}`,
				location: `${e.line}${e.column ? `:${e.column}` : ""}`,
				expanded: d.includes(n),
				onExpandedChange: (e) => p((t) => {
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
								style: a >= e.line && a <= (e.endLine ?? e.line) ? xn(vn[e.severity].stroke) : void 0,
								children: /* @__PURE__ */ (0, P.jsx)(f, {
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/blame.ts
var An = (e) => /^0+$/.test(e.sha);
function jn(e, t) {
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
var Mn = 864e5;
function Nn(e, t = /* @__PURE__ */ new Date()) {
	let n = Math.floor((t.getTime() - Date.parse(e)) / Mn);
	return n < 1 ? "today" : n < 30 ? `${n}d` : n < 365 ? `${Math.floor(n / 30)}mo` : `${Math.floor(n / 365)}y`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/source-marks.ts
var Pn = (e) => JSON.parse(e.dataset.sourceComment);
function Fn(e, t, n, r) {
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
function In(e) {
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
function Ln(e) {
	let t = In(e);
	return !t || t.start.line !== t.end.line ? null : {
		line: t.start.line,
		start: t.start.column,
		end: t.end.column
	};
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/SourceView.tsx
function Rn({ blame: e }) {
	if (!e) return /* @__PURE__ */ (0, P.jsx)("span", {});
	let { commit: t, first: n, recency: r } = e, i = An(t);
	return /* @__PURE__ */ (0, P.jsxs)("span", {
		"data-blame-sha": n ? t.sha : void 0,
		title: i ? "Not committed yet" : `${t.sha.slice(0, 12)} · ${t.author} · ${t.date.slice(0, 10)}\n${t.summary}`,
		className: s("relative flex min-w-0 items-start gap-1.5 pl-2 pr-2 font-sans text-[11px] text-muted-foreground", n && "border-t border-border/60"),
		children: [/* @__PURE__ */ (0, P.jsx)("span", {
			"aria-hidden": !0,
			className: s("absolute inset-y-0 left-0 w-[3px]", i ? "bg-amber-500" : "bg-sky-500"),
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
				children: Nn(t.date)
			})
		] }))]
	});
}
function zn({ source: e, tokens: t, blame: n, findings: r, showFindings: i, marks: a, comments: o, selected: c, onSelect: l, onFocus: d, getFindingActions: f }) {
	return /* @__PURE__ */ (0, P.jsx)("ol", {
		className: "py-2 font-mono text-xs leading-5",
		children: e.split("\n").map((e, p) => {
			let m = p + 1, h = r.filter((e) => e.line === m).sort(Sn), g = h[0] ? vn[yn(h[0].severity)] : null, v = a.filter((e) => e.spans.at(-1)?.line === m), y = i || c === $(m) ? h : [];
			return /* @__PURE__ */ (0, P.jsxs)("li", {
				"data-source-line": m,
				className: s("grid pr-3", n ? "grid-cols-[13rem_1.25rem_2.5rem_1fr]" : "grid-cols-[1.25rem_2.5rem_1fr]"),
				children: [
					n && /* @__PURE__ */ (0, P.jsx)(Rn, { blame: n.get(m) }),
					g && !i ? /* @__PURE__ */ (0, P.jsx)("button", {
						type: "button",
						"data-source-mark": m,
						"aria-label": `${h.length} ${h.length === 1 ? "finding" : "findings"} on line ${m}`,
						"aria-pressed": c === $(m),
						onClick: () => c === $(m) ? d(null) : l(h[0]),
						className: s("flex h-5 items-center justify-center rounded-sm hover:bg-muted", g.tone),
						children: /* @__PURE__ */ (0, P.jsx)(u, {
							icon: g.icon,
							className: "text-[13px]"
						})
					}) : /* @__PURE__ */ (0, P.jsx)("span", {
						className: s("flex h-5 items-center justify-center", g?.tone),
						children: g && /* @__PURE__ */ (0, P.jsx)(u, {
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
								style: g ? xn(g.stroke) : void 0,
								children: e ? Fn(e, m, a, t?.[p]).map((e) => {
									let t = e.token ? _(e.token) : void 0;
									return e.anchors.length === 0 ? /* @__PURE__ */ (0, P.jsx)("span", {
										style: t,
										children: e.text
									}, e.start) : /* @__PURE__ */ (0, P.jsx)("mark", {
										"data-source-comment": JSON.stringify(e.anchors),
										onClick: () => window.getSelection()?.isCollapsed && d(e.anchors[0]),
										style: t,
										className: s("cursor-pointer rounded-sm text-inherit", e.anchors.includes(c ?? "") ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25"),
										children: e.text
									}, e.start);
								}) : " "
							}),
							v.map(({ anchor: e }) => {
								let t = Re(o, e), n = V(t)[0];
								return n ? /* @__PURE__ */ (0, P.jsxs)("button", {
									type: "button",
									"data-source-thread": e,
									onClick: () => d(e),
									className: "flex w-full items-center gap-1.5 text-left font-sans text-[11px] text-amber-600 hover:underline [[data-theme=dark]_&]:text-amber-400",
									children: [
										/* @__PURE__ */ (0, P.jsx)(K, {
											author: n.author ?? null,
											size: "xs",
											bare: !0
										}),
										/* @__PURE__ */ (0, P.jsx)("span", {
											className: "shrink-0 font-medium text-foreground",
											children: n.author?.name
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
							y.map((e) => /* @__PURE__ */ (0, P.jsxs)("span", {
								"data-source-finding": e.line,
								className: "flex items-center gap-2",
								children: [/* @__PURE__ */ (0, P.jsxs)("button", {
									type: "button",
									"aria-pressed": c === $(e.line),
									onClick: () => l(e),
									className: s("flex min-w-0 flex-1 items-center gap-1.5 text-left font-sans text-[11px] hover:underline aria-pressed:font-semibold aria-pressed:underline", vn[yn(e.severity)].tone),
									children: [
										/* @__PURE__ */ (0, P.jsx)(u, { icon: vn[yn(e.severity)].icon }),
										/* @__PURE__ */ (0, P.jsx)("span", {
											className: "text-foreground",
											children: e.message
										}),
										/* @__PURE__ */ (0, P.jsxs)("span", {
											className: "text-muted-foreground",
											children: [
												e.source,
												" ",
												e.rule
											]
										})
									]
								}), f && /* @__PURE__ */ (0, P.jsx)(Ce, { items: f(e) })]
							}, `${e.rule}:${e.message}`))
						]
					})
				]
			}, m);
		})
	});
}
var Bn = 24, Vn = 60;
function Hn(e) {
	return `s:${JSON.stringify([
		e.line,
		e.prefix,
		e.quote,
		e.suffix
	])}`;
}
function Un(e) {
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
function Wn(e) {
	let t = [0];
	for (let n = e.indexOf("\n"); n !== -1; n = e.indexOf("\n", n + 1)) t.push(n + 1);
	return t;
}
function Gn(e, t, n) {
	let r = t[n.line - 1];
	if (r === void 0) throw Error(`line ${n.line} is outside the ${t.length}-line source`);
	let i = t[n.line] === void 0 ? e.length : t[n.line] - 1;
	return Math.min(r + n.column, i);
}
function Kn(e, t) {
	let n = 1;
	for (; n < e.length && e[n] <= t;) n++;
	return n;
}
function qn(e, t, n) {
	let r = Wn(e), i = Gn(e, r, t), a = Gn(e, r, n), o = e.slice(i, a);
	return o.trim() ? {
		line: Kn(r, i),
		prefix: e.slice(Math.max(0, i - Bn), i),
		quote: o,
		suffix: e.slice(a, a + Bn)
	} : null;
}
function Jn(e, t) {
	let n = t.prefix + t.quote + t.suffix, r = Wn(e), i = null, a = Infinity;
	for (let o = e.indexOf(n); o !== -1; o = e.indexOf(n, o + 1)) {
		let e = o + t.prefix.length, n = Math.abs(Kn(r, e) - t.line);
		n < a && (i = e, a = n);
	}
	if (i === null) return [];
	let o = i + t.quote.length, s = [];
	for (let t = Kn(r, i); t <= r.length && r[t - 1] < o; t++) {
		let n = r[t - 1], a = r[t] === void 0 ? e.length : r[t] - 1, c = {
			line: t,
			start: Math.max(i, n) - n,
			end: Math.min(o, a) - n
		};
		c.end > c.start && s.push(c);
	}
	return s;
}
function Yn(e) {
	let t = e.quote.replace(/\s+/g, " ").trim();
	return `Line ${e.line} · “${t.length > Vn ? `${t.slice(0, Vn)}…` : t}”`;
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/InlineThread.tsx
var Xn = 340, Zn = 6;
function Qn({ anchor: e, label: t, boxes: n, containerWidth: r, thread: i, config: a, callbacks: o, pinned: c, onPin: l, onClose: d, onPointerEnter: f, onPointerLeave: p, seamless: m = !1, renderRootDetails: h }) {
	(0, N.useEffect)(() => {
		let e = (e) => e.key === "Escape" && d();
		return window.addEventListener("keydown", e), () => window.removeEventListener("keydown", e);
	}, [d]);
	let g = n[0], _ = n.at(-1);
	if (!g || !_) return null;
	let v = Math.min(Xn, r), y = Math.max(0, Math.min(g.left, r - v)), b = !!o.onCreate;
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		"data-inline-thread": e,
		role: "dialog",
		"aria-label": `Comments on ${t}`,
		onMouseEnter: f,
		onMouseLeave: p,
		onMouseDown: l,
		className: s("absolute z-30 bg-card", m ? "rounded-md font-sans shadow-[0_4px_16px_rgba(0,0,0,0.12)] ring-1 ring-border/60" : "rounded-lg border shadow-lg", !m && (c ? "border-primary/50" : "border-border")),
		style: {
			top: _.top + _.height + (m ? 2 : Zn),
			left: y,
			width: v
		},
		children: [!m && /* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: t
				}),
				!c && /* @__PURE__ */ (0, P.jsx)("span", { children: "click to pin" }),
				/* @__PURE__ */ (0, P.jsx)("button", {
					type: "button",
					"aria-label": "Close comments",
					onClick: d,
					className: "rounded p-0.5 hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
					children: /* @__PURE__ */ (0, P.jsx)(u, {
						icon: ie,
						className: "text-xs"
					})
				})
			]
		}), /* @__PURE__ */ (0, P.jsx)("div", {
			className: "max-h-[360px] overflow-y-auto",
			children: /* @__PURE__ */ (0, P.jsx)(xt, {
				comments: i,
				config: a,
				anchor: e,
				compact: !0,
				defaultExpanded: !0,
				hideComposer: !b || i.length > 0 && !c,
				autoFocusComposer: i.length === 0,
				composerPlaceholder: i.length > 0 ? "Add another comment…" : "Add a comment…",
				...o,
				...h ? { renderRootDetails: h } : {}
			})
		})]
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/InlineAuthors.tsx
function $n({ markdown: e, comments: t, focused: n, onFocus: r }) {
	let [i, a] = (0, N.useState)(/* @__PURE__ */ new Map());
	(0, N.useLayoutEffect)(() => {
		let n = e.current;
		if (!n) return;
		let r = /* @__PURE__ */ new Map(), i = () => {
			let e = !1;
			for (let i of V(t)) {
				let t = i.anchor;
				if (!t?.startsWith("q:") || r.get(i.id)?.isConnected) continue;
				let a = Zt(n, Vt(t));
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
	let o = H(t);
	return V(t).map((e) => {
		let t = i.get(e.id), a = e.anchor;
		if (!t || !a) return null;
		let c = o.get(e.id)?.length ?? 0;
		return (0, q.createPortal)(/* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			"aria-label": `Open comment by ${B(e.author ?? null)}`,
			"aria-pressed": n === a,
			title: e.body,
			className: s("inline-flex max-w-[min(40ch,100%)] items-center gap-1 rounded-full border border-border bg-muted py-px pl-px pr-2 text-xs font-normal text-muted-foreground ring-offset-background hover:text-foreground hover:ring-2 hover:ring-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary", n === a && "text-foreground ring-2 ring-primary"),
			onClick: () => r(a),
			children: [
				/* @__PURE__ */ (0, P.jsx)(K, {
					author: e.author ?? null,
					size: "xs",
					bare: !0
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "truncate",
					children: He(e.body)
				}),
				c > 0 && /* @__PURE__ */ (0, P.jsxs)("span", {
					className: "shrink-0 tabular-nums",
					children: ["· ", c]
				})
			]
		}), t, e.id);
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/CommentDraft.tsx
var er = {
	width: 104,
	height: 28
}, tr = 320, nr = 190, rr = 6;
function ir({ anchor: e, label: t, position: n, containerWidth: r, expanded: i, config: a, onCreate: o, onExpand: c, onClose: l }) {
	(0, N.useEffect)(() => {
		if (!i) return;
		let e = (e) => e.key === "Escape" && l();
		return window.addEventListener("keydown", e), () => window.removeEventListener("keydown", e);
	}, [i, l]);
	let d = n.top - er.height - 8, f = n.top >= nr, p = Math.min(tr, r), m = Math.max(p / 2, Math.min(n.left, r - p / 2)), h = i ? {
		top: f ? d + er.height : n.bottom + rr,
		left: m,
		width: p,
		transform: `translate(-50%, ${f ? "-100%" : "0"})`
	} : {
		top: d,
		left: n.left,
		width: er.width,
		transform: "translate(-50%, 0)"
	};
	return /* @__PURE__ */ (0, P.jsx)("div", {
		"data-comment-draft": e,
		role: i ? "dialog" : void 0,
		"aria-label": i ? `New comment on ${t}` : void 0,
		onMouseDown: (e) => !i && e.preventDefault(),
		className: s("absolute z-30 overflow-hidden border bg-card shadow-md transition-[top,left,width,transform,border-radius,box-shadow] duration-200 ease-out", i ? "rounded-lg border-primary/50 shadow-lg" : "rounded-full border-border"),
		style: h,
		children: i ? /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex items-center gap-1 border-b border-border px-2 py-1 text-[11px] text-muted-foreground",
			children: [
				/* @__PURE__ */ (0, P.jsx)(u, {
					icon: y,
					className: "text-xs"
				}),
				/* @__PURE__ */ (0, P.jsx)("span", {
					className: "min-w-0 flex-1 truncate",
					children: t
				}),
				/* @__PURE__ */ (0, P.jsx)("button", {
					type: "button",
					"aria-label": "Discard comment",
					onClick: l,
					className: "rounded p-0.5 hover:bg-muted hover:text-foreground [@media(pointer:coarse)]:min-h-11 [@media(pointer:coarse)]:min-w-11",
					children: /* @__PURE__ */ (0, P.jsx)(u, {
						icon: ie,
						className: "text-xs"
					})
				})
			]
		}), /* @__PURE__ */ (0, P.jsx)("div", {
			className: "p-2",
			children: /* @__PURE__ */ (0, P.jsx)(tt, {
				config: a,
				anchor: e,
				collapsible: !1,
				autoFocus: !0,
				placeholder: "Add a comment…",
				onCreate: o
			})
		})] }) : /* @__PURE__ */ (0, P.jsxs)("button", {
			type: "button",
			onClick: c,
			className: "inline-flex w-full items-center justify-center gap-1 px-2.5 text-xs font-medium hover:bg-muted",
			style: { height: er.height },
			children: [/* @__PURE__ */ (0, P.jsx)(u, {
				icon: y,
				className: "text-sm"
			}), "Comment"]
		})
	});
}
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/file-review/FileReview.tsx
var ar = { statuses: Ne }, or = 96, sr = 450, cr = 300, lr = 16, ur = [], dr = [], fr = {}, pr = $(1).replace(/\d+$/, ""), mr = (e) => e.startsWith(pr), hr = (e) => /\.mdx?$/i.test(e), gr = (e) => e.startsWith("q:") || e.startsWith("s:"), _r = (e) => ({
	anchor: Bt(e),
	label: sn(e)
});
function vr({ mode: e, active: t }) {
	return e === "block" ? s("border-l-[3px] border-amber-400", t ? "bg-amber-200/50 [[data-theme=dark]_&]:bg-amber-400/20" : "bg-amber-100/30 [[data-theme=dark]_&]:bg-amber-400/5") : s("rounded-sm", t ? "bg-amber-300 [[data-theme=dark]_&]:bg-amber-400/50" : "bg-amber-200/60 [[data-theme=dark]_&]:bg-amber-400/25");
}
function yr(e, t) {
	let n = e.startsWith("q:") ? sn(Vt(e)) : e.startsWith("s:") ? Yn(Un(e)) : e.replace("line:", "line ");
	return t ? n : `Text not found · ${n}`;
}
var br = /[\p{L}\p{N}'’-]/u;
function xr(e, t) {
	let n = document.caretRangeFromPoint?.(e, t), r = n?.startContainer;
	if (!n || !r || r.nodeType !== Node.TEXT_NODE) return null;
	let i = r.textContent ?? "", a = n.startOffset, o = n.startOffset;
	for (; a > 0 && br.test(i[a - 1]);) a--;
	for (; o < i.length && br.test(i[o]);) o++;
	if (a === o) return null;
	let s = document.createRange();
	return s.setStart(r, a), s.setEnd(r, o), s;
}
var Sr = [{
	id: "inline",
	label: "Inline comments"
}, {
	id: "margin",
	label: "Comment sidebar"
}], Cr = [{
	id: "line",
	label: "Align to line"
}, {
	id: "stack",
	label: "Stacked list"
}], wr = 48;
function Tr({ path: e, source: t, view: n, onViewChange: r, findings: i, problemFilters: a, onProblemFiltersChange: o, comments: l, mode: d = "text", selectToComment: f = !1, authorFilter: p = !1, layout: h = "margin", onLayoutChange: _, onSourceChange: y, blame: b, line: S, onLineChange: C, getFindingActions: w, toolbar: T, className: E }) {
	let D = i !== void 0, O = i ?? ur, k = hr(e);
	if (n === "problems" && !D) throw Error("FileReview: view \"problems\" needs `findings`");
	if (n === "problems" && (!a || !o)) throw Error("FileReview: view \"problems\" needs `problemFilters` and `onProblemFiltersChange`");
	if (n === "preview" && !k) throw Error(`FileReview: view "preview" needs a markdown file, got ${e}`);
	let [ne, re] = (0, N.useState)(!1), [ie, ae] = (0, N.useState)(!0), oe = ne && b !== void 0, ce = y !== void 0 && !oe, ue = l?.items ?? dr, de = l?.callbacks ?? fr, fe = l?.config ?? ar, A = (0, N.useMemo)(() => ({
		...fe,
		statuses: [...fe.statuses, ...Dt.statuses],
		facets: [...fe.facets ?? [], ...(Dt.facets ?? []).filter((e) => e.key !== "severity")]
	}), [fe]), he = f && !!de.onCreate, j = (0, N.useMemo)(() => e.endsWith(".mdx") ? It(t) : t, [e, t]), ge = (0, N.useMemo)(() => j.split("\n"), [j]), _e = (0, N.useMemo)(() => oe ? jn(b, ge.length) : null, [
		oe,
		b,
		ge.length
	]), ve = m(), ye = n === "source" && !ce || n === "problems", [be, xe] = (0, N.useState)(null);
	(0, N.useEffect)(() => {
		if (!ye) return;
		let t = !0, n = ve === "dark" ? "github-dark" : "github-light";
		return g(j, {
			lang: v(e),
			theme: n
		}).then((e) => t && xe({
			source: j,
			lines: e
		})), () => {
			t = !1;
		};
	}, [
		ye,
		j,
		e,
		ve
	]);
	let Se = be?.source === j ? be.lines : null, [M, Te] = (0, N.useState)("line"), [F, I] = (0, N.useState)(null), [Ee, De] = (0, N.useState)(null), [L, R] = (0, N.useState)(null), [ke, Ae] = (0, N.useState)("all"), [je, Me] = (0, N.useState)(void 0), [z, Ne] = (0, N.useState)(null), Pe = (0, N.useRef)(null), B = (0, N.useRef)(null), Fe = (0, N.useRef)(null), Ie = (0, N.useRef)(null), Le = (0, N.useRef)(null), H = (0, N.useRef)(null), U = gn(Ie), { heights: ze, refFor: Be } = mn(), Ve = (0, N.useMemo)(() => At(O), [O]), He = (0, N.useMemo)(() => [...Ve, ...ue], [Ve, ue]), Ue = (0, N.useMemo)(() => new Map(Ve.map((e, t) => [e.id, O[t]])), [Ve, O]), We = (0, N.useMemo)(() => {
		if (w) return (e) => {
			let t = Ue.get(e.id);
			return t ? /* @__PURE__ */ (0, P.jsx)(Ce, { items: w(t) }) : null;
		};
	}, [w, Ue]), G = (0, N.useMemo)(() => He.filter((e) => e.parentId || ke === "all" || (e.author?.kind ?? "user") === ke), [He, ke]), K = (0, N.useMemo)(() => Ge(G, A, "active"), [G, A]), Ke = V(G).filter((e) => W(A, e.status) !== "active").length, q = (0, N.useMemo)(() => [...new Set(V(K).map((e) => e.anchor ?? ""))].filter(Boolean), [K]), qe = L?.expanded ? L.anchor : null, Je = qe && !q.includes(qe) ? qe : null, Ye = (0, N.useMemo)(() => Je ? [...q, Je] : q, [q, Je]), Xe = (0, N.useMemo)(() => {
		let e = /* @__PURE__ */ new Map();
		for (let t of V(K)) {
			if (!t.anchor || !mr(t.anchor)) continue;
			let n = yn(t.status), r = e.get(t.anchor);
			(!r || _n[n] > _n[r]) && e.set(t.anchor, n);
		}
		return e;
	}, [K]), { boxes: J, height: Ze, width: Qe } = hn({
		origin: Le,
		markdown: H,
		anchors: Ye,
		mode: d,
		visible: n === "preview"
	}), $e = (0, N.useRef)(null), [et, tt] = (0, N.useState)(0), [nt, rt] = (0, N.useState)([]);
	(0, N.useLayoutEffect)(() => {
		if (n !== "source" || !$e.current) return;
		let e = H.current, t = () => rt(Ye.flatMap((t) => t.startsWith("s:") ? [{
			anchor: t,
			spans: Jn(j, Un(t))
		}] : e && t.startsWith("q:") ? [{
			anchor: t,
			spans: an(e, ge, Vt(t))
		}] : []).filter((e) => e.spans.length > 0));
		t();
		let r = new MutationObserver(t);
		e && r.observe(e, {
			childList: !0,
			subtree: !0,
			characterData: !0
		});
		let i = $e.current, a = new ResizeObserver(() => tt(i.offsetWidth));
		return a.observe(i), () => {
			r.disconnect(), a.disconnect();
		};
	}, [
		n,
		Ye,
		j,
		ge,
		ce
	]);
	let it = (0, N.useRef)(null), at = (0, N.useRef)(null), Y = F ?? Ee;
	(0, N.useEffect)(() => () => clearTimeout(it.current ?? void 0), []);
	let [X, ot] = (0, N.useState)(null);
	(0, N.useLayoutEffect)(() => {
		let e = $e.current;
		if (n !== "source" || !e || !Y || !gr(Y)) return ot(null);
		let t = e.getBoundingClientRect(), r = [...e.querySelectorAll("[data-source-comment]")].filter((e) => Pn(e).includes(Y)).flatMap((e) => [...e.getClientRects()]).map((e) => ({
			top: e.top - t.top,
			left: e.left - t.left,
			width: e.width,
			height: e.height
		}));
		ot(r.length > 0 ? {
			anchor: Y,
			boxes: r
		} : null);
	}, [
		n,
		Y,
		nt,
		et
	]);
	let st = z ? J.get(z)?.[0] : void 0;
	(0, N.useEffect)(() => {
		let e = (U ? Pe : B).current;
		if (n !== "preview" || !st || !e || !Le.current) return;
		let t = Le.current.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop + st.top;
		t >= e.scrollTop && t + st.height <= e.scrollTop + e.clientHeight || (e.scrollTop = t - wr), Ne(null);
	}, [
		n,
		st,
		U
	]), (0, N.useEffect)(() => {
		let e = Pe.current;
		if (n !== "source" || !z || !e) return;
		let t = mr(z) ? z.slice(pr.length) : null, r = t ? e.querySelector(`[data-source-line="${t}"]`) : null;
		if (r) {
			let t = r.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
			t >= e.scrollTop && t + r.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = t - wr);
		}
		Ne(null);
	}, [n, z]);
	let Z = h === "margin" && !U, ct = () => {
		Z && M === "line" && Fe.current && B.current && (Fe.current.scrollTop = B.current.scrollTop);
	};
	(0, N.useEffect)(ct, [
		F,
		M,
		Z
	]), (0, N.useEffect)(() => {
		let e = Fe.current;
		if (!Z || M !== "stack" || !F || !e) return;
		let t = [...e.querySelectorAll("[data-card-anchor]")].find((e) => e.dataset.cardAnchor === F);
		if (!t) return;
		let n = t.getBoundingClientRect().top - e.getBoundingClientRect().top + e.scrollTop;
		n >= e.scrollTop && n + t.offsetHeight <= e.scrollTop + e.clientHeight || (e.scrollTop = n - wr);
	}, [
		F,
		M,
		Z
	]);
	function lt(e) {
		I(e), Ne(e), mr(e) && C?.(Number(e.slice(pr.length)));
	}
	function ut(e) {
		let t = $(e.line);
		r(k ? "preview" : "source"), I(t), Ne(t), k || Me(e.line), C?.(e.line);
	}
	(0, N.useEffect)(() => {
		if (S === void 0) return;
		let e = $(S);
		I(e), Ne(e), Me(S);
	}, [S]);
	let dt = [...J.keys()], ft = fn({
		cards: q.map((e) => ({
			anchor: e,
			desired: J.get(e)?.[0]?.top ?? Ze,
			order: J.has(e) ? dt.indexOf(e) : dt.length,
			height: ze[e] ?? or
		})),
		focused: F
	}), pt = Math.max(Ze, ...q.map((e) => (ft.get(e) ?? 0) + (ze[e] ?? or))), mt = [...q].sort((e, t) => (ft.get(e) ?? 0) - (ft.get(t) ?? 0));
	function ht(e, t) {
		let n = e.currentTarget.getBoundingClientRect(), r = e.clientX - n.left, i = e.clientY - n.top, a = [...J].filter(([e, n]) => {
			if (e === Je) return !1;
			if (mr(e)) {
				let e = n[n.length - 1];
				return r <= lr && i >= n[0].top && i <= e.top + e.height || t === "click" && n.some((e) => i >= e.top && i <= e.top + e.height && r >= e.left && r <= e.left + e.width);
			}
			return n.some((e) => {
				if (i < e.top || i > e.top + e.height) return !1;
				let n = d === "text";
				return t === "click" || n ? r >= e.left - (n ? 0 : 12) && r <= e.left + e.width : r >= e.left - lr && r <= e.left;
			});
		}).map(([e]) => e);
		return a.find((e) => e.startsWith("q:")) ?? a[0] ?? null;
	}
	let Q = (e) => e.target.closest("[data-inline-thread], [data-comment-draft], [data-source-thread]") !== null;
	function gt(e, t) {
		if (!k) {
			let e = In(t), n = e && qn(j, e.start, e.end);
			return n ? {
				anchor: Hn(n),
				label: Yn(n)
			} : null;
		}
		if (!H.current) return null;
		if (e === "preview") {
			let e = Qt(H.current, t);
			return e && _r(e);
		}
		let n = Ln(t), r = n && rn(H.current, ge, n);
		return r && _r(r);
	}
	let _t = (e) => (e === "preview" ? Le : $e).current;
	function vt(e) {
		!window.getSelection()?.isCollapsed || Q(e) || e.target.closest("button, a") || I(ht(e, "click"));
	}
	function yt(e) {
		e !== at.current && (at.current = e, clearTimeout(it.current ?? void 0), it.current = setTimeout(() => De(e), e ? sr : cr));
	}
	function bt(e) {
		clearTimeout(it.current ?? void 0), at.current = e;
	}
	function St() {
		clearTimeout(it.current ?? void 0), at.current = null, I(null), De(null);
	}
	function Ct(e) {
		if (Q(e)) return;
		let t = e.buttons !== 0 || !window.getSelection()?.isCollapsed, n = e.target.closest("[data-source-comment]");
		yt(t || !n ? null : Pn(n)[0]);
	}
	function wt(e) {
		Q(e) || e.target.closest("[data-source-comment], button, a") || F && gr(F) && window.getSelection()?.isCollapsed && St();
	}
	function Tt(e) {
		h !== "inline" || Q(e) || yt(e.buttons !== 0 || !window.getSelection()?.isCollapsed ? null : ht(e, "hover"));
	}
	function Et(e, t) {
		if (!he || Q(e) || e.target.closest("[data-comment-author]")) return;
		let n = window.getSelection(), r = n && !n.isCollapsed && n.rangeCount > 0 ? n.getRangeAt(0) : null, i = r ? gt(t, r) : null;
		if (!r || !i) {
			R((e) => e?.expanded ? e : null);
			return;
		}
		let a = _t(t).getBoundingClientRect(), o = r.getBoundingClientRect();
		R({
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
		if (!he || Q(e) || e.target.closest("[data-comment-author], button, a")) return;
		let n = window.getSelection(), r = n && !n.isCollapsed && n.rangeCount > 0 ? n.getRangeAt(0) : null, i = r && [...r.getClientRects()].some((t) => e.clientX >= t.left && e.clientX <= t.right && e.clientY >= t.top && e.clientY <= t.bottom), a = i ? r : xr(e.clientX, e.clientY), o = a ? gt(t, a) : null;
		if (!a || !o) return;
		e.preventDefault(), i || (n?.removeAllRanges(), n?.addRange(a));
		let s = _t(t).getBoundingClientRect(), c = a.getBoundingClientRect();
		R({
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
		R((e) => e && {
			...e,
			expanded: !0
		}), window.getSelection()?.removeAllRanges();
	}
	let jt = (0, N.useCallback)(() => R(null), []);
	function Nt(e) {
		let t = e.expanded && e.surface === "preview" ? J.get(e.anchor) : void 0, n = t?.[0], r = t?.at(-1);
		return !n || !r ? e.position : {
			top: n.top,
			bottom: r.top + r.height,
			left: n.left + n.width / 2
		};
	}
	function Pt(e, t) {
		let n = de.onCreate;
		if (!n) return null;
		let { anchor: r, label: i } = e;
		return /* @__PURE__ */ (0, P.jsx)(ir, {
			anchor: r,
			label: i,
			position: Nt(e),
			containerWidth: t,
			expanded: e.expanded,
			config: A,
			onExpand: kt,
			onClose: jt,
			onCreate: async (e) => {
				await n(e), R(null), I(r);
			}
		}, r);
	}
	return /* @__PURE__ */ (0, P.jsxs)("div", {
		ref: Ie,
		"data-file-review": e,
		className: s("flex min-h-0 flex-col gap-2", E ?? "h-[600px]"),
		children: [
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex shrink-0 flex-wrap items-center gap-2 text-xs text-muted-foreground",
				children: [
					T,
					(l || D) && /* @__PURE__ */ (0, P.jsxs)("span", { children: [
						V(K).length,
						" open · ",
						Ke,
						" resolved (hidden)"
					] }),
					l && h === "inline" && /* @__PURE__ */ (0, P.jsx)("span", { children: "· hover a highlight to view or reply" }),
					he && /* @__PURE__ */ (0, P.jsx)("span", { children: "· select or right-click text to comment" }),
					p && /* @__PURE__ */ (0, P.jsx)("div", {
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
							"aria-checked": ke === e,
							onClick: () => Ae(e),
							className: s("rounded px-2 py-0.5 font-medium", ke === e ? "bg-primary text-primary-foreground" : "hover:text-foreground"),
							children: e === "all" ? "All" : e === "user" ? "People" : "Tools"
						}, e))
					})
				]
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				className: "flex shrink-0 items-center border-b border-border",
				children: [
					/* @__PURE__ */ (0, P.jsx)(we, {
						className: "min-w-0 flex-1 border-b-0",
						value: n,
						onChange: (e) => {
							r(e), R(null);
						},
						tabs: [
							...k ? [{
								id: "preview",
								label: "Preview",
								icon: pe
							}] : [],
							{
								id: "source",
								label: "Source",
								icon: se
							},
							...D ? [{
								id: "problems",
								label: "Problems",
								icon: le,
								count: O.length
							}] : []
						]
					}),
					D && O.length > 0 && n === "source" && !ce && /* @__PURE__ */ (0, P.jsxs)(c, {
						size: "sm",
						variant: ie ? "secondary" : "ghost",
						"aria-pressed": ie,
						title: ie ? "Show only the gutter marks; pick one to read its findings" : "Show each line's findings underneath it",
						onClick: () => ae((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, P.jsx)(u, { icon: le }), "Findings"]
					}),
					b && n === "source" && /* @__PURE__ */ (0, P.jsxs)(c, {
						size: "sm",
						variant: ne ? "secondary" : "ghost",
						"aria-pressed": ne,
						title: ne ? "Hide who last changed each line" : "Show who last changed each line",
						onClick: () => re((e) => !e),
						className: "shrink-0 gap-1.5",
						children: [/* @__PURE__ */ (0, P.jsx)(u, { icon: Oe }), "Blame"]
					}),
					l && _ && k && n !== "problems" && /* @__PURE__ */ (0, P.jsx)(ee, {
						align: "right",
						icon: x,
						hideChevron: !0,
						variant: "ghost",
						size: "sm",
						title: "View options",
						menuLabel: "View options",
						className: "shrink-0",
						items: [...Sr.map((e) => ({
							label: e.label,
							group: "Comments",
							icon: te,
							iconClassName: e.id === h ? void 0 : "invisible",
							onSelect: () => _(e.id)
						})), ...h === "margin" ? Cr.map((e) => ({
							label: e.label,
							group: "Sidebar cards",
							icon: te,
							iconClassName: e.id === M ? void 0 : "invisible",
							onSelect: () => Te(e.id)
						})) : []]
					})
				]
			}),
			/* @__PURE__ */ (0, P.jsxs)("div", {
				ref: Pe,
				className: s("relative min-h-0 flex-1 rounded-md border border-border bg-background", ce && n === "source" ? "overflow-hidden" : "overflow-y-auto"),
				children: [
					ce && n === "source" && /* @__PURE__ */ (0, P.jsx)(Mt, {
						path: e,
						source: t,
						findings: O,
						line: je,
						onLineChange: Me,
						onChange: y,
						inlineChips: !0
					}),
					!ce && n === "source" && /* @__PURE__ */ (0, P.jsxs)("div", {
						ref: $e,
						className: "relative",
						onMouseUp: (e) => e.button === 0 && Et(e, "source"),
						onContextMenu: (e) => Ot(e, "source"),
						onClick: wt,
						onMouseMove: Ct,
						onMouseLeave: () => yt(null),
						children: [
							/* @__PURE__ */ (0, P.jsx)(zn, {
								source: j,
								tokens: Se,
								blame: _e,
								findings: O,
								showFindings: ie,
								marks: nt,
								comments: K,
								selected: Y ?? qe,
								onSelect: (e) => {
									I($(e.line)), C?.(e.line);
								},
								getFindingActions: w,
								onFocus: I
							}),
							L?.surface === "source" && Pt(L, et),
							X && /* @__PURE__ */ (0, P.jsx)(Qn, {
								seamless: !0,
								anchor: X.anchor,
								label: yr(X.anchor, !0),
								boxes: X.boxes,
								containerWidth: et,
								thread: Re(K, X.anchor),
								config: A,
								callbacks: de,
								pinned: F === X.anchor,
								onPin: () => I(X.anchor),
								onClose: St,
								onPointerEnter: () => bt(X.anchor),
								onPointerLeave: () => yt(null)
							}, X.anchor)
						]
					}),
					n === "problems" && a && o && /* @__PURE__ */ (0, P.jsx)(kn, {
						source: j,
						tokens: Se,
						findings: O,
						filters: a,
						onFiltersChange: o,
						onOpen: ut,
						openLabel: k ? "Open in preview" : "Open in source",
						getFindingActions: w
					}),
					k && /* @__PURE__ */ (0, P.jsxs)("div", {
						className: s("flex", U ? "flex-col gap-4 p-4" : "h-full", n !== "preview" && "hidden"),
						children: [/* @__PURE__ */ (0, P.jsx)("div", {
							ref: B,
							onScroll: ct,
							className: s("min-w-0 flex-1", !U && "overflow-y-auto p-4"),
							children: /* @__PURE__ */ (0, P.jsxs)("div", {
								ref: Le,
								className: "relative",
								onMouseUp: (e) => e.button === 0 && Et(e, "preview"),
								onContextMenu: (e) => Ot(e, "preview"),
								onClick: vt,
								onMouseMove: Tt,
								onMouseLeave: () => h === "inline" && yt(null),
								children: [
									[...J].flatMap(([e, t]) => {
										let n = e === F || e === Ee || e === qe;
										if (mr(e)) {
											let r = Xe.get(e);
											if (!r) throw Error(`lint anchor ${e} has no open finding`);
											let i = vn[r], a = t[0];
											return [/* @__PURE__ */ (0, P.jsx)("span", {
												"data-lint-rail": r,
												"aria-hidden": !0,
												className: s("pointer-events-none absolute z-20 flex size-4 items-center justify-center rounded-sm", i.tone, n && i.tint),
												style: {
													top: a.top + a.height / 2 - 8,
													left: 0
												},
												children: /* @__PURE__ */ (0, P.jsx)(u, {
													icon: i.icon,
													className: "text-[13px]"
												})
											}, `${e}:rail`), ...t.map((t, a) => /* @__PURE__ */ (0, P.jsx)("div", {
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
												children: /* @__PURE__ */ (0, P.jsx)("span", {
													className: "absolute inset-x-0 bottom-0 h-[3px]",
													style: {
														backgroundImage: bn(i.stroke),
														backgroundRepeat: "repeat-x",
														backgroundSize: "6px 3px"
													}
												})
											}, `${e}:${a}`))];
										}
										return t.map((t, r) => /* @__PURE__ */ (0, P.jsx)("div", {
											"data-anchor-box": e,
											"aria-hidden": !0,
											className: s("pointer-events-none absolute z-0 transition-colors", vr({
												mode: d,
												active: n
											})),
											style: d === "text" ? {
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
										ref: H,
										"data-annotated-markdown": "",
										className: "relative z-10 pl-4",
										children: /* @__PURE__ */ (0, P.jsx)(me, {
											sourceBlocks: !0,
											text: j
										})
									}),
									l && h === "inline" && /* @__PURE__ */ (0, P.jsx)($n, {
										markdown: H,
										comments: K,
										focused: F,
										onFocus: I
									}),
									L?.surface === "preview" && Pt(L, Qe),
									h === "inline" && Y && /* @__PURE__ */ (0, P.jsx)(Qn, {
										anchor: Y,
										label: yr(Y, !0),
										boxes: J.get(Y) ?? [],
										containerWidth: Qe,
										thread: Re(K, Y),
										config: A,
										callbacks: mr(Y) ? fr : de,
										renderRootDetails: We,
										pinned: F === Y,
										onPin: () => I(Y),
										onClose: St,
										onPointerEnter: () => bt(Y),
										onPointerLeave: () => yt(null)
									}, Y)
								]
							})
						}), h === "margin" && /* @__PURE__ */ (0, P.jsx)("div", {
							ref: Fe,
							"data-comment-sidebar": U ? "below" : M,
							className: s(U ? "w-full" : "w-[316px] shrink-0 overflow-y-auto overscroll-contain border-l border-border py-4 pl-2 pr-4"),
							children: /* @__PURE__ */ (0, P.jsx)("div", {
								className: s("relative", (U || M === "stack") && "space-y-2"),
								style: Z && M === "line" ? { height: pt } : void 0,
								children: mt.map((e) => {
									let t = Re(K, e), n = e === F;
									return /* @__PURE__ */ (0, P.jsxs)("div", {
										ref: Be(e),
										"data-card-anchor": e,
										onMouseEnter: () => De(e),
										onMouseLeave: () => De(null),
										onClick: () => lt(e),
										className: s("rounded-lg border bg-card transition-[top,transform] duration-150", Z && M === "line" && "absolute inset-x-0", n ? "z-10 -translate-x-2 border-primary/50 shadow-lg" : "border-border shadow-sm"),
										style: Z && M === "line" ? { top: ft.get(e) ?? 0 } : void 0,
										children: [/* @__PURE__ */ (0, P.jsx)("div", {
											className: "truncate px-2 pt-2 text-[11px] text-muted-foreground",
											children: yr(e, J.has(e))
										}), /* @__PURE__ */ (0, P.jsx)(xt, {
											comments: t,
											config: A,
											anchor: e,
											compact: !0,
											defaultExpanded: n || mr(e),
											hideComposer: t.length > 0,
											autoFocusComposer: t.length === 0,
											composerPlaceholder: "Add a comment…",
											...mr(e) ? fr : de,
											...We ? { renderRootDetails: We } : {}
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
// ../../../../../../../../../Users/moshe/go/src/github.com/flanksource/clicky-ui/packages/ui/src/file-findings/FileFindingsView.tsx
function Er(e) {
	return /\.mdx?$/i.test(e);
}
function Dr({ findings: e }) {
	let t = wn(e);
	return /* @__PURE__ */ (0, P.jsx)("span", {
		className: "flex shrink-0 items-center gap-2 text-xs text-muted-foreground",
		children: Cn.filter((e) => t[e] > 0).map((e) => /* @__PURE__ */ (0, P.jsxs)("span", {
			className: "inline-flex items-center gap-0.5",
			title: `${t[e]} ${e}`,
			children: [/* @__PURE__ */ (0, P.jsx)(fe, { severity: e }), /* @__PURE__ */ (0, P.jsx)("span", {
				className: "tabular-nums",
				children: t[e]
			})]
		}, e))
	});
}
function Or({ node: e, selected: t }) {
	return /* @__PURE__ */ (0, P.jsxs)(P.Fragment, { children: [
		/* @__PURE__ */ (0, P.jsx)(u, {
			icon: e.kind === "folder" ? ue : se,
			className: "shrink-0 text-muted-foreground"
		}),
		/* @__PURE__ */ (0, P.jsx)("span", {
			className: s("min-w-0 truncate text-foreground", t ? "font-semibold" : "font-medium"),
			title: e.name,
			children: e.name
		}),
		/* @__PURE__ */ (0, P.jsx)("span", { className: "flex-1" }),
		/* @__PURE__ */ (0, P.jsx)(Dr, { findings: e.findings })
	] });
}
function kr(e, t) {
	if (!t) return null;
	for (let n of e) {
		if (n.kind === "file" && he(n.ref ?? null, t)) return n;
		let e = kr(n.children ?? [], t);
		if (e) return e;
	}
	return null;
}
function Ar(e, t) {
	let [n, r] = (0, N.useState)(null), i = e ? ve(e) : null, a = (0, N.useRef)({
		ref: e,
		loadFile: t
	});
	return a.current = {
		ref: e,
		loadFile: t
	}, (0, N.useEffect)(() => {
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
function jr({ children: e }) {
	return /* @__PURE__ */ (0, P.jsx)("div", {
		className: "flex h-full items-center justify-center p-8 text-sm text-muted-foreground",
		children: e
	});
}
function Mr({ file: e, findings: t, state: n, line: r, onLineChange: i, view: a, onViewChange: o, problemFilters: s, onProblemFiltersChange: c, getFindingActions: l }) {
	let d = !Er(e.file) && a === "preview" ? "source" : a, f;
	return f = !n || n.status === "loading" ? /* @__PURE__ */ (0, P.jsxs)(jr, { children: [
		/* @__PURE__ */ (0, P.jsx)(u, {
			icon: de,
			className: "mr-2 animate-spin"
		}),
		"Loading ",
		e.file
	] }) : n.status === "error" ? /* @__PURE__ */ (0, P.jsxs)("div", {
		role: "alert",
		className: "m-density-4 rounded-md border border-destructive/40 p-3 text-sm text-destructive",
		children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			className: "font-semibold",
			children: ["Could not load ", e.file]
		}), /* @__PURE__ */ (0, P.jsx)("pre", {
			className: "mt-1 whitespace-pre-wrap break-words font-mono text-xs",
			children: n.message
		})]
	}) : /* @__PURE__ */ (0, P.jsx)(Tr, {
		path: e.file,
		source: n.text,
		view: d,
		onViewChange: o,
		findings: t,
		problemFilters: s,
		onProblemFiltersChange: c,
		line: r,
		onLineChange: i,
		getFindingActions: l,
		className: "h-full p-density-2"
	}, ve(e)), /* @__PURE__ */ (0, P.jsxs)("div", {
		className: "flex h-full min-h-0 flex-col",
		children: [/* @__PURE__ */ (0, P.jsxs)("div", {
			className: "flex min-w-0 items-center gap-2 border-b border-border px-density-4 py-density-2",
			"data-file-header": "",
			children: [/* @__PURE__ */ (0, P.jsx)("span", {
				className: "min-w-0 truncate font-mono text-sm",
				title: e.workDir ? `${e.workDir}/${e.file}` : e.file,
				children: e.file
			}), /* @__PURE__ */ (0, P.jsx)(Dr, { findings: t })]
		}), /* @__PURE__ */ (0, P.jsx)("div", {
			className: "min-h-0 flex-1",
			children: f
		})]
	});
}
function Nr({ findings: e, selectedFile: t, onSelectFile: n, line: r, onLineChange: i, view: a, onViewChange: o, problemFilters: c, onProblemFiltersChange: l, loadFile: u, getFindingActions: d, defaultSplit: f = 28, className: p }) {
	let m = (0, N.useMemo)(() => Se(e), [e]), h = (0, N.useMemo)(() => kr(m, t), [m, t]), g = Ar(t, u), _ = /* @__PURE__ */ (0, P.jsx)(oe, {
		className: "min-h-0",
		roots: m,
		getKey: (e) => e.key,
		getChildren: (e) => e.children,
		selected: h,
		onSelect: (e) => {
			e.kind === "file" && e.ref && n(e.ref);
		},
		defaultOpen: () => !0,
		showSearch: !0,
		searchLabel: "Filter files",
		getSearchText: (e) => e.name,
		empty: /* @__PURE__ */ (0, P.jsx)("div", {
			className: "p-8 text-center text-sm text-muted-foreground",
			children: "No lint violations or failed tests with a file location"
		}),
		renderRow: ({ node: e, selected: t }) => /* @__PURE__ */ (0, P.jsx)(Or, {
			node: e,
			selected: t
		})
	}), v = t ? h ? /* @__PURE__ */ (0, P.jsx)(Mr, {
		file: t,
		findings: h.findings,
		state: g,
		line: r,
		onLineChange: i,
		view: a,
		onViewChange: o,
		problemFilters: c,
		onProblemFiltersChange: l,
		getFindingActions: d
	}) : /* @__PURE__ */ (0, P.jsxs)(jr, { children: [t.file, " has no findings in this run."] }) : /* @__PURE__ */ (0, P.jsx)(jr, { children: "Select a file to see its findings in context." });
	return /* @__PURE__ */ (0, P.jsx)("div", {
		className: s("flex h-full min-h-0 flex-col bg-background", p),
		"data-file-findings": "",
		children: /* @__PURE__ */ (0, P.jsx)(ce, {
			className: "min-h-0 flex-1",
			defaultSplit: f,
			minLeft: 15,
			minRight: 40,
			leftClass: "border-r-2 border-border bg-muted/30",
			left: _,
			right: v
		})
	});
}
// src/FilesTab.tsx
var Pr = /* @__PURE__ */ e({ default: () => Fr });
function Fr(e) {
	return /* @__PURE__ */ (0, P.jsx)(Nr, { ...e });
}
export { U as _, Tt as a, bt as c, H as d, Fr as default, We as f, Re as g, V as h, $ as i, W as l, Be as m, Dt as n, Et as o, ze as p, At as r, xt as s, Pr as t, Ge as u, z as v };
