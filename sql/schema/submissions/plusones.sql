CREATE TABLE plus_ones (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    pipeline_stage INT NOT NULL REFERENCES pipeline_stages(id),
    made_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, submission_id)
);
