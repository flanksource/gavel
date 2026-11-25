SET client_min_messages = NOTICE;
GO

CREATE FUNCTION pg_temp.report_progress() RETURNS integer
LANGUAGE plpgsql AS $$
DECLARE
    completed integer;
    total constant integer := 5;
BEGIN
    RAISE NOTICE '::description::Processing PostgreSQL batches';
    RAISE NOTICE '::status::Starting';
    RAISE NOTICE '::progress::0/%', total;
    FOR completed IN 1..total LOOP
        PERFORM pg_sleep(1);
        RAISE NOTICE 'Completed batch % of %', completed, total;
        RAISE NOTICE '::progress::%/%', completed, total;
        RAISE NOTICE '::status::Completed batch % of %', completed, total;
    END LOOP;
    RAISE NOTICE '::description::PostgreSQL batches complete';
    RETURN total;
END;
$$;
GO

SELECT pg_temp.report_progress() AS completed;
