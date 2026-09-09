-- phase: post

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
