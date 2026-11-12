CREATE TABLE published_submissions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP NOT NULL DEFAULT NOW(),
    content TEXT NOT NULL,
    author BIGINT NOT NULL REFERENCES users(id),
    hash VARCHAR(128) NOT NULL UNIQUE,
    context TEXT NOT NULL
);
