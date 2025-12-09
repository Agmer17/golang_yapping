package repository

import (
	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepositoryInterface interface {
	Save(d model.ChatModel) error
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

func (r *ChatRepository) Save(d model.ChatModel) error {
	// TODO: implement insert chat
	return nil
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
