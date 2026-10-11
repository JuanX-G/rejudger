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
