package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type PasswordRepo struct {
	db DB
}

func NewPasswordRepo(db DB) (*PasswordRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &PasswordRepo{db: db}, nil
}

func (repo *PasswordRepo) Add(ctx context.Context, pwd *password.Password) error {
	model, err := mapPasswordToModel(pwd)
	if err != nil {
		return fmt.Errorf("map password to model failed: %w", err)
	}

	sql := `INSERT INTO passwords
			(id, user_id, hash, created_at, updated_at)
			VALUES
			(@id, @user_id, @hash, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("password insert failed: %w", err)
	}

	return nil
}

// UpdateHash guards the update with the hash the change was based on, so of two
// concurrent changes of the same password only one is applied. When nothing is
// updated, it checks whether the password exists, so a missing password is not
// reported as a lost race.
//
// See docs/architecture/persistence/repositories.md#guarded-writes.
func (repo *PasswordRepo) UpdateHash(
	ctx context.Context, pwd *password.Password, previous password.Hashed,
) error {
	model, err := mapPasswordToModel(pwd)
	if err != nil {
		return fmt.Errorf("map password to model failed: %w", err)
	}
	if previous.IsZero() {
		return errors.New("previous hash zero")
	}

	sql := `UPDATE passwords
			SET hash = @hash,
				updated_at = @updated_at
			WHERE id = @id AND hash = @previous_hash`
	tag, err := repo.db.Exec(ctx, sql, pgx.NamedArgs{
		"id":            model.ID,
		"hash":          model.Hash,
		"updated_at":    model.UpdatedAt,
		"previous_hash": previous.Value(),
	})

	if err != nil {
		return fmt.Errorf("password hash update failed: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	exists, err := repo.existsByID(ctx, model.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("password not found")
	}

	return password.ErrHashChanged
}

func (repo *PasswordRepo) existsByID(ctx context.Context, id string) (bool, error) {
	sql := "SELECT EXISTS(SELECT 1 FROM passwords WHERE id = $1)"

	var exists bool
	if err := repo.db.QueryRow(ctx, sql, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("password exists query failed: %w", err)
	}

	return exists, nil
}

func (repo *PasswordRepo) FindByID(ctx context.Context, id password.ID) (*password.Password, error) {
	sql := "SELECT * FROM passwords WHERE id = $1"

	return repo.findOne(ctx, sql, id.Value().String())
}

func (repo *PasswordRepo) FindByUserID(
	ctx context.Context, userID user.ID,
) (*password.Password, error) {
	sql := "SELECT * FROM passwords WHERE user_id = $1"

	return repo.findOne(ctx, sql, userID.Value().String())
}

func (repo *PasswordRepo) findOne(
	ctx context.Context, sql string, args ...any,
) (*password.Password, error) {
	rows, err := repo.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[passwordModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	pwd, err := mapPasswordToEntity(model)
	if err != nil {
		return nil, err
	}

	return pwd, nil
}
