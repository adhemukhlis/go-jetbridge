package credential

import (
	"context"
	"database/sql"
	"fmt"
	"go-jetbridge/gen/jet/public/model"
	"go-jetbridge/gen/jet/public/table"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new instance of the credential repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{
		db: db,
	}
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (model.Credential, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return model.Credential{}, fmt.Errorf("invalid uuid: %w", err)
	}

	var dest model.Credential
	stmt := postgres.SELECT(
		table.Credential.AllColumns,
	).FROM(
		table.Credential,
	).WHERE(
		table.Credential.ID.EQ(postgres.UUID(parsedID)),
	)

	err = stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

func (r *postgresRepository) FindAll(ctx context.Context) ([]model.Credential, error) {
	var dest []model.Credential
	stmt := postgres.SELECT(
		table.Credential.AllColumns,
	).FROM(
		table.Credential,
	)

	err := stmt.QueryContext(ctx, r.db, &dest)
	return dest, err
}

func (r *postgresRepository) Create(ctx context.Context, c model.Credential) (model.Credential, error) {
	if c.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return model.Credential{}, err
		}
		c.ID = newID
	}

	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now

	stmt := table.Credential.INSERT(
		table.Credential.ID,
		table.Credential.CreatedAt,
		table.Credential.UpdatedAt,
		table.Credential.PasswordHash,
		table.Credential.SuspendedAt,
		table.Credential.SuspendReason,
		table.Credential.IsNeedPasswordChange,
		table.Credential.UserId,
	).MODEL(c).RETURNING(table.Credential.AllColumns)

	var created model.Credential
	err := stmt.QueryContext(ctx, r.db, &created)
	if err != nil {
		return model.Credential{}, err
	}

	return created, nil
}

func (r *postgresRepository) Update(ctx context.Context, c model.Credential) (model.Credential, error) {
	c.UpdatedAt = time.Now()

	stmt := table.Credential.UPDATE(
		table.Credential.UpdatedAt,
		table.Credential.PasswordHash,
		table.Credential.SuspendedAt,
		table.Credential.SuspendReason,
		table.Credential.IsNeedPasswordChange,
		table.Credential.UserId,
	).MODEL(c).
		WHERE(table.Credential.ID.EQ(postgres.UUID(c.ID))).
		RETURNING(table.Credential.AllColumns)

	var updated model.Credential
	err := stmt.QueryContext(ctx, r.db, &updated)
	if err != nil {
		return model.Credential{}, err
	}

	return updated, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuid: %w", err)
	}

	stmt := table.Credential.DELETE().WHERE(table.Credential.ID.EQ(postgres.UUID(parsedID)))
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
