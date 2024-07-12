package platform

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
	SaltKey string
}

type Project struct {
	OrgID       string         `db:"org_id"`
	ID          string         `db:"id"`
	Platform    string         `db:"platform"`
	Credentials string         `db:"credentials"`
	Name        string         `db:"name"`
	Environment string         `db:"environment"`
	Bucket      string         `db:"bucket"`
	DNSZone     sql.NullString `db:"dns_zone"`
	Subdomain   sql.NullString `db:"subdomain"`
	Region      sql.NullString `db:"region"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

func (s Store) CreateProject(ctx context.Context, proj Project) error {
	query := fmt.Sprintf(`INSERT into projects
        (id, org_id, platform, credentials, name, environment, bucket) values
        (:id, :org_id, :platform, pgp_sym_encrypt(:credentials, '%s'), :name, :environment, :bucket)
        ON CONFLICT (org_id, environment)
        DO UPDATE SET
        credentials = pgp_sym_encrypt(:credentials, '%s')
        `, s.SaltKey, s.SaltKey)
	query = s.db.Rebind(query)
	_, err := s.db.NamedExecContext(ctx, query, proj)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) fetchProject(ctx context.Context, oid string, env string) (Project, error) {
	query := fmt.Sprintf(`SELECT id,
                            pgp_sym_decrypt(credentials::bytea, '%s') as credentials,
                            platform,
                            dns_zone,
                            subdomain,
                            region,
                            bucket from projects
                        WHERE org_id = ?
                        AND environment = ?`, s.SaltKey)
	query = s.db.Rebind(query)
	var res Project
	err := s.db.GetContext(ctx, &res, query, oid, env)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return Project{}, cerr.ErrNoProjectFound
	} else if err != nil {
		return Project{}, err
	}
	return res, nil
}

func NewStore(db *sqlx.DB, key string) Store {
	return Store{db, key}
}
