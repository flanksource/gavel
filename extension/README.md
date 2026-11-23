# Gavel Chrome extension

`@flanksource/gavel-extension` lives in Gavel's `pr/ui` pnpm workspace. It connects directly to a Gavel dashboard; the Clicky UI playground is not a backend.

From the Gavel repository:

```sh
pnpm --dir pr/ui install
pnpm --dir pr/ui --filter @flanksource/gavel-extension build
```

Open `chrome://extensions`, enable Developer mode, choose Load unpacked, and select `extension/dist`. Open the toolbar popup on the page you want to comment on. Choose a project, then switch Gavel on for the site; the first project you choose becomes the default for every site, and a different project chosen later applies to this site only. A project belongs to its dashboard, so a site on its own dashboard never uses the default dashboard's project and chooses its own. Under **Comments from this site go to**, pick **This site's agent**, the **Global session** or **Automatic**, where Gavel starts sessions itself (for each page or one for all pages). For an agent or the global session, press **Start session** (Gavel names it `<name>-<6 characters>`, starting from `gavel-<host>` or `gavel-global`), choose a running session with **Switch** or **use a running session**, or press **New session** to replace the current one. The choice is stored as `target` in the page's `flanksource.grabber` entry (`page` is the default and is not stored, `global` and `auto` are), the site's own session as `session` in the same entry, and the global session in the extension's global defaults. Every control applies at once; there are no save buttons. **Settings** in the popup header holds the Gavel dashboard address (default `http://localhost:9092`, committed on Enter or when the field loses focus) and the default model for new sessions. A value stored for this site only shows a **This site** menu, where you can use the default here again or make the value the default for all sites. If Chrome withholds access to the dashboard, the popup says so and **Allow access** asks for it. The footer shows the sync status of captures and comments. The same card is under the panel's settings gear. Enabling the site shows the element picker and comment panel. Rebuild and reload the extension after changing its source.

The page grabber's menu captures a todo, and a sync error in the popup footer opens the saved captures; the bundled form opens even when Gavel is stopped. Projects, labels, existing todos, lifecycle steps, presets and runtime catalogs are cached per dashboard and project in IndexedDB. Load them once while Gavel is available. Captures, screenshots and requested runs are committed to IndexedDB before the form confirms success, then the background worker retries delivery on startup and every minute. A failed run can be retried against its already-created todo. An interrupted request with an unknown outcome requires review before retrying to avoid duplicates. Offline run selection requires a previously cached runtime catalog.

The form includes the project picker, labels, priority, status, description, screenshots, and Start new run with the shared runtime bar. The page panel also supports comments, assignment, managed sessions and the session inspector. Page frames are validated against their parent origin and the background worker's recorded activation.

Site settings use the page's `flanksource.grabber` localStorage entry. Previous Gavel entries are read and rewritten without the backend discriminator; playground entries must be discarded in the popup. Chrome associates an unpacked extension's IndexedDB with its extension ID: moving the install directory can create a new ID. Sync saved captures in the previous installation before replacing it, and keep that installation available until its queue is empty.

The page composer previews where Gavel will deliver the comment. Its **Runs on** menu changes that draft's destination only: the default (the site's page or global session, or a session Gavel starts and routes when the site is on Automatic), the global session, this page's session, any other active session, a new session with a model, or an `@` target. The Session tab and Assign All use the site's default destination, which is the route chosen in the settings card: this site's agent, the global session or Automatic.

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
