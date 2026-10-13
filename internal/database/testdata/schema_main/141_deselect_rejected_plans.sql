-- phase: post
-- dependsOn: 140_todo_prompt_run_step_open.sql

-- Rejection retains the immutable Captain plan and its Gavel link as history,
-- but a rejected plan cannot remain the issue's active selection. Historical
-- rejection events already record the decision, so this projection repair does
-- not synthesize another event or advance the issue version.
UPDATE public.todo_issues AS issue
SET selected_plan_id = NULL,
    status = 'open'
FROM public.captain_plans AS plan
WHERE issue.selected_plan_id = plan.id
  AND plan.approval_state = 'rejected';
