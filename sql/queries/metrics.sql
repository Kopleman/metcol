-- name: GetAllMetrics :many
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics ORDER BY name ASC;

-- name: GetMetric :one
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics WHERE type=$1 AND name=$2 LIMIT 1;

-- name: UpdateMetric :exec
UPDATE metrics
SET value=$1, delta=$2, updated_at=now()
WHERE type=$3 AND name=$4
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at;