package poker

import (
	"testing"
)

func TestDeuce7TableGeneration(t *testing.T) {
	// Force table initialization
	initDeuce7Table()

	if deuce7RootTable == nil {
		t.Fatal("deuce7RootTable was not initialized")
	}

	// Test cases from deuce_seven_test.go
	testCases := []struct {
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
			hand1:       "D2 H3 C4 S5 HA",
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
			name:        "Pair of Aces vs Ace high",
			hand1:       "HA CA D3 C4 H5",
			hand2:       "HA D3 C4 H5 S7",
			want1Better: false,
		},
		{
			name:        "Flush vs high card",
			hand1:       "H2 H3 H4 H5 H7",
			hand2:       "D2 C3 H4 S5 HA",
			want1Better: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hand1 := parseHandForTest(t, tc.hand1)
			hand2 := parseHandForTest(t, tc.hand2)

			var h1, h2 [5]Card
			copy(h1[:], hand1)
			copy(h2[:], hand2)

			score1 := Eval27Fast(&h1)
			score2 := Eval27Fast(&h2)
			result := int(score1) - int(score2)

			// Print scores for debugging
			t.Logf("Hand1 (%s) score: %d", tc.hand1, score1)
			t.Logf("Hand2 (%s) score: %d", tc.hand2, score2)

			if tc.want1Better && result >= 0 {
				t.Errorf("Expected %s to be better than %s", tc.hand1, tc.hand2)
			}
			if !tc.want1Better && result <= 0 {
				t.Errorf("Expected %s to be worse than %s", tc.hand1, tc.hand2)
			}
		})
	}
}

func TestDeuce7TableCompleteness(t *testing.T) {
	// Force table initialization
	initDeuce7Table()

	// Count how many unique ranks we find
	rankCounts := make(map[int16]int)
	totalHands := 0

	// Generate all possible 5-card hands
	var cards [5]Card
	for i := 0; i < 48; i++ {
		cards[0] = Card(i)
		for j := i + 1; j < 49; j++ {
			cards[1] = Card(j)
			for k := j + 1; k < 50; k++ {
				cards[2] = Card(k)
				for l := k + 1; l < 51; l++ {
					cards[3] = Card(l)
					for m := l + 1; m < 52; m++ {
						cards[4] = Card(m)

						// Get the rank from the lookup table
						rank := Eval27Fast(&cards)
						rankCounts[rank]++
						totalHands++

						// Also evaluate using slow method to verify consistency
						slowEval, err := evalSlow27(cards[:], true, false)
						if err != nil {
							t.Errorf("evalSlow27 failed for hand %v: %v", cards, err)
							continue
						}
						slowRank := evalInfo.slowRankToPacked[slowEval.rank]

						if rank != slowRank {
							t.Errorf("Inconsistent evaluation for hand %v: fast=%d, slow=%d",
								cards, rank, slowRank)
						}
					}
				}
			}
		}
	}

	// Verify we have the expected number of hands
	expectedHands := 2598960 // (52 choose 5)
	if totalHands != expectedHands {
		t.Errorf("Expected %d total hands, got %d", expectedHands, totalHands)
	}

	// Verify we have a reasonable number of unique ranks
	// In 2-7 lowball, we expect several thousand unique ranks
	if len(rankCounts) < 1000 {
		t.Errorf("Expected at least 1000 unique ranks, got %d", len(rankCounts))
	}

	// Print some statistics
	t.Logf("Total hands evaluated: %d", totalHands)
	t.Logf("Unique ranks found: %d", len(rankCounts))
}

// func BenchmarkDeuce7FastEval(b *testing.B) {
// 	// Force table initialization before benchmarking
// 	initDeuce7Table()

// 	// Create a sample hand (perfect 2-7 lowball hand)
// 	hand := [5]Card{
// 		mustMakeCard(Club, Two),
// 		mustMakeCard(Heart, Three),
// 		mustMakeCard(Diamond, Four),
// 		mustMakeCard(Spade, Five),
// 		mustMakeCard(Club, Seven),
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		Eval27Fast(&hand)
// 	}
// }
