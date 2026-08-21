-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetBasicInfoByEmail :one
SELECT password, id, name FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: GetUserByInternalId :one
SELECT * FROM users
WHERE internal_id = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetBasicInfoByUserName :one
SELECT password, id, name FROM users
WHERE name = $1 LIMIT 1;

-- name: InsertUser :exec
INSERT INTO users (
  name, internal_id, password, email
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserPermissions :many
SELECT DISTINCT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_roles ur ON ur.role_id = rp.role_id
WHERE ur.user_id = $1;


-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;
