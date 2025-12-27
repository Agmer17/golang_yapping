package repository

import (
	"context"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepositoryInterface interface {
	Save(d model.ChatModel, ctx context.Context) (model.ChatModel, error)
	GetChatBeetween(r uuid.UUID, s uuid.UUID) ([]model.ChatModel, error)
	GetLastChat(r uuid.UUID, s uuid.UUID) (model.ChatModel, error)
	MarkConversationAsRead(sender uuid.UUID, receiver uuid.UUID) error
	Delete(id uuid.UUID) error
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

func (r *ChatRepository) GetChatBeetween(a uuid.UUID, b uuid.UUID) ([]model.ChatModel, error) {
	// TODO: implement get chats between user A and B
	return []model.ChatModel{}, nil
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
