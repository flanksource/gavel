import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiLink } from '@flanksource/clicky-ui/icons';
import { copyText } from '../../clipboard';
import { useCopyFeedback } from '../../useCopyFeedback';

// CopyLinkButton copies a PR's URL to the clipboard. It uses copyText so it
// works inside the native menu-bar WKWebView, where navigator.clipboard is
// denied and the async Clipboard API silently fails.
export function CopyLinkButton({ url }: { url: string }) {
  const { copyState, beginCopy, resetCopyFeedback } = useCopyFeedback();
  const copied = copyState === 'copied';
  const errored = copyState === 'error';

  async function onCopy() {
    beginCopy();
    try {
      await copyText(url);
      resetCopyFeedback('copied');
    } catch {
      resetCopyFeedback('error');
    }
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onCopy}
      className={`h-8 w-8 rounded-md hover:bg-muted hover:text-foreground ${copied ? 'text-green-600' : errored ? 'text-red-600' : 'text-muted-foreground'}`}
      title={copied ? 'Link copied' : errored ? 'Copy failed' : 'Copy link to this PR'}
      aria-label="Copy link to this PR"
    >
      {copied ? <UiCheck className="text-base" /> : <UiLink className="text-base" />}
    </Button>
  );
}
