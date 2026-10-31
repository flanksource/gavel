// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.3/6ddc4bb142089e0be26db8a3e2cc436edc8eddd035daf11477b72a920064aef3/node_modules/@codemirror/legacy-modes/mode/diff.js
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
