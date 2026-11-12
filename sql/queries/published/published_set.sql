-- name: InsertIntoPublishedSubmissions :exec
INSERT INTO published_submissions
(created_at, content, author, hash, context)
VALUES ($1, $2, $3, $4, $5);
