SELECT
  user_id,
  value,
  deleted_at
FROM
  strings
WHERE
  id = $1;
