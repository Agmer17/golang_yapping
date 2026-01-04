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

type LatestChatQuery struct {
	ChatData    model.ChatModel
	PartnerData model.User
}

type ChatRepositoryInterface interface {
	Save(d model.ChatModel, ctx context.Context) (model.ChatModel, error)
	GetChatBeetween(ctx context.Context, r uuid.UUID, s uuid.UUID) ([]model.ChatModel, error)
	GetLastChat(ctx context.Context, userId uuid.UUID) ([]LatestChatQuery, error)
	MarkConversationAsRead(sender uuid.UUID, receiver uuid.UUID) error
	GetChatById(ctx context.Context, chatId uuid.UUID) (model.ChatModel, error)
	Delete(ctx context.Context, id uuid.UUID) ([]string, error)
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

func (r *ChatRepository) GetLastChat(ctx context.Context, userID uuid.UUID) ([]LatestChatQuery, error) {
	query := `
		SELECT DISTINCT ON (chat_partner_id)
			pm.id                          AS chat_id,
			pm.sender_id,
			pm.receiver_id,
			pm.reply_to,
			pm.chat_text,
			pm.post_id,
			pm.is_read,
			pm.created_at AT TIME ZONE 'UTC' AS created_at,

			(pm.sender_id = $1)            AS is_own,

			-- partner
			chat_partner_id                AS partner_id,
			u.full_name                    AS partner_fullname,
			u.username                     AS partner_username,
			u.profile_picture              AS partner_profile_picture,

			-- attachments as JSON
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
				'[]'::json
			) AS attachments

		FROM (
			SELECT
				pm.*,
				CASE
					WHEN pm.sender_id = $1 THEN pm.receiver_id
					ELSE pm.sender_id
				END AS chat_partner_id
			FROM private_messages pm
			WHERE
				pm.sender_id = $1
				OR pm.receiver_id = $1
		) pm
		JOIN users u ON u.id = pm.chat_partner_id
		ORDER BY chat_partner_id, pm.created_at DESC;
	`

	rows, err := r.Pool.Query(ctx, query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var queryData []LatestChatQuery = make([]LatestChatQuery, 0)

	for rows.Next() {

		var chatData model.ChatModel
		var partnerData model.User
		var rawAttachments json.RawMessage

		err := rows.Scan(
			&chatData.Id,
			&chatData.SenderId,
			&chatData.ReceiverId,
			&chatData.ReplyTo,
			&chatData.ChatText,
			&chatData.PostId,
			&chatData.IsRead,
			&chatData.CreatedAt,
			&chatData.IsOwn,

			&partnerData.Id,
			&partnerData.FullName,
			&partnerData.Username,
			&partnerData.ProfilePicture,

			&rawAttachments,
		)
		if err != nil {
			return nil, err
		}

		var attachments []model.ChatAttachment
		if err := json.Unmarshal(rawAttachments, &attachments); err != nil {
			return nil, err
		}

		chatData.Attachment = attachments
		queryData = append(queryData, LatestChatQuery{
			ChatData:    chatData,
			PartnerData: partnerData,
		})
	}

	return queryData, nil

}

func (r *ChatRepository) MarkConversationAsRead(sender uuid.UUID, receiver uuid.UUID) error {
	// TODO: implement mark chats as read
	return nil
}

func (r *ChatRepository) GetChatById(ctx context.Context, chatId uuid.UUID) (model.ChatModel, error) {

	query := `
	select 
		pm.id,
		pm.sender_id,
		pm.receiver_id,
		pm.reply_to,
		pm.chat_text,
		pm.post_id,
		pm.is_read,
		pm.created_at 
	from private_messages pm
	where pm.id = $1; 
	`
	var tmpData model.ChatModel

	row := r.Pool.QueryRow(ctx, query, chatId)

	err := row.Scan(
		&tmpData.Id,
		&tmpData.SenderId,
		&tmpData.ReceiverId,
		&tmpData.ReplyTo,
		&tmpData.ChatText,
		&tmpData.PostId,
		&tmpData.IsRead,
		&tmpData.CreatedAt,
	)

	if err != nil {
		return model.ChatModel{}, err
	}

	return tmpData, nil

}

func (r *ChatRepository) Delete(ctx context.Context, id uuid.UUID) ([]string, error) {
	// TODO: implement delete chat by ID
	query := `
	WITH deleted_files AS (
		DELETE FROM private_messages_attachment
		WHERE chat_id = $1
		RETURNING file_name
	),
	delete_chat AS (
		DELETE FROM private_messages
		WHERE id = $1
	)
	SELECT file_name FROM deleted_files;
	`

	var filesToDelete []string = make([]string, 0)

	err := pgx.BeginFunc(ctx, r.Pool, func(tx pgx.Tx) error {

		rows, err := r.Pool.Query(ctx, query, id)

		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var fname string
			if err := rows.Scan(&fname); err != nil {
				return err
			}

			filesToDelete = append(filesToDelete, fname)
		}

		if err = rows.Err(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filesToDelete, nil
}
