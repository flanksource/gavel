import { useEffect } from "react";
import { sendToHost } from "./useHostBridge";

export function useHostDialog(visible = true, notify = sendToHost) {
  useEffect(() => {
    if (!visible) return;
    let previous = false;
    const update = () => {
      const open = !!document.querySelector('[aria-modal="true"]');
      if (open !== previous) {
        previous = open;
        notify({ type: "dialog", open });
      }
    };
    const observer = new MutationObserver(update);
    observer.observe(document.body, { subtree: true, childList: true });
    update();
    return () => {
      observer.disconnect();
      if (previous) notify({ type: "dialog", open: false });
    };
  }, [visible, notify]);
}
