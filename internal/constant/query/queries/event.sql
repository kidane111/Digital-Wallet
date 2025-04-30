-- name: CreateEvent :one
INSERT INTO webhook_events (
    id,
    event_id,
    event_type,
    status,
    payload,
    attempts,
    last_attempt_at,
    next_attempt_at,
    error_message
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetEventByID :one
SELECT * FROM webhook_events WHERE id = $1 LIMIT 1;

-- name: GetEventByEventID :one
SELECT * FROM webhook_events WHERE event_id = $1 LIMIT 1;

-- name: UpdateEventStatus :exec
UPDATE webhook_events
SET 
    status = $2,
    attempts = $3,
    last_attempt_at = $4,
    next_attempt_at = $5,
    error_message = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetPendingEvents :many
SELECT * FROM webhook_events
WHERE status = 'pending' AND (next_attempt_at IS NULL OR next_attempt_at <= CURRENT_TIMESTAMP)
ORDER BY created_at ASC
LIMIT $1;

-- name: DeleteEvent :exec
DELETE FROM webhook_events WHERE id = $1;