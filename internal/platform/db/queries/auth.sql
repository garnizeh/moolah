-- name: CreateOTPRequest :one
INSERT INTO otp_requests (
    id, email, code_hash, used, expires_at, created_at
) VALUES (
    $1, $2, $3, $4, $5, NOW()
) RETURNING id, email, code_hash, used, expires_at, created_at;

-- name: GetActiveOTPByEmail :one
SELECT id, email, code_hash, used, expires_at, created_at
FROM otp_requests
WHERE email = $1 AND used = FALSE AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: MarkOTPUsed :exec
UPDATE otp_requests
SET used = TRUE
WHERE id = $1;

-- name: DeleteExpiredOTPs :exec
DELETE FROM otp_requests
WHERE expires_at < NOW() OR (used = TRUE AND created_at < NOW() - INTERVAL '24 hours');
