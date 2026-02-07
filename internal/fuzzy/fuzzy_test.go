package fuzzy

import "testing"

func TestMatch_EmptyQuery(t *testing.T) {
	matched, score := Match("", "anything")
	if !matched || score != 0 {
		t.Errorf("empty query should match everything: matched=%v score=%d", matched, score)
	}
}

func TestMatch_ExactMatch(t *testing.T) {
	matched, score := Match("Bug Report", "Bug Report")
	if !matched {
		t.Error("exact match should succeed")
	}
	if score <= 0 {
		t.Errorf("exact match should have positive score: %d", score)
	}
}

func TestMatch_FuzzyChars(t *testing.T) {
	matched, _ := Match("bg", "Bug Report")
	if !matched {
		t.Error("'bg' should fuzzy-match 'Bug Report'")
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	matched, _ := Match("BUG", "bug report")
	if !matched {
		t.Error("case-insensitive match should succeed")
	}
}

func TestMatch_NoMatch(t *testing.T) {
	matched, _ := Match("xyz", "Bug Report")
	if matched {
		t.Error("'xyz' should not match 'Bug Report'")
	}
}

func TestMatch_OutOfOrder(t *testing.T) {
	matched, _ := Match("rb", "Bug Report")
	if matched {
		t.Error("'rb' should not match 'Bug Report' (out of order)")
	}
}

func TestMatch_ConsecutiveBonus(t *testing.T) {
	_, scoreConsec := Match("bug", "Bug Report")
	_, scoreSpread := Match("brt", "Bug Report")
	if scoreConsec <= scoreSpread {
		t.Errorf("consecutive match should score higher: consec=%d spread=%d", scoreConsec, scoreSpread)
	}
}

func TestMatch_WordBoundaryBonus(t *testing.T) {
	_, scoreBoundary := Match("br", "Bug Report")
	_, scoreMiddle := Match("ug", "Bug Report")
	if scoreBoundary <= scoreMiddle {
		t.Errorf("word-boundary match should score higher: boundary=%d middle=%d", scoreBoundary, scoreMiddle)
	}
}
