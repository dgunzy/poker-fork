package poker

import (
	"sort"
	"strings"
	"testing"
)

// This is copied from poker_test.go to avoid export issues
func parseHandForTest(t *testing.T, s string) []Card {
	cs := strings.Split(s, " ")
	r := make([]Card, len(cs))
	for i, p := range cs {
		var ok bool
		r[i], ok = NameToCard[p]
		if !ok {
			t.Fatalf("can't parse card %s", p)
		}
	}
	return r
}

// sortCards sorts cards in descending order by rank for 2-7 evaluation
func sortCards(cards []Card) {
	sort.Slice(cards, func(i, j int) bool {
		// Get ranks, treating ace as highest
		ri := (int(cards[i]>>2) & 15) + 1
		rj := (int(cards[j]>>2) & 15) + 1
		if ri == 1 {
			ri = 14 // Ace high
		}
		if rj == 1 {
			rj = 14 // Ace high
		}
		if ri != rj {
			return ri > rj // Higher ranks first
		}
		// If ranks equal, sort by suit for consistency
		return cards[i]&3 > cards[j]&3
	})
}

func TestEval27(t *testing.T) {
	tests := []struct {
		name        string
		hand1       string
		hand2       string
		want1Better bool
	}{
		{
			name:        "Perfect vs Second best",
			hand1:       "D2 H3 C4 S5 H7",
			hand2:       "D2 H3 C4 S5 H8",
			want1Better: true,
		},
		{
			name:        "Ace should be high",
			hand1:       "D2 H3 C4 S5 HA", // Ace makes this worse
			hand2:       "D2 H3 C4 S5 H7",
			want1Better: false,
		},
		{
			name:        "Ace high vs King high",
			hand1:       "D2 H3 C4 S5 HA",
			hand2:       "D2 H3 C4 S5 HK",
			want1Better: false,
		},
		{
			name:        "Wheel should not be recognized, first hand should be much lower.",
			hand1:       "DA H2 C3 S4 H5", // Not a wheel
			hand2:       "D6 H3 C4 S5 H7", // straight
			want1Better: true,
		},
		{
			name:        "Wheel should not be recognized again - look for a lower number on first hand.",
			hand1:       "D2 H3 C4 S5 HA", // Wheel
			hand2:       "DK H3 C4 S5 CA", // Not a straight - but Ace king is worse than ace high
			want1Better: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand1 := parseHandForTest(t, tt.hand1)
			hand2 := parseHandForTest(t, tt.hand2)

			// Sort both hands before evaluation
			sortCards(hand1)
			sortCards(hand2)

			var h1, h2 [5]Card
			copy(h1[:], hand1)
			copy(h2[:], hand2)

			score1 := Eval27(&h1)
			score2 := Eval27(&h2)
			result := int(score1) - int(score2)

			// Print scores and sorted hands
			t.Logf("Hand1 (%s) sorted to (%s) score: %d",
				tt.hand1, Hand(hand1).String(), score1)
			t.Logf("Hand2 (%s) sorted to (%s) score: %d",
				tt.hand2, Hand(hand2).String(), score2)

			if tt.want1Better && result >= 0 {
				t.Errorf("Expected %s to be better than %s", tt.hand1, tt.hand2)
			}
			if !tt.want1Better && result <= 0 {
				t.Errorf("Expected %s to be worse than %s", tt.hand1, tt.hand2)
			}
		})
	}
}
