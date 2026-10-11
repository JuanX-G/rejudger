-- name: GetPermissionByContext :many
SELECT action FROM permissions WHERE context = $1 LIMIT 1;

-- name: GetPermissionId :one
SELECT id FROM permissions WHERE context = $1 AND action = $2;
