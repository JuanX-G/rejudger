-- name: InsertRole :exec
INSERT INTO roles (name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING
RETURNING *;

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: AddUserToRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2);

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: DeleteRoleByName :exec
DELETE FROM roles
WHERE name = $1;
