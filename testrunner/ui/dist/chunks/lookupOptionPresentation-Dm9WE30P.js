import { p as e } from "./JsonSchemaForm-C40jV51n.js";
// ../../../clicky-ui/packages/ui/src/rpc/lookupOptionPresentation.ts
function t(t, r) {
	return r ? {
		...r.icon ? { icon: r.icon } : {},
		...r.description ? { description: r.description } : {},
		...r.tone ? { tone: e(r.tone) } : {},
		...r.group ? { group: r.group } : {},
		...r.badges && r.badges.length > 0 ? { badges: n(t, r.badges) } : {}
	} : {};
}
function n(t, n) {
	return n.map((n, r) => {
		if (typeof n.label != "string" || n.label === "") throw Error(`lookup option "${t}" badge ${r} needs a label, got ${JSON.stringify(n)}`);
		return {
			label: n.label,
			...n.icon ? { icon: n.icon } : {},
			...n.tone ? { tone: e(n.tone) } : {}
		};
	});
}
export { t };
