package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Store manages todo persistence with atomic file writes.
type Store struct {
	path string
	data Data
}

// TodosPath returns the XDG-compliant path for the todos data file.
// Uses os.UserConfigDir to resolve ~/.config/todo-calendar/todos.json.
func TodosPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "todo-calendar", "todos.json"), nil
}

// NewStore creates a Store backed by the given file path.
// If the file is missing or empty, the store starts with an empty todo list.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	s.EnsureSortOrder()
	return s, nil
}

// load reads the data file into memory.
// Missing or empty files are treated as a fresh store.
func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) || (err == nil && len(b) == 0) {
		s.data = Data{NextID: 1, Todos: []Todo{}}
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.data)
}

// Save writes the current state to disk using an atomic write pattern
// (write to temp file, sync, rename) to prevent data corruption.
func (s *Store) Save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".todos-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, s.path)
}

// Add creates a new todo with the given text and optional date (YYYY-MM-DD or "").
// It persists the change and returns the newly created todo.
func (s *Store) Add(text string, date string) Todo {
	maxOrder := 0
	for _, t := range s.data.Todos {
		if t.SortOrder > maxOrder {
			maxOrder = t.SortOrder
		}
	}
	t := Todo{
		ID:        s.data.NextID,
		Text:      text,
		Date:      date,
		Done:      false,
		CreatedAt: time.Now().Format(dateFormat),
		SortOrder: maxOrder + 10,
	}
	s.data.NextID++
	s.data.Todos = append(s.data.Todos, t)
	s.Save()
	return t
}

// Toggle flips the Done status of the todo with the given ID and persists.
func (s *Store) Toggle(id int) {
	for i := range s.data.Todos {
		if s.data.Todos[i].ID == id {
			s.data.Todos[i].Done = !s.data.Todos[i].Done
			s.Save()
			return
		}
	}
}

// Delete removes the todo with the given ID and persists.
func (s *Store) Delete(id int) {
	for i, t := range s.data.Todos {
		if t.ID == id {
			s.data.Todos = append(s.data.Todos[:i], s.data.Todos[i+1:]...)
			s.Save()
			return
		}
	}
}

// Find returns a pointer to the todo with the given ID, or nil if not found.
// This is a read-only lookup and does not call Save.
func (s *Store) Find(id int) *Todo {
	for i := range s.data.Todos {
		if s.data.Todos[i].ID == id {
			return &s.data.Todos[i]
		}
	}
	return nil
}

// Update modifies the text and date of the todo with the given ID and persists.
// Date="" means floating (no date). If ID is not found, does nothing.
func (s *Store) Update(id int, text string, date string) {
	for i := range s.data.Todos {
		if s.data.Todos[i].ID == id {
			s.data.Todos[i].Text = text
			s.data.Todos[i].Date = date
			s.Save()
			return
		}
	}
}

// Todos returns all todos in the store.
func (s *Store) Todos() []Todo {
	return s.data.Todos
}

// TodosForMonth returns todos whose date falls in the given year and month,
// sorted by SortOrder ascending, then date, then ID for stability.
func (s *Store) TodosForMonth(year int, month time.Month) []Todo {
	var result []Todo
	for _, t := range s.data.Todos {
		if t.InMonth(year, month) {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder != result[j].SortOrder {
			return result[i].SortOrder < result[j].SortOrder
		}
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].ID < result[j].ID
	})
	return result
}

// IncompleteTodosPerDay returns a map from day-of-month to count of
// incomplete (not done) todos for the specified year and month.
// Days with zero incomplete todos are omitted from the map.
func (s *Store) IncompleteTodosPerDay(year int, month time.Month) map[int]int {
	counts := make(map[int]int)
	for _, t := range s.data.Todos {
		if t.Done || !t.InMonth(year, month) {
			continue
		}
		d, err := time.Parse(dateFormat, t.Date)
		if err != nil {
			continue
		}
		counts[d.Day()]++
	}
	return counts
}

// FloatingTodos returns todos with no date assigned, sorted by SortOrder then ID.
func (s *Store) FloatingTodos() []Todo {
	var result []Todo
	for _, t := range s.data.Todos {
		if !t.HasDate() {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder != result[j].SortOrder {
			return result[i].SortOrder < result[j].SortOrder
		}
		return result[i].ID < result[j].ID
	})
	return result
}

// EnsureSortOrder assigns unique SortOrder values to any todos that
// have the zero value (legacy data). Called once at load time.
func (s *Store) EnsureSortOrder() {
	needsSave := false
	for i := range s.data.Todos {
		if s.data.Todos[i].SortOrder == 0 {
			s.data.Todos[i].SortOrder = (i + 1) * 10
			needsSave = true
		}
	}
	if needsSave {
		s.Save()
	}
}

// MonthCount holds the total number of todos for a given year and month.
type MonthCount struct {
	Year  int
	Month time.Month
	Count int
}

// TodoCountsByMonth returns the number of todos per month across all dated
// todos, sorted chronologically (year ascending, then month ascending).
// Undated (floating) todos are excluded.
func (s *Store) TodoCountsByMonth() []MonthCount {
	type ym struct {
		y int
		m time.Month
	}
	counts := make(map[ym]int)
	for _, t := range s.data.Todos {
		if t.Date == "" {
			continue
		}
		d, err := time.Parse(dateFormat, t.Date)
		if err != nil {
			continue
		}
		counts[ym{d.Year(), d.Month()}]++
	}
	result := make([]MonthCount, 0, len(counts))
	for k, c := range counts {
		result = append(result, MonthCount{Year: k.y, Month: k.m, Count: c})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Year != result[j].Year {
			return result[i].Year < result[j].Year
		}
		return result[i].Month < result[j].Month
	})
	return result
}

// FloatingTodoCount returns the number of todos with no date assigned.
func (s *Store) FloatingTodoCount() int {
	count := 0
	for _, t := range s.data.Todos {
		if !t.HasDate() {
			count++
		}
	}
	return count
}

// SwapOrder swaps the SortOrder values of two todos identified by ID
// and persists the change. If either ID is not found, does nothing.
func (s *Store) SwapOrder(id1, id2 int) {
	var t1, t2 *Todo
	for i := range s.data.Todos {
		switch s.data.Todos[i].ID {
		case id1:
			t1 = &s.data.Todos[i]
		case id2:
			t2 = &s.data.Todos[i]
		}
	}
	if t1 != nil && t2 != nil {
		t1.SortOrder, t2.SortOrder = t2.SortOrder, t1.SortOrder
		s.Save()
	}
}
