-- name: InsertRole :exec
INSERT INTO roles (name)
VALUES ($1) RETURNING *;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE name = $1 LIMIT 1;

-- name: GetRoleNames :many
SELECT name FROM roles;

-- name: AddPermissionToRole :exec
insert into role_permissions (role_id, permission_id)
values ($1, $2);

-- name: AddUserToRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2);

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1;

-- name: DeleteRoleByName :exec
DELETE FROM roles
WHERE name = $1;

-- name: GetUnusedRoles :many
SELECT r.*
FROM roles r
WHERE NOT EXISTS (
    SELECT 1
    FROM user_roles ur
    WHERE r.id = ur.role_id
);
