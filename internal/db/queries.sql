-- name: PutDecision :one
INSERT INTO decisions (
    actor_user_id, recipient_user_id, liked
) VALUES (
             $1, $2, $3
         )
ON CONFLICT (actor_user_id, recipient_user_id)
    DO UPDATE SET liked = EXCLUDED.liked, updated_at = NOW()
RETURNING (
    EXISTS (
        SELECT 1 FROM decisions
        WHERE actor_user_id = $2
          AND recipient_user_id = $1
          AND liked = true
    )
    ) mutual_likes;

-- name: ListLikers :many
SELECT
    actor_user_id,
    created_at
FROM decisions
WHERE recipient_user_id = sqlc.arg(recipient_user_id)
  AND liked = true
  AND (
    CASE
        WHEN sqlc.arg(created_at_cursor)::TIMESTAMPTZ > '0001-01-02'::TIMESTAMPTZ THEN
            created_at < sqlc.arg(created_at_cursor)
        ELSE true
        END
    )
ORDER BY created_at DESC, actor_user_id DESC
LIMIT sqlc.arg(page_limit);

-- name: ListNewLikers :many
SELECT
    d1.actor_user_id,
    d1.created_at
FROM decisions d1
         LEFT JOIN decisions d2 ON
    d1.actor_user_id = d2.recipient_user_id
        AND d1.recipient_user_id = d2.actor_user_id
        AND d2.liked = true
WHERE d1.recipient_user_id = sqlc.arg(recipient_user_id)
  AND d1.liked = true
  AND d2.actor_user_id IS NULL
  AND (
    CASE
        WHEN sqlc.arg(created_at_cursor)::TIMESTAMPTZ > '0001-01-02'::TIMESTAMPTZ THEN
            d1.created_at < sqlc.arg(created_at_cursor)
        ELSE true
        END
    )
ORDER BY d1.created_at DESC, d1.actor_user_id DESC
LIMIT sqlc.arg(page_limit);

-- name: CountLikers :one
SELECT COUNT(*)
FROM decisions
WHERE recipient_user_id = $1
  AND liked = true;