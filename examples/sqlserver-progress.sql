CREATE PROCEDURE #report_progress
AS
BEGIN
    SET NOCOUNT ON;
    DECLARE @completed int = 0, @total int = 5;
    RAISERROR('::description::Processing SQL Server batches', 0, 1) WITH NOWAIT;
    RAISERROR('::status::Starting', 0, 1) WITH NOWAIT;
    RAISERROR('::progress::%d/%d', 0, 1, @completed, @total) WITH NOWAIT;
    WHILE @completed < @total
    BEGIN
        WAITFOR DELAY '00:00:01';
        SET @completed += 1;
        RAISERROR('Completed batch %d of %d', 0, 1, @completed, @total) WITH NOWAIT;
        RAISERROR('::progress::%d/%d', 0, 1, @completed, @total) WITH NOWAIT;
        RAISERROR('::status::Completed batch %d of %d', 0, 1, @completed, @total) WITH NOWAIT;
    END;
    RAISERROR('::description::SQL Server batches complete', 0, 1) WITH NOWAIT;
    SELECT @completed AS completed;
END;
GO
EXEC #report_progress;
