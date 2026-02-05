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
	t := Todo{
		ID:        s.data.NextID,
		Text:      text,
		Date:      date,
		Done:      false,
		CreatedAt: time.Now().Format(dateFormat),
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

// Todos returns all todos in the store.
func (s *Store) Todos() []Todo {
	return s.data.Todos
}

// TodosForMonth returns todos whose date falls in the given year and month,
// sorted by date ascending then by ID for same-date stability.
func (s *Store) TodosForMonth(year int, month time.Month) []Todo {
	var result []Todo
	for _, t := range s.data.Todos {
		if t.InMonth(year, month) {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
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

// FloatingTodos returns todos with no date assigned, sorted by ID ascending.
func (s *Store) FloatingTodos() []Todo {
	var result []Todo
	for _, t := range s.data.Todos {
		if !t.HasDate() {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}
