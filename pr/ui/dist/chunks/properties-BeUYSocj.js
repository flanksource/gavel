// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.3/6ddc4bb142089e0be26db8a3e2cc436edc8eddd035daf11477b72a920064aef3/node_modules/@codemirror/legacy-modes/mode/properties.js
var e = {
	name: "properties",
	token: function(e, t) {
		var n = e.sol() || t.afterSection, r = e.eol();
		if (t.afterSection = !1, n && (t.nextMultiline ? (t.inMultiline = !0, t.nextMultiline = !1) : t.position = "def"), r && !t.nextMultiline && (t.inMultiline = !1, t.position = "def"), n) for (; e.eatSpace(););
		var i = e.next();
		return n && (i === "#" || i === "!" || i === ";") ? (t.position = "comment", e.skipToEnd(), "comment") : n && i === "[" ? (t.afterSection = !0, e.skipTo("]"), e.eat("]"), "header") : i === "=" || i === ":" ? (t.position = "quote", null) : (i === "\\" && t.position === "quote" && e.eol() && (t.nextMultiline = !0), t.position);
	},
	startState: function() {
		return {
			position: "def",
			nextMultiline: !1,
			inMultiline: !1,
			afterSection: !1
		};
	}
};
export { e as properties };
