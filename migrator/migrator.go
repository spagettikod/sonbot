package migrator

import (
	"database/sql"
	"log/slog"
)

type Migrator interface {
	// Init will set up the Migrator for the current database.
	Init(db *sql.DB) error
	// Initialized will check if the Migrator is setup in this database.
	Initialized(db *sql.DB) (bool, error)
	// Version returns the current version from the database.
	Version() (int, error)
	// SetVersion updates the current version in the database.
	SetVersion(version int) error
	// Migrate will run the forward migrations in the array.
	Migrate(migrations []string) error
}

type SqliteMigrator struct {
	db *sql.DB
}

func NewSqliteMigrator(db *sql.DB) SqliteMigrator {
	return SqliteMigrator{db: db}
}

// Init will set up the Migrator for the current database. If already initialized it does nothing.
func (sm SqliteMigrator) Init() error {
	initialized, err := sm.Initialized()
	if err != nil {
		return err
	}
	if !initialized {
		sql := `CREATE TABLE _migrator_ (
			version INTEGER NOT NULL
		) STRICT`
		_, err = sm.db.Exec(sql)
	}
	return err
}

// Initialized will check if the Migrator is setup in this database.
func (sm SqliteMigrator) Initialized() (bool, error) {
	sql := `SELECT name
			FROM sqlite_master
			WHERE type='table'
			AND name=?1`
	rows, err := sm.db.Query(sql, "_migrator_")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		return true, nil
	}
	return false, nil
}

// Version returns the current version from the database.
func (sm SqliteMigrator) Version() (int, error) {
	query := `SELECT version FROM _migrator_`
	row := sm.db.QueryRow(query)

	version := -1
	if err := row.Scan(&version); err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return 0, err
	}
	return version, nil
}

// SetVersion updates the current version in the database.
func (sm SqliteMigrator) SetVersion(version int) error {
	sql := `REPLACE INTO _migrator_ (version) VALUES (?1)`
	_, err := sm.db.Exec(sql, version)
	return err
}

// Migrate will run the forward migrations in the array.
func (sm SqliteMigrator) Migrate(migrations []string) error {
	v, err := sm.Version()
	if err != nil {
		return err
	}
	// if no migrations have run we're at version -1, to kickstart migrations we must start at v==0
	if v == -1 {
		v = 0
	}
	slog.Debug("migration check", "current_version", v, "available_migrations", len(migrations)-v)
	for i := v; i < len(migrations); i++ {
		slog.Debug("migrating", "current_version", i, "migrating_to_version", i+1)
		if _, err := sm.db.Exec(migrations[i]); err != nil {
			return err
		}
		sm.SetVersion(i)
	}
	return nil
}
