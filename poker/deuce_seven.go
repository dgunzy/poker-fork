package poker

import (
	"fmt"
	"sort"
)

// Eval27 evaluates a 5-card poker hand for 2-7 lowball rules.
// Higher scores indicate worse hands in 2-7.
// Eval27 evaluates a 5-card poker hand for 2-7 lowball rules
func Eval27(hand *[5]Card) int16 {
	// Try to use the lookup table first
	if deuce7RootTable != nil {
		return Eval27Fast(hand)
	}

	// Fall back to slow evaluation if table isn't initialized
	ev, err := evalSlow27(hand[:], true, false)
	if err != nil {
		return 0 // Or some error value
	}
	return evalInfo.slowRankToPacked[ev.rank]
}

// Compare27 compares two 5-card 2-7 lowball hands.
// Returns:
//
//	negative if hand1 is better (has lower score in 2-7)
//	0 if hands are equal
//	positive if hand2 is better
func Compare27(hand1, hand2 *[5]Card) int {
	score1 := Eval27(hand1)
	score2 := Eval27(hand2)
	return int(score1) - int(score2)
}

// sortCards sorts cards in descending order by rank for 2-7 evaluation
func SortCards(cards []Card) {
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

// evalSlow evaluates a 3- or 5- card poker hand.
// The result is a number which can be compared
// with other hand's evaluations to correctly rank them as poker
// hands.
// If "replace" is false, then details are dropped of hands that can't be
// used for comparing against hands which are drawn from the same
// deck (for example: the kickers with trip aces don't matter).
//
// This function is used to build tables for fast hand evaluation.
// It's slow, but a little bit optimized so that the table construction
// is relatively fast.

func evalSlow27(c []Card, replace, text bool) (eval, error) {

	// fmt.Printf("Evaluating hand: %v\n", Hand(c).String())
	if len(c) != 5 {
		return eval{}, fmt.Errorf("evalSlow27: need 5 cards, got %d", len(c))
	}

	// Initial setup for hand evaluation
	flush := isFlush(c)
	ranks := [13]int{}
	dupes := [6]int{}  // uniqs, pairs, trips, quads, quins
	str8s := [13]int{} // finds straights
	str8top := 0       // set to the top card of the straight, if any
	var rankBits [6]uint16

	// First pass - check for straights
	for _, ci := range c {
		cr := (int(ci>>2) & 15) + 1
		if cr == 1 { // If it's an Ace
			cr = 14 // Treat it as higher than King
		}

		// We keep the original ranking logic for straight detection
		rawr := (cr + 11) % 13
		rankBits[ranks[rawr]] |= 1 << rawr
		ranks[rawr]++
		dupes[ranks[rawr]]++
		dupes[ranks[rawr]-1]--

		// Straight detection logic
		for i := 0; i < 5; i++ {
			var idx int
			if cr != 14 {
				idx = ((cr - 1) + i) % 13
			} else {
				idx = (13 + i) % 13
			}
			str8s[idx] |= 1 << uint(i)

			if str8s[idx] == 31 && idx >= 5 {
				str8top = (idx+12)%13 + 1
			}
		}
	}

	// Second pass - reset and properly evaluate hand strength
	for _, ci := range c {
		cr := (int(ci>>2) & 15) + 1
		if cr == 1 { // Found an Ace
			// Clear all tracking arrays for fresh evaluation
			ranks = [13]int{}
			dupes = [6]int{}
			rankBits = [6]uint16{}

			// Reprocess all cards with corrected ranking
			for j := 0; j < len(c); j++ {
				cr := (int(c[j]>>2) & 15) + 1
				if cr == 1 {
					// Ace becomes highest rank (12)
					rawr := 12
					rankBits[ranks[rawr]] |= 1 << rawr
					ranks[rawr]++
					dupes[ranks[rawr]]++
					dupes[ranks[rawr]-1]--
				} else {
					// Other cards use positions 1-11
					// This is the key change - we use (cr - 1) instead of (cr + 11) % 13
					// This ensures no bits are set at position 0
					rawr := (cr - 1)
					rankBits[ranks[rawr]] |= 1 << rawr
					ranks[rawr]++
					dupes[ranks[rawr]]++
					dupes[ranks[rawr]-1]--
				}
			}
			break // Only need to do this once if we find an Ace
		}
	}
	// fmt.Printf("After ace reset: rankBits[0]: %016b\n", rankBits[0])
	// fmt.Printf("After ace reset: ranks: %v\n", ranks)
	// fmt.Printf("After ace reset: dupes: %v\n", dupes)
	// fmt.Printf("flush: %v, str8top: %v, dupes[1]: %v\n", flush, str8top, dupes[1])
	rankBits[0] &^= rankBits[1]
	rankBits[1] &^= rankBits[2]
	rankBits[2] &^= rankBits[3]
	rankBits[3] &^= rankBits[4]
	rankBits[4] &^= rankBits[5]
	// fmt.Printf("After masking: rankBits[0]: %016b\n", rankBits[0])
	if !flush && str8top == 0 && dupes[1] == len(c) { // No pair

		var a, b, c, d, e int
		// fmt.Printf("Before popping: rankBits[0]: %016b\n", rankBits[0])
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		d, rankBits[0] = poptop(rankBits[0])
		e, rankBits[0] = poptop(rankBits[0])

		if text {
			return evalScore("%s-%s-%s-%s-%s", 0, a, b, c, d, e), nil
		}
		return evalScore5(0, a, b, c, d, e), nil
	}
	if dupes[2] == 1 && dupes[3] == 0 { // One pair
		var p, a, b, c int
		p, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%s-%s-%s", 1, p, a, b, c), nil
		}
		return evalScore5(1, p, a, b, c, 0), nil
	}
	if dupes[2] == 2 { // Two pair
		var p, q, a int
		p, rankBits[1] = poptop(rankBits[1])
		q, rankBits[1] = poptop(rankBits[1])
		a, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%[1]s%[1]s-%[2]s%[2]s-%[3]s", 2, p, q, a), nil
		}
		return evalScore5(2, p, q, a, 0, 0), nil
	}
	if dupes[3] == 1 && dupes[2] == 0 { // Trips
		if replace {
			var t, a, b int
			a, rankBits[0] = poptop(rankBits[0])
			b, rankBits[0] = poptop(rankBits[0])
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%s-%s", 3, t, a, b), nil
			}
			return evalScore5(3, t, a, b, 0, 0), nil
		}
		if len(c) == 5 {
			var t int
			t, rankBits[2] = poptop(rankBits[2])
			if text {
				return evalScore("%[1]s%[1]s%[1]s-x-y", 3, t), nil // ignore kickers
			}
			return evalScore5(3, t, 0, 0, 0, 0), nil
		}
		var t int
		t, rankBits[2] = poptop(rankBits[2])
		if text {
			return evalScore("%[1]s%[1]s%[1]s", 3, t), nil
		}
		return evalScore5(3, t, 0, 0, 0, 0), nil
	}
	if str8top != 0 && !flush { // Straight
		if text {
			return evalScore("%s straight", 4, (str8top+11)%13+2), nil
		}
		return evalScore5(4, (str8top+11)%13+2, 0, 0, 0, 0), nil
	}
	if flush && str8top == 0 { // Flush
		var a, b, c, d, e int
		a, rankBits[0] = poptop(rankBits[0])
		b, rankBits[0] = poptop(rankBits[0])
		c, rankBits[0] = poptop(rankBits[0])
		d, rankBits[0] = poptop(rankBits[0])
		e, rankBits[0] = poptop(rankBits[0])
		if text {
			return evalScore("%s%s%s%s%s flush", 5, a, b, c, d, e), nil
		}
		return evalScore5(5, a, b, c, d, e), nil
	}
	if dupes[2] == 1 && dupes[3] == 1 { // Full house
		var t, p int
		t, rankBits[2] = poptop(rankBits[2])
		p, rankBits[1] = poptop(rankBits[1])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s-%[2]s%[2]s", 6, t, p), nil
			}
			return evalScore5(6, t, p, 0, 0, 0), nil
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s-xx", 6, t), nil // ignore lower pair
		}
		return evalScore5(6, t, 0, 0, 0, 0), nil // ignore lower pair
	}
	if dupes[4] == 1 { // Quads
		var q, a int
		q, rankBits[3] = poptop(rankBits[3])
		a, rankBits[0] = poptop(rankBits[0])
		if replace {
			if text {
				return evalScore("%[1]s%[1]s%[1]s%[1]s-%[2]s", 7, q, a), nil
			}
			return evalScore5(7, q, a, 0, 0, 0), nil
		}
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s-x", 7, q), nil // ignore kicker
		}
		return evalScore5(7, q, 0, 0, 0, 0), nil
	}
	if str8top != 0 && flush { // Straight flush
		if text {
			return evalScore("%s straight flush", 8, (str8top+11)%13+2), nil
		}
		return evalScore5(8, (str8top+11)%13+2, 0, 0, 0, 0), nil
	}
	if dupes[5] == 1 { // 5-kind
		var q int
		q, rankBits[4] = poptop(rankBits[4])
		if text {
			return evalScore("%[1]s%[1]s%[1]s%[1]s%[1]s", 9, q), nil
		}
		return evalScore5(9, q, 0, 0, 0, 0), nil
	}
	return eval{}, fmt.Errorf("failed to eval hand %v", c)
}
