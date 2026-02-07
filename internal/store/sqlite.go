package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Compile-time check: *SQLiteStore implements TodoStore.
var _ TodoStore = (*SQLiteStore)(nil)

// SQLiteStore implements TodoStore backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at dbPath and returns
// a ready-to-use store. The parent directory is created if it does not exist.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite does not support concurrent writers; restrict to one connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return s, nil
}

// migrate applies schema migrations using PRAGMA user_version for tracking.
func (s *SQLiteStore) migrate() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	if version < 1 {
		if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS todos (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			text       TEXT    NOT NULL,
			body       TEXT    NOT NULL DEFAULT '',
			date       TEXT,
			done       INTEGER NOT NULL DEFAULT 0,
			created_at TEXT    NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0
		)`); err != nil {
			return fmt.Errorf("create todos table: %w", err)
		}
		if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_date ON todos(date)`); err != nil {
			return fmt.Errorf("create date index: %w", err)
		}
		if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_todos_done ON todos(done)`); err != nil {
			return fmt.Errorf("create done index: %w", err)
		}
		if _, err := s.db.Exec(`PRAGMA user_version = 1`); err != nil {
			return fmt.Errorf("set user_version: %w", err)
		}
	}

	if version < 2 {
		if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS templates (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL UNIQUE,
			content    TEXT    NOT NULL,
			created_at TEXT    NOT NULL
		)`); err != nil {
			return fmt.Errorf("create templates table: %w", err)
		}
		if _, err := s.db.Exec(`PRAGMA user_version = 2`); err != nil {
			return fmt.Errorf("set user_version: %w", err)
		}
	}

	if version < 3 {
		for _, t := range defaultTemplates() {
			s.db.Exec(
				"INSERT OR IGNORE INTO templates (name, content, created_at) VALUES (?, ?, ?)",
				t.Name, t.Content, time.Now().Format(dateFormat),
			)
		}
		if _, err := s.db.Exec("PRAGMA user_version = 3"); err != nil {
			return fmt.Errorf("set user_version: %w", err)
		}
	}

	return nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// todoColumns is the column list used in SELECT statements.
const todoColumns = "id, text, body, date, done, created_at, sort_order"

// scanTodo scans a single todo row from the given scanner.
func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
	var t Todo
	var date sql.NullString
	var done int
	err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder)
	if err != nil {
		return Todo{}, err
	}
	t.Done = done != 0
	if date.Valid {
		t.Date = date.String
	}
	return t, nil
}

