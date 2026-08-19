-- name: InsertPermission :exec
INSERT INTO permissions (action, context)
VALUES ($1, $2);


-- name: DeletePermission :exec
DELETE FROM permissions WHERE context = $1 AND action = $2;

-- name: DeletePermissionContext :exec
DELETE FROM permissions WHERE context = $1;

-- name: GetPermissionByContext :many
SELECT action FROM permissions WHERE context = $1 LIMIT 1;

-- name: GetPermissionId :one
SELECT id FROM permissions WHERE context = $1 AND action = $2;
