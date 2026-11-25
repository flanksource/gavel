import { DOCUMENT_ANCHOR } from "@flanksource/clicky-ui";
import type { Selection } from "../host/model";

/**
 * The page snapshot a whole-page comment is filed with: `request` asks the host
 * for the page, and the panel hands each selection the host reports to `claim`,
 * which takes the whole-page one so it posts from the footer instead of opening
 * a floating draft.
 */
export function pageSnapshot({ requestPage, timeoutMs }: { requestPage: () => void; timeoutMs: number }) {
  let waiter: ((selection: Selection) => void) | null = null;
  return {
    claim(selection: Selection): boolean {
      if (!waiter || selection.anchor !== DOCUMENT_ANCHOR) return false;
      waiter(selection);
      return true;
    },
    request(): Promise<Selection> {
      if (waiter) return Promise.reject(new Error("A whole-page comment is already being posted"));
      return new Promise<Selection>((resolve, reject) => {
        const timer = setTimeout(() => {
          waiter = null;
          reject(new Error("The page did not report its snapshot, so the comment was not posted"));
        }, timeoutMs);
        waiter = next => { clearTimeout(timer); waiter = null; resolve(next); };
        requestPage();
      });
    },
  };
}
