-- name: GetUnusedPipelines :many
SELECT p.*
FROM pipelines p
WHERE NOT EXISTS (
    SELECT 1
    FROM submissions s
    WHERE s.pipeline_id = p.id
);

-- name: GetPipelineByVersion :one
SELECT * FROM pipelines
WHERE version = $1;

-- name: GetPipelinesIds :many
SELECT id FROM pipelines;

-- name: GetHowManyUseByPipelineId :one
SELECT COUNT(s.id)
FROM submissions s
JOIN pipelines p
    ON s.pipeline = p.id
WHERE p.id = $1;
