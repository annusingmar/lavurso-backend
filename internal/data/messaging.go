package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrNoSuchThread           = errors.New("no such thread")
	ErrNoSuchMessage          = errors.New("no such message")
	ErrThreadAlreadyLocked    = errors.New("thread already locked")
	ErrThreadAlreadyUnlocked  = errors.New("thread already unlocked")
	ErrCantDeleteFirstMessage = errors.New("can't delete first message of thread")
)

const (
	MsgTypeNormal      = "normal"
	MsgTypeThreadStart = "thread_start"
)

type Thread struct {
	ID           int       `json:"id"`
	User         *User     `json:"user"`
	Title        string    `json:"title"`
	Locked       bool      `json:"locked"`
	MessageCount *int      `json:"message_count,omitempty"`
	Read         *bool     `json:"read,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	User      *User     `json:"user"`
	Body      string    `json:"body,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NThread struct {
	model.Threads
	User *model.Users `json:"user"`
	Read *bool        `json:"read,omitempty"`
}

type NMessage struct {
	model.Messages
	User *model.Users `json:"user"`
}

type MessagingModel struct {
	DB *sql.DB
}

func (m MessagingModel) GetThreadByID(threadID int) (*NThread, error) {
	query := postgres.SELECT(table.Threads.AllColumns, table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Threads.
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.Threads.UserID))).
		WHERE(table.Threads.ID.EQ(helpers.PostgresInt(threadID)))

	var thread NThread

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &thread)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchThread
		default:
			return nil, err
		}
	}

	return &thread, nil
}

