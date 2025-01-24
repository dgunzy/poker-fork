package poker

import (
	"fmt"
	"sync"
)

var (
	deuce7RootTable []int16
	deuce7TableInit sync.Once
)

// initDeuce7Table initializes the lookup table for 2-7 lowball evaluation
func initDeuce7Table() {
	deuce7TableInit.Do(func() {
		// Initialize table with sentinel values
		deuce7RootTable = make([]int16, 7462)
		for i := range deuce7RootTable {
			deuce7RootTable[i] = -1
		}

		// Generate all possible 5-card hands
		var cards [5]Card
		for i := 0; i < 52; i++ {
			cards[0] = Card(i)
			for j := i + 1; j < 52; j++ {
				cards[1] = Card(j)
				for k := j + 1; k < 52; k++ {
					cards[2] = Card(k)
					for l := k + 1; l < 52; l++ {
						cards[3] = Card(l)
						for m := l + 1; m < 52; m++ {
							cards[4] = Card(m)

							// Create a copy for evaluation
							cardSlice := make([]Card, 5)
							copy(cardSlice, cards[:])

							// Sort before evaluation (same as in evalSlow27)
							SortCards(cardSlice)

							// Evaluate using slow method
							eval, err := evalSlow27(cardSlice, true, false)
							if err != nil {
								panic(fmt.Sprintf("Failed to evaluate hand %v: %v", cardSlice, err))
							}

							// Store using the same sorted hand
							idx := perfectHash(cardSlice)
							if idx < 0 || idx >= len(deuce7RootTable) {
								panic(fmt.Sprintf("Invalid hash index %d for hand %v", idx, cardSlice))
							}

							slowRank := evalInfo.slowRankToPacked[eval.rank]
							if slowRank == 0 {
								panic(fmt.Sprintf("Invalid rank 0 for hand %v", cardSlice))
							}

							deuce7RootTable[idx] = slowRank
						}
					}
				}
			}
		}

		// Verify no unfilled entries
		for i, v := range deuce7RootTable {
			if v == -1 {
				panic(fmt.Sprintf("Unfilled table entry at index %d", i))
			}
		}
	})
}

// getHandIndex calculates a unique index for a 5-card hand
func getHandIndex(hand *[5]Card) int {
	// This is a simplified version - you'll need to implement
	// a proper perfect hash function for the hands
	var idx int
	for _, c := range hand {
		idx = idx*52 + int(c)
	}
	return idx % 7462
}

// Eval27Fast uses the lookup table for faster evaluation
func Eval27Fast(hand *[5]Card) int16 {
	if deuce7RootTable == nil {
		initDeuce7Table()
	}

	// Make a copy and sort the cards for consistent lookup
	var sortedHand [5]Card
	copy(sortedHand[:], hand[:])
	SortCards(sortedHand[:])

	// Use the same hash function used during table initialization
	return deuce7RootTable[perfectHash(sortedHand[:])]
}

// perfectHash implements a perfect hash for 5-card poker hands
func perfectHash(cards []Card) int {
	var val uint32
	for i, c := range cards {
		rank := uint32((c >> 2) & 0xF)
		suit := uint32(c & 0x3)
		val += (rank*4 + suit) * uint32(pow(53, i))
	}
	return int(val % 7462)
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// binomial calculates "n choose k"
func binomial(n, k int) int {
	if k > n {
		return 0
	}
	if k*2 > n {
		k = n - k
	}
	if k == 0 {
		return 1
	}

	result := n
	for i := 2; i <= k; i++ {
		result *= (n - i + 1)
		result /= i
	}
	return result
}
