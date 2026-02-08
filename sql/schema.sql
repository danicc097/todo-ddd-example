CREATE TABLE todos(
  id uuid PRIMARY KEY,
  title text NOT NULL,
  status text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);

