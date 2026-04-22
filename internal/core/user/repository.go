package user

import (
	"context"
	"database/sql"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/jet/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type User struct {
	DB *sql.DB
}

func (r *User) FindByID(ctx context.Context, id string) (WithRoles, error) {
	var dest WithRoles
	stmt := postgres.SELECT(
		table.User.AllColumns,
		table.Role.AllColumns,
	).FROM(
		table.User.
			LEFT_JOIN(table.UserRole, table.UserRole.UserId.EQ(table.User.ID)).
			LEFT_JOIN(table.Role, table.Role.ID.EQ(table.UserRole.RoleId)),
	).WHERE(
		table.User.ID.EQ(postgres.CAST(postgres.String(id)).AS_UUID()),
	)

	err := stmt.QueryContext(ctx, r.DB, &dest)
	return dest, err
}

func (r *User) FindAll(ctx context.Context) ([]WithRoles, error) {
	var dest []WithRoles
	stmt := postgres.SELECT(
		table.User.AllColumns,
		table.Role.AllColumns,
	).FROM(
		table.User.
			LEFT_JOIN(table.UserRole, table.UserRole.UserId.EQ(table.User.ID)).
			LEFT_JOIN(table.Role, table.Role.ID.EQ(table.UserRole.RoleId)),
	)

	err := stmt.QueryContext(ctx, r.DB, &dest)
	return dest, err
}

// Create inserts a new user into the database and returns the created user record.
func (r *User) Create(ctx context.Context, u model.User) (WithRoles, error) {
	// Generate a new UUID if not provided
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Construct the INSERT statement for mutable and primary key columns
	stmt := table.User.INSERT(
		table.User.ID,
		table.User.Name,
		table.User.Email,
		table.User.Username,
	).MODEL(u).RETURNING(table.User.AllColumns)

	var createdUser model.User
	err := stmt.QueryContext(ctx, r.DB, &createdUser)
	if err != nil {
		return WithRoles{}, err
	}

	// Return the user with an empty roles slice as roles are not assigned during initial creation.
	return WithRoles{
		User: createdUser,
		Role: []model.Role{},
	}, nil
}
