export const COMMENT_ELEMENT_HTML_LIMIT = 4096;
export const COMMENT_ELEMENT_HTML_TRUNCATION = "\n<!-- …truncated to 4KB -->";

export const COMMENT_SCREENSHOT_UNAVAILABLE_REASONS = [
  "unsupported",
  "cancelled",
  "failed",
] as const;

export type CommentScreenshotUnavailableReason =
  (typeof COMMENT_SCREENSHOT_UNAVAILABLE_REASONS)[number];

export type CommentScreenshot =
  | { status: "captured"; filename: "screenshot.png"; url: string }
  | {
      status: "unavailable";
      reason: CommentScreenshotUnavailableReason;
    };

export type CommentScreenshotCapture =
  | { status: "captured"; dataUrl: string }
  | {
      status: "unavailable";
      reason: CommentScreenshotUnavailableReason;
    };

/** Longest normalised text a text locator or heading may carry. */
export const COMMENT_LOCATOR_TEXT_LIMIT = 200;
/** Longest selector a locator may carry. */
export const COMMENT_LOCATOR_SELECTOR_LIMIT = 1024;

/** A heading that scopes a text locator: the CSS that matches it plus its normalised text. */
export type CommentLocatorHeading = { selector: string; text: string };

/**
 * One independent way to find the commented element again. A comment stores
 * every locator that singled the element out when it was left; resolution
 * tries them all, so one broken selector does not orphan the comment.
 */
export type CommentLocator =
  | { kind: "css"; selector: string }
  | { kind: "role"; tag: string; role: string; name: string }
  | {
      kind: "text";
      tag: string;
      text: string;
      heading?: CommentLocatorHeading;
      /** Selector of the landmark the text is searched within. */
      scope?: string;
    }
  | {
      /** A rendered markdown block, found by its section and its text. */
      kind: "markdown";
      /** Heading texts from the outermost section in; empty before the first heading. */
      path: string[];
      tag: string;
      text: string;
      /** Source line the block started on when captured; breaks ties only. */
      line?: number;
    };

export type CommentElementContext = {
  pageUrl?: string;
  componentName?: string;
  source: string;
  html: string;
  /** Absent on comments left before locators were captured. */
  locators?: CommentLocator[];
  /** Present when screenshot capture was requested. */
  screenshot?: CommentScreenshot;
};

export type CommentElementCaptureContext = Omit<
  CommentElementContext,
  "screenshot"
> & {
  screenshot?: CommentScreenshotCapture;
};
