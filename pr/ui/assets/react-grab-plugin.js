/* gavel · React Grab plugin
 *
 * Registers an "Add to gavel todo" action with React Grab. When invoked on a
 * grabbed element it opens a modal dialog embedding gavel's /todos/new form,
 * prefilled with the component name (title), its source/component stack (body),
 * and the grabbed page URL so the form can create a new todo or comment on an
 * existing one in the right project.
 *
 * gavel serves this file with __GAVEL_ORIGIN__ replaced by its own origin, so
 * the iframe + API calls always target the serving gavel server — even when the
 * script is injected (via bookmarklet) into a different app's dev browser. If
 * React Grab isn't present yet (a foreign app), the global build is loaded from
 * unpkg before registering.
 */
(function () {
  "use strict";

  var GAVEL_ORIGIN = "__GAVEL_ORIGIN__";
  var PLUGIN_NAME = "gavel-todo";
  var MSG_SOURCE = "gavel-react-grab";
  var STYLE_ID = "gavel-rg-style";
  var HTML_LIMIT = 2048; // ~2KB of raw outerHTML

  // grabHtml returns up to ~2KB of the grabbed element's raw outerHTML (the
  // element and its children), truncated with a marker so the reader knows the
  // markup was clipped.
  function grabHtml(el) {
    if (!el || typeof el.outerHTML !== "string") return "";
    var html = el.outerHTML;
    if (html.length > HTML_LIMIT) {
      html = html.slice(0, HTML_LIMIT) + "\n<!-- …truncated to 2KB -->";
    }
    return html;
  }

  // captureFrame screenshots the page or — where Region Capture is supported —
  // just the grabbed element. It opens a getDisplayMedia stream (the browser
  // prompts the user to pick a surface; preferCurrentTab offers the current tab
  // first), crops the track to the element via CropTarget when available, paints
  // one frame to a canvas, and resolves a PNG Blob. Resolves null on cancel /
  // unsupported so the todo still opens, just without an image.
  function captureFrame(el) {
    var md = navigator.mediaDevices;
    if (!md || !md.getDisplayMedia) return Promise.resolve(null);
    var stream;
    return md
      .getDisplayMedia({ video: { displaySurface: "browser" }, preferCurrentTab: true, audio: false })
      .then(function (s) {
        stream = s;
        var track = s.getVideoTracks()[0];
        return cropToElement(track, el);
      })
      .then(function () { return frameToBlob(stream); })
      .then(
        function (blob) { stopStream(stream); return blob; },
        function () { stopStream(stream); return null; },
      );
  }

  // cropToElement crops the capture track to el's bounding box via the Region
  // Capture API. It never rejects: where CropTarget is unsupported (non-Chromium)
  // or the shared surface isn't the current tab, it resolves and the full frame is
  // captured — the agreed page fallback.
  function cropToElement(track, el) {
    if (!el || !track || !track.cropTo || typeof CropTarget === "undefined" || !CropTarget.fromElement) {
      return Promise.resolve();
    }
    return CropTarget.fromElement(el)
      .then(function (target) { return track.cropTo(target); })
      .catch(function () { /* keep full frame */ });
  }

  function frameToBlob(stream) {
    var video = document.createElement("video");
    video.muted = true;
    video.srcObject = stream;
    return video.play().then(function () {
      return new Promise(function (resolve) {
        // Two RAFs so cropTo's first cropped frame has reached the track before paint.
        requestAnimationFrame(function () {
          requestAnimationFrame(function () {
            var w = video.videoWidth || 1;
            var h = video.videoHeight || 1;
            var canvas = document.createElement("canvas");
            canvas.width = w;
            canvas.height = h;
            canvas.getContext("2d").drawImage(video, 0, 0, w, h);
            video.pause();
            video.srcObject = null;
            canvas.toBlob(function (blob) { resolve(blob); }, "image/png");
          });
        });
      });
    });
  }

  function stopStream(stream) {
    if (!stream) return;
    try {
      stream.getTracks().forEach(function (t) { t.stop(); });
    } catch {
      /* noop */
    }
  }

  function ensureStyle() {
    if (document.getElementById(STYLE_ID)) return;
    var style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent =
      // Center on the viewport explicitly. A modal <dialog> is centered by the UA
      // via margin:auto, but a host page's margin reset (we're often injected via
      // bookmarklet) zeroes that and pins it to the top — so override with
      // !important to win against the host's styles.
      ".gavel-rg-dialog{position:fixed!important;top:50%!important;left:50%!important;" +
      "transform:translate(-50%,-50%)!important;margin:0!important;" +
      "padding:0;border:none;border-radius:10px;width:min(680px,92vw);" +
      "height:min(640px,88vh);overflow:hidden;box-shadow:0 12px 48px rgba(0,0,0,.35);background:#fff;color:#111}" +
      ".gavel-rg-dialog::backdrop{background:rgba(0,0,0,.45)}" +
      ".gavel-rg-bar{display:flex;align-items:center;justify-content:space-between;padding:8px 12px;" +
      "background:#111;color:#fff;font:600 13px/1.4 system-ui,sans-serif}" +
      ".gavel-rg-bar button{border:none;background:none;color:#fff;font-size:16px;line-height:1;cursor:pointer;padding:2px 6px}" +
      ".gavel-rg-frame{border:none;display:block;width:100%;height:calc(100% - 37px)}";
    document.head.appendChild(style);
  }

  function buildUrl(title, body, sourceUrl) {
    var u = new URL(GAVEL_ORIGIN + "/todos/new");
    u.searchParams.set("embed", "1");
    if (title) u.searchParams.set("title", title);
    if (body) u.searchParams.set("body", body);
    if (sourceUrl) u.searchParams.set("sourceUrl", sourceUrl);
    return u.toString();
  }

  // buildBody renders the grabbed element as Markdown for the todo body: the
  // component, the page it was grabbed from, the source/component stack, and up
  // to 2KB of the element's raw HTML — each in its own code block.
  function buildBody(componentName, source, html) {
    var parts = [];
    if (componentName) parts.push("**Component:** `<" + componentName + ">`");
    parts.push("**Page:** " + window.location.href);
    if (source) {
      parts.push("");
      parts.push("```");
      parts.push(source);
      parts.push("```");
    }
    if (html) {
      parts.push("");
      parts.push("**HTML:**");
      parts.push("```html");
      parts.push(html);
      parts.push("```");
    }
    return parts.join("\n");
  }

  function fileLine(filePath, lineNumber) {
    if (!filePath) return "";
    return filePath + ":" + (lineNumber || "");
  }

  function resolveWithTimeout(promise, fallback, timeoutMs) {
    return new Promise(function (resolve) {
      var done = false;
      var timer = setTimeout(function () {
        if (done) return;
        done = true;
        resolve(fallback);
      }, timeoutMs);
      Promise.resolve(promise).then(
        function (value) {
          if (done) return;
          done = true;
          clearTimeout(timer);
          resolve(value || fallback);
        },
        function () {
          if (done) return;
          done = true;
          clearTimeout(timer);
          resolve(fallback);
        },
      );
    });
  }

  function displayNameForElement(el, fallback) {
    var api = window.__REACT_GRAB__;
    if (fallback) return fallback;
    try {
      if (api && api.getDisplayName) return api.getDisplayName(el) || "";
    } catch {
      /* noop */
    }
    return "";
  }

  function fallbackSourceForElement(el, fallback) {
    if (fallback) return Promise.resolve(fallback);
    var api = window.__REACT_GRAB__;
    if (!api || !api.getSource) return Promise.resolve("");
    return Promise.resolve()
      .then(function () { return api.getSource(el); })
      .then(function (source) { return source ? fileLine(source.filePath, source.lineNumber) : ""; })
      .catch(function () { return ""; });
  }

  function stackSourceForElement(el, fallback) {
    var api = window.__REACT_GRAB__;
    if (!api || !api.getStackContext) return Promise.resolve(fallback);
    return resolveWithTimeout(
      Promise.resolve().then(function () { return api.getStackContext(el); }),
      fallback,
      1500,
    );
  }

  // todoBodyForElement is the single capture formatter used by both the gavel
  // todo dialog and React Grab's built-in Copy command, so clipboard content and
  // created/commented todos stay identical.
  function todoBodyForElement(opts) {
    var el = opts.element;
    var html = opts.html != null ? opts.html : grabHtml(el);
    var componentName = displayNameForElement(el, opts.componentName);
    var fallback = fileLine(opts.filePath, opts.lineNumber);
    return fallbackSourceForElement(el, fallback)
      .then(function (sourceFallback) {
        return stackSourceForElement(el, sourceFallback);
      })
      .then(function (source) {
        return buildBody(componentName, source, html);
      });
  }

  function todoBodiesForElements(elements) {
    var selected = Array.isArray(elements) ? elements.filter(Boolean) : [];
    if (selected.length === 0) return Promise.resolve("");
    return Promise.all(selected.map(function (el) {
      return todoBodyForElement({ element: el });
    })).then(function (parts) {
      return parts.filter(Boolean).join("\n\n---\n\n");
    });
  }

  function todoBodyForContext(ctx) {
    var selected = Array.isArray(ctx.elements) ? ctx.elements.filter(Boolean) : [];
    if (selected.length > 1) return todoBodiesForElements(selected);
    return todoBodyForElement({
      element: ctx.element,
      componentName: ctx.componentName,
      filePath: ctx.filePath,
      lineNumber: ctx.lineNumber,
      html: grabHtml(ctx.element),
    });
  }

  function absoluteGavelURL(pathOrURL) {
    if (!pathOrURL) return "";
    try {
      return new URL(pathOrURL, GAVEL_ORIGIN).toString();
    } catch {
      return pathOrURL;
    }
  }

  function uploadScreenshot(blob) {
    if (!blob) return Promise.reject(new Error("screenshot capture canceled"));
    var form = new FormData();
    form.append("attachment", blob, "screenshot.png");
    return fetch(GAVEL_ORIGIN + "/api/todos/attachments", { method: "POST", body: form })
      .then(function (res) {
        return res.json().catch(function () { return {}; }).then(function (data) {
          if (!res.ok) throw new Error(data.error || "screenshot upload failed");
          var attachment = data.attachments && data.attachments[0];
          if (!attachment || !attachment.url) throw new Error("screenshot upload returned no URL");
          attachment.url = absoluteGavelURL(attachment.url);
          return attachment;
        });
      });
  }

  function screenshotMarkdown(attachment) {
    var filename = attachment.filename || "screenshot.png";
    return "![" + filename + "](" + attachment.url + ")";
  }

  function todoBodyWithScreenshot(body, attachment) {
    var image = screenshotMarkdown(attachment);
    body = (body || "").trim();
    if (!body) return image;
    return body + "\n\n## Attachments\n\n" + image;
  }

  function fallbackCopyText(text) {
    return new Promise(function (resolve, reject) {
      var textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.setAttribute("readonly", "readonly");
      textarea.style.position = "fixed";
      textarea.style.left = "-9999px";
      document.body.appendChild(textarea);
      textarea.select();
      try {
        if (!document.execCommand("copy")) throw new Error("copy failed");
        resolve(true);
      } catch (err) {
        reject(err);
      } finally {
        textarea.remove();
      }
    });
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(
        function () { return true; },
        function () { return fallbackCopyText(text); },
      );
    }
    return fallbackCopyText(text);
  }

  // openDialog shows the todo form in a modal iframe. When attachment ({blob,name})
  // is set it ships the blob into the iframe once the form signals "embed-ready" —
  // the form (same-origin to gavel) then uploads it as multipart, so no CORS is
  // needed even when this plugin runs injected into a foreign app.
  function openDialog(url, attachment) {
    ensureStyle();
    var dlg = document.createElement("dialog");
    dlg.className = "gavel-rg-dialog";

    var bar = document.createElement("div");
    bar.className = "gavel-rg-bar";
    var label = document.createElement("span");
    label.textContent = "Gavel issue";
    var close = document.createElement("button");
    close.type = "button";
    close.setAttribute("aria-label", "Close");
    close.textContent = "✕";
    bar.appendChild(label);
    bar.appendChild(close);

    var frame = document.createElement("iframe");
    frame.className = "gavel-rg-frame";
    frame.src = url;

    dlg.appendChild(bar);
    dlg.appendChild(frame);
    document.body.appendChild(dlg);

    function teardown() {
      window.removeEventListener("message", onMessage);
      if (dlg.open) {
        try {
          dlg.close();
        } catch {
          /* already closing */
        }
      }
      dlg.remove();
    }
    function onMessage(e) {
      if (e.source !== frame.contentWindow) return;
      var d = e.data;
      if (!d || d.source !== MSG_SOURCE) return;
      if (d.type === "embed-ready") {
        if (attachment && attachment.blob && frame.contentWindow) {
          frame.contentWindow.postMessage(
            { source: MSG_SOURCE, type: "attachment", blob: attachment.blob, name: attachment.name },
            GAVEL_ORIGIN,
          );
        }
      } else if (d.type === "todo-created" || d.type === "todo-commented" || d.type === "cancel") {
        teardown();
      }
    }

    close.addEventListener("click", teardown);
    dlg.addEventListener("cancel", function (e) {
      e.preventDefault();
      teardown();
    });
    window.addEventListener("message", onMessage);
    dlg.showModal();
  }

  // dismissGrab removes react-grab's overlay and UNFREEZES the page. Toggle
  // activation (the default) freezes the page and shows the "Grabbing…" overlay;
  // cleanup() is what unfreezes it. Skipping it leaves the page stuck on "Grabbing…".
  function dismissGrab(ctx) {
    try { if (ctx.hideContextMenu) ctx.hideContextMenu(); } catch { /* noop */ }
    try { if (ctx.cleanup) ctx.cleanup(); } catch { /* noop */ }
  }

  // openTodo opens the new-todo dialog prefilled from the grabbed element,
  // optionally carrying a captured screenshot ({blob,name}) to attach.
  function openTodo(ctx, attachment) {
    var el = ctx.element;
    var title = ctx.componentName || ctx.tagName || "UI element";
    // Snapshot the markup now, before cleanup() unfreezes and may re-render the page.
    var html = grabHtml(el);
    var sourceUrl = window.location.href;
    var body = todoBodyForElement({
      element: el,
      componentName: ctx.componentName,
      filePath: ctx.filePath,
      lineNumber: ctx.lineNumber,
      html: html,
    });
    dismissGrab(ctx);

    // Open the dialog from the best context we can get without ever blocking:
    // getStackContext is richer but may be slow/hang for some elements, so race
    // it against a short timeout and fall back to the synchronous file:line.
    body.then(function (markdown) {
      openDialog(buildUrl(title, markdown, sourceUrl), attachment);
    });
  }

  function onAction(ctx) {
    openTodo(ctx, null);
  }

  // onScreenshot dismisses the grab overlay first so it isn't in the shot, then
  // captures a frame (still within the action's user gesture, required by
  // getDisplayMedia) and opens the todo with the screenshot attached.
  function onScreenshot(ctx) {
    dismissGrab(ctx);
    captureFrame(ctx.element).then(function (blob) {
      openTodo(ctx, blob ? { blob: blob, name: "screenshot.png" } : null);
    });
  }

  function onCopyContent(content, elements) {
    return todoBodiesForElements(elements).then(function (markdown) {
      return markdown || content;
    });
  }

  function copyScreenshot(ctx, bodyMode) {
    var body = bodyMode ? todoBodyForContext(ctx) : Promise.resolve("");
    var el = ctx.element;
    dismissGrab(ctx);
    return Promise.all([
      body,
      captureFrame(el).then(uploadScreenshot),
    ]).then(function (result) {
      var markdown = result[0];
      var attachment = result[1];
      var text = bodyMode ? todoBodyWithScreenshot(markdown, attachment) : attachment.url;
      return copyText(text);
    }).catch(function (err) {
      console.warn("[gavel] copy screenshot failed:", err);
      return false;
    });
  }

  function onCopyScreenshot(ctx) {
    return copyScreenshot(ctx, true);
  }

  function onScreenshotOnly(ctx) {
    return copyScreenshot(ctx, false);
  }

  var plugin = {
    name: PLUGIN_NAME,
    hooks: {
      transformCopyContent: onCopyContent,
    },
    actions: [
      { id: "gavel-todo", label: "Add to gavel todo", shortcut: "T", onAction: onAction },
      { id: "gavel-screenshot", label: "Screenshot to gavel todo", shortcut: "S", onAction: onScreenshot },
      { id: "gavel-copy-screenshot", label: "Copy screenshot", shortcut: "C", onAction: onCopyScreenshot },
      { id: "gavel-screenshot-only", label: "Screenshot only", shortcut: "O", onAction: onScreenshotOnly },
    ],
  };

  var registered = false;
  function register() {
    if (registered) return true;
    var api = window.__REACT_GRAB__;
    if (api && typeof api.registerPlugin === "function") {
      api.registerPlugin(plugin);
      registered = true;
      // Make "Add to gavel todo" the default action so the primary grab gesture
      // (select) runs it instead of copy.
      try {
        if (api.setToolbarState) api.setToolbarState({ defaultAction: "gavel-todo" });
      } catch {
        /* noop */
      }
      return true;
    }
    return false;
  }

  function loadReactGrab() {
    if (document.getElementById("gavel-rg-loader")) return;
    var s = document.createElement("script");
    s.id = "gavel-rg-loader";
    s.crossOrigin = "anonymous";
    s.src = "https://unpkg.com/react-grab/dist/index.global.js";
    document.body.appendChild(s);
  }

  if (!register()) {
    var tries = 0;
    var loaderKicked = false;
    var timer = setInterval(function () {
      tries++;
      if (register() || tries > 100) {
        clearInterval(timer);
        return;
      }
      if (tries === 5 && !loaderKicked) {
        loaderKicked = true;
        loadReactGrab();
      }
    }, 100);
  }

  // React Grab and this plugin are per-window: an iframe has its own document and
  // window, so the parent's instance can't grab inside it. injectIntoFrame loads
  // this same script into each SAME-ORIGIN iframe, where it self-bootstraps (no
  // __REACT_GRAB__ there → loadReactGrab() pulls react-grab from unpkg → registers
  // the gavel action locally). Nested iframes are handled recursively, per-window.
  function injectIntoFrame(frame) {
    if (frame.classList && frame.classList.contains("gavel-rg-frame")) return; // our own todo modal
    var doc;
    try {
      doc = frame.contentDocument; // cross-origin / sandboxed-without-allow-same-origin throws or null
    } catch {
      return;
    }
    if (!doc || doc.getElementById("gavel-rg-plugin-script")) return;
    var mount = doc.head || doc.body || doc.documentElement;
    if (!mount) return; // mid-parse/empty doc — the load listener retries
    var s = doc.createElement("script");
    s.id = "gavel-rg-plugin-script";
    s.src = GAVEL_ORIGIN + "/react-grab-plugin.js";
    mount.appendChild(s);
  }

  // watchFrame injects now and again on load — navigation replaces the iframe's
  // document (dropping gavel-rg-plugin-script), so re-injecting is correct. The
  // one-time element flag keeps the load listener from stacking across re-scans.
  function watchFrame(frame) {
    if (frame.__gavelRgWatched) return;
    frame.__gavelRgWatched = true;
    injectIntoFrame(frame);
    frame.addEventListener("load", function () {
      injectIntoFrame(frame);
    });
  }

  function scanFrames(root) {
    var frames = (root || document).querySelectorAll("iframe");
    for (var i = 0; i < frames.length; i++) watchFrame(frames[i]);
  }

  // queueScan coalesces a full re-scan into one frame. React Grab's overlay mutates
  // the DOM constantly, so re-scanning on every MutationObserver record would storm.
  var scanQueued = false;
  function queueScan() {
    if (scanQueued) return;
    scanQueued = true;
    var run = function () {
      scanQueued = false;
      scanFrames(document);
    };
    if (typeof requestAnimationFrame === "function") requestAnimationFrame(run);
    else setTimeout(run, 0);
  }

  function observeFrames() {
    if (typeof MutationObserver === "undefined") return;
    new MutationObserver(function (mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var added = mutations[i].addedNodes;
        for (var j = 0; j < added.length; j++) {
          var node = added[j];
          if (node.nodeType !== 1) continue;
          if (node.tagName === "IFRAME") watchFrame(node);
          else if (node.querySelector && node.querySelector("iframe")) queueScan();
        }
      }
    }).observe(document.documentElement, { childList: true, subtree: true });
  }

  function bootstrapFrames() {
    scanFrames(document);
    observeFrames();
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootstrapFrames);
  } else {
    bootstrapFrames();
  }
})();
