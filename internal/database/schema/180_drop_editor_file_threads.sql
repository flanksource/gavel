-- phase: post

-- `gavel edit` threads used to live in their own tenant, scoped
-- gavel://edit/file/<sha256 of the path>; they now live under the project
-- holding the file (gavel://project/<name>/files/<path>). That tenant held only
-- those threads, which cannot be mapped back to a path, so it is dropped.
-- Children first: the comment foreign keys do not cascade.
DELETE FROM comment_reads WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_attachments WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_blobs WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_operations WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_messages WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_events WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_bindings WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_lint_runs WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comment_sequences WHERE tenant_id = 'gavel-local-editor';
DELETE FROM comments WHERE tenant_id = 'gavel-local-editor';
