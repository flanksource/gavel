// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.5/016bc79896894b8ba1a40f928f9bb248a78667ccdad19c4225df0f983f1ced27/node_modules/@codemirror/legacy-modes/mode/diff.js
var e = {
	"+": "inserted",
	"-": "deleted",
	"@": "meta"
}, t = {
	name: "diff",
	token: function(t) {
		var n = t.string.search(/[\t ]+?$/);
		if (!t.sol() || n === 0) return t.skipToEnd(), ("error " + (e[t.string.charAt(0)] || "")).replace(/ $/, "");
		var r = e[t.peek()] || t.skipToEnd();
		return n === -1 ? t.skipToEnd() : t.pos = n, r;
	}
};
export { t as diff };
