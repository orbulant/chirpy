-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password, created_at, updated_at, is_chirpy_red)
VALUES (
    gen_random_uuid(), $1, $2, NOW(), NOW(), FALSE
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteAllUsers :execrows
DELETE FROM users;

-- name: UpdateUserEmailAndPassword :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetIsUserChirpyRedByID :one
UPDATE users
SET is_chirpy_red = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;