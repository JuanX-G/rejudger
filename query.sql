-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetBasicInfoByEmail :one
SELECT password, id, name FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: GetBasicInfoByUserName :one
SELECT password, id, name FROM users
WHERE name = $1 LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (
  name, internal_id, password, email
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;

-- name: GetUserPermissions :many
SELECT DISTINCT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;

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

-- name: InsertRole :exec
INSERT INTO roles (name)
VALUES ($1) RETURNING *;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE name = $1 LIMIT 1;

-- name: AddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2);

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1;
