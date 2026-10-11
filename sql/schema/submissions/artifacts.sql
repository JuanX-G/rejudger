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
