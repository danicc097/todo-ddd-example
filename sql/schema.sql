CREATE TABLE todos(
  id uuid PRIMARY KEY,
  title text NOT NULL,
  completed boolean NOT NULL DEFAULT FALSE,
  created_at timestamptz NOT NULL DEFAULT NOW()
);

