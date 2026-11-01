import { useCallback, useEffect, useRef, useState } from "react";
import { AnsiHtml, Badge } from "@flanksource/clicky-ui/data";
import { jsx, jsxs } from "react/jsx-runtime";
import { Modal } from "@flanksource/clicky-ui/components";
// src/components/RerunDialog.tsx
var MAX_CONNECT_ATTEMPTS = 20;
var RETRY_MS = 250;
var STATUS_TONE = {
	running: "info",
	success: "success",
	failed: "danger",
	canceled: "warning"
};
function lineText(line) {
	return line.stream === "stderr" ? `\x1b[31m${line.text}\x1b[0m` : line.text;
}
/**
* The output of the rerun in progress, streamed from the server. The server
* sends each write once, and replays them all to a client that reconnects, so
* every (re)connection starts the output over.
*/
function RerunDialog({ open, onClose, streamUrl }) {
	const [lines, setLines] = useState([]);
	const [status, setStatus] = useState("running");
	const [command, setCommand] = useState("");
	const [error, setError] = useState(void 0);
	const scrollRef = useRef(null);
	const autoScrollRef = useRef(true);
	const statusRef = useRef("running");
	const handleScroll = useCallback(() => {
		const el = scrollRef.current;
		if (!el) return;
		autoScrollRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
	}, []);
	const [attempt, setAttempt] = useState(0);
	const retryTimer = useRef(null);
	useEffect(() => {
		if (!open) return;
		setStatus("running");
		statusRef.current = "running";
		setCommand("");
		setError(void 0);
		setAttempt(0);
		autoScrollRef.current = true;
	}, [open, streamUrl]);
	useEffect(() => {
		if (!open) return;
		const es = new EventSource(streamUrl);
		es.addEventListener("open", () => setLines([]));
		es.addEventListener("message", (e) => {
			const snap = JSON.parse(e.data);
			if (snap.command) setCommand(snap.command);
			statusRef.current = snap.status;
			setStatus(snap.status);
			if (snap.lines?.length) setLines((prev) => [...prev, ...snap.lines]);
		});
		es.addEventListener("done", () => es.close());
		es.onerror = () => {
			if (statusRef.current !== "running") {
				es.close();
				return;
			}
			if (es.readyState !== EventSource.CLOSED) return;
			if (attempt >= MAX_CONNECT_ATTEMPTS) {
				setError("The rerun output is not available");
				return;
			}
			retryTimer.current = window.setTimeout(() => setAttempt((n) => n + 1), RETRY_MS);
		};
		return () => {
			es.close();
			if (retryTimer.current !== null) window.clearTimeout(retryTimer.current);
		};
	}, [
		open,
		streamUrl,
		attempt
	]);
	useEffect(() => {
		if (autoScrollRef.current && scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
	}, [lines]);
	return /* @__PURE__ */ jsx(Modal, {
		open,
		onClose,
		size: "xl",
		closeOnBackdrop: true,
		scrollBody: false,
		title: /* @__PURE__ */ jsx("span", {
			className: "truncate font-mono text-sm",
			title: command || "rerun",
			children: command || "rerun"
		}),
		headerSlot: /* @__PURE__ */ jsx("span", {
			"data-rerun-status": status,
			children: /* @__PURE__ */ jsx(Badge, {
				tone: STATUS_TONE[status],
				variant: "soft",
				size: "xs",
				children: status
			})
		}),
		children: /* @__PURE__ */ jsxs("pre", {
			ref: scrollRef,
			onScroll: handleScroll,
			className: "h-[70vh] overflow-auto bg-zinc-950 p-4 font-mono text-sm leading-relaxed text-zinc-100",
			"data-rerun-output": "",
			children: [
				lines.length === 0 && status === "running" && !error && /* @__PURE__ */ jsx("span", {
					className: "text-zinc-500",
					children: "Waiting for output..."
				}),
				lines.length === 0 && status !== "running" && !error && /* @__PURE__ */ jsx("span", {
					className: "text-zinc-500",
					children: "The rerun wrote no output."
				}),
				error && /* @__PURE__ */ jsx("span", {
					className: "text-red-400",
					children: error
				}),
				/* @__PURE__ */ jsx(AnsiHtml, {
					as: "span",
					text: lines.map(lineText).join("")
				})
			]
		})
	});
}
export { RerunDialog };

//# sourceMappingURL=RerunDialog.js.map