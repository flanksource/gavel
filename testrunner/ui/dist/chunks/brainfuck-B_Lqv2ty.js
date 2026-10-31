// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.3/6ddc4bb142089e0be26db8a3e2cc436edc8eddd035daf11477b72a920064aef3/node_modules/@codemirror/legacy-modes/mode/brainfuck.js
var e = "><+-.,[]".split(""), t = {
	name: "brainfuck",
	startState: function() {
		return {
			commentLine: !1,
			left: 0,
			right: 0,
			commentLoop: !1
		};
	},
	token: function(t, n) {
		if (t.eatSpace()) return null;
		t.sol() && (n.commentLine = !1);
		var r = t.next().toString();
		if (e.indexOf(r) !== -1) {
			if (n.commentLine === !0) return t.eol() && (n.commentLine = !1), "comment";
			if (r === "]" || r === "[") return r === "[" ? n.left++ : n.right++, "bracket";
			if (r === "+" || r === "-") return "keyword";
			if (r === "<" || r === ">") return "atom";
			if (r === "." || r === ",") return "def";
		} else return n.commentLine = !0, t.eol() && (n.commentLine = !1), "comment";
		t.eol() && (n.commentLine = !1);
	}
};
export { t as brainfuck };
