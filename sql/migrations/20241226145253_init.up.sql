CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
DROP TYPE IF EXISTS METRIC_TYPE;
CREATE TYPE METRIC_TYPE AS ENUM ('gauge', 'counter');

CREATE TABLE IF NOT EXISTS metrics
(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    type       METRIC_TYPE  NOT NULL,
    value      DOUBLE PRECISION NULL,
    delta      BIGINT  NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NULL,
    deleted_at TIMESTAMP    NULL,
    CONSTRAINT name_type_uniq UNIQUE(name, type),
    CONSTRAINT value_delta_null_check CHECK (NOT(value IS NULL AND delta IS NULL))
);