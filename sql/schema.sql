CREATE TABLE todos(
  id uuid PRIMARY KEY,
  title text NOT NULL,
  status text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE users(
  id uuid PRIMARY KEY,
  email text NOT NULL UNIQUE,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);

