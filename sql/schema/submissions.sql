CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    content TEXT,
    author BIGINT NOT NULL REFERENCES users(id),
    hash VARCHAR(128) NOT NULL UNIQUE
);

CREATE TABLE plus_ones (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    made_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, submission_id)
);

CREATE TABLE submission_artifacts (
    id SERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    artifact_url TEXT NOT NULL,
    file_type TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_artifacts_submission_id ON submission_artifacts(submission_id);
