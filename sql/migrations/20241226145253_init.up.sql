DROP TYPE IF EXISTS METRIC_TYPE;
CREATE TYPE METRIC_TYPE AS ENUM ('gauge', 'counter');

CREATE TABLE metrics
(
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    type       METRIC_TYPE  NOT NULL,
    value      DOUBLE PRECISION NULL,
    delta      INTEGER  NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NULL,
    deleted_at TIMESTAMP    NULL,
    UNIQUE (name, type)
);