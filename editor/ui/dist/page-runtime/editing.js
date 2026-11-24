import { createContext, createElement, useContext, useSyncExternalStore } from "react";
// src/editing.ts
const scope = globalThis;
const { removed, listeners } = scope.__pageEditing ??= {
	removed: /* @__PURE__ */ new Set(),
	listeners: /* @__PURE__ */ new Set()
};
const subscribe = (listener) => {
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
};
const reviewScope = createContext("");
function EditingScope({ scope, children }) {
	return createElement(reviewScope.Provider, { value: scope }, children);
}
function setRemoved(identity, value) {
	if (value) removed.add(identity);
	else removed.delete(identity);
	for (const listener of listeners) listener();
}
function EditableBoundary({ identities, children }) {
	const scope = useContext(reviewScope);
	return useSyncExternalStore(subscribe, () => identities.some((identity) => removed.has(scope && identity.startsWith("review:") ? `review:${scope}:${identity.slice(7)}` : identity)), () => false) ? null : children;
}
function editableElement(element) {
	const review = element.props["data-review-id"];
	const source = element.props["data-source"];
	const identities = [...typeof review === "string" ? [`review:${review}`] : [], ...typeof source === "string" && source.includes("#") ? [`source:${source}`] : []];
	return identities.length ? createElement(EditableBoundary, {
		identities,
		key: element.key,
		children: element
	}) : element;
}
export { EditingScope, editableElement, setRemoved };
