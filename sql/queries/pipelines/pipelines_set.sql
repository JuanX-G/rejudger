-- name: InsertPipeline :one
INSERT INTO pipelines (
    name,
    version,
    description
)
VALUES ($1,$2,$3)
RETURNING *;

-- Insert a stage into a pipeline
-- name: InsertPipelineStage :exec
INSERT INTO pipeline_stages (
    pipeline_id,
    stage_id,
    position
)
VALUES ($1,$2,$3);

-- name: DeletePipelineById :exec
DELETE FROM pipelines
WHERE id = $1;
