CREATE TABLE users (
	id   		BIGSERIAL PRIMARY KEY,
	name 		TEXT NOT NULL UNIQUE,
	internal_id	TEXT UNIQUE,
	password 	TEXT NOT NULL,
	email		TEXT UNIQUE
);


CREATE TABLE categories (
	id 		SERIAL PRIMARY KEY,
	name 		TEXT NOT NULL UNIQUE
);

CREATE TABLE sections (
	id 		SERIAL PRIMARY KEY,
	name 		TEXT NOT NULL UNIQUE,
	category	INT REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE section_pipelines (
	section_id INT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
	pipeline_id INT NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE
);

CREATE TABLE pipelines (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    version     JSONB NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE stages (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    version JSONB NOT NULL UNIQUE,
    plus_one_required SMALLINT NOT NULL,
    blind BOOL NOT NULL,
    soft_veto BOOL NOT NULL,
    has_deadline BOOL NOT NULL,
    deadline_str TEXT
);

CREATE TABLE pipeline_stages (
    pipeline_id BIGINT NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    stage_id    BIGINT NOT NULL REFERENCES stages(id),
    position    INT NOT NULL,

    PRIMARY KEY (pipeline_id, position),
    UNIQUE (pipeline_id, stage_id)
);

CREATE TABLE roles (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

CREATE TABLE permissions (
	id SERIAL PRIMARY KEY,
	action TEXT NOT NULL,
	context TEXT NOT NULL,


	UNIQUE(action, context)
);

CREATE TABLE role_permissions (
	role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

	PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, role_id)
);
