// The entry of the extension's `grabber.js`, injected into the page's own world:
// the shared grabber host plus the screenshot capture and upload only the extension has.
export { installBookmarklet } from "../host/host";
export { extensionCapture } from "./capture-client";
export { extensionUpload } from "./upload-client";
