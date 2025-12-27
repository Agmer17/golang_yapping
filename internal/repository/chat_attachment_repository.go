package repository

import (
	"context"
	"errors"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatAttachmentInterface interface {
	Save(m model.ChatAttachment, ctx context.Context) error
	SaveAll(l []model.ChatAttachment, ctx context.Context) error
	Delete(chatId uuid.UUID, ctx context.Context) error
	DeleteAll(chatsId []uuid.UUID, ctx context.Context) error
}

type ChatAttachmentRepository struct {
	Pool *pgxpool.Pool
}

func NewChatAttachmentRepo(pool *pgxpool.Pool) *ChatAttachmentRepository {
	return &ChatAttachmentRepository{
		Pool: pool,
	}

}

func (c *ChatAttachmentRepository) Save(m model.ChatAttachment, ctx context.Context) error {

	query := `
		insert into private_messages_attachment(chat_id, file_name, media_type, size)
		values($1, $2, $3, $4)	
	`

	err := pgx.BeginFunc(ctx, c.Pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			m.ChatId,
			m.FileName,
			m.MediaType,
			m.Size,
		)

		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (c *ChatAttachmentRepository) SaveAll(list []model.ChatAttachment, ctx context.Context) error {

	if len(list) == 0 {
		return errors.New("The len of the list less than 1!")
	}

	err := pgx.BeginFunc(ctx, c.Pool, func(tx pgx.Tx) error {
		rows := make([][]any, 0, len(list))
		for _, m := range list {
			rows = append(rows, []any{
				m.ChatId,
				m.FileName,
				m.MediaType,
				m.Size,
			})
		}

		_, err := tx.CopyFrom(
			ctx,
			pgx.Identifier{"private_messages_attachment"},
			[]string{"chat_id", "file_name", "media_type", "size"},
			pgx.CopyFromRows(rows),
		)

		return err

	})

	return err

}

func (c *ChatAttachmentRepository) Delete(chatId uuid.UUID, ctx context.Context) error {
	return nil
}

func (c *ChatAttachmentRepository) DeleteAll(chatsId []uuid.UUID, ctx context.Context) error {
	return nil
}
