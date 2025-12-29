package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepositoryInterface interface {
	Save(d model.ChatModel, ctx context.Context) (model.ChatModel, error)
	GetChatBeetween(ctx context.Context, r uuid.UUID, s uuid.UUID) ([]model.ChatModel, error)
	GetLastChat(r uuid.UUID, s uuid.UUID) (model.ChatModel, error)
	MarkConversationAsRead(sender uuid.UUID, receiver uuid.UUID) error
	Delete(id uuid.UUID) error
}

type Attachment struct {
	Filename  string    `json:"file_name"`
	MediaType string    `json:"media_type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatRepository struct {
	Pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{
		Pool: pool,
	}
}

func (r *ChatRepository) Save(d model.ChatModel, ctx context.Context) (model.ChatModel, error) {

	query := `
		INSERT INTO private_messages
			(sender_id, receiver_id, reply_to, chat_text, post_id)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING
			id, sender_id, receiver_id, reply_to, chat_text, post_id, is_read, created_at
	`

	var result model.ChatModel

	err := pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {

		return tx.QueryRow(ctx,
			query, d.SenderId,
			d.ReceiverId,
			d.ReplyTo,
			d.ChatText,
			d.PostId,
		).Scan(
			&result.Id,
			&result.SenderId,
			&result.ReceiverId,
			&result.ReplyTo,
			&result.ChatText,
			&result.PostId,
			&result.IsRead,
			&result.CreatedAt,
		)
	})

	return result, err
}

func (r *ChatRepository) GetChatBeetween(
	ctx context.Context,
	re uuid.UUID,
	se uuid.UUID,
) ([]model.ChatModel, error) {

	query := `
		SELECT
		pm.id,
		pm.sender_id,
		pm.receiver_id,
		pm.post_id,
		pm.reply_to,
		pm.chat_text,
		pm.created_at,
		pm.is_read,

		(pm.sender_id = $3) AS is_own,

		COALESCE(
			(
				SELECT json_agg(
					json_build_object(
						'id', pma.id,
						'chat_id', pma.chat_id,
						'file_name', pma.file_name,
						'media_type', pma.media_type,
						'size', pma.size,
						'created_at', pma.created_at AT TIME ZONE 'UTC'
					)
				)
				FROM private_messages_attachment pma
				WHERE pma.chat_id = pm.id
			),
			'[]'
		) AS attachments
	FROM private_messages pm
	WHERE
		(pm.sender_id = $1 AND pm.receiver_id = $2)
		OR
		(pm.sender_id = $2 AND pm.receiver_id = $1)
	ORDER BY pm.created_at;
	`

	rows, err := r.Pool.Query(ctx, query, re, se, se)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []model.ChatModel

	for rows.Next() {
		var chat model.ChatModel
		var rawAttachments json.RawMessage

		if err := rows.Scan(
			&chat.Id,
			&chat.SenderId,
			&chat.ReceiverId,
			&chat.PostId,
			&chat.ReplyTo,
			&chat.ChatText,
			&chat.CreatedAt,
			&chat.IsRead,
			&chat.IsOwn,
			&rawAttachments,
		); err != nil {
			return nil, err
		}

		var attachments []model.ChatAttachment
		if err := json.Unmarshal(rawAttachments, &attachments); err != nil {
			return nil, err
		}

		chat.Attachment = attachments
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *ChatRepository) GetLastChat(a uuid.UUID, b uuid.UUID) (model.ChatModel, error) {
	// TODO: implement get last chat between user A and B
	return model.ChatModel{}, nil
}

func (r *ChatRepository) MarkConversationAsRead(sender uuid.UUID, receiver uuid.UUID) error {
	// TODO: implement mark chats as read
	return nil
}

func (r *ChatRepository) Delete(id uuid.UUID) error {
	// TODO: implement delete chat by ID
	return nil
}
