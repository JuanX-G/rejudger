CREATE TABLE stage_deadlines (
    id          BIGSERIAL PRIMARY KEY,
    stage_id    BIGINT UNIQUE NOT NULL REFERENCES stages(id),
    last_ran    TIMESTAMP NOT NULL,
    duration    BIGINT NOT NULL
);
