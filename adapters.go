package migrations

type SQLiteAdapter struct{}

func NewSQLiteAdapter() *SQLiteAdapter {
	return &SQLiteAdapter{}
}

func (S SQLiteAdapter) CreateTableSQL() string {
	return `CREATE TABLE IF NOT EXISTS migrations (number int)`
}

func (S SQLiteAdapter) SelectLatestMigrationSQL() string {
	return `SELECT number FROM migrations ORDER BY number DESC LIMIT 1`
}

func (S SQLiteAdapter) InsertMigrationSQL() string {
	return `INSERT INTO migrations (number) VALUES (?)`
}
