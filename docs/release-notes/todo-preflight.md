# TODO preflight

TODO CLI dry-runs and dashboard previews now validate the complete Captain run input before execution. Capability warnings appear with the preview; invalid attachments, judge declarations, runtime policy, and resolved constraints fail before a run is admitted. Dispatch repeats the same checks against its actual provider and configuration.

Unsupported per-tool policies, including a built-in plan or triage prompt selected with an incompatible runtime, are now rejected during preview. Their policy is preserved; this moves an existing execution failure to the earlier boundary.

Preview remains read-only: it does not start an agent, create a checkout, execute verification, or persist a run. The dashboard shows warnings for the latest request and clears stale warnings when options change or validation fails.

Saved Captain settings now fill unresolved runtime fields across TODOs, named prompts, AI fixes, commit/PR content, status and outline summaries, and fixture grading. Authored settings retain precedence, including explicit false, zero and empty values. Complete resolved specifications reach provider creation and execution, and field provenance identifies the actual source. An authoritative fixture specification is validated without applying ambient saved defaults.

Generating requests without a model now fail with configuration guidance. A named model with no configured mode warns during the compatibility window. Run controls no longer select the first catalog model or inject defaults into otherwise empty requests. Invalid saved configuration is reported as a configuration error even when an explicit request supplies the same field.
