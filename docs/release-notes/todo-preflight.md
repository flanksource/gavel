# TODO preflight

TODO CLI dry-runs and dashboard previews now validate the complete Captain run input before execution. Capability warnings appear with the preview; invalid attachments, judge declarations, runtime policy, and resolved constraints fail before a run is admitted. Dispatch repeats the same checks against its actual provider and configuration.

Unsupported per-tool policies, including a built-in plan or triage prompt selected with an incompatible runtime, are now rejected during preview. Their policy is preserved; this moves an existing execution failure to the earlier boundary.

Preview remains read-only: it does not start an agent, create a checkout, execute verification, or persist a run. The dashboard shows warnings for the latest request and clears stale warnings when options change or validation fails.
