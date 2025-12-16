UPDATE strings
SET
  deleted_at = $2
WHERE
  id = $1;
