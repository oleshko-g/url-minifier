INSERT INTO
  strings (
    user_id,
    id,
    value,
    created_at,
    updated_at,
    deleted_at
  )
VALUES
  ($1, $2, $3, $4, $5, $6)
ON CONFLICT ON CONSTRAINT value_idx DO NOTHING
RETURNING
  id,
  value;
