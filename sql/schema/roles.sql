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
