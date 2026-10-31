// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-editor/0.48.0/79f9a4ebbdbe59855876269cb57343d612db5a7cea425c949c960b7cf68538da/node_modules/monaco-editor/esm/vs/basic-languages/pascaligo/pascaligo.js
var e = {
	comments: {
		lineComment: "//",
		blockComment: ["(*", "*)"]
	},
	brackets: [
		["{", "}"],
		["[", "]"],
		["(", ")"],
		["<", ">"]
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
			open: "<",
			close: ">"
		},
		{
			open: "'",
			close: "'"
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
			open: "<",
			close: ">"
		},
		{
			open: "'",
			close: "'"
		}
	]
}, t = {
	defaultToken: "",
	tokenPostfix: ".pascaligo",
	ignoreCase: !0,
	brackets: [
		{
			open: "{",
			close: "}",
			token: "delimiter.curly"
		},
		{
			open: "[",
			close: "]",
			token: "delimiter.square"
		},
		{
			open: "(",
			close: ")",
			token: "delimiter.parenthesis"
		},
		{
			open: "<",
			close: ">",
			token: "delimiter.angle"
		}
	],
	keywords: [
		"begin",
		"block",
		"case",
		"const",
		"else",
		"end",
		"fail",
		"for",
		"from",
		"function",
		"if",
		"is",
		"nil",
		"of",
		"remove",
		"return",
		"skip",
		"then",
		"type",
		"var",
		"while",
		"with",
		"option",
		"None",
		"transaction"
	],
	typeKeywords: [
		"bool",
		"int",
		"list",
		"map",
		"nat",
		"record",
		"string",
		"unit",
		"address",
		"map",
		"mtz",
		"xtz"
	],
	operators: [
		"=",
		">",
		"<",
		"<=",
		">=",
		"<>",
		":",
		":=",
		"and",
		"mod",
		"or",
		"+",
		"-",
		"*",
		"/",
		"@",
		"&",
		"^",
		"%"
	],
	symbols: /[=><:@\^&|+\-*\/\^%]+/,
	tokenizer: {
		root: [
			[/[a-zA-Z_][\w]*/, { cases: {
				"@keywords": { token: "keyword.$0" },
				"@default": "identifier"
			} }],
			{ include: "@whitespace" },
			[/[{}()\[\]]/, "@brackets"],
			[/[<>](?!@symbols)/, "@brackets"],
			[/@symbols/, { cases: {
				"@operators": "delimiter",
				"@default": ""
			} }],
			[/\d*\.\d+([eE][\-+]?\d+)?/, "number.float"],
			[/\$[0-9a-fA-F]{1,16}/, "number.hex"],
			[/\d+/, "number"],
			[/[;,.]/, "delimiter"],
			[/'([^'\\]|\\.)*$/, "string.invalid"],
			[
				/'/,
				"string",
				"@string"
			],
			[/'[^\\']'/, "string"],
			[/'/, "string.invalid"],
			[/\#\d+/, "string"]
		],
		comment: [
			[/[^\(\*]+/, "comment"],
			[
				/\*\)/,
				"comment",
				"@pop"
			],
			[/\(\*/, "comment"]
		],
		string: [
			[/[^\\']+/, "string"],
			[/\\./, "string.escape.invalid"],
			[/'/, {
				token: "string.quote",
				bracket: "@close",
				next: "@pop"
			}]
		],
		whitespace: [
			[/[ \t\r\n]+/, "white"],
			[
				/\(\*/,
				"comment",
				"@comment"
			],
			[/\/\/.*$/, "comment"]
		]
	}
};
export { e as conf, t as language };
