// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-editor/0.48.0/79f9a4ebbdbe59855876269cb57343d612db5a7cea425c949c960b7cf68538da/node_modules/monaco-editor/esm/vs/basic-languages/sparql/sparql.js
var e = {
	comments: { lineComment: "#" },
	brackets: [
		["{", "}"],
		["[", "]"],
		["(", ")"]
	],
	autoClosingPairs: [
		{
			open: "'",
			close: "'",
			notIn: ["string"]
		},
		{
			open: "\"",
			close: "\"",
			notIn: ["string"]
		},
		{
			open: "{",
			close: "}"
		},
		{
			open: "[",
			close: "]"
		},
		{
			open: "(",
			close: ")"
		}
	]
}, t = {
	defaultToken: "",
	tokenPostfix: ".rq",
	brackets: [
		{
			token: "delimiter.curly",
			open: "{",
			close: "}"
		},
		{
			token: "delimiter.parenthesis",
			open: "(",
			close: ")"
		},
		{
			token: "delimiter.square",
			open: "[",
			close: "]"
		},
		{
			token: "delimiter.angle",
			open: "<",
			close: ">"
		}
	],
	keywords: /* @__PURE__ */ "add.as.asc.ask.base.by.clear.construct.copy.create.data.delete.desc.describe.distinct.drop.false.filter.from.graph.group.having.in.insert.limit.load.minus.move.named.not.offset.optional.order.prefix.reduced.select.service.silent.to.true.undef.union.using.values.where.with".split("."),
	builtinFunctions: /* @__PURE__ */ "a.abs.avg.bind.bnode.bound.ceil.coalesce.concat.contains.count.datatype.day.encode_for_uri.exists.floor.group_concat.hours.if.iri.isblank.isiri.isliteral.isnumeric.isuri.lang.langmatches.lcase.max.md5.min.minutes.month.now.rand.regex.replace.round.sameterm.sample.seconds.sha1.sha256.sha384.sha512.str.strafter.strbefore.strdt.strends.strlang.strlen.strstarts.struuid.substr.sum.timezone.tz.ucase.uri.uuid.year".split("."),
	ignoreCase: !0,
	tokenizer: {
		root: [
			[/<[^\s\u00a0>]*>?/, "tag"],
			{ include: "@strings" },
			[/#.*/, "comment"],
			[/[{}()\[\]]/, "@brackets"],
			[/[;,.]/, "delimiter"],
			[/[_\w\d]+:(\.(?=[\w_\-\\%])|[:\w_-]|\\[-\\_~.!$&'()*+,;=/?#@%]|%[a-f\d][a-f\d])*/, "tag"],
			[/:(\.(?=[\w_\-\\%])|[:\w_-]|\\[-\\_~.!$&'()*+,;=/?#@%]|%[a-f\d][a-f\d])+/, "tag"],
			[/[$?]?[_\w\d]+/, { cases: {
				"@keywords": { token: "keyword" },
				"@builtinFunctions": { token: "predefined.sql" },
				"@default": "identifier"
			} }],
			[/\^\^/, "operator.sql"],
			[/\^[*+\-<>=&|^\/!?]*/, "operator.sql"],
			[/[*+\-<>=&|\/!?]/, "operator.sql"],
			[/@[a-z\d\-]*/, "metatag.html"],
			[/\s+/, "white"]
		],
		strings: [
			[/'([^'\\]|\\.)*$/, "string.invalid"],
			[
				/'$/,
				"string.sql",
				"@pop"
			],
			[
				/'/,
				"string.sql",
				"@stringBody"
			],
			[/"([^"\\]|\\.)*$/, "string.invalid"],
			[
				/"$/,
				"string.sql",
				"@pop"
			],
			[
				/"/,
				"string.sql",
				"@dblStringBody"
			]
		],
		stringBody: [
			[/[^\\']+/, "string.sql"],
			[/\\./, "string.escape"],
			[
				/'/,
				"string.sql",
				"@pop"
			]
		],
		dblStringBody: [
			[/[^\\"]+/, "string.sql"],
			[/\\./, "string.escape"],
			[
				/"/,
				"string.sql",
				"@pop"
			]
		]
	}
};
export { e as conf, t as language };
