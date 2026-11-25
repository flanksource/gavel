import type { CommentScreenshotCapture } from "../contracts/comments-model";
import { viewportCrop, type HostViewportRect, type ScreenshotTarget } from "../comments/screenshot-viewport";
import { CAPTURE_REQUEST_EVENT, CAPTURE_RESULT_EVENT, type CaptureRequestDetail, type CaptureResultDetail } from "./messages";

const BRIDGE_TIMEOUT_MS = 10_000;

/** Asks the extension (through the isolated-world bridge, which asks the background worker) for a PNG data URL of the visible tab. */
export function requestTabImage(): Promise<string> {
  const id = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    const finish = () => { clearTimeout(timeout); window.removeEventListener(CAPTURE_RESULT_EVENT, onResult); };
    const onResult = (event: Event) => {
      const result = (event as CustomEvent<CaptureResultDetail>).detail;
      if (result.id !== id) return;
      finish();
      if (result.ok) resolve(result.dataUrl);
      else reject(new Error(result.error));
    };
    const timeout = setTimeout(() => { finish(); reject(new Error("the grabber's extension bridge did not answer the screenshot request")); }, BRIDGE_TIMEOUT_MS);
    window.addEventListener(CAPTURE_RESULT_EVENT, onResult);
    window.dispatchEvent(new CustomEvent<CaptureRequestDetail>(CAPTURE_REQUEST_EVENT, { detail: { id } }));
  });
}

function loadImage(dataUrl: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error("the captured tab image could not be decoded"));
    image.src = dataUrl;
  });
}

function cropToDataUrl(image: HTMLImageElement, crop: HostViewportRect): string {
  const canvas = document.createElement("canvas");
  canvas.width = crop.width;
  canvas.height = crop.height;
  const context = canvas.getContext("2d");
  if (!context) throw new Error("browser could not create a screenshot canvas");
  context.drawImage(image, crop.x, crop.y, crop.width, crop.height, 0, 0, crop.width, crop.height);
  return canvas.toDataURL("image/png");
}

/**
 * The host's `capture` inside the extension: the visible tab as the extension
 * sees it, cropped to the target (the image is `devicePixelRatio` times the
 * viewport, which `viewportCrop` derives from the image size). A failure is
 * logged and recorded as `failed`, never turned into an empty screenshot.
 */
export async function extensionCapture(target: Element | ScreenshotTarget): Promise<CommentScreenshotCapture> {
  try {
    const image = await loadImage(await requestTabImage());
    const crop = viewportCrop({ target, frame: { width: image.naturalWidth, height: image.naturalHeight }, viewport: { width: window.innerWidth, height: window.innerHeight } });
    return { status: "captured", dataUrl: cropToDataUrl(image, crop) };
  } catch (cause) {
    console.error("[gavel] Screenshot failed:", cause);
    return { status: "unavailable", reason: "failed" };
  }
}
