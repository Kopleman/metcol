-- name: CreateMetric :one
INSERT INTO metrics (name, type, value, delta, created_at)
	VALUES ($1, $2, $3, $4, now())
	RETURNING *;

-- name: ExistsMetric :one
SELECT EXISTS (SELECT * FROM metrics WHERE name=$1 AND type=$2)::boolean;

