package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Migration struct {
	Order int
	Up    string
	Down  string
}
type Migrations []*Migration

func (m Migrations) Len() int { return len(m) }
func (m Migrations) Less(i, j int) bool {
	return m[i].Order < m[j].Order
}
func (m Migrations) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func fromDir(directory string) []*Migration {
	out := make([]*Migration, 0)
	if err := filepath.Walk(directory, BuildMigrations(&out)); err != nil {
		panic(err)
	}
	sort.Sort(Migrations(out))
	return out
}

func BuildMigrations(m *[]*Migration) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ignore nested dirs, why are these here?
		if info.IsDir() {
			return nil
		}
		up, down := getUpAndDownFromFile(path)
		migration := &Migration{
			Order: getOrderFromFilename(info.Name()),
			Up:    up,
			Down:  down,
		}
		*m = append(*m, migration)
		return nil
	}
}

func getOrderFromFilename(name string) int {
	i, err := strconv.Atoi(strings.TrimSuffix(name, ".sql"))
	if err != nil {
		panic(err)
	}
	return i
}

const SplitMarker = "-- SPLIT --"

func getUpAndDownFromFile(name string) (string, string) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(string(b), SplitMarker)
	return parts[0], parts[1]
}

type Database interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type QueryAdapter interface {
	CreateTableSQL() string
	SelectLatestMigrationSQL() string
	InsertMigrationSQL() string
}

// Initialize runs the migrations that have not yet been run according to the migrations table
func Initialize(ctx context.Context, migrationDir string, db Database, queries QueryAdapter) error {
	// make sure migrations can persist
	_, err := db.ExecContext(ctx, queries.CreateTableSQL())
	if err != nil {
		return errors.WithStack(err)
	}
	// Get the migration with the highest ID number
	row := db.QueryRowContext(ctx, queries.SelectLatestMigrationSQL())
	if err := row.Err(); err != nil {
		return errors.WithStack(err)
	}
	latest := 0

	if err := row.Scan(&latest); err != nil {
		if err != sql.ErrNoRows {
			return errors.WithStack(err)
		}
	}

	// Load up the migrations and run starting from the last known run migration
	migs := fromDir(migrationDir)
	for _, migration := range migs[:latest] {
		fmt.Printf("Already ran migration: %d\n", migration.Order)
	}

	for _, migration := range migs[latest:] {
		fmt.Printf("Running migration %d\n", migration.Order)
		if _, err := db.ExecContext(ctx, migration.Up); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, queries.InsertMigrationSQL(), migration.Order); err != nil {
			return err
		}
	}
	return nil
}
