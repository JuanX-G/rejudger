-- name: SetDeadlineLastRanToNowByStageId :exec
UPDATE stage_deadlines
SET last_ran = NOW()
WHERE stage_id = $1;

-- name: InsertStageDeadline :exec
INSERT INTO stage_deadlines (stage_id, last_ran, duration)
VALUES ($1, $2, $3);
