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
	ErrNoSuchGroup   = errors.New("no such group")
	ErrNoSuchGroups  = errors.New("no such groups")
	ErrGroupArchived = errors.New("group is archived")
)

type NGroup struct {
	model.Groups
	MemberCount *int `json:"member_count,omitempty"`
}

type GroupModel struct {
	DB *sql.DB
}

func (m GroupModel) GetGroupByID(groupID int) (*model.Groups, error) {
	query := postgres.SELECT(table.Groups.AllColumns).
		FROM(table.Groups).
		WHERE(table.Groups.ID.EQ(helpers.PostgresInt(groupID)))

	var group model.Groups

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &group)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchGroup
		default:
			return nil, err
		}
	}

	return &group, nil
}

func (m GroupModel) UpdateGroup(g *model.Groups) error {
	stmt := table.Groups.UPDATE(table.Groups.Name, table.Groups.Archived).
		MODEL(g).
		WHERE(table.Groups.ID.EQ(helpers.PostgresInt(g.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) GetAllGroups(archived bool) ([]*NGroup, error) {
	query := postgres.SELECT(table.Groups.AllColumns, postgres.COUNT(table.UsersGroups.UserID).AS("ngroup.member_count")).
		FROM(table.Groups.
			LEFT_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.Groups.ID))).
		WHERE(table.Groups.Archived.EQ(postgres.Bool(archived))).
		GROUP_BY(table.Groups.ID)

	var groups []*NGroup

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m GroupModel) GetAllGroupIDsForUser(userID int) ([]int, error) {
	query := postgres.SELECT(table.Groups.ID).
		FROM(table.Groups.
			INNER_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.Groups.ID))).
		WHERE(table.Groups.Archived.IS_FALSE().
			AND(table.UsersGroups.UserID.EQ(helpers.PostgresInt(userID))))

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m GroupModel) GetUsersByGroupID(groupID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role, table.Users.ClassID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			INNER_JOIN(table.UsersGroups, table.UsersGroups.UserID.EQ(table.Users.ID)).
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.UsersGroups.GroupID.EQ(helpers.PostgresInt(groupID))).
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

func (m GroupModel) GetGroupsByUserID(userID int) ([]*model.Groups, error) {
	query := postgres.SELECT(table.Groups.ID, table.Groups.Name).
		FROM(table.Groups.
			INNER_JOIN(table.UsersGroups, table.UsersGroups.GroupID.EQ(table.Groups.ID))).
		WHERE(table.UsersGroups.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.Groups.Archived.IS_FALSE()))

	var groups []*model.Groups

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &groups)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (m GroupModel) InsertGroup(g *model.Groups) error {
	stmt := table.Groups.INSERT(table.Groups.Name).
		MODEL(g).
		RETURNING(table.Groups.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, g)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) InsertUsersIntoGroup(userIDs []int, groupID int) error {
	var ug []model.UsersGroups
	for _, uid := range userIDs {
		uid := uid
		ug = append(ug, model.UsersGroups{
			UserID:  &uid,
			GroupID: &groupID,
		})
	}

	stmt := table.UsersGroups.INSERT(table.UsersGroups.AllColumns).
		MODELS(ug).
		ON_CONFLICT(table.UsersGroups.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m GroupModel) RemoveUsersFromGroup(userIDs []int, groupID int) error {
	var uids []postgres.Expression
	for _, id := range userIDs {
		uids = append(uids, helpers.PostgresInt(id))
	}

	stmt := table.UsersGroups.DELETE().
		WHERE(table.UsersGroups.UserID.IN(uids...).
			AND(table.UsersGroups.GroupID.EQ(helpers.PostgresInt(groupID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
