// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package db

import (
	"context"
	"time"
)

const countLikers = `-- name: CountLikers :one
SELECT COUNT(*)
FROM decisions
WHERE recipient_user_id = $1
  AND liked = true
`

func (q *Queries) CountLikers(ctx context.Context, recipientUserID string) (int64, error) {
	row := q.db.QueryRow(ctx, countLikers, recipientUserID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listLikers = `-- name: ListLikers :many
SELECT
    actor_user_id,
    created_at
FROM decisions
WHERE recipient_user_id = $1
  AND liked = true
  AND (
    CASE
        WHEN $2::TIMESTAMPTZ > '0001-01-02'::TIMESTAMPTZ THEN
            created_at < $2
        ELSE true
        END
    )
ORDER BY created_at DESC, actor_user_id DESC
LIMIT $3
`

type ListLikersParams struct {
	RecipientUserID string    `json:"recipientUserId"`
	CreatedAtCursor time.Time `json:"createdAtCursor"`
	PageLimit       int32     `json:"pageLimit"`
}

type ListLikersRow struct {
	ActorUserID string    `json:"actorUserId"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (q *Queries) ListLikers(ctx context.Context, arg ListLikersParams) ([]ListLikersRow, error) {
	rows, err := q.db.Query(ctx, listLikers, arg.RecipientUserID, arg.CreatedAtCursor, arg.PageLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListLikersRow
	for rows.Next() {
		var i ListLikersRow
		if err := rows.Scan(&i.ActorUserID, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listNewLikers = `-- name: ListNewLikers :many
SELECT
    d1.actor_user_id,
    d1.created_at
FROM decisions d1
         LEFT JOIN decisions d2 ON
    d1.actor_user_id = d2.recipient_user_id
        AND d1.recipient_user_id = d2.actor_user_id
        AND d2.liked = true
WHERE d1.recipient_user_id = $1
  AND d1.liked = true
  AND d2.actor_user_id IS NULL
  AND (
    CASE
        WHEN $2::TIMESTAMPTZ > '0001-01-02'::TIMESTAMPTZ THEN
            d1.created_at < $2
        ELSE true
        END
    )
ORDER BY d1.created_at DESC, d1.actor_user_id DESC
LIMIT $3
`

type ListNewLikersParams struct {
	RecipientUserID string    `json:"recipientUserId"`
	CreatedAtCursor time.Time `json:"createdAtCursor"`
	PageLimit       int32     `json:"pageLimit"`
}

type ListNewLikersRow struct {
	ActorUserID string    `json:"actorUserId"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (q *Queries) ListNewLikers(ctx context.Context, arg ListNewLikersParams) ([]ListNewLikersRow, error) {
	rows, err := q.db.Query(ctx, listNewLikers, arg.RecipientUserID, arg.CreatedAtCursor, arg.PageLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListNewLikersRow
	for rows.Next() {
		var i ListNewLikersRow
		if err := rows.Scan(&i.ActorUserID, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const putDecision = `-- name: PutDecision :one
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
    ) mutual_likes
`

type PutDecisionParams struct {
	ActorUserID     string `json:"actorUserId"`
	RecipientUserID string `json:"recipientUserId"`
	Liked           bool   `json:"liked"`
}

func (q *Queries) PutDecision(ctx context.Context, arg PutDecisionParams) (bool, error) {
	row := q.db.QueryRow(ctx, putDecision, arg.ActorUserID, arg.RecipientUserID, arg.Liked)
	var mutual_likes bool
	err := row.Scan(&mutual_likes)
	return mutual_likes, err
}
