SELECT
  user_id,
  id,
  value
FROM
  strings
WHERE
  user_id = $1
  AND deleted_at IS NOT NULL;
