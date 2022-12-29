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

type Thread = model.Threads

type ThreadExt struct {
	Thread
	User         *User `json:"user"`
	Read         *bool `json:"read,omitempty"`
	MessageCount *int  `json:"message_count,omitempty"`
}

type Message = model.Messages

type MessageExt struct {
	Message
	User *User `json:"user"`
}

type MessagingModel struct {
	DB *sql.DB
}

func (m MessagingModel) GetThreadByID(threadID int) (*ThreadExt, error) {
	query := postgres.SELECT(table.Threads.AllColumns, table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Threads.
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.Threads.UserID))).
		WHERE(table.Threads.ID.EQ(helpers.PostgresInt(threadID)))

	var thread ThreadExt

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

func (m MessagingModel) InsertThread(t *Thread) error {
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

func (m MessagingModel) GetUsersInThread(threadID int) ([]*UserExt, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.UserID.EQ(table.Users.ID))).
		WHERE(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Users.Name.ASC())

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m MessagingModel) GetGroupsInThread(threadID int) ([]*Group, error) {
	query := postgres.SELECT(table.Groups.ID, table.Groups.Name).
		FROM(table.Groups.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.GroupID.EQ(table.Groups.ID))).
		WHERE(table.ThreadsRecipients.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Groups.Name.ASC())

	var groups []*Group

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m MessagingModel) GetMessageByID(messageID int) (*Message, error) {
	query := postgres.SELECT(table.Messages.AllColumns).
		FROM(table.Messages).
		WHERE(table.Messages.ID.EQ(helpers.PostgresInt(messageID)))

	var message Message

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

func (m MessagingModel) InsertMessage(ms *Message) error {
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

func (m MessagingModel) UpdateMessage(ms *Message) error {
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

func (m MessagingModel) GetAllMessagesByThreadID(threadID int) ([]*MessageExt, error) {
	query := postgres.SELECT(table.Messages.AllColumns, table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Messages.
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.Messages.UserID))).
		WHERE(table.Messages.ThreadID.EQ(helpers.PostgresInt(threadID))).
		ORDER_BY(table.Messages.CreatedAt.ASC())

	var messages []*MessageExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (m MessagingModel) GetThreadsForUser(userID int, search string) ([]*ThreadExt, error) {
	uid := helpers.PostgresInt(userID)

	from := table.Threads.
		INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.ThreadID.EQ(table.Threads.ID)).
		LEFT_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.ThreadsRecipients.GroupID).
			AND(table.UsersGroups.UserID.EQ(uid))).
		LEFT_JOIN(table.ThreadsRead, table.ThreadsRead.ThreadID.EQ(table.Threads.ID).
			AND(table.ThreadsRead.UserID.EQ(uid))).
		INNER_JOIN(table.Users, table.Users.ID.EQ(table.Threads.UserID))

	var where postgres.BoolExpression
	if search != "" {
		from = from.LEFT_JOIN(table.Messages, table.Messages.ThreadID.EQ(table.Threads.ID))
		where = postgres.AND(
			table.ThreadsRecipients.UserID.EQ(uid).OR(table.UsersGroups.UserID.EQ(uid)),
			postgres.CAST(
				postgres.Raw("(to_tsvector('simple', threads.title) @@ plainto_tsquery('simple', #search)) OR (to_tsvector('simple', messages.body) @@ plainto_tsquery('simple', #search))",
					postgres.RawArgs{"#search": search}),
			).AS_BOOL(),
		)
	} else {
		where = table.ThreadsRecipients.UserID.EQ(uid).OR(table.UsersGroups.UserID.EQ(uid))
	}

	query := postgres.SELECT(
		table.Threads.ID, table.Threads.UserID, table.Threads.Title, table.Threads.Locked, table.Threads.CreatedAt, table.Threads.UpdatedAt,
		table.Users.ID, table.Users.Name, table.Users.Role,
		postgres.CASE().WHEN(table.ThreadsRead.UserID.IS_NOT_NULL()).THEN(postgres.Bool(true)).ELSE(postgres.Bool(false)).AS("threadext.read"),
		postgres.SELECT(postgres.COUNT(table.Messages.ID)).FROM(table.Messages).WHERE(table.Messages.ThreadID.EQ(table.Threads.ID)).AS("threadext.message_count"),
	).DISTINCT().
		FROM(from).
		WHERE(where).
		ORDER_BY(table.Threads.UpdatedAt.DESC())

	var threads []*ThreadExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &threads)
	if err != nil {
		return nil, err
	}

	return threads, nil
}

func (m MessagingModel) DoesUserHaveUnread(userID int) (bool, error) {
	uid := helpers.PostgresInt(userID)

	query := postgres.SELECT(postgres.COUNT(postgres.STAR)).
		FROM(table.Threads.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.ThreadID.EQ(table.Threads.ID)).
			LEFT_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.ThreadsRecipients.GroupID).
				AND(table.UsersGroups.UserID.EQ(uid))).
			LEFT_JOIN(table.ThreadsRead, table.ThreadsRead.ThreadID.EQ(table.Threads.ID).
				AND(table.ThreadsRead.UserID.EQ(uid)))).
		WHERE(postgres.AND(
			table.ThreadsRecipients.UserID.EQ(uid).OR(table.UsersGroups.UserID.EQ(uid)),
			table.ThreadsRead.UserID.IS_NULL(),
		))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m MessagingModel) IsUserInThread(userID, threadID int) (bool, error) {
	uid := helpers.PostgresInt(userID)

	query := postgres.SELECT(postgres.COUNT(postgres.STAR)).
		FROM(table.Threads.
			INNER_JOIN(table.ThreadsRecipients, table.ThreadsRecipients.ThreadID.EQ(table.Threads.ID)).
			LEFT_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.ThreadsRecipients.GroupID).
				AND(table.UsersGroups.UserID.EQ(uid)))).
		WHERE(postgres.AND(
			table.ThreadsRecipients.UserID.EQ(uid).OR(table.UsersGroups.UserID.EQ(uid)),
			table.Threads.ID.EQ(helpers.PostgresInt(threadID)),
		))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m MessagingModel) SetThreadAsReadForUser(threadID, userID int) error {
	stmt := table.ThreadsRead.INSERT(table.ThreadsRead.AllColumns).
		MODEL(model.ThreadsRead{ThreadID: &threadID, UserID: &userID}).
		ON_CONFLICT(table.ThreadsRead.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadAsUnreadForAll(threadID int) error {
	stmt := table.ThreadsRead.DELETE().
		WHERE(table.ThreadsRead.ThreadID.EQ(helpers.PostgresInt(threadID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadLocked(threadID int, locked bool) error {
	stmt := table.Threads.UPDATE(table.Threads.Locked).
		SET(locked).
		WHERE(table.Threads.ID.EQ(helpers.PostgresInt(threadID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m MessagingModel) SetThreadUpdatedAt(threadID int) error {
	stmt := table.Threads.UPDATE(table.Threads.UpdatedAt).
		SET(time.Now().UTC()).
		WHERE(table.Threads.ID.EQ(helpers.PostgresInt(threadID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
