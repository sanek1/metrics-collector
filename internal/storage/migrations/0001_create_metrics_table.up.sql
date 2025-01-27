CREATE TABLE metrics (
    id serial PRIMARY KEY,
    "key" text NOT NULL,
    m_type text NOT NULL,
    delta bigint,
    value double precision
);