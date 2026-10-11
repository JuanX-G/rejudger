-- name: InsertSubmission :exec
INSERT INTO submissions (content, author, hash)
VALUES ($1, $2, $3);

-- name: AddPlusOne :exec
INSERT INTO plus_ones (user_id, submission_id, pipeline_stage)
VALUES ($1, $2, $3);

-- name: DeletePlusOne :exec
DELETE FROM plus_ones WHERE user_id = $1 AND submission_id = $2 AND pipeline_stage = $3;

-- name: ResetPlusOneForSubmission :exec
DELETE FROM plus_ones WHERE submission_id = $1;

-- name: AddOneToSubmissionStageById :exec
UPDATE submissions
SET pipeline_stage = pipeline_stage + 1
WHERE id = $1;
