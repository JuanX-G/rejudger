CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    content TEXT NOT NULL,
    author BIGINT NOT NULL REFERENCES users(id),
    hash VARCHAR(128) NOT NULL UNIQUE,
    pipeline BIGINT NOT NULL REFERENCES pipelines(id),
    pipeline_stage BIGINT NOT NULL REFERENCES pipeline_stages(id)
);

CREATE TABLE plus_ones (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    made_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, submission_id)
);

CREATE TABLE submission_artifacts (
    id BIGSERIAL PRIMARY KEY,
    hash VARCHAR(128) NOT NULL UNIQUE,
    name TEXT NOT NULL,
    generation SERIAL NOT NULL,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    file_type TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(hash, generation)
);

CREATE INDEX idx_artifacts_submission_id ON submission_artifacts(submission_id);
