-- name: GetUnusedRoles :many
SELECT r.*
FROM roles r
WHERE NOT EXISTS (
    SELECT 1
    FROM user_roles ur
    WHERE r.id = ur.role_id
);

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE name = $1 LIMIT 1;

-- name: GetRoleNames :many
SELECT name FROM roles;
