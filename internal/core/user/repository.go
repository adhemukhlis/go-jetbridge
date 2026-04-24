package user

import (
	"context"
	"database/sql"
	"fmt"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/jet/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new instance of the user repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (WithRoles, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return WithRoles{}, fmt.Errorf("invalid uuid: %w", err)
	}

	var dest WithRoles
	stmt := postgres.SELECT(
		table.User.AllColumns,
		table.Role.AllColumns,
	).FROM(
		table.User.
			LEFT_JOIN(table.UserRole, table.UserRole.UserId.EQ(table.User.ID)).
			LEFT_JOIN(table.Role, table.Role.ID.EQ(table.UserRole.RoleId)),
	).WHERE(
		table.User.ID.EQ(postgres.UUID(parsedID)),
	)

	err = stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]WithRoles, error) {
	var dest []WithRoles
	stmt := postgres.SELECT(
		table.User.AllColumns,
		table.Role.AllColumns,
	).FROM(
		table.User.
			LEFT_JOIN(table.UserRole, table.UserRole.UserId.EQ(table.User.ID)).
			LEFT_JOIN(table.Role, table.Role.ID.EQ(table.UserRole.RoleId)),
	)

	err := stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

// Create inserts a new user into the database and returns the created user record.
func (r *postgresRepository) Create(ctx context.Context, u model.User) (WithRoles, error) {
	if u.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return WithRoles{}, err
		}
		u.ID = newID
	}

	// Construct the INSERT statement for mutable and primary key columns
	stmt := table.User.INSERT(
		table.User.ID,
		table.User.Name,
		table.User.Email,
		table.User.Username,
	).MODEL(u).RETURNING(table.User.AllColumns)

	var createdUser model.User
	err := stmt.QueryContext(ctx, r.db, &createdUser)
	if err != nil {
		return WithRoles{}, err
	}

	// Return the user with an empty roles slice as roles are not assigned during initial creation.
	return WithRoles{
		User: createdUser,
		Role: []model.Role{},
	}, nil
}

// Update updates an existing user record and returns the updated record.
func (r *postgresRepository) Update(ctx context.Context, u model.User) (WithRoles, error) {
	stmt := table.User.UPDATE(
		table.User.Name,
		table.User.Email,
		table.User.Username,
	).MODEL(u).
		WHERE(table.User.ID.EQ(postgres.UUID(u.ID))).
		RETURNING(table.User.AllColumns)

	var updatedUser model.User
	err := stmt.QueryContext(ctx, r.db, &updatedUser)
	if err != nil {
		return WithRoles{}, err
	}

	// Fetch roles as well, or just return empty for now if not updating roles
	// To be consistent with Create, we can return WithRoles
	return WithRoles{
		User: updatedUser,
		Role: []model.Role{}, // Simplified: roles are not updated here
	}, nil
}

// Delete removes a user record from the database by their ID.
func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	stmt := table.User.DELETE().WHERE(table.User.ID.EQ(postgres.UUID(parsedID)))
	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
