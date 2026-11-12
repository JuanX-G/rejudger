CREATE TABLE permissions (
	id SERIAL PRIMARY KEY,
	action TEXT NOT NULL,
	context BIGINT NOT NULL REFERENCES contexts(id) ON DELETE CASCADE,


	UNIQUE(action, context)
);

CREATE TABLE role_permissions (
	role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
	permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

	PRIMARY KEY (role_id, permission_id)
);
