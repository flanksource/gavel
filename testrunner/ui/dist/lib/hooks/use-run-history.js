import { fetchRunsPage } from "../runHistory.js";
import { useCallback, useEffect, useRef, useState } from "react";
// src/hooks/use-run-history.ts
/** Loads the run history page by page from `/api/runs` under the page's base path. */
function useRunHistory() {
	const [pages, setPages] = useState([]);
	const [available, setAvailable] = useState(false);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState();
	const inFlight = useRef(false);
	const load = useCallback(async (before) => {
		if (inFlight.current) return;
		inFlight.current = true;
		setLoading(true);
		try {
			const page = await fetchRunsPage(before ? { before } : {});
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
	useEffect(() => {
		load();
	}, [load]);
	const first = pages[0];
	const next = pages.at(-1)?.next;
	const loadMore = useCallback(() => {
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
export { useRunHistory };

//# sourceMappingURL=use-run-history.js.map