// scanTodos scans multiple rows into a slice of Todo.
func scanTodos(rows *sql.Rows) ([]Todo, error) {
	var todos []Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

// Add creates a new todo and returns it. Date="" means floating (NULL in DB).
func (s *SQLiteStore) Add(text string, date string) Todo {
	createdAt := time.Now().Format(dateFormat)

	// Compute next sort_order as MAX(sort_order) + 10.
	var maxOrder int
	_ = s.db.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM todos").Scan(&maxOrder)
	sortOrder := maxOrder + 10

	var dateVal any
	if date != "" {
		dateVal = date
	}

	result, err := s.db.Exec(
		"INSERT INTO todos (text, body, date, done, created_at, sort_order) VALUES (?, '', ?, 0, ?, ?)",
		text, dateVal, createdAt, sortOrder,
	)
	if err != nil {
		return Todo{}
	}

	id, _ := result.LastInsertId()
	return Todo{
		ID:        int(id),
		Text:      text,
		Date:      date,
		Done:      false,
		CreatedAt: createdAt,
		SortOrder: sortOrder,
	}
}

// Toggle flips the Done status of the todo with the given ID.
func (s *SQLiteStore) Toggle(id int) {
	s.db.Exec("UPDATE todos SET done = NOT done WHERE id = ?", id)
}

// Delete removes the todo with the given ID.
func (s *SQLiteStore) Delete(id int) {
	s.db.Exec("DELETE FROM todos WHERE id = ?", id)
}

// Find returns a pointer to the todo with the given ID, or nil if not found.
func (s *SQLiteStore) Find(id int) *Todo {
	row := s.db.QueryRow("SELECT "+todoColumns+" FROM todos WHERE id = ?", id)
	t, err := scanTodo(row)
	if err != nil {
		return nil
	}
	return &t
}

// Update modifies the text and date of the todo with the given ID.
// Date="" means floating (NULL in DB).
func (s *SQLiteStore) Update(id int, text string, date string) {
	var dateVal any
	if date != "" {
		dateVal = date
	}
	s.db.Exec("UPDATE todos SET text = ?, date = ? WHERE id = ?", text, dateVal, id)
}

// UpdateBody sets the markdown body of the todo with the given ID.
func (s *SQLiteStore) UpdateBody(id int, body string) {
	s.db.Exec("UPDATE todos SET body = ? WHERE id = ?", body, id)
}

// Todos returns all todos ordered by sort_order, then id.
func (s *SQLiteStore) Todos() []Todo {
	rows, err := s.db.Query("SELECT " + todoColumns + " FROM todos ORDER BY sort_order, id")
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// TodosForMonth returns todos whose date falls in the given year and month,
// sorted by sort_order, date, then id.
func (s *SQLiteStore) TodosForMonth(year int, month time.Month) []Todo {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	// Last day: go to first of next month, subtract one day.
	end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE date >= ? AND date <= ? ORDER BY sort_order, date, id",
		start, end,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// FloatingTodos returns todos with no date, sorted by sort_order then id.
func (s *SQLiteStore) FloatingTodos() []Todo {
	rows, err := s.db.Query("SELECT " + todoColumns + " FROM todos WHERE date IS NULL ORDER BY sort_order, id")
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// IncompleteTodosPerDay returns a map from day-of-month to count of
// incomplete todos for the specified year and month.
func (s *SQLiteStore) IncompleteTodosPerDay(year int, month time.Month) map[int]int {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

	rows, err := s.db.Query(
		"SELECT CAST(substr(date, 9, 2) AS INTEGER) AS day, COUNT(*) FROM todos WHERE done = 0 AND date >= ? AND date <= ? GROUP BY day",
		start, end,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var day, count int
		if err := rows.Scan(&day, &count); err == nil {
			counts[day] = count
		}
	}
	return counts
}

// TotalTodosPerDay returns a map from day-of-month to count of all todos
// (both done and not done) for the specified year and month.
func (s *SQLiteStore) TotalTodosPerDay(year int, month time.Month) map[int]int {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

	rows, err := s.db.Query(
		"SELECT CAST(substr(date, 9, 2) AS INTEGER) AS day, COUNT(*) FROM todos WHERE date >= ? AND date <= ? GROUP BY day",
		start, end,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var day, count int
		if err := rows.Scan(&day, &count); err == nil {
			counts[day] = count
		}
	}
	return counts
}

// TodoCountsByMonth returns pending and completed counts per month across
// all dated todos, sorted chronologically.
func (s *SQLiteStore) TodoCountsByMonth() []MonthCount {
	rows, err := s.db.Query(`
		SELECT substr(date, 1, 7) AS ym,
		       SUM(CASE WHEN done = 0 THEN 1 ELSE 0 END) AS pending,
		       SUM(CASE WHEN done = 1 THEN 1 ELSE 0 END) AS completed
		FROM todos
		WHERE date IS NOT NULL
		GROUP BY ym
		ORDER BY ym
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []MonthCount
	for rows.Next() {
		var ym string
		var mc MonthCount
		if err := rows.Scan(&ym, &mc.Pending, &mc.Completed); err != nil {
			continue
		}
		t, err := time.Parse("2006-01", ym)
		if err != nil {
			continue
		}
		mc.Year = t.Year()
		mc.Month = t.Month()
		result = append(result, mc)
	}
	return result
}

// FloatingTodoCounts returns pending and completed counts for todos with no date.
func (s *SQLiteStore) FloatingTodoCounts() FloatingCount {
	var fc FloatingCount
	s.db.QueryRow(`
		SELECT COALESCE(SUM(CASE WHEN done = 0 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN done = 1 THEN 1 ELSE 0 END), 0)
		FROM todos WHERE date IS NULL
	`).Scan(&fc.Pending, &fc.Completed)
	return fc
}

// SwapOrder swaps the sort_order values of two todos in a transaction.
func (s *SQLiteStore) SwapOrder(id1, id2 int) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	var order1, order2 int
	if err := tx.QueryRow("SELECT sort_order FROM todos WHERE id = ?", id1).Scan(&order1); err != nil {
		return
	}
	if err := tx.QueryRow("SELECT sort_order FROM todos WHERE id = ?", id2).Scan(&order2); err != nil {
		return
	}

	tx.Exec("UPDATE todos SET sort_order = ? WHERE id = ?", order2, id1)
	tx.Exec("UPDATE todos SET sort_order = ? WHERE id = ?", order1, id2)
	tx.Commit()
}

// SearchTodos returns todos whose text contains the query (case-insensitive),
// sorted: dated first by date ascending, then floating by id.
func (s *SQLiteStore) SearchTodos(query string) []Todo {
	if query == "" {
		return nil
	}
	pattern := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"

	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE text LIKE ? ESCAPE '\\' ORDER BY CASE WHEN date IS NULL THEN 1 ELSE 0 END, date, id",
		pattern,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// EnsureSortOrder assigns sort_order = id * 10 to any todos with sort_order = 0.
func (s *SQLiteStore) EnsureSortOrder() {
	s.db.Exec("UPDATE todos SET sort_order = id * 10 WHERE sort_order = 0")
}

// AddTemplate creates a new template with the given name and content.
// Returns an error if the name is not unique.
func (s *SQLiteStore) AddTemplate(name, content string) (Template, error) {
	createdAt := time.Now().Format(dateFormat)
	result, err := s.db.Exec(
		"INSERT INTO templates (name, content, created_at) VALUES (?, ?, ?)",
		name, content, createdAt,
	)
	if err != nil {
		return Template{}, fmt.Errorf("add template: %w", err)
	}
	id, _ := result.LastInsertId()
	return Template{
		ID:        int(id),
		Name:      name,
		Content:   content,
		CreatedAt: createdAt,
	}, nil
}

// ListTemplates returns all templates ordered by name.
func (s *SQLiteStore) ListTemplates() []Template {
	rows, err := s.db.Query("SELECT id, name, content, created_at FROM templates ORDER BY name")
	if err != nil {
		return []Template{}
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Content, &t.CreatedAt); err != nil {
			continue
		}
		templates = append(templates, t)
	}
	if templates == nil {
		return []Template{}
	}
	return templates
}

// FindTemplate returns the template with the given ID, or nil if not found.
func (s *SQLiteStore) FindTemplate(id int) *Template {
	var t Template
	err := s.db.QueryRow(
		"SELECT id, name, content, created_at FROM templates WHERE id = ?", id,
	).Scan(&t.ID, &t.Name, &t.Content, &t.CreatedAt)
	if err != nil {
		return nil
	}
	return &t
}

// DeleteTemplate removes the template with the given ID.
func (s *SQLiteStore) DeleteTemplate(id int) {
	s.db.Exec("DELETE FROM templates WHERE id = ?", id)
}

// Save is a no-op for SQLiteStore since all mutations are immediately persisted.
func (s *SQLiteStore) Save() error {
	return nil
}
