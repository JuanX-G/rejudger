-- name: GetPipelineStages :many
SELECT s.*
FROM pipeline_stages ps
JOIN stages s
    ON s.id = ps.stage_id
WHERE ps.pipeline_id = $1
ORDER BY ps.position;

-- name: GetStageByVersion :one
SELECT * FROM stages
WHERE version = $1;

-- name: GetStageById :one
SELECT * FROM stages
WHERE id = $1;

-- name: GetPipelineStage :one
SELECT * FROM pipeline_stages
WHERE pipeline_id = $1 AND position = $2;
