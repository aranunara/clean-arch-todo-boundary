CREATE TABLE todos (
	id text PRIMARY KEY,
	title text NOT NULL CHECK (btrim(title) <> ''),
	completed boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);
