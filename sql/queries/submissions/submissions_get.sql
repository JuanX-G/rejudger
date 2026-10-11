-- name: GetSubmissionsByHash :one
SELECT * FROM submissions
WHERE hash = $1;

-- name: GetSubmissionsByAuthor :many
SELECT * FROM submissions
WHERE author = $1;

-- name: GetSubmissionsByAuthorOffset :many
SELECT * FROM submissions
WHERE author = $1
OFFSET  $2
LIMIT $3;

-- name: GetSubmissionsById :one
SELECT * FROM submissions
WHERE id = $1;

-- name: GetSubmissionsInStage :many
SELECT s.* FROM submissions
WHERE pipeline = $1 AND pipeline_stage = $2;

-- name: GetSubmissionsByStageId :many
SELECT s.*
FROM submissions s
JOIN pipeline_stages ps ON ps.id = s.pipeline_stage
WHERE ps.stage_id = $1;


-- name: GetVetosBySubmissionId :many
SELECT * FROM vetoes
WHERE submission_id = $1;

-- name: CountPlusOnes :one
SELECT COUNT(*) FROM plus_ones
WHERE submission_id = $1 AND pipeline_stage = $2;


-- name: AdvanceSubmissionToNextStage :execrows
UPDATE submissions s
SET pipeline_stage = next_ps.id
FROM pipeline_stages current_ps
JOIN pipeline_stages next_ps
  ON next_ps.pipeline_id = current_ps.pipeline_id
 AND next_ps.position = current_ps.position + 1
WHERE s.id = $1
  AND current_ps.id = s.pipeline_stage;
