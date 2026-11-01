// src/lintRerun.ts
function normalize(path) {
	return path.replace(/\\/g, "/").replace(/\/+/g, "/").replace(/\/$/, "");
}
function relPath(file, workDir) {
	if (!file) return void 0;
	const normalized = normalize(file);
	if (!workDir) return normalized;
	const prefix = `${normalize(workDir)}/`;
	return normalized.startsWith(prefix) ? normalized.slice(prefix.length) : normalized;
}
function collectWorkDirs(node, out) {
	if (node.work_dir) out.add(node.work_dir);
	(node.children ?? []).forEach((child) => collectWorkDirs(child, out));
}
/**
* The rerun request that re-lints what a lint tree node covers: its linter, the
* files beneath it that still have findings, and the single work dir they share.
* Without a node it reruns every linter. `work_dir` is empty when the scope
* spans several lint roots, which the server cannot rerun as one.
*/
function buildLintRerunBody(node, lint) {
	const workDirs = /* @__PURE__ */ new Set();
	const linters = /* @__PURE__ */ new Set();
	const files = /* @__PURE__ */ new Set();
	const addMatches = (targetPath, source, rule) => {
		for (const result of lint) {
			if (node?.work_dir && result.work_dir && result.work_dir !== node.work_dir) continue;
			if (source && result.linter !== source) continue;
			let matched = false;
			for (const violation of result.violations ?? []) {
				const rawFile = relPath(violation.file, result.work_dir);
				if (rule && (violation.rule?.method || "(no rule)") !== rule) continue;
				if (!rawFile) {
					matched = rule !== void 0 && targetPath === void 0;
					continue;
				}
				if (!(!targetPath || rawFile === targetPath || rawFile.startsWith(`${targetPath}/`))) continue;
				matched = true;
				files.add(rawFile);
			}
			if (matched) {
				if (result.work_dir) workDirs.add(result.work_dir);
				linters.add(result.linter);
			}
		}
	};
	if (!node) for (const result of lint) {
		linters.add(result.linter);
		if (result.work_dir) workDirs.add(result.work_dir);
	}
	else if (node.kind === "linter" || node.kind === "linter-error") {
		if (node.linterName) linters.add(node.linterName);
		collectWorkDirs(node, workDirs);
	} else if (node.kind === "lint-rule" || node.kind === "lint-rule-group") {
		if (node.linterName) linters.add(node.linterName);
		addMatches(node.target_path, node.linterName, node.ruleName);
	} else addMatches(node.target_path, node.linterName, node.ruleName);
	if (node?.work_dir) workDirs.add(node.work_dir);
	return {
		lint: true,
		work_dir: workDirs.size === 1 ? Array.from(workDirs)[0] : "",
		lint_linters: Array.from(linters),
		lint_files: Array.from(files)
	};
}
exports.buildLintRerunBody = buildLintRerunBody;

//# sourceMappingURL=lintRerun.cjs.map