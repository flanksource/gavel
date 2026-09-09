// copyText writes text to the clipboard, falling back to a hidden-textarea
// execCommand('copy') when the async Clipboard API is unavailable or rejects —
// e.g. inside the native menu-bar WKWebView, where navigator.clipboard.writeText
// is denied. Rejects when both paths fail so callers can surface the failure
// instead of a copy button that silently does nothing.
export async function copyText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch {
      // Fall through to the user-gesture copy path below.
    }
  }
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  let copied = false;
  try {
    copied = document.execCommand('copy');
  } finally {
    document.body.removeChild(textarea);
  }
  if (!copied) throw new Error('Browser rejected the copy request');
}
