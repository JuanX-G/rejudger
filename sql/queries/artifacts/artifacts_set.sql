-- name: DeleteArtifact :exec
DELETE FROM submission_artifacts
WHERE hash = $1 AND file_type = $2;

-- name: InsertArtifact :exec
INSERT INTO submission_artifacts (hash, name, submission_id, file_type)
VALUES ($1, $2, $3, $4);


-- name: DeleteArtifactByGen :exec
SELECT * FROM submission_artifacts
WHERE hash = $1 AND file_type = $2 AND generation = $3;
