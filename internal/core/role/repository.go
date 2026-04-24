package role

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

// NewRepository creates a new instance of the role repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (model.Role, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return model.Role{}, fmt.Errorf("invalid uuid: %w", err)
	}

	var dest model.Role
	stmt := postgres.SELECT(
		table.Role.AllColumns,
	).FROM(
		table.Role,
	).WHERE(
		table.Role.ID.EQ(postgres.UUID(parsedID)),
	)

	err = stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]model.Role, error) {
	var dest []model.Role
	stmt := postgres.SELECT(
		table.Role.AllColumns,
	).FROM(
		table.Role,
	)

	err := stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

func (r *postgresRepository) Create(ctx context.Context, m model.Role) (model.Role, error) {
	if m.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return model.Role{}, err
		}
		m.ID = newID
	}

	stmt := table.Role.INSERT(
		table.Role.ID,
		table.Role.Key,
		table.Role.Name,
	).MODEL(m).RETURNING(table.Role.AllColumns)

	var createdRole model.Role
	err := stmt.QueryContext(ctx, r.db, &createdRole)
	return createdRole, err
}

func (r *postgresRepository) Update(ctx context.Context, m model.Role) (model.Role, error) {
	stmt := table.Role.UPDATE(
		table.Role.Key,
		table.Role.Name,
	).MODEL(m).
		WHERE(table.Role.ID.EQ(postgres.UUID(m.ID))).
		RETURNING(table.Role.AllColumns)

	var updatedRole model.Role
	err := stmt.QueryContext(ctx, r.db, &updatedRole)
	return updatedRole, err
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	stmt := table.Role.DELETE().WHERE(table.Role.ID.EQ(postgres.UUID(parsedID)))
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
