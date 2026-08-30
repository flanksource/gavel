-- phase: post
-- dependsOn: 115_backfill_todo_activity.sql, 102_todo_prompt_run_step_kind.sql

-- One-time reclassification for triage runs recorded before triage became its
-- own step kind.
--
-- Triage has always been a plan-CLASS run (todos/prompt/catalog.go declares it
-- Class: ModePlan), so every historical triage pass was linked as
-- step_kind = 'plan' and is distinguishable only by the prompt name Captain
-- recorded on the run itself (spec_profile). Left alone, every triage a project
-- has ever run would report as a planning pass.
--
-- Safe to move: the link's unique index is (issue_id, step_kind, ordinal), so
-- lifting a row out of the 'plan' sequence into 'triage' cannot collide with a
-- row already there. It does leave both sequences non-contiguous, which nothing
-- depends on — nextPromptOrdinal takes max+1 and readers order by ordinal DESC.
--
-- 102 widens the step_kind CHECK on databases Atlas leaves on the old
-- expression, so it has to run first or every UPDATE below is rejected.
UPDATE public.todo_issue_prompt_runs AS link
SET step_kind = 'triage'
FROM public.captain_prompt_runs AS run
WHERE run.id = link.prompt_run_id
  AND link.step_kind = 'plan'
  AND run.spec_profile = 'triage';
