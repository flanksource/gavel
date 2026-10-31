// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.3/6ddc4bb142089e0be26db8a3e2cc436edc8eddd035daf11477b72a920064aef3/node_modules/@codemirror/legacy-modes/mode/protobuf.js
function e(e) {
	return RegExp("^((" + e.join(")|(") + "))\\b", "i");
}
var t = /* @__PURE__ */ "package.message.import.syntax.required.optional.repeated.reserved.default.extensions.packed.bool.bytes.double.enum.float.string.int32.int64.uint32.uint64.sint32.sint64.fixed32.fixed64.sfixed32.sfixed64.option.service.rpc.returns".split("."), n = e(t), r = /* @__PURE__ */ RegExp("^[_A-Za-z¡-￿][_A-Za-z0-9¡-￿]*");
function i(e) {
	return e.eatSpace() ? null : e.match("//") ? (e.skipToEnd(), "comment") : e.match(/^[0-9\.+-]/, !1) && (e.match(/^[+-]?0x[0-9a-fA-F]+/) || e.match(/^[+-]?\d*\.\d+([EeDd][+-]?\d+)?/) || e.match(/^[+-]?\d+([EeDd][+-]?\d+)?/)) ? "number" : e.match(/^"([^"]|(""))*"/) || e.match(/^'([^']|(''))*'/) ? "string" : e.match(n) ? "keyword" : e.match(r) ? "variable" : (e.next(), null);
}
var a = {
	name: "protobuf",
	token: i,
	languageData: { autocomplete: t }
};
export { a as protobuf };
