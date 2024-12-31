package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spagettikod/sonbot/energy"
	"github.com/spagettikod/sonbot/migrator"
)

type SQLiteStore struct {
	DB *sql.DB
}

func NewSQLiteStore(file string) (SQLiteStore, error) {
	slog.Debug("opening database", "file", file)
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return SQLiteStore{}, err
	}
	store := SQLiteStore{DB: db}
	mg := migrator.NewSqliteMigrator(db)
	if err := mg.Init(); err != nil {
		return store, err
	}
	return store, mg.Migrate(migrations)
}

var (
	migrations = []string{
		`CREATE TABLE migration (
			next_version INTEGER NOT NULL
		) STRICT`,
		`CREATE TABLE ts_sek_per_kwh_se3 (
			timestamp INTEGER PRIMARY KEY,
			value REAL NOT NULL
		) STRICT`,
		`CREATE TABLE ts_sek_per_kwh_se1 (
			timestamp INTEGER PRIMARY KEY,
			value REAL NOT NULL
		) STRICT`,
		`CREATE TABLE ts_sek_per_kwh_se2 (
			timestamp INTEGER PRIMARY KEY,
			value REAL NOT NULL
		) STRICT`,
		`CREATE TABLE ts_sek_per_kwh_se4 (
			timestamp INTEGER PRIMARY KEY,
			value REAL NOT NULL
		) STRICT`,
		`CREATE TABLE ts_consumption (
			timestamp INTEGER PRIMARY KEY,
			value REAL NOT NULL
		) STRICT`,
		`CREATE TABLE ts_hourly_bucket (
			timestamp INTEGER PRIMARY KEY,
			area_code REAL NOT NULL,
			price REAL NOT NULL,
			exchange_rate_sek REAL NOT NULL,
			consumption REAL NOT NULL,
			production REAL NOT NULL
		) STRICT`,
	}
)

func ObsToAny(obs Observation) []any {
	return []any{obs.Timestamp.UTC().Unix(), obs.Value}
}

func table(sql, table, areaCode string) string {
	fullTable := fmt.Sprintf("%s_%s", table, strings.ToLower(areaCode))
	return strings.ReplaceAll(sql, "<TABLE_NAME>", fullTable)
}

func (store SQLiteStore) PutSekPerKwh(areaCode string, observations []Observation) error {
	sd := [][]any{}
	for _, obs := range observations {
		sd = append(sd, ObsToAny(obs))
	}
	sql := "REPLACE INTO <TABLE_NAME> (timestamp, value) VALUES (?1, ?2)"
	return store.doUpsert(table(sql, "ts_sek_per_kwh", areaCode), sd)
}

func (store SQLiteStore) GetSekPerKwh(areaCode string, from, to time.Time) ([]Observation, error) {
	obs := []Observation{}
	sql := `SELECT  timestamp,
					value
			FROM <TABLE_NAME>
			WHERE timestamp BETWEEN ?1 AND ?2
			ORDER BY timestamp`
	rows, err := store.DB.Query(table(sql, "ts_sek_per_kwh", areaCode), from.Unix(), to.Unix())
	if err != nil {
		return obs, err
	}
	defer rows.Close()

	for rows.Next() {
		ts := int64(0)
		value := float64(0)
		err = rows.Scan(&ts, &value)
		if err != nil {
			return obs, err
		}
		obs = append(obs, NewObservation(time.Unix(ts, 0).UTC(), value))
	}
	return obs, nil
}

func (store SQLiteStore) PutConsumption(stat energy.Stat) error {
	sd := [][]any{}
	sdStat := []any{stat.Timestamp.UTC().Unix(), stat.Consumption}
	sd = append(sd, sdStat)
	sql := "REPLACE INTO ts_consumption (timestamp, value) VALUES (?1, ?2)"
	return store.doUpsert(sql, sd)
}

func (store SQLiteStore) GetConsumption(from, to time.Time) ([]Observation, error) {
	obs := []Observation{}
	sql := `SELECT  timestamp,
					value
			FROM ts_consumption
			WHERE timestamp BETWEEN ?1 AND ?2
			ORDER BY timestamp`
	rows, err := store.DB.Query(sql, from.Unix(), to.Unix())
	if err != nil {
		return obs, err
	}
	defer rows.Close()

	for rows.Next() {
		ts := int64(0)
		value := float64(0)
		err = rows.Scan(&ts, &value)
		if err != nil {
			return obs, err
		}
		obs = append(obs, NewObservation(time.Unix(ts, 0).UTC(), value))
	}
	return obs, nil
}

func (store SQLiteStore) doUpsert(sql string, data [][]any) error {
	tx, err := store.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		tx.Rollback()
	}()

	for _, val := range data {
		_, err := tx.Exec(sql, val...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
