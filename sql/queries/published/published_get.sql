-- name: GetPublishedByHash :one
SELECT * FROM published_submissions
WHERE hash = $1;
