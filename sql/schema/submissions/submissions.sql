CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    content TEXT NOT NULL,
    author BIGINT NOT NULL REFERENCES users(id),
    hash VARCHAR(128) NOT NULL UNIQUE,
    pipeline BIGINT NOT NULL REFERENCES pipelines(id),
    pipeline_stage INT NOT NULL REFERENCES pipeline_stages(id)
);
