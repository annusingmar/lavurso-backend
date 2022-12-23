package data

import (
	"context"
	"database/sql"
	"errors"
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
	User         *model.Users `json:"user"`
	Read         *bool        `json:"read,omitempty"`
	MessageCount *int         `json:"message_count,omitempty"`
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

func (m MessagingModel) GetUsersInThread(threadID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.UserID.EQ(table.Users.ID))).
		WHERE(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Users.Name.ASC())

	var users []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m MessagingModel) GetGroupsInThread(threadID int) ([]*model.Groups, error) {
	query := postgres.SELECT(table.Groups.ID, table.Groups.Name).
		FROM(table.Groups.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.GroupID.EQ(table.Groups.ID))).
		WHERE(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Groups.Name.ASC())

	var groups []*model.Groups

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m MessagingModel) GetMessageByID(messageID int) (*model.Messages, error) {
	query := postgres.SELECT(table.Messages.AllColumns).
		FROM(table.Messages).
		WHERE(table.Messages.ID.EQ(helpers.PostgresInt(messageID)))

	var message model.Messages

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &message)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchMessage
		default:
			return nil, err
		}
	}

	return &message, nil
}

func (m MessagingModel) InsertMessage(ms *model.Messages) error {
	stmt := table.Messages.INSERT(table.Messages.MutableColumns).
		MODEL(ms).
		RETURNING(table.Messages.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, ms)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) DeleteMessage(messageID int) error {
	stmt := table.Messages.DELETE().
		WHERE(table.Messages.ID.EQ(helpers.PostgresInt(messageID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) UpdateMessage(ms *model.Messages) error {
	stmt := table.Messages.UPDATE(table.Messages.Body, table.Messages.UpdatedAt).
		MODEL(ms).
		WHERE(table.Messages.ID.EQ(helpers.PostgresInt(ms.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) GetAllMessagesByThreadID(threadID int) ([]*NMessage, error) {
	query := postgres.SELECT(table.Messages.AllColumns, table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Messages.
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.Messages.UserID))).
		WHERE(table.Messages.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Messages.CreatedAt.ASC())

	var messages []*NMessage

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &messages)
	if err != nil {
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
