// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-editor/0.48.0/79f9a4ebbdbe59855876269cb57343d612db5a7cea425c949c960b7cf68538da/node_modules/monaco-editor/esm/vs/basic-languages/ini/ini.js
var e = {
	comments: { lineComment: "#" },
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
			close: "\""
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
			open: "\"",
			close: "\""
		},
		{
			open: "'",
			close: "'"
		}
	]
}, t = {
	defaultToken: "",
	tokenPostfix: ".ini",
	escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,
	tokenizer: {
		root: [
			[/^\[[^\]]*\]/, "metatag"],
			[/(^\w+)(\s*)(\=)/, [
				"key",
				"",
				"delimiter"
			]],
			{ include: "@whitespace" },
			[/\d+/, "number"],
			[/"([^"\\]|\\.)*$/, "string.invalid"],
			[/'([^'\\]|\\.)*$/, "string.invalid"],
			[
				/"/,
				"string",
				"@string.\""
			],
			[
				/'/,
				"string",
				"@string.'"
			]
		],
		whitespace: [[/[ \t\r\n]+/, ""], [/^\s*[#;].*$/, "comment"]],
		string: [
			[/[^\\"']+/, "string"],
			[/@escapes/, "string.escape"],
			[/\\./, "string.escape.invalid"],
			[/["']/, { cases: {
				"$#==$S2": {
					token: "string",
					next: "@pop"
				},
				"@default": "string"
			} }]
		]
	}
};
export { e as conf, t as language };
