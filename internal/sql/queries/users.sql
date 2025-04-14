-- name: CreateUser :one
INSERT INTO users (email, password_hash, role) 
VALUES ($1, md5($2), $3)
RETURNING id, email, role;

-- name: GetUserByCredentials :one
SELECT role 
FROM users 
WHERE email = $1 AND password_hash = md5($2);
