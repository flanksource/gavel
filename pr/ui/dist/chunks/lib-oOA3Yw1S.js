import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { t as r } from "./react-dom-Cx3BtKbT.js";
import { t as i } from "./clsx-BitgEGBo.js";
import { t as a } from "./prop-types-PMJd4NkD.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/react-draggable/4.6.0/3327e58d4b110e6cb2ec1f9b9984ccc3dad5f3fc484a69508e9b4a050639dc60/node_modules/react-draggable/build/cjs/chunk-D5BXCJ5G.mjs
var o = /* @__PURE__ */ e(n(), 1), s = /* @__PURE__ */ e(a(), 1), c = /* @__PURE__ */ e(r(), 1);
function l(e, t) {
	for (let n = 0, r = e.length; n < r; n++) if (t.apply(t, [
		e[n],
		n,
		e
	])) return e[n];
}
function u(e) {
	return typeof e == "function" || Object.prototype.toString.call(e) === "[object Function]";
}
function d(e) {
	return typeof e == "number" && !isNaN(e);
}
function f(e) {
	return parseInt(e, 10);
}
function p(e, t, n) {
	if (e[t]) return /* @__PURE__ */ Error(`Invalid prop ${t} passed to ${n} - do not set this, set it on the child.`);
}
var m = [
	"Moz",
	"Webkit",
	"O",
	"ms"
];
function h(e = "transform") {
	if (typeof window > "u") return "";
	let t = window.document?.documentElement?.style;
	if (!t || e in t) return "";
	for (let n = 0; n < m.length; n++) if (g(e, m[n]) in t) return m[n];
	return "";
}
function g(e, t) {
	return t ? `${t}${_(e)}` : e;
}
function _(e) {
	let t = "", n = !0;
	for (let r = 0; r < e.length; r++) n ? (t += e[r].toUpperCase(), n = !1) : e[r] === "-" ? n = !0 : t += e[r];
	return t;
}
var v = h(), y = "";
function b(e, t) {
	y ||= l([
		"matches",
		"webkitMatchesSelector",
		"mozMatchesSelector",
		"msMatchesSelector",
		"oMatchesSelector"
	], function(t) {
		return u(e[t]);
	}) ?? "";
	let n = e[y];
	return u(n) ? !!n.call(e, t) : !1;
}
function x(e, t, n) {
	let r = e;
	do {
		if (b(r, t)) return !0;
		if (r === n) return !1;
		r = r.parentNode;
	} while (r);
	return !1;
}
function S(e, t, n, r) {
	if (!e) return;
	let i = {
		capture: !0,
		...r
	}, a = n;
	e.addEventListener ? e.addEventListener(t, a, i) : e.attachEvent ? e.attachEvent("on" + t, a) : e["on" + t] = a;
}
function C(e, t, n, r) {
	if (!e) return;
	let i = {
		capture: !0,
		...r
	}, a = n;
	e.removeEventListener ? e.removeEventListener(t, a, i) : e.detachEvent ? e.detachEvent("on" + t, a) : e["on" + t] = null;
}
function w(e) {
	let t = e.clientHeight, n = e.ownerDocument.defaultView.getComputedStyle(e);
	return t += f(n.borderTopWidth), t += f(n.borderBottomWidth), t;
}
function T(e) {
	let t = e.clientWidth, n = e.ownerDocument.defaultView.getComputedStyle(e);
	return t += f(n.borderLeftWidth), t += f(n.borderRightWidth), t;
}
function E(e) {
	let t = e.clientHeight, n = e.ownerDocument.defaultView.getComputedStyle(e);
	return t -= f(n.paddingTop), t -= f(n.paddingBottom), t;
}
function D(e) {
	let t = e.clientWidth, n = e.ownerDocument.defaultView.getComputedStyle(e);
	return t -= f(n.paddingLeft), t -= f(n.paddingRight), t;
}
function O(e, t, n) {
	let r = t === t.ownerDocument.body ? {
		left: 0,
		top: 0
	} : t.getBoundingClientRect();
	return {
		x: (e.clientX + t.scrollLeft - r.left) / n,
		y: (e.clientY + t.scrollTop - r.top) / n
	};
}
function k(e, t) {
	let n = j(e, t, "px");
	return { [g("transform", v)]: n };
}
function A(e, t) {
	return j(e, t, "");
}
function j({ x: e, y: t }, n, r) {
	let i = `translate(${e}${r},${t}${r})`;
	return n && (i = `translate(${`${typeof n.x == "string" ? n.x : n.x + r}`}, ${`${typeof n.y == "string" ? n.y : n.y + r}`})` + i), i;
}
function M(e, t) {
	return e.targetTouches && l(e.targetTouches, (e) => t === e.identifier) || e.changedTouches && l(e.changedTouches, (e) => t === e.identifier);
}
function ee(e) {
	if (e.targetTouches && e.targetTouches[0]) return e.targetTouches[0].identifier;
	if (e.changedTouches && e.changedTouches[0]) return e.changedTouches[0].identifier;
}
function N(e) {
	if (!e) return;
	let t = e.getElementById("react-draggable-style-el");
	t || (t = e.createElement("style"), t.type = "text/css", t.id = "react-draggable-style-el", t.innerHTML = ".react-draggable-transparent-selection *::-moz-selection {all: inherit;}\n", t.innerHTML += ".react-draggable-transparent-selection *::selection {all: inherit;}\n", e.getElementsByTagName("head")[0].appendChild(t)), e.body && ne(e.body, "react-draggable-transparent-selection");
}
function P(e) {
	window.requestAnimationFrame ? window.requestAnimationFrame(() => {
		te(e);
	}) : te(e);
}
function te(e) {
	if (e) try {
		e.body && re(e.body, "react-draggable-transparent-selection");
		let t = e.selection;
		if (t) t.empty();
		else {
			let t = (e.defaultView || window).getSelection();
			t && t.type !== "Caret" && t.removeAllRanges();
		}
	} catch {}
}
function ne(e, t) {
	e.classList ? e.classList.add(t) : e.className.match(RegExp(`(?:^|\\s)${t}(?!\\S)`)) || (e.className += ` ${t}`);
}
function re(e, t) {
	e.classList ? e.classList.remove(t) : e.className = e.className.replace(RegExp(`(?:^|\\s)${t}(?!\\S)`, "g"), "");
}
function ie(e, t, n) {
	if (!e.props.bounds) return [t, n];
	let { bounds: r } = e.props;
	r = typeof r == "string" ? r : ce(r);
	let i = R(e);
	if (typeof r == "string") {
		let { ownerDocument: e } = i, t = e.defaultView;
		if (!t) throw Error("Cannot resolve the owner window of the draggable node.");
		let n;
		if (n = r === "parent" ? i.parentNode : i.getRootNode().querySelector(r), !(n instanceof t.HTMLElement)) throw Error("Bounds selector \"" + r + "\" could not find an element.");
		let a = n, o = t.getComputedStyle(i), s = t.getComputedStyle(a);
		r = {
			left: -i.offsetLeft + f(s.paddingLeft) + f(o.marginLeft),
			top: -i.offsetTop + f(s.paddingTop) + f(o.marginTop),
			right: D(a) - T(i) - i.offsetLeft + f(s.paddingRight) - f(o.marginRight),
			bottom: E(a) - w(i) - i.offsetTop + f(s.paddingBottom) - f(o.marginBottom)
		};
	}
	return d(r.right) && (t = Math.min(t, r.right)), d(r.bottom) && (n = Math.min(n, r.bottom)), d(r.left) && (t = Math.max(t, r.left)), d(r.top) && (n = Math.max(n, r.top)), [t, n];
}
function ae(e, t, n) {
	return [Math.round(t / e[0]) * e[0], Math.round(n / e[1]) * e[1]];
}
function oe(e) {
	return e.props.axis === "both" || e.props.axis === "x";
}
function se(e) {
	return e.props.axis === "both" || e.props.axis === "y";
}
function F(e, t, n) {
	let r = typeof t == "number" ? M(e, t) : null;
	if (typeof t == "number" && !r) return null;
	let i = R(n), a = n.props.offsetParent || i.offsetParent || i.ownerDocument.body;
	return O(r || e, a, n.props.scale);
}
function I(e, t, n) {
	let r = !d(e.lastX), i = R(e);
	return r ? {
		node: i,
		deltaX: 0,
		deltaY: 0,
		lastX: t,
		lastY: n,
		x: t,
		y: n
	} : {
		node: i,
		deltaX: t - e.lastX,
		deltaY: n - e.lastY,
		lastX: e.lastX,
		lastY: e.lastY,
		x: t,
		y: n
	};
}
function L(e, t) {
	let n = e.props.scale;
	return {
		node: t.node,
		x: e.state.x + t.deltaX / n,
		y: e.state.y + t.deltaY / n,
		deltaX: t.deltaX / n,
		deltaY: t.deltaY / n,
		lastX: e.state.x,
		lastY: e.state.y
	};
}
function ce(e) {
	return {
		left: e.left,
		top: e.top,
		right: e.right,
		bottom: e.bottom
	};
}
function R(e) {
	let t = e.findDOMNode();
	if (!t) throw Error("<DraggableCore>: Unmounted during event!");
	return t;
}
function z(...e) {
	process.env.DRAGGABLE_DEBUG && console.log(...e);
}
var B = {
	touch: {
		start: "touchstart",
		move: "touchmove",
		stop: "touchend"
	},
	mouse: {
		start: "mousedown",
		move: "mousemove",
		stop: "mouseup"
	}
}, V = B.mouse, H = class extends o.Component {
	constructor() {
		super(...arguments), this.dragging = !1, this.lastX = NaN, this.lastY = NaN, this.touchIdentifier = null, this.mounted = !1, this.handleDragStart = (e) => {
			if (this.props.onMouseDown(e), !this.props.allowAnyClick && (typeof e.button == "number" && e.button !== 0 || e.ctrlKey)) return !1;
			let t = this.findDOMNode();
			if (!t || !t.ownerDocument || !t.ownerDocument.body) throw Error("<DraggableCore> not mounted on DragStart!");
			let { ownerDocument: n } = t;
			if (this.props.disabled || !(e.target instanceof n.defaultView.Node) || this.props.handle && !x(e.target, this.props.handle, t) || this.props.cancel && x(e.target, this.props.cancel, t)) return;
			e.type === "touchstart" && !this.props.allowMobileScroll && e.preventDefault();
			let r = ee(e);
			this.touchIdentifier = r;
			let i = F(e, r, this);
			if (i == null) return;
			let { x: a, y: o } = i, s = I(this, a, o);
			z("DraggableCore: handleDragStart: %j", s), z("calling", this.props.onStart), this.props.onStart(e, s) !== !1 && this.mounted !== !1 && (this.props.enableUserSelectHack && N(n), this.dragging = !0, this.lastX = a, this.lastY = o, S(n, V.move, this.handleDrag), S(n, V.stop, this.handleDragStop));
		}, this.handleDrag = (e) => {
			let t = F(e, this.touchIdentifier, this);
			if (t == null) return;
			let { x: n, y: r } = t;
			if (Array.isArray(this.props.grid)) {
				let e = n - this.lastX, t = r - this.lastY;
				if ([e, t] = ae(this.props.grid, e, t), !e && !t) return;
				n = this.lastX + e, r = this.lastY + t;
			}
			let i = I(this, n, r);
			if (z("DraggableCore: handleDrag: %j", i), this.props.onDrag(e, i) === !1 || this.mounted === !1) try {
				this.handleDragStop(new MouseEvent("mouseup"));
			} catch {
				let e = document.createEvent("MouseEvents");
				e.initMouseEvent("mouseup", !0, !0, window, 0, 0, 0, 0, 0, !1, !1, !1, !1, 0, null), this.handleDragStop(e);
			}
			else this.lastX = n, this.lastY = r;
		}, this.handleDragStop = (e) => {
			if (!this.dragging) return;
			let t = F(e, this.touchIdentifier, this);
			if (t == null) return;
			let { x: n, y: r } = t;
			if (Array.isArray(this.props.grid)) {
				let e = n - this.lastX || 0, t = r - this.lastY || 0;
				[e, t] = ae(this.props.grid, e, t), n = this.lastX + e, r = this.lastY + t;
			}
			let i = I(this, n, r);
			if (this.props.onStop(e, i) === !1 || this.mounted === !1) return !1;
			let a = this.findDOMNode();
			a && this.props.enableUserSelectHack && P(a.ownerDocument), z("DraggableCore: handleDragStop: %j", i), this.dragging = !1, this.lastX = NaN, this.lastY = NaN, a && (z("DraggableCore: Removing handlers"), C(a.ownerDocument, V.move, this.handleDrag), C(a.ownerDocument, V.stop, this.handleDragStop));
		}, this.onMouseDown = (e) => (V = B.mouse, this.handleDragStart(e)), this.onMouseUp = (e) => (V = B.mouse, this.handleDragStop(e)), this.onTouchStart = (e) => (V = B.touch, this.handleDragStart(e)), this.onTouchEnd = (e) => (V = B.touch, this.handleDragStop(e));
	}
	componentDidMount() {
		this.mounted = !0;
		let e = this.findDOMNode();
		e && S(e, B.touch.start, this.onTouchStart, { passive: !1 });
	}
	componentWillUnmount() {
		this.mounted = !1;
		let e = this.findDOMNode();
		if (e) {
			let { ownerDocument: t } = e;
			C(t, B.mouse.move, this.handleDrag), C(t, B.touch.move, this.handleDrag), C(t, B.mouse.stop, this.handleDragStop), C(t, B.touch.stop, this.handleDragStop), C(e, B.touch.start, this.onTouchStart, { passive: !1 }), this.props.enableUserSelectHack && P(t);
		}
	}
	findDOMNode() {
		if (this.props?.nodeRef) return this.props.nodeRef.current;
		let e = c.default;
		return typeof e.findDOMNode == "function" ? e.findDOMNode(this) : (z("react-draggable: ReactDOM.findDOMNode is not available in React 19+. You must provide a nodeRef prop. See: https://github.com/react-grid-layout/react-draggable#noderef"), null);
	}
	render() {
		return o.cloneElement(o.Children.only(this.props.children), {
			onMouseDown: this.onMouseDown,
			onMouseUp: this.onMouseUp,
			onTouchEnd: this.onTouchEnd
		});
	}
};
H.displayName = "DraggableCore", H.propTypes = {
	allowAnyClick: s.default.bool,
	allowMobileScroll: s.default.bool,
	children: s.default.node.isRequired,
	disabled: s.default.bool,
	enableUserSelectHack: s.default.bool,
	offsetParent: function(e, t) {
		if (e[t] && e[t].nodeType !== 1) throw Error("Draggable's offsetParent must be a DOM Node.");
	},
	grid: s.default.arrayOf(s.default.number),
	handle: s.default.string,
	cancel: s.default.string,
	nodeRef: s.default.object,
	onStart: s.default.func,
	onDrag: s.default.func,
	onStop: s.default.func,
	onMouseDown: s.default.func,
	scale: s.default.number,
	className: p,
	style: p,
	transform: p
}, H.defaultProps = {
	allowAnyClick: !1,
	allowMobileScroll: !1,
	disabled: !1,
	enableUserSelectHack: !0,
	onStart: function() {},
	onDrag: function() {},
	onStop: function() {},
	onMouseDown: function() {},
	scale: 1
};
var U = class extends o.Component {
	constructor(e) {
		super(e), this.onDragStart = (e, t) => {
			if (z("Draggable: onDragStart: %j", t), this.props.onStart(e, L(this, t)) === !1) return !1;
			this.setState({
				dragging: !0,
				dragged: !0
			});
		}, this.onDrag = (e, t) => {
			if (!this.state.dragging) return !1;
			z("Draggable: onDrag: %j", t);
			let n = L(this, t), r = {
				x: n.x,
				y: n.y,
				slackX: 0,
				slackY: 0
			};
			if (this.props.bounds) {
				let { x: e, y: t } = r;
				r.x += this.state.slackX, r.y += this.state.slackY;
				let [i, a] = ie(this, r.x, r.y);
				r.x = i, r.y = a, r.slackX = this.state.slackX + (e - r.x), r.slackY = this.state.slackY + (t - r.y), n.x = r.x, n.y = r.y, n.deltaX = r.x - this.state.x, n.deltaY = r.y - this.state.y;
			}
			if (this.props.onDrag(e, n) === !1) return !1;
			this.setState(r);
		}, this.onDragStop = (e, t) => {
			if (!this.state.dragging || this.props.onStop(e, L(this, t)) === !1) return !1;
			z("Draggable: onDragStop: %j", t);
			let n = {
				dragging: !1,
				slackX: 0,
				slackY: 0
			};
			if (this.props.position) {
				let { x: e, y: t } = this.props.position;
				n.x = e, n.y = t;
			}
			this.setState(n);
		}, this.state = {
			dragging: !1,
			dragged: !1,
			x: e.position ? e.position.x : e.defaultPosition.x,
			y: e.position ? e.position.y : e.defaultPosition.y,
			prevPropsPosition: { ...e.position },
			slackX: 0,
			slackY: 0,
			isElementSVG: !1
		}, e.position && !(e.onDrag || e.onStop) && console.warn("A `position` was applied to this <Draggable>, without drag handlers. This will make this component effectively undraggable. Please attach `onDrag` or `onStop` handlers so you can adjust the `position` of this element.");
	}
	static getDerivedStateFromProps({ position: e }, { prevPropsPosition: t }) {
		return e && (!t || e.x !== t.x || e.y !== t.y) ? (z("Draggable: getDerivedStateFromProps %j", {
			position: e,
			prevPropsPosition: t
		}), {
			x: e.x,
			y: e.y,
			prevPropsPosition: { ...e }
		}) : null;
	}
	componentDidMount() {
		window.SVGElement !== void 0 && this.findDOMNode() instanceof window.SVGElement && this.setState({ isElementSVG: !0 });
	}
	componentWillUnmount() {
		this.state.dragging && this.setState({ dragging: !1 });
	}
	findDOMNode() {
		if (this.props?.nodeRef) return this.props.nodeRef.current;
		let e = c.default;
		return typeof e.findDOMNode == "function" ? e.findDOMNode(this) : null;
	}
	render() {
		let { axis: e, bounds: t, children: n, defaultPosition: r, defaultClassName: a, defaultClassNameDragging: s, defaultClassNameDragged: c, position: l, positionOffset: u, scale: d, ...f } = this.props, p = {}, m = null, h = !l || this.state.dragging, g = l || r, _ = {
			x: oe(this) && h ? this.state.x : g.x,
			y: se(this) && h ? this.state.y : g.y
		};
		this.state.isElementSVG ? m = A(_, u) : p = k(_, u);
		let v = o.Children.only(n), y = i(v.props.className || "", a, {
			[s]: this.state.dragging,
			[c]: this.state.dragged
		});
		return /* @__PURE__ */ o.createElement(H, {
			...f,
			onStart: this.onDragStart,
			onDrag: this.onDrag,
			onStop: this.onDragStop
		}, o.cloneElement(v, {
			className: y,
			style: {
				...v.props.style,
				...p
			},
			transform: m
		}));
	}
};
U.displayName = "Draggable", U.propTypes = {
	...H.propTypes,
	axis: s.default.oneOf([
		"both",
		"x",
		"y",
		"none"
	]),
	bounds: s.default.oneOfType([
		s.default.shape({
			left: s.default.number,
			right: s.default.number,
			top: s.default.number,
			bottom: s.default.number
		}),
		s.default.string,
		s.default.oneOf([!1])
	]),
	defaultClassName: s.default.string,
	defaultClassNameDragging: s.default.string,
	defaultClassNameDragged: s.default.string,
	defaultPosition: s.default.shape({
		x: s.default.number,
		y: s.default.number
	}),
	positionOffset: s.default.shape({
		x: s.default.oneOfType([s.default.number, s.default.string]),
		y: s.default.oneOfType([s.default.number, s.default.string])
	}),
	position: s.default.shape({
		x: s.default.number,
		y: s.default.number
	}),
	className: p,
	style: p,
	transform: p
}, U.defaultProps = {
	...H.defaultProps,
	axis: "both",
	bounds: !1,
	defaultClassName: "react-draggable",
	defaultClassNameDragging: "react-draggable-dragging",
	defaultClassNameDragged: "react-draggable-dragged",
	defaultPosition: {
		x: 0,
		y: 0
	},
	scale: 1
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/react-draggable/4.6.0/3327e58d4b110e6cb2ec1f9b9984ccc3dad5f3fc484a69508e9b4a050639dc60/node_modules/react-draggable/build/cjs/cjs.mjs
var le = U, W = /* @__PURE__ */ e(t()), G = function() {
	return G = Object.assign || function(e) {
		for (var t, n = 1, r = arguments.length; n < r; n++) for (var i in t = arguments[n], t) Object.prototype.hasOwnProperty.call(t, i) && (e[i] = t[i]);
		return e;
	}, G.apply(this, arguments);
}, ue = {
	width: "100%",
	height: "10px",
	top: "0px",
	left: "0px",
	cursor: "row-resize"
}, de = {
	width: "10px",
	height: "100%",
	top: "0px",
	left: "0px",
	cursor: "col-resize"
}, K = {
	width: "20px",
	height: "20px",
	position: "absolute",
	zIndex: 1
}, fe = {
	top: G(G({}, ue), { top: "-5px" }),
	right: G(G({}, de), {
		left: void 0,
		right: "-5px"
	}),
	bottom: G(G({}, ue), {
		top: void 0,
		bottom: "-5px"
	}),
	left: G(G({}, de), { left: "-5px" }),
	topRight: G(G({}, K), {
		right: "-10px",
		top: "-10px",
		cursor: "ne-resize"
	}),
	bottomRight: G(G({}, K), {
		right: "-10px",
		bottom: "-10px",
		cursor: "se-resize"
	}),
	bottomLeft: G(G({}, K), {
		left: "-10px",
		bottom: "-10px",
		cursor: "sw-resize"
	}),
	topLeft: G(G({}, K), {
		left: "-10px",
		top: "-10px",
		cursor: "nw-resize"
	})
}, pe = (0, o.memo)(function(e) {
	var t = e.onResizeStart, n = e.direction, r = e.children, i = e.replaceStyles, a = e.className, s = (0, o.useCallback)(function(e) {
		t(e, n);
	}, [t, n]), c = (0, o.useCallback)(function(e) {
		t(e, n);
	}, [t, n]), l = (0, o.useMemo)(function() {
		return G(G({
			position: "absolute",
			userSelect: "none"
		}, fe[n]), i ?? {});
	}, [i, n]);
	return (0, W.jsx)("div", {
		className: a || void 0,
		style: l,
		onMouseDown: s,
		onTouchStart: c,
		children: r
	});
}), me = (function() {
	var e = function(t, n) {
		return e = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(e, t) {
			e.__proto__ = t;
		} || function(e, t) {
			for (var n in t) Object.prototype.hasOwnProperty.call(t, n) && (e[n] = t[n]);
		}, e(t, n);
	};
	return function(t, n) {
		if (typeof n != "function" && n !== null) throw TypeError("Class extends value " + String(n) + " is not a constructor or null");
		e(t, n);
		function r() {
			this.constructor = t;
		}
		t.prototype = n === null ? Object.create(n) : (r.prototype = n.prototype, new r());
	};
})(), q = function() {
	return q = Object.assign || function(e) {
		for (var t, n = 1, r = arguments.length; n < r; n++) for (var i in t = arguments[n], t) Object.prototype.hasOwnProperty.call(t, i) && (e[i] = t[i]);
		return e;
	}, q.apply(this, arguments);
}, he = {
	width: "auto",
	height: "auto"
}, J = function(e, t, n) {
	return Math.max(Math.min(e, n), t);
}, ge = function(e, t, n) {
	var r = Math.round(e / t);
	return r * t + n * (r - 1);
}, Y = function(e, t) {
	return new RegExp(e, "i").test(t);
}, X = function(e) {
	return !!(e.touches && e.touches.length);
}, _e = function(e) {
	return !(!e.clientX && e.clientX !== 0 || !e.clientY && e.clientY !== 0);
}, ve = function(e, t, n) {
	n === void 0 && (n = 0);
	var r = t.reduce(function(n, r, i) {
		return Math.abs(r - e) < Math.abs(t[n] - e) ? i : n;
	}, 0), i = Math.abs(t[r] - e);
	return n === 0 || i < n ? t[r] : e;
}, Z = function(e) {
	return e = e.toString(), e === "auto" || e.endsWith("px") || e.endsWith("%") || e.endsWith("vh") || e.endsWith("vw") || e.endsWith("vmax") || e.endsWith("vmin") ? e : `${e}px`;
}, Q = function(e, t, n, r) {
	if (e && typeof e == "string") {
		if (e.endsWith("px")) return Number(e.replace("px", ""));
		if (e.endsWith("%")) {
			var i = Number(e.replace("%", "")) / 100;
			return t * i;
		}
		if (e.endsWith("vw")) {
			var i = Number(e.replace("vw", "")) / 100;
			return n * i;
		}
		if (e.endsWith("vh")) {
			var i = Number(e.replace("vh", "")) / 100;
			return r * i;
		}
	}
	return e;
}, ye = function(e, t, n, r, i, a, o) {
	return r = Q(r, e.width, t, n), i = Q(i, e.height, t, n), a = Q(a, e.width, t, n), o = Q(o, e.height, t, n), {
		maxWidth: r === void 0 ? void 0 : Number(r),
		maxHeight: i === void 0 ? void 0 : Number(i),
		minWidth: a === void 0 ? void 0 : Number(a),
		minHeight: o === void 0 ? void 0 : Number(o)
	};
}, be = function(e) {
	return Array.isArray(e) ? e : [e, e];
}, xe = /* @__PURE__ */ "as.ref.style.className.grid.gridGap.snap.bounds.boundsByDirection.size.defaultSize.minWidth.minHeight.maxWidth.maxHeight.lockAspectRatio.lockAspectRatioExtraWidth.lockAspectRatioExtraHeight.enable.handleStyles.handleClasses.handleWrapperStyle.handleWrapperClass.children.onResizeStart.onResize.onResizeStop.handleComponent.scale.resizeRatio.snapGap".split("."), Se = "__resizable_base__", Ce = function(e) {
	me(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.ratio = 1, n.resizable = null, n.parentLeft = 0, n.parentTop = 0, n.resizableLeft = 0, n.resizableRight = 0, n.resizableTop = 0, n.resizableBottom = 0, n.targetLeft = 0, n.targetTop = 0, n.delta = {
			width: 0,
			height: 0
		}, n.appendBase = function() {
			if (!n.resizable || !n.window) return null;
			var e = n.parentNode;
			if (!e) return null;
			var t = n.window.document.createElement("div");
			return t.style.width = "100%", t.style.height = "100%", t.style.position = "absolute", t.style.transform = "scale(0, 0)", t.style.left = "0", t.style.flex = "0 0 100%", t.classList ? t.classList.add(Se) : t.className += Se, e.appendChild(t), t;
		}, n.removeBase = function(e) {
			var t = n.parentNode;
			t && t.removeChild(e);
		}, n.state = {
			isResizing: !1,
			width: n.propsSize?.width ?? "auto",
			height: n.propsSize?.height ?? "auto",
			direction: "right",
			original: {
				x: 0,
				y: 0,
				width: 0,
				height: 0
			},
			backgroundStyle: {
				height: "100%",
				width: "100%",
				backgroundColor: "rgba(0,0,0,0)",
				cursor: "auto",
				opacity: 0,
				position: "fixed",
				zIndex: 9999,
				top: "0",
				left: "0",
				bottom: "0",
				right: "0"
			},
			flexBasis: void 0
		}, n.onResizeStart = n.onResizeStart.bind(n), n.onMouseMove = n.onMouseMove.bind(n), n.onMouseUp = n.onMouseUp.bind(n), n;
	}
	return Object.defineProperty(t.prototype, "parentNode", {
		get: function() {
			return this.resizable ? this.resizable.parentNode : null;
		},
		enumerable: !1,
		configurable: !0
	}), Object.defineProperty(t.prototype, "window", {
		get: function() {
			return !this.resizable || !this.resizable.ownerDocument ? null : this.resizable.ownerDocument.defaultView;
		},
		enumerable: !1,
		configurable: !0
	}), Object.defineProperty(t.prototype, "propsSize", {
		get: function() {
			return this.props.size || this.props.defaultSize || he;
		},
		enumerable: !1,
		configurable: !0
	}), Object.defineProperty(t.prototype, "size", {
		get: function() {
			var e = 0, t = 0;
			if (this.resizable && this.window) {
				var n = this.resizable.offsetWidth, r = this.resizable.offsetHeight, i = this.resizable.style.position;
				i !== "relative" && (this.resizable.style.position = "relative"), e = this.resizable.style.width === "auto" ? n : this.resizable.offsetWidth, t = this.resizable.style.height === "auto" ? r : this.resizable.offsetHeight, this.resizable.style.position = i;
			}
			return {
				width: e,
				height: t
			};
		},
		enumerable: !1,
		configurable: !0
	}), Object.defineProperty(t.prototype, "sizeStyle", {
		get: function() {
			var e = this, t = this.props.size, n = function(t) {
				if (e.state[t] === void 0 || e.state[t] === "auto") return "auto";
				if (e.propsSize && e.propsSize[t] && e.propsSize[t]?.toString().endsWith("%")) {
					if (e.state[t].toString().endsWith("%")) return e.state[t].toString();
					var n = e.getParentSize();
					return `${Number(e.state[t].toString().replace("px", "")) / n[t] * 100}%`;
				}
				return Z(e.state[t]);
			};
			return {
				width: t && t.width !== void 0 && !this.state.isResizing ? Z(t.width) : n("width"),
				height: t && t.height !== void 0 && !this.state.isResizing ? Z(t.height) : n("height")
			};
		},
		enumerable: !1,
		configurable: !0
	}), t.prototype.getParentSize = function() {
		if (!this.parentNode) return this.window ? {
			width: this.window.innerWidth,
			height: this.window.innerHeight
		} : {
			width: 0,
			height: 0
		};
		var e = this.appendBase();
		if (!e) return {
			width: 0,
			height: 0
		};
		var t = !1, n = this.parentNode.style.flexWrap;
		n !== "wrap" && (t = !0, this.parentNode.style.flexWrap = "wrap"), e.style.position = "relative", e.style.minWidth = "100%", e.style.minHeight = "100%";
		var r = {
			width: e.offsetWidth,
			height: e.offsetHeight
		};
		return t && (this.parentNode.style.flexWrap = n), this.removeBase(e), r;
	}, t.prototype.bindEvents = function() {
		this.window && (this.window.addEventListener("mouseup", this.onMouseUp), this.window.addEventListener("mousemove", this.onMouseMove), this.window.addEventListener("mouseleave", this.onMouseUp), this.window.addEventListener("touchmove", this.onMouseMove, {
			capture: !0,
			passive: !1
		}), this.window.addEventListener("touchend", this.onMouseUp));
	}, t.prototype.unbindEvents = function() {
		this.window && (this.window.removeEventListener("mouseup", this.onMouseUp), this.window.removeEventListener("mousemove", this.onMouseMove), this.window.removeEventListener("mouseleave", this.onMouseUp), this.window.removeEventListener("touchmove", this.onMouseMove, !0), this.window.removeEventListener("touchend", this.onMouseUp));
	}, t.prototype.componentDidMount = function() {
		if (this.resizable && this.window) {
			var e = this.window.getComputedStyle(this.resizable);
			this.setState({
				width: this.state.width || this.size.width,
				height: this.state.height || this.size.height,
				flexBasis: e.flexBasis === "auto" ? void 0 : e.flexBasis
			});
		}
	}, t.prototype.componentWillUnmount = function() {
		this.window && this.unbindEvents();
	}, t.prototype.createSizeForCssProperty = function(e, t) {
		var n = this.propsSize && this.propsSize[t];
		return this.state[t] === "auto" && this.state.original[t] === e && (n === void 0 || n === "auto") ? "auto" : e;
	}, t.prototype.calculateNewMaxFromBoundary = function(e, t) {
		var n = this.props.boundsByDirection, r = this.state.direction, i = n && Y("left", r), a = n && Y("top", r), o, s;
		if (this.props.bounds === "parent") {
			var c = this.parentNode;
			c && (o = i ? this.resizableRight - this.parentLeft : c.offsetWidth + (this.parentLeft - this.resizableLeft), s = a ? this.resizableBottom - this.parentTop : c.offsetHeight + (this.parentTop - this.resizableTop));
		} else this.props.bounds === "window" ? this.window && (o = i ? this.resizableRight : this.window.innerWidth - this.resizableLeft, s = a ? this.resizableBottom : this.window.innerHeight - this.resizableTop) : this.props.bounds && (o = i ? this.resizableRight - this.targetLeft : this.props.bounds.offsetWidth + (this.targetLeft - this.resizableLeft), s = a ? this.resizableBottom - this.targetTop : this.props.bounds.offsetHeight + (this.targetTop - this.resizableTop));
		return o && Number.isFinite(o) && (e = e && e < o ? e : o), s && Number.isFinite(s) && (t = t && t < s ? t : s), {
			maxWidth: e,
			maxHeight: t
		};
	}, t.prototype.calculateNewSizeFromDirection = function(e, t) {
		var n = this.props.scale || 1, r = be(this.props.resizeRatio || 1), i = r[0], a = r[1], o = this.state, s = o.direction, c = o.original, l = this.props, u = l.lockAspectRatio, d = l.lockAspectRatioExtraHeight, f = l.lockAspectRatioExtraWidth, p = c.width, m = c.height, h = d || 0, g = f || 0;
		return Y("right", s) && (p = c.width + (e - c.x) * i / n, u && (m = (p - g) / this.ratio + h)), Y("left", s) && (p = c.width - (e - c.x) * i / n, u && (m = (p - g) / this.ratio + h)), Y("bottom", s) && (m = c.height + (t - c.y) * a / n, u && (p = (m - h) * this.ratio + g)), Y("top", s) && (m = c.height - (t - c.y) * a / n, u && (p = (m - h) * this.ratio + g)), {
			newWidth: p,
			newHeight: m
		};
	}, t.prototype.calculateNewSizeFromAspectRatio = function(e, t, n, r) {
		var i = this.props, a = i.lockAspectRatio, o = i.lockAspectRatioExtraHeight, s = i.lockAspectRatioExtraWidth, c = r.width === void 0 ? 10 : r.width, l = n.width === void 0 || n.width < 0 ? e : n.width, u = r.height === void 0 ? 10 : r.height, d = n.height === void 0 || n.height < 0 ? t : n.height, f = o || 0, p = s || 0;
		if (a) {
			var m = (u - f) * this.ratio + p, h = (d - f) * this.ratio + p, g = (c - p) / this.ratio + f, _ = (l - p) / this.ratio + f, v = Math.max(c, m), y = Math.min(l, h), b = Math.max(u, g), x = Math.min(d, _);
			e = J(e, v, y), t = J(t, b, x);
		} else e = J(e, c, l), t = J(t, u, d);
		return {
			newWidth: e,
			newHeight: t
		};
	}, t.prototype.setBoundingClientRect = function() {
		var e = 1 / (this.props.scale || 1);
		if (this.props.bounds === "parent") {
			var t = this.parentNode;
			if (t) {
				var n = t.getBoundingClientRect();
				this.parentLeft = n.left * e, this.parentTop = n.top * e;
			}
		}
		if (this.props.bounds && typeof this.props.bounds != "string") {
			var r = this.props.bounds.getBoundingClientRect();
			this.targetLeft = r.left * e, this.targetTop = r.top * e;
		}
		if (this.resizable) {
			var i = this.resizable.getBoundingClientRect(), a = i.left, o = i.top, s = i.right, c = i.bottom;
			this.resizableLeft = a * e, this.resizableRight = s * e, this.resizableTop = o * e, this.resizableBottom = c * e;
		}
	}, t.prototype.onResizeStart = function(e, t) {
		if (this.resizable && this.window) {
			var n = 0, r = 0;
			if (e.nativeEvent && _e(e.nativeEvent) ? (n = e.nativeEvent.clientX, r = e.nativeEvent.clientY) : e.nativeEvent && X(e.nativeEvent) && (n = e.nativeEvent.touches[0].clientX, r = e.nativeEvent.touches[0].clientY), !(this.props.onResizeStart && this.resizable && this.props.onResizeStart(e, t, this.resizable) === !1)) {
				this.props.size && (this.props.size.height !== void 0 && this.props.size.height !== this.state.height && this.setState({ height: this.props.size.height }), this.props.size.width !== void 0 && this.props.size.width !== this.state.width && this.setState({ width: this.props.size.width })), this.ratio = typeof this.props.lockAspectRatio == "number" ? this.props.lockAspectRatio : this.size.width / this.size.height;
				var i, a = this.window.getComputedStyle(this.resizable);
				if (a.flexBasis !== "auto") {
					var o = this.parentNode;
					if (o) {
						var s = this.window.getComputedStyle(o).flexDirection;
						this.flexDir = s.startsWith("row") ? "row" : "column", i = a.flexBasis;
					}
				}
				this.setBoundingClientRect(), this.bindEvents();
				var c = {
					original: {
						x: n,
						y: r,
						width: this.size.width,
						height: this.size.height
					},
					isResizing: !0,
					backgroundStyle: q(q({}, this.state.backgroundStyle), { cursor: this.window.getComputedStyle(e.target).cursor || "auto" }),
					direction: t,
					flexBasis: i
				};
				this.setState(c);
			}
		}
	}, t.prototype.onMouseMove = function(e) {
		var t = this;
		if (this.state.isResizing && this.resizable && this.window) {
			if (this.window.TouchEvent && X(e)) try {
				e.preventDefault(), e.stopPropagation();
			} catch {}
			var n = this.props, r = n.maxWidth, i = n.maxHeight, a = n.minWidth, o = n.minHeight, s = X(e) ? e.touches[0].clientX : e.clientX, l = X(e) ? e.touches[0].clientY : e.clientY, u = this.state, d = u.direction, f = u.original, p = u.width, m = u.height, h = this.getParentSize(), g = ye(h, this.window.innerWidth, this.window.innerHeight, r, i, a, o);
			r = g.maxWidth, i = g.maxHeight, a = g.minWidth, o = g.minHeight;
			var _ = this.calculateNewSizeFromDirection(s, l), v = _.newHeight, y = _.newWidth, b = this.calculateNewMaxFromBoundary(r, i);
			this.props.snap && this.props.snap.x && (y = ve(y, this.props.snap.x, this.props.snapGap)), this.props.snap && this.props.snap.y && (v = ve(v, this.props.snap.y, this.props.snapGap));
			var x = this.calculateNewSizeFromAspectRatio(y, v, {
				width: b.maxWidth,
				height: b.maxHeight
			}, {
				width: a,
				height: o
			});
			if (y = x.newWidth, v = x.newHeight, this.props.grid) {
				var S = ge(y, this.props.grid[0], this.props.gridGap ? this.props.gridGap[0] : 0), C = ge(v, this.props.grid[1], this.props.gridGap ? this.props.gridGap[1] : 0), w = this.props.snapGap || 0, T = w === 0 || Math.abs(S - y) <= w ? S : y, E = w === 0 || Math.abs(C - v) <= w ? C : v;
				y = T, v = E;
			}
			var D = {
				width: y - f.width,
				height: v - f.height
			};
			if (this.delta = D, p && typeof p == "string") {
				if (p.endsWith("%")) {
					var O = y / h.width * 100;
					y = `${O}%`;
				} else if (p.endsWith("vw")) {
					var k = y / this.window.innerWidth * 100;
					y = `${k}vw`;
				} else if (p.endsWith("vh")) {
					var A = y / this.window.innerHeight * 100;
					y = `${A}vh`;
				}
			}
			if (m && typeof m == "string") {
				if (m.endsWith("%")) {
					var O = v / h.height * 100;
					v = `${O}%`;
				} else if (m.endsWith("vw")) {
					var k = v / this.window.innerWidth * 100;
					v = `${k}vw`;
				} else if (m.endsWith("vh")) {
					var A = v / this.window.innerHeight * 100;
					v = `${A}vh`;
				}
			}
			var j = {
				width: this.createSizeForCssProperty(y, "width"),
				height: this.createSizeForCssProperty(v, "height")
			};
			this.flexDir === "row" ? j.flexBasis = j.width : this.flexDir === "column" && (j.flexBasis = j.height);
			var M = this.state.width !== j.width, ee = this.state.height !== j.height, N = this.state.flexBasis !== j.flexBasis, P = M || ee || N;
			P && (0, c.flushSync)(function() {
				t.setState(j);
			}), this.props.onResize && P && this.props.onResize(e, d, this.resizable, D);
		}
	}, t.prototype.onMouseUp = function(e) {
		var t = this.state, n = t.isResizing, r = t.direction;
		t.original, n && this.resizable && (this.props.onResizeStop && this.props.onResizeStop(e, r, this.resizable, this.delta), this.props.size && this.setState({
			width: this.props.size.width ?? "auto",
			height: this.props.size.height ?? "auto"
		}), this.unbindEvents(), this.setState({
			isResizing: !1,
			backgroundStyle: q(q({}, this.state.backgroundStyle), { cursor: "auto" })
		}));
	}, t.prototype.updateSize = function(e) {
		this.setState({
			width: e.width ?? "auto",
			height: e.height ?? "auto"
		});
	}, t.prototype.renderResizer = function() {
		var e = this, t = this.props, n = t.enable, r = t.handleStyles, i = t.handleClasses, a = t.handleWrapperStyle, o = t.handleWrapperClass, s = t.handleComponent;
		if (!n) return null;
		var c = Object.keys(n).map(function(t) {
			return n[t] === !1 ? null : (0, W.jsx)(pe, {
				direction: t,
				onResizeStart: e.onResizeStart,
				replaceStyles: r && r[t],
				className: i && i[t],
				children: s && s[t] ? s[t] : null
			}, t);
		});
		return (0, W.jsx)("div", {
			className: o,
			style: a,
			children: c
		});
	}, t.prototype.render = function() {
		var e = this, t = Object.keys(this.props).reduce(function(t, n) {
			return xe.indexOf(n) === -1 && (t[n] = e.props[n]), t;
		}, {}), n = q(q(q({
			position: "relative",
			userSelect: this.state.isResizing ? "none" : "auto"
		}, this.props.style), this.sizeStyle), {
			maxWidth: this.props.maxWidth,
			maxHeight: this.props.maxHeight,
			minWidth: this.props.minWidth,
			minHeight: this.props.minHeight,
			boxSizing: "border-box",
			flexShrink: 0
		});
		this.state.flexBasis && (n.flexBasis = this.state.flexBasis);
		var r = this.props.as || "div";
		return (0, W.jsxs)(r, q({
			style: n,
			className: this.props.className
		}, t, {
			ref: function(t) {
				t && (e.resizable = t);
			},
			children: [
				this.state.isResizing && (0, W.jsx)("div", { style: this.state.backgroundStyle }),
				this.props.children,
				this.renderResizer()
			]
		}));
	}, t.defaultProps = {
		as: "div",
		onResizeStart: function() {},
		onResize: function() {},
		onResizeStop: function() {},
		enable: {
			top: !0,
			right: !0,
			bottom: !0,
			left: !0,
			topRight: !0,
			bottomRight: !0,
			bottomLeft: !0,
			topLeft: !0
		},
		style: {},
		grid: [1, 1],
		gridGap: [0, 0],
		lockAspectRatio: !1,
		lockAspectRatioExtraWidth: 0,
		lockAspectRatioExtraHeight: 0,
		scale: 1,
		resizeRatio: 1,
		snapGap: 0
	}, t;
}(o.PureComponent), we = function(e, t) {
	return we = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(e, t) {
		e.__proto__ = t;
	} || function(e, t) {
		for (var n in t) t.hasOwnProperty(n) && (e[n] = t[n]);
	}, we(e, t);
};
function Te(e, t) {
	we(e, t);
	function n() {
		this.constructor = e;
	}
	e.prototype = t === null ? Object.create(t) : (n.prototype = t.prototype, new n());
}
var $ = function() {
	return $ = Object.assign || function(e) {
		for (var t, n = 1, r = arguments.length; n < r; n++) for (var i in t = arguments[n], t) Object.prototype.hasOwnProperty.call(t, i) && (e[i] = t[i]);
		return e;
	}, $.apply(this, arguments);
};
function Ee(e, t) {
	var n = {};
	for (var r in e) Object.prototype.hasOwnProperty.call(e, r) && t.indexOf(r) < 0 && (n[r] = e[r]);
	if (e != null && typeof Object.getOwnPropertySymbols == "function") for (var i = 0, r = Object.getOwnPropertySymbols(e); i < r.length; i++) t.indexOf(r[i]) < 0 && Object.prototype.propertyIsEnumerable.call(e, r[i]) && (n[r[i]] = e[r[i]]);
	return n;
}
var De = {
	width: "auto",
	height: "auto",
	display: "inline-block",
	position: "absolute",
	top: 0,
	left: 0
}, Oe = function(e) {
	return {
		bottom: e,
		bottomLeft: e,
		bottomRight: e,
		left: e,
		right: e,
		top: e,
		topLeft: e,
		topRight: e
	};
}, ke = function(e) {
	Te(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.resizingPosition = {
			x: 0,
			y: 0
		}, n.offsetFromParent = {
			left: 0,
			top: 0
		}, n.resizableElement = { current: null }, n.originalPosition = {
			x: 0,
			y: 0
		}, n.state = {
			resizing: !1,
			bounds: {
				top: 0,
				right: 0,
				bottom: 0,
				left: 0
			},
			maxWidth: t.maxWidth,
			maxHeight: t.maxHeight
		}, n.onResizeStart = n.onResizeStart.bind(n), n.onResize = n.onResize.bind(n), n.onResizeStop = n.onResizeStop.bind(n), n.onDragStart = n.onDragStart.bind(n), n.onDrag = n.onDrag.bind(n), n.onDragStop = n.onDragStop.bind(n), n.getMaxSizesFromProps = n.getMaxSizesFromProps.bind(n), n;
	}
	return t.prototype.componentDidMount = function() {
		this.updateOffsetFromParent();
		var e = this.offsetFromParent, t = e.left, n = e.top, r = this.getDraggablePosition(), i = r.x, a = r.y;
		this.draggable.setState({
			x: i - t,
			y: a - n
		}), this.forceUpdate();
	}, t.prototype.getDraggablePosition = function() {
		var e = this.draggable.state;
		return {
			x: e.x,
			y: e.y
		};
	}, t.prototype.getParent = function() {
		return this.resizable && this.resizable.parentNode;
	}, t.prototype.getParentSize = function() {
		return this.resizable.getParentSize();
	}, t.prototype.getMaxSizesFromProps = function() {
		return {
			maxWidth: this.props.maxWidth === void 0 ? 2 ** 53 - 1 : this.props.maxWidth,
			maxHeight: this.props.maxHeight === void 0 ? 2 ** 53 - 1 : this.props.maxHeight
		};
	}, t.prototype.getSelfElement = function() {
		return this.resizable && this.resizable.resizable;
	}, t.prototype.getOffsetHeight = function(e) {
		var t = this.props.scale;
		switch (this.props.bounds) {
			case "window": return window.innerHeight / t;
			case "body": return document.body.offsetHeight / t;
			default: return e.offsetHeight;
		}
	}, t.prototype.getOffsetWidth = function(e) {
		var t = this.props.scale;
		switch (this.props.bounds) {
			case "window": return window.innerWidth / t;
			case "body": return document.body.offsetWidth / t;
			default: return e.offsetWidth;
		}
	}, t.prototype.onDragStart = function(e, t) {
		if (this.props.onDragStart && this.props.onDragStart(e, t) === !1) return !1;
		var n = this.getDraggablePosition();
		if (this.originalPosition = n, this.props.bounds) {
			var r = this.getParent(), i = this.props.scale, a;
			if (this.props.bounds === "parent") a = r;
			else if (this.props.bounds === "body") {
				var o = r.getBoundingClientRect(), s = o.left, c = o.top, l = document.body.getBoundingClientRect(), u = -(s - r.offsetLeft * i - l.left) / i, d = -(c - r.offsetTop * i - l.top) / i, f = (document.body.offsetWidth - this.resizable.size.width * i) / i + u, p = (document.body.offsetHeight - this.resizable.size.height * i) / i + d;
				return this.setState({ bounds: {
					top: d,
					right: f,
					bottom: p,
					left: u
				} });
			} else if (this.props.bounds === "window") {
				if (!this.resizable) return;
				var m = r.getBoundingClientRect(), h = m.left, g = m.top, _ = -(h - r.offsetLeft * i) / i, v = -(g - r.offsetTop * i) / i, f = (window.innerWidth - this.resizable.size.width * i) / i + _, p = (window.innerHeight - this.resizable.size.height * i) / i + v;
				return this.setState({ bounds: {
					top: v,
					right: f,
					bottom: p,
					left: _
				} });
			} else typeof this.props.bounds == "string" ? a = document.querySelector(this.props.bounds) : this.props.bounds instanceof HTMLElement && (a = this.props.bounds);
			if (a instanceof HTMLElement && r instanceof HTMLElement) {
				var y = a.getBoundingClientRect(), b = y.left, x = y.top, S = r.getBoundingClientRect(), C = S.left, w = S.top, T = (b - C) / i, E = x - w;
				if (this.resizable) {
					this.updateOffsetFromParent();
					var D = this.offsetFromParent;
					this.setState({ bounds: {
						top: E - D.top,
						right: T + (a.offsetWidth - this.resizable.size.width) - D.left / i,
						bottom: E + (a.offsetHeight - this.resizable.size.height) - D.top,
						left: T - D.left / i
					} });
				}
			}
		}
	}, t.prototype.onDrag = function(e, t) {
		if (this.props.onDrag) {
			var n = this.offsetFromParent, r = n.left, i = n.top;
			if (!this.props.dragAxis || this.props.dragAxis === "both") return this.props.onDrag(e, $($({}, t), {
				x: t.x + r,
				y: t.y + i
			}));
			if (this.props.dragAxis === "x") return this.props.onDrag(e, $($({}, t), {
				x: t.x + r,
				y: this.originalPosition.y + i,
				deltaY: 0
			}));
			if (this.props.dragAxis === "y") return this.props.onDrag(e, $($({}, t), {
				x: this.originalPosition.x + r,
				y: t.y + i,
				deltaX: 0
			}));
		}
	}, t.prototype.onDragStop = function(e, t) {
		if (this.props.onDragStop) {
			var n = this.offsetFromParent, r = n.left, i = n.top;
			if (!this.props.dragAxis || this.props.dragAxis === "both") return this.props.onDragStop(e, $($({}, t), {
				x: t.x + r,
				y: t.y + i
			}));
			if (this.props.dragAxis === "x") return this.props.onDragStop(e, $($({}, t), {
				x: t.x + r,
				y: this.originalPosition.y + i,
				deltaY: 0
			}));
			if (this.props.dragAxis === "y") return this.props.onDragStop(e, $($({}, t), {
				x: this.originalPosition.x + r,
				y: t.y + i,
				deltaX: 0
			}));
		}
	}, t.prototype.onResizeStart = function(e, t, n) {
		if (this.props.onResizeStart && this.props.onResizeStart(e, t, n) === !1) return !1;
		e.stopPropagation(), this.setState({ resizing: !0 });
		var r = this.props.scale, i = this.offsetFromParent, a = this.getDraggablePosition();
		if (this.resizingPosition = {
			x: a.x + i.left,
			y: a.y + i.top
		}, this.originalPosition = a, this.props.bounds) {
			var o = this.getParent(), s = void 0;
			this.props.bounds === "parent" ? s = o : this.props.bounds === "body" ? s = document.body : this.props.bounds === "window" ? s = window : typeof this.props.bounds == "string" ? s = document.querySelector(this.props.bounds) : this.props.bounds instanceof HTMLElement && (s = this.props.bounds);
			var c = this.getSelfElement();
			if (c instanceof Element && (s instanceof HTMLElement || s === window) && o instanceof HTMLElement) {
				var l = this.getMaxSizesFromProps(), u = l.maxWidth, d = l.maxHeight, f = this.getParentSize();
				if (u && typeof u == "string") {
					if (u.endsWith("%")) {
						var p = Number(u.replace("%", "")) / 100;
						u = f.width * p;
					} else u.endsWith("px") && (u = Number(u.replace("px", "")));
				}
				if (d && typeof d == "string") {
					if (d.endsWith("%")) {
						var p = Number(d.replace("%", "")) / 100;
						d = f.height * p;
					} else d.endsWith("px") && (d = Number(d.replace("px", "")));
				}
				var m = c.getBoundingClientRect(), h = m.left, g = m.top, _ = this.props.bounds === "window" ? {
					left: 0,
					top: 0
				} : s.getBoundingClientRect(), v = _.left, y = _.top, b = this.getOffsetWidth(s), x = this.getOffsetHeight(s), S = t.toLowerCase().endsWith("left"), C = t.toLowerCase().endsWith("right"), w = t.startsWith("top"), T = t.startsWith("bottom");
				if ((S || w) && this.resizable) {
					var E = (h - v) / r + this.resizable.size.width;
					this.setState({ maxWidth: E > Number(u) ? u : E });
				}
				if (C || this.props.lockAspectRatio && !S && !w) {
					var E = b + (v - h) / r;
					this.setState({ maxWidth: E > Number(u) ? u : E });
				}
				if ((w || S) && this.resizable) {
					var E = (g - y) / r + this.resizable.size.height;
					this.setState({ maxHeight: E > Number(d) ? d : E });
				}
				if (T || this.props.lockAspectRatio && !w && !S) {
					var E = x + (y - g) / r;
					this.setState({ maxHeight: E > Number(d) ? d : E });
				}
			}
		} else this.setState({
			maxWidth: this.props.maxWidth,
			maxHeight: this.props.maxHeight
		});
	}, t.prototype.onResize = function(e, t, n, r) {
		var i = this, a = {
			x: this.originalPosition.x,
			y: this.originalPosition.y
		}, o = -r.width, s = -r.height;
		[
			"top",
			"left",
			"topLeft",
			"bottomLeft",
			"topRight"
		].includes(t) && (t === "bottomLeft" ? a.x += o : (t === "topRight" || (a.x += o), a.y += s));
		var l = this.draggable.state;
		(a.x !== l.x || a.y !== l.y) && (0, c.flushSync)(function() {
			i.draggable.setState(a);
		}), this.updateOffsetFromParent();
		var u = this.offsetFromParent, d = this.getDraggablePosition().x + u.left, f = this.getDraggablePosition().y + u.top;
		this.resizingPosition = {
			x: d,
			y: f
		}, this.props.onResize && this.props.onResize(e, t, n, r, {
			x: d,
			y: f
		});
	}, t.prototype.onResizeStop = function(e, t, n, r) {
		this.setState({ resizing: !1 });
		var i = this.getMaxSizesFromProps(), a = i.maxWidth, o = i.maxHeight;
		this.setState({
			maxWidth: a,
			maxHeight: o
		}), this.props.onResizeStop && this.props.onResizeStop(e, t, n, r, this.resizingPosition);
	}, t.prototype.updateSize = function(e) {
		this.resizable && this.resizable.updateSize({
			width: e.width,
			height: e.height
		});
	}, t.prototype.updatePosition = function(e) {
		this.draggable.setState(e);
	}, t.prototype.updateOffsetFromParent = function() {
		var e = this.props.scale, t = this.getParent(), n = this.getSelfElement();
		if (!t || n === null) return {
			top: 0,
			left: 0
		};
		var r = t.getBoundingClientRect(), i = r.left, a = r.top, o = n.getBoundingClientRect(), s = this.getDraggablePosition(), c = t.scrollLeft, l = t.scrollTop;
		this.offsetFromParent = {
			left: o.left - i + c - s.x * e,
			top: o.top - a + l - s.y * e
		};
	}, t.prototype.render = function() {
		var e = this, t = this.props, n = t.disableDragging, r = t.style, i = t.dragHandleClassName, a = t.position, s = t.onMouseDown, c = t.onMouseUp, l = t.dragAxis, u = t.dragGrid, d = t.bounds, f = t.enableUserSelectHack, p = t.cancel, m = t.children;
		t.onResizeStart, t.onResize, t.onResizeStop, t.onDragStart, t.onDrag, t.onDragStop;
		var h = t.resizeHandleStyles, g = t.resizeHandleClasses, _ = t.resizeHandleComponent, v = t.enableResizing, y = t.resizeGrid, b = t.resizeHandleWrapperClass, x = t.resizeHandleWrapperStyle, S = t.scale, C = t.allowAnyClick, w = t.dragPositionOffset, T = Ee(t, /* @__PURE__ */ "disableDragging.style.dragHandleClassName.position.onMouseDown.onMouseUp.dragAxis.dragGrid.bounds.enableUserSelectHack.cancel.children.onResizeStart.onResize.onResizeStop.onDragStart.onDrag.onDragStop.resizeHandleStyles.resizeHandleClasses.resizeHandleComponent.enableResizing.resizeGrid.resizeHandleWrapperClass.resizeHandleWrapperStyle.scale.allowAnyClick.dragPositionOffset".split(".")), E = this.props.default ? $({}, this.props.default) : void 0;
		delete T.default;
		var D = n || i ? { cursor: "auto" } : { cursor: "move" }, O = $($($({}, De), D), r), k = this.offsetFromParent, A = k.left, j = k.top, M;
		a && (M = {
			x: a.x - A,
			y: a.y - j
		});
		var ee = this.state.resizing ? void 0 : M, N = this.state.resizing ? "both" : l;
		return (0, o.createElement)(le, {
			ref: function(t) {
				t && (e.draggable = t);
			},
			handle: i ? `.${i}` : void 0,
			defaultPosition: E,
			onMouseDown: s,
			onMouseUp: c,
			onStart: this.onDragStart,
			onDrag: this.onDrag,
			onStop: this.onDragStop,
			axis: N,
			disabled: n,
			grid: u,
			bounds: d ? this.state.bounds : void 0,
			position: ee,
			enableUserSelectHack: f,
			cancel: p,
			scale: S,
			allowAnyClick: C,
			nodeRef: this.resizableElement,
			positionOffset: w
		}, (0, o.createElement)(Ce, $({}, T, {
			ref: function(t) {
				t && (e.resizable = t, e.resizableElement.current = t.resizable);
			},
			defaultSize: E,
			size: this.props.size,
			enable: typeof v == "boolean" ? Oe(v) : v,
			onResizeStart: this.onResizeStart,
			onResize: this.onResize,
			onResizeStop: this.onResizeStop,
			style: O,
			minWidth: this.props.minWidth,
			minHeight: this.props.minHeight,
			maxWidth: this.state.resizing ? this.state.maxWidth : this.props.maxWidth,
			maxHeight: this.state.resizing ? this.state.maxHeight : this.props.maxHeight,
			grid: y,
			handleWrapperClass: b,
			handleWrapperStyle: x,
			lockAspectRatio: this.props.lockAspectRatio,
			lockAspectRatioExtraWidth: this.props.lockAspectRatioExtraWidth,
			lockAspectRatioExtraHeight: this.props.lockAspectRatioExtraHeight,
			handleStyles: h,
			handleClasses: g,
			handleComponent: _,
			scale: this.props.scale
		}), m));
	}, t.defaultProps = {
		maxWidth: 2 ** 53 - 1,
		maxHeight: 2 ** 53 - 1,
		scale: 1,
		onResizeStart: function() {},
		onResize: function() {},
		onResizeStop: function() {},
		onDragStart: function() {},
		onDrag: function() {},
		onDragStop: function() {}
	}, t;
}(o.PureComponent);
export { ke as Rnd };
