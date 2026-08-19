CREATE TABLE users (
	id   		BIGSERIAL PRIMARY KEY,
	name 		TEXT NOT NULL UNIQUE,
	internal_id	TEXT UNIQUE,
	password 	TEXT NOT NULL,
	email		TEXT UNIQUE
);

CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, role_id)
);
