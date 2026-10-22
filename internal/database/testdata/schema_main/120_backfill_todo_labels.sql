-- phase: post
-- dependsOn: 116_backfill_triage_step_kind.sql

-- Adopt the labels each workspace already uses as editable project labels.
--
-- Before per-project labels existed, a label was just a string on a TODO: it
-- rendered with a colour derived from its own name and appeared in no editable
-- list anywhere. todo_labels only ever holds what someone deliberately wrote, so
-- without this backfill a project that has been tagging TODOs for months opens
-- the Tags editor to a list of built-ins and none of its own vocabulary.
--
-- The colour each adopted label gets is the colour it is ALREADY rendering with,
-- recomputed here rather than reassigned: adoption must not repaint a backlog.
-- That is why this needs the hash below instead of picking hues round-robin.
--
-- Skipped, deliberately:
--   * status:/priority:/session: — lifecycle state Gavel writes and reads back,
--     not vocabulary anyone should be editing the colour of.
--   * the built-in names — they already resolve, already appear in the editor,
--     and already have hand-picked colours that the hash would overwrite. The
--     list is a snapshot of todos/labels/builtin.go taken when this migration
--     was written; it is deliberately NOT kept in sync, because a run-once
--     migration records what was true when it ran.
--   * anything already defined, at either scope — this adopts, never overrides.

-- gavel_backfill_label_hue mirrors labels.Hash in todos/labels/palette.go:
-- FNV-1a over the label's namespace key (or the whole label when it has none),
-- modulo the palette. Byte-wise, like the Go original, so a non-ASCII label
-- hashes identically rather than drifting on codepoints.
CREATE OR REPLACE FUNCTION public.gavel_backfill_label_hue(label text)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
  -- The order is load-bearing: it indexes the hash, so it must match
  -- todos/labels/palette.go exactly or every adopted label changes colour.
  palette CONSTANT text[] := ARRAY[
    'slate', 'red', 'orange', 'amber', 'yellow', 'lime',
    'green', 'emerald', 'teal', 'cyan', 'sky', 'blue',
    'indigo', 'violet', 'purple', 'fuchsia', 'pink', 'rose'];
  seed text := lower(btrim(label));
  boundary int;
  bytes bytea;
  hash bigint := 2166136261;  -- FNV-1a 32-bit offset basis
  i int;
BEGIN
  -- labels.Key: the first ':' or '/' splits a namespace, so every "area/*"
  -- shares one hue. LEAST ignores NULLs, so this is "whichever comes first, if
  -- either does"; position 1 is not a split, matching Go's idx > 0.
  boundary := LEAST(NULLIF(strpos(seed, ':'), 0), NULLIF(strpos(seed, '/'), 0));
  IF boundary IS NOT NULL AND boundary > 1 THEN
    seed := substr(seed, 1, boundary - 1);
  END IF;

  bytes := convert_to(seed, 'UTF8');
  FOR i IN 0 .. octet_length(bytes) - 1 LOOP
    hash := hash # get_byte(bytes, i);
    hash := (hash * 16777619) % 4294967296;  -- FNV prime, truncated to uint32
  END LOOP;

  RETURN palette[(hash % array_length(palette, 1)) + 1];
END;
$$;

INSERT INTO public.todo_labels (workspace_id, name, color)
SELECT used.workspace_id, used.name, public.gavel_backfill_label_hue(used.name)
FROM (
  SELECT DISTINCT issue.workspace_id, lower(btrim(label)) AS name
  FROM public.todo_issues AS issue, unnest(issue.labels) AS label
  WHERE lower(btrim(label)) <> ''
) AS used
WHERE used.name NOT LIKE 'status:%'
  AND used.name NOT LIKE 'priority:%'
  AND used.name NOT LIKE 'session:%'
  AND used.name NOT IN (
    'bug', 'security', 'docs', 'perf', 'test',
    'refactor', 'ui', 'api', 'ci', 'breaking')
  AND NOT EXISTS (
    SELECT 1 FROM public.todo_labels AS defined
    WHERE defined.name = used.name
      AND (defined.workspace_id = used.workspace_id OR defined.workspace_id IS NULL)
  )
ON CONFLICT DO NOTHING;

DROP FUNCTION public.gavel_backfill_label_hue(text);
