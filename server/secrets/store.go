package secrets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scalescape/dolores/server/cerr"
)

type Store struct {
	db      *sqlx.DB
	saltKey string
}

type Secret struct {
	ID        string    `db:"id" json:"id,omitempty"`
	ProjectID string    `db:"project_id" json:"project_id,omitempty"`
	Name      string    `db:"name" json:"name"`
	Location  string    `db:"location" json:"location"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Key struct {
	UserID      string         `db:"user_id"`
	ProjectID   string         `db:"project_id"`
	Environment string         `db:"environment"`
	PublicKey   string         `db:"public_key"`
	PrivateKey  sql.NullString `db:"private_key"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (s Store) ListUsersPublicKeys(ctx context.Context, env, orgID string) ([]Key, error) {
	query := `select user_id, public_key from user_keys
        where user_id in (
            select id from users where org_id = ?
        ) and
        environment = ?`
	query = s.db.Rebind(query)
	var result []Key
	err := s.db.SelectContext(ctx, &result, query, orgID, env)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error finding user_keys for org %s: %w", orgID, err)
	} else if err != nil {
		return nil, fmt.Errorf("error fetching user_keys for org %s: %w", orgID, err)
	}
	return result, nil
}

func (s Store) SaveSecret(ctx context.Context, sec Secret) error {
	query := `INSERT into secrets
        (id, project_id, location, name, created_at, updated_at) values
        (:id, :project_id, :location, :name, :created_at, now())
        ON CONFLICT (name)
        DO UPDATE SET updated_at = now()
    `
	query = s.db.Rebind(query)
	if _, err := s.db.NamedExecContext(ctx, query, sec); err != nil {
		return err
	}
	return nil
}

func (s Store) fetchSecret(ctx context.Context, name, pid string) (Secret, error) {
	query := `select * from secrets where
        name = ? and
        project_id = ?
        `
	query = s.db.Rebind(query)
	var result Secret
	err := s.db.GetContext(ctx, &result, query, name, pid)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return Secret{}, fmt.Errorf("error finding secret for %s: %w", name, cerr.ErrNoSecretFound)
	} else if err != nil {
		return Secret{}, fmt.Errorf("error fetching secret for %s: %w", name, cerr.ErrNoSecretFound)
	}
	return result, nil
}

func NewStore(db *sqlx.DB, key string) Store {
	return Store{db, key}
}
