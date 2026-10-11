-- name: GetDeadlineByStageId :one
SELECT * FROM stage_deadlines
WHERE stage_id = $1;
