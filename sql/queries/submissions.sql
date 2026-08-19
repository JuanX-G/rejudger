-- name: GetUnusedPipelines :many
SELECT p.*
FROM pipelines p
WHERE NOT EXISTS (
    SELECT 1
    FROM submissions s
    WHERE s.pipeline_id = p.id
);
