-- name: GetAllMetrics :many
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics ORDER BY name ASC;

-- name: GetMetric :one
SELECT id, name, type, value, delta, created_at, updated_at, deleted_at FROM metrics WHERE type=$1 AND name=$2 LIMIT 1;

-- name: UpdateMetricAndGet :one
UPDATE metrics
SET value=$1, delta=$2, updated_at=now()
WHERE type=$3 AND name=$4
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at;

-- name: UpdateMetric :exec
UPDATE metrics
SET value=$1, delta=$2, updated_at=now()
WHERE type=$3 AND name=$4;

-- name: CreateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
VALUES ($1, $2, $3, $4, now())
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at;

-- name: ExistsMetric :one
SELECT EXISTS (SELECT * FROM metrics WHERE name=$1 AND type=$2)::boolean;

-- name: CreateOrUpdateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
VALUES ($1, $2, $3, $4, now())
    ON CONFLICT ON CONSTRAINT name_type_uniq DO UPDATE SET value=$3, delta=$4, updated_at=now()
    RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at;