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

  function ensureStyle() {
    if (document.getElementById(STYLE_ID)) return;
    var style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent =
      ".gavel-rg-dialog{padding:0;border:none;border-radius:10px;width:min(680px,92vw);" +
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
  // component, the page it was grabbed from, and the source/component stack in a
  // code block.
  function buildBody(componentName, source) {
    var parts = [];
    if (componentName) parts.push("**Component:** `<" + componentName + ">`");
    parts.push("**Page:** " + window.location.href);
    if (source) {
      parts.push("");
      parts.push("```");
      parts.push(source);
      parts.push("```");
    }
    return parts.join("\n");
  }

  function openDialog(url) {
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
      if (d && d.source === MSG_SOURCE && (d.type === "todo-created" || d.type === "cancel")) {
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

  function onAction(ctx) {
    var el = ctx.element;
    var title = ctx.componentName || ctx.tagName || "UI element";
    var fileLine = ctx.filePath ? ctx.filePath + ":" + (ctx.lineNumber || "") : "";

    // Dismiss react-grab's UI and UNFREEZE the page immediately. Toggle activation
    // (the default) freezes the page and shows the "Grabbing…" overlay; cleanup()
    // is what unfreezes it. Skipping it leaves the page stuck on "Grabbing…".
    try { if (ctx.hideContextMenu) ctx.hideContextMenu(); } catch (e) { /* noop */ }
    try { if (ctx.cleanup) ctx.cleanup(); } catch (e) { /* noop */ }

    // Open the dialog from the best context we can get without ever blocking:
    // getStackContext is richer but may be slow/hang for some elements, so race
    // it against a short timeout and fall back to the synchronous file:line.
    var done = false;
    var open = function (source) {
      if (done) return;
      done = true;
      openDialog(buildUrl(title, buildBody(ctx.componentName, source || fileLine)));
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

  var plugin = {
    name: PLUGIN_NAME,
    actions: [{ id: "gavel-todo", label: "Add to gavel todo", shortcut: "T", onAction: onAction }],
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
