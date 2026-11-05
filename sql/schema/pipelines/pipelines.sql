CREATE TABLE pipelines (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    version     JSONB NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE pipeline_stages (
    id          BIGSERIAL PRIMARY KEY,
    pipeline_id BIGINT NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    stage_id    BIGINT NOT NULL REFERENCES stages(id),
    position    INT NOT NULL,

    UNIQUE (pipeline_id, position),
    UNIQUE (pipeline_id, stage_id)
);
