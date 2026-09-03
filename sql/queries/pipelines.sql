-- name: InsertStage :one
INSERT INTO stages (
    name,
    version,
    plus_one_required,
    blind,
    soft_veto,
    has_deadline,
    deadline_str
)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: InsertPipeline :one
INSERT INTO pipelines (
    name,
    version,
    description
)
VALUES ($1,$2,$3)
RETURNING *;

-- name: InsertPipelineStage :exec
INSERT INTO pipeline_stages (
    pipeline_id,
    stage_id,
    position
)
VALUES ($1,$2,$3);

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


-- name: GetPipelineByVersion :one
SELECT * FROM pipelines
WHERE version = $1;

-- name: DeletePipelineById :exec
DELETE FROM pipelines
WHERE id = $1;

-- name: GetHowManyUseByPipelineId :one
SELECT COUNT(s.id)
FROM submissions s
JOIN pipelines p
    ON s.pipeline = p.id
WHERE p.id = $1;

-- name: GetPipelinesIds :many
SELECT id FROM pipelines;
