-- name: InsertStage :one
INSERT INTO stages (
    name,
    version,
    plus_one_required,
    blind,
    soft_veto,
    has_deadline,
    deadline_str,
    final
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;
