package role

import (
	"context"
	"database/sql"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/jet/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type Role struct {
	DB *sql.DB
}

func (r *Role) FindByID(ctx context.Context, id string) (model.Role, error) {
	var dest model.Role
	stmt := postgres.SELECT(
		table.Role.AllColumns,
	).FROM(
		table.Role,
	).WHERE(
		table.Role.ID.EQ(postgres.CAST(postgres.String(id)).AS_UUID()),
	)

	err := stmt.QueryContext(ctx, r.DB, &dest)
	return dest, err
}

func (r *Role) FindAll(ctx context.Context) ([]model.Role, error) {
	var dest []model.Role
	stmt := postgres.SELECT(
		table.Role.AllColumns,
	).FROM(
		table.Role,
	)

	err := stmt.QueryContext(ctx, r.DB, &dest)
	return dest, err
}

func (r *Role) Create(ctx context.Context, m model.Role) (model.Role, error) {
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
	err := stmt.QueryContext(ctx, r.DB, &createdRole)
	return createdRole, err
}

func (r *Role) Update(ctx context.Context, m model.Role) (model.Role, error) {
	stmt := table.Role.UPDATE(
		table.Role.Key,
		table.Role.Name,
	).MODEL(m).
		WHERE(table.Role.ID.EQ(postgres.CAST(postgres.String(m.ID.String())).AS_UUID())).
		RETURNING(table.Role.AllColumns)

	var updatedRole model.Role
	err := stmt.QueryContext(ctx, r.DB, &updatedRole)
	return updatedRole, err
}

func (r *Role) Delete(ctx context.Context, id string) error {
	stmt := table.Role.DELETE().WHERE(table.Role.ID.EQ(postgres.CAST(postgres.String(id)).AS_UUID()))
	res, err := stmt.ExecContext(ctx, r.DB)
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
