-- name: SetProcessed :batchexec
UPDATE
    queue
SET
    processed_count = processed_count+1
WHERE
    id = $1::integer;

-- name: BatchGet :many
SELECT
    id,
    due_date
FROM queue
WHERE
    processed_count = 0
AND due_date < now()
ORDER BY
    due_date ASC
LIMIT
    $1::integer
for update skip locked;