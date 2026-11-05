-- name: InsertPermission :exec
INSERT INTO permissions (action, context)
VALUES ($1, $2)
ON CONFLICT (action, context) DO NOTHING;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE context = $1 AND action = $2;

-- name: DeletePermissionContext :exec
DELETE FROM permissions WHERE context = $1;
