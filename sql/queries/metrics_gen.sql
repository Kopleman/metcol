-- name: CreateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
	VALUES ($1, $2, $3, $4, now())
	RETURNING id, name, type, value, delta, created_at, updated_at, deleted_at;

-- name: ExistsMetric :one
SELECT EXISTS (SELECT * FROM metrics WHERE name=$1 AND type=$2)::boolean;