func (m MessagingModel) InsertThread(t *model.Threads) error {
	stmt := table.Threads.INSERT(table.Threads.MutableColumns).
		MODEL(t).
		RETURNING(table.Threads.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, t)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) DeleteThread(threadID int) error {
	stmt := table.Threads.DELETE().
		WHERE(table.Threads.ID.EQ(helpers.PostgresInt(threadID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) AddUsersToThread(threadID int, userIDs []int) error {
	var ut []model.ThreadsRecipients
	for _, uid := range userIDs {
		uid := uid
		ut = append(ut, model.ThreadsRecipients{
			UserID:   &uid,
			ThreadID: &threadID,
		})
	}

	stmt := table.ThreadsRecipients.INSERT(table.ThreadsRecipients.UserID, table.ThreadsRecipients.ThreadID).
		MODELS(ut).
		ON_CONFLICT(table.ThreadsRecipients.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) RemoveUsersFromThread(threadID int, userIDs []int) error {
	var uids []postgres.Expression
	for _, id := range userIDs {
		uids = append(uids, helpers.PostgresInt(id))
	}

	stmt := table.ThreadsRecipients.DELETE().
		WHERE(table.ThreadsRecipients.UserID.IN(uids...).
			AND(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) AddGroupsToThread(threadID int, groupIDs []int) error {
	var gt []model.ThreadsRecipients
	for _, gid := range groupIDs {
		gid := gid
		gt = append(gt, model.ThreadsRecipients{
			GroupID:  &gid,
			ThreadID: &threadID,
		})
	}

	stmt := table.ThreadsRecipients.INSERT(table.ThreadsRecipients.GroupID, table.ThreadsRecipients.ThreadID).
		MODELS(gt).
		ON_CONFLICT(table.ThreadsRecipients.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) RemoveGroupsFromThread(threadID int, groupIDs []int) error {
	var gids []postgres.Expression
	for _, id := range groupIDs {
		gids = append(gids, helpers.PostgresInt(id))
	}

	stmt := table.ThreadsRecipients.DELETE().
		WHERE(table.ThreadsRecipients.GroupID.IN(gids...).
			AND(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) GetUsersInThread(threadID int) ([]*User, error) {
	query := `SELECT tr.user_id, u.name, u.role
	FROM threads_recipients tr
	INNER JOIN users u
	ON tr.user_id = u.id
	WHERE tr.thread_id = $1 and tr.group_id is NULL
	ORDER BY u.name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, threadID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m MessagingModel) GetGroupsInThread(threadID int) ([]*Group, error) {
	query := `SELECT tr.group_id, g.name
	FROM threads_recipients tr
	INNER JOIN groups g
	ON tr.group_id = g.id
	WHERE tr.thread_id = $1 and tr.group_id is NOT NULL
	ORDER BY g.name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, threadID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var groups []*Group

	for rows.Next() {
		var group Group
		err := rows.Scan(
			&group.ID,
			&group.Name,
		)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (m MessagingModel) GetMessageByID(messageID int) (*Message, error) {
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.type, m.created_at, m.updated_at
	FROM messages m
	INNER JOIN users u
	ON m.user_id = u.id
	WHERE m.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var message Message
	message.User = new(User)

	err := m.DB.QueryRowContext(ctx, query, messageID).Scan(
		&message.ID,
		&message.ThreadID,
		&message.User.ID,
		&message.User.Name,
		&message.User.Role,
		&message.Body,
		&message.Type,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchMessage
		default:
			return nil, err
		}
	}

	return &message, nil
}

func (m MessagingModel) InsertMessage(ms *Message) error {
	stmt := `INSERT INTO messages
	(thread_id, user_id, body, type, created_at, updated_at)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, ms.ThreadID, ms.User.ID, ms.Body, ms.Type, ms.CreatedAt, ms.UpdatedAt).Scan(&ms.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) DeleteMessage(messageID int) error {
	stmt := `DELETE FROM messages
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, messageID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) UpdateMessage(ms *Message) error {
	stmt := `UPDATE messages
	SET (body, updated_at)
	= ($1, $2)
	WHERE id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, ms.Body, ms.UpdatedAt, ms.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m MessagingModel) GetAllMessagesByThreadID(threadID int) ([]*Message, error) {
	query := `SELECT m.id, m.thread_id, m.user_id, u.name, u.role, m.body, m.type, m.created_at, m.updated_at
	FROM messages m
	INNER JOIN users u
	ON m.user_id = u.id
	WHERE thread_id = $1
	ORDER BY created_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, threadID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []*Message

	for rows.Next() {
		var message Message
		message.User = new(User)

		err = rows.Scan(
			&message.ID,
			&message.ThreadID,
			&message.User.ID,
			&message.User.Name,
			&message.User.Role,
			&message.Body,
			&message.Type,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (m MessagingModel) GetThreadsForUser(userID int, search string) ([]*Thread, error) {
	baseQuery := `SELECT DISTINCT t.id, t.user_id, u.name, u.role, t.title, t.locked, (CASE WHEN r.user_id is NOT NULL THEN TRUE ELSE FALSE END), t.created_at, t.updated_at, (SELECT COUNT(id) FROM messages WHERE thread_id = t.id)
	FROM threads t
	INNER JOIN threads_recipients tr
	ON t.id = tr.thread_id
    LEFT JOIN users_groups ug
    ON ug.group_id = tr.group_id AND ug.user_id = $1
    LEFT JOIN threads_read r
    ON r.thread_id = t.id AND r.user_id = $1
    INNER JOIN users u
	ON t.user_id = u.id
	INNER JOIN messages m
	ON m.thread_id = t.id
    WHERE (tr.user_id = $1 OR ug.user_id = $1)%s
	ORDER BY updated_at DESC`

	var query string

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rows *sql.Rows
	var err error

	if search != "" {
		query = fmt.Sprintf(baseQuery, " AND ((to_tsvector('simple', t.title) @@ plainto_tsquery('simple', $2)) OR (to_tsvector('simple', m.body) @@ plainto_tsquery('simple', $2)))")
		rows, err = m.DB.QueryContext(ctx, query, userID, search)
	} else {
		query = fmt.Sprintf(baseQuery, "")
		rows, err = m.DB.QueryContext(ctx, query, userID)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var threads []*Thread

	for rows.Next() {
		var thread Thread
		thread.User = new(User)

		err = rows.Scan(
			&thread.ID,
			&thread.User.ID,
			&thread.User.Name,
			&thread.User.Role,
			&thread.Title,
			&thread.Locked,
			&thread.Read,
			&thread.CreatedAt,
			&thread.UpdatedAt,
			&thread.MessageCount,
		)
		if err != nil {
			return nil, err
		}

		threads = append(threads, &thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

func (m MessagingModel) DoesUserHaveUnread(userID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM threads t
	INNER JOIN threads_recipients tr
	ON t.id = tr.thread_id
    LEFT JOIN users_groups ug
    ON ug.group_id = tr.group_id AND ug.user_id = $1
    LEFT JOIN threads_read r
    ON r.thread_id = t.id AND r.user_id = $1
    WHERE ( tr.user_id = $1 OR ug.user_id = $1 ) AND r.user_id IS NULL`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (m MessagingModel) IsUserInThread(userID, threadID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM threads t
	INNER JOIN threads_recipients tr
	ON t.id = tr.thread_id
    LEFT JOIN users_groups ug
    ON ug.group_id = tr.group_id AND ug.user_id = $1
    WHERE ( tr.user_id = $1 OR ug.user_id = $1 ) AND t.id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, userID, threadID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (m MessagingModel) SetThreadAsReadForUser(threadID, userID int) error {
	stmt := `INSERT INTO threads_read
	(thread_id, user_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, threadID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadAsUnreadForAll(threadID int) error {
	stmt := `DELETE FROM threads_read
	WHERE thread_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, threadID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadLocked(threadID int, locked bool) error {
	stmt := `UPDATE threads
	SET locked = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, locked, threadID)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadUpdatedAt(threadID int) error {
	stmt := `UPDATE threads
	SET updated_at = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, time.Now().UTC(), threadID)
	if err != nil {
		return err
	}

	return nil
}
