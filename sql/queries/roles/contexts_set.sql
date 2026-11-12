-- name: AddContext :exec
INSERT INTO contexts (name) VALUES ($1);

-- name: DeleteContextsById :exec
DELETE FROM contexts WHERE id = $1;
