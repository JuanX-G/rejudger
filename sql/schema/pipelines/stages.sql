CREATE TABLE stages (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    version JSONB NOT NULL UNIQUE,
    plus_one_required SMALLINT NOT NULL,
    blind BOOL NOT NULL,
    soft_veto BOOL NOT NULL,
    has_deadline BOOL NOT NULL,
    deadline_str TEXT,
    final BOOL NOT NULL
);
