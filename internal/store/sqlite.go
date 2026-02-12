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

	if version < 4 {
		if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schedules (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			template_id          INTEGER NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
			cadence_type         TEXT    NOT NULL,
			cadence_value        TEXT    NOT NULL DEFAULT '',
			placeholder_defaults TEXT    NOT NULL DEFAULT '{}',
			created_at           TEXT    NOT NULL
		)`); err != nil {
			return fmt.Errorf("create schedules table: %w", err)
		}
		if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_schedules_template ON schedules(template_id)`); err != nil {
			return fmt.Errorf("create schedules template index: %w", err)
		}
		if _, err := s.db.Exec(`PRAGMA user_version = 4`); err != nil {
			return fmt.Errorf("set user_version: %w", err)
		}
	}

	if version < 5 {
		if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN schedule_id INTEGER REFERENCES schedules(id) ON DELETE SET NULL`); err != nil {
			return fmt.Errorf("add schedule_id column: %w", err)
		}
		if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN schedule_date TEXT`); err != nil {
			return fmt.Errorf("add schedule_date column: %w", err)
		}
		if _, err := s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_todos_schedule_dedup ON todos(schedule_id, schedule_date)`); err != nil {
			return fmt.Errorf("create schedule dedup index: %w", err)
		}
		if _, err := s.db.Exec(`PRAGMA user_version = 5`); err != nil {
			return fmt.Errorf("set user_version: %w", err)
		}
	}

	if version < 6 {
		if _, err := s.db.Exec(`ALTER TABLE todos ADD COLUMN date_precision TEXT NOT NULL DEFAULT 'day'`); err != nil {
			return fmt.Errorf("add date_precision column: %w", err)
		}
		// Floating todos (date IS NULL) should have empty date_precision, not 'day'.
		if _, err := s.db.Exec(`UPDATE todos SET date_precision = '' WHERE date IS NULL`); err != nil {
			return fmt.Errorf("fix floating date_precision: %w", err)
		}
		if _, err := s.db.Exec(`PRAGMA user_version = 6`); err != nil {
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
const todoColumns = "id, text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision"

// scanTodo scans a single todo row from the given scanner.
func scanTodo(scanner interface{ Scan(...any) error }) (Todo, error) {
	var t Todo
	var date sql.NullString
	var done int
	var scheduleID sql.NullInt64
	var scheduleDate sql.NullString
	err := scanner.Scan(&t.ID, &t.Text, &t.Body, &date, &done, &t.CreatedAt, &t.SortOrder, &scheduleID, &scheduleDate, &t.DatePrecision)
	if err != nil {
		return Todo{}, err
	}
	t.Done = done != 0
	if date.Valid {
		t.Date = date.String
	}
	if scheduleID.Valid {
		t.ScheduleID = int(scheduleID.Int64)
	}
	if scheduleDate.Valid {
		t.ScheduleDate = scheduleDate.String
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
// datePrecision is "day", "month", "year", or "" (floating).
func (s *SQLiteStore) Add(text string, date string, datePrecision string) Todo {
	createdAt := time.Now().Format(dateFormat)

	// Compute next sort_order as MAX(sort_order) + 10.
	var maxOrder int
	_ = s.db.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM todos").Scan(&maxOrder)
	sortOrder := maxOrder + 10

	var dateVal any
	if date != "" {
		dateVal = date
	}

	// Floating todos get empty precision.
	if date == "" {
		datePrecision = ""
	}

	result, err := s.db.Exec(
		"INSERT INTO todos (text, body, date, done, created_at, sort_order, date_precision) VALUES (?, '', ?, 0, ?, ?, ?)",
		text, dateVal, createdAt, sortOrder, datePrecision,
	)
	if err != nil {
		return Todo{}
	}

	id, _ := result.LastInsertId()
	return Todo{
		ID:            int(id),
		Text:          text,
		Date:          date,
		Done:          false,
		CreatedAt:     createdAt,
		SortOrder:     sortOrder,
		DatePrecision: datePrecision,
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

// Update modifies the text, date, and date precision of the todo with the given ID.
// Date="" means floating (NULL in DB). datePrecision is "day", "month", "year", or "" (floating).
func (s *SQLiteStore) Update(id int, text string, date string, datePrecision string) {
	var dateVal any
	if date != "" {
		dateVal = date
	}
	// Floating todos get empty precision.
	if date == "" {
		datePrecision = ""
	}
	s.db.Exec("UPDATE todos SET text = ?, date = ?, date_precision = ? WHERE id = ?", text, dateVal, datePrecision, id)
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

// TodosForMonth returns day-precision todos whose date falls in the given year and month,
// sorted by sort_order, date, then id. Excludes fuzzy-date (month/year precision) todos.
func (s *SQLiteStore) TodosForMonth(year int, month time.Month) []Todo {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	// Last day: go to first of next month, subtract one day.
	end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE date >= ? AND date <= ? AND date_precision = 'day' ORDER BY sort_order, date, id",
		start, end,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// TodosForDateRange returns day-precision todos whose date falls within [startDate, endDate] inclusive,
// sorted by sort_order, date, then id. Parameters are ISO date strings ("YYYY-MM-DD").
// Excludes fuzzy-date (month/year precision) todos.
func (s *SQLiteStore) TodosForDateRange(startDate, endDate string) []Todo {
	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE date >= ? AND date <= ? AND date_precision = 'day' ORDER BY sort_order, date, id",
		startDate, endDate,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// MonthTodos returns month-precision todos for the given year and month,
// sorted by sort_order, then id.
func (s *SQLiteStore) MonthTodos(year int, month time.Month) []Todo {
	ym := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE date_precision = 'month' AND substr(date, 1, 7) = ? ORDER BY sort_order, id",
		ym,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	todos, _ := scanTodos(rows)
	return todos
}

// YearTodos returns year-precision todos for the given year,
// sorted by sort_order, then id.
func (s *SQLiteStore) YearTodos(year int) []Todo {
	y := fmt.Sprintf("%04d", year)
	rows, err := s.db.Query(
		"SELECT "+todoColumns+" FROM todos WHERE date_precision = 'year' AND substr(date, 1, 4) = ? ORDER BY sort_order, id",
		y,
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
		"SELECT CAST(substr(date, 9, 2) AS INTEGER) AS day, COUNT(*) FROM todos WHERE done = 0 AND date >= ? AND date <= ? AND date_precision = 'day' GROUP BY day",
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
// Excludes fuzzy-date (month/year precision) todos.
func (s *SQLiteStore) TotalTodosPerDay(year int, month time.Month) map[int]int {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Format(dateFormat)

	rows, err := s.db.Query(
		"SELECT CAST(substr(date, 9, 2) AS INTEGER) AS day, COUNT(*) FROM todos WHERE date >= ? AND date <= ? AND date_precision = 'day' GROUP BY day",
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

// UpdateTemplate updates both the name and content of a template by ID.
// Returns an error if the name violates the UNIQUE constraint.
func (s *SQLiteStore) UpdateTemplate(id int, name, content string) error {
	_, err := s.db.Exec("UPDATE templates SET name = ?, content = ? WHERE id = ?", name, content, id)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	return nil
}

// Save is a no-op for SQLiteStore since all mutations are immediately persisted.
func (s *SQLiteStore) Save() error {
	return nil
}

// scanSchedule scans a single schedule row from the given scanner.
func scanSchedule(scanner interface{ Scan(...any) error }) (Schedule, error) {
	var sc Schedule
	err := scanner.Scan(&sc.ID, &sc.TemplateID, &sc.CadenceType, &sc.CadenceValue, &sc.PlaceholderDefaults, &sc.CreatedAt)
	if err != nil {
		return Schedule{}, err
	}
	return sc, nil
}

// scanSchedules scans multiple rows into a slice of Schedule.
func scanSchedules(rows *sql.Rows) ([]Schedule, error) {
	var schedules []Schedule
	for rows.Next() {
		sc, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, sc)
	}
	return schedules, rows.Err()
}

// AddSchedule creates a new schedule linked to a template.
func (s *SQLiteStore) AddSchedule(templateID int, cadenceType, cadenceValue, placeholderDefaults string) (Schedule, error) {
	createdAt := time.Now().Format(dateFormat)
	result, err := s.db.Exec(
		"INSERT INTO schedules (template_id, cadence_type, cadence_value, placeholder_defaults, created_at) VALUES (?, ?, ?, ?, ?)",
		templateID, cadenceType, cadenceValue, placeholderDefaults, createdAt,
	)
	if err != nil {
		return Schedule{}, fmt.Errorf("add schedule: %w", err)
	}
	id, _ := result.LastInsertId()
	return Schedule{
		ID:                 int(id),
		TemplateID:         templateID,
		CadenceType:        cadenceType,
		CadenceValue:       cadenceValue,
		PlaceholderDefaults: placeholderDefaults,
		CreatedAt:          createdAt,
	}, nil
}

// ListSchedules returns all schedules ordered by ID.
func (s *SQLiteStore) ListSchedules() []Schedule {
	rows, err := s.db.Query("SELECT id, template_id, cadence_type, cadence_value, placeholder_defaults, created_at FROM schedules ORDER BY id")
	if err != nil {
		return nil
	}
	defer rows.Close()
	schedules, _ := scanSchedules(rows)
	return schedules
}

// ListSchedulesForTemplate returns schedules for a given template ordered by ID.
func (s *SQLiteStore) ListSchedulesForTemplate(templateID int) []Schedule {
	rows, err := s.db.Query(
		"SELECT id, template_id, cadence_type, cadence_value, placeholder_defaults, created_at FROM schedules WHERE template_id = ? ORDER BY id",
		templateID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	schedules, _ := scanSchedules(rows)
	return schedules
}

// DeleteSchedule removes a schedule by ID.
func (s *SQLiteStore) DeleteSchedule(id int) {
	s.db.Exec("DELETE FROM schedules WHERE id = ?", id)
}

// UpdateSchedule modifies the cadence and placeholder defaults of a schedule.
func (s *SQLiteStore) UpdateSchedule(id int, cadenceType, cadenceValue, placeholderDefaults string) error {
	_, err := s.db.Exec(
		"UPDATE schedules SET cadence_type = ?, cadence_value = ?, placeholder_defaults = ? WHERE id = ?",
		cadenceType, cadenceValue, placeholderDefaults, id,
	)
	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}
	return nil
}

// TodoExistsForSchedule checks if a todo already exists for a schedule and date.
func (s *SQLiteStore) TodoExistsForSchedule(scheduleID int, date string) bool {
	var exists int
	err := s.db.QueryRow(
		"SELECT 1 FROM todos WHERE schedule_id = ? AND schedule_date = ? LIMIT 1",
		scheduleID, date,
	).Scan(&exists)
	return err == nil
}

// AddScheduledTodo creates a todo linked to a schedule with schedule_date set.
// Scheduled todos are always day-precision.
func (s *SQLiteStore) AddScheduledTodo(text, date, body string, scheduleID int) Todo {
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
		"INSERT INTO todos (text, body, date, done, created_at, sort_order, schedule_id, schedule_date, date_precision) VALUES (?, ?, ?, 0, ?, ?, ?, ?, 'day')",
		text, body, dateVal, createdAt, sortOrder, scheduleID, date,
	)
	if err != nil {
		return Todo{}
	}

	id, _ := result.LastInsertId()
	return Todo{
		ID:            int(id),
		Text:          text,
		Body:          body,
		Date:          date,
		Done:          false,
		CreatedAt:     createdAt,
		SortOrder:     sortOrder,
		ScheduleID:    scheduleID,
		ScheduleDate:  date,
		DatePrecision: "day",
	}
}
