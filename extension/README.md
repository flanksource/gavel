# Gavel Chrome extension

`@flanksource/gavel-extension` lives in Gavel's `pr/ui` pnpm workspace. It connects directly to a Gavel dashboard; the Clicky UI playground is not a backend.

From the Gavel repository:

```sh
pnpm --dir pr/ui install
pnpm --dir pr/ui --filter @flanksource/gavel-extension build
```

Open `chrome://extensions`, enable Developer mode, choose Load unpacked, and select `extension/dist`. Open the toolbar popup, set the Gavel dashboard address (default `http://localhost:9092`), and choose a project for page comments. Global defaults apply to every site; This site saves only that site's overrides. Enable the site to show the element picker and comment panel. Rebuild and reload the extension after changing its source.

Capture todo / saved captures opens the bundled form even when Gavel is stopped. Projects, labels, existing todos, lifecycle steps, presets and runtime catalogs are cached per dashboard and project in IndexedDB. Load them once while Gavel is available. Captures, screenshots and requested runs are committed to IndexedDB before the form confirms success, then the background worker retries delivery on startup and every minute. A failed run can be retried against its already-created todo. An interrupted request with an unknown outcome requires review before retrying to avoid duplicates. Offline run selection requires a previously cached runtime catalog.

The form includes the project picker, labels, priority, status, description, screenshots, and Start new run with the shared runtime bar. The page panel also supports comments, assignment, managed sessions and the session inspector. Page frames are validated against their parent origin and the background worker's recorded activation.

Site settings use the page's `flanksource.grabber` localStorage entry. Previous Gavel entries are read and rewritten without the backend discriminator; playground entries must be discarded in the popup. Chrome associates an unpacked extension's IndexedDB with its extension ID: moving the install directory can create a new ID. Sync saved captures in the previous installation before replacing it, and keep that installation available until its queue is empty.

## Development

Normal builds resolve installed packages without sibling source aliases. The migration needs a Clicky UI release containing the current comment/session APIs and the React picker at `@flanksource/clicky-ui/react-grab`. The picker ships within Clicky UI. Clicky UI 0.3.59 does not contain those APIs, and an unpublished dependency cannot support a clean install. Until that release exists, dependency installation, lockfile generation and the normal Make targets remain blocked; the explicit development checks below validate the local source only. Build the local picker with `pnpm --filter @flanksource/clicky-ui build:react-grab` in the Clicky UI checkout before development checks.

With a built sibling `clicky-ui` checkout and development dependencies installed, use explicit development mode:

```sh
pnpm --dir pr/ui --filter @flanksource/gavel-extension build:dev
pnpm --dir pr/ui --filter @flanksource/gavel-extension check:dev
pnpm --dir pr/ui --filter @flanksource/gavel-extension test --mode development
pnpm --dir pr/ui --filter @flanksource/gavel-extension lint
```

The extension owns its popup, page bridge, Gavel adapters, TODO queue and build. Generic components and the shared page agent remain dependencies of Clicky UI and page-runtime. No playground server, middleware or source path is imported by the extension.
