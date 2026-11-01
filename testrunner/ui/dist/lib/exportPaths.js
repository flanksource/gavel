// src/exportPaths.ts
var GO_TEST_FRAMEWORK = "go test";
function insertSpaces(name) {
	let out = "";
	for (let i = 0; i < name.length; i++) {
		const char = name[i];
		const prev = i > 0 ? name[i - 1] : "";
		if (i > 0 && /\p{Lu}/u.test(char) && (/\p{Ll}/u.test(prev) || /\p{Nd}/u.test(prev))) out += " ";
		out += char;
	}
	return out;
}
function displayTestName(name, framework) {
	if (framework !== GO_TEST_FRAMEWORK || name.endsWith("/")) return name;
	return name.split("/").map((part, i) => i === 0 ? insertSpaces(part.replace(/^Test/, "")) : part.replaceAll("_", " ")).join(" / ");
}
function slugify(input) {
	return input.toLowerCase().replace(/[^\p{L}\p{N}]+/gu, "-").replace(/^-+|-+$/g, "") || "node";
}
/** Annotates every node with the path the server's export of that node is requested at. */
function withExportPaths(tests, parent = []) {
	const slugs = tests.map((test) => slugify(displayTestName(test.name, test.framework)));
	const counts = /* @__PURE__ */ new Map();
	slugs.forEach((slug) => counts.set(slug, (counts.get(slug) ?? 0) + 1));
	const seen = /* @__PURE__ */ new Map();
	return tests.map((test, i) => {
		const slug = slugs[i];
		const ordinal = (seen.get(slug) ?? 0) + 1;
		seen.set(slug, ordinal);
		const path = [...parent, counts.get(slug) > 1 ? `${slug}~${ordinal}` : slug];
		return {
			...test,
			export_path: path.join("/"),
			...test.children ? { children: withExportPaths(test.children, path) } : {}
		};
	});
}
export { withExportPaths };

//# sourceMappingURL=exportPaths.js.map