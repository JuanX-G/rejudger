-- name: ExistsArtifact :one
SELECT EXISTS(
    SELECT 1 FROM submission_artifacts
    WHERE hash = $1 AND file_type = $2
);

-- name: ExistsArtifactRev :one
SELECT EXISTS(
    SELECT 1 FROM submission_artifacts
    WHERE hash = $1 AND file_type = $2
    AND generation = $3
);

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
