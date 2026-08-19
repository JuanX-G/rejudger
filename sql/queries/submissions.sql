-- name: GetUnusedPipelines :many
SELECT p.*
FROM pipelines p
WHERE NOT EXISTS (
    SELECT 1
    FROM submissions s
    WHERE s.pipeline_id = p.id
);

-- name: GetSubmissionsByHash :one
SELECT * FROM submissions
WHERE hash = $1;

-- name: GetSubmissionsByAuthor :many
SELECT * FROM submissions
WHERE author = $1;

-- name: GetSubmissionsById :one
SELECT * FROM submissions
WHERE id = $1;

-- name: InsertSubmission :exec
INSERT INTO submissions (content, author, hash)
VALUES ($1, $2, $3);

-- name: AddPlusOne :exec
INSERT INTO plus_ones (user_id, submission_id)
VALUES ($1, $2);

-- name: DeletePlusOne :exec
DELETE FROM plus_ones WHERE user_id = $1 AND submission_id = $2;

-- name: ResetPlusOneForSubmission :exec
DELETE FROM plus_ones WHERE submission_id = $1;

-- name: CountPlusOnes :one
SELECT COUNT(*) FROM plus_ones
WHERE submission_id = $1;
