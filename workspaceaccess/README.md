# Workspace access

Open **Workspace access…** from Gavel edit's all-files menu or the playground's repository controls. Enter an absolute directory, choose **Review access**, then explicitly confirm the canonical directory and permissions. Gavel edit offers **Grant and open**; the playground refreshes the repository tree after granting access.

Saved grants permit file reads/writes and local tool execution in the directory and its descendants. They are scoped to the OS user and the machine identity in `~/.config/gavel/workspace-machine-id` and stored in Gavel's PostgreSQL database. They apply across local Gavel servers using that database and survive restarts. Database failures deny access; there is no file-based grant registry. Nothing is granted automatically.

Directories supplied on the command line are fixed for that server session. They are not saved automatically, and the UI cannot revoke them. Removing a saved grant blocks subsequent operations but does not undo files already downloaded or stop already-running jobs. Unsaved drafts remain in the browser.

Management requires a loopback caller, a loopback Host, and same-origin browser requests. The playground checks these conditions before proxying and forwards the backend origin only for validated management requests. Hosted dashboards do not expose local workspace management.

The shared API exposes `GET/POST /api/workspace/access`, `POST /api/workspace/access/preview`, `DELETE /api/workspace/access/{id}` and SSE at `GET /api/workspace/access/events`. Grant bodies contain `path` and the reviewed `canonicalPath`; a changed symlink target requires a new preview. `GET/POST /api/workspace/open` is editor-only and returns its opened roots and current file catalog. Saved permissions do not automatically open folders in other editor sessions.
