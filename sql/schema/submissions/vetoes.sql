CREATE TABLE vetoes (
    id BIGSERIAL PRIMARY KEY,
    vetoer BIGINT NOT NULL REFERENCES users(id),
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    hard BOOL NOT NULL,

    UNIQUE(vetoer, submission_id)
);
