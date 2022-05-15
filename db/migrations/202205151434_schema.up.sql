CREATE TABLE queue(
  id INTEGER PRIMARY KEY,
  due_date timestamptz NOT NULL,
  processed_count INTEGER NOT NULL DEFAULT 0
);

INSERT INTO
  queue(id, due_date)
SELECT
  x,
  now()
FROM
  generate_series(1, 2000000) x;