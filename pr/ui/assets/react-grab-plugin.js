/* gavel · React Grab plugin
 *
 * Registers an "Add to gavel todo" action with React Grab. When invoked on a
 * grabbed element it opens a modal dialog embedding gavel's /todos/new form,
 * prefilled with the component name (title) and its source/component stack
 * (body), and closes the dialog once the form posts back that a todo was
 * created (or cancelled).
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
    } catch (e) {
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

  function buildUrl(title, body) {
    var u = new URL(GAVEL_ORIGIN + "/todos/new");
    u.searchParams.set("embed", "1");
    if (title) u.searchParams.set("title", title);
    if (body) u.searchParams.set("body", body);
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
    label.textContent = "New gavel todo";
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
        } catch (e) {
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
      } else if (d.type === "todo-created" || d.type === "cancel") {
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
    try { if (ctx.hideContextMenu) ctx.hideContextMenu(); } catch (e) { /* noop */ }
    try { if (ctx.cleanup) ctx.cleanup(); } catch (e) { /* noop */ }
  }

  // openTodo opens the new-todo dialog prefilled from the grabbed element,
  // optionally carrying a captured screenshot ({blob,name}) to attach.
  function openTodo(ctx, attachment) {
    var el = ctx.element;
    var title = ctx.componentName || ctx.tagName || "UI element";
    var fileLine = ctx.filePath ? ctx.filePath + ":" + (ctx.lineNumber || "") : "";
    // Snapshot the markup now, before cleanup() unfreezes and may re-render the page.
    var html = grabHtml(el);
    dismissGrab(ctx);

    // Open the dialog from the best context we can get without ever blocking:
    // getStackContext is richer but may be slow/hang for some elements, so race
    // it against a short timeout and fall back to the synchronous file:line.
    var done = false;
    var open = function (source) {
      if (done) return;
      done = true;
      openDialog(buildUrl(title, buildBody(ctx.componentName, source || fileLine, html)), attachment);
    };
    var api = window.__REACT_GRAB__;
    if (api && api.getStackContext) {
      var timer = setTimeout(function () { open(fileLine); }, 1500);
      Promise.resolve()
        .then(function () { return api.getStackContext(el); })
        .then(
          function (source) { clearTimeout(timer); open(source); },
          function () { clearTimeout(timer); open(fileLine); },
        );
    } else {
      open(fileLine);
    }
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

  var plugin = {
    name: PLUGIN_NAME,
    actions: [
      { id: "gavel-todo", label: "Add to gavel todo", shortcut: "T", onAction: onAction },
      { id: "gavel-screenshot", label: "Screenshot to gavel todo", shortcut: "S", onAction: onScreenshot },
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
      } catch (e) {
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
})();
