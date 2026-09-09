-- phase: post

-- Plan paths are optional metadata. This Gavel-owned cross-owner view exposes
-- the durable revision body without copying Captain content into TODO tables.
CREATE OR REPLACE VIEW public.todo_issue_plan_revision_details AS
SELECT
  link.issue_id,
  link.ordinal AS plan_ordinal,
  plan.id AS plan_id,
  COALESCE(plan.id = issue.selected_plan_id, false) AS selected,
  plan.approval_state::text AS approval_state,
  plan.approved_revision_id,
  revision.id AS revision_id,
  revision.revision,
  revision.plan_markdown,
  revision.content_hash,
  revision.feedback,
  revision.created_by,
  revision.created_at,
  COALESCE(revision.id = plan.approved_revision_id, false) AS approved
FROM public.todo_issue_plans AS link
JOIN public.todo_issues AS issue
  ON issue.id = link.issue_id
JOIN public.captain_plans AS plan
  ON plan.id = link.plan_id
JOIN public.captain_plan_revisions AS revision
  ON revision.plan_id = plan.id;

COMMENT ON VIEW public.todo_issue_plan_revision_details IS
  'Gavel issue plan links resolved to Captain-owned durable plan revisions';
