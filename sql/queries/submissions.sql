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

-- name: GetSubmissionsByAuthorOffset :many
SELECT * FROM submissions
WHERE author = $1
OFFSET  $2;

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

-- name: ExistsArtifact :one
SELECT EXISTS(SELECT 1 FROM submission_artifacts WHERE hash = $1 AND file_type = $2);

-- name: ExistsArtifactRev :one
SELECT EXISTS(SELECT 1 FROM submission_artifacts WHERE hash = $1 AND file_type = $2 AND generation = $3);

-- name: InsertArtifact :exec
INSERT INTO submission_artifacts (hash, name, submission_id, file_type)
VALUES ($1, $2, $3, $4);

-- name: GetNewestArtifact :one
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2
ORDER BY generation DESC
LIMIT 1;

-- name: GetArtifactByGeneration :one
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2 AND generation = $3;

-- name: GetArtifacts :many
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2
ORDER BY generation DESC;

-- name: GetArtifactsBySubmission :many
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2 AND submission_id = $3
ORDER BY generation DESC;

-- name: DeleteArtifactByGen :exec
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2 AND generation = $3;

-- name: DeleteArtifact :exec
DELETE FROM submission_artifacts
WHERE hash = $1 AND file_type = $2;
