-- phase: post

-- Mirror every issue's primary PR (todo_issues.pr) into todo_pr_links as an
-- implements link, so the links table is complete from the moment it exists.
-- From here on native.Repository keeps the two in step: setting the pr column
-- upserts the implements link, clearing it deletes that link.
--
-- The repository key must be exactly the one types.PRRefFromURL derives, or a
-- backfilled link and a link written later for the same PR would be two rows.
-- Both work from the stored URL's escaped path and lower-case the result:
--   GitHub   https://github.com/<owner>/<repo>/pull/<n>                    -> owner/repo
--   Azure    https://dev.azure.com/<org>/[<project>/]_git/<repo>/pullrequest/<n>
--            https://<org>.visualstudio.com/[DefaultCollection/][<project>/]_git/<repo>/pullrequest/<n>
--                                                                          -> org[/project]/repo
-- concat_ws skips the NULL an absent optional project matches as.
WITH matches AS (
  SELECT issue.id, issue.workspace_id, issue.updated_at, issue.pr ->> 'url' AS url,
    regexp_match(issue.pr ->> 'url', '^https://github\.com/([^/]+)/([^/]+)/pull/([0-9]+)$') AS gh,
    regexp_match(issue.pr ->> 'url', '^https://dev\.azure\.com/([^/]+)/(?:([^/]+)/)?_git/([^/]+)/pullrequest/([0-9]+)$') AS az,
    regexp_match(issue.pr ->> 'url', '^https://([^./]+)\.visualstudio\.com/(?:[Dd][Ee][Ff][Aa][Uu][Ll][Tt][Cc][Oo][Ll][Ll][Ee][Cc][Tt][Ii][Oo][Nn]/)?(?:([^/]+)/)?_git/([^/]+)/pullrequest/([0-9]+)$') AS vs
  FROM public.todo_issues AS issue
  WHERE jsonb_typeof(issue.pr) = 'object'
), refs AS (
  SELECT id, workspace_id, updated_at, url,
    CASE
      WHEN gh IS NOT NULL THEN 'github'
      WHEN az IS NOT NULL OR vs IS NOT NULL THEN 'azure'
    END AS provider,
    CASE
      WHEN gh IS NOT NULL THEN lower(gh[1] || '/' || gh[2])
      WHEN az IS NOT NULL THEN lower(concat_ws('/', az[1], az[2], az[3]))
      WHEN vs IS NOT NULL THEN lower(concat_ws('/', vs[1], vs[2], vs[3]))
    END AS repo,
    COALESCE(gh[3], az[4], vs[4])::bigint AS number
  FROM matches
)
INSERT INTO public.todo_pr_links
  (workspace_id, issue_id, provider, repo, number, url, relation, source, created_at)
SELECT workspace_id, id, provider, repo, number, url, 'implements', 'manual', updated_at
FROM refs
WHERE provider IS NOT NULL
ON CONFLICT (workspace_id, issue_id, provider, repo, number) DO NOTHING;

-- Every stored primary PR was validated by types.PRFromURL on write, so each one
-- must have keyed. One that did not means the patterns above have drifted from
-- PRRefFromURL; failing names it instead of leaving that TODO's PR off its links.
DO $$
DECLARE
  unkeyed text;
BEGIN
  SELECT string_agg(issue.id::text || ' (' || (issue.pr ->> 'url') || ')', ', ')
  INTO unkeyed
  FROM public.todo_issues AS issue
  WHERE jsonb_typeof(issue.pr) = 'object'
    AND NOT EXISTS (
      SELECT 1 FROM public.todo_pr_links AS link
      WHERE link.issue_id = issue.id AND link.relation = 'implements'
    );
  IF unkeyed IS NOT NULL THEN
    RAISE EXCEPTION 'todo_pr_links backfill could not key the primary PR of: %', unkeyed;
  END IF;
END
$$;
