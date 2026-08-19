-- name: SetPassword :exec
UPDATE users
SET hashed_password = $1
WHERE id = $2;

-- name: SetEmail :exec
UPDATE users
SET email = $1
WHERE id = $2;

