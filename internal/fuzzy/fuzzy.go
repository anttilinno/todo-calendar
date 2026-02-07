package fuzzy

import "strings"

// Match performs a fuzzy match of query against target. Characters in query
// must appear in target in the same order (case-insensitive). Returns whether
// the match succeeded and a relevance score. Higher scores indicate better
// matches (consecutive characters, word-boundary alignment).
func Match(query, target string) (bool, int) {
	if query == "" {
		return true, 0
	}

	q := strings.ToLower(query)
	t := strings.ToLower(target)

	qi := 0
	score := 0
	prevMatch := -1

	for ti := 0; ti < len(t) && qi < len(q); ti++ {
		if t[ti] == q[qi] {
			score += 1

			// Consecutive match bonus
			if prevMatch == ti-1 {
				score += 3
			}

			// Word-boundary bonus: first char or preceded by space/punctuation
			if ti == 0 || t[ti-1] == ' ' || t[ti-1] == '-' || t[ti-1] == '_' || t[ti-1] == '/' {
				score += 5
			}

			// Exact case bonus
			if ti < len(target) && qi < len(query) && target[ti] == query[qi] {
				score += 1
			}

			prevMatch = ti
			qi++
		}
	}

	if qi < len(q) {
		return false, 0
	}

	return true, score
}
