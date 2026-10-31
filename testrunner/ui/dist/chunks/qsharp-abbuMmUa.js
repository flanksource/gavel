// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-editor/0.48.0/79f9a4ebbdbe59855876269cb57343d612db5a7cea425c949c960b7cf68538da/node_modules/monaco-editor/esm/vs/basic-languages/qsharp/qsharp.js
var e = {
	comments: { lineComment: "//" },
	brackets: [
		["{", "}"],
		["[", "]"],
		["(", ")"]
	],
	autoClosingPairs: [
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
		},
		{
			open: "\"",
			close: "\"",
			notIn: ["string", "comment"]
		}
	],
	surroundingPairs: [
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
		},
		{
			open: "\"",
			close: "\""
		}
	]
}, t = {
	keywords: /* @__PURE__ */ "namespace.open.as.operation.function.body.adjoint.newtype.controlled.if.elif.else.repeat.until.fixup.for.in.while.return.fail.within.apply.Adjoint.Controlled.Adj.Ctl.is.self.auto.distribute.invert.intrinsic.let.set.w/.new.not.and.or.use.borrow.using.borrowing.mutable.internal".split("."),
	typeKeywords: [
		"Unit",
		"Int",
		"BigInt",
		"Double",
		"Bool",
		"String",
		"Qubit",
		"Result",
		"Pauli",
		"Range"
	],
	invalidKeywords: /* @__PURE__ */ "abstract.base.bool.break.byte.case.catch.char.checked.class.const.continue.decimal.default.delegate.do.double.enum.event.explicit.extern.finally.fixed.float.foreach.goto.implicit.int.interface.lock.long.null.object.operator.out.override.params.private.protected.public.readonly.ref.sbyte.sealed.short.sizeof.stackalloc.static.string.struct.switch.this.throw.try.typeof.unit.ulong.unchecked.unsafe.ushort.virtual.void.volatile".split("."),
	constants: [
		"true",
		"false",
		"PauliI",
		"PauliX",
		"PauliY",
		"PauliZ",
		"One",
		"Zero"
	],
	builtin: /* @__PURE__ */ "X.Y.Z.H.HY.S.T.SWAP.CNOT.CCNOT.MultiX.R.RFrac.Rx.Ry.Rz.R1.R1Frac.Exp.ExpFrac.Measure.M.MultiM.Message.Length.Assert.AssertProb.AssertEqual".split("."),
	operators: /* @__PURE__ */ "and=,<-,->,*,*=,@,!,^,^=,:,::,..,==,...,=,=>,>,>=,<,<=,-,-=,!=,or=,%,%=,|,+,+=,?,/,/=,&&&,&&&=,^^^,^^^=,>>>,>>>=,<<<,<<<=,|||,|||=,~~~,_,w/,w/=".split(","),
	namespaceFollows: ["namespace", "open"],
	symbols: /[=><!~?:&|+\-*\/\^%@._]+/,
	escapes: /\\[\s\S]/,
	tokenizer: {
		root: [
			[/[a-zA-Z_$][\w$]*/, { cases: {
				"@namespaceFollows": {
					token: "keyword.$0",
					next: "@namespace"
				},
				"@typeKeywords": "type",
				"@keywords": "keyword",
				"@constants": "constant",
				"@builtin": "keyword",
				"@invalidKeywords": "invalid",
				"@default": "identifier"
			} }],
			{ include: "@whitespace" },
			[/[{}()\[\]]/, "@brackets"],
			[/@symbols/, { cases: {
				"@operators": "operator",
				"@default": ""
			} }],
			[/\d*\.\d+([eE][\-+]?\d+)?/, "number.float"],
			[/\d+/, "number"],
			[/[;,.]/, "delimiter"],
			[/"/, {
				token: "string.quote",
				bracket: "@open",
				next: "@string"
			}]
		],
		string: [
			[/[^\\"]+/, "string"],
			[/@escapes/, "string.escape"],
			[/"/, {
				token: "string.quote",
				bracket: "@close",
				next: "@pop"
			}]
		],
		namespace: [
			{ include: "@whitespace" },
			[/[A-Za-z]\w*/, "namespace"],
			[/[\.=]/, "delimiter"],
			[
				"",
				"",
				"@pop"
			]
		],
		whitespace: [[/[ \t\r\n]+/, "white"], [/(\/\/).*/, "comment"]]
	}
};
export { e as conf, t as language };
