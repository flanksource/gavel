// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.5/016bc79896894b8ba1a40f928f9bb248a78667ccdad19c4225df0f983f1ced27/node_modules/@codemirror/legacy-modes/mode/troff.js
var e = {};
function t(t) {
	if (t.eatSpace()) return null;
	var n = t.sol(), r = t.next();
	if (r === "\\") return t.match("fB") || t.match("fR") || t.match("fI") || t.match("u") || t.match("d") || t.match("%") || t.match("&") ? "string" : t.match("m[") ? (t.skipTo("]"), t.next(), "string") : t.match("s+") || t.match("s-") ? (t.eatWhile(/[\d-]/), "string") : ((t.match("(") || t.match("*(")) && t.eatWhile(/[\w-]/), "string");
	if (n && (r === "." || r === "'") && t.eat("\\") && t.eat("\"")) return t.skipToEnd(), "comment";
	if (n && r === ".") {
		if (t.match("B ") || t.match("I ") || t.match("R ")) return "attribute";
		if (t.match("TH ") || t.match("SH ") || t.match("SS ") || t.match("HP ")) return t.skipToEnd(), "quote";
		if (t.match(/[A-Z]/) && t.match(/[A-Z]/) || t.match(/[a-z]/) && t.match(/[a-z]/)) return "attribute";
	}
	t.eatWhile(/[\w-]/);
	var i = t.current();
	return e.hasOwnProperty(i) ? e[i] : null;
}
function n(e, n) {
	return (n.tokens[0] || t)(e, n);
}
var r = {
	name: "troff",
	startState: function() {
		return { tokens: [] };
	},
	token: function(e, t) {
		return n(e, t);
	}
};
export { r as troff };
