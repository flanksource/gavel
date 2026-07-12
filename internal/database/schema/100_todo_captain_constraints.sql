-- phase: post
-- runs: always

-- Captain owns the referenced rows. Gavel owns these links and deliberately
-- preserves them: deleting a linked Captain run or plan is rejected until the
-- Gavel link (and any active pointer) is explicitly removed. This makes the
-- deletion policy visible and prevents execution history from disappearing via
-- a Captain session cascade.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.todo_issue_prompt_runs'::regclass
      AND conname = 'todo_issue_prompt_runs_captain_prompt_run_fkey'
  ) THEN
    ALTER TABLE public.todo_issue_prompt_runs
      ADD CONSTRAINT todo_issue_prompt_runs_captain_prompt_run_fkey
      FOREIGN KEY (prompt_run_id)
      REFERENCES public.captain_prompt_runs (id)
      ON UPDATE NO ACTION
      ON DELETE RESTRICT;
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conrelid = 'public.todo_issue_plans'::regclass
      AND conname = 'todo_issue_plans_captain_plan_fkey'
  ) THEN
    ALTER TABLE public.todo_issue_plans
      ADD CONSTRAINT todo_issue_plans_captain_plan_fkey
      FOREIGN KEY (plan_id)
      REFERENCES public.captain_plans (id)
      ON UPDATE NO ACTION
      ON DELETE RESTRICT;
  END IF;
END
$$;

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
