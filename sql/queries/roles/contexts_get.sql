-- name: GetUnusedContexts :many
SELECT c.*
FROM contexts c
WHERE NOT EXISTS (
    SELECT 1
    FROM pipelines pl
    WHERE pl.context = c.id
) AND
NOT EXISTS (
    SELECT 1
    FROM permissions ps
    WHERE ps.context = c.id
);

-- name: GetContextByName :one
SELECT * FROM contexts
WHERE name = $1
LIMIT 1;


-- name: GetContextById :one
SELECT * FROM contexts
WHERE id = $1
LIMIT 1;
