const require_runHistory = require("../runHistory.cjs");
let react = require("react");
// src/hooks/use-run-history.ts
/** Loads the run history page by page from `/api/runs` under the page's base path. */
function useRunHistory() {
	const [pages, setPages] = (0, react.useState)([]);
	const [available, setAvailable] = (0, react.useState)(false);
	const [loading, setLoading] = (0, react.useState)(true);
	const [error, setError] = (0, react.useState)();
	const inFlight = (0, react.useRef)(false);
	const load = (0, react.useCallback)(async (before) => {
		if (inFlight.current) return;
		inFlight.current = true;
		setLoading(true);
		try {
			const page = await require_runHistory.fetchRunsPage(before ? { before } : {});
			if (page) {
				setAvailable(true);
				setPages((prev) => before ? [...prev, page] : [page]);
			}
			setError(void 0);
		} catch (e) {
			setError(e instanceof Error ? e.message : String(e));
		} finally {
			inFlight.current = false;
			setLoading(false);
		}
	}, []);
	(0, react.useEffect)(() => {
		load();
	}, [load]);
	const first = pages[0];
	const next = pages.at(-1)?.next;
	const loadMore = (0, react.useCallback)(() => {
		if (next) load(next);
	}, [next, load]);
	return {
		available,
		...first?.current ? { current: first.current } : {},
		latest: first?.latest ?? [],
		latestHref: first?.latestHref ?? "/",
		runs: pages.flatMap((page) => page.runs),
		hasMore: !!next,
		loading,
		...error ? { error } : {},
		loadMore
	};
}
exports.useRunHistory = useRunHistory;

//# sourceMappingURL=use-run-history.cjs.